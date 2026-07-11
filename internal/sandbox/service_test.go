package sandbox

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/zzm/opcv2/internal/account"
	"github.com/zzm/opcv2/internal/ai"
	"github.com/zzm/opcv2/internal/membership"
)

type fakeRepository struct {
	created  Session
	session  Session
	sessions []Session
	err      error
	messages []Message
	statuses []string
}

func (r *fakeRepository) UpdateSessionStatus(_ context.Context, userID, id int64, status string) error {
	if r.err != nil {
		return r.err
	}
	if r.session.UserID != userID || r.session.ID != id {
		return ErrSessionNotFound
	}
	r.session.Status = status
	r.statuses = append(r.statuses, status)
	return nil
}

func (r *fakeRepository) CreateMessage(_ context.Context, message Message) (Message, error) {
	message.ID = int64(len(r.messages) + 1)
	r.messages = append(r.messages, message)
	return message, r.err
}

func (r *fakeRepository) ListMessages(_ context.Context, userID, sessionID int64) ([]Message, error) {
	rows := make([]Message, 0, len(r.messages))
	for _, message := range r.messages {
		if message.UserID == userID && message.SessionID == sessionID {
			rows = append(rows, message)
		}
	}
	return rows, r.err
}

func (r *fakeRepository) UpdateSessionDraft(_ context.Context, userID, id int64, update DraftUpdate) (Session, error) {
	if r.err != nil {
		return Session{}, r.err
	}
	if r.session.UserID != userID || r.session.ID != id || r.session.Status != StatusDraft {
		return Session{}, ErrSessionNotFound
	}
	if update.Goal != nil {
		r.session.Goal = *update.Goal
	}
	if update.TargetUsers != nil {
		r.session.TargetUsers = *update.TargetUsers
	}
	if update.Product != nil {
		r.session.Product = *update.Product
	}
	if update.Roles != nil {
		r.session.Roles = *update.Roles
	}
	return r.session, nil
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

type fakeProfileContextProvider struct {
	context account.ProfileContext
	err     error
}

func (p *fakeProfileContextProvider) GetProfileContext(context.Context, int64) (account.ProfileContext, error) {
	return p.context, p.err
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

func TestServiceUpdatesOwnedDraftAndListsRoles(t *testing.T) {
	repository := &fakeRepository{session: Session{ID: 99, UserID: 42, Goal: "旧目标", TargetUsers: "连锁门店", Product: "AI运营平台", Status: StatusDraft}}
	service := NewService(repository, nil)
	goal := " 验证企业AI项目 "
	roles := []string{" 用户视角 ", "用户视角", "投资人视角"}

	session, err := service.UpdateSessionDraft(context.Background(), 42, 99, DraftUpdate{Goal: &goal, Roles: &roles})
	if err != nil {
		t.Fatalf("UpdateSessionDraft() error = %v", err)
	}
	if session.Goal != "验证企业AI项目" || len(session.Roles) != 2 {
		t.Fatalf("session = %+v", session)
	}
	roleCatalog := service.ListRoles()
	if len(roleCatalog) < 7 || roleCatalog[0].Key == "" || roleCatalog[0].Label == "" {
		t.Fatalf("roles = %+v", roleCatalog)
	}
}

func TestServiceRejectsRunningIncompleteDraft(t *testing.T) {
	service := NewService(&fakeRepository{session: Session{ID: 99, UserID: 42, Goal: "验证项目", Status: StatusDraft}}, &fakeGenerator{})

	_, err := service.RunSession(context.Background(), 42, 99)

	if !errors.Is(err, ErrInvalidSession) {
		t.Fatalf("err = %v, want ErrInvalidSession", err)
	}
}

func TestServiceRunSessionAddsProfileContextToPrompt(t *testing.T) {
	report := Report{
		Score:         83,
		Summary:       "可以先做小范围验证",
		Metrics:       []Metric{{Label: "市场吸引力", Value: "8.4"}},
		RoleSummaries: []RoleSummary{{Role: "用户", View: "关注响应效率和数据安全"}},
		Risks:         []string{"客户教育成本高"},
		NextActions:   []string{"访谈 10 个目标客户"},
	}
	payload, _ := json.Marshal(report)
	generator := &fakeGenerator{content: payload}
	profile := &fakeProfileContextProvider{context: account.ProfileContext{
		UserID:    42,
		Completed: true,
		Groups: []account.ProfileGroup{
			{Key: account.ProfileGroupBusiness, Title: "我的业务/公司", Fields: map[string]string{"company": "智活AI", "stage": "启动"}},
			{Key: account.ProfileGroupResources, Title: "能力与资源", Fields: map[string]string{"budget": "3万以内"}},
		},
	}}
	service := NewService(&fakeRepository{session: Session{
		ID:          99,
		UserID:      42,
		Status:      StatusDraft,
		Goal:        "验证企业服务项目",
		TargetUsers: "中小企业老板",
		Product:     "AI 顾问服务",
		Roles:       []string{"用户"},
	}}, generator, WithProfileContextProvider(profile))

	_, err := service.RunSession(context.Background(), 42, 99)

	if err != nil {
		t.Fatalf("RunSession() error = %v", err)
	}
	if !strings.Contains(generator.request.UserPrompt, "用户画像上下文") ||
		!strings.Contains(generator.request.UserPrompt, "智活AI") ||
		!strings.Contains(generator.request.UserPrompt, "3万以内") {
		t.Fatalf("prompt missing profile context: %s", generator.request.UserPrompt)
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
	if len(repository.statuses) != 1 || repository.statuses[0] != StatusRunning {
		t.Fatalf("statuses = %+v", repository.statuses)
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
	if repository.session.Status != StatusFailed {
		t.Fatalf("status = %s, want failed", repository.session.Status)
	}
}

func TestServiceAsksSelectedSandboxRoleAndPersistsAnswer(t *testing.T) {
	repository := &fakeRepository{session: Session{ID: 99, UserID: 42, Product: "AI运营平台", Roles: []string{"投资人视角"}, Status: StatusCompleted}}
	generator := &fakeGenerator{content: []byte(`{"answer":"我会重点看客户留存与单位经济模型。"}`)}
	service := NewService(repository, generator)

	message, err := service.AskRole(context.Background(), AskRoleInput{UserID: 42, SessionID: 99, Role: " 投资人视角 ", Question: " 你最关注哪些指标？ "})
	if err != nil {
		t.Fatalf("AskRole() error = %v", err)
	}
	if message.Role != "投资人视角" || message.Question != "你最关注哪些指标？" || message.Answer == "" {
		t.Fatalf("message = %+v", message)
	}
	if generator.request.Feature != "sandbox.follow_up" || !strings.Contains(generator.request.UserPrompt, "投资人视角") {
		t.Fatalf("request = %+v", generator.request)
	}
}
