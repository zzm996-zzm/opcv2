package leads

import (
	"errors"
	"time"
)

const (
	StatusQueued    = "queued"
	StatusRunning   = "running"
	StatusSucceeded = "succeeded"
	StatusFailed    = "failed"
	StatusCancelled = "cancelled"
	StatusRefunded  = "refunded"
)

const (
	ErrorProviderUnavailable   = "provider_unavailable"
	ErrorProviderQuotaExceeded = "provider_quota_exceeded"
)

const defaultCreditCost = 1

var (
	ErrTaskNotFound          = errors.New("lead task not found")
	ErrServiceNotReady       = errors.New("leads service is not configured")
	ErrInvalidTaskInput      = errors.New("invalid lead task input")
	ErrProviderUnavailable   = errors.New("lead provider unavailable")
	ErrProviderQuotaExceeded = errors.New("lead provider quota exceeded")
)

type CreateTaskInput struct {
	UserID         int64  `json:"-"`
	Query          string `json:"query"`
	IdempotencyKey string `json:"idempotency_key"`
}

type Task struct {
	ID             int64     `json:"id"`
	UserID         int64     `json:"user_id"`
	Query          string    `json:"query"`
	Status         string    `json:"status"`
	IdempotencyKey string    `json:"idempotency_key"`
	CreditCost     int       `json:"credit_cost"`
	ErrorCode      string    `json:"error_code,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type TaskDetail struct {
	Task            Task   `json:"task"`
	ProgressPercent int    `json:"progress_percent"`
	Message         string `json:"message"`
	ResultsCount    int    `json:"results_count"`
}

type Lead struct {
	Name     string     `json:"name"`
	Phone    string     `json:"phone,omitempty"`
	Email    string     `json:"email,omitempty"`
	Website  string     `json:"website,omitempty"`
	Evidence []Evidence `json:"evidence,omitempty"`
}

type LeadResult struct {
	ID        int64      `json:"id"`
	TaskID    int64      `json:"task_id"`
	Name      string     `json:"name"`
	Phone     string     `json:"phone,omitempty"`
	Email     string     `json:"email,omitempty"`
	Website   string     `json:"website,omitempty"`
	Evidence  []Evidence `json:"evidence,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

type Evidence struct {
	Type  string `json:"type"`
	Title string `json:"title"`
	URL   string `json:"url"`
}

type SearchInput struct {
	Query string
}
