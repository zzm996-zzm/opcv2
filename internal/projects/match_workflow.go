package projects

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/zzm/opcv2/internal/ai"
)

const (
	MatchWorkflowVersion  = 2
	MatchStatusClarifying = "clarifying"
	MatchStatusReady      = "ready"
)

var (
	ErrInvalidMatchRequest   = errors.New("invalid project match request")
	ErrMatchRevisionConflict = errors.New("project match revision conflict")
)

type MatchInputSnapshot struct {
	Need         string         `json:"need"`
	ProfilePatch map[string]any `json:"profile_patch,omitempty"`
}

type MatchFieldSource struct {
	Type      string `json:"type"`
	Locator   string `json:"locator"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

type ClarificationQuestion struct {
	ID       string   `json:"id"`
	Field    string   `json:"field"`
	Type     string   `json:"type"`
	Question string   `json:"question"`
	Options  []string `json:"options,omitempty"`
	Required bool     `json:"required"`
	Reason   string   `json:"reason,omitempty"`
}

type ClarificationAnswer struct {
	QuestionID string `json:"question_id"`
	Field      string `json:"field,omitempty"`
	Value      any    `json:"value"`
}

type MatchAnswerEvent struct {
	Round          int                   `json:"round"`
	IdempotencyKey string                `json:"idempotency_key,omitempty"`
	Answers        []ClarificationAnswer `json:"answers"`
	CreatedAt      time.Time             `json:"created_at"`
}

type MatchRun struct {
	ID              int64                         `json:"match_id"`
	UserID          int64                         `json:"-"`
	WorkflowVersion int                           `json:"workflow_version"`
	Name            string                        `json:"name,omitempty"`
	Need            string                        `json:"need"`
	IdempotencyKey  string                        `json:"-"`
	Status          string                        `json:"status"`
	InputSnapshot   MatchInputSnapshot            `json:"input_snapshot"`
	ParsedProfile   map[string]any                `json:"parsed_profile"`
	FieldSources    map[string][]MatchFieldSource `json:"field_sources"`
	AnalysisSummary string                        `json:"analysis_summary,omitempty"`
	Completeness    float64                       `json:"completeness"`
	MissingFields   []string                      `json:"missing_fields,omitempty"`
	Questions       []ClarificationQuestion       `json:"questions,omitempty"`
	QuestionCount   int                           `json:"question_count"`
	Rounds          int                           `json:"rounds"`
	AnswerEvents    []MatchAnswerEvent            `json:"answer_events,omitempty"`
	Assumptions     []string                      `json:"assumptions,omitempty"`
	Revision        int                           `json:"revision"`
	SkippedAt       *time.Time                    `json:"skipped_at,omitempty"`
	CreatedAt       time.Time                     `json:"created_at"`
	UpdatedAt       time.Time                     `json:"updated_at"`
}

type CreateProjectMatchInput struct {
	UserID         int64          `json:"-"`
	Need           string         `json:"need"`
	ProfilePatch   map[string]any `json:"profile_patch,omitempty"`
	IdempotencyKey string         `json:"-"`
}

type AnswerProjectMatchInput struct {
	UserID         int64                 `json:"-"`
	MatchID        int64                 `json:"-"`
	Revision       int                   `json:"revision,omitempty"`
	Answers        []ClarificationAnswer `json:"answers,omitempty"`
	Skip           bool                  `json:"skip,omitempty"`
	IdempotencyKey string                `json:"-"`
}

type MatchWorkflowResponse struct {
	MatchID         int64                         `json:"match_id"`
	Status          string                        `json:"status"`
	AnalysisSummary string                        `json:"analysis_summary,omitempty"`
	ParsedProfile   map[string]any                `json:"parsed_profile,omitempty"`
	FieldSources    map[string][]MatchFieldSource `json:"field_sources,omitempty"`
	Completeness    float64                       `json:"completeness"`
	MissingFields   []string                      `json:"missing_fields,omitempty"`
	Questions       []ClarificationQuestion       `json:"questions,omitempty"`
	Assumptions     []string                      `json:"assumptions,omitempty"`
	Revision        int                           `json:"revision"`
}

type MatchWorkflowRepository interface {
	FindMatchRunByIdempotency(context.Context, int64, string) (MatchRun, error)
	CreateMatchRun(context.Context, MatchRun) (MatchRun, bool, error)
	GetMatchRun(context.Context, int64, int64) (MatchRun, error)
	UpdateMatchRun(context.Context, MatchRun, int) (MatchRun, error)
}

func (s *Service) workflowRepository() (MatchWorkflowRepository, error) {
	if s.repository == nil {
		return nil, ErrServiceNotReady
	}
	repository, ok := s.repository.(MatchWorkflowRepository)
	if !ok {
		return nil, ErrServiceNotReady
	}
	return repository, nil
}

func (s *Service) CreateProjectMatch(ctx context.Context, input CreateProjectMatchInput) (MatchWorkflowResponse, error) {
	repository, err := s.workflowRepository()
	if err != nil {
		return MatchWorkflowResponse{}, err
	}
	input.Need = strings.TrimSpace(input.Need)
	input.IdempotencyKey = strings.TrimSpace(input.IdempotencyKey)
	if input.Need == "" || input.UserID <= 0 {
		return MatchWorkflowResponse{}, ErrInvalidMatchRequest
	}
	if s.generator == nil {
		return MatchWorkflowResponse{}, ErrServiceNotReady
	}
	if input.IdempotencyKey != "" {
		if existing, findErr := repository.FindMatchRunByIdempotency(ctx, input.UserID, input.IdempotencyKey); findErr == nil {
			return matchWorkflowResponse(existing), nil
		} else if !errors.Is(findErr, ErrSessionNotFound) {
			return MatchWorkflowResponse{}, findErr
		}
	}

	analysis, err := s.generateWorkflowAnalysis(ctx, input.UserID, input.Need, input.ProfilePatch, nil)
	if err != nil {
		return MatchWorkflowResponse{}, err
	}
	profile := cloneMap(analysis.ParsedProfile)
	if profile == nil {
		profile = map[string]any{}
	}
	sources := cloneFieldSources(analysis.FieldSources)
	if sources == nil {
		sources = map[string][]MatchFieldSource{}
	}
	for field, value := range input.ProfilePatch {
		profile[field] = value
		sources[field] = []MatchFieldSource{{Type: "profile_patch", Locator: field}}
	}
	questions := filterAnsweredQuestions(normalizeQuestions(analysis.Questions, 3), profile)
	status := MatchStatusClarifying
	if analysis.Completeness >= 0.8 || len(questions) == 0 {
		status = MatchStatusReady
		questions = []ClarificationQuestion{}
	}
	now := s.now()
	run := MatchRun{
		UserID: input.UserID, WorkflowVersion: MatchWorkflowVersion, Need: input.Need,
		IdempotencyKey: input.IdempotencyKey, Status: status,
		InputSnapshot: MatchInputSnapshot{Need: input.Need, ProfilePatch: input.ProfilePatch},
		ParsedProfile: profile, FieldSources: sources, AnalysisSummary: analysis.AnalysisSummary,
		Completeness: clampCompleteness(analysis.Completeness), MissingFields: filterPresentFields(analysis.MissingFields, profile),
		Questions: questions, QuestionCount: len(questions), Revision: 1, CreatedAt: now, UpdatedAt: now,
	}
	run, created, err := repository.CreateMatchRun(ctx, run)
	if err != nil {
		return MatchWorkflowResponse{}, err
	}
	if !created {
		return matchWorkflowResponse(run), nil
	}
	return matchWorkflowResponse(run), nil
}

func (s *Service) AnswerProjectMatch(ctx context.Context, input AnswerProjectMatchInput) (MatchWorkflowResponse, error) {
	repository, err := s.workflowRepository()
	if err != nil {
		return MatchWorkflowResponse{}, err
	}
	run, err := repository.GetMatchRun(ctx, input.UserID, input.MatchID)
	if err != nil {
		return MatchWorkflowResponse{}, err
	}
	input.IdempotencyKey = strings.TrimSpace(input.IdempotencyKey)
	if input.IdempotencyKey != "" && hasAnswerEvent(run.AnswerEvents, input.IdempotencyKey) {
		return matchWorkflowResponse(run), nil
	}
	if run.Status != MatchStatusClarifying {
		return matchWorkflowResponse(run), nil
	}
	expectedRevision := input.Revision
	if expectedRevision == 0 {
		expectedRevision = run.Revision
	}
	if expectedRevision != run.Revision {
		return MatchWorkflowResponse{}, ErrMatchRevisionConflict
	}
	now := s.now()
	if input.Skip {
		run.Status = MatchStatusReady
		run.Assumptions = appendUnique(run.Assumptions, run.MissingFields...)
		run.Questions = nil
		run.SkippedAt = &now
	} else {
		if s.generator == nil {
			return MatchWorkflowResponse{}, ErrServiceNotReady
		}
		if len(input.Answers) == 0 || len(input.Answers) > len(run.Questions) {
			return MatchWorkflowResponse{}, ErrInvalidMatchAnswers
		}
		questionsByID := map[string]ClarificationQuestion{}
		for _, question := range run.Questions {
			questionsByID[question.ID] = question
		}
		answersByQuestion := map[string]ClarificationAnswer{}
		for _, answer := range input.Answers {
			answer.QuestionID = strings.TrimSpace(answer.QuestionID)
			question, known := questionsByID[answer.QuestionID]
			if answer.QuestionID == "" || !known {
				return MatchWorkflowResponse{}, ErrInvalidMatchAnswers
			}
			if answer.Field == "" {
				answer.Field = question.Field
			}
			answersByQuestion[answer.QuestionID] = answer
		}
		for _, question := range run.Questions {
			if question.Required {
				answer, ok := answersByQuestion[question.ID]
				if !ok || !hasMatchValue(answer.Value) {
					return MatchWorkflowResponse{}, ErrInvalidMatchAnswers
				}
			}
		}
		run.Rounds++
		storedAnswers := make([]ClarificationAnswer, 0, len(input.Answers))
		for _, answer := range input.Answers {
			storedAnswers = append(storedAnswers, answersByQuestion[strings.TrimSpace(answer.QuestionID)])
		}
		run.AnswerEvents = append(run.AnswerEvents, MatchAnswerEvent{Round: run.Rounds, IdempotencyKey: input.IdempotencyKey, Answers: storedAnswers, CreatedAt: now})
		profile := cloneMap(run.ParsedProfile)
		if profile == nil {
			profile = map[string]any{}
		}
		for _, answer := range storedAnswers {
			if answer.Field != "" {
				profile[answer.Field] = answer.Value
			}
		}
		analysis, genErr := s.generateWorkflowAnalysis(ctx, input.UserID, run.Need, run.InputSnapshot.ProfilePatch, profile)
		if genErr != nil {
			return MatchWorkflowResponse{}, genErr
		}
		run.ParsedProfile = profile
		for field, value := range analysis.ParsedProfile {
			if _, exists := profile[field]; !exists {
				profile[field] = value
			}
		}
		run.FieldSources = mergeFieldSources(run.FieldSources, analysis.FieldSources)
		for field := range run.InputSnapshot.ProfilePatch {
			run.FieldSources[field] = []MatchFieldSource{{Type: "profile_patch", Locator: field}}
		}
		for _, answer := range storedAnswers {
			if answer.Field != "" {
				run.FieldSources[answer.Field] = []MatchFieldSource{{Type: "answer", Locator: answer.QuestionID}}
			}
		}
		run.AnalysisSummary, run.Completeness = analysis.AnalysisSummary, clampCompleteness(analysis.Completeness)
		run.MissingFields = filterPresentFields(analysis.MissingFields, profile)
		remaining := 8 - run.QuestionCount
		if remaining < 0 {
			remaining = 0
		}
		if run.Rounds >= 3 {
			remaining = 0
		}
		run.Questions = filterAnsweredQuestions(normalizeQuestions(analysis.Questions, remaining), profile)
		run.QuestionCount += len(run.Questions)
		if run.Completeness >= 0.8 || run.Rounds >= 3 || run.QuestionCount >= 8 || len(run.Questions) == 0 {
			run.Status = MatchStatusReady
			run.Questions = []ClarificationQuestion{}
		}
	}
	run.UpdatedAt = now
	updated, err := repository.UpdateMatchRun(ctx, run, expectedRevision)
	if err != nil {
		return MatchWorkflowResponse{}, err
	}
	return matchWorkflowResponse(updated), nil
}

func (s *Service) GetProjectMatch(ctx context.Context, userID, id int64) (MatchWorkflowResponse, error) {
	repository, err := s.workflowRepository()
	if err != nil {
		return MatchWorkflowResponse{}, err
	}
	run, err := repository.GetMatchRun(ctx, userID, id)
	if err != nil {
		return MatchWorkflowResponse{}, err
	}
	return matchWorkflowResponse(run), nil
}

type workflowAnalysis struct {
	AnalysisSummary string                        `json:"analysis_summary"`
	ParsedProfile   map[string]any                `json:"parsed_profile"`
	FieldSources    map[string][]MatchFieldSource `json:"field_sources"`
	Completeness    float64                       `json:"completeness"`
	MissingFields   []string                      `json:"missing_fields"`
	Questions       []ClarificationQuestion       `json:"questions"`
}

func (s *Service) generateWorkflowAnalysis(ctx context.Context, userID int64, need string, patch map[string]any, profile map[string]any) (workflowAnalysis, error) {
	prompt, err := json.Marshal(struct {
		Need           string         `json:"need"`
		ProfilePatch   map[string]any `json:"profile_patch,omitempty"`
		CurrentProfile map[string]any `json:"current_profile,omitempty"`
	}{Need: need, ProfilePatch: patch, CurrentProfile: profile})
	if err != nil {
		return workflowAnalysis{}, fmt.Errorf("%w: %v", ErrInvalidMatchRequest, err)
	}
	request := ai.GenerateJSONRequest{UserID: userID, Feature: "projects.match_analysis", PromptVersion: "project_match_analysis_v1", SystemPrompt: "你是项目超市的需求分析助手。根据用户当前输入提取结构化画像、字段来源、完整度、缺失字段和最多三个澄清问题。用户明确填写的 profile_patch 和回答优先级最高。只返回 JSON，不输出思维链。", UserPrompt: string(prompt), SchemaName: "project_match_analysis", RepairAttempts: 1}
	request.Validate = func(content []byte) error {
		var payload workflowAnalysis
		if err := json.Unmarshal(content, &payload); err != nil {
			return err
		}
		if payload.Completeness < 0 || payload.Completeness > 1 {
			return errors.New("completeness must be between 0 and 1")
		}
		return nil
	}
	result, err := s.generator.GenerateJSON(ctx, request)
	if err != nil {
		return workflowAnalysis{}, fmt.Errorf("%w: %v", ErrInvalidAIResult, err)
	}
	var payload workflowAnalysis
	if err := json.Unmarshal(result.Content, &payload); err != nil {
		return workflowAnalysis{}, fmt.Errorf("%w: %v", ErrInvalidAIResult, err)
	}
	return payload, nil
}

func matchWorkflowResponse(run MatchRun) MatchWorkflowResponse {
	return MatchWorkflowResponse{MatchID: run.ID, Status: run.Status, AnalysisSummary: run.AnalysisSummary, ParsedProfile: cloneMap(run.ParsedProfile), FieldSources: cloneFieldSources(run.FieldSources), Completeness: run.Completeness, MissingFields: append([]string(nil), run.MissingFields...), Questions: append([]ClarificationQuestion(nil), run.Questions...), Assumptions: append([]string(nil), run.Assumptions...), Revision: run.Revision}
}

func normalizeQuestions(questions []ClarificationQuestion, limit int) []ClarificationQuestion {
	if limit <= 0 {
		return []ClarificationQuestion{}
	}
	if limit > 3 {
		limit = 3
	}
	seen := map[string]bool{}
	result := make([]ClarificationQuestion, 0, limit)
	for _, question := range questions {
		question.ID, question.Field = strings.TrimSpace(question.ID), strings.TrimSpace(question.Field)
		if question.ID == "" || seen[question.ID] {
			continue
		}
		seen[question.ID] = true
		result = append(result, question)
		if len(result) == limit {
			break
		}
	}
	return result
}

func hasAnswerEvent(events []MatchAnswerEvent, idempotencyKey string) bool {
	for _, event := range events {
		if event.IdempotencyKey == idempotencyKey {
			return true
		}
	}
	return false
}

func hasMatchValue(value any) bool {
	if value == nil {
		return false
	}
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed) != ""
	case []any:
		return len(typed) > 0
	case []string:
		return len(typed) > 0
	default:
		return strings.TrimSpace(fmt.Sprint(value)) != ""
	}
}

func filterAnsweredQuestions(questions []ClarificationQuestion, profile map[string]any) []ClarificationQuestion {
	result := make([]ClarificationQuestion, 0, len(questions))
	for _, question := range questions {
		if value, found := profile[question.Field]; found && hasMatchValue(value) {
			continue
		}
		result = append(result, question)
	}
	return result
}

func filterPresentFields(fields []string, profile map[string]any) []string {
	result := make([]string, 0, len(fields))
	for _, field := range fields {
		field = strings.TrimSpace(field)
		if field == "" {
			continue
		}
		if value, found := profile[field]; found && hasMatchValue(value) {
			continue
		}
		result = appendUnique(result, field)
	}
	return result
}

func clampCompleteness(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}
	return value
}
func cloneMap(input map[string]any) map[string]any {
	if input == nil {
		return nil
	}
	output := map[string]any{}
	for key, value := range input {
		output[key] = value
	}
	return output
}
func cloneFieldSources(input map[string][]MatchFieldSource) map[string][]MatchFieldSource {
	if input == nil {
		return nil
	}
	output := map[string][]MatchFieldSource{}
	for key, value := range input {
		output[key] = append([]MatchFieldSource(nil), value...)
	}
	return output
}
func mergeFieldSources(left, right map[string][]MatchFieldSource) map[string][]MatchFieldSource {
	output := cloneFieldSources(left)
	if output == nil {
		output = map[string][]MatchFieldSource{}
	}
	for key, value := range right {
		seen := map[string]bool{}
		for _, source := range output[key] {
			seen[source.Type+"\x00"+source.Locator] = true
		}
		for _, source := range value {
			sourceKey := source.Type + "\x00" + source.Locator
			if !seen[sourceKey] {
				output[key] = append(output[key], source)
				seen[sourceKey] = true
			}
		}
	}
	return output
}
func appendUnique(values []string, additions ...string) []string {
	seen := map[string]bool{}
	for _, value := range values {
		seen[value] = true
	}
	for _, value := range additions {
		value = strings.TrimSpace(value)
		if value != "" && !seen[value] {
			values = append(values, value)
			seen[value] = true
		}
	}
	return values
}
