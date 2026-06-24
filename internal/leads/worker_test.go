package leads

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hibiken/asynq"
	"github.com/zzm/opcv2/internal/jobs"
)

type fakeProcessor struct {
	taskID int64
	err    error
}

func (p *fakeProcessor) ProcessTask(_ context.Context, taskID int64) error {
	p.taskID = taskID
	return p.err
}

func TestWorkerProcessesLeadSearchEnvelope(t *testing.T) {
	processor := &fakeProcessor{}
	handler := NewWorkerHandler(processor)
	payload, err := json.Marshal(jobs.Envelope{
		IdempotencyKey: "lead-task-42",
		Payload:        map[string]any{"task_id": float64(99)},
	})
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	err = handler.Handle(context.Background(), payload)
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if processor.taskID != 99 {
		t.Fatalf("taskID = %d, want 99", processor.taskID)
	}
}

func TestWorkerRegistersLeadSearchHandler(t *testing.T) {
	processor := &fakeProcessor{}
	mux := asynq.NewServeMux()
	RegisterWorker(mux, processor)
	payload, err := json.Marshal(jobs.Envelope{
		IdempotencyKey: "lead-task-42",
		Payload:        map[string]any{"task_id": float64(99)},
	})
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	err = mux.ProcessTask(context.Background(), asynq.NewTask(jobs.TypeLeadSearch, payload))
	if err != nil {
		t.Fatalf("ProcessTask() error = %v", err)
	}
	if processor.taskID != 99 {
		t.Fatalf("taskID = %d, want 99", processor.taskID)
	}
}
