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

func (s *Service) ListRoles() []Role {
	return DefaultRoles()
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

type Option func(*Service)

type Service struct {
	repository Repository
	generator  JSONGenerator
	quota      QuotaConsumer
	queue      Queue
	profile    ProfileContextProvider
	now        func() time.Time
}

func NewService(repository Repository, generator JSONGenerator, options ...Option) *Service {
	service := &Service{repository: repository, generator: generator, now: time.Now}
	for _, option := range options {
		option(service)
	}
	return service
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

func (s *Service) CreateSession(ctx context.Context, input CreateInput) (Session, error) {
	if s.repository == nil {
		return Session{}, ErrServiceNotReady
	}
	session := Session{
		UserID:      input.UserID,
		Goal:        strings.TrimSpace(input.Goal),
		TargetUsers: strings.TrimSpace(input.TargetUsers),
		Product:     strings.TrimSpace(input.Product),
		Roles:       normalizeRoles(input.Roles),
		Status:      StatusDraft,
		CurrentStep: StatusDraft,
		CreatedAt:   s.now(),
	}
	return s.repository.CreateSession(ctx, session)
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
	if goal == "" || targetUsers == "" || product == "" {
		return Session{}, ErrInvalidSession
	}
	return s.repository.UpdateSessionDraft(ctx, userID, id, DraftUpdate{
		Goal: &goal, TargetUsers: &targetUsers, Product: &product, Roles: &roles,
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
	if (session.Status != StatusDraft && session.Status != StatusFailed && session.Status != StatusCanceled) || session.Goal == "" || session.TargetUsers == "" || session.Product == "" || len(session.Roles) == 0 {
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
	return fmt.Sprintf(
		"目标：%s\n目标用户：%s\n产品/方案：%s\n推演角色：%s",
		session.Goal,
		session.TargetUsers,
		session.Product,
		strings.Join(session.Roles, "、"),
	)
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
