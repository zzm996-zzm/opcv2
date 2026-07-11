package tasks

import (
	"errors"
	"time"
)

const (
	StatusTodo       = "todo"
	StatusInProgress = "in_progress"
	StatusCompleted  = "completed"
	StatusReminder   = "reminder"

	PriorityLow    = "low"
	PriorityMedium = "medium"
	PriorityHigh   = "high"

	ReminderRecurrenceOnce   = "once"
	ReminderRecurrenceDaily  = "daily"
	ReminderRecurrenceWeekly = "weekly"

	SourceAnalysisSession      = "analysis_session"
	SourceProjectMatch         = "project_match"
	SourceSandboxSession       = "sandbox_session"
	SourceCompetitorScan       = "competitor_scan"
	SourceCompetitorMonitoring = "competitor_monitoring"
	SourceEnterpriseDiagnosis  = "enterprise_diagnosis"
	SourceLeadTask             = "lead_task"
	SourceCRMCustomer          = "crm_customer"
	SourceLearningDiagnosis    = "learning_diagnosis"
	SourceGrowthModel          = "growth_model"
)

var (
	ErrServiceNotReady                     = errors.New("tasks service is not configured")
	ErrTaskNotFound                        = errors.New("task not found")
	ErrSubtaskNotFound                     = errors.New("subtask not found")
	ErrReminderNotFound                    = errors.New("task reminder not found")
	ErrInvalidReminderTime                 = errors.New("invalid task reminder time")
	ErrInvalidReminderRecurrence           = errors.New("invalid task reminder recurrence")
	ErrRecurringReminderRequiresMembership = errors.New("recurring task reminder requires membership")
	ErrInvalidGeneratedTasks               = errors.New("invalid generated task plan")
	ErrInvalidTaskSource                   = errors.New("invalid task source")
	ErrInvalidTaskBatch                    = errors.New("invalid task batch")
)

type BatchTaskStatusInput struct {
	IDs    []int64 `json:"ids"`
	Status string  `json:"status"`
}

type BatchTaskIDsInput struct {
	IDs []int64 `json:"ids"`
}

type CreateInput struct {
	UserID      int64      `json:"-"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Assignee    string     `json:"assignee"`
	Project     string     `json:"project"`
	Priority    string     `json:"priority"`
	Tags        []string   `json:"tags"`
	DueAt       *time.Time `json:"due_at,omitempty"`
	Tools       []string   `json:"tools"`
	Learning    string     `json:"learning"`
	SourceType  string     `json:"source_type"`
	SourceID    *int64     `json:"source_id,omitempty"`
	SourceTitle string     `json:"source_title"`
	SourceURL   string     `json:"source_url"`
}

type GenerateTasksInput struct {
	UserID      int64  `json:"-"`
	Goal        string `json:"goal"`
	SourceType  string `json:"source_type"`
	SourceID    *int64 `json:"source_id,omitempty"`
	SourceTitle string `json:"source_title"`
	SourceURL   string `json:"source_url"`
}

type GeneratedTaskDraft struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Project     string   `json:"project"`
	Priority    string   `json:"priority"`
	Tags        []string `json:"tags"`
	DueInDays   int      `json:"due_in_days"`
	Tools       []string `json:"tools"`
	Learning    string   `json:"learning"`
}

type GeneratedTaskPlan struct {
	Tasks []GeneratedTaskDraft `json:"tasks"`
}

type GenerateTasksResult struct {
	Tasks []Task `json:"tasks"`
}

type Subtask struct {
	ID        int64      `json:"id"`
	TaskID    int64      `json:"task_id"`
	UserID    int64      `json:"user_id"`
	Title     string     `json:"title"`
	Assignee  string     `json:"assignee"`
	DueAt     *time.Time `json:"due_at,omitempty"`
	Completed bool       `json:"completed"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type CreateSubtaskInput struct {
	UserID   int64      `json:"-"`
	TaskID   int64      `json:"-"`
	Title    string     `json:"title"`
	Assignee string     `json:"assignee"`
	DueAt    *time.Time `json:"due_at,omitempty"`
}

type SubtaskUpdate struct {
	Title      *string    `json:"title,omitempty"`
	Assignee   *string    `json:"assignee,omitempty"`
	DueAt      *time.Time `json:"due_at,omitempty"`
	ClearDueAt bool       `json:"clear_due_at,omitempty"`
	Completed  *bool      `json:"completed,omitempty"`
}

type TaskReminder struct {
	ID         int64      `json:"id"`
	TaskID     int64      `json:"task_id"`
	UserID     int64      `json:"user_id"`
	RemindAt   time.Time  `json:"remind_at"`
	Recurrence string     `json:"recurrence"`
	SentAt     *time.Time `json:"sent_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

type UpsertTaskReminderInput struct {
	UserID     int64     `json:"-"`
	TaskID     int64     `json:"-"`
	RemindAt   time.Time `json:"remind_at"`
	Recurrence string    `json:"recurrence"`
}

type TaskUpdate struct {
	Title       *string    `json:"title,omitempty"`
	Description *string    `json:"description,omitempty"`
	Assignee    *string    `json:"assignee,omitempty"`
	Project     *string    `json:"project,omitempty"`
	Status      *string    `json:"status,omitempty"`
	Priority    *string    `json:"priority,omitempty"`
	Tags        *[]string  `json:"tags,omitempty"`
	DueAt       *time.Time `json:"due_at,omitempty"`
	ClearDueAt  bool       `json:"clear_due_at,omitempty"`
	Tools       *[]string  `json:"tools,omitempty"`
	Learning    *string    `json:"learning,omitempty"`
}

type ListFilters struct {
	Status   string
	Project  string
	Priority string
	Tag      string
	Query    string
	Limit    int
	Offset   int
}

type TaskPage struct {
	Tasks  []Task `json:"tasks"`
	Total  int    `json:"total"`
	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
}

type Stats struct {
	Total      int `json:"total"`
	Todo       int `json:"todo"`
	InProgress int `json:"in_progress"`
	Completed  int `json:"completed"`
	Reminder   int `json:"reminder"`
	Overdue    int `json:"overdue"`
}

type Task struct {
	ID          int64      `json:"id"`
	UserID      int64      `json:"user_id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Assignee    string     `json:"assignee"`
	Project     string     `json:"project"`
	Status      string     `json:"status"`
	Priority    string     `json:"priority"`
	Tags        []string   `json:"tags"`
	DueAt       *time.Time `json:"due_at,omitempty"`
	Tools       []string   `json:"tools"`
	Learning    string     `json:"learning"`
	SourceType  string     `json:"source_type"`
	SourceID    *int64     `json:"source_id,omitempty"`
	SourceTitle string     `json:"source_title"`
	SourceURL   string     `json:"source_url"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}
