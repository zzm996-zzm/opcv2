package learning

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
)

type Repository interface {
	ListCourses(ctx context.Context, filter CourseFilter) ([]Course, error)
	GetCourse(ctx context.Context, slug string) (Course, error)
	ListProgress(ctx context.Context, userID int64) ([]Progress, error)
	GetProgress(ctx context.Context, userID int64, courseSlug string) (Progress, error)
	UpsertProgress(ctx context.Context, progress Progress) (Progress, error)
	CreateDiagnosis(ctx context.Context, diagnosis Diagnosis) (Diagnosis, error)
	GetDiagnosis(ctx context.Context, userID, id int64) (Diagnosis, error)
	LatestDiagnosis(ctx context.Context, userID int64) (Diagnosis, error)
}

type JSONGenerator interface {
	GenerateJSON(ctx context.Context, request ai.GenerateJSONRequest) (ai.GenerateJSONResult, error)
}

type ProfileContextProvider interface {
	GetProfileContext(ctx context.Context, userID int64) (account.ProfileContext, error)
}

type Option func(*Service)

func WithGenerator(generator JSONGenerator) Option {
	return func(service *Service) { service.generator = generator }
}

func WithProfileContextProvider(provider ProfileContextProvider) Option {
	return func(service *Service) { service.profile = provider }
}

type Service struct {
	repository Repository
	generator  JSONGenerator
	profile    ProfileContextProvider
	now        func() time.Time
}

func NewService(repository Repository, options ...Option) *Service {
	service := &Service{repository: repository, now: time.Now}
	for _, option := range options {
		option(service)
	}
	return service
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
	if s.repository == nil || s.generator == nil {
		return Diagnosis{}, ErrServiceNotReady
	}
	now := s.now()
	goal := strings.TrimSpace(input.Goal)
	project := strings.TrimSpace(input.Project)
	if input.UserID <= 0 || goal == "" || project == "" {
		return Diagnosis{}, ErrInvalidDiagnosis
	}
	input.Goal = goal
	input.Project = project
	input.FocusAbilities = normalizeUniqueStrings(input.FocusAbilities)
	input.WeeklyTime = strings.TrimSpace(input.WeeklyTime)
	input.Bottleneck = strings.TrimSpace(input.Bottleneck)
	input.Answers = normalizeAssessmentAnswers(input.Answers)

	profilePrompt, err := s.profilePrompt(ctx, input.UserID)
	if err != nil {
		return Diagnosis{}, err
	}
	result, err := s.generateDiagnosis(ctx, input, profilePrompt)
	if err != nil {
		return Diagnosis{}, err
	}
	diagnosis := Diagnosis{
		UserID:          input.UserID,
		Goal:            input.Goal,
		Project:         input.Project,
		FocusAbilities:  input.FocusAbilities,
		WeeklyTime:      input.WeeklyTime,
		Bottleneck:      input.Bottleneck,
		Answers:         input.Answers,
		Status:          DiagnosisCompleted,
		OverallScore:    result.OverallScore,
		Dimensions:      result.Dimensions,
		Recommendations: result.Recommendations,
		Basis:           "model_assessment",
		Disclaimer:      "本诊断由 AI 基于用户提交信息与可用画像进行模型评估，不代表标准化考试成绩或客观能力认证。",
		Assumptions:     result.Assumptions,
		EvidenceSources: diagnosisEvidenceSources(now, profilePrompt),
		CreatedAt:       now,
	}
	if len(diagnosis.Assumptions) == 0 {
		diagnosis.Assumptions = []string{
			fmt.Sprintf("诊断目标以“%s”为前提。", diagnosis.Goal),
			fmt.Sprintf("目标项目以“%s”的当前描述为前提。", diagnosis.Project),
			"当前诊断未接入标准化考试或外部能力认证数据。",
		}
	}
	diagnosis = attachSnapshots(diagnosis, now)
	created, err := s.repository.CreateDiagnosis(ctx, diagnosis)
	if err != nil {
		return Diagnosis{}, err
	}
	return withSnapshotDiagnosisID(created), nil
}

type diagnosisModelResult struct {
	OverallScore    int         `json:"overall_score"`
	Dimensions      []Dimension `json:"dimensions"`
	Recommendations []string    `json:"recommendations"`
	Assumptions     []string    `json:"assumptions"`
}

func (s *Service) generateDiagnosis(ctx context.Context, input CreateDiagnosisInput, profilePrompt string) (diagnosisModelResult, error) {
	result, err := s.generator.GenerateJSON(ctx, ai.GenerateJSONRequest{
		UserID:         input.UserID,
		Feature:        "learning.diagnosis",
		PromptVersion:  "learning_diagnosis_v2",
		SystemPrompt:   "你是学习能力诊断助手。只返回 JSON，字段为 overall_score、dimensions、recommendations、assumptions。所有分数都是基于输入的模型评估，不是考试成绩；不得声称读取了未提供的项目、任务、课程或工具数据。",
		UserPrompt:     appendPromptSection(diagnosisUserPrompt(input), profilePrompt),
		SchemaName:     "learning_diagnosis",
		Validate:       validateDiagnosisModelJSON,
		RepairAttempts: 1,
	})
	if err != nil {
		return diagnosisModelResult{}, fmt.Errorf("%w: %v", ErrInvalidAIResult, err)
	}
	var diagnosis diagnosisModelResult
	if err := json.Unmarshal(result.Content, &diagnosis); err != nil {
		return diagnosisModelResult{}, fmt.Errorf("%w: %v", ErrInvalidAIResult, err)
	}
	if err := validateDiagnosisModel(diagnosis); err != nil {
		return diagnosisModelResult{}, fmt.Errorf("%w: %v", ErrInvalidAIResult, err)
	}
	diagnosis.Assumptions = normalizeUniqueStrings(diagnosis.Assumptions)
	diagnosis.Recommendations = normalizeUniqueStrings(diagnosis.Recommendations)
	return diagnosis, nil
}

func validateDiagnosisModelJSON(data []byte) error {
	var diagnosis diagnosisModelResult
	if err := json.Unmarshal(data, &diagnosis); err != nil {
		return err
	}
	return validateDiagnosisModel(diagnosis)
}

func validateDiagnosisModel(diagnosis diagnosisModelResult) error {
	if diagnosis.OverallScore <= 0 || diagnosis.OverallScore > 100 || len(diagnosis.Dimensions) == 0 || len(diagnosis.Recommendations) == 0 {
		return errors.New("learning diagnosis is missing required fields")
	}
	for _, dimension := range diagnosis.Dimensions {
		if strings.TrimSpace(dimension.Name) == "" || strings.TrimSpace(dimension.Summary) == "" || dimension.Score < 0 || dimension.Score > 100 || dimension.Gap < 0 || dimension.Gap > 100 {
			return errors.New("learning diagnosis dimension is invalid")
		}
	}
	return nil
}

func normalizeAssessmentAnswers(values []AssessmentAnswer) []AssessmentAnswer {
	result := make([]AssessmentAnswer, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		value.Key = strings.TrimSpace(value.Key)
		value.Question = strings.TrimSpace(value.Question)
		value.Answer = strings.TrimSpace(value.Answer)
		if value.Key == "" || value.Answer == "" {
			continue
		}
		if _, exists := seen[value.Key]; exists {
			continue
		}
		seen[value.Key] = struct{}{}
		result = append(result, value)
	}
	return result
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

func diagnosisUserPrompt(input CreateDiagnosisInput) string {
	lines := []string{
		"学习目标：" + input.Goal,
		"目标项目：" + input.Project,
		"重点能力：" + strings.Join(input.FocusAbilities, "、"),
		"每周投入：" + input.WeeklyTime,
		"当前瓶颈：" + input.Bottleneck,
	}
	for _, answer := range input.Answers {
		lines = append(lines, fmt.Sprintf("评估回答[%s]：%s", answer.Key, answer.Answer))
	}
	return strings.Join(lines, "\n")
}

func appendPromptSection(base, section string) string {
	section = strings.TrimSpace(section)
	if section == "" {
		return base
	}
	return base + "\n\n" + section
}

func (s *Service) profilePrompt(ctx context.Context, userID int64) (string, error) {
	if s.profile == nil {
		return "", nil
	}
	profile, err := s.profile.GetProfileContext(ctx, userID)
	if err != nil {
		return "", err
	}
	if len(profile.Groups) == 0 {
		return "", nil
	}
	lines := []string{"用户画像上下文："}
	for _, group := range profile.Groups {
		fields := make([]string, 0, len(group.Fields))
		for key, value := range group.Fields {
			value = strings.TrimSpace(value)
			if value != "" {
				fields = append(fields, key+"="+value)
			}
		}
		if len(fields) == 0 {
			continue
		}
		sort.Strings(fields)
		lines = append(lines, "- "+group.Title+"："+strings.Join(fields, "；"))
	}
	if len(lines) == 1 {
		return "", nil
	}
	return strings.Join(lines, "\n"), nil
}

func diagnosisEvidenceSources(now time.Time, profilePrompt string) []DiagnosisEvidenceSource {
	sources := []DiagnosisEvidenceSource{{Type: "assessment_input", Label: "用户本次提交的诊断目标、项目与自述信息", CapturedAt: now}}
	if strings.TrimSpace(profilePrompt) != "" {
		sources = append(sources, DiagnosisEvidenceSource{Type: "user_profile", Label: "用户已保存的画像上下文", CapturedAt: now})
	}
	return sources
}

func attachSnapshots(diagnosis Diagnosis, generatedAt time.Time) Diagnosis {
	diagnosis.UpdatedAt = generatedAt
	diagnosis.GapsSnapshot = deriveGaps(diagnosis)
	diagnosis.RecommendationsSnapshot = deriveRecommendations(diagnosis)
	diagnosis.PlanSnapshot = derivePlan(diagnosis)
	diagnosis.ReportSnapshot = deriveReport(diagnosis)
	return diagnosis
}

func withSnapshotDiagnosisID(diagnosis Diagnosis) Diagnosis {
	diagnosis.GapsSnapshot.DiagnosisID = diagnosis.ID
	diagnosis.RecommendationsSnapshot.DiagnosisID = diagnosis.ID
	diagnosis.PlanSnapshot.DiagnosisID = diagnosis.ID
	diagnosis.ReportSnapshot.DiagnosisID = diagnosis.ID
	return diagnosis
}

func (s *Service) LatestDiagnosis(ctx context.Context, userID int64) (Diagnosis, error) {
	if s.repository == nil {
		return Diagnosis{}, ErrServiceNotReady
	}
	diagnosis, err := s.repository.LatestDiagnosis(ctx, userID)
	return withSnapshotDiagnosisID(diagnosis), err
}

func (s *Service) GetDiagnosis(ctx context.Context, userID, id int64) (Diagnosis, error) {
	if s.repository == nil {
		return Diagnosis{}, ErrServiceNotReady
	}
	if userID <= 0 || id <= 0 {
		return Diagnosis{}, ErrInvalidDiagnosis
	}
	diagnosis, err := s.repository.GetDiagnosis(ctx, userID, id)
	return withSnapshotDiagnosisID(diagnosis), err
}

func (s *Service) LatestGaps(ctx context.Context, userID int64) (DiagnosisGaps, error) {
	diagnosis, err := s.LatestDiagnosis(ctx, userID)
	if err != nil {
		return DiagnosisGaps{}, err
	}
	if len(diagnosis.GapsSnapshot.Gaps) > 0 {
		return diagnosis.GapsSnapshot, nil
	}
	return deriveGaps(diagnosis), nil
}

func (s *Service) LatestRecommendations(ctx context.Context, userID int64) (DiagnosisRecommendations, error) {
	diagnosis, err := s.LatestDiagnosis(ctx, userID)
	if err != nil {
		return DiagnosisRecommendations{}, err
	}
	if len(diagnosis.RecommendationsSnapshot.Focus) > 0 {
		return diagnosis.RecommendationsSnapshot, nil
	}
	return deriveRecommendations(diagnosis), nil
}

func (s *Service) LatestPlan(ctx context.Context, userID int64) (DiagnosisPlan, error) {
	diagnosis, err := s.LatestDiagnosis(ctx, userID)
	if err != nil {
		return DiagnosisPlan{}, err
	}
	if len(diagnosis.PlanSnapshot.Stages) > 0 {
		return diagnosis.PlanSnapshot, nil
	}
	return derivePlan(diagnosis), nil
}

func (s *Service) LatestReport(ctx context.Context, userID int64) (DiagnosisReport, error) {
	diagnosis, err := s.LatestDiagnosis(ctx, userID)
	if err != nil {
		return DiagnosisReport{}, err
	}
	if len(diagnosis.ReportSnapshot.Dimensions) > 0 {
		return diagnosis.ReportSnapshot, nil
	}
	return deriveReport(diagnosis), nil
}

func deriveGaps(diagnosis Diagnosis) DiagnosisGaps {
	return DiagnosisGaps{
		DiagnosisID:     diagnosis.ID,
		Goal:            diagnosis.Goal,
		Project:         diagnosis.Project,
		OverallScore:    diagnosis.OverallScore,
		Gaps:            prioritizedGaps(diagnosis),
		Evidence:        diagnosisEvidence(diagnosis),
		Basis:           diagnosis.Basis,
		Disclaimer:      diagnosis.Disclaimer,
		Assumptions:     diagnosis.Assumptions,
		EvidenceSources: diagnosis.EvidenceSources,
		GeneratedAt:     diagnosis.UpdatedAt,
	}
}

func deriveRecommendations(diagnosis Diagnosis) DiagnosisRecommendations {
	gaps := prioritizedGaps(diagnosis)
	focus := make([]RecommendationFocus, 0, len(gaps))
	for _, gap := range gaps {
		focus = append(focus, RecommendationFocus{Name: gap.Name, Priority: gap.Priority, Summary: gap.Recommended})
	}
	methods := []LearningMethod{{Title: "建议学习顺序", Value: learningOrder(gaps), Detail: "按模型评估的差距从高到低安排。"}}
	if diagnosis.WeeklyTime != "" {
		methods = append([]LearningMethod{{Title: "用户计划投入", Value: diagnosis.WeeklyTime, Detail: "来自本次诊断提交，实际节奏由用户自行确认。"}}, methods...)
	}
	return DiagnosisRecommendations{
		DiagnosisID: diagnosis.ID, Goal: diagnosis.Goal, Project: diagnosis.Project,
		Focus: focus, Recommendations: diagnosis.Recommendations, Methods: methods,
		Basis: diagnosis.Basis, Disclaimer: diagnosis.Disclaimer, Assumptions: diagnosis.Assumptions,
		EvidenceSources: diagnosis.EvidenceSources, GeneratedAt: diagnosis.UpdatedAt,
	}
}

func derivePlan(diagnosis Diagnosis) DiagnosisPlan {
	gaps := prioritizedGaps(diagnosis)
	stages := make([]PlanStage, 0, len(gaps))
	for index, gap := range gaps {
		stages = append(stages, PlanStage{
			Number: index + 1, Title: gap.Name, Status: "not_started", Courses: []string{},
			Goal: gap.Recommended, Milestone: fmt.Sprintf("提交一份%s的项目应用练习", gap.Name),
		})
	}
	return DiagnosisPlan{
		DiagnosisID: diagnosis.ID, Title: formatPlanTitle(diagnosis.Goal),
		Description:     fmt.Sprintf("基于“%s”目标与本次模型评估快照生成。", diagnosis.Project),
		Recommendations: diagnosis.Recommendations, Stages: stages, WeeklySuggestion: diagnosis.WeeklyTime,
		Basis: diagnosis.Basis, Disclaimer: diagnosis.Disclaimer, Assumptions: diagnosis.Assumptions,
		EvidenceSources: diagnosis.EvidenceSources, GeneratedAt: diagnosis.UpdatedAt,
	}
}

func deriveReport(diagnosis Diagnosis) DiagnosisReport {
	return DiagnosisReport{
		DiagnosisID: diagnosis.ID, Goal: diagnosis.Goal, Project: diagnosis.Project,
		OverallScore: diagnosis.OverallScore, Dimensions: diagnosis.Dimensions,
		PriorityGaps: prioritizedGaps(diagnosis), Recommendations: diagnosis.Recommendations,
		Evidence: diagnosisEvidence(diagnosis), Basis: diagnosis.Basis, Disclaimer: diagnosis.Disclaimer,
		Assumptions: diagnosis.Assumptions, EvidenceSources: diagnosis.EvidenceSources, GeneratedAt: diagnosis.UpdatedAt,
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
	evidence := make([]string, 0, len(diagnosis.EvidenceSources))
	for _, source := range diagnosis.EvidenceSources {
		evidence = append(evidence, source.Label)
	}
	if len(evidence) == 0 {
		return []string{"历史诊断未记录结构化输入依据，使用前需重新核验。"}
	}
	return evidence
}

func learningOrder(gaps []GapItem) string {
	names := make([]string, 0, len(gaps))
	for _, gap := range gaps {
		names = append(names, gap.Name)
	}
	if len(names) == 0 {
		return "等待有效诊断结果"
	}
	return strings.Join(names, " → ")
}

func formatPlanTitle(goal string) string {
	goal = strings.TrimSpace(strings.TrimPrefix(goal, "提升"))
	if goal == "" {
		return "系统学习路径"
	}
	return goal + "路径"
}
