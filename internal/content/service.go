package content

import (
	"context"
	"strings"
	"time"
)

type Repository interface {
	IsAdmin(ctx context.Context, userID int64) (bool, error)
	ListArticles(ctx context.Context) ([]Article, error)
	GetArticle(ctx context.Context, slug string) (Article, error)
	UpsertArticle(ctx context.Context, article Article) (Article, error)
	ListTools(ctx context.Context) ([]Tool, error)
	UpsertTool(ctx context.Context, tool Tool) (Tool, error)
	GetCommunityConfig(ctx context.Context) (CommunityConfig, error)
	UpsertCommunityConfig(ctx context.Context, config CommunityConfig) (CommunityConfig, error)
	ListBrandMetrics(ctx context.Context) ([]BrandMetric, error)
	UpsertBrandMetric(ctx context.Context, metric BrandMetric) (BrandMetric, error)
	ListBrandCases(ctx context.Context) ([]BrandCase, error)
	UpsertBrandCase(ctx context.Context, brandCase BrandCase) (BrandCase, error)
}

type Service struct {
	repository Repository
	now        func() time.Time
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository, now: time.Now}
}

func (s *Service) ListArticles(ctx context.Context) ([]Article, error) {
	if s.repository == nil {
		return nil, ErrServiceNotReady
	}
	articles, err := s.repository.ListArticles(ctx)
	if articles == nil {
		articles = []Article{}
	}
	return articles, err
}

func (s *Service) GetArticle(ctx context.Context, slug string) (Article, error) {
	if s.repository == nil {
		return Article{}, ErrServiceNotReady
	}
	slug = strings.TrimSpace(slug)
	if slug == "" {
		return Article{}, ErrInvalidInput
	}
	return s.repository.GetArticle(ctx, slug)
}

func (s *Service) CreateArticle(ctx context.Context, userID int64, input ArticleInput) (Article, error) {
	if err := s.requireAdmin(ctx, userID); err != nil {
		return Article{}, err
	}
	input = normalizeArticleInput(input)
	if input.Slug == "" || input.Title == "" || input.Body == "" || !validStatus(input.Status) {
		return Article{}, ErrInvalidInput
	}
	now := s.now()
	article := Article{
		Slug:      input.Slug,
		Title:     input.Title,
		Summary:   input.Summary,
		Body:      input.Body,
		Status:    input.Status,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if input.Status == StatusPublished {
		article.PublishedAt = now
	}
	return s.repository.UpsertArticle(ctx, article)
}

func (s *Service) ListTools(ctx context.Context) ([]Tool, error) {
	if s.repository == nil {
		return nil, ErrServiceNotReady
	}
	tools, err := s.repository.ListTools(ctx)
	if tools == nil {
		tools = []Tool{}
	}
	return tools, err
}

func (s *Service) UpsertTool(ctx context.Context, userID int64, input ToolInput) (Tool, error) {
	if err := s.requireAdmin(ctx, userID); err != nil {
		return Tool{}, err
	}
	input = normalizeToolInput(input)
	if input.Slug == "" || input.Name == "" || !validStatus(input.Status) {
		return Tool{}, ErrInvalidInput
	}
	now := s.now()
	return s.repository.UpsertTool(ctx, Tool{
		Slug:        input.Slug,
		Name:        input.Name,
		Description: input.Description,
		URL:         input.URL,
		Status:      input.Status,
		CreatedAt:   now,
		UpdatedAt:   now,
	})
}

func (s *Service) GetCommunityConfig(ctx context.Context) (CommunityConfig, error) {
	if s.repository == nil {
		return CommunityConfig{}, ErrServiceNotReady
	}
	return s.repository.GetCommunityConfig(ctx)
}

func (s *Service) UpdateCommunityConfig(ctx context.Context, userID int64, input CommunityConfigInput) (CommunityConfig, error) {
	if err := s.requireAdmin(ctx, userID); err != nil {
		return CommunityConfig{}, err
	}
	input.Headline = strings.TrimSpace(input.Headline)
	input.Description = strings.TrimSpace(input.Description)
	input.JoinURL = strings.TrimSpace(input.JoinURL)
	if input.Headline == "" {
		return CommunityConfig{}, ErrInvalidInput
	}
	now := s.now()
	return s.repository.UpsertCommunityConfig(ctx, CommunityConfig{
		Headline:    input.Headline,
		Description: input.Description,
		JoinURL:     input.JoinURL,
		CreatedAt:   now,
		UpdatedAt:   now,
	})
}

func (s *Service) GetBrand(ctx context.Context) (BrandContent, error) {
	if s.repository == nil {
		return BrandContent{}, ErrServiceNotReady
	}
	metrics, err := s.repository.ListBrandMetrics(ctx)
	if err != nil {
		return BrandContent{}, err
	}
	cases, err := s.repository.ListBrandCases(ctx)
	if err != nil {
		return BrandContent{}, err
	}
	if metrics == nil {
		metrics = []BrandMetric{}
	}
	if cases == nil {
		cases = []BrandCase{}
	}
	return BrandContent{Metrics: metrics, Cases: cases}, nil
}

func (s *Service) UpsertBrandMetric(ctx context.Context, userID int64, input BrandMetricInput) (BrandMetric, error) {
	if err := s.requireAdmin(ctx, userID); err != nil {
		return BrandMetric{}, err
	}
	input.Key = strings.TrimSpace(input.Key)
	input.Label = strings.TrimSpace(input.Label)
	input.Value = strings.TrimSpace(input.Value)
	if input.Key == "" || input.Label == "" || input.Value == "" {
		return BrandMetric{}, ErrInvalidInput
	}
	now := s.now()
	return s.repository.UpsertBrandMetric(ctx, BrandMetric{
		Key:       input.Key,
		Label:     input.Label,
		Value:     input.Value,
		CreatedAt: now,
		UpdatedAt: now,
	})
}

func (s *Service) UpsertBrandCase(ctx context.Context, userID int64, input BrandCaseInput) (BrandCase, error) {
	if err := s.requireAdmin(ctx, userID); err != nil {
		return BrandCase{}, err
	}
	input.Slug = strings.TrimSpace(input.Slug)
	input.Title = strings.TrimSpace(input.Title)
	input.Summary = strings.TrimSpace(input.Summary)
	input.URL = strings.TrimSpace(input.URL)
	if input.Slug == "" || input.Title == "" {
		return BrandCase{}, ErrInvalidInput
	}
	now := s.now()
	return s.repository.UpsertBrandCase(ctx, BrandCase{
		Slug:      input.Slug,
		Title:     input.Title,
		Summary:   input.Summary,
		URL:       input.URL,
		CreatedAt: now,
		UpdatedAt: now,
	})
}

func (s *Service) requireAdmin(ctx context.Context, userID int64) error {
	if s.repository == nil {
		return ErrServiceNotReady
	}
	if userID <= 0 {
		return ErrAdminRequired
	}
	ok, err := s.repository.IsAdmin(ctx, userID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrAdminRequired
	}
	return nil
}

func normalizeArticleInput(input ArticleInput) ArticleInput {
	input.Slug = strings.TrimSpace(input.Slug)
	input.Title = strings.TrimSpace(input.Title)
	input.Summary = strings.TrimSpace(input.Summary)
	input.Body = strings.TrimSpace(input.Body)
	input.Status = strings.TrimSpace(input.Status)
	if input.Status == "" {
		input.Status = StatusDraft
	}
	return input
}

func normalizeToolInput(input ToolInput) ToolInput {
	input.Slug = strings.TrimSpace(input.Slug)
	input.Name = strings.TrimSpace(input.Name)
	input.Description = strings.TrimSpace(input.Description)
	input.URL = strings.TrimSpace(input.URL)
	input.Status = strings.TrimSpace(input.Status)
	if input.Status == "" {
		input.Status = StatusDraft
	}
	return input
}

func validStatus(status string) bool {
	return status == StatusDraft || status == StatusPublished
}
