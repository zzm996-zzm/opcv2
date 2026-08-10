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

type fakeV2Processor struct {
	*fakeProcessor
	runID    int64
	revision int
	roleCode string
}

func (p *fakeV2Processor) ProcessV2Run(_ context.Context, userID, runID int64, revision int) error {
	p.userID, p.runID, p.revision = userID, runID, revision
	return nil
}

func (p *fakeV2Processor) ProcessV2RoleRetry(_ context.Context, userID, runID int64, revision int, roleCode string) error {
	p.userID, p.runID, p.revision, p.roleCode = userID, runID, revision, roleCode
	return nil
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

func TestWorkerDispatchesV2RunAndRoleRetry(t *testing.T) {
	processor := &fakeV2Processor{fakeProcessor: &fakeProcessor{}}
	handler := NewWorkerHandler(processor)
	if err := handler.HandleV2(context.Background(), jobs.Envelope{Payload: map[string]any{
		"user_id": int64(42), "run_id": int64(99), "revision": 2,
	}}); err != nil {
		t.Fatalf("HandleV2 run error = %v", err)
	}
	if processor.runID != 99 || processor.revision != 2 || processor.roleCode != "" {
		t.Fatalf("run dispatch = %+v", processor)
	}
	if err := handler.HandleV2(context.Background(), jobs.Envelope{Payload: map[string]any{
		"user_id": int64(42), "run_id": int64(99), "revision": 3, "role_code": "skeptic",
	}}); err != nil {
		t.Fatalf("HandleV2 retry error = %v", err)
	}
	if processor.revision != 3 || processor.roleCode != "skeptic" {
		t.Fatalf("retry dispatch = %+v", processor)
	}
}
