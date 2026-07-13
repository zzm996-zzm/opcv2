package sandbox

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hibiken/asynq"
	"github.com/zzm/opcv2/internal/jobs"
)

type fakeProcessor struct {
	userID    int64
	sessionID int64
	attempt   int
}

func (p *fakeProcessor) ProcessSession(_ context.Context, userID, sessionID int64, attempt int) error {
	p.userID, p.sessionID, p.attempt = userID, sessionID, attempt
	return nil
}

func TestWorkerProcessesSandboxRun(t *testing.T) {
	processor := &fakeProcessor{}
	handler := NewWorkerHandler(processor)
	err := handler.Handle(context.Background(), jobs.Envelope{Payload: map[string]any{"user_id": float64(42), "session_id": float64(99), "attempt": float64(2)}})
	if err != nil || processor.userID != 42 || processor.sessionID != 99 || processor.attempt != 2 {
		t.Fatalf("processed = %d/%d/%d err=%v", processor.userID, processor.sessionID, processor.attempt, err)
	}
}

func TestWorkerRegistersSandboxRunHandler(t *testing.T) {
	processor := &fakeProcessor{}
	mux := asynq.NewServeMux()
	RegisterWorker(mux, processor)
	payload, err := json.Marshal(jobs.Envelope{IdempotencyKey: "sandbox-99-1", Payload: map[string]any{"user_id": int64(42), "session_id": int64(99), "attempt": 1}})
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	task := asynq.NewTask(jobs.TypeSandboxRun, payload)
	if err := mux.ProcessTask(context.Background(), task); err != nil {
		t.Fatalf("ProcessTask() error = %v", err)
	}
	if processor.sessionID != 99 {
		t.Fatalf("processor = %+v", processor)
	}
}
