package leads

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/hibiken/asynq"
	"github.com/zzm/opcv2/internal/jobs"
)

var ErrInvalidJobPayload = errors.New("invalid lead job payload")

type Processor interface {
	ProcessTask(ctx context.Context, taskID int64) error
}

type WorkerHandler struct {
	processor Processor
}

func NewWorkerHandler(processor Processor) *WorkerHandler {
	return &WorkerHandler{processor: processor}
}

func (h *WorkerHandler) Handle(ctx context.Context, payload []byte) error {
	var envelope jobs.Envelope
	if err := json.Unmarshal(payload, &envelope); err != nil {
		return err
	}
	taskID, ok := envelope.Payload["task_id"].(float64)
	if !ok || taskID <= 0 {
		return ErrInvalidJobPayload
	}
	return h.processor.ProcessTask(ctx, int64(taskID))
}

func RegisterWorker(mux *asynq.ServeMux, processor Processor) {
	handler := NewWorkerHandler(processor)
	jobs.Register(mux, jobs.Handler{
		Type: jobs.TypeLeadSearch,
		Handle: func(ctx context.Context, envelope jobs.Envelope) error {
			payload, err := json.Marshal(envelope)
			if err != nil {
				return err
			}
			return handler.Handle(ctx, payload)
		},
	})
}
