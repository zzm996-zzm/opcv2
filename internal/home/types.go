package home

import "time"

type Card struct {
	Title   string `json:"title"`
	Summary string `json:"summary,omitempty"`
	URL     string `json:"url,omitempty"`
}

type Metric struct {
	Label string `json:"label"`
	Value string `json:"value"`
	Icon  string `json:"icon,omitempty"`
}

type ActionItem struct {
	Type     string `json:"type"`
	Priority string `json:"priority"`
	Title    string `json:"title"`
	Summary  string `json:"summary,omitempty"`
	URL      string `json:"url,omitempty"`
	CTA      string `json:"cta,omitempty"`
}

type RecentTask struct {
	ID      int64      `json:"id"`
	Title   string     `json:"title"`
	Project string     `json:"project"`
	Status  string     `json:"status"`
	DueAt   *time.Time `json:"due_at,omitempty"`
}

type NotificationItem struct {
	ID        int64     `json:"id"`
	Type      string    `json:"type"`
	Title     string    `json:"title"`
	Summary   string    `json:"summary,omitempty"`
	ActionURL string    `json:"action_url,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type NotificationSummary struct {
	Unread int                `json:"unread"`
	Latest []NotificationItem `json:"latest"`
}

type QuotaWarning struct {
	Key     string `json:"key"`
	Label   string `json:"label"`
	Used    int    `json:"used"`
	Limit   int    `json:"limit"`
	Message string `json:"message"`
}

type AccountSummary struct {
	PlanName      string         `json:"plan_name"`
	CreditBalance int            `json:"credit_balance"`
	QuotaWarnings []QuotaWarning `json:"quota_warnings"`
}

type Summary struct {
	Metrics             []Metric            `json:"metrics"`
	HeroCards           []Card              `json:"hero_cards"`
	Recommendations     []Card              `json:"recommendations"`
	ActionItems         []ActionItem        `json:"action_items"`
	RecentTasks         []RecentTask        `json:"recent_tasks"`
	NotificationSummary NotificationSummary `json:"notification_summary"`
	AccountSummary      AccountSummary      `json:"account_summary"`
}
