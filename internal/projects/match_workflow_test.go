package projects

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/zzm/opcv2/internal/ai"
)

type workflowMemoryRepository struct {
	*memoryRepository
	runs   []MatchRun
	nextID int64
}

func (r *workflowMemoryRepository) FindMatchRunByIdempotency(_ context.Context, userID int64, key string) (MatchRun, error) {
	for _, run := range r.runs {
		if run.UserID == userID && run.IdempotencyKey == key {
			return run, nil
		}
	}
	return MatchRun{}, ErrSessionNotFound
}

func (r *workflowMemoryRepository) CreateMatchRun(_ context.Context, run MatchRun) (MatchRun, bool, error) {
	for _, existing := range r.runs {
		if existing.UserID == run.UserID && existing.IdempotencyKey == run.IdempotencyKey {
			return existing, false, nil
		}
	}
	if r.nextID == 0 {
		r.nextID = 200
	}
	run.ID = r.nextID
	r.nextID++
	r.runs = append(r.runs, run)
	return run, true, nil
}

func (r *workflowMemoryRepository) GetMatchRun(_ context.Context, userID, id int64) (MatchRun, error) {
	for _, run := range r.runs {
		if run.UserID == userID && run.ID == id {
			return run, nil
		}
	}
	return MatchRun{}, ErrSessionNotFound
}

func (r *workflowMemoryRepository) ListMatchRuns(_ context.Context, userID int64, limit int) ([]MatchRun, error) {
	items := make([]MatchRun, 0)
	for index := len(r.runs) - 1; index >= 0 && len(items) < limit; index-- {
		if r.runs[index].UserID == userID {
			items = append(items, r.runs[index])
		}
	}
	return items, nil
}

func (r *workflowMemoryRepository) UpdateMatchRun(_ context.Context, run MatchRun, expectedRevision int) (MatchRun, error) {
	for index, existing := range r.runs {
		if existing.UserID == run.UserID && existing.ID == run.ID {
			if existing.Revision != expectedRevision {
				return MatchRun{}, ErrMatchRevisionConflict
			}
			run.Revision = expectedRevision + 1
			r.runs[index] = run
			return run, nil
		}
	}
	return MatchRun{}, ErrSessionNotFound
}

type sequenceWorkflowGenerator struct {
	contents [][]byte
	calls    int
}

type failingWorkflowGenerator struct {
	err error
}

func (g failingWorkflowGenerator) GenerateJSON(_ context.Context, _ ai.GenerateJSONRequest) (ai.GenerateJSONResult, error) {
	return ai.GenerateJSONResult{}, g.err
}

func (g *sequenceWorkflowGenerator) GenerateJSON(_ context.Context, request ai.GenerateJSONRequest) (ai.GenerateJSONResult, error) {
	if request.Feature != "projects.match_analysis" {
		return ai.GenerateJSONResult{}, errors.New("unexpected feature")
	}
	if g.calls >= len(g.contents) {
		return ai.GenerateJSONResult{}, errors.New("missing workflow response")
	}
	content := g.contents[g.calls]
	g.calls++
	if err := request.Validate(content); err != nil {
		return ai.GenerateJSONResult{}, err
	}
	return ai.GenerateJSONResult{Content: content}, nil
}

func TestCreateProjectMatchPersistsStructuredAnalysisAndIsIdempotent(t *testing.T) {
	repository := &workflowMemoryRepository{memoryRepository: &memoryRepository{}}
	generator := &sequenceWorkflowGenerator{contents: [][]byte{[]byte(`{
		"analysis_summary":"已识别内容能力，预算和时间仍需确认",
		"parsed_profile":{"skills":["内容创作"],"team_size":1},
		"field_sources":{"skills":[{"type":"text","locator":"need"}]},
		"completeness":0.45,
		"missing_fields":["budget_band","time_per_week","risk_preference","location"],
		"questions":[
			{"id":"budget","field":"budget_band","type":"single","question":"启动预算是多少？","options":["0-5k","5k-2w"],"required":true,"reason":"预算会改变候选范围"},
			{"id":"time","field":"time_per_week","type":"single","question":"每周投入多久？","options":["5小时","20小时"],"required":true,"reason":"时间影响项目复杂度"},
			{"id":"risk","field":"risk_preference","type":"single","question":"风险偏好？","options":["低","中"],"required":true,"reason":"风险影响排序"},
			{"id":"location","field":"location","type":"text","question":"主要经营地域？","options":[],"required":false,"reason":"地域影响渠道"}
		]
	}`)}}
	service := NewService(repository, generator)
	service.now = func() time.Time { return time.Date(2026, 8, 9, 10, 0, 0, 0, time.UTC) }
	input := CreateProjectMatchInput{
		UserID: 42, Need: "我会内容创作，想做一人公司", ProfilePatch: map[string]any{"team_size": 1}, IdempotencyKey: "match-create-42-1",
	}

	first, err := service.CreateProjectMatch(context.Background(), input)
	if err != nil {
		t.Fatalf("CreateProjectMatch() error = %v", err)
	}
	second, err := service.CreateProjectMatch(context.Background(), input)
	if err != nil {
		t.Fatalf("second CreateProjectMatch() error = %v", err)
	}
	if first.MatchID != second.MatchID || generator.calls != 1 {
		t.Fatalf("idempotency first/second/calls = %+v/%+v/%d", first, second, generator.calls)
	}
	if first.Status != MatchStatusClarifying || len(first.Questions) != 3 || first.Completeness != 0.45 {
		t.Fatalf("response = %+v", first)
	}
	if first.Need != input.Need || second.Need != input.Need {
		t.Fatalf("response needs = %q/%q, want %q", first.Need, second.Need, input.Need)
	}
	run := repository.runs[0]
	if run.WorkflowVersion != 2 || run.InputSnapshot.Need == "" || run.ParsedProfile["team_size"] == nil || len(run.FieldSources["skills"]) != 1 {
		t.Fatalf("run = %+v", run)
	}
}

func TestCreateProjectMatchFallsBackToDeterministicQuestionsWhenAnalysisFails(t *testing.T) {
	repository := &workflowMemoryRepository{memoryRepository: &memoryRepository{}}
	service := NewService(repository, failingWorkflowGenerator{err: ai.ErrInvalidModelJSON})

	response, err := service.CreateProjectMatch(context.Background(), CreateProjectMatchInput{UserID: 42, Need: "想做线上服务"})
	if err != nil {
		t.Fatalf("CreateProjectMatch() error = %v", err)
	}
	if response.Status != MatchStatusClarifying || len(response.Questions) != 3 || response.Questions[0].Field != "budget_band" {
		t.Fatalf("response = %+v", response)
	}
}

func TestAnswerProjectMatchFallsBackToReadyWhenAnalysisFails(t *testing.T) {
	repository := &workflowMemoryRepository{memoryRepository: &memoryRepository{}, runs: []MatchRun{{
		ID: 200, UserID: 42, WorkflowVersion: 2, Need: "做内容项目", Status: MatchStatusClarifying,
		Questions: []ClarificationQuestion{
			{ID: "budget", Field: "budget_band", Required: true},
			{ID: "time", Field: "time_per_week", Required: true},
			{ID: "risk", Field: "risk_preference", Required: true},
		},
		QuestionCount: 3, Revision: 1, InputSnapshot: MatchInputSnapshot{Need: "做内容项目"}, ParsedProfile: map[string]any{}, FieldSources: map[string][]MatchFieldSource{},
	}}}
	service := NewService(repository, failingWorkflowGenerator{err: ai.ErrInvalidModelJSON})

	response, err := service.AnswerProjectMatch(context.Background(), AnswerProjectMatchInput{
		UserID: 42, MatchID: 200, Revision: 1,
		Answers: []ClarificationAnswer{
			{QuestionID: "budget", Value: "2w以上"},
			{QuestionID: "time", Value: "20小时以上"},
			{QuestionID: "risk", Value: "低"},
		},
	})
	if err != nil {
		t.Fatalf("AnswerProjectMatch() error = %v", err)
	}
	if response.Status != MatchStatusReady || response.Completeness != 0.9 || len(response.Questions) != 0 {
		t.Fatalf("response = %+v", response)
	}
	if repository.runs[0].ParsedProfile["risk_preference"] != "低" {
		t.Fatalf("profile = %+v", repository.runs[0].ParsedProfile)
	}
}

func TestAnswerProjectMatchReanalyzesAndBecomesReady(t *testing.T) {
	repository := &workflowMemoryRepository{memoryRepository: &memoryRepository{}, runs: []MatchRun{{
		ID: 200, UserID: 42, WorkflowVersion: 2, Need: "做内容项目", Status: MatchStatusClarifying,
		Questions:     []ClarificationQuestion{{ID: "budget", Field: "budget_band", Required: true}},
		QuestionCount: 1, Revision: 1, InputSnapshot: MatchInputSnapshot{Need: "做内容项目"}, ParsedProfile: map[string]any{}, FieldSources: map[string][]MatchFieldSource{},
	}}}
	generator := &sequenceWorkflowGenerator{contents: [][]byte{[]byte(`{
		"analysis_summary":"信息完整，可以生成匹配结果",
		"parsed_profile":{"budget_band":"0-5k","team_size":1},
		"field_sources":{"budget_band":[{"type":"answer","locator":"budget"}]},
		"completeness":0.9,
		"missing_fields":[],
		"questions":[]
	}`)}}
	service := NewService(repository, generator)

	response, err := service.AnswerProjectMatch(context.Background(), AnswerProjectMatchInput{
		UserID: 42, MatchID: 200, Answers: []ClarificationAnswer{{QuestionID: "budget", Field: "budget_band", Value: "0-5k"}},
	})
	if err != nil {
		t.Fatalf("AnswerProjectMatch() error = %v", err)
	}
	if response.Status != MatchStatusReady || response.Completeness != 0.9 || len(response.Questions) != 0 {
		t.Fatalf("response = %+v", response)
	}
	run := repository.runs[0]
	if run.Rounds != 1 || len(run.AnswerEvents) != 1 || run.Revision != 2 || run.QuestionCount != 1 {
		t.Fatalf("run = %+v", run)
	}
}

func TestSkipProjectMatchAddsAssumptionsAndStopsClarification(t *testing.T) {
	repository := &workflowMemoryRepository{memoryRepository: &memoryRepository{}, runs: []MatchRun{{
		ID: 200, UserID: 42, WorkflowVersion: 2, Need: "做内容项目", Status: MatchStatusClarifying,
		MissingFields: []string{"budget_band", "risk_preference"}, Questions: []ClarificationQuestion{{ID: "budget"}},
		Revision: 1, InputSnapshot: MatchInputSnapshot{Need: "做内容项目"}, ParsedProfile: map[string]any{}, FieldSources: map[string][]MatchFieldSource{},
	}}}
	service := NewService(repository, &sequenceWorkflowGenerator{})

	response, err := service.AnswerProjectMatch(context.Background(), AnswerProjectMatchInput{UserID: 42, MatchID: 200, Skip: true})
	if err != nil {
		t.Fatalf("AnswerProjectMatch(skip) error = %v", err)
	}
	if response.Status != MatchStatusReady || len(response.Assumptions) != 2 || repository.runs[0].SkippedAt == nil {
		t.Fatalf("response/run = %+v/%+v", response, repository.runs[0])
	}
}

func TestProjectMatchWorkflowIsUserScoped(t *testing.T) {
	repository := &workflowMemoryRepository{memoryRepository: &memoryRepository{}, runs: []MatchRun{{ID: 200, UserID: 7, WorkflowVersion: 2}}}
	service := NewService(repository, &sequenceWorkflowGenerator{})

	_, err := service.GetProjectMatch(context.Background(), 42, 200)
	if !errors.Is(err, ErrSessionNotFound) {
		t.Fatalf("GetProjectMatch() error = %v, want not found", err)
	}
}

func TestListProjectMatchesReturnsUserScopedWorkflowResponses(t *testing.T) {
	now := time.Date(2026, 8, 12, 10, 0, 0, 0, time.UTC)
	repository := &workflowMemoryRepository{memoryRepository: &memoryRepository{}, runs: []MatchRun{
		{ID: 200, UserID: 42, WorkflowVersion: 2, Need: "做内容项目", Status: MatchStatusReady, CreatedAt: now, UpdatedAt: now},
		{ID: 201, UserID: 7, WorkflowVersion: 2, Need: "其他项目", Status: MatchStatusReady, CreatedAt: now, UpdatedAt: now},
	}}
	service := NewService(repository, &sequenceWorkflowGenerator{})
	items, err := service.ListProjectMatches(context.Background(), 42, 20)
	if err != nil || len(items) != 1 || items[0].MatchID != 200 || !items[0].CreatedAt.Equal(now) {
		t.Fatalf("items = %+v err=%v", items, err)
	}
}

func TestAnswerProjectMatchRejectsStaleRevision(t *testing.T) {
	repository := &workflowMemoryRepository{memoryRepository: &memoryRepository{}, runs: []MatchRun{{
		ID: 200, UserID: 42, WorkflowVersion: 2, Status: MatchStatusClarifying, Revision: 3,
	}}}
	service := NewService(repository, &sequenceWorkflowGenerator{})
	_, err := service.AnswerProjectMatch(context.Background(), AnswerProjectMatchInput{UserID: 42, MatchID: 200, Revision: 2, Skip: true})
	if !errors.Is(err, ErrMatchRevisionConflict) {
		t.Fatalf("AnswerProjectMatch() error = %v, want revision conflict", err)
	}
}

func TestAnswerProjectMatchIsIdempotentAndCapsTotalQuestions(t *testing.T) {
	repository := &workflowMemoryRepository{memoryRepository: &memoryRepository{}, runs: []MatchRun{{
		ID: 200, UserID: 42, WorkflowVersion: 2, Need: "做内容项目", Status: MatchStatusClarifying,
		Questions:     []ClarificationQuestion{{ID: "budget", Field: "budget_band", Required: true}},
		QuestionCount: 8, Revision: 1, InputSnapshot: MatchInputSnapshot{Need: "做内容项目"},
		ParsedProfile: map[string]any{}, FieldSources: map[string][]MatchFieldSource{},
	}}}
	generator := &sequenceWorkflowGenerator{contents: [][]byte{[]byte(`{
		"analysis_summary":"仍有信息待确认",
		"parsed_profile":{"budget_band":"0-5k"},
		"field_sources":{"budget_band":[{"type":"answer","locator":"budget"}]},
		"completeness":0.7,
		"missing_fields":["location"],
		"questions":[{"id":"location","field":"location","type":"text","question":"经营地域？","required":true}]
	}`)}}
	service := NewService(repository, generator)
	input := AnswerProjectMatchInput{
		UserID: 42, MatchID: 200, Revision: 1, IdempotencyKey: "answer-200-1",
		Answers: []ClarificationAnswer{{QuestionID: "budget", Value: "0-5k"}},
	}
	first, err := service.AnswerProjectMatch(context.Background(), input)
	if err != nil {
		t.Fatalf("AnswerProjectMatch() error = %v", err)
	}
	second, err := service.AnswerProjectMatch(context.Background(), input)
	if err != nil {
		t.Fatalf("second AnswerProjectMatch() error = %v", err)
	}
	if first.Status != MatchStatusReady || len(first.Questions) != 0 || first.Revision != 2 || second.Revision != 2 || generator.calls != 1 {
		t.Fatalf("first/second/calls = %+v/%+v/%d", first, second, generator.calls)
	}
}
