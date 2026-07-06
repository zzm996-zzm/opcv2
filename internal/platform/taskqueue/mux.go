package taskqueue

import (
	"github.com/hibiken/asynq"
	"github.com/zzm/opcv2/internal/jobs"
)

func DefaultRegistry() jobs.Registry {
	return jobs.Registry{
		jobs.TypeLeadSearch:     true,
		jobs.TypeGeoAnalysis:    true,
		jobs.TypeCompetitorScan: true,
	}
}

func NewClient(redisAddr string) *jobs.Client {
	return jobs.NewClient(asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr}), DefaultRegistry())
}

func NewMux() *asynq.ServeMux {
	return asynq.NewServeMux()
}

func NewServer(redisAddr string) *asynq.Server {
	return asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddr},
		asynq.Config{Concurrency: 10},
	)
}
