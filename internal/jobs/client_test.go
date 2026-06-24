package jobs

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/hibiken/asynq"
)

type fakeEnqueuer struct {
	task *asynq.Task
	opts []asynq.Option
}

func (e *fakeEnqueuer) EnqueueContext(_ context.Context, task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	e.task = task
	e.opts = opts
	return &asynq.TaskInfo{ID: "task-id"}, nil
}

func TestClientEnqueueIncludesIdempotencyKeyInPayload(t *testing.T) {
	enqueuer := &fakeEnqueuer{}
	client := NewClient(enqueuer, Registry{TypeLeadSearch: true})

	err := client.Enqueue(context.Background(), Job{
		Type:           TypeLeadSearch,
		IdempotencyKey: "lead-task-42",
		Payload:        map[string]any{"task_id": float64(42)},
		MaxRetry:       3,
		Timeout:        2 * time.Minute,
	})
	if err != nil {
		t.Fatalf("Enqueue() error = %v", err)
	}
	if enqueuer.task == nil || enqueuer.task.Type() != TypeLeadSearch {
		t.Fatalf("task = %+v", enqueuer.task)
	}
	var payload Envelope
	if err := json.Unmarshal(enqueuer.task.Payload(), &payload); err != nil {
		t.Fatalf("payload is not envelope JSON: %v", err)
	}
	if payload.IdempotencyKey != "lead-task-42" {
		t.Fatalf("idempotency key = %q", payload.IdempotencyKey)
	}
	if payload.Payload["task_id"] != float64(42) {
		t.Fatalf("payload = %+v", payload.Payload)
	}
	if len(enqueuer.opts) == 0 {
		t.Fatal("enqueue options were not set")
	}
}

func TestClientRejectsUnknownJobType(t *testing.T) {
	client := NewClient(&fakeEnqueuer{}, Registry{TypeLeadSearch: true})

	err := client.Enqueue(context.Background(), Job{
		Type:           "unknown",
		IdempotencyKey: "key",
		Payload:        map[string]any{},
	})
	if !errors.Is(err, ErrUnknownType) {
		t.Fatalf("Enqueue() error = %v, want ErrUnknownType", err)
	}
}

func TestClientRejectsMissingIdempotencyKey(t *testing.T) {
	client := NewClient(&fakeEnqueuer{}, Registry{TypeLeadSearch: true})

	err := client.Enqueue(context.Background(), Job{
		Type:    TypeLeadSearch,
		Payload: map[string]any{},
	})
	if !errors.Is(err, ErrMissingIdempotencyKey) {
		t.Fatalf("Enqueue() error = %v, want ErrMissingIdempotencyKey", err)
	}
}

func TestRegisterAddsHandlerForKnownType(t *testing.T) {
	mux := asynq.NewServeMux()
	called := false

	Register(mux, Handler{
		Type: TypeLeadSearch,
		Handle: func(context.Context, Envelope) error {
			called = true
			return nil
		},
	})

	payload, err := json.Marshal(Envelope{IdempotencyKey: "lead-task-42", Payload: map[string]any{"task_id": float64(42)}})
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	err = mux.ProcessTask(context.Background(), asynq.NewTask(TypeLeadSearch, payload))
	if err != nil {
		t.Fatalf("ProcessTask() error = %v", err)
	}
	if !called {
		t.Fatal("handler was not called")
	}
}
