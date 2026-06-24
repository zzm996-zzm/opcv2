package ai

import "time"

const (
	StatusPending   = "pending"
	StatusCompleted = "completed"
	StatusFailed    = "failed"
)

type Run struct {
	ID            int64
	UserID        int64
	Feature       string
	PromptVersion string
	Provider      string
	Model         string
	Status        string
	Request       []byte
	Response      []byte
	ErrorCode     string
	ErrorMessage  string
	InputTokens   int
	OutputTokens  int
	LatencyMS     int
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type RunResult struct {
	Response     []byte
	InputTokens  int
	OutputTokens int
	LatencyMS    int
}

type RunFailure struct {
	Code      string
	Message   string
	LatencyMS int
}
