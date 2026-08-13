package sandbox

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/zzm/opcv2/internal/account"
	"github.com/zzm/opcv2/internal/ai"
	"github.com/zzm/opcv2/internal/jobs"
	"github.com/zzm/opcv2/internal/membership"
)

type Repository interface {
	CreateSession(ctx context.Context, session Session) (Session, error)
	UpdateSessionDraft(ctx context.Context, userID, id int64, update DraftUpdate) (Session, error)
	UpdateSessionIntake(ctx context.Context, userID, id int64, intake Intake) (Session, error)
	UpdateSessionSettings(ctx context.Context, userID, id int64, settings RunSettings) (Session, error)
	PrepareSessionRun(ctx context.Context, userID, id int64) (Session, error)
	UpdateSessionProgress(ctx context.Context, userID, id int64, attempt int, status string, progress int, step, errorMessage string) (Session, error)
	CancelSession(ctx context.Context, userID, id int64) (Session, error)
	CreateMessage(ctx context.Context, message Message) (Message, error)
	ListMessages(ctx context.Context, userID, sessionID int64) ([]Message, error)
	UpdateSessionResult(ctx context.Context, userID, id int64, attempt int, result Report) (Session, error)
	ListSessions(ctx context.Context, userID int64, limit int) ([]Session, error)
	GetSession(ctx context.Context, userID, id int64) (Session, error)
}

type roleAnswer struct {
	Answer string `json:"answer"`
}

type intakeGeneration struct {
	Goal             string            `json:"goal"`
	TargetUsers      string            `json:"target_users"`
	Product          string            `json:"product"`
	RecognizedFields []RecognizedField `json:"recognized_fields"`
	Questions        []IntakeQuestion  `json:"questions"`
}

func (s *Service) ListRoles() []Role {
	roles := append([]Role{}, DefaultRoles()...)
	return append(roles, DefaultSystemPerspectives()...)
}

func (s *Service) Options() Options {
	return DefaultOptions()
}

func (s *Service) ListExamples() []Session {
	return ExampleSessions()
}

type JSONGenerator interface {
	GenerateJSON(ctx context.Context, request ai.GenerateJSONRequest) (ai.GenerateJSONResult, error)
}

type QuotaConsumer interface {
	CheckAndConsume(ctx context.Context, input membership.ConsumeInput) (membership.UsageItem, error)
	RefundUsage(ctx context.Context, input membership.ConsumeInput) (membership.UsageItem, error)
}

type Queue interface {
	Enqueue(ctx context.Context, job jobs.Job) error
}

type ProfileContextProvider interface {
	GetProfileContext(ctx context.Context, userID int64) (account.ProfileContext, error)
}

type V2PDFRenderer func(run V2SandboxRun) ([]byte, error)

type Option func(*Service)

type Service struct {
	repository  Repository
	generator   JSONGenerator
	quota       QuotaConsumer
	queue       Queue
	profile     ProfileContextProvider
	taskCreator SandboxTaskCreator
	renderPDF   V2PDFRenderer
	now         func() time.Time
}

func NewService(repository Repository, generator JSONGenerator, options ...Option) *Service {
	service := &Service{repository: repository, generator: generator, renderPDF: RenderV2SandboxPDF, now: time.Now}
	for _, option := range options {
		option(service)
	}
	return service
}

func WithV2PDFRenderer(renderer V2PDFRenderer) Option {
	return func(service *Service) {
		service.renderPDF = renderer
	}
}

func WithQuotaConsumer(quota QuotaConsumer) Option {
	return func(service *Service) {
		service.quota = quota
	}
}

func WithProfileContextProvider(provider ProfileContextProvider) Option {
	return func(service *Service) {
		service.profile = provider
	}
}

func WithQueue(queue Queue) Option {
	return func(service *Service) {
		service.queue = queue
	}
}

func WithTaskCreator(creator SandboxTaskCreator) Option {
	return func(service *Service) { service.taskCreator = creator }
}

func (s *Service) CreateSession(ctx context.Context, input CreateInput) (Session, error) {
	if s.repository == nil {
		return Session{}, ErrServiceNotReady
	}
	now := s.now()
	session := Session{
		UserID:      input.UserID,
		Goal:        strings.TrimSpace(input.Goal),
		TargetUsers: strings.TrimSpace(input.TargetUsers),
		Product:     strings.TrimSpace(input.Product),
		Roles:       normalizeRoles(input.Roles),
		Status:      StatusDraft,
		CurrentStep: StatusDraft,
		Intake:      DefaultReadyIntake(),
		Settings:    DefaultRunSettings(),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	return s.repository.CreateSession(ctx, session)
}

func (s *Service) CreateIntake(ctx context.Context, input IntakeCreateInput) (Session, error) {
	if s.repository == nil || s.generator == nil {
		return Session{}, ErrServiceNotReady
	}
	input.InitialIdea = strings.TrimSpace(input.InitialIdea)
	if input.UserID <= 0 || input.InitialIdea == "" || len([]rune(input.InitialIdea)) > 5000 {
		return Session{}, ErrInvalidIntake
	}
	result, err := s.generator.GenerateJSON(ctx, ai.GenerateJSONRequest{
		UserID:         input.UserID,
		Feature:        "sandbox.intake",
		PromptVersion:  "sandbox_intake_v1",
		SystemPrompt:   "你是商业沙盘信息整理助手。只返回 JSON，提取目标、用户、产品，并生成结构化追问。不得编造用户未提供的事实；未知字段用空字符串表示。",
		UserPrompt:     "请根据以下项目描述生成商业沙盘初始信息和追问：\n" + input.InitialIdea,
		SchemaName:     "sandbox_intake",
		Validate:       validateIntakeJSON,
		RepairAttempts: 1,
	})
	if err != nil {
		return Session{}, fmt.Errorf("%w: %v", ErrInvalidAIResult, err)
	}
	var generated intakeGeneration
	if err := json.Unmarshal(result.Content, &generated); err != nil {
		return Session{}, fmt.Errorf("%w: %v", ErrInvalidAIResult, err)
	}
	generated.Goal = strings.TrimSpace(generated.Goal)
	generated.TargetUsers = strings.TrimSpace(generated.TargetUsers)
	generated.Product = strings.TrimSpace(generated.Product)
	if generated.Goal == "" || generated.TargetUsers == "" || generated.Product == "" || len(generated.Questions) == 0 {
		return Session{}, ErrInvalidAIResult
	}
	questions := normalizeIntakeQuestions(generated.Questions)
	if len(questions) == 0 {
		return Session{}, ErrInvalidAIResult
	}
	intake := Intake{
		Status:           IntakeStatusQuestions,
		InitialIdea:      input.InitialIdea,
		RecognizedFields: normalizeRecognizedFields(generated.RecognizedFields),
		Questions:        questions,
		TotalQuestions:   len(questions),
	}
	now := s.now()
	return s.repository.CreateSession(ctx, Session{
		UserID:      input.UserID,
		Goal:        generated.Goal,
		TargetUsers: generated.TargetUsers,
		Product:     generated.Product,
		Status:      StatusDraft,
		CurrentStep: IntakeStatusQuestions,
		Intake:      intake,
		Settings:    DefaultRunSettings(),
		CreatedAt:   now,
		UpdatedAt:   now,
	})
}

func (s *Service) AnswerIntake(ctx context.Context, input IntakeAnswerInput) (Session, error) {
	if s.repository == nil {
		return Session{}, ErrServiceNotReady
	}
	input.QuestionKey = strings.TrimSpace(input.QuestionKey)
	input.Answer = strings.TrimSpace(input.Answer)
	if input.UserID <= 0 || input.SessionID <= 0 || input.QuestionKey == "" || (input.Answer == "" && !input.Skipped) {
		return Session{}, ErrInvalidIntake
	}
	session, err := s.repository.GetSession(ctx, input.UserID, input.SessionID)
	if err != nil {
		return Session{}, err
	}
	if session.Status != StatusDraft || session.Intake.Status == IntakeStatusReady {
		return Session{}, ErrInvalidIntake
	}
	intake := normalizeIntake(session.Intake)
	questionIndex := -1
	for index := range intake.Questions {
		if intake.Questions[index].Key == input.QuestionKey {
			questionIndex = index
			break
		}
	}
	if questionIndex < 0 {
		return Session{}, ErrInvalidIntake
	}
	question := &intake.Questions[questionIndex]
	if question.MaxLength > 0 && len([]rune(input.Answer)) > question.MaxLength {
		return Session{}, ErrInvalidIntake
	}
	question.Answer = input.Answer
	question.Skipped = input.Skipped
	if question.Answer != "" {
		question.Skipped = false
	}
	intake.AnsweredCount = countCompletedQuestions(intake.Questions)
	return s.repository.UpdateSessionIntake(ctx, input.UserID, input.SessionID, intake)
}

func (s *Service) CompleteIntake(ctx context.Context, userID, id int64) (Session, error) {
	if s.repository == nil {
		return Session{}, ErrServiceNotReady
	}
	session, err := s.repository.GetSession(ctx, userID, id)
	if err != nil {
		return Session{}, err
	}
	if session.Status != StatusDraft {
		return Session{}, ErrInvalidIntake
	}
	intake := normalizeIntake(session.Intake)
	if len(intake.Questions) == 0 {
		return Session{}, ErrIntakeIncomplete
	}
	for _, question := range intake.Questions {
		if strings.TrimSpace(question.Answer) == "" && !question.Skipped {
			return Session{}, ErrIntakeIncomplete
		}
	}
	intake.Status = IntakeStatusReady
	intake.AnsweredCount = countCompletedQuestions(intake.Questions)
	return s.repository.UpdateSessionIntake(ctx, userID, id, intake)
}

func (s *Service) UpdateSessionSettings(ctx context.Context, userID, id int64, settings RunSettings) (Session, error) {
	if s.repository == nil {
		return Session{}, ErrServiceNotReady
	}
	settings, err := normalizeRunSettings(settings)
	if err != nil {
		return Session{}, err
	}
	return s.repository.UpdateSessionSettings(ctx, userID, id, settings)
}

func (s *Service) UpdateSessionDraft(ctx context.Context, userID, id int64, update DraftUpdate) (Session, error) {
	if s.repository == nil {
		return Session{}, ErrServiceNotReady
	}
	session, err := s.repository.GetSession(ctx, userID, id)
	if err != nil {
		return Session{}, err
	}
	if session.Status != StatusDraft {
		return Session{}, ErrInvalidSession
	}
	hasDraftFields := update.Goal != nil || update.TargetUsers != nil || update.Product != nil || update.Roles != nil
	goal, targetUsers, product, roles := session.Goal, session.TargetUsers, session.Product, session.Roles
	if update.Goal != nil {
		goal = strings.TrimSpace(*update.Goal)
	}
	if update.TargetUsers != nil {
		targetUsers = strings.TrimSpace(*update.TargetUsers)
	}
	if update.Product != nil {
		product = strings.TrimSpace(*update.Product)
	}
	if update.Roles != nil {
		roles = normalizeRoles(*update.Roles)
	}
	if hasDraftFields && (goal == "" || targetUsers == "" || product == "") {
		return Session{}, ErrInvalidSession
	}
	var settings *RunSettings
	if update.Settings != nil {
		normalized, err := normalizeRunSettings(*update.Settings)
		if err != nil {
			return Session{}, err
		}
		settings = &normalized
	}
	if !hasDraftFields {
		if settings == nil {
			return Session{}, ErrInvalidSession
		}
		return s.repository.UpdateSessionSettings(ctx, userID, id, *settings)
	}
	return s.repository.UpdateSessionDraft(ctx, userID, id, DraftUpdate{
		Goal: &goal, TargetUsers: &targetUsers, Product: &product, Roles: &roles, Settings: settings,
	})
}

func (s *Service) RunSession(ctx context.Context, userID, id int64) (Session, error) {
	if s.repository == nil || s.queue == nil {
		return Session{}, ErrServiceNotReady
	}
	session, err := s.repository.GetSession(ctx, userID, id)
	if err != nil {
		return Session{}, err
	}
	intakeIncomplete := len(session.Intake.Questions) > 0 && session.Intake.Status != IntakeStatusReady
	if (session.Status != StatusDraft && session.Status != StatusFailed && session.Status != StatusCanceled) || session.Goal == "" || session.TargetUsers == "" || session.Product == "" || len(session.Roles) == 0 || intakeIncomplete {
		return Session{}, ErrInvalidSession
	}
	queued, err := s.repository.PrepareSessionRun(ctx, userID, id)
	if err != nil {
		return Session{}, err
	}
	if s.quota != nil {
		if _, err := s.quota.CheckAndConsume(ctx, membership.ConsumeInput{
			UserID:         userID,
			FeatureKey:     membership.FeatureSandboxRuns,
			Amount:         1,
			IdempotencyKey: sandboxConsumeKey(id, queued.RunAttempt),
		}); err != nil {
			_, _ = s.repository.UpdateSessionProgress(ctx, userID, id, queued.RunAttempt, StatusFailed, 100, "failed", "quota_check_failed")
			return Session{}, err
		}
	}
	if err := s.queue.Enqueue(ctx, jobs.Job{
		Type:           jobs.TypeSandboxRun,
		IdempotencyKey: fmt.Sprintf("sandbox-run-%d-attempt-%d", id, queued.RunAttempt),
		Payload:        map[string]any{"user_id": userID, "session_id": id, "attempt": queued.RunAttempt},
		MaxRetry:       1,
		Timeout:        5 * time.Minute,
	}); err != nil {
		_, _ = s.repository.UpdateSessionProgress(ctx, userID, id, queued.RunAttempt, StatusFailed, 100, "failed", "queue_unavailable")
		if refundErr := s.refundRun(ctx, userID, id, queued.RunAttempt); refundErr != nil {
			return Session{}, refundErr
		}
		return Session{}, err
	}
	return queued, nil
}

func (s *Service) RetrySession(ctx context.Context, userID, id int64) (Session, error) {
	return s.RunSession(ctx, userID, id)
}

func (s *Service) CancelSession(ctx context.Context, userID, id int64) (Session, error) {
	if s.repository == nil {
		return Session{}, ErrServiceNotReady
	}
	session, err := s.repository.CancelSession(ctx, userID, id)
	if err != nil {
		return Session{}, err
	}
	if err := s.refundRun(ctx, userID, id, session.RunAttempt); err != nil {
		return Session{}, err
	}
	return session, nil
}

func (s *Service) ProcessSession(ctx context.Context, userID, id int64, attempt int) error {
	if s.repository == nil || s.generator == nil {
		return ErrServiceNotReady
	}
	session, err := s.repository.GetSession(ctx, userID, id)
	if err != nil {
		return err
	}
	if session.RunAttempt != attempt || session.Status == StatusCanceled || session.Status == StatusCompleted {
		return nil
	}
	if session.Status != StatusQueued {
		return ErrStaleRun
	}
	if _, err := s.repository.UpdateSessionProgress(ctx, userID, id, attempt, StatusRunning, 20, "generating_report", ""); err != nil {
		return err
	}
	report, err := s.generateReport(ctx, session)
	if err != nil {
		_, _ = s.repository.UpdateSessionProgress(ctx, userID, id, attempt, StatusFailed, 100, "failed", "invalid_ai_result")
		if refundErr := s.refundRun(ctx, userID, id, attempt); refundErr != nil {
			return refundErr
		}
		return err
	}
	if _, err := s.repository.UpdateSessionProgress(ctx, userID, id, attempt, StatusRunning, 80, "storing_report", ""); err != nil {
		if errors.Is(err, ErrStaleRun) {
			return nil
		}
		return err
	}
	_, err = s.repository.UpdateSessionResult(ctx, userID, id, attempt, report)
	if errors.Is(err, ErrStaleRun) {
		return nil
	}
	return err
}

func (s *Service) refundRun(ctx context.Context, userID, id int64, attempt int) error {
	if s.quota == nil {
		return nil
	}
	_, err := s.quota.RefundUsage(ctx, membership.ConsumeInput{
		UserID: userID, FeatureKey: membership.FeatureSandboxRuns, Amount: 1,
		IdempotencyKey: fmt.Sprintf("sandbox-run-%d-attempt-%d-refund", id, attempt),
	})
	return err
}

func sandboxConsumeKey(id int64, attempt int) string {
	return fmt.Sprintf("sandbox-run-%d-attempt-%d", id, attempt)
}

func (s *Service) ListSessions(ctx context.Context, userID int64, limit int) ([]Session, error) {
	if s.repository == nil {
		return nil, ErrServiceNotReady
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return s.repository.ListSessions(ctx, userID, limit)
}

func (s *Service) GetSession(ctx context.Context, userID, id int64) (Session, error) {
	if s.repository == nil {
		return Session{}, ErrServiceNotReady
	}
	return s.repository.GetSession(ctx, userID, id)
}

func (s *Service) ListMessages(ctx context.Context, userID, sessionID int64) ([]Message, error) {
	if s.repository == nil {
		return nil, ErrServiceNotReady
	}
	if _, err := s.repository.GetSession(ctx, userID, sessionID); err != nil {
		return nil, err
	}
	return s.repository.ListMessages(ctx, userID, sessionID)
}

func (s *Service) AskRole(ctx context.Context, input AskRoleInput) (Message, error) {
	if s.repository == nil || s.generator == nil {
		return Message{}, ErrServiceNotReady
	}
	input.Role = strings.TrimSpace(input.Role)
	input.Question = strings.TrimSpace(input.Question)
	if input.UserID <= 0 || input.SessionID <= 0 || input.Role == "" || input.Question == "" || len([]rune(input.Question)) > 2000 {
		return Message{}, ErrInvalidSession
	}
	session, err := s.repository.GetSession(ctx, input.UserID, input.SessionID)
	if err != nil {
		return Message{}, err
	}
	if !containsRole(session.Roles, input.Role) {
		return Message{}, ErrInvalidSession
	}
	if session.Status != StatusCompleted {
		return Message{}, ErrInvalidSession
	}
	result, err := s.generator.GenerateJSON(ctx, ai.GenerateJSONRequest{
		UserID: input.UserID, Feature: "sandbox.follow_up", PromptVersion: "sandbox_role_follow_up_v1",
		SystemPrompt: "你正在商业沙盘中扮演指定角色。只返回JSON，格式为{\"answer\":\"...\"}。",
		UserPrompt:   fmt.Sprintf("产品：%s\n角色：%s\n问题：%s", session.Product, input.Role, input.Question),
		SchemaName:   "sandbox_role_answer", Validate: validateRoleAnswerJSON, RepairAttempts: 1,
	})
	if err != nil {
		return Message{}, fmt.Errorf("%w: %v", ErrInvalidAIResult, err)
	}
	var answer roleAnswer
	if err := json.Unmarshal(result.Content, &answer); err != nil || strings.TrimSpace(answer.Answer) == "" {
		return Message{}, ErrInvalidAIResult
	}
	return s.repository.CreateMessage(ctx, Message{
		SessionID: input.SessionID, UserID: input.UserID, Role: input.Role,
		Question: input.Question, Answer: strings.TrimSpace(answer.Answer), CreatedAt: s.now(),
	})
}

func validateRoleAnswerJSON(data []byte) error {
	var answer roleAnswer
	if err := json.Unmarshal(data, &answer); err != nil {
		return err
	}
	if strings.TrimSpace(answer.Answer) == "" {
		return errors.New("sandbox role answer is empty")
	}
	return nil
}

func containsRole(roles []string, role string) bool {
	for _, candidate := range roles {
		if candidate == role {
			return true
		}
	}
	return false
}

func (s *Service) generateReport(ctx context.Context, session Session) (Report, error) {
	profilePrompt, err := s.profilePrompt(ctx, session.UserID)
	if err != nil {
		return Report{}, err
	}
	aiResult, err := s.generator.GenerateJSON(ctx, ai.GenerateJSONRequest{
		UserID:         session.UserID,
		Feature:        "sandbox.run",
		PromptVersion:  "sandbox_run_v2",
		SystemPrompt:   "你是商业沙盘推演助手。必须只返回 JSON，字段严格匹配 sandbox_report。所有分数和判断均为模型推演；assumptions 必须列出关键假设；不得编造外部证据或来源链接。",
		UserPrompt:     appendPromptSection(sandboxUserPrompt(session), profilePrompt),
		SchemaName:     "sandbox_report",
		Validate:       validateReportJSON,
		RepairAttempts: 1,
	})
	if err != nil {
		return Report{}, fmt.Errorf("%w: %v", ErrInvalidAIResult, err)
	}
	var report Report
	if err := json.Unmarshal(aiResult.Content, &report); err != nil {
		return Report{}, fmt.Errorf("%w: %v", ErrInvalidAIResult, err)
	}
	report = normalizeReportContent(report)
	report = labelModelReport(report, session)
	if err := validateReport(report); err != nil {
		return Report{}, fmt.Errorf("%w: %v", ErrInvalidAIResult, err)
	}
	return report, nil
}

func labelModelReport(report Report, session Session) Report {
	report.Basis = "model_simulation"
	report.Disclaimer = "本报告由 AI 基于用户输入进行情景推演，不代表真实市场统计、收益承诺或已验证事实。"
	if len(report.Assumptions) == 0 {
		report.Assumptions = []string{
			fmt.Sprintf("目标用户范围以“%s”为前提。", session.TargetUsers),
			fmt.Sprintf("产品方案以“%s”的当前描述为前提。", session.Product),
			"当前推演未接入外部市场数据或真实用户实验结果。",
		}
	}
	report.EvidenceSources = []ReportEvidence{}
	return report
}

func (s *Service) profilePrompt(ctx context.Context, userID int64) (string, error) {
	if s.profile == nil {
		return "", nil
	}
	profile, err := s.profile.GetProfileContext(ctx, userID)
	if err != nil {
		return "", err
	}
	return formatProfileContext(profile), nil
}

func sandboxUserPrompt(session Session) string {
	prompt := fmt.Sprintf(
		"目标：%s\n目标用户：%s\n产品/方案：%s\n推演角色：%s",
		session.Goal,
		session.TargetUsers,
		session.Product,
		strings.Join(session.Roles, "、"),
	)
	settings, err := normalizeRunSettings(session.Settings)
	if err != nil {
		settings = DefaultRunSettings()
	}
	prompt += fmt.Sprintf("\n推演深度：%s\n输出风格：%s\n生成大纲：%t", settings.Depth, settings.OutputStyle, settings.GenerateOutline)
	if len(settings.Variables) > 0 {
		keys := make([]string, 0, len(settings.Variables))
		for key := range settings.Variables {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			prompt += fmt.Sprintf("\n高级变量[%s]：%s", key, settings.Variables[key])
		}
	}
	for _, question := range session.Intake.Questions {
		if strings.TrimSpace(question.Answer) != "" {
			prompt += fmt.Sprintf("\n补充回答[%s]：%s", question.Title, question.Answer)
		}
	}
	return prompt
}

func appendPromptSection(base string, section string) string {
	section = strings.TrimSpace(section)
	if section == "" {
		return base
	}
	return base + "\n\n" + section
}

func formatProfileContext(profile account.ProfileContext) string {
	if len(profile.Groups) == 0 {
		return ""
	}
	lines := []string{"用户画像上下文："}
	for _, group := range profile.Groups {
		if len(group.Fields) == 0 {
			continue
		}
		fields := make([]string, 0, len(group.Fields))
		for key, value := range group.Fields {
			value = strings.TrimSpace(value)
			if value == "" {
				continue
			}
			fields = append(fields, key+"="+value)
		}
		if len(fields) == 0 {
			continue
		}
		sort.Strings(fields)
		lines = append(lines, "- "+group.Title+"："+strings.Join(fields, "；"))
	}
	if len(lines) == 1 {
		return ""
	}
	return strings.Join(lines, "\n")
}

func validateReportJSON(data []byte) error {
	var report Report
	if err := json.Unmarshal(data, &report); err != nil {
		return err
	}
	report = normalizeReportContent(report)
	return validateReportContent(report)
}

func validateReport(report Report) error {
	if err := validateReportContent(report); err != nil {
		return err
	}
	if report.Basis != "model_simulation" || report.Disclaimer == "" || len(report.Assumptions) == 0 {
		return errors.New("sandbox report metadata is missing")
	}
	return nil
}

func validateReportContent(report Report) error {
	if report.Score <= 0 || report.Score > 100 ||
		report.Summary == "" ||
		len(report.Metrics) == 0 ||
		len(report.RoleSummaries) == 0 ||
		len(report.Risks) == 0 ||
		len(report.NextActions) == 0 {
		return errors.New("sandbox report is missing required fields")
	}
	for _, metric := range report.Metrics {
		if metric.Label == "" || metric.Value == "" {
			return errors.New("sandbox metric is missing required fields")
		}
	}
	for _, summary := range report.RoleSummaries {
		if summary.Role == "" || summary.View == "" {
			return errors.New("sandbox role summary is missing required fields")
		}
	}
	return nil
}

func validateIntakeJSON(data []byte) error {
	var generated intakeGeneration
	if err := json.Unmarshal(data, &generated); err != nil {
		return err
	}
	if strings.TrimSpace(generated.Goal) == "" || strings.TrimSpace(generated.TargetUsers) == "" || strings.TrimSpace(generated.Product) == "" {
		return errors.New("sandbox intake is missing core fields")
	}
	if len(normalizeIntakeQuestions(generated.Questions)) == 0 {
		return errors.New("sandbox intake has no questions")
	}
	return nil
}

func normalizeIntakeQuestions(questions []IntakeQuestion) []IntakeQuestion {
	normalized := make([]IntakeQuestion, 0, len(questions))
	seen := make(map[string]struct{}, len(questions))
	for _, question := range questions {
		question.Key = strings.TrimSpace(question.Key)
		question.Title = strings.TrimSpace(question.Title)
		question.Hint = strings.TrimSpace(question.Hint)
		question.Placeholder = strings.TrimSpace(question.Placeholder)
		if question.Key == "" || question.Title == "" {
			continue
		}
		if _, exists := seen[question.Key]; exists {
			continue
		}
		seen[question.Key] = struct{}{}
		if question.MaxLength <= 0 || question.MaxLength > 2000 {
			question.MaxLength = 1000
		}
		question.Position = len(normalized) + 1
		question.Answer = strings.TrimSpace(question.Answer)
		question.Skipped = question.Skipped && question.Answer == ""
		normalized = append(normalized, question)
	}
	return normalized
}

func normalizeRecognizedFields(fields []RecognizedField) []RecognizedField {
	normalized := make([]RecognizedField, 0, len(fields))
	seen := make(map[string]struct{}, len(fields))
	for _, field := range fields {
		field.Key = strings.TrimSpace(field.Key)
		field.Label = strings.TrimSpace(field.Label)
		field.Value = strings.TrimSpace(field.Value)
		if field.Key == "" || field.Label == "" || field.Value == "" {
			continue
		}
		if _, exists := seen[field.Key]; exists {
			continue
		}
		seen[field.Key] = struct{}{}
		normalized = append(normalized, field)
	}
	return normalized
}

func normalizeIntake(intake Intake) Intake {
	intake.InitialIdea = strings.TrimSpace(intake.InitialIdea)
	intake.RecognizedFields = normalizeRecognizedFields(intake.RecognizedFields)
	intake.Questions = normalizeIntakeQuestions(intake.Questions)
	intake.TotalQuestions = len(intake.Questions)
	intake.AnsweredCount = countCompletedQuestions(intake.Questions)
	return intake
}

func countCompletedQuestions(questions []IntakeQuestion) int {
	count := 0
	for _, question := range questions {
		if strings.TrimSpace(question.Answer) != "" || question.Skipped {
			count++
		}
	}
	return count
}

func normalizeRunSettings(settings RunSettings) (RunSettings, error) {
	defaults := DefaultRunSettings()
	if settings.Depth == "" {
		settings.Depth = defaults.Depth
	}
	if settings.OutputStyle == "" {
		settings.OutputStyle = defaults.OutputStyle
	}
	if settings.Depth != RunDepthStandard && settings.Depth != RunDepthDeep {
		return RunSettings{}, ErrInvalidSession
	}
	if settings.OutputStyle != OutputStyleStructured && settings.OutputStyle != OutputStyleConcise {
		return RunSettings{}, ErrInvalidSession
	}
	if settings.Variables == nil {
		settings.Variables = map[string]string{}
	}
	if len(settings.Variables) > 50 {
		return RunSettings{}, ErrInvalidSession
	}
	variables := make(map[string]string, len(settings.Variables))
	for key, value := range settings.Variables {
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key == "" || len([]rune(key)) > 100 || len([]rune(value)) > 1000 {
			return RunSettings{}, ErrInvalidSession
		}
		variables[key] = value
	}
	settings.Variables = variables
	return settings, nil
}

func normalizeReportContent(report Report) Report {
	if report.Score == 0 && report.ConsumerProbability > 0 {
		report.Score = report.ConsumerProbability
	}
	if strings.TrimSpace(report.Summary) == "" && len(report.CoreConclusions) > 0 {
		report.Summary = strings.TrimSpace(report.CoreConclusions[0])
	}
	if len(report.Metrics) == 0 && len(report.ValidationMetrics) > 0 {
		report.Metrics = make([]Metric, 0, len(report.ValidationMetrics))
		for _, metric := range report.ValidationMetrics {
			value := strings.TrimSpace(metric.Target)
			if value == "" {
				value = strings.TrimSpace(metric.Current)
			}
			report.Metrics = append(report.Metrics, Metric{Label: strings.TrimSpace(metric.Label), Value: value})
		}
	}
	if len(report.RoleSummaries) == 0 {
		for _, insight := range report.OpportunityAnalysis {
			if strings.TrimSpace(insight.Detail) == "" {
				continue
			}
			report.RoleSummaries = append(report.RoleSummaries, RoleSummary{Role: "综合分析", View: strings.TrimSpace(insight.Detail)})
			break
		}
	}
	if len(report.Risks) == 0 {
		for _, insight := range report.RiskAnalysis {
			if strings.TrimSpace(insight.Detail) != "" {
				report.Risks = append(report.Risks, strings.TrimSpace(insight.Detail))
			}
		}
	}
	if len(report.NextActions) == 0 {
		for _, action := range report.ActionPlan {
			if strings.TrimSpace(action.Title) == "" && strings.TrimSpace(action.Detail) == "" {
				continue
			}
			title := strings.TrimSpace(action.Title)
			if title == "" {
				title = strings.TrimSpace(action.Detail)
			}
			report.NextActions = append(report.NextActions, title)
		}
	}
	return report
}

func normalizeRoles(roles []string) []string {
	normalized := make([]string, 0, len(roles))
	seen := make(map[string]struct{}, len(roles))
	for _, role := range roles {
		role = strings.TrimSpace(role)
		if role == "" {
			continue
		}
		if _, exists := seen[role]; exists {
			continue
		}
		seen[role] = struct{}{}
		normalized = append(normalized, role)
	}
	return normalized
}
