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

type WorkerHandler struct {
	processor Processor
}

func NewWorkerHandler(processor Processor) *WorkerHandler {
	return &WorkerHandler{processor: processor}
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

func RegisterWorker(mux *asynq.ServeMux, processor Processor) {
	handler := NewWorkerHandler(processor)
	jobs.Register(mux, jobs.Handler{Type: jobs.TypeSandboxRun, Handle: handler.Handle})
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
