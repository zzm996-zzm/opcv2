package learning

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/zzm/opcv2/internal/account"
	"github.com/zzm/opcv2/internal/ai"
)

type fakeRepository struct {
	courses    []Course
	materials  []CourseMaterial
	progress   []Progress
	diagnosis  Diagnosis
	planItems  []PlanItem
	created    Diagnosis
	repository error
}

func (r *fakeRepository) GetProgress(_ context.Context, userID int64, courseSlug string) (Progress, error) {
	for _, item := range r.progress {
		if item.UserID == userID && item.CourseSlug == courseSlug {
			return item, nil
		}
	}
	return Progress{}, ErrProgressNotFound
}

func (r *fakeRepository) UpsertProgress(_ context.Context, progress Progress) (Progress, error) {
	if r.repository != nil {
		return Progress{}, r.repository
	}
	progress.ID = 88
	for index, item := range r.progress {
		if item.UserID == progress.UserID && item.CourseSlug == progress.CourseSlug {
			r.progress[index] = progress
			return progress, nil
		}
	}
	r.progress = append(r.progress, progress)
	return progress, nil
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

func (r *fakeRepository) ListCourseMaterials(_ context.Context, courseSlug string) ([]CourseMaterial, error) {
	rows := make([]CourseMaterial, 0, len(r.materials))
	for _, material := range r.materials {
		if material.CourseSlug == courseSlug {
			rows = append(rows, material)
		}
	}
	return rows, r.repository
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

func (r *fakeRepository) GetDiagnosis(_ context.Context, userID, id int64) (Diagnosis, error) {
	if r.repository != nil {
		return Diagnosis{}, r.repository
	}
	if r.diagnosis.UserID != userID || r.diagnosis.ID != id {
		return Diagnosis{}, ErrDiagnosisNotFound
	}
	return r.diagnosis, nil
}

func (r *fakeRepository) ListPlanItems(_ context.Context, userID, diagnosisID int64) ([]PlanItem, error) {
	rows := make([]PlanItem, 0, len(r.planItems))
	for _, item := range r.planItems {
		if item.UserID == userID && item.DiagnosisID == diagnosisID {
			rows = append(rows, item)
		}
	}
	return rows, r.repository
}

func (r *fakeRepository) UpsertPlanItem(_ context.Context, item PlanItem) (PlanItem, error) {
	if r.repository != nil {
		return PlanItem{}, r.repository
	}
	item.ID = 77
	for index, current := range r.planItems {
		if current.UserID == item.UserID && current.DiagnosisID == item.DiagnosisID && current.StageNumber == item.StageNumber {
			r.planItems[index] = item
			return item, nil
		}
	}
	r.planItems = append(r.planItems, item)
	return item, nil
}

type fakeGenerator struct {
	content []byte
	request ai.GenerateJSONRequest
}

func (g *fakeGenerator) GenerateJSON(_ context.Context, request ai.GenerateJSONRequest) (ai.GenerateJSONResult, error) {
	g.request = request
	if request.Validate != nil {
		if err := request.Validate(g.content); err != nil {
			return ai.GenerateJSONResult{}, err
		}
	}
	return ai.GenerateJSONResult{Content: g.content}, nil
}

type fakeProfileProvider struct {
	profile account.ProfileContext
}

func (p *fakeProfileProvider) GetProfileContext(context.Context, int64) (account.ProfileContext, error) {
	return p.profile, nil
}

var diagnosisPayload = []byte(`{
	"overall_score":74,
	"dimensions":[
		{"name":"数据洞察能力","score":62,"gap":24,"summary":"尚未提供可验证作品"},
		{"name":"目标拆解能力","score":72,"gap":14,"summary":"目标明确但交付物不足"}
	],
	"recommendations":["完成一次项目数据分析练习","保留学习复盘记录"],
	"assumptions":["用户自述代表当前学习需求"]
}`)

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

func TestServiceListsPersistedCourseMaterials(t *testing.T) {
	repository := &fakeRepository{
		courses:   []Course{{Slug: "ai-basics", Title: "AI基础入门"}},
		materials: []CourseMaterial{{ID: 9, CourseSlug: "ai-basics", Title: "第一章讲义", MaterialType: "article"}},
	}
	service := NewService(repository)

	materials, err := service.ListCourseMaterials(context.Background(), " ai-basics ")

	if err != nil || len(materials) != 1 || materials[0].Title != "第一章讲义" {
		t.Fatalf("materials = %+v err=%v", materials, err)
	}
}

func TestServiceUpdatesUserCourseProgress(t *testing.T) {
	repository := &fakeRepository{courses: []Course{{Slug: "ai-market-analysis", Title: "AI行业分析方法"}}}
	service := NewService(repository)

	progress, err := service.UpdateProgress(context.Background(), UpdateProgressInput{
		UserID:            42,
		CourseSlug:        " ai-market-analysis ",
		Percent:           38,
		LastLesson:        " 2.3 行业规模与增长趋势分析 ",
		RecommendedAction: " 继续完成第2章 ",
	})

	if err != nil {
		t.Fatalf("UpdateProgress() error = %v", err)
	}
	if progress.UserID != 42 || progress.CourseSlug != "ai-market-analysis" || progress.CourseTitle != "AI行业分析方法" || progress.Percent != 38 {
		t.Fatalf("progress = %+v", progress)
	}
	if progress.LastLesson != "2.3 行业规模与增长趋势分析" || progress.RecommendedAction != "继续完成第2章" {
		t.Fatalf("progress text = %+v", progress)
	}
}

func TestServiceRejectsInvalidCourseProgress(t *testing.T) {
	service := NewService(&fakeRepository{})

	_, err := service.UpdateProgress(context.Background(), UpdateProgressInput{UserID: 42, CourseSlug: "ai-market-analysis", Percent: 101})

	if !errors.Is(err, ErrInvalidProgress) {
		t.Fatalf("err = %v, want ErrInvalidProgress", err)
	}
}

func TestServiceCreatesCompletedDiagnosis(t *testing.T) {
	now := time.Date(2026, 6, 30, 15, 0, 0, 0, time.UTC)
	repository := &fakeRepository{}
	generator := &fakeGenerator{content: diagnosisPayload}
	profile := &fakeProfileProvider{profile: account.ProfileContext{Groups: []account.ProfileGroup{{Title: "目标", Fields: map[string]string{"方向": "智能客服"}}}}}
	service := NewService(repository, WithGenerator(generator), WithProfileContextProvider(profile))
	service.now = func() time.Time { return now }

	diagnosis, err := service.CreateDiagnosis(context.Background(), CreateDiagnosisInput{
		UserID:         42,
		Goal:           " 提升智能客服和市场分析能力 ",
		Project:        " 智能客服系统 ",
		FocusAbilities: []string{" 数据洞察能力 ", "数据洞察能力", "提示词工程实战"},
		WeeklyTime:     " 5-8 小时 ",
		Bottleneck:     " 缺少真实项目案例 ",
		Answers: []AssessmentAnswer{
			{Key: "experience", Question: "是否有项目经验？", Answer: " 有一次试点 "},
			{Key: "experience", Question: "重复项", Answer: "忽略"},
		},
	})

	if err != nil {
		t.Fatalf("CreateDiagnosis() error = %v", err)
	}
	if diagnosis.ID != 99 || diagnosis.Status != DiagnosisCompleted {
		t.Fatalf("diagnosis = %+v", diagnosis)
	}
	if diagnosis.OverallScore != 74 || len(diagnosis.Dimensions) == 0 || len(diagnosis.Recommendations) == 0 {
		t.Fatalf("diagnosis missing generated content: %+v", diagnosis)
	}
	if repository.created.UserID != 42 || repository.created.CreatedAt != now {
		t.Fatalf("created = %+v", repository.created)
	}
	if repository.created.Goal != "提升智能客服和市场分析能力" || repository.created.Project != "智能客服系统" {
		t.Fatalf("created text fields = %+v", repository.created)
	}
	if len(repository.created.FocusAbilities) != 2 || repository.created.FocusAbilities[0] != "数据洞察能力" {
		t.Fatalf("created focus abilities = %+v", repository.created.FocusAbilities)
	}
	if repository.created.WeeklyTime != "5-8 小时" || repository.created.Bottleneck != "缺少真实项目案例" {
		t.Fatalf("created intake = %+v", repository.created)
	}
	if diagnosis.Basis != "model_assessment" || diagnosis.Disclaimer == "" || len(diagnosis.Assumptions) == 0 || len(diagnosis.EvidenceSources) != 2 {
		t.Fatalf("diagnosis provenance = %+v", diagnosis)
	}
	if len(diagnosis.Answers) != 1 || diagnosis.Answers[0].Answer != "有一次试点" {
		t.Fatalf("diagnosis answers = %+v", diagnosis.Answers)
	}
	if generator.request.Feature != "learning.diagnosis" || generator.request.PromptVersion != "learning_diagnosis_v2" || !strings.Contains(generator.request.UserPrompt, "用户画像上下文") {
		t.Fatalf("generator request = %+v", generator.request)
	}
	if len(repository.created.GapsSnapshot.Gaps) == 0 || len(repository.created.ReportSnapshot.EvidenceSources) != 2 {
		t.Fatalf("snapshots = %+v", repository.created)
	}
}

func TestServiceRejectsInvalidAIDiagnosis(t *testing.T) {
	repository := &fakeRepository{}
	service := NewService(repository, WithGenerator(&fakeGenerator{content: []byte(`{"overall_score":0}`)}))

	_, err := service.CreateDiagnosis(context.Background(), CreateDiagnosisInput{UserID: 42, Goal: "提升AI能力", Project: "智能客服"})

	if !errors.Is(err, ErrInvalidAIResult) {
		t.Fatalf("err = %v, want ErrInvalidAIResult", err)
	}
	if repository.created.ID != 0 {
		t.Fatalf("invalid diagnosis should not persist: %+v", repository.created)
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
	if recommendations.Focus[0].Name != "业务场景拆解" || len(recommendations.Methods) == 0 || recommendations.Methods[0].Value == "" {
		t.Fatalf("recommendations = %+v", recommendations)
	}

	plan, err := service.LatestPlan(context.Background(), 42)
	if err != nil {
		t.Fatalf("LatestPlan() error = %v", err)
	}
	if plan.Title != "企业AI落地能力路径" || len(plan.Stages) < 2 || plan.Stages[0].Title != "业务场景拆解" {
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

func TestServicePersistsAndMergesPlanItemCompletion(t *testing.T) {
	now := time.Date(2026, 7, 13, 9, 0, 0, 0, time.UTC)
	repository := &fakeRepository{diagnosis: Diagnosis{
		ID: 99, UserID: 42, Goal: "提升数据能力", Project: "智能客服", UpdatedAt: now,
		PlanSnapshot: DiagnosisPlan{DiagnosisID: 99, Stages: []PlanStage{{Number: 1, Title: "数据洞察能力", Status: "not_started"}}},
	}}
	service := NewService(repository)
	service.now = func() time.Time { return now }

	item, err := service.UpdatePlanItem(context.Background(), UpdatePlanItemInput{UserID: 42, DiagnosisID: 99, StageNumber: 1, Completed: true})
	if err != nil || !item.Completed || item.CompletedAt == nil || item.Title != "数据洞察能力" {
		t.Fatalf("item = %+v err=%v", item, err)
	}
	plan, err := service.GetPlan(context.Background(), 42, 99)
	if err != nil || len(plan.Items) != 1 || !plan.Items[0].Completed || plan.Stages[0].Status != "completed" {
		t.Fatalf("plan = %+v err=%v", plan, err)
	}
}

func TestServiceRejectsUnknownPlanStage(t *testing.T) {
	repository := &fakeRepository{diagnosis: Diagnosis{ID: 99, UserID: 42, PlanSnapshot: DiagnosisPlan{Stages: []PlanStage{{Number: 1, Title: "数据洞察能力"}}}}}
	service := NewService(repository)

	_, err := service.UpdatePlanItem(context.Background(), UpdatePlanItemInput{UserID: 42, DiagnosisID: 99, StageNumber: 2, Completed: true})

	if !errors.Is(err, ErrInvalidPlanItem) {
		t.Fatalf("err = %v, want ErrInvalidPlanItem", err)
	}
}
