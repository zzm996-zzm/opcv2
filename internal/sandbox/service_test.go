package sandbox

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/zzm/opcv2/internal/ai"
	"github.com/zzm/opcv2/internal/membership"
)

type fakeRepository struct {
	created  Session
	session  Session
	sessions []Session
	err      error
}

func (r *fakeRepository) CreateSession(_ context.Context, session Session) (Session, error) {
	r.created = session
	session.ID = 99
	session.UpdatedAt = session.CreatedAt
	r.session = session
	return session, r.err
}

func (r *fakeRepository) UpdateSessionResult(_ context.Context, userID, id int64, result Report) (Session, error) {
	if r.err != nil {
		return Session{}, r.err
	}
	if r.session.UserID != userID || r.session.ID != id {
		return Session{}, ErrSessionNotFound
	}
	r.session.Status = StatusCompleted
	r.session.Report = result
	return r.session, nil
}

func (r *fakeRepository) ListSessions(_ context.Context, userID int64, limit int) ([]Session, error) {
	if r.err != nil {
		return nil, r.err
	}
	var rows []Session
	for _, session := range r.sessions {
		if session.UserID == userID {
			rows = append(rows, session)
		}
	}
	return rows[:min(len(rows), limit)], nil
}

func (r *fakeRepository) GetSession(_ context.Context, userID, id int64) (Session, error) {
	if r.err != nil {
		return Session{}, r.err
	}
	if r.session.UserID != userID || r.session.ID != id {
		return Session{}, ErrSessionNotFound
	}
	return r.session, nil
}

type fakeGenerator struct {
	request ai.GenerateJSONRequest
	content []byte
	calls   int
	err     error
}

func (g *fakeGenerator) GenerateJSON(_ context.Context, request ai.GenerateJSONRequest) (ai.GenerateJSONResult, error) {
	g.calls++
	g.request = request
	if g.err != nil {
		return ai.GenerateJSONResult{}, g.err
	}
	if request.Validate != nil {
		if err := request.Validate(g.content); err != nil {
			return ai.GenerateJSONResult{}, err
		}
	}
	return ai.GenerateJSONResult{Content: g.content}, nil
}

type fakeQuotaConsumer struct {
	consumed []membership.ConsumeInput
	err      error
}

func (c *fakeQuotaConsumer) CheckAndConsume(_ context.Context, input membership.ConsumeInput) (membership.UsageItem, error) {
	c.consumed = append(c.consumed, input)
	if c.err != nil {
		return membership.UsageItem{}, c.err
	}
	return membership.UsageItem{Key: input.FeatureKey, Used: 1, Limit: 20}, nil
}

func TestServiceCreatesDraftSession(t *testing.T) {
	now := time.Date(2026, 6, 30, 10, 0, 0, 0, time.UTC)
	repository := &fakeRepository{}
	service := NewService(repository, nil)
	service.now = func() time.Time { return now }

	session, err := service.CreateSession(context.Background(), CreateInput{
		UserID:      42,
		Goal:        "验证面向教培机构的 AI 客服工具是否值得启动",
		TargetUsers: "30-200 人的本地教培机构",
		Product:     "AI 客服 + 企微转化助手",
		Roles:       []string{"用户", "投资人", "增长顾问"},
	})

	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}
	if session.ID != 99 || session.Status != StatusDraft {
		t.Fatalf("session = %+v", session)
	}
	if repository.created.UserID != 42 || repository.created.Goal == "" || len(repository.created.Roles) != 3 {
		t.Fatalf("created = %+v", repository.created)
	}
}

func TestServiceRunSessionConsumesSandboxQuota(t *testing.T) {
	report := Report{
		Score:         83,
		Summary:       "可以先做小范围验证",
		Metrics:       []Metric{{Label: "市场吸引力", Value: "8.4"}},
		RoleSummaries: []RoleSummary{{Role: "用户", View: "关注响应效率和数据安全"}},
		Risks:         []string{"客户教育成本高"},
		NextActions:   []string{"访谈 10 个目标客户"},
	}
	payload, _ := json.Marshal(report)
	quota := &fakeQuotaConsumer{}
	service := NewService(&fakeRepository{session: Session{
		ID:          99,
		UserID:      42,
		Status:      StatusDraft,
		Goal:        "验证项目",
		TargetUsers: "教培机构",
		Product:     "AI 工具",
		Roles:       []string{"用户"},
	}}, &fakeGenerator{content: payload}, WithQuotaConsumer(quota))

	_, err := service.RunSession(context.Background(), 42, 99)

	if err != nil {
		t.Fatalf("RunSession() error = %v", err)
	}
	if len(quota.consumed) != 1 {
		t.Fatalf("consumed = %+v, want one quota consume", quota.consumed)
	}
	consumed := quota.consumed[0]
	if consumed.FeatureKey != membership.FeatureSandboxRuns || consumed.IdempotencyKey != "sandbox-run-99" {
		t.Fatalf("consumed = %+v", consumed)
	}
}

func TestServiceRunSessionStopsWhenSandboxQuotaExceeded(t *testing.T) {
	generator := &fakeGenerator{content: []byte(`{}`)}
	service := NewService(&fakeRepository{session: Session{
		ID:          99,
		UserID:      42,
		Status:      StatusDraft,
		Goal:        "验证项目",
		TargetUsers: "教培机构",
		Product:     "AI 工具",
		Roles:       []string{"用户"},
	}}, generator, WithQuotaConsumer(&fakeQuotaConsumer{err: membership.ErrQuotaExceeded}))

	_, err := service.RunSession(context.Background(), 42, 99)

	if !errors.Is(err, membership.ErrQuotaExceeded) {
		t.Fatalf("err = %v, want ErrQuotaExceeded", err)
	}
	if generator.calls != 0 {
		t.Fatalf("generator calls = %d, want 0", generator.calls)
	}
}

func TestServiceRunSessionGeneratesReportThroughAI(t *testing.T) {
	report := Report{
		Score:         83,
		Summary:       "可以先做小范围验证",
		Metrics:       []Metric{{Label: "市场吸引力", Value: "8.4"}},
		RoleSummaries: []RoleSummary{{Role: "用户", View: "关注响应效率和数据安全"}},
		Risks:         []string{"客户教育成本高"},
		NextActions:   []string{"访谈 10 个目标客户"},
	}
	payload, _ := json.Marshal(report)
	repository := &fakeRepository{session: Session{
		ID:          99,
		UserID:      42,
		Status:      StatusDraft,
		Goal:        "验证面向教培机构的 AI 客服工具是否值得启动",
		TargetUsers: "30-200 人的本地教培机构",
		Product:     "AI 客服 + 企微转化助手",
		Roles:       []string{"用户", "投资人"},
	}}
	generator := &fakeGenerator{content: payload}
	service := NewService(repository, generator)

	session, err := service.RunSession(context.Background(), 42, 99)

	if err != nil {
		t.Fatalf("RunSession() error = %v", err)
	}
	if session.Status != StatusCompleted || session.Report.Score != 83 {
		t.Fatalf("session = %+v", session)
	}
	if generator.request.Feature != "sandbox.run" || generator.request.SchemaName != "sandbox_report" {
		t.Fatalf("ai request = %+v", generator.request)
	}
}

func TestServiceRunSessionRejectsOtherUsersSession(t *testing.T) {
	repository := &fakeRepository{session: Session{ID: 99, UserID: 7, Status: StatusDraft}}
	service := NewService(repository, &fakeGenerator{content: []byte(`{}`)})

	_, err := service.RunSession(context.Background(), 42, 99)

	if !errors.Is(err, ErrSessionNotFound) {
		t.Fatalf("err = %v, want ErrSessionNotFound", err)
	}
}

func TestServiceReturnsSafeErrorForInvalidAIReport(t *testing.T) {
	repository := &fakeRepository{session: Session{
		ID:          99,
		UserID:      42,
		Status:      StatusDraft,
		Goal:        "验证一个项目",
		TargetUsers: "教培机构",
		Product:     "AI 工具",
		Roles:       []string{"用户"},
	}}
	service := NewService(repository, &fakeGenerator{content: []byte(`{"score":0}`)})

	_, err := service.RunSession(context.Background(), 42, 99)

	if !errors.Is(err, ErrInvalidAIResult) {
		t.Fatalf("err = %v, want ErrInvalidAIResult", err)
	}
}
