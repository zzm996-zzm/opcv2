package projects

import (
	"errors"
	"time"
)

const (
	StatusNeedsInput = "needs_input"
	StatusCompleted  = "completed"
)

var (
	ErrServiceNotReady = errors.New("projects service is not configured")
	ErrInvalidAIResult = errors.New("invalid projects ai result")
	ErrSessionNotFound = errors.New("project match session not found")
)

type MatchInput struct {
	UserID  int64    `json:"-"`
	Intent  string   `json:"intent"`
	Answers []Answer `json:"answers,omitempty"`
}

type Answer struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Question struct {
	Key     string   `json:"key"`
	Text    string   `json:"text"`
	Options []string `json:"options"`
}

type MatchResult struct {
	SessionID int64          `json:"session_id"`
	Status    string         `json:"status"`
	Questions []Question     `json:"questions,omitempty"`
	Projects  []ProjectMatch `json:"projects,omitempty"`
}

type ProjectMatch struct {
	Rank    int      `json:"rank"`
	Title   string   `json:"title"`
	Score   int      `json:"score"`
	Tags    []string `json:"tags"`
	Budget  string   `json:"budget"`
	Reasons []string `json:"reasons"`
	Risk    string   `json:"risk"`
}

type MatchSession struct {
	ID        int64       `json:"id"`
	UserID    int64       `json:"user_id"`
	Intent    string      `json:"intent"`
	Status    string      `json:"status"`
	Questions []Question  `json:"questions,omitempty"`
	Result    MatchResult `json:"result,omitempty"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}

type Favorite struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	SessionID int64     `json:"session_id"`
	CreatedAt time.Time `json:"created_at"`
}
