package learning

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeRepository struct {
	courses    []Course
	progress   []Progress
	diagnosis  Diagnosis
	created    Diagnosis
	repository error
}

func (r *fakeRepository) ListCourses(_ context.Context, filter CourseFilter) ([]Course, error) {
	if r.repository != nil {
		return nil, r.repository
	}
	rows := make([]Course, 0, len(r.courses))
	for _, course := range r.courses {
		if filter.Category != "" && course.Category != filter.Category {
			continue
		}
		rows = append(rows, course)
	}
	if filter.Limit > 0 && len(rows) > filter.Limit {
		rows = rows[:filter.Limit]
	}
	return rows, nil
}

func (r *fakeRepository) GetCourse(_ context.Context, slug string) (Course, error) {
	if r.repository != nil {
		return Course{}, r.repository
	}
	for _, course := range r.courses {
		if course.Slug == slug {
			return course, nil
		}
	}
	return Course{}, ErrCourseNotFound
}

func (r *fakeRepository) ListProgress(_ context.Context, userID int64) ([]Progress, error) {
	if r.repository != nil {
		return nil, r.repository
	}
	rows := make([]Progress, 0, len(r.progress))
	for _, progress := range r.progress {
		if progress.UserID == userID {
			rows = append(rows, progress)
		}
	}
	return rows, nil
}

func (r *fakeRepository) CreateDiagnosis(_ context.Context, diagnosis Diagnosis) (Diagnosis, error) {
	if r.repository != nil {
		return Diagnosis{}, r.repository
	}
	r.created = diagnosis
	diagnosis.ID = 99
	diagnosis.UpdatedAt = diagnosis.CreatedAt
	r.diagnosis = diagnosis
	return diagnosis, nil
}

func (r *fakeRepository) LatestDiagnosis(_ context.Context, userID int64) (Diagnosis, error) {
	if r.repository != nil {
		return Diagnosis{}, r.repository
	}
	if r.diagnosis.UserID != userID {
		return Diagnosis{}, ErrDiagnosisNotFound
	}
	return r.diagnosis, nil
}

func TestServiceListsCoursesByCategory(t *testing.T) {
	repository := &fakeRepository{courses: []Course{
		{ID: 1, Slug: "ai-basics", Title: "AI基础入门", Category: "入门"},
		{ID: 2, Slug: "prompt-engineering", Title: "提示词工程实战", Category: "实战"},
	}}
	service := NewService(repository)

	courses, err := service.ListCourses(context.Background(), CourseFilter{Category: "实战", Limit: 20})

	if err != nil {
		t.Fatalf("ListCourses() error = %v", err)
	}
	if len(courses) != 1 || courses[0].Slug != "prompt-engineering" {
		t.Fatalf("courses = %+v", courses)
	}
}

func TestServiceCreatesCompletedDiagnosis(t *testing.T) {
	now := time.Date(2026, 6, 30, 15, 0, 0, 0, time.UTC)
	repository := &fakeRepository{}
	service := NewService(repository)
	service.now = func() time.Time { return now }

	diagnosis, err := service.CreateDiagnosis(context.Background(), CreateDiagnosisInput{
		UserID:  42,
		Goal:    "提升智能客服和市场分析能力",
		Project: "智能客服系统",
	})

	if err != nil {
		t.Fatalf("CreateDiagnosis() error = %v", err)
	}
	if diagnosis.ID != 99 || diagnosis.Status != DiagnosisCompleted {
		t.Fatalf("diagnosis = %+v", diagnosis)
	}
	if len(diagnosis.Dimensions) == 0 || len(diagnosis.Recommendations) == 0 {
		t.Fatalf("diagnosis missing generated content: %+v", diagnosis)
	}
	if repository.created.UserID != 42 || repository.created.CreatedAt != now {
		t.Fatalf("created = %+v", repository.created)
	}
}

func TestServiceKeepsDiagnosisUserScoped(t *testing.T) {
	service := NewService(&fakeRepository{diagnosis: Diagnosis{ID: 7, UserID: 99}})

	_, err := service.LatestDiagnosis(context.Background(), 42)

	if !errors.Is(err, ErrDiagnosisNotFound) {
		t.Fatalf("err = %v, want ErrDiagnosisNotFound", err)
	}
}

func TestServiceDerivesLearningViewsFromLatestDiagnosis(t *testing.T) {
	now := time.Date(2026, 6, 30, 15, 0, 0, 0, time.UTC)
	service := NewService(&fakeRepository{diagnosis: Diagnosis{
		ID:           99,
		UserID:       42,
		Goal:         "提升企业AI落地能力",
		Project:      "企业AI运营项目",
		OverallScore: 82,
		Dimensions: []Dimension{
			{Name: "自动化运营能力", Score: 88, Gap: 6, Summary: "表现较好"},
			{Name: "业务场景拆解", Score: 58, Gap: 24, Summary: "需要补齐场景拆解方法"},
			{Name: "数据洞察能力", Score: 64, Gap: 18, Summary: "需要加强洞察输出"},
		},
		Recommendations: []string{"优先补齐业务场景拆解"},
		UpdatedAt:       now,
	}})

	gaps, err := service.LatestGaps(context.Background(), 42)
	if err != nil {
		t.Fatalf("LatestGaps() error = %v", err)
	}
	if len(gaps.Gaps) != 3 || gaps.Gaps[0].Name != "业务场景拆解" || gaps.Gaps[0].Priority != "high" {
		t.Fatalf("gaps = %+v", gaps)
	}

	recommendations, err := service.LatestRecommendations(context.Background(), 42)
	if err != nil {
		t.Fatalf("LatestRecommendations() error = %v", err)
	}
	if recommendations.Focus[0].Name != "业务场景拆解" || recommendations.Methods[2].Value == "" {
		t.Fatalf("recommendations = %+v", recommendations)
	}

	plan, err := service.LatestPlan(context.Background(), 42)
	if err != nil {
		t.Fatalf("LatestPlan() error = %v", err)
	}
	if plan.Title != "企业AI落地能力路径" || len(plan.Stages) < 2 || plan.Stages[1].Title != "业务场景拆解" {
		t.Fatalf("plan = %+v", plan)
	}

	report, err := service.LatestReport(context.Background(), 42)
	if err != nil {
		t.Fatalf("LatestReport() error = %v", err)
	}
	if report.OverallScore != 82 || report.PriorityGaps[0].Name != "业务场景拆解" || len(report.Evidence) == 0 {
		t.Fatalf("report = %+v", report)
	}
}
