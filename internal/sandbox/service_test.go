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
	"github.com/zzm/opcv2/internal/jobs"
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

func (r *fakeRepository) PrepareSessionRun(_ context.Context, userID, id int64) (Session, error) {
	if r.err != nil {
		return Session{}, r.err
	}
	if r.session.UserID != userID || r.session.ID != id {
		return Session{}, ErrSessionNotFound
	}
	if r.session.Status != StatusDraft && r.session.Status != StatusFailed && r.session.Status != StatusCanceled {
		return Session{}, ErrInvalidSession
	}
	r.session.Status = StatusQueued
	r.session.ProgressPercent = 0
	r.session.CurrentStep = StatusQueued
	r.session.ErrorMessage = ""
	r.session.RunAttempt++
	r.statuses = append(r.statuses, StatusQueued)
	return r.session, nil
}

func (r *fakeRepository) UpdateSessionProgress(_ context.Context, userID, id int64, attempt int, status string, progress int, step, errorMessage string) (Session, error) {
	if r.err != nil {
		return Session{}, r.err
	}
	if r.session.UserID != userID || r.session.ID != id || r.session.RunAttempt != attempt {
		return Session{}, ErrStaleRun
	}
	r.session.Status, r.session.ProgressPercent, r.session.CurrentStep, r.session.ErrorMessage = status, progress, step, errorMessage
	r.statuses = append(r.statuses, status)
	return r.session, nil
}

func (r *fakeRepository) CancelSession(_ context.Context, userID, id int64) (Session, error) {
	if r.session.UserID != userID || r.session.ID != id || (r.session.Status != StatusQueued && r.session.Status != StatusRunning) {
		return Session{}, ErrInvalidSession
	}
	r.session.Status, r.session.ProgressPercent, r.session.CurrentStep = StatusCanceled, 100, StatusCanceled
	return r.session, nil
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
	if update.Settings != nil {
		r.session.Settings = *update.Settings
	}
	return r.session, nil
}

func (r *fakeRepository) UpdateSessionIntake(_ context.Context, userID, id int64, intake Intake) (Session, error) {
	if r.err != nil {
		return Session{}, r.err
	}
	if r.session.UserID != userID || r.session.ID != id || r.session.Status != StatusDraft {
		return Session{}, ErrSessionNotFound
	}
	r.session.Intake = intake
	r.session.CurrentStep = intake.Status
	return r.session, nil
}

func (r *fakeRepository) UpdateSessionSettings(_ context.Context, userID, id int64, settings RunSettings) (Session, error) {
	if r.err != nil {
		return Session{}, r.err
	}
	if r.session.UserID != userID || r.session.ID != id || r.session.Status != StatusDraft {
		return Session{}, ErrSessionNotFound
	}
	r.session.Settings = settings
	return r.session, nil
}

func (r *fakeRepository) CreateSession(_ context.Context, session Session) (Session, error) {
	r.created = session
	session.ID = 99
	session.UpdatedAt = session.CreatedAt
	r.session = session
	return session, r.err
}

func (r *fakeRepository) UpdateSessionResult(_ context.Context, userID, id int64, attempt int, result Report) (Session, error) {
	if r.err != nil {
		return Session{}, r.err
	}
	if r.session.UserID != userID || r.session.ID != id || r.session.RunAttempt != attempt || r.session.Status != StatusRunning {
		return Session{}, ErrStaleRun
	}
	r.session.Status = StatusCompleted
	r.session.Report = result
	r.session.ProgressPercent = 100
	r.session.CurrentStep = StatusCompleted
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
	refunded []membership.ConsumeInput
	err      error
}

func (c *fakeQuotaConsumer) CheckAndConsume(_ context.Context, input membership.ConsumeInput) (membership.UsageItem, error) {
	c.consumed = append(c.consumed, input)
	if c.err != nil {
		return membership.UsageItem{}, c.err
	}
	return membership.UsageItem{Key: input.FeatureKey, Used: 1, Limit: 20}, nil
}

func (c *fakeQuotaConsumer) RefundUsage(_ context.Context, input membership.ConsumeInput) (membership.UsageItem, error) {
	c.refunded = append(c.refunded, input)
	return membership.UsageItem{Key: input.FeatureKey}, c.err
}

type fakeQueue struct {
	jobs []jobs.Job
	err  error
}

func (q *fakeQueue) Enqueue(_ context.Context, job jobs.Job) error {
	q.jobs = append(q.jobs, job)
	return q.err
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
	service := NewService(&fakeRepository{session: Session{ID: 99, UserID: 42, Goal: "验证项目", Status: StatusDraft}}, &fakeGenerator{}, WithQueue(&fakeQueue{}))

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
	repository := &fakeRepository{session: Session{
		ID:          99,
		UserID:      42,
		Status:      StatusDraft,
		Goal:        "验证企业服务项目",
		TargetUsers: "中小企业老板",
		Product:     "AI 顾问服务",
		Roles:       []string{"用户"},
	}}
	service := NewService(repository, generator, WithProfileContextProvider(profile), WithQueue(&fakeQueue{}))

	queued, err := service.RunSession(context.Background(), 42, 99)

	if err != nil {
		t.Fatalf("RunSession() error = %v", err)
	}
	if err := service.ProcessSession(context.Background(), 42, 99, queued.RunAttempt); err != nil {
		t.Fatalf("ProcessSession() error = %v", err)
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
	}}, &fakeGenerator{content: payload}, WithQuotaConsumer(quota), WithQueue(&fakeQueue{}))

	_, err := service.RunSession(context.Background(), 42, 99)

	if err != nil {
		t.Fatalf("RunSession() error = %v", err)
	}
	if len(quota.consumed) != 1 {
		t.Fatalf("consumed = %+v, want one quota consume", quota.consumed)
	}
	consumed := quota.consumed[0]
	if consumed.FeatureKey != membership.FeatureSandboxRuns || consumed.IdempotencyKey != "sandbox-run-99-attempt-1" {
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
	}}, generator, WithQuotaConsumer(&fakeQuotaConsumer{err: membership.ErrQuotaExceeded}), WithQueue(&fakeQueue{}))

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
	queue := &fakeQueue{}
	service := NewService(repository, generator, WithQueue(queue))

	queued, err := service.RunSession(context.Background(), 42, 99)

	if err != nil {
		t.Fatalf("RunSession() error = %v", err)
	}
	if queued.Status != StatusQueued || len(queue.jobs) != 1 || queue.jobs[0].Type != jobs.TypeSandboxRun {
		t.Fatalf("queued/job = %+v/%+v", queued, queue.jobs)
	}
	if err := service.ProcessSession(context.Background(), 42, 99, queued.RunAttempt); err != nil {
		t.Fatalf("ProcessSession() error = %v", err)
	}
	session := repository.session
	if session.Status != StatusCompleted || session.Report.Score != 83 {
		t.Fatalf("session = %+v", session)
	}
	if session.Report.Basis != "model_simulation" || session.Report.Disclaimer == "" || len(session.Report.Assumptions) == 0 || len(session.Report.EvidenceSources) != 0 {
		t.Fatalf("report metadata = %+v", session.Report)
	}
	if generator.request.Feature != "sandbox.run" || generator.request.SchemaName != "sandbox_report" {
		t.Fatalf("ai request = %+v", generator.request)
	}
	if len(repository.statuses) != 3 || repository.statuses[0] != StatusQueued || repository.statuses[1] != StatusRunning {
		t.Fatalf("statuses = %+v", repository.statuses)
	}
}

func TestServiceRunSessionRejectsOtherUsersSession(t *testing.T) {
	repository := &fakeRepository{session: Session{ID: 99, UserID: 7, Status: StatusDraft}}
	service := NewService(repository, &fakeGenerator{content: []byte(`{}`)}, WithQueue(&fakeQueue{}))

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
	quota := &fakeQuotaConsumer{}
	service := NewService(repository, &fakeGenerator{content: []byte(`{"score":0}`)}, WithQueue(&fakeQueue{}), WithQuotaConsumer(quota))

	queued, err := service.RunSession(context.Background(), 42, 99)
	if err == nil {
		err = service.ProcessSession(context.Background(), 42, 99, queued.RunAttempt)
	}

	if !errors.Is(err, ErrInvalidAIResult) {
		t.Fatalf("err = %v, want ErrInvalidAIResult", err)
	}
	if repository.session.Status != StatusFailed {
		t.Fatalf("status = %s, want failed", repository.session.Status)
	}
	if len(quota.refunded) != 1 || quota.refunded[0].IdempotencyKey != "sandbox-run-99-attempt-1-refund" {
		t.Fatalf("refunds = %+v", quota.refunded)
	}
}

func TestServiceCancelsQueuedRunAndRefundsQuota(t *testing.T) {
	quota := &fakeQuotaConsumer{}
	repository := &fakeRepository{session: Session{ID: 99, UserID: 42, Status: StatusDraft, Goal: "验证", TargetUsers: "客户", Product: "产品", Roles: []string{"用户"}}}
	service := NewService(repository, &fakeGenerator{}, WithQueue(&fakeQueue{}), WithQuotaConsumer(quota))
	queued, err := service.RunSession(context.Background(), 42, 99)
	if err != nil {
		t.Fatalf("RunSession() error = %v", err)
	}
	canceled, err := service.CancelSession(context.Background(), 42, 99)
	if err != nil || canceled.Status != StatusCanceled || queued.RunAttempt != canceled.RunAttempt {
		t.Fatalf("canceled = %+v err=%v", canceled, err)
	}
	if len(quota.refunded) != 1 {
		t.Fatalf("refunds = %+v", quota.refunded)
	}
	if err := service.ProcessSession(context.Background(), 42, 99, queued.RunAttempt); err != nil {
		t.Fatalf("canceled stale job should no-op: %v", err)
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

func TestServiceCreatesIntakeDraftFromAI(t *testing.T) {
	repository := &fakeRepository{}
	generator := &fakeGenerator{content: []byte(`{
		"goal":"验证项目是否值得投入",
		"target_users":"本地门店老板",
		"product":"AI运营助手",
		"recognized_fields":[{"key":"industry","label":"行业","value":"零售"}],
		"questions":[
			{"key":"pain","title":"最急需解决什么问题？","hint":"描述问题","placeholder":"请输入","required":true,"max_length":1000,"position":1},
			{"key":"model","title":"如何收费？","hint":"描述模式","placeholder":"请输入","required":false,"max_length":1000,"position":2}
		]
	}`)}
	service := NewService(repository, generator)

	session, err := service.CreateIntake(context.Background(), IntakeCreateInput{UserID: 42, InitialIdea: "想做一个门店运营助手"})
	if err != nil {
		t.Fatalf("CreateIntake() error = %v", err)
	}
	if session.Status != StatusDraft || session.CurrentStep != IntakeStatusQuestions || session.Intake.Status != IntakeStatusQuestions || session.Intake.TotalQuestions != 2 {
		t.Fatalf("session = %+v", session)
	}
	if session.Goal != "验证项目是否值得投入" || generator.request.Feature != "sandbox.intake" || generator.request.SchemaName != "sandbox_intake" {
		t.Fatalf("session/request = %+v/%+v", session, generator.request)
	}
}

func TestServiceAnswersAndCompletesIntake(t *testing.T) {
	repository := &fakeRepository{session: Session{
		ID: 99, UserID: 42, Status: StatusDraft,
		Intake: Intake{Status: IntakeStatusQuestions, Questions: []IntakeQuestion{
			{Key: "pain", Title: "痛点", Required: true, MaxLength: 20},
			{Key: "model", Title: "收费", Required: false, MaxLength: 20},
		}},
	}}
	service := NewService(repository, nil)

	if _, err := service.AnswerIntake(context.Background(), IntakeAnswerInput{UserID: 42, SessionID: 99, QuestionKey: "pain", Answer: "回复太慢"}); err != nil {
		t.Fatalf("required answer error = %v", err)
	}
	if _, err := service.AnswerIntake(context.Background(), IntakeAnswerInput{UserID: 42, SessionID: 99, QuestionKey: "model", Skipped: true}); err != nil {
		t.Fatalf("optional skip error = %v", err)
	}
	session, err := service.CompleteIntake(context.Background(), 42, 99)
	if err != nil {
		t.Fatalf("CompleteIntake() error = %v", err)
	}
	if session.Intake.Status != IntakeStatusReady || session.Intake.AnsweredCount != 2 || session.CurrentStep != IntakeStatusReady {
		t.Fatalf("session = %+v", session)
	}
}

func TestServiceAllowsExplicitlySkippingRequiredIntakeQuestion(t *testing.T) {
	repository := &fakeRepository{session: Session{
		ID: 99, UserID: 42, Status: StatusDraft,
		Intake: Intake{Status: IntakeStatusQuestions, Questions: []IntakeQuestion{
			{Key: "pain", Title: "痛点", Required: true, MaxLength: 20},
		}},
	}}
	service := NewService(repository, nil)

	if _, err := service.AnswerIntake(context.Background(), IntakeAnswerInput{UserID: 42, SessionID: 99, QuestionKey: "pain", Skipped: true}); err != nil {
		t.Fatalf("required question explicit skip error = %v", err)
	}
	completed, err := service.CompleteIntake(context.Background(), 42, 99)
	if err != nil || completed.Intake.Status != IntakeStatusReady {
		t.Fatalf("completed = %+v err=%v", completed, err)
	}
}

func TestServiceRejectsIncompleteIntakeAndInvalidSettings(t *testing.T) {
	repository := &fakeRepository{session: Session{ID: 99, UserID: 42, Status: StatusDraft, Intake: Intake{
		Status:    IntakeStatusQuestions,
		Questions: []IntakeQuestion{{Key: "pain", Title: "痛点", Required: true}},
	}}}
	service := NewService(repository, nil)
	if _, err := service.CompleteIntake(context.Background(), 42, 99); !errors.Is(err, ErrIntakeIncomplete) {
		t.Fatalf("CompleteIntake() error = %v, want incomplete", err)
	}
	if _, err := service.UpdateSessionSettings(context.Background(), 42, 99, RunSettings{Depth: "invalid"}); !errors.Is(err, ErrInvalidSession) {
		t.Fatalf("UpdateSessionSettings() error = %v, want invalid session", err)
	}
}

func TestServiceAcceptsRichReportAndBackfillsLegacyFields(t *testing.T) {
	repository := &fakeRepository{session: Session{
		ID: 99, UserID: 42, Status: StatusDraft, Goal: "验证项目", TargetUsers: "门店", Product: "助手", Roles: []string{"用户"},
	}}
	generator := &fakeGenerator{content: []byte(`{
		"report_version":"sandbox_report_v3",
		"consumer_probability":77,
		"core_conclusions":["先做三家试点"],
		"opportunity_analysis":[{"title":"高频场景","detail":"咨询量大","tags":["效率"]}],
		"risk_analysis":[{"title":"合规","detail":"需要授权","tags":["风险"]}],
		"action_plan":[{"order":1,"title":"客户访谈","detail":"访谈十家","duration":"一周"}],
		"validation_metrics":[{"label":"响应时间","current":"15分钟","target":"1分钟内","confidence_percent":70}]
	}`)}
	service := NewService(repository, generator, WithQueue(&fakeQueue{}))
	queued, err := service.RunSession(context.Background(), 42, 99)
	if err != nil {
		t.Fatalf("RunSession() error = %v", err)
	}
	if err := service.ProcessSession(context.Background(), 42, 99, queued.RunAttempt); err != nil {
		t.Fatalf("ProcessSession() error = %v", err)
	}
	report := repository.session.Report
	if report.Score != 77 || report.Summary != "先做三家试点" || len(report.Metrics) != 1 || len(report.RoleSummaries) != 1 || len(report.Risks) != 1 || len(report.NextActions) != 1 {
		t.Fatalf("report = %+v", report)
	}
}
