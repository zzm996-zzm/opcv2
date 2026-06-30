package sandbox

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/zzm/opcv2/internal/ai"
)

type Repository interface {
	CreateSession(ctx context.Context, session Session) (Session, error)
	UpdateSessionResult(ctx context.Context, userID, id int64, result Report) (Session, error)
	ListSessions(ctx context.Context, userID int64, limit int) ([]Session, error)
	GetSession(ctx context.Context, userID, id int64) (Session, error)
}

type JSONGenerator interface {
	GenerateJSON(ctx context.Context, request ai.GenerateJSONRequest) (ai.GenerateJSONResult, error)
}

type Service struct {
	repository Repository
	generator  JSONGenerator
	now        func() time.Time
}

func NewService(repository Repository, generator JSONGenerator) *Service {
	return &Service{repository: repository, generator: generator, now: time.Now}
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

func (s *Service) RunSession(ctx context.Context, userID, id int64) (Session, error) {
	if s.repository == nil || s.generator == nil {
		return Session{}, ErrServiceNotReady
	}
	session, err := s.repository.GetSession(ctx, userID, id)
	if err != nil {
		return Session{}, err
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
	aiResult, err := s.generator.GenerateJSON(ctx, ai.GenerateJSONRequest{
		UserID:         session.UserID,
		Feature:        "sandbox.run",
		PromptVersion:  "sandbox_run_v1",
		SystemPrompt:   "你是商业沙盘推演助手。必须只返回 JSON，字段严格匹配 sandbox_report。",
		UserPrompt:     sandboxUserPrompt(session),
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

func sandboxUserPrompt(session Session) string {
	return fmt.Sprintf(
		"目标：%s\n目标用户：%s\n产品/方案：%s\n推演角色：%s",
		session.Goal,
		session.TargetUsers,
		session.Product,
		strings.Join(session.Roles, "、"),
	)
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
	for _, role := range roles {
		role = strings.TrimSpace(role)
		if role != "" {
			normalized = append(normalized, role)
		}
	}
	return normalized
}
