package sandbox

import (
	"errors"
	"time"
)

const (
	StatusDraft     = "draft"
	StatusCompleted = "completed"
)

var (
	ErrServiceNotReady = errors.New("sandbox service is not configured")
	ErrInvalidAIResult = errors.New("invalid sandbox ai result")
	ErrSessionNotFound = errors.New("sandbox session not found")
)

type CreateInput struct {
	UserID      int64    `json:"-"`
	Goal        string   `json:"goal"`
	TargetUsers string   `json:"target_users"`
	Product     string   `json:"product"`
	Roles       []string `json:"roles"`
}

type Session struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"user_id"`
	Goal        string    `json:"goal"`
	TargetUsers string    `json:"target_users"`
	Product     string    `json:"product"`
	Roles       []string  `json:"roles"`
	Status      string    `json:"status"`
	Report      Report    `json:"report,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Report struct {
	Score         int           `json:"score"`
	Summary       string        `json:"summary"`
	Metrics       []Metric      `json:"metrics"`
	RoleSummaries []RoleSummary `json:"role_summaries"`
	Risks         []string      `json:"risks"`
	NextActions   []string      `json:"next_actions"`
}

type Metric struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

type RoleSummary struct {
	Role string `json:"role"`
	View string `json:"view"`
}
