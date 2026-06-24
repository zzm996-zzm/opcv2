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
	ErrServiceNotReady = errors.New("content service is not configured")
	ErrInvalidInput    = errors.New("invalid content input")
	ErrAdminRequired   = errors.New("admin role required")
	ErrArticleNotFound = errors.New("content article not found")
	ErrToolNotFound    = errors.New("content tool not found")
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
}

type Tool struct {
	ID          int64     `json:"id"`
	Slug        string    `json:"slug"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	URL         string    `json:"url,omitempty"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
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
