package competitor

import (
	"encoding/json"
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
	ErrServiceNotReady          = errors.New("competitor service is not configured")
	ErrScanNotFound             = errors.New("competitor scan not found")
	ErrInvalidWatchItem         = errors.New("invalid competitor watch item")
	ErrWatchItemNotFound        = errors.New("competitor watch item not found")
	ErrAdminRequired            = errors.New("admin role required")
	ErrInvalidScriptAccount     = errors.New("invalid competitor script account")
	ErrScriptAccountNotFound    = errors.New("competitor script account not found")
	ErrScriptAccountUnavailable = errors.New("competitor script account unavailable")
)

const (
	ScriptAccountAvailable = "available"
	ScriptAccountInUse     = "in_use"
	ScriptAccountCooldown  = "cooldown"
	ScriptAccountDisabled  = "disabled"
)

type CreateScanInput struct {
	UserID  int64    `json:"-"`
	Targets []string `json:"targets"`
	Focus   string   `json:"focus"`
}

type CreateWatchItemInput struct {
	UserID   int64    `json:"-"`
	Name     string   `json:"name"`
	Category string   `json:"category"`
	Channels []string `json:"channels"`
}

type AddScanCompetitorToWatchlistInput struct {
	CompetitorName string `json:"competitor_name"`
}

type ScanResult struct {
	Competitors     []Competitor
	Conclusions     []Conclusion
	EvidenceSources []EvidenceSource
	RawSnapshots    []RawSnapshot
}

type RawSnapshot struct {
	Platform   string          `json:"-"`
	Payload    json.RawMessage `json:"-"`
	ObjectKey  string          `json:"-"`
	CapturedAt time.Time       `json:"-"`
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
	SourceType          string    `json:"source_type"`
	Platform            string    `json:"platform,omitempty"`
	Title               string    `json:"title"`
	URL                 string    `json:"url"`
	Summary             string    `json:"summary"`
	ScreenshotObjectKey string    `json:"screenshot_object_key,omitempty"`
	CapturedAt          time.Time `json:"captured_at"`
}

type WatchItem struct {
	ID         int64     `json:"id"`
	UserID     int64     `json:"-"`
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

type ScriptAccount struct {
	ID             int64      `json:"id"`
	Platform       string     `json:"platform"`
	AccountLabel   string     `json:"account_label"`
	CredentialRef  string     `json:"-"`
	HasCredential  bool       `json:"has_credential"`
	Status         string     `json:"status"`
	CooldownUntil  *time.Time `json:"cooldown_until,omitempty"`
	FailureCount   int        `json:"failure_count"`
	MaxRunsPerHour int        `json:"max_runs_per_hour"`
	LastUsedAt     *time.Time `json:"last_used_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type ScriptAccountInput struct {
	ID             int64  `json:"-"`
	AdminUserID    int64  `json:"-"`
	Platform       string `json:"platform"`
	AccountLabel   string `json:"account_label"`
	CredentialRef  string `json:"credential_ref"`
	Status         string `json:"status"`
	MaxRunsPerHour int    `json:"max_runs_per_hour"`
}
