package geo

import "time"

const (
	AnalysisRequestStatusQueued    = "queued"
	AnalysisRequestStatusRunning   = "running"
	AnalysisRequestStatusSucceeded = "succeeded"
	AnalysisRequestStatusFailed    = "failed"
)

type Metric struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	Value string `json:"value"`
}

type AnalysisRequestInput struct {
	Target string `json:"target"`
}

type AnalysisRequest struct {
	ID           int64     `json:"id"`
	UserID       int64     `json:"user_id"`
	Target       string    `json:"target"`
	Status       string    `json:"status"`
	ErrorMessage string    `json:"error_message,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type EngineCoverage struct {
	Name            string `json:"name"`
	CoveragePercent int    `json:"coverage_percent"`
	Status          string `json:"status,omitempty"`
}

type LeadSignal struct {
	Title  string `json:"title"`
	Detail string `json:"detail"`
}

type KeywordOpportunity struct {
	ID       int64  `json:"id"`
	Query    string `json:"query"`
	Intent   string `json:"intent,omitempty"`
	Coverage string `json:"coverage,omitempty"`
	Score    int    `json:"score"`
	Action   string `json:"action,omitempty"`
}

type ContentTask struct {
	ID       int64  `json:"id"`
	Type     string `json:"type,omitempty"`
	Title    string `json:"title"`
	Priority string `json:"priority,omitempty"`
	DueAt    string `json:"due_at,omitempty"`
}

type Overview struct {
	Stats        []Metric             `json:"stats"`
	Engines      []EngineCoverage     `json:"engines"`
	LeadSignals  []LeadSignal         `json:"lead_signals"`
	Keywords     []KeywordOpportunity `json:"keywords"`
	ContentTasks []ContentTask        `json:"content_tasks"`
}
