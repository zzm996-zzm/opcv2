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
)

var (
	ErrServiceNotReady = errors.New("tasks service is not configured")
	ErrTaskNotFound    = errors.New("task not found")
)

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
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}
