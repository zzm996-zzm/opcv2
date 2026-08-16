package tasks

import (
	"errors"
	"time"
)

const (
	StatusTodo       = "todo"
	StatusInProgress = "in_progress"
	StatusReview     = "review"
	StatusBlocked    = "blocked"
	StatusCompleted  = "completed"
	StatusCancelled  = "cancelled"
	StatusReminder   = "reminder"

	PriorityLow    = "low"
	PriorityMedium = "medium"
	PriorityHigh   = "high"

	TaskSortCreated  = "created_at"
	TaskSortUpdated  = "updated_at"
	TaskSortDue      = "due_at"
	TaskSortPriority = "priority"
	TaskSortProgress = "progress"

	TaskGroupStatus   = "status"
	TaskGroupAssignee = "assignee"
	TaskGroupProject  = "project"
	TaskGroupPriority = "priority"
	TaskGroupSource   = "source"

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
	SourceCopilotMessage       = "copilot_message"
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
	ErrInvalidTaskStatusTransition         = errors.New("invalid task status transition")
	ErrInvalidTaskProgress                 = errors.New("invalid task progress")
	ErrTaskVersionConflict                 = errors.New("task version conflict")
	ErrTaskAIDraftNotFound                 = errors.New("task ai draft not found")
	ErrTaskAIDraftAlreadyAdopted           = errors.New("task ai draft already adopted")
	ErrInvalidTaskAIDraft                  = errors.New("invalid task ai draft")
	ErrInvalidTaskCalendarRange            = errors.New("invalid task calendar range")
	ErrTaskCommentNotFound                 = errors.New("task comment not found")
	ErrInvalidTaskComment                  = errors.New("invalid task comment")
	ErrInvalidTaskViewPreference           = errors.New("invalid task view preference")
	ErrTaskAttachmentNotFound              = errors.New("task attachment not found")
	ErrTaskAttachmentTooLarge              = errors.New("task attachment is too large")
	ErrTaskAttachmentCountExceeded         = errors.New("task attachment count exceeded")
	ErrInvalidTaskAttachment               = errors.New("invalid task attachment")
	ErrUnsupportedTaskAttachmentType       = errors.New("unsupported task attachment type")
	ErrTaskAttachmentMIMEMismatch          = errors.New("task attachment MIME does not match content")
	ErrUnsafeTaskAttachment                = errors.New("unsafe task attachment")
	ErrTaskAttachmentSignatureInvalid      = errors.New("invalid task attachment signature")
	ErrTaskAttachmentSignatureUnavailable  = errors.New("task attachment signing is not configured")
)

const (
	TaskAIDraftStatusDraft   = "draft"
	TaskAIDraftStatusAdopted = "adopted"
)

type BatchTaskStatusInput struct {
	IDs    []int64 `json:"ids"`
	Status string  `json:"status"`
}

type BatchTaskIDsInput struct {
	IDs []int64 `json:"ids"`
}

type BatchTaskUpdateInput struct {
	IDs        []int64    `json:"ids"`
	Status     *string    `json:"status,omitempty"`
	Assignee   *string    `json:"assignee,omitempty"`
	Priority   *string    `json:"priority,omitempty"`
	DueAt      *time.Time `json:"due_at,omitempty"`
	ClearDueAt bool       `json:"clear_due_at,omitempty"`
	Tags       *[]string  `json:"tags,omitempty"`
}

type BatchTaskUpdateFailure struct {
	ID     int64  `json:"id"`
	Code   string `json:"code"`
	Detail string `json:"detail,omitempty"`
}

type BatchTaskUpdateResult struct {
	Updated int                      `json:"updated"`
	Failed  []BatchTaskUpdateFailure `json:"failed"`
}

type CreateInput struct {
	UserID         int64      `json:"-"`
	Title          string     `json:"title"`
	Description    string     `json:"description"`
	Assignee       string     `json:"assignee"`
	Project        string     `json:"project"`
	Priority       string     `json:"priority"`
	Tags           []string   `json:"tags"`
	DueAt          *time.Time `json:"due_at,omitempty"`
	Tools          []string   `json:"tools"`
	Learning       string     `json:"learning"`
	SourceType     string     `json:"source_type"`
	SourceID       *int64     `json:"source_id,omitempty"`
	SourceTitle    string     `json:"source_title"`
	SourceURL      string     `json:"source_url"`
	IdempotencyKey string     `json:"-"`
}

type BatchCreateInput struct {
	UserID int64         `json:"-"`
	Tasks  []CreateInput `json:"tasks"`
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
	Draft TaskAIDraft `json:"draft"`
	Tasks []Task      `json:"tasks"`
}

type TaskAIDraft struct {
	ID             int64      `json:"id"`
	UserID         int64      `json:"user_id"`
	Goal           string     `json:"goal"`
	SourceType     string     `json:"source_type"`
	SourceID       *int64     `json:"source_id,omitempty"`
	SourceTitle    string     `json:"source_title"`
	SourceURL      string     `json:"source_url"`
	Tasks          []Task     `json:"tasks"`
	Status         string     `json:"status"`
	AdoptedTaskIDs []int64    `json:"adopted_task_ids,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	AdoptedAt      *time.Time `json:"adopted_at,omitempty"`
}

type AIDraftTaskInput struct {
	DraftIndex  int        `json:"draft_index"`
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

type AdoptTaskAIDraftInput struct {
	Tasks []AIDraftTaskInput `json:"tasks"`
}

type Subtask struct {
	ID              int64      `json:"id"`
	TaskID          int64      `json:"task_id"`
	UserID          int64      `json:"user_id"`
	ParentSubtaskID *int64     `json:"parent_subtask_id,omitempty"`
	Title           string     `json:"title"`
	Assignee        string     `json:"assignee"`
	DueAt           *time.Time `json:"due_at,omitempty"`
	Completed       bool       `json:"completed"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type CreateSubtaskInput struct {
	UserID          int64      `json:"-"`
	TaskID          int64      `json:"-"`
	ParentSubtaskID *int64     `json:"parent_subtask_id,omitempty"`
	Title           string     `json:"title"`
	Assignee        string     `json:"assignee"`
	DueAt           *time.Time `json:"due_at,omitempty"`
}

type SubtaskUpdate struct {
	Title      *string    `json:"title,omitempty"`
	Assignee   *string    `json:"assignee,omitempty"`
	DueAt      *time.Time `json:"due_at,omitempty"`
	ClearDueAt bool       `json:"clear_due_at,omitempty"`
	Completed  *bool      `json:"completed,omitempty"`
}

const TaskViewList = "list"

var (
	taskListDefaultColumns = []string{"title", "project", "assignee", "due_at", "priority", "status"}
	taskListAllowedColumns = []string{"title", "project", "assignee", "due_at", "priority", "status", "tags", "progress", "source", "created_at", "updated_at"}
)

type TaskViewPreference struct {
	View      string    `json:"view"`
	Columns   []string  `json:"columns"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UpdateTaskViewPreferenceInput struct {
	UserID  int64    `json:"-"`
	View    string   `json:"view"`
	Columns []string `json:"columns"`
}

func defaultTaskListColumns() []string {
	return append([]string(nil), taskListDefaultColumns...)
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
	Progress    *int       `json:"progress,omitempty"`
	Version     *int64     `json:"version,omitempty"`
}

type ListFilters struct {
	Status   string
	Project  string
	Priority string
	Tag      string
	Query    string
	Sort     string
	Group    string
	Limit    int
	Offset   int
}

type CalendarFilters struct {
	Status   string
	Project  string
	Priority string
	Tag      string
	Query    string
	From     time.Time
	To       time.Time
}

type CalendarPage struct {
	Tasks       []Task    `json:"tasks"`
	Unscheduled []Task    `json:"unscheduled"`
	Total       int       `json:"total"`
	From        time.Time `json:"from"`
	To          time.Time `json:"to"`
}

type TaskPage struct {
	Tasks  []Task `json:"tasks"`
	Total  int    `json:"total"`
	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
	Sort   string `json:"sort"`
	Group  string `json:"group,omitempty"`
}

type Stats struct {
	Total      int `json:"total"`
	Todo       int `json:"todo"`
	InProgress int `json:"in_progress"`
	Completed  int `json:"completed"`
	Reminder   int `json:"reminder"`
	DueSoon    int `json:"due_soon"`
	TimedOut   int `json:"timed_out"`
	Overdue    int `json:"overdue"`
}

type TaskActivity struct {
	ID         int64          `json:"id"`
	TaskID     int64          `json:"task_id"`
	UserID     int64          `json:"user_id"`
	Action     string         `json:"action"`
	BeforeData map[string]any `json:"before"`
	AfterData  map[string]any `json:"after"`
	Metadata   map[string]any `json:"metadata"`
	CreatedAt  time.Time      `json:"created_at"`
}

type TaskActivityPage struct {
	Activities []TaskActivity `json:"activities"`
	Total      int            `json:"total"`
	Limit      int            `json:"limit"`
	Offset     int            `json:"offset"`
}

type TaskComment struct {
	ID              int64     `json:"id"`
	TaskID          int64     `json:"task_id"`
	UserID          int64     `json:"user_id"`
	ParentCommentID *int64    `json:"parent_comment_id,omitempty"`
	Content         string    `json:"content"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type TaskAttachment struct {
	ID         int64      `json:"id"`
	TaskID     int64      `json:"task_id"`
	CommentID  *int64     `json:"comment_id,omitempty"`
	UserID     int64      `json:"-"`
	Name       string     `json:"name"`
	MIMEType   string     `json:"mime_type"`
	SizeBytes  int64      `json:"size_bytes"`
	SHA256     string     `json:"sha256"`
	CreatedAt  time.Time  `json:"created_at"`
	DeletedAt  *time.Time `json:"-"`
	StorageKey string     `json:"-"`
}

type TaskAttachmentDownload struct {
	Attachment TaskAttachment `json:"attachment"`
	URL        string         `json:"url"`
	ExpiresAt  time.Time      `json:"expires_at"`
}

type CreateTaskAttachmentInput struct {
	UserID    int64
	TaskID    int64
	CommentID *int64
	Name      string
	MIMEType  string
	Data      []byte
}

type TaskCommentPage struct {
	Comments []TaskComment `json:"comments"`
	Total    int           `json:"total"`
	Limit    int           `json:"limit"`
	Offset   int           `json:"offset"`
}

type CreateTaskCommentInput struct {
	UserID          int64  `json:"-"`
	TaskID          int64  `json:"-"`
	ParentCommentID *int64 `json:"parent_comment_id,omitempty"`
	Content         string `json:"content"`
}

type UpdateTaskCommentInput struct {
	Content string `json:"content"`
}

type Task struct {
	ID             int64      `json:"id"`
	UserID         int64      `json:"user_id"`
	Title          string     `json:"title"`
	Description    string     `json:"description"`
	Assignee       string     `json:"assignee"`
	Project        string     `json:"project"`
	Status         string     `json:"status"`
	Priority       string     `json:"priority"`
	Tags           []string   `json:"tags"`
	DueAt          *time.Time `json:"due_at,omitempty"`
	Tools          []string   `json:"tools"`
	Learning       string     `json:"learning"`
	Progress       int        `json:"progress"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
	Version        int64      `json:"version"`
	SourceType     string     `json:"source_type"`
	SourceID       *int64     `json:"source_id,omitempty"`
	SourceTitle    string     `json:"source_title"`
	SourceURL      string     `json:"source_url"`
	IdempotencyKey string     `json:"-"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}
