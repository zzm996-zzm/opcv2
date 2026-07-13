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
	ErrInvalidAIResult     = errors.New("invalid content ai result")
	ErrInsightSourcesEmpty = errors.New("insight sources are unavailable")
)

type ArticleCitation struct {
	ID          string `json:"id"`
	Label       string `json:"label"`
	SourceName  string `json:"source_name"`
	SourceURL   string `json:"source_url"`
	Excerpt     string `json:"excerpt,omitempty"`
	PublishedAt string `json:"published_at,omitempty"`
}

type ArticleInput struct {
	Slug              string            `json:"slug"`
	Title             string            `json:"title"`
	Summary           string            `json:"summary,omitempty"`
	Body              string            `json:"body"`
	Status            string            `json:"status,omitempty"`
	SourceName        string            `json:"source_name,omitempty"`
	SourceURL         string            `json:"source_url,omitempty"`
	Author            string            `json:"author,omitempty"`
	Category          string            `json:"category,omitempty"`
	Tags              []string          `json:"tags"`
	Citations         []ArticleCitation `json:"citations"`
	SourcePublishedAt *time.Time        `json:"source_published_at,omitempty"`
}

type Article struct {
	ID                int64             `json:"id"`
	Slug              string            `json:"slug"`
	Title             string            `json:"title"`
	Summary           string            `json:"summary,omitempty"`
	Body              string            `json:"body,omitempty"`
	Status            string            `json:"status"`
	SourceName        string            `json:"source_name,omitempty"`
	SourceURL         string            `json:"source_url,omitempty"`
	Author            string            `json:"author,omitempty"`
	Category          string            `json:"category,omitempty"`
	Tags              []string          `json:"tags"`
	Citations         []ArticleCitation `json:"citations"`
	SourcePublishedAt *time.Time        `json:"source_published_at,omitempty"`
	PublishedAt       time.Time         `json:"published_at,omitempty"`
	CreatedAt         time.Time         `json:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at"`
}

type ArticleFilters struct {
	Category string
	Query    string
	Limit    int
}

type ToolInput struct {
	Slug            string     `json:"slug"`
	Name            string     `json:"name"`
	Description     string     `json:"description,omitempty"`
	URL             string     `json:"url,omitempty"`
	Status          string     `json:"status,omitempty"`
	Category        string     `json:"category,omitempty"`
	Tags            []string   `json:"tags"`
	ProviderName    string     `json:"provider_name,omitempty"`
	PriceLabel      string     `json:"price_label,omitempty"`
	Platforms       []string   `json:"platforms"`
	Features        []string   `json:"features"`
	UseCases        []string   `json:"use_cases"`
	Limitations     []string   `json:"limitations"`
	SourceURL       string     `json:"source_url,omitempty"`
	SourceUpdatedAt *time.Time `json:"source_updated_at,omitempty"`
	SortWeight      int        `json:"sort_weight,omitempty"`
}

type Tool struct {
	ID              int64      `json:"id"`
	Slug            string     `json:"slug"`
	Name            string     `json:"name"`
	Description     string     `json:"description,omitempty"`
	URL             string     `json:"url,omitempty"`
	Status          string     `json:"status"`
	Category        string     `json:"category,omitempty"`
	Tags            []string   `json:"tags"`
	ProviderName    string     `json:"provider_name,omitempty"`
	PriceLabel      string     `json:"price_label,omitempty"`
	Platforms       []string   `json:"platforms"`
	Features        []string   `json:"features"`
	UseCases        []string   `json:"use_cases"`
	Limitations     []string   `json:"limitations"`
	SourceURL       string     `json:"source_url,omitempty"`
	SourceUpdatedAt *time.Time `json:"source_updated_at,omitempty"`
	SortWeight      int        `json:"sort_weight"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type ToolFilters struct {
	Category string
	Query    string
	Sort     string
	Limit    int
}

type ToolRecommendationInput struct {
	Goal     string `json:"goal"`
	Scenario string `json:"scenario"`
	Category string `json:"category,omitempty"`
	Limit    int    `json:"limit,omitempty"`
}

type ToolRecommendations struct {
	Basis    string                  `json:"basis"`
	Criteria ToolRecommendationInput `json:"criteria"`
	Tools    []Tool                  `json:"tools"`
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
	Headline    string               `json:"headline"`
	Description string               `json:"description,omitempty"`
	JoinURL     string               `json:"join_url,omitempty"`
	QRVariants  []CommunityQRVariant `json:"qr_variants"`
}

type CommunityQRVariant struct {
	Key         string `json:"key"`
	Label       string `json:"label"`
	Description string `json:"description,omitempty"`
	ImageURL    string `json:"image_url,omitempty"`
	JoinURL     string `json:"join_url,omitempty"`
	Status      string `json:"status"`
}

type CommunityConfig struct {
	ID          int64                `json:"id"`
	Headline    string               `json:"headline"`
	Description string               `json:"description,omitempty"`
	JoinURL     string               `json:"join_url,omitempty"`
	QRVariants  []CommunityQRVariant `json:"qr_variants"`
	CreatedAt   time.Time            `json:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"`
}

type InsightQuestionInput struct {
	Question     string   `json:"question"`
	ArticleSlugs []string `json:"article_slugs"`
}

type InsightAnswerCitation struct {
	ID          string `json:"id"`
	ArticleSlug string `json:"article_slug"`
	Label       string `json:"label"`
	SourceName  string `json:"source_name"`
	SourceURL   string `json:"source_url"`
	Excerpt     string `json:"excerpt,omitempty"`
}

type InsightAnswer struct {
	Answer      string                  `json:"answer"`
	Citations   []InsightAnswerCitation `json:"citations"`
	Assumptions []string                `json:"assumptions"`
	Basis       string                  `json:"basis"`
	Disclaimer  string                  `json:"disclaimer"`
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
