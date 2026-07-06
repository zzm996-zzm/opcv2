package jobs

import (
	"errors"
	"time"
)

const (
	TypeLeadSearch     = "leads.search"
	TypeGeoAnalysis    = "geo.analysis"
	TypeCompetitorScan = "competitor.scan"
)

var (
	ErrUnknownType           = errors.New("unknown job type")
	ErrMissingIdempotencyKey = errors.New("missing job idempotency key")
)

type Registry map[string]bool

type Job struct {
	Type           string
	IdempotencyKey string
	Payload        map[string]any
	MaxRetry       int
	Timeout        time.Duration
}

type Envelope struct {
	IdempotencyKey string         `json:"idempotency_key"`
	Payload        map[string]any `json:"payload"`
}
