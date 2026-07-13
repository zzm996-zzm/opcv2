package content

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/zzm/opcv2/internal/ai"
)

type fakeRepository struct {
	admin          bool
	article        Article
	articles       []Article
	articleFilters ArticleFilters
	tool           Tool
	tools          []Tool
	filters        ToolFilters
	community      CommunityConfig
	favorite       FavoriteResult
	bookmark       BookmarkResult
	joinInput      CommunityJoinInput
	joinRequest    CommunityJoinRequest
	helpTopics     []HelpTopic
	helpArticles   []HelpArticle
	helpArticle    HelpArticle
	err            error
}

func (r *fakeRepository) IsAdmin(context.Context, int64) (bool, error) {
	return r.admin, r.err
}

func (r *fakeRepository) ListArticles(_ context.Context, filters ArticleFilters) ([]Article, error) {
	r.articleFilters = filters
	return r.articles, r.err
}

func (r *fakeRepository) GetArticle(context.Context, string) (Article, error) {
	return r.article, r.err
}

type fakeGenerator struct {
	content []byte
	request ai.GenerateJSONRequest
	err     error
}

func (g *fakeGenerator) GenerateJSON(_ context.Context, request ai.GenerateJSONRequest) (ai.GenerateJSONResult, error) {
	g.request = request
	if g.err != nil {
		return ai.GenerateJSONResult{}, g.err
	}
	if request.Validate != nil {
		if err := request.Validate(g.content); err != nil {
			return ai.GenerateJSONResult{}, err
		}
	}
	return ai.GenerateJSONResult{Content: g.content}, nil
}

func (r *fakeRepository) UpsertArticle(_ context.Context, article Article) (Article, error) {
	r.article = article
	return article, r.err
}

func (r *fakeRepository) BookmarkArticle(context.Context, int64, string) (BookmarkResult, error) {
	return r.bookmark, r.err
}

func (r *fakeRepository) UnbookmarkArticle(context.Context, int64, string) (BookmarkResult, error) {
	return r.bookmark, r.err
}

func (r *fakeRepository) ListTools(_ context.Context, filters ToolFilters) ([]Tool, error) {
	r.filters = filters
	return r.tools, r.err
}

func (r *fakeRepository) GetTool(context.Context, string) (Tool, error) {
	return r.tool, r.err
}

func (r *fakeRepository) FavoriteTool(context.Context, int64, string) (FavoriteResult, error) {
	return r.favorite, r.err
}

func (r *fakeRepository) UnfavoriteTool(context.Context, int64, string) (FavoriteResult, error) {
	return r.favorite, r.err
}

func (r *fakeRepository) UpsertTool(_ context.Context, tool Tool) (Tool, error) {
	r.tool = tool
	return tool, r.err
}

func (r *fakeRepository) GetCommunityConfig(context.Context) (CommunityConfig, error) {
	return r.community, r.err
}

func (r *fakeRepository) UpsertCommunityConfig(_ context.Context, config CommunityConfig) (CommunityConfig, error) {
	r.community = config
	return config, r.err
}

func (r *fakeRepository) CreateCommunityJoinRequest(_ context.Context, userID int64, input CommunityJoinInput) (CommunityJoinRequest, error) {
	r.joinInput = input
	r.joinRequest.UserID = userID
	return r.joinRequest, r.err
}

func (r *fakeRepository) ListHelpTopics(context.Context) ([]HelpTopic, error) {
	return r.helpTopics, r.err
}

func (r *fakeRepository) ListHelpArticles(context.Context, HelpArticleFilters) ([]HelpArticle, error) {
	return r.helpArticles, r.err
}

func (r *fakeRepository) GetHelpArticle(context.Context, string) (HelpArticle, error) {
	return r.helpArticle, r.err
}

func (r *fakeRepository) ListBrandMetrics(context.Context) ([]BrandMetric, error) {
	return nil, r.err
}

func (r *fakeRepository) UpsertBrandMetric(context.Context, BrandMetric) (BrandMetric, error) {
	return BrandMetric{}, r.err
}

func (r *fakeRepository) ListBrandCases(context.Context) ([]BrandCase, error) {
	return nil, r.err
}

func (r *fakeRepository) UpsertBrandCase(context.Context, BrandCase) (BrandCase, error) {
	return BrandCase{}, r.err
}

func TestServiceRejectsNonAdminArticleWrite(t *testing.T) {
	service := NewService(&fakeRepository{admin: false})

	_, err := service.CreateArticle(context.Background(), 42, ArticleInput{
		Slug:  "growth-playbook",
		Title: "增长手册",
		Body:  "正文",
	})

	if err != ErrAdminRequired {
		t.Fatalf("err = %v, want ErrAdminRequired", err)
	}
}

func TestServiceNormalizesArticleAndAllowsAdmin(t *testing.T) {
	repository := &fakeRepository{admin: true}
	service := NewService(repository)
	now := time.Date(2026, 6, 24, 9, 0, 0, 0, time.UTC)
	service.now = func() time.Time { return now }

	article, err := service.CreateArticle(context.Background(), 42, ArticleInput{
		Slug:   " growth-playbook ",
		Title:  " 增长手册 ",
		Body:   " 正文 ",
		Status: StatusPublished,
	})

	if err != nil {
		t.Fatalf("CreateArticle() error = %v", err)
	}
	if article.Slug != "growth-playbook" || article.Title != "增长手册" || !article.PublishedAt.Equal(now) {
		t.Fatalf("article = %+v", article)
	}
}

func TestServiceEmptyListsAreArrays(t *testing.T) {
	service := NewService(&fakeRepository{})

	articles, err := service.ListArticles(context.Background(), ArticleFilters{})

	if err != nil {
		t.Fatalf("ListArticles() error = %v", err)
	}
	if articles == nil || len(articles) != 0 {
		t.Fatalf("articles = %#v, want empty slice", articles)
	}
}

func TestServiceNormalizesArticleFilters(t *testing.T) {
	repository := &fakeRepository{}
	service := NewService(repository)

	_, err := service.ListArticles(context.Background(), ArticleFilters{Category: " 行业趋势 ", Query: " 客服 ", Limit: 500})

	if err != nil {
		t.Fatalf("ListArticles() error = %v", err)
	}
	if repository.articleFilters.Category != "行业趋势" || repository.articleFilters.Query != "客服" || repository.articleFilters.Limit != 100 {
		t.Fatalf("filters = %+v", repository.articleFilters)
	}
}

func TestServiceListsToolsWithFilters(t *testing.T) {
	repository := &fakeRepository{tools: []Tool{{Slug: "canva-ai", Name: "Canva AI"}}}
	service := NewService(repository)

	tools, err := service.ListTools(context.Background(), ToolFilters{
		Category: " 创业获客 ",
		Query:    " canva ",
		Sort:     "hot",
		Limit:    500,
	})

	if err != nil {
		t.Fatalf("ListTools() error = %v", err)
	}
	if len(tools) != 1 || repository.filters.Category != "创业获客" || repository.filters.Query != "canva" || repository.filters.Limit != 100 {
		t.Fatalf("tools/filters = %+v/%+v", tools, repository.filters)
	}
}

func TestServiceFavoritesToolForUser(t *testing.T) {
	repository := &fakeRepository{favorite: FavoriteResult{Slug: "canva-ai", Favorited: true}}
	service := NewService(repository)

	result, err := service.FavoriteTool(context.Background(), 42, " canva-ai ")

	if err != nil {
		t.Fatalf("FavoriteTool() error = %v", err)
	}
	if result.Slug != "canva-ai" || !result.Favorited {
		t.Fatalf("result = %+v", result)
	}
}

func TestServiceRecommendsOnlyCatalogMatches(t *testing.T) {
	repository := &fakeRepository{tools: []Tool{{Slug: "canva-ai", Name: "Canva AI"}}}
	service := NewService(repository)

	result, err := service.RecommendTools(context.Background(), ToolRecommendationInput{
		Goal: " 获取更多线索 ", Scenario: " 社媒海报 ", Category: " 创业获客 ", Limit: 30,
	})

	if err != nil {
		t.Fatalf("RecommendTools() error = %v", err)
	}
	if result.Basis != "catalog_match" || len(result.Tools) != 1 || repository.filters.Query != "社媒海报" || repository.filters.Limit != 20 {
		t.Fatalf("result/filters = %+v/%+v", result, repository.filters)
	}
}

func TestServiceAnswersInsightQuestionWithStoredCitations(t *testing.T) {
	repository := &fakeRepository{article: Article{
		Slug: "growth-playbook", Title: "增长手册", Summary: "摘要", Body: "正文",
		Citations: []ArticleCitation{{ID: "source-1", Label: "行业报告", SourceName: "研究机构", SourceURL: "https://example.com/source", Excerpt: "市场增长"}},
	}}
	generator := &fakeGenerator{content: []byte(`{"answer":"市场仍在增长。","citation_ids":["growth-playbook:source-1"],"assumptions":["仅覆盖已选文章"]}`)}
	service := NewService(repository, WithGenerator(generator))

	answer, err := service.AnswerInsightQuestion(context.Background(), 42, InsightQuestionInput{
		Question: " 市场趋势是什么？ ", ArticleSlugs: []string{" growth-playbook ", "growth-playbook"},
	})

	if err != nil {
		t.Fatalf("AnswerInsightQuestion() error = %v", err)
	}
	if answer.Basis != "catalog_citations" || len(answer.Citations) != 1 || answer.Citations[0].SourceURL != "https://example.com/source" {
		t.Fatalf("answer = %+v", answer)
	}
	if generator.request.Feature != "content.insight_qa" || !strings.Contains(generator.request.UserPrompt, "growth-playbook:source-1") {
		t.Fatalf("request = %+v", generator.request)
	}
}

func TestServiceRejectsInsightQuestionWithoutStoredSources(t *testing.T) {
	service := NewService(&fakeRepository{article: Article{Slug: "growth-playbook"}}, WithGenerator(&fakeGenerator{}))

	_, err := service.AnswerInsightQuestion(context.Background(), 42, InsightQuestionInput{
		Question: "趋势？", ArticleSlugs: []string{"growth-playbook"},
	})

	if err != ErrInsightSourcesEmpty {
		t.Fatalf("err = %v, want ErrInsightSourcesEmpty", err)
	}
}

func TestServiceNormalizesCommunityQRVariants(t *testing.T) {
	repository := &fakeRepository{admin: true}
	service := NewService(repository)

	config, err := service.UpdateCommunityConfig(context.Background(), 42, CommunityConfigInput{
		Headline: " 社群 ", QRVariants: []CommunityQRVariant{
			{Key: " members ", Label: " 免费社群 ", ImageURL: " /qr/members.png ", Status: "published"},
			{Key: "members", Label: "重复", Status: "published"},
		},
	})

	if err != nil {
		t.Fatalf("UpdateCommunityConfig() error = %v", err)
	}
	if len(config.QRVariants) != 1 || config.QRVariants[0].Key != "members" || config.QRVariants[0].ImageURL != "/qr/members.png" {
		t.Fatalf("config = %+v", config)
	}
}

func TestServiceBookmarksArticleForUser(t *testing.T) {
	repository := &fakeRepository{bookmark: BookmarkResult{Slug: "growth-playbook", Bookmarked: true}}
	service := NewService(repository)

	result, err := service.BookmarkArticle(context.Background(), 42, " growth-playbook ")

	if err != nil {
		t.Fatalf("BookmarkArticle() error = %v", err)
	}
	if result.Slug != "growth-playbook" || !result.Bookmarked {
		t.Fatalf("result = %+v", result)
	}
}

func TestServiceCreatesCommunityJoinRequest(t *testing.T) {
	repository := &fakeRepository{joinRequest: CommunityJoinRequest{ID: 9, Status: "submitted"}}
	service := NewService(repository)

	request, err := service.CreateCommunityJoinRequest(context.Background(), 42, CommunityJoinInput{
		Community: " members ",
		Contact:   " 13800138000 ",
		Note:      " 希望加入 ",
	})

	if err != nil {
		t.Fatalf("CreateCommunityJoinRequest() error = %v", err)
	}
	if request.UserID != 42 || repository.joinInput.Community != "members" || repository.joinInput.Contact != "13800138000" {
		t.Fatalf("request/input = %+v/%+v", request, repository.joinInput)
	}
}

func TestServiceRejectsInvalidCommunityJoinRequest(t *testing.T) {
	service := NewService(&fakeRepository{})

	_, err := service.CreateCommunityJoinRequest(context.Background(), 42, CommunityJoinInput{Community: "other", Contact: "138"})

	if err != ErrInvalidInput {
		t.Fatalf("err = %v, want ErrInvalidInput", err)
	}
}

func TestServiceListsHelpContent(t *testing.T) {
	repository := &fakeRepository{
		helpTopics:   []HelpTopic{{Key: "account", Name: "账号与安全"}},
		helpArticles: []HelpArticle{{Slug: "login-help", Title: "如何登录账号"}},
		helpArticle:  HelpArticle{Slug: "login-help", Title: "如何登录账号", Body: "正文"},
	}
	service := NewService(repository)

	topics, err := service.ListHelpTopics(context.Background())
	if err != nil {
		t.Fatalf("ListHelpTopics() error = %v", err)
	}
	articles, err := service.ListHelpArticles(context.Background(), HelpArticleFilters{Topic: " account ", Query: " 登录 ", Limit: 500})
	if err != nil {
		t.Fatalf("ListHelpArticles() error = %v", err)
	}
	article, err := service.GetHelpArticle(context.Background(), " login-help ")
	if err != nil {
		t.Fatalf("GetHelpArticle() error = %v", err)
	}
	if len(topics) != 1 || len(articles) != 1 || article.Body == "" {
		t.Fatalf("topics/articles/article = %+v/%+v/%+v", topics, articles, article)
	}
}
