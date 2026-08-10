package sandbox

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/hibiken/asynq"
	"github.com/zzm/opcv2/internal/jobs"
)

var ErrInvalidJobPayload = errors.New("invalid sandbox job payload")

type Processor interface {
	ProcessSession(ctx context.Context, userID, sessionID int64, attempt int) error
}

type V2Processor interface {
	ProcessV2Run(ctx context.Context, userID, runID int64, revision int) error
}

type WorkerHandler struct {
	processor   Processor
	v2Processor V2Processor
}

func NewWorkerHandler(processor Processor) *WorkerHandler {
	handler := &WorkerHandler{processor: processor}
	if v2Processor, ok := processor.(V2Processor); ok {
		handler.v2Processor = v2Processor
	}
	return handler
}

func (h *WorkerHandler) Handle(ctx context.Context, envelope jobs.Envelope) error {
	userID, userOK := numberPayload(envelope.Payload["user_id"])
	sessionID, sessionOK := numberPayload(envelope.Payload["session_id"])
	attempt, attemptOK := numberPayload(envelope.Payload["attempt"])
	if !userOK || !sessionOK || !attemptOK || userID <= 0 || sessionID <= 0 || attempt <= 0 {
		return ErrInvalidJobPayload
	}
	return h.processor.ProcessSession(ctx, userID, sessionID, int(attempt))
}

func (h *WorkerHandler) HandleV2(ctx context.Context, envelope jobs.Envelope) error {
	if h.v2Processor == nil {
		return ErrServiceNotReady
	}
	userID, userOK := numberPayload(envelope.Payload["user_id"])
	runID, runOK := numberPayload(envelope.Payload["run_id"])
	revision, revisionOK := numberPayload(envelope.Payload["revision"])
	if !userOK || !runOK || !revisionOK || userID <= 0 || runID <= 0 || revision <= 0 {
		return ErrInvalidJobPayload
	}
	return h.v2Processor.ProcessV2Run(ctx, userID, runID, int(revision))
}

func RegisterWorker(mux *asynq.ServeMux, processor Processor) {
	handler := NewWorkerHandler(processor)
	jobs.Register(mux, jobs.Handler{Type: jobs.TypeSandboxRun, Handle: handler.Handle})
	jobs.Register(mux, jobs.Handler{Type: jobs.TypeSandboxV2Run, Handle: handler.HandleV2})
}

func numberPayload(value any) (int64, bool) {
	switch typed := value.(type) {
	case float64:
		return int64(typed), typed > 0 && typed == float64(int64(typed))
	case int:
		return int64(typed), typed > 0
	case int64:
		return typed, typed > 0
	case json.Number:
		parsed, err := typed.Int64()
		return parsed, err == nil && parsed > 0
	default:
		return 0, false
	}
}
