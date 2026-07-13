package sandbox

import (
	"errors"
	"time"
)

const (
	StatusDraft     = "draft"
	StatusQueued    = "queued"
	StatusRunning   = "running"
	StatusCompleted = "completed"
	StatusFailed    = "failed"
	StatusCanceled  = "canceled"
)

var (
	ErrServiceNotReady = errors.New("sandbox service is not configured")
	ErrInvalidAIResult = errors.New("invalid sandbox ai result")
	ErrInvalidSession  = errors.New("invalid sandbox session")
	ErrSessionNotFound = errors.New("sandbox session not found")
	ErrStaleRun        = errors.New("stale sandbox run")
)

type Role struct {
	Key         string `json:"key"`
	Label       string `json:"label"`
	Description string `json:"description"`
	Badge       string `json:"badge"`
}

type CreateInput struct {
	UserID      int64    `json:"-"`
	Goal        string   `json:"goal"`
	TargetUsers string   `json:"target_users"`
	Product     string   `json:"product"`
	Roles       []string `json:"roles"`
}

type DraftUpdate struct {
	Goal        *string   `json:"goal,omitempty"`
	TargetUsers *string   `json:"target_users,omitempty"`
	Product     *string   `json:"product,omitempty"`
	Roles       *[]string `json:"roles,omitempty"`
}

type AskRoleInput struct {
	UserID    int64  `json:"-"`
	SessionID int64  `json:"-"`
	Role      string `json:"role"`
	Question  string `json:"question"`
}

type Message struct {
	ID        int64     `json:"id"`
	SessionID int64     `json:"session_id"`
	UserID    int64     `json:"user_id"`
	Role      string    `json:"role"`
	Question  string    `json:"question"`
	Answer    string    `json:"answer"`
	CreatedAt time.Time `json:"created_at"`
}

func DefaultRoles() []Role {
	return []Role{
		{Key: "user", Label: "用户视角", Description: "评估产品体验与价值", Badge: "推荐优先"},
		{Key: "investor", Label: "投资人视角", Description: "评估市场潜力与回报", Badge: "热门选择"},
		{Key: "channel", Label: "代理商 / 渠道方视角", Description: "评估项目落地可行性", Badge: "渠道必选"},
		{Key: "competitor", Label: "竞争对手视角", Description: "评估竞争格局与策略", Badge: "深度分析"},
		{Key: "operator", Label: "运营视角", Description: "评估执行与增长策略", Badge: "运营必选"},
		{Key: "growth", Label: "增长策略", Description: "评估获客与规模化路径", Badge: "增长视角"},
		{Key: "risk", Label: "风险研判", Description: "识别合规、交付与经营风险", Badge: "风险视角"},
	}
}

type Session struct {
	ID              int64     `json:"id"`
	UserID          int64     `json:"user_id"`
	Goal            string    `json:"goal"`
	TargetUsers     string    `json:"target_users"`
	Product         string    `json:"product"`
	Roles           []string  `json:"roles"`
	Status          string    `json:"status"`
	ProgressPercent int       `json:"progress_percent"`
	CurrentStep     string    `json:"current_step"`
	ErrorMessage    string    `json:"error_message,omitempty"`
	RunAttempt      int       `json:"run_attempt"`
	Report          Report    `json:"report,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
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
