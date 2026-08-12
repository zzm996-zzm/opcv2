package projects

import (
	"errors"
	"time"

	projectfiles "github.com/zzm/opcv2/internal/projects/files"
)

const (
	StatusNeedsInput           = "needs_input"
	StatusCompleted            = "completed"
	OpportunityStatusDraft     = "draft"
	OpportunityStatusPublished = "published"
	CaseStatusDraft            = "draft"
	CaseStatusPublished        = "published"
	ExportSourceMatch          = "match"
	ExportSourceComparison     = "comparison"
)

var (
	ErrServiceNotReady     = errors.New("projects service is not configured")
	ErrInvalidAIResult     = errors.New("invalid projects ai result")
	ErrSessionNotFound     = errors.New("project match session not found")
	ErrOpportunityNotFound = errors.New("project opportunity not found")
	ErrCaseNotFound        = errors.New("project case not found")
	ErrInvalidMatchAnswers = errors.New("invalid project match answers")
	ErrComparisonNotFound  = errors.New("project comparison not found")
	ErrInvalidComparison   = errors.New("invalid project comparison")
	ErrExportNotFound      = errors.New("project export not found")
	ErrInvalidExport       = errors.New("invalid project export")
	ErrExportExpired       = errors.New("project export expired")
	ErrCompareLimit        = errors.New("project comparison limit reached")
)

type CreateExportInput struct {
	UserID     int64    `json:"-"`
	SourceType string   `json:"source_type"`
	SourceID   int64    `json:"source_id"`
	Format     string   `json:"format,omitempty"`
	Includes   []string `json:"includes,omitempty"`
}
type Export struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"user_id"`
	SourceType  string    `json:"source_type"`
	SourceID    int64     `json:"source_id"`
	Status      string    `json:"status"`
	Payload     []byte    `json:"-"`
	DownloadURL string    `json:"download_url"`
	Format      string    `json:"format"`
	Includes    []string  `json:"includes,omitempty"`
	ExpiresAt   time.Time `json:"expires_at"`
	CreatedAt   time.Time `json:"created_at"`
}

type CreateComparisonInput struct {
	UserID           int64    `json:"-"`
	OpportunitySlugs []string `json:"opportunity_slugs"`
}
type Comparison struct {
	ID        int64         `json:"id"`
	UserID    int64         `json:"user_id"`
	Items     []Opportunity `json:"items"`
	CreatedAt time.Time     `json:"created_at"`
}

type AnswerMatchInput struct {
	UserID    int64    `json:"-"`
	SessionID int64    `json:"-"`
	Answers   []Answer `json:"answers"`
}

type CaseFilters struct {
	CaseType        string
	OpportunitySlug string
	Industry        string
	Limit           int
}

type CaseStudy struct {
	ID               int64      `json:"id"`
	Slug             string     `json:"slug"`
	OpportunityID    *int64     `json:"opportunity_id,omitempty"`
	OpportunitySlug  string     `json:"opportunity_slug,omitempty"`
	OpportunityTitle string     `json:"opportunity_title,omitempty"`
	Industry         string     `json:"industry,omitempty"`
	Title            string     `json:"title"`
	Summary          string     `json:"summary"`
	CaseType         string     `json:"case_type"`
	Outcome          string     `json:"outcome"`
	KeyActions       []string   `json:"key_actions"`
	Lessons          []string   `json:"lessons"`
	Pitfalls         []string   `json:"pitfalls"`
	SourceTitle      string     `json:"source_title"`
	SourceURL        string     `json:"source_url"`
	CapturedAt       time.Time  `json:"captured_at"`
	Status           string     `json:"-"`
	PublishedAt      *time.Time `json:"published_at,omitempty"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type OpportunityFilters struct {
	Query    string
	Industry string
	Limit    int
}

type OpportunitySectionItem struct {
	Title    string   `json:"title,omitempty"`
	Value    string   `json:"value,omitempty"`
	Detail   string   `json:"detail,omitempty"`
	Meta     string   `json:"meta,omitempty"`
	Tone     string   `json:"tone,omitempty"`
	Progress float64  `json:"progress,omitempty"`
	Tags     []string `json:"tags,omitempty"`
}

type OpportunitySectionBlock struct {
	Type     string                   `json:"type,omitempty"`
	Title    string                   `json:"title,omitempty"`
	Subtitle string                   `json:"subtitle,omitempty"`
	Columns  int                      `json:"columns,omitempty"`
	Items    []OpportunitySectionItem `json:"items,omitempty"`
	Series   []OpportunitySectionItem `json:"series,omitempty"`
}

type OpportunitySection struct {
	Key    string                    `json:"key,omitempty"`
	Title  string                    `json:"title"`
	Body   string                    `json:"body"`
	Items  []string                  `json:"items"`
	Blocks []OpportunitySectionBlock `json:"blocks,omitempty"`
}

type Opportunity struct {
	ID                   int64                `json:"id"`
	Slug                 string               `json:"slug"`
	Title                string               `json:"title"`
	Summary              string               `json:"summary"`
	Industry             string               `json:"industry"`
	Tags                 []string             `json:"tags"`
	BudgetBand           string               `json:"budget_band"`
	Difficulty           string               `json:"difficulty"`
	ResourceRequirements []string             `json:"resource_requirements"`
	Sections             []OpportunitySection `json:"sections"`
	Status               string               `json:"-"`
	PublishedAt          *time.Time           `json:"published_at,omitempty"`
	UpdatedAt            time.Time            `json:"updated_at"`
}

type MatchInput struct {
	UserID     int64    `json:"-"`
	Intent     string   `json:"intent"`
	FileIDs    []int64  `json:"file_ids,omitempty"`
	FilePrompt string   `json:"-"`
	Answers    []Answer `json:"answers,omitempty"`
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
	Rank            int      `json:"rank"`
	OpportunitySlug string   `json:"opportunity_slug,omitempty"`
	Title           string   `json:"title"`
	Score           int      `json:"score"`
	Tags            []string `json:"tags"`
	Budget          string   `json:"budget"`
	Reasons         []string `json:"reasons"`
	Risk            string   `json:"risk"`
}

type MatchSession struct {
	ID        int64               `json:"id"`
	UserID    int64               `json:"user_id"`
	Intent    string              `json:"intent"`
	Answers   []Answer            `json:"answers,omitempty"`
	Status    string              `json:"status"`
	Questions []Question          `json:"questions,omitempty"`
	Result    MatchResult         `json:"result,omitempty"`
	Files     []projectfiles.File `json:"files,omitempty"`
	CreatedAt time.Time           `json:"created_at"`
	UpdatedAt time.Time           `json:"updated_at"`
}

type Favorite struct {
	ID        int64         `json:"id"`
	UserID    int64         `json:"user_id"`
	SessionID int64         `json:"session_id"`
	CreatedAt time.Time     `json:"created_at"`
	Session   *MatchSession `json:"session,omitempty"`
}
