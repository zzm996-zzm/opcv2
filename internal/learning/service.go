package learning

import (
	"context"
	"strings"
	"time"
)

type Repository interface {
	ListCourses(ctx context.Context, filter CourseFilter) ([]Course, error)
	GetCourse(ctx context.Context, slug string) (Course, error)
	ListProgress(ctx context.Context, userID int64) ([]Progress, error)
	CreateDiagnosis(ctx context.Context, diagnosis Diagnosis) (Diagnosis, error)
	LatestDiagnosis(ctx context.Context, userID int64) (Diagnosis, error)
}

type Service struct {
	repository Repository
	now        func() time.Time
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository, now: time.Now}
}

func (s *Service) ListCourses(ctx context.Context, filter CourseFilter) ([]Course, error) {
	if s.repository == nil {
		return nil, ErrServiceNotReady
	}
	if filter.Limit <= 0 || filter.Limit > 100 {
		filter.Limit = 20
	}
	filter.Category = strings.TrimSpace(filter.Category)
	return s.repository.ListCourses(ctx, filter)
}

func (s *Service) GetCourse(ctx context.Context, slug string) (Course, error) {
	if s.repository == nil {
		return Course{}, ErrServiceNotReady
	}
	return s.repository.GetCourse(ctx, strings.TrimSpace(slug))
}

func (s *Service) ListProgress(ctx context.Context, userID int64) ([]Progress, error) {
	if s.repository == nil {
		return nil, ErrServiceNotReady
	}
	return s.repository.ListProgress(ctx, userID)
}

func (s *Service) CreateDiagnosis(ctx context.Context, input CreateDiagnosisInput) (Diagnosis, error) {
	if s.repository == nil {
		return Diagnosis{}, ErrServiceNotReady
	}
	now := s.now()
	goal := strings.TrimSpace(input.Goal)
	project := strings.TrimSpace(input.Project)
	return s.repository.CreateDiagnosis(ctx, Diagnosis{
		UserID:       input.UserID,
		Goal:         goal,
		Project:      project,
		Status:       DiagnosisCompleted,
		OverallScore: 72,
		Dimensions: []Dimension{
			{Name: "市场分析能力", Score: 78, Gap: 12, Summary: "具备基础判断能力，需要补充竞品拆解方法。"},
			{Name: "数据分析能力", Score: 64, Gap: 22, Summary: "能理解核心指标，但需要加强漏斗和转化分析。"},
			{Name: "运营执行能力", Score: 70, Gap: 18, Summary: "执行意识较强，建议补齐任务拆解和复盘机制。"},
		},
		Recommendations: []string{
			"优先学习 AI行业分析方法，建立市场与竞品拆解框架。",
			"补充提示词工程实战，提升 AI 辅助分析和内容生成质量。",
			"把学习任务同步到任务中心，每周复盘一次应用效果。",
		},
		CreatedAt: now,
	})
}

func (s *Service) LatestDiagnosis(ctx context.Context, userID int64) (Diagnosis, error) {
	if s.repository == nil {
		return Diagnosis{}, ErrServiceNotReady
	}
	return s.repository.LatestDiagnosis(ctx, userID)
}
