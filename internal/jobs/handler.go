package jobs

import (
	"context"
	"encoding/json"

	"github.com/hibiken/asynq"
)

type Handler struct {
	Type   string
	Handle func(context.Context, Envelope) error
}

func Register(mux *asynq.ServeMux, handler Handler) {
	mux.HandleFunc(handler.Type, func(ctx context.Context, task *asynq.Task) error {
		var envelope Envelope
		if err := json.Unmarshal(task.Payload(), &envelope); err != nil {
			return err
		}
		return handler.Handle(ctx, envelope)
	})
}
