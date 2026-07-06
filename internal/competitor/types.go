package competitor

import (
	"errors"
	"time"
)

const (
	StatusQueued    = "queued"
	StatusRunning   = "running"
	StatusSucceeded = "succeeded"
	StatusFailed    = "failed"
	StatusCompleted = StatusSucceeded
)

var (
	ErrServiceNotReady = errors.New("competitor service is not configured")
	ErrScanNotFound    = errors.New("competitor scan not found")
)

type CreateScanInput struct {
	UserID  int64    `json:"-"`
	Targets []string `json:"targets"`
	Focus   string   `json:"focus"`
}

type ScanResult struct {
	Competitors     []Competitor
	Conclusions     []Conclusion
	EvidenceSources []EvidenceSource
}

type Scan struct {
	ID              int64            `json:"id"`
	UserID          int64            `json:"user_id"`
	Targets         []string         `json:"targets"`
	Focus           string           `json:"focus"`
	Status          string           `json:"status"`
	ProgressPercent int              `json:"progress_percent"`
	CurrentStep     string           `json:"current_step"`
	ErrorMessage    string           `json:"error_message,omitempty"`
	Competitors     []Competitor     `json:"competitors"`
	Conclusions     []Conclusion     `json:"conclusions"`
	EvidenceSources []EvidenceSource `json:"evidence_sources"`
	CreatedAt       time.Time        `json:"created_at"`
	UpdatedAt       time.Time        `json:"updated_at"`
}

type Competitor struct {
	Name     string   `json:"name"`
	Category string   `json:"category"`
	Score    int      `json:"score"`
	Signal   string   `json:"signal"`
	Risk     string   `json:"risk"`
	Tags     []string `json:"tags"`
}

type Conclusion struct {
	Title  string `json:"title"`
	Detail string `json:"detail"`
}

type EvidenceSource struct {
	SourceType string    `json:"source_type"`
	Title      string    `json:"title"`
	URL        string    `json:"url"`
	Summary    string    `json:"summary"`
	CapturedAt time.Time `json:"captured_at"`
}

type WatchItem struct {
	Name       string    `json:"name"`
	Category   string    `json:"category"`
	Status     string    `json:"status"`
	Threat     string    `json:"threat"`
	LastSeenAt time.Time `json:"last_seen_at"`
	Channels   []string  `json:"channels"`
	Signal     string    `json:"signal"`
}

type Event struct {
	OccurredAt time.Time `json:"occurred_at"`
	Company    string    `json:"company"`
	Title      string    `json:"title"`
	Detail     string    `json:"detail"`
	Level      string    `json:"level"`
}

type MonitoringSnapshot struct {
	Watchlist []WatchItem `json:"watchlist"`
	Events    []Event     `json:"events"`
}
