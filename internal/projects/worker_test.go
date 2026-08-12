package projects

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hibiken/asynq"
	"github.com/zzm/opcv2/internal/jobs"
)

type fakeProjectMatchProcessor struct {
	userID       int64
	matchID      int64
	attempt      int
	projectID    int64
	reindexID    int64
	contentRunID int64
}

func (p *fakeProjectMatchProcessor) ProcessProjectHeat(_ context.Context, projectID int64) error {
	p.projectID = projectID
	return nil
}
func (p *fakeProjectMatchProcessor) ProcessProjectKBReindex(_ context.Context, id int64) error {
	p.reindexID = id
	return nil
}
func (p *fakeProjectMatchProcessor) ProcessProjectContentBatch(_ context.Context, id int64) error {
	p.contentRunID = id
	return nil
}

func (p *fakeProjectMatchProcessor) ProcessProjectExport(_ context.Context, userID, exportID int64) error {
	p.userID, p.matchID = userID, exportID
	return nil
}

func (p *fakeProjectMatchProcessor) ProcessProjectMatch(_ context.Context, userID, matchID int64, attempt int) error {
	p.userID, p.matchID, p.attempt = userID, matchID, attempt
	return nil
}

func TestProjectExportWorkerProcessesEnvelope(t *testing.T) {
	processor := &fakeProjectMatchProcessor{}
	mux := asynq.NewServeMux()
	RegisterWorker(mux, processor)
	payload, err := json.Marshal(jobs.Envelope{IdempotencyKey: "project-export-71", Payload: map[string]any{"user_id": 42, "export_id": 71}})
	if err != nil {
		t.Fatal(err)
	}
	if err := mux.ProcessTask(context.Background(), asynq.NewTask(jobs.TypeProjectExportRender, payload)); err != nil {
		t.Fatalf("ProcessTask() error = %v", err)
	}
	if processor.userID != 42 || processor.matchID != 71 {
		t.Fatalf("processor = %+v", processor)
	}
}

func TestProjectMatchWorkerProcessesEnvelope(t *testing.T) {
	processor := &fakeProjectMatchProcessor{}
	mux := asynq.NewServeMux()
	RegisterWorker(mux, processor)
	payload, err := json.Marshal(jobs.Envelope{IdempotencyKey: "project-match-200-attempt-1", Payload: map[string]any{"user_id": 42, "match_id": 200, "attempt": 1}})
	if err != nil {
		t.Fatal(err)
	}
	if err := mux.ProcessTask(context.Background(), asynq.NewTask(jobs.TypeProjectMatchGenerate, payload)); err != nil {
		t.Fatalf("ProcessTask() error = %v", err)
	}
	if processor.userID != 42 || processor.matchID != 200 || processor.attempt != 1 {
		t.Fatalf("processor = %+v", processor)
	}
}

func TestProjectHeatWorkerProcessesEnvelope(t *testing.T) {
	processor := &fakeProjectMatchProcessor{}
	mux := asynq.NewServeMux()
	RegisterWorker(mux, processor)
	payload, err := json.Marshal(jobs.Envelope{IdempotencyKey: "project-heat-42", Payload: map[string]any{"project_id": 42}})
	if err != nil {
		t.Fatal(err)
	}
	if err := mux.ProcessTask(context.Background(), asynq.NewTask(jobs.TypeProjectHeatAggregate, payload)); err != nil {
		t.Fatalf("ProcessTask() error = %v", err)
	}
	if processor.projectID != 42 {
		t.Fatalf("projectID = %d, want 42", processor.projectID)
	}
}

func TestProjectOperationsWorkerProcessesReindexAndContentJobs(t *testing.T) {
	processor := &fakeProjectMatchProcessor{}
	mux := asynq.NewServeMux()
	RegisterWorker(mux, processor)
	checks := []struct {
		typeName string
		key      string
		payload  map[string]any
	}{
		{jobs.TypeProjectKBReindex, "kb-v2", map[string]any{"reindex_id": 91}},
		{jobs.TypeProjectContentBatch, "content-2026-08-11", map[string]any{"run_id": 92}},
	}
	for _, check := range checks {
		payload, err := json.Marshal(jobs.Envelope{IdempotencyKey: check.key, Payload: check.payload})
		if err != nil {
			t.Fatal(err)
		}
		if err := mux.ProcessTask(context.Background(), asynq.NewTask(check.typeName, payload)); err != nil {
			t.Fatalf("ProcessTask(%s) error = %v", check.typeName, err)
		}
	}
	if processor.reindexID != 91 || processor.contentRunID != 92 {
		t.Fatalf("processor = %+v", processor)
	}
}
