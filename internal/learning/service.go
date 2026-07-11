package learning

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"
)

type Repository interface {
	ListCourses(ctx context.Context, filter CourseFilter) ([]Course, error)
	GetCourse(ctx context.Context, slug string) (Course, error)
	ListProgress(ctx context.Context, userID int64) ([]Progress, error)
	GetProgress(ctx context.Context, userID int64, courseSlug string) (Progress, error)
	UpsertProgress(ctx context.Context, progress Progress) (Progress, error)
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

func (s *Service) GetProgress(ctx context.Context, userID int64, courseSlug string) (Progress, error) {
	if s.repository == nil {
		return Progress{}, ErrServiceNotReady
	}
	courseSlug = strings.TrimSpace(courseSlug)
	if userID <= 0 || courseSlug == "" {
		return Progress{}, ErrInvalidProgress
	}
	return s.repository.GetProgress(ctx, userID, courseSlug)
}

func (s *Service) UpdateProgress(ctx context.Context, input UpdateProgressInput) (Progress, error) {
	if s.repository == nil {
		return Progress{}, ErrServiceNotReady
	}
	input.CourseSlug = strings.TrimSpace(input.CourseSlug)
	if input.UserID <= 0 || input.CourseSlug == "" || input.Percent < 0 || input.Percent > 100 {
		return Progress{}, ErrInvalidProgress
	}
	course, err := s.repository.GetCourse(ctx, input.CourseSlug)
	if err != nil {
		return Progress{}, err
	}
	progress, err := s.repository.UpsertProgress(ctx, Progress{
		UserID:            input.UserID,
		CourseSlug:        input.CourseSlug,
		Percent:           input.Percent,
		LastLesson:        strings.TrimSpace(input.LastLesson),
		RecommendedAction: strings.TrimSpace(input.RecommendedAction),
		UpdatedAt:         s.now(),
	})
	if err != nil {
		return Progress{}, err
	}
	progress.CourseTitle = course.Title
	return progress, nil
}

func (s *Service) CreateDiagnosis(ctx context.Context, input CreateDiagnosisInput) (Diagnosis, error) {
	if s.repository == nil {
		return Diagnosis{}, ErrServiceNotReady
	}
	now := s.now()
	goal := strings.TrimSpace(input.Goal)
	project := strings.TrimSpace(input.Project)
	return s.repository.CreateDiagnosis(ctx, Diagnosis{
		UserID:         input.UserID,
		Goal:           goal,
		Project:        project,
		FocusAbilities: normalizeUniqueStrings(input.FocusAbilities),
		WeeklyTime:     strings.TrimSpace(input.WeeklyTime),
		Bottleneck:     strings.TrimSpace(input.Bottleneck),
		Status:         DiagnosisCompleted,
		OverallScore:   72,
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

func normalizeUniqueStrings(values []string) []string {
	result := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func (s *Service) LatestDiagnosis(ctx context.Context, userID int64) (Diagnosis, error) {
	if s.repository == nil {
		return Diagnosis{}, ErrServiceNotReady
	}
	return s.repository.LatestDiagnosis(ctx, userID)
}

func (s *Service) LatestGaps(ctx context.Context, userID int64) (DiagnosisGaps, error) {
	diagnosis, err := s.LatestDiagnosis(ctx, userID)
	if err != nil {
		return DiagnosisGaps{}, err
	}
	return deriveGaps(diagnosis), nil
}

func (s *Service) LatestRecommendations(ctx context.Context, userID int64) (DiagnosisRecommendations, error) {
	diagnosis, err := s.LatestDiagnosis(ctx, userID)
	if err != nil {
		return DiagnosisRecommendations{}, err
	}
	gaps := prioritizedGaps(diagnosis)
	focus := make([]RecommendationFocus, 0, len(gaps))
	for _, gap := range gaps {
		focus = append(focus, RecommendationFocus{
			Name:     gap.Name,
			Priority: gap.Priority,
			Summary:  gap.Recommended,
		})
	}
	return DiagnosisRecommendations{
		DiagnosisID:     diagnosis.ID,
		Goal:            diagnosis.Goal,
		Project:         diagnosis.Project,
		Focus:           focus,
		Recommendations: diagnosis.Recommendations,
		Methods: []LearningMethod{
			{Title: "建议每周学习节奏", Value: "每周 6-8 小时", Detail: "建议每周学习 2-3 次，并保留复盘时间。"},
			{Title: "预计完成周期", Value: "3-4 周", Detail: "约 24-32 小时学习量。"},
			{Title: "建议学习顺序", Value: learningOrder(gaps), Detail: "先补关键短板，再巩固通用能力。"},
			{Title: "学习目标产出", Value: "3 个能力交付物", Detail: "完成实战练习、项目拆解和复盘报告。"},
		},
		GeneratedAt: diagnosis.UpdatedAt,
	}, nil
}

func (s *Service) LatestPlan(ctx context.Context, userID int64) (DiagnosisPlan, error) {
	diagnosis, err := s.LatestDiagnosis(ctx, userID)
	if err != nil {
		return DiagnosisPlan{}, err
	}
	gaps := prioritizedGaps(diagnosis)
	stages := make([]PlanStage, 0, len(gaps)+1)
	stages = append(stages, PlanStage{
		Number:    1,
		Title:     "AI基础认知",
		Status:    "进行中",
		Courses:   []string{"AI基础入门", "AI能力地图与应用场景"},
		Duration:  "4.5 小时",
		Goal:      "理解AI基本概念与能力边界，建立用AI解决问题的思维框架。",
		Milestone: "完成AI基础测验",
	})
	for index, gap := range gaps {
		stages = append(stages, PlanStage{
			Number:    index + 2,
			Title:     gap.Name,
			Status:    "未开始",
			Courses:   coursesForGap(gap.Name),
			Duration:  "6.0 小时",
			Goal:      gap.Recommended,
			Milestone: fmt.Sprintf("完成%s实战任务", gap.Name),
		})
	}
	return DiagnosisPlan{
		DiagnosisID:      diagnosis.ID,
		Title:            formatPlanTitle(diagnosis.Goal),
		Description:      fmt.Sprintf("基于你的项目方向与能力诊断结果，为你定制学习路径，助你掌握“%s”相关能力。", diagnosis.Project),
		Recommendations:  diagnosis.Recommendations,
		Stages:           stages,
		EstimatedHours:   24,
		WeeklySuggestion: "每周 6-8 小时",
		GeneratedAt:      diagnosis.UpdatedAt,
	}, nil
}

func (s *Service) LatestReport(ctx context.Context, userID int64) (DiagnosisReport, error) {
	diagnosis, err := s.LatestDiagnosis(ctx, userID)
	if err != nil {
		return DiagnosisReport{}, err
	}
	return DiagnosisReport{
		DiagnosisID:     diagnosis.ID,
		Goal:            diagnosis.Goal,
		Project:         diagnosis.Project,
		OverallScore:    diagnosis.OverallScore,
		Dimensions:      diagnosis.Dimensions,
		PriorityGaps:    prioritizedGaps(diagnosis),
		Recommendations: diagnosis.Recommendations,
		Evidence:        diagnosisEvidence(diagnosis),
		GeneratedAt:     diagnosis.UpdatedAt,
	}, nil
}

func deriveGaps(diagnosis Diagnosis) DiagnosisGaps {
	return DiagnosisGaps{
		DiagnosisID:  diagnosis.ID,
		Goal:         diagnosis.Goal,
		Project:      diagnosis.Project,
		OverallScore: diagnosis.OverallScore,
		Gaps:         prioritizedGaps(diagnosis),
		Evidence:     diagnosisEvidence(diagnosis),
		GeneratedAt:  diagnosis.UpdatedAt,
	}
}

func prioritizedGaps(diagnosis Diagnosis) []GapItem {
	dimensions := append([]Dimension(nil), diagnosis.Dimensions...)
	sort.SliceStable(dimensions, func(i, j int) bool {
		return dimensions[i].Gap > dimensions[j].Gap
	})
	if len(dimensions) > 3 {
		dimensions = dimensions[:3]
	}
	gaps := make([]GapItem, 0, len(dimensions))
	for index, dimension := range dimensions {
		target := dimension.Score + dimension.Gap
		if target > 100 {
			target = 100
		}
		gaps = append(gaps, GapItem{
			Name:        dimension.Name,
			Current:     dimension.Score,
			Target:      target,
			Gap:         dimension.Gap,
			Priority:    gapPriority(index),
			Summary:     dimension.Summary,
			Evidence:    fmt.Sprintf("诊断显示%s当前为%d分，目标差距%d分。", dimension.Name, dimension.Score, dimension.Gap),
			Recommended: fmt.Sprintf("优先补齐%s，结合%s项目做一次可交付练习。", dimension.Name, diagnosis.Project),
		})
	}
	return gaps
}

func gapPriority(index int) string {
	switch index {
	case 0:
		return "high"
	case 1:
		return "medium"
	default:
		return "normal"
	}
}

func diagnosisEvidence(diagnosis Diagnosis) []string {
	return []string{
		fmt.Sprintf("学习目标：%s", diagnosis.Goal),
		fmt.Sprintf("项目方向：%s", diagnosis.Project),
		"结合课程进度、任务应用和工具使用深度生成。",
	}
}

func learningOrder(gaps []GapItem) string {
	names := make([]string, 0, len(gaps))
	for _, gap := range gaps {
		names = append(names, gap.Name)
	}
	if len(names) == 0 {
		return "AI基础认知 → 提示词工程 → 行业分析"
	}
	return strings.Join(names, " → ")
}

func coursesForGap(name string) []string {
	if strings.Contains(name, "提示词") {
		return []string{"提示词工程实战", "AI工具箱实践指南"}
	}
	if strings.Contains(name, "数据") {
		return []string{"数据洞察与竞品研究实战", "AI行业分析方法"}
	}
	if strings.Contains(name, "市场") || strings.Contains(name, "行业") {
		return []string{"AI行业分析方法", "竞品全盘数据破解实战"}
	}
	return []string{name, "实战项目复盘"}
}

func formatPlanTitle(goal string) string {
	goal = strings.TrimSpace(strings.TrimPrefix(goal, "提升"))
	if goal == "" {
		return "系统学习路径"
	}
	return goal + "路径"
}
