package taskqueue

import "github.com/hibiken/asynq"

func NewMux() *asynq.ServeMux {
	return asynq.NewServeMux()
}

func NewServer(redisAddr string) *asynq.Server {
	return asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddr},
		asynq.Config{Concurrency: 10},
	)
}
