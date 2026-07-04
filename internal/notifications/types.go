package notifications

import "time"

const (
	TypeSystem     = "system"
	TypeTask       = "task"
	TypeAnalysis   = "analysis"
	TypeLead       = "lead"
	TypeCRM        = "crm"
	TypeMembership = "membership"

	StatusAll    = "all"
	StatusUnread = "unread"
	StatusRead   = "read"
)

type Notification struct {
	ID          int64      `json:"id"`
	UserID      int64      `json:"user_id"`
	Type        string     `json:"type"`
	Title       string     `json:"title"`
	Summary     string     `json:"summary"`
	Body        string     `json:"body"`
	SourceType  string     `json:"source_type"`
	SourceID    *int64     `json:"source_id,omitempty"`
	ActionLabel string     `json:"action_label"`
	ActionURL   string     `json:"action_url"`
	ReadAt      *time.Time `json:"read_at"`
	CreatedAt   time.Time  `json:"created_at"`
}

type ListFilters struct {
	Type   string
	Status string
	Limit  int
}

type TypeCount struct {
	Type  string `json:"type"`
	Count int    `json:"count"`
}

type Summary struct {
	Unread int            `json:"unread"`
	ByType []TypeCount    `json:"by_type"`
	Latest []Notification `json:"latest"`
}
