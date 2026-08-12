package projects

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hibiken/asynq"
	"github.com/zzm/opcv2/internal/jobs"
)

type fakeProjectMatchProcessor struct {
	userID  int64
	matchID int64
	attempt int
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
