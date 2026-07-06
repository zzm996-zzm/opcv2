package competitor

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/hibiken/asynq"
	"github.com/zzm/opcv2/internal/jobs"
)

var ErrInvalidJobPayload = errors.New("invalid competitor job payload")

type Processor interface {
	ProcessScan(ctx context.Context, scanID int64) error
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
	scanID, ok := envelope.Payload["scan_id"].(float64)
	if !ok || scanID <= 0 {
		return ErrInvalidJobPayload
	}
	return h.processor.ProcessScan(ctx, int64(scanID))
}

func RegisterWorker(mux *asynq.ServeMux, processor Processor) {
	handler := NewWorkerHandler(processor)
	jobs.Register(mux, jobs.Handler{
		Type: jobs.TypeCompetitorScan,
		Handle: func(ctx context.Context, envelope jobs.Envelope) error {
			payload, err := json.Marshal(envelope)
			if err != nil {
				return err
			}
			return handler.Handle(ctx, payload)
		},
	})
}
