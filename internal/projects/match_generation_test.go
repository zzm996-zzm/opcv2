package projects

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/zzm/opcv2/internal/ai"
	"github.com/zzm/opcv2/internal/jobs"
)

type generationMemoryRepository struct {
	*workflowMemoryRepository
	events []MatchProgressEvent
}

func (r *generationMemoryRepository) PrepareMatchGeneration(_ context.Context, userID, matchID int64) (MatchRun, error) {
	run, err := r.GetMatchRun(context.Background(), userID, matchID)
	if err != nil {
		return MatchRun{}, err
	}
	if run.Status != MatchStatusReady && run.Status != MatchStatusFailed && run.Status != MatchStatusCanceled {
		return MatchRun{}, ErrMatchNotReady
	}
	run.Status, run.GenerationAttempt, run.ProgressPercent, run.CurrentStep = MatchStatusQueued, run.GenerationAttempt+1, 0, MatchStepQueued
	for index := range r.runs {
		if r.runs[index].ID == run.ID {
			r.runs[index] = run
		}
	}
	r.events = append(r.events, MatchProgressEvent{ID: int64(len(r.events) + 1), MatchID: matchID, Attempt: run.GenerationAttempt, Event: MatchStepQueued})
	return run, nil
}

func (r *generationMemoryRepository) UpdateMatchGeneration(_ context.Context, userID, matchID int64, attempt int, status string, progress int, step, errorCode string, result *MatchResult) (MatchRun, error) {
	run, err := r.GetMatchRun(context.Background(), userID, matchID)
	if err != nil {
		return MatchRun{}, err
	}
	if run.GenerationAttempt != attempt || (run.Status != MatchStatusQueued && run.Status != MatchStatusRunning) {
		return MatchRun{}, ErrStaleMatchGeneration
	}
	run.Status, run.ProgressPercent, run.CurrentStep, run.ErrorCode = status, progress, step, errorCode
	if result != nil {
		run.Result = *result
	}
	for index := range r.runs {
		if r.runs[index].ID == run.ID {
			r.runs[index] = run
		}
	}
	r.events = append(r.events, MatchProgressEvent{ID: int64(len(r.events) + 1), MatchID: matchID, Attempt: attempt, Event: step, ProgressPercent: progress})
	return run, nil
}

func (r *generationMemoryRepository) CancelMatchGeneration(_ context.Context, userID, matchID int64) (MatchRun, error) {
	run, err := r.GetMatchRun(context.Background(), userID, matchID)
	if err != nil {
		return MatchRun{}, err
	}
	if run.Status != MatchStatusQueued && run.Status != MatchStatusRunning {
		return MatchRun{}, ErrMatchNotReady
	}
	run.Status, run.ProgressPercent, run.CurrentStep = MatchStatusCanceled, 100, MatchStepCanceled
	for index := range r.runs {
		if r.runs[index].ID == run.ID {
			r.runs[index] = run
		}
	}
	r.events = append(r.events, MatchProgressEvent{ID: int64(len(r.events) + 1), MatchID: matchID, Attempt: run.GenerationAttempt, Event: MatchStepCanceled, ProgressPercent: 100})
	return run, nil
}

func (r *generationMemoryRepository) ListMatchProgressEvents(_ context.Context, userID, matchID, afterID int64) ([]MatchProgressEvent, error) {
	if _, err := r.GetMatchRun(context.Background(), userID, matchID); err != nil {
		return nil, err
	}
	var events []MatchProgressEvent
	for _, event := range r.events {
		if event.MatchID == matchID && event.ID > afterID {
			events = append(events, event)
		}
	}
	return events, nil
}

type generationQueue struct{ jobs []jobs.Job }

func (q *generationQueue) Enqueue(_ context.Context, job jobs.Job) error {
	q.jobs = append(q.jobs, job)
	return nil
}

func TestGenerateProjectMatchEnqueuesOnce(t *testing.T) {
	repository := &generationMemoryRepository{workflowMemoryRepository: &workflowMemoryRepository{memoryRepository: &memoryRepository{}, runs: []MatchRun{{ID: 200, UserID: 42, WorkflowVersion: 2, Status: MatchStatusReady}}}}
	queue := &generationQueue{}
	service := NewService(repository, &fakeJSONGenerator{}, WithProjectMatchQueue(queue))
	first, err := service.GenerateProjectMatch(context.Background(), 42, 200)
	if err != nil {
		t.Fatalf("GenerateProjectMatch() error = %v", err)
	}
	second, err := service.GenerateProjectMatch(context.Background(), 42, 200)
	if err != nil {
		t.Fatalf("second GenerateProjectMatch() error = %v", err)
	}
	if first.Status != MatchStatusQueued || first.Attempt != 1 || second.Attempt != 1 || len(queue.jobs) != 1 || queue.jobs[0].Type != jobs.TypeProjectMatchGenerate {
		t.Fatalf("first/second/jobs = %+v/%+v/%+v", first, second, queue.jobs)
	}
}

func TestProcessProjectMatchPersistsProgressAndResult(t *testing.T) {
	content, _ := json.Marshal(MatchResult{Projects: []ProjectMatch{{Rank: 1, OpportunitySlug: "ai-sales", Title: "AI销售顾问", Score: 90, Tags: []string{"B端"}, Budget: "1万", Reasons: []string{"经验匹配"}, Risk: "需验证获客"}}})
	repository := &generationMemoryRepository{workflowMemoryRepository: &workflowMemoryRepository{
		memoryRepository: &memoryRepository{opportunities: []Opportunity{{Slug: "ai-sales", Title: "AI销售顾问", Status: OpportunityStatusPublished}}},
		runs:             []MatchRun{{ID: 200, UserID: 42, WorkflowVersion: 2, Need: "做销售项目", Status: MatchStatusQueued, GenerationAttempt: 1}},
	}}
	service := NewService(repository, &fakeJSONGenerator{result: ai.GenerateJSONResult{Content: content}})
	service.now = func() time.Time { return time.Date(2026, 8, 10, 8, 0, 0, 0, time.UTC) }
	if err := service.ProcessProjectMatch(context.Background(), 42, 200, 1); err != nil {
		t.Fatalf("ProcessProjectMatch() error = %v", err)
	}
	run := repository.runs[0]
	if run.Status != MatchStatusCompleted || run.ProgressPercent != 100 || run.CurrentStep != MatchStepDone || len(run.Result.Projects) != 1 || len(repository.events) != 6 {
		t.Fatalf("run/events = %+v/%+v", run, repository.events)
	}
}

func TestCanceledProjectMatchStopsBeforeGeneration(t *testing.T) {
	repository := &generationMemoryRepository{workflowMemoryRepository: &workflowMemoryRepository{memoryRepository: &memoryRepository{}, runs: []MatchRun{{ID: 200, UserID: 42, WorkflowVersion: 2, Status: MatchStatusCanceled, GenerationAttempt: 1}}}}
	service := NewService(repository, &fakeJSONGenerator{err: context.Canceled})
	if err := service.ProcessProjectMatch(context.Background(), 42, 200, 1); err != nil {
		t.Fatalf("ProcessProjectMatch() error = %v", err)
	}
}
