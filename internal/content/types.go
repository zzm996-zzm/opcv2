package content

import (
	"errors"
	"time"
)

const (
	StatusDraft     = "draft"
	StatusPublished = "published"
)

var (
	ErrServiceNotReady     = errors.New("content service is not configured")
	ErrInvalidInput        = errors.New("invalid content input")
	ErrAdminRequired       = errors.New("admin role required")
	ErrArticleNotFound     = errors.New("content article not found")
	ErrToolNotFound        = errors.New("content tool not found")
	ErrHelpArticleNotFound = errors.New("help article not found")
)

type ArticleInput struct {
	Slug    string `json:"slug"`
	Title   string `json:"title"`
	Summary string `json:"summary,omitempty"`
	Body    string `json:"body"`
	Status  string `json:"status,omitempty"`
}

type Article struct {
	ID          int64     `json:"id"`
	Slug        string    `json:"slug"`
	Title       string    `json:"title"`
	Summary     string    `json:"summary,omitempty"`
	Body        string    `json:"body,omitempty"`
	Status      string    `json:"status"`
	PublishedAt time.Time `json:"published_at,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ToolInput struct {
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	URL         string `json:"url,omitempty"`
	Status      string `json:"status,omitempty"`
	Category    string `json:"category,omitempty"`
}

type Tool struct {
	ID          int64     `json:"id"`
	Slug        string    `json:"slug"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	URL         string    `json:"url,omitempty"`
	Status      string    `json:"status"`
	Category    string    `json:"category,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ToolFilters struct {
	Category string
	Query    string
	Sort     string
	Limit    int
}

type FavoriteResult struct {
	Slug      string `json:"slug"`
	Favorited bool   `json:"favorited"`
}

type BookmarkResult struct {
	Slug       string `json:"slug"`
	Bookmarked bool   `json:"bookmarked"`
}

type CommunityConfigInput struct {
	Headline    string `json:"headline"`
	Description string `json:"description,omitempty"`
	JoinURL     string `json:"join_url,omitempty"`
}

type CommunityConfig struct {
	ID          int64     `json:"id"`
	Headline    string    `json:"headline"`
	Description string    `json:"description,omitempty"`
	JoinURL     string    `json:"join_url,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CommunityJoinInput struct {
	Community string `json:"community"`
	Contact   string `json:"contact"`
	Note      string `json:"note,omitempty"`
}

type CommunityJoinRequest struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id,omitempty"`
	Community string    `json:"community"`
	Contact   string    `json:"contact,omitempty"`
	Note      string    `json:"note,omitempty"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type HelpTopic struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

type HelpArticleFilters struct {
	Topic string
	Query string
	Limit int
}

type HelpArticle struct {
	ID        int64     `json:"id,omitempty"`
	Slug      string    `json:"slug"`
	Topic     string    `json:"topic"`
	Title     string    `json:"title"`
	Summary   string    `json:"summary,omitempty"`
	Body      string    `json:"body,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

type BrandMetricInput struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	Value string `json:"value"`
}

type BrandMetric struct {
	ID        int64     `json:"id"`
	Key       string    `json:"key"`
	Label     string    `json:"label"`
	Value     string    `json:"value"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type BrandCaseInput struct {
	Slug    string `json:"slug"`
	Title   string `json:"title"`
	Summary string `json:"summary,omitempty"`
	URL     string `json:"url,omitempty"`
}

type BrandCase struct {
	ID        int64     `json:"id"`
	Slug      string    `json:"slug"`
	Title     string    `json:"title"`
	Summary   string    `json:"summary,omitempty"`
	URL       string    `json:"url,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type BrandContent struct {
	Metrics []BrandMetric `json:"metrics"`
	Cases   []BrandCase   `json:"cases"`
}
