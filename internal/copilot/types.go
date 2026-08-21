package copilot

import (
	"encoding/json"
	"errors"
	"time"
)

const (
	ModeChat    = "chat"
	ModeCompare = "compare"

	RoleUser      = "user"
	RoleAssistant = "assistant"
	RoleSystem    = "system"

	MessageStatusCompleted = "completed"
	MessageStatusFailed    = "failed"

	FileStatusReady  = "ready"
	FileStatusFailed = "failed"
	FileSourcePasted = "pasted"
	FileSourceUpload = "upload"
)

var (
	ErrServiceNotReady     = errors.New("copilot service is not configured")
	ErrInvalidInput        = errors.New("invalid copilot input")
	ErrInvalidAIResult     = errors.New("invalid copilot ai result")
	ErrToolPreviewNotFound = errors.New("copilot tool preview not found")
	ErrToolPreviewExpired  = errors.New("copilot tool preview expired")
	ErrToolPreviewConflict = errors.New("copilot tool preview is no longer pending")
	ErrThreadNotFound      = errors.New("copilot thread not found")
	ErrMemoryNotFound      = errors.New("copilot memory not found")
	ErrFileNotFound        = errors.New("copilot file not found")
	ErrFileTooLarge        = errors.New("copilot file is too large")
	ErrUnsupportedFileType = errors.New("unsupported copilot file type")
	ErrInvalidFileEncoding = errors.New("invalid copilot file encoding")
	ErrToolNotAvailable    = errors.New("copilot tool is not available")
)

type CreateThreadInput struct {
	UserID int64  `json:"-"`
	Title  string `json:"title"`
	Mode   string `json:"mode,omitempty"`
	Model  string `json:"model,omitempty"`
}

type SendMessageInput struct {
	UserID        int64             `json:"-"`
	ThreadID      int64             `json:"-"`
	Content       string            `json:"content"`
	Model         string            `json:"model,omitempty"`
	ReferenceIDs  []int64           `json:"reference_ids,omitempty"`
	RequestID     string            `json:"request_id,omitempty"`
	TaskID        int64             `json:"task_id,omitempty"`
	CurrentView   string            `json:"current_view,omitempty"`
	ActiveFilters map[string]string `json:"active_filters,omitempty"`
}

type CompareMessagesInput struct {
	UserID    int64    `json:"-"`
	ThreadID  int64    `json:"-"`
	Content   string   `json:"content"`
	Models    []string `json:"models,omitempty"`
	RequestID string   `json:"request_id,omitempty"`
}

type CompareSummaryInput struct {
	UserID    int64           `json:"-"`
	ThreadID  int64           `json:"-"`
	Content   string          `json:"content"`
	Model     string          `json:"model,omitempty"`
	Answers   []CompareAnswer `json:"answers"`
	RequestID string          `json:"request_id,omitempty"`
}

type ModelSmokeInput struct {
	UserID int64  `json:"-"`
	Model  string `json:"model,omitempty"`
	Prompt string `json:"prompt,omitempty"`
}

type MemoryInput struct {
	UserID     int64   `json:"-"`
	Key        string  `json:"key"`
	Value      string  `json:"value"`
	Confidence float64 `json:"confidence,omitempty"`
	Source     string  `json:"source,omitempty"`
	Status     string  `json:"status,omitempty"`
}

type MemoryUpdateInput struct {
	UserID int64  `json:"-"`
	ID     int64  `json:"-"`
	Key    string `json:"key,omitempty"`
	Value  string `json:"value,omitempty"`
	Status string `json:"status,omitempty"`
}

type FileInput struct {
	UserID   int64  `json:"-"`
	Name     string `json:"name"`
	MimeType string `json:"mime_type,omitempty"`
	Content  string `json:"content"`
}

type UploadFileInput struct {
	UserID   int64  `json:"-"`
	Name     string `json:"name"`
	MimeType string `json:"mime_type,omitempty"`
	Data     []byte `json:"-"`
}

type Thread struct {
	ID         int64      `json:"id"`
	UserID     int64      `json:"user_id"`
	Title      string     `json:"title"`
	Mode       string     `json:"mode"`
	Model      string     `json:"model,omitempty"`
	ArchivedAt *time.Time `json:"archived_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

type Message struct {
	ID           int64           `json:"id"`
	UserID       int64           `json:"user_id"`
	ThreadID     int64           `json:"thread_id"`
	Role         string          `json:"role"`
	Content      string          `json:"content"`
	Status       string          `json:"status"`
	Model        string          `json:"model,omitempty"`
	ErrorCode    string          `json:"error_code,omitempty"`
	InputTokens  int             `json:"input_tokens,omitempty"`
	OutputTokens int             `json:"output_tokens,omitempty"`
	Metadata     json.RawMessage `json:"metadata,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
}

type Memory struct {
	ID         int64     `json:"id"`
	UserID     int64     `json:"user_id"`
	Key        string    `json:"key"`
	Value      string    `json:"value"`
	Confidence float64   `json:"confidence"`
	Source     string    `json:"source,omitempty"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

const (
	MemoryStatusPending  = "pending"
	MemoryStatusActive   = "active"
	MemoryStatusInactive = "inactive"
)

type File struct {
	ID             int64     `json:"id"`
	UserID         int64     `json:"user_id"`
	Name           string    `json:"name"`
	MimeType       string    `json:"mime_type"`
	SizeBytes      int       `json:"size_bytes"`
	Content        string    `json:"content,omitempty"`
	Status         string    `json:"status"`
	Source         string    `json:"source"`
	SHA256         string    `json:"sha256,omitempty"`
	ExtractedChars int       `json:"extracted_chars"`
	ErrorCode      string    `json:"error_code,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type MemoryCandidate struct {
	Key        string  `json:"key"`
	Value      string  `json:"value"`
	Confidence float64 `json:"confidence"`
	Source     string  `json:"source"`
}

type ModelOption struct {
	Name      string `json:"name"`
	Value     string `json:"value"`
	Provider  string `json:"provider,omitempty"`
	IsDefault bool   `json:"is_default"`
}

type AIRun struct {
	ID            int64     `json:"id"`
	UserID        int64     `json:"user_id"`
	Feature       string    `json:"feature"`
	PromptVersion string    `json:"prompt_version"`
	Provider      string    `json:"provider"`
	Model         string    `json:"model"`
	Status        string    `json:"status"`
	ErrorCode     string    `json:"error_code,omitempty"`
	ErrorMessage  string    `json:"error_message,omitempty"`
	InputTokens   int       `json:"input_tokens,omitempty"`
	OutputTokens  int       `json:"output_tokens,omitempty"`
	LatencyMS     int       `json:"latency_ms,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type SendMessageResult struct {
	UserMessage      Message `json:"user_message"`
	AssistantMessage Message `json:"assistant_message"`
}

type ToolConfirmationInput struct {
	UserID    int64  `json:"-"`
	ThreadID  int64  `json:"-"`
	MessageID int64  `json:"-"`
	Decision  string `json:"decision"`
}

type ToolConfirmationResult struct {
	Message    Message              `json:"message"`
	ToolResult *ToolExecutionResult `json:"tool_result,omitempty"`
}

const (
	StreamEventUserMessage      = "user_message"
	StreamEventDelta            = "delta"
	StreamEventAssistantMessage = "assistant_message"
)

type StreamEvent struct {
	Type             string   `json:"type"`
	Delta            string   `json:"delta,omitempty"`
	UserMessage      *Message `json:"user_message,omitempty"`
	AssistantMessage *Message `json:"assistant_message,omitempty"`
}

const (
	ToolNone             = "none"
	ToolCreateTask       = "create_task"
	ToolProjectMatch     = "project_match"
	ToolPreviewPending   = "pending"
	ToolPreviewExecuting = "executing"
	ToolPreviewConfirmed = "confirmed"
	ToolPreviewCancelled = "cancelled"
	ToolPreviewExpired   = "expired"
)

type ToolArguments struct {
	Title       string   `json:"title,omitempty"`
	Description string   `json:"description,omitempty"`
	Priority    string   `json:"priority,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Intent      string   `json:"intent,omitempty"`
}

type ToolCall struct {
	Tool      string        `json:"tool"`
	Arguments ToolArguments `json:"arguments"`
}

type ToolPreview struct {
	ID              string    `json:"id"`
	SourceMessageID int64     `json:"source_message_id"`
	Call            ToolCall  `json:"call"`
	Status          string    `json:"status"`
	ExpiresAt       time.Time `json:"expires_at"`
	Error           string    `json:"error,omitempty"`
}

type TaskContext struct {
	ID                int64                `json:"id"`
	Title             string               `json:"title"`
	Description       string               `json:"description,omitempty"`
	Assignee          string               `json:"assignee,omitempty"`
	Project           string               `json:"project,omitempty"`
	Status            string               `json:"status"`
	Priority          string               `json:"priority"`
	Progress          int                  `json:"progress"`
	DueAt             *time.Time           `json:"due_at,omitempty"`
	Subtasks          []TaskSubtaskContext `json:"subtasks,omitempty"`
	CompletedSubtasks int                  `json:"completed_subtasks"`
	TotalSubtasks     int                  `json:"total_subtasks"`
	CurrentView       string               `json:"current_view,omitempty"`
	ActiveFilters     map[string]string    `json:"active_filters,omitempty"`
}

type TaskSubtaskContext struct {
	ID              int64      `json:"id"`
	ParentSubtaskID *int64     `json:"parent_subtask_id,omitempty"`
	Title           string     `json:"title"`
	Assignee        string     `json:"assignee,omitempty"`
	DueAt           *time.Time `json:"due_at,omitempty"`
	Completed       bool       `json:"completed"`
}

type ToolExecutionResult struct {
	Tool     string `json:"tool"`
	Status   string `json:"status"`
	EntityID int64  `json:"entity_id"`
	Title    string `json:"title"`
	URL      string `json:"url"`
	Message  string `json:"message"`
}

type CompareAnswer struct {
	Model            string  `json:"model"`
	AssistantMessage Message `json:"assistant_message"`
	ErrorCode        string  `json:"error_code,omitempty"`
}

type CompareMessagesResult struct {
	UserMessage Message         `json:"user_message"`
	Answers     []CompareAnswer `json:"answers"`
}

type CompareSummaryResult struct {
	SummaryMessage Message `json:"summary_message"`
}

type ModelSmokeResult struct {
	OK           bool   `json:"ok"`
	Model        string `json:"model"`
	Reply        string `json:"reply"`
	InputTokens  int    `json:"input_tokens,omitempty"`
	OutputTokens int    `json:"output_tokens,omitempty"`
}
