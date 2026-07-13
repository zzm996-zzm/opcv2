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
)

var (
	ErrServiceNotReady = errors.New("copilot service is not configured")
	ErrInvalidInput    = errors.New("invalid copilot input")
	ErrInvalidAIResult = errors.New("invalid copilot ai result")
	ErrThreadNotFound  = errors.New("copilot thread not found")
	ErrMemoryNotFound  = errors.New("copilot memory not found")
	ErrFileNotFound    = errors.New("copilot file not found")
)

type CreateThreadInput struct {
	UserID int64  `json:"-"`
	Title  string `json:"title"`
	Mode   string `json:"mode,omitempty"`
	Model  string `json:"model,omitempty"`
}

type SendMessageInput struct {
	UserID       int64   `json:"-"`
	ThreadID     int64   `json:"-"`
	Content      string  `json:"content"`
	Model        string  `json:"model,omitempty"`
	ReferenceIDs []int64 `json:"reference_ids,omitempty"`
	RequestID    string  `json:"request_id,omitempty"`
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
}

type FileInput struct {
	UserID   int64  `json:"-"`
	Name     string `json:"name"`
	MimeType string `json:"mime_type,omitempty"`
	Content  string `json:"content"`
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
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type File struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Name      string    `json:"name"`
	MimeType  string    `json:"mime_type"`
	SizeBytes int       `json:"size_bytes"`
	Content   string    `json:"content,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
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
