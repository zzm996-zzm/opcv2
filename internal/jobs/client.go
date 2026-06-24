package jobs

import (
	"context"
	"encoding/json"
	"time"

	"github.com/hibiken/asynq"
)

type Enqueuer interface {
	EnqueueContext(ctx context.Context, task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error)
}

type Client struct {
	enqueuer Enqueuer
	registry Registry
}

func NewClient(enqueuer Enqueuer, registry Registry) *Client {
	return &Client{enqueuer: enqueuer, registry: registry}
}

func (c *Client) Enqueue(ctx context.Context, job Job) error {
	if !c.registry[job.Type] {
		return ErrUnknownType
	}
	if job.IdempotencyKey == "" {
		return ErrMissingIdempotencyKey
	}
	payload, err := json.Marshal(Envelope{
		IdempotencyKey: job.IdempotencyKey,
		Payload:        job.Payload,
	})
	if err != nil {
		return err
	}
	opts := enqueueOptions(job)
	_, err = c.enqueuer.EnqueueContext(ctx, asynq.NewTask(job.Type, payload), opts...)
	return err
}

func enqueueOptions(job Job) []asynq.Option {
	maxRetry := job.MaxRetry
	if maxRetry <= 0 {
		maxRetry = 3
	}
	timeout := job.Timeout
	if timeout <= 0 {
		timeout = 5 * time.Minute
	}
	return []asynq.Option{
		asynq.TaskID(job.IdempotencyKey),
		asynq.MaxRetry(maxRetry),
		asynq.Timeout(timeout),
	}
}
