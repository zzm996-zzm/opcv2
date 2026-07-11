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
	"github.com/zzm/opcv2/internal/membership"
)

type Repository interface {
	CreateSession(ctx context.Context, session Session) (Session, error)
	UpdateSessionDraft(ctx context.Context, userID, id int64, update DraftUpdate) (Session, error)
	UpdateSessionResult(ctx context.Context, userID, id int64, result Report) (Session, error)
	ListSessions(ctx context.Context, userID int64, limit int) ([]Session, error)
	GetSession(ctx context.Context, userID, id int64) (Session, error)
}

func (s *Service) ListRoles() []Role {
	return DefaultRoles()
}

type JSONGenerator interface {
	GenerateJSON(ctx context.Context, request ai.GenerateJSONRequest) (ai.GenerateJSONResult, error)
}

type QuotaConsumer interface {
	CheckAndConsume(ctx context.Context, input membership.ConsumeInput) (membership.UsageItem, error)
}

type ProfileContextProvider interface {
	GetProfileContext(ctx context.Context, userID int64) (account.ProfileContext, error)
}

type Option func(*Service)

type Service struct {
	repository Repository
	generator  JSONGenerator
	quota      QuotaConsumer
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
	if s.repository == nil || s.generator == nil {
		return Session{}, ErrServiceNotReady
	}
	session, err := s.repository.GetSession(ctx, userID, id)
	if err != nil {
		return Session{}, err
	}
	if session.Status != StatusDraft || session.Goal == "" || session.TargetUsers == "" || session.Product == "" || len(session.Roles) == 0 {
		return Session{}, ErrInvalidSession
	}
	if s.quota != nil {
		if _, err := s.quota.CheckAndConsume(ctx, membership.ConsumeInput{
			UserID:         userID,
			FeatureKey:     membership.FeatureSandboxRuns,
			Amount:         1,
			IdempotencyKey: fmt.Sprintf("sandbox-run-%d", id),
		}); err != nil {
			return Session{}, err
		}
	}
	report, err := s.generateReport(ctx, session)
	if err != nil {
		return Session{}, err
	}
	return s.repository.UpdateSessionResult(ctx, userID, id, report)
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

func (s *Service) generateReport(ctx context.Context, session Session) (Report, error) {
	profilePrompt, err := s.profilePrompt(ctx, session.UserID)
	if err != nil {
		return Report{}, err
	}
	aiResult, err := s.generator.GenerateJSON(ctx, ai.GenerateJSONRequest{
		UserID:         session.UserID,
		Feature:        "sandbox.run",
		PromptVersion:  "sandbox_run_v1",
		SystemPrompt:   "你是商业沙盘推演助手。必须只返回 JSON，字段严格匹配 sandbox_report。",
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
	if err := validateReport(report); err != nil {
		return Report{}, fmt.Errorf("%w: %v", ErrInvalidAIResult, err)
	}
	return report, nil
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
	return validateReport(report)
}

func validateReport(report Report) error {
	if report.Score <= 0 ||
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
