package content

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/zzm/opcv2/internal/ai"
)

type Repository interface {
	IsAdmin(ctx context.Context, userID int64) (bool, error)
	ListArticles(ctx context.Context, filters ArticleFilters) ([]Article, error)
	GetArticle(ctx context.Context, slug string) (Article, error)
	UpsertArticle(ctx context.Context, article Article) (Article, error)
	BookmarkArticle(ctx context.Context, userID int64, slug string) (BookmarkResult, error)
	UnbookmarkArticle(ctx context.Context, userID int64, slug string) (BookmarkResult, error)
	ListTools(ctx context.Context, filters ToolFilters) ([]Tool, error)
	GetTool(ctx context.Context, slug string) (Tool, error)
	FavoriteTool(ctx context.Context, userID int64, slug string) (FavoriteResult, error)
	UnfavoriteTool(ctx context.Context, userID int64, slug string) (FavoriteResult, error)
	UpsertTool(ctx context.Context, tool Tool) (Tool, error)
	GetCommunityConfig(ctx context.Context) (CommunityConfig, error)
	UpsertCommunityConfig(ctx context.Context, config CommunityConfig) (CommunityConfig, error)
	CreateCommunityJoinRequest(ctx context.Context, userID int64, input CommunityJoinInput) (CommunityJoinRequest, error)
	ListHelpTopics(ctx context.Context) ([]HelpTopic, error)
	ListHelpArticles(ctx context.Context, filters HelpArticleFilters) ([]HelpArticle, error)
	GetHelpArticle(ctx context.Context, slug string) (HelpArticle, error)
	ListBrandMetrics(ctx context.Context) ([]BrandMetric, error)
	UpsertBrandMetric(ctx context.Context, metric BrandMetric) (BrandMetric, error)
	ListBrandCases(ctx context.Context) ([]BrandCase, error)
	UpsertBrandCase(ctx context.Context, brandCase BrandCase) (BrandCase, error)
}

type JSONGenerator interface {
	GenerateJSON(ctx context.Context, request ai.GenerateJSONRequest) (ai.GenerateJSONResult, error)
}

type Option func(*Service)

func WithGenerator(generator JSONGenerator) Option {
	return func(service *Service) { service.generator = generator }
}

type Service struct {
	repository Repository
	generator  JSONGenerator
	now        func() time.Time
}

func NewService(repository Repository, options ...Option) *Service {
	service := &Service{repository: repository, now: time.Now}
	for _, option := range options {
		option(service)
	}
	return service
}

func (s *Service) ListArticles(ctx context.Context, filters ArticleFilters) ([]Article, error) {
	if s.repository == nil {
		return nil, ErrServiceNotReady
	}
	filters.Category = strings.TrimSpace(filters.Category)
	filters.Query = strings.TrimSpace(filters.Query)
	if filters.Limit <= 0 {
		filters.Limit = 20
	}
	if filters.Limit > 100 {
		filters.Limit = 100
	}
	articles, err := s.repository.ListArticles(ctx, filters)
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
		Slug: input.Slug, Title: input.Title, Summary: input.Summary, Body: input.Body,
		Status: input.Status, SourceName: input.SourceName, SourceURL: input.SourceURL,
		Author: input.Author, Category: input.Category, Tags: input.Tags,
		Citations: input.Citations, SourcePublishedAt: input.SourcePublishedAt,
		CreatedAt: now, UpdatedAt: now,
	}
	if input.Status == StatusPublished {
		article.PublishedAt = now
	}
	return s.repository.UpsertArticle(ctx, article)
}

func (s *Service) BookmarkArticle(ctx context.Context, userID int64, slug string) (BookmarkResult, error) {
	if s.repository == nil {
		return BookmarkResult{}, ErrServiceNotReady
	}
	slug = strings.TrimSpace(slug)
	if userID <= 0 || slug == "" {
		return BookmarkResult{}, ErrInvalidInput
	}
	return s.repository.BookmarkArticle(ctx, userID, slug)
}

func (s *Service) UnbookmarkArticle(ctx context.Context, userID int64, slug string) (BookmarkResult, error) {
	if s.repository == nil {
		return BookmarkResult{}, ErrServiceNotReady
	}
	slug = strings.TrimSpace(slug)
	if userID <= 0 || slug == "" {
		return BookmarkResult{}, ErrInvalidInput
	}
	return s.repository.UnbookmarkArticle(ctx, userID, slug)
}

func (s *Service) ListTools(ctx context.Context, filters ToolFilters) ([]Tool, error) {
	if s.repository == nil {
		return nil, ErrServiceNotReady
	}
	filters.Category = strings.TrimSpace(filters.Category)
	filters.Query = strings.TrimSpace(filters.Query)
	filters.Sort = strings.TrimSpace(filters.Sort)
	if filters.Sort == "" {
		filters.Sort = "featured"
	}
	if filters.Limit <= 0 {
		filters.Limit = 20
	}
	if filters.Limit > 100 {
		filters.Limit = 100
	}
	tools, err := s.repository.ListTools(ctx, filters)
	if tools == nil {
		tools = []Tool{}
	}
	return tools, err
}

func (s *Service) GetTool(ctx context.Context, slug string) (Tool, error) {
	if s.repository == nil {
		return Tool{}, ErrServiceNotReady
	}
	slug = strings.TrimSpace(slug)
	if slug == "" {
		return Tool{}, ErrInvalidInput
	}
	return s.repository.GetTool(ctx, slug)
}

func (s *Service) RecommendTools(ctx context.Context, input ToolRecommendationInput) (ToolRecommendations, error) {
	if s.repository == nil {
		return ToolRecommendations{}, ErrServiceNotReady
	}
	input.Goal = strings.TrimSpace(input.Goal)
	input.Scenario = strings.TrimSpace(input.Scenario)
	input.Category = strings.TrimSpace(input.Category)
	if input.Goal == "" || input.Scenario == "" {
		return ToolRecommendations{}, ErrInvalidInput
	}
	if input.Limit <= 0 {
		input.Limit = 6
	}
	if input.Limit > 20 {
		input.Limit = 20
	}
	tools, err := s.repository.ListTools(ctx, ToolFilters{
		Category: input.Category,
		Query:    input.Scenario,
		Sort:     "match",
		Limit:    input.Limit,
	})
	if err != nil {
		return ToolRecommendations{}, err
	}
	if tools == nil {
		tools = []Tool{}
	}
	return ToolRecommendations{Basis: "catalog_match", Criteria: input, Tools: tools}, nil
}

func (s *Service) FavoriteTool(ctx context.Context, userID int64, slug string) (FavoriteResult, error) {
	if s.repository == nil {
		return FavoriteResult{}, ErrServiceNotReady
	}
	slug = strings.TrimSpace(slug)
	if userID <= 0 || slug == "" {
		return FavoriteResult{}, ErrInvalidInput
	}
	return s.repository.FavoriteTool(ctx, userID, slug)
}

func (s *Service) UnfavoriteTool(ctx context.Context, userID int64, slug string) (FavoriteResult, error) {
	if s.repository == nil {
		return FavoriteResult{}, ErrServiceNotReady
	}
	slug = strings.TrimSpace(slug)
	if userID <= 0 || slug == "" {
		return FavoriteResult{}, ErrInvalidInput
	}
	return s.repository.UnfavoriteTool(ctx, userID, slug)
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
		Slug: input.Slug, Name: input.Name, Description: input.Description, URL: input.URL,
		Status: input.Status, Category: input.Category, Tags: input.Tags,
		ProviderName: input.ProviderName, PriceLabel: input.PriceLabel,
		Platforms: input.Platforms, Features: input.Features, UseCases: input.UseCases,
		Limitations: input.Limitations, SourceURL: input.SourceURL,
		SourceUpdatedAt: input.SourceUpdatedAt, SortWeight: input.SortWeight,
		CreatedAt: now, UpdatedAt: now,
	})
}

type insightModelResult struct {
	Answer      string   `json:"answer"`
	CitationIDs []string `json:"citation_ids"`
	Assumptions []string `json:"assumptions"`
}

func (s *Service) AnswerInsightQuestion(ctx context.Context, userID int64, input InsightQuestionInput) (InsightAnswer, error) {
	if s.repository == nil || s.generator == nil {
		return InsightAnswer{}, ErrServiceNotReady
	}
	input.Question = strings.TrimSpace(input.Question)
	input.ArticleSlugs = normalizeUniqueStrings(input.ArticleSlugs)
	if userID <= 0 || input.Question == "" || len(input.ArticleSlugs) == 0 || len(input.ArticleSlugs) > 5 {
		return InsightAnswer{}, ErrInvalidInput
	}

	articles := make([]Article, 0, len(input.ArticleSlugs))
	available := make(map[string]InsightAnswerCitation)
	for _, slug := range input.ArticleSlugs {
		article, err := s.repository.GetArticle(ctx, slug)
		if err != nil {
			return InsightAnswer{}, err
		}
		articles = append(articles, article)
		for _, citation := range article.Citations {
			key := article.Slug + ":" + citation.ID
			if citation.ID == "" || citation.SourceURL == "" {
				continue
			}
			available[key] = InsightAnswerCitation{
				ID: key, ArticleSlug: article.Slug, Label: citation.Label,
				SourceName: citation.SourceName, SourceURL: citation.SourceURL, Excerpt: citation.Excerpt,
			}
		}
	}
	if len(available) == 0 {
		return InsightAnswer{}, ErrInsightSourcesEmpty
	}

	validate := func(data []byte) error {
		var result insightModelResult
		if err := json.Unmarshal(data, &result); err != nil {
			return err
		}
		if strings.TrimSpace(result.Answer) == "" || len(result.CitationIDs) == 0 {
			return fmt.Errorf("answer and citation_ids are required")
		}
		for _, id := range result.CitationIDs {
			if _, ok := available[id]; !ok {
				return fmt.Errorf("unknown citation id %q", id)
			}
		}
		return nil
	}
	generated, err := s.generator.GenerateJSON(ctx, ai.GenerateJSONRequest{
		UserID: userID, Feature: "content.insight_qa", PromptVersion: "content_insight_qa_v1",
		SystemPrompt: "你是资讯问答助手。只能依据提供的文章和引用作答；只返回 JSON，字段为 answer、citation_ids、assumptions。citation_ids 必须逐字使用提供的引用编号；证据不足时要明确说明，不得补造事实或来源。",
		UserPrompt:   insightQuestionPrompt(input.Question, articles), SchemaName: "content_insight_answer",
		Validate: validate, RepairAttempts: 1,
	})
	if err != nil {
		return InsightAnswer{}, fmt.Errorf("%w: %v", ErrInvalidAIResult, err)
	}
	var result insightModelResult
	if err := json.Unmarshal(generated.Content, &result); err != nil {
		return InsightAnswer{}, fmt.Errorf("%w: %v", ErrInvalidAIResult, err)
	}
	citations := make([]InsightAnswerCitation, 0, len(result.CitationIDs))
	seen := map[string]struct{}{}
	for _, id := range result.CitationIDs {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		citations = append(citations, available[id])
	}
	return InsightAnswer{
		Answer: strings.TrimSpace(result.Answer), Citations: citations,
		Assumptions: normalizeUniqueStrings(result.Assumptions), Basis: "catalog_citations",
		Disclaimer: "AI 回答仅基于所选已发布资讯及其已保存引用，不代表对外部来源完成实时核验。",
	}, nil
}

func insightQuestionPrompt(question string, articles []Article) string {
	var builder strings.Builder
	builder.WriteString("问题：")
	builder.WriteString(question)
	for _, article := range articles {
		builder.WriteString("\n\n文章[")
		builder.WriteString(article.Slug)
		builder.WriteString("]：")
		builder.WriteString(article.Title)
		builder.WriteString("\n摘要：")
		builder.WriteString(article.Summary)
		builder.WriteString("\n正文：")
		builder.WriteString(article.Body)
		for _, citation := range article.Citations {
			if citation.ID == "" || citation.SourceURL == "" {
				continue
			}
			builder.WriteString("\n引用编号=")
			builder.WriteString(article.Slug + ":" + citation.ID)
			builder.WriteString("；来源=")
			builder.WriteString(citation.SourceName)
			builder.WriteString("；标题=")
			builder.WriteString(citation.Label)
			builder.WriteString("；URL=")
			builder.WriteString(citation.SourceURL)
			builder.WriteString("；摘录=")
			builder.WriteString(citation.Excerpt)
		}
	}
	return builder.String()
}

func (s *Service) GetCommunityConfig(ctx context.Context) (CommunityConfig, error) {
	if s.repository == nil {
		return CommunityConfig{}, ErrServiceNotReady
	}
	config, err := s.repository.GetCommunityConfig(ctx)
	if err != nil {
		return CommunityConfig{}, err
	}
	published := make([]CommunityQRVariant, 0, len(config.QRVariants))
	for _, variant := range config.QRVariants {
		if variant.Status == StatusPublished {
			published = append(published, variant)
		}
	}
	config.QRVariants = published
	return config, nil
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
		QRVariants:  normalizeQRVariants(input.QRVariants),
		CreatedAt:   now,
		UpdatedAt:   now,
	})
}

func (s *Service) CreateCommunityJoinRequest(ctx context.Context, userID int64, input CommunityJoinInput) (CommunityJoinRequest, error) {
	if s.repository == nil {
		return CommunityJoinRequest{}, ErrServiceNotReady
	}
	input.Community = strings.TrimSpace(input.Community)
	input.Contact = strings.TrimSpace(input.Contact)
	input.Note = strings.TrimSpace(input.Note)
	if userID <= 0 || input.Contact == "" || (input.Community != "members" && input.Community != "enterprise") {
		return CommunityJoinRequest{}, ErrInvalidInput
	}
	return s.repository.CreateCommunityJoinRequest(ctx, userID, input)
}

func (s *Service) ListHelpTopics(ctx context.Context) ([]HelpTopic, error) {
	if s.repository == nil {
		return nil, ErrServiceNotReady
	}
	topics, err := s.repository.ListHelpTopics(ctx)
	if topics == nil {
		topics = []HelpTopic{}
	}
	return topics, err
}

func (s *Service) ListHelpArticles(ctx context.Context, filters HelpArticleFilters) ([]HelpArticle, error) {
	if s.repository == nil {
		return nil, ErrServiceNotReady
	}
	filters.Topic = strings.TrimSpace(filters.Topic)
	filters.Query = strings.TrimSpace(filters.Query)
	if filters.Limit <= 0 {
		filters.Limit = 20
	}
	if filters.Limit > 100 {
		filters.Limit = 100
	}
	articles, err := s.repository.ListHelpArticles(ctx, filters)
	if articles == nil {
		articles = []HelpArticle{}
	}
	return articles, err
}

func (s *Service) GetHelpArticle(ctx context.Context, slug string) (HelpArticle, error) {
	if s.repository == nil {
		return HelpArticle{}, ErrServiceNotReady
	}
	slug = strings.TrimSpace(slug)
	if slug == "" {
		return HelpArticle{}, ErrInvalidInput
	}
	return s.repository.GetHelpArticle(ctx, slug)
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
	input.SourceName = strings.TrimSpace(input.SourceName)
	input.SourceURL = strings.TrimSpace(input.SourceURL)
	input.Author = strings.TrimSpace(input.Author)
	input.Category = strings.TrimSpace(input.Category)
	input.Tags = normalizeUniqueStrings(input.Tags)
	input.Citations = normalizeArticleCitations(input.Citations)
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
	input.Category = strings.TrimSpace(input.Category)
	input.Tags = normalizeUniqueStrings(input.Tags)
	input.ProviderName = strings.TrimSpace(input.ProviderName)
	input.PriceLabel = strings.TrimSpace(input.PriceLabel)
	input.Platforms = normalizeUniqueStrings(input.Platforms)
	input.Features = normalizeUniqueStrings(input.Features)
	input.UseCases = normalizeUniqueStrings(input.UseCases)
	input.Limitations = normalizeUniqueStrings(input.Limitations)
	input.SourceURL = strings.TrimSpace(input.SourceURL)
	input.Status = strings.TrimSpace(input.Status)
	if input.Status == "" {
		input.Status = StatusDraft
	}
	return input
}

func normalizeUniqueStrings(values []string) []string {
	result := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func normalizeArticleCitations(values []ArticleCitation) []ArticleCitation {
	result := make([]ArticleCitation, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for index, value := range values {
		value.ID = strings.TrimSpace(value.ID)
		value.Label = strings.TrimSpace(value.Label)
		value.SourceName = strings.TrimSpace(value.SourceName)
		value.SourceURL = strings.TrimSpace(value.SourceURL)
		value.Excerpt = strings.TrimSpace(value.Excerpt)
		value.PublishedAt = strings.TrimSpace(value.PublishedAt)
		if value.ID == "" {
			value.ID = fmt.Sprintf("source-%d", index+1)
		}
		if value.Label == "" || value.SourceName == "" || value.SourceURL == "" {
			continue
		}
		if _, exists := seen[value.ID]; exists {
			continue
		}
		seen[value.ID] = struct{}{}
		result = append(result, value)
	}
	return result
}

func normalizeQRVariants(values []CommunityQRVariant) []CommunityQRVariant {
	result := make([]CommunityQRVariant, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		value.Key = strings.TrimSpace(value.Key)
		value.Label = strings.TrimSpace(value.Label)
		value.Description = strings.TrimSpace(value.Description)
		value.ImageURL = strings.TrimSpace(value.ImageURL)
		value.JoinURL = strings.TrimSpace(value.JoinURL)
		value.Status = strings.TrimSpace(value.Status)
		if value.Status == "" {
			value.Status = StatusDraft
		}
		if value.Key == "" || value.Label == "" || !validStatus(value.Status) {
			continue
		}
		if _, exists := seen[value.Key]; exists {
			continue
		}
		seen[value.Key] = struct{}{}
		result = append(result, value)
	}
	return result
}

func validStatus(status string) bool {
	return status == StatusDraft || status == StatusPublished
}
