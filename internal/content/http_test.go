package content

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zzm/opcv2/internal/auth"
)

type fakeApplication struct {
	userID       int64
	input        ArticleInput
	article      Article
	articles     []Article
	tool         Tool
	tools        []Tool
	toolFilters  ToolFilters
	toolSlug     string
	favorite     FavoriteResult
	bookmark     BookmarkResult
	joinInput    CommunityJoinInput
	joinRequest  CommunityJoinRequest
	helpTopics   []HelpTopic
	helpArticles []HelpArticle
	helpArticle  HelpArticle
	helpFilters  HelpArticleFilters
	err          error
}

func (a *fakeApplication) ListArticles(context.Context) ([]Article, error) {
	return a.articles, a.err
}

func (a *fakeApplication) GetArticle(_ context.Context, slug string) (Article, error) {
	a.input.Slug = slug
	return a.article, a.err
}

func (a *fakeApplication) CreateArticle(_ context.Context, userID int64, input ArticleInput) (Article, error) {
	a.userID = userID
	a.input = input
	return a.article, a.err
}

func (a *fakeApplication) BookmarkArticle(_ context.Context, userID int64, slug string) (BookmarkResult, error) {
	a.userID = userID
	a.input.Slug = slug
	return a.bookmark, a.err
}

func (a *fakeApplication) UnbookmarkArticle(_ context.Context, userID int64, slug string) (BookmarkResult, error) {
	a.userID = userID
	a.input.Slug = slug
	return a.bookmark, a.err
}

func (a *fakeApplication) ListTools(_ context.Context, filters ToolFilters) ([]Tool, error) {
	a.toolFilters = filters
	return a.tools, a.err
}

func (a *fakeApplication) GetTool(_ context.Context, slug string) (Tool, error) {
	a.toolSlug = slug
	return a.tool, a.err
}

func (a *fakeApplication) FavoriteTool(_ context.Context, userID int64, slug string) (FavoriteResult, error) {
	a.userID = userID
	a.toolSlug = slug
	return a.favorite, a.err
}

func (a *fakeApplication) UnfavoriteTool(_ context.Context, userID int64, slug string) (FavoriteResult, error) {
	a.userID = userID
	a.toolSlug = slug
	return a.favorite, a.err
}

func (a *fakeApplication) UpsertTool(context.Context, int64, ToolInput) (Tool, error) {
	return Tool{}, a.err
}

func (a *fakeApplication) GetCommunityConfig(context.Context) (CommunityConfig, error) {
	return CommunityConfig{}, a.err
}

func (a *fakeApplication) UpdateCommunityConfig(context.Context, int64, CommunityConfigInput) (CommunityConfig, error) {
	return CommunityConfig{}, a.err
}

func (a *fakeApplication) CreateCommunityJoinRequest(_ context.Context, userID int64, input CommunityJoinInput) (CommunityJoinRequest, error) {
	a.userID = userID
	a.joinInput = input
	return a.joinRequest, a.err
}

func (a *fakeApplication) ListHelpTopics(context.Context) ([]HelpTopic, error) {
	return a.helpTopics, a.err
}

func (a *fakeApplication) ListHelpArticles(_ context.Context, filters HelpArticleFilters) ([]HelpArticle, error) {
	a.helpFilters = filters
	return a.helpArticles, a.err
}

func (a *fakeApplication) GetHelpArticle(_ context.Context, slug string) (HelpArticle, error) {
	a.helpArticle.Slug = slug
	return a.helpArticle, a.err
}

func (a *fakeApplication) GetBrand(context.Context) (BrandContent, error) {
	return BrandContent{Metrics: []BrandMetric{}, Cases: []BrandCase{}}, a.err
}

func (a *fakeApplication) UpsertBrandMetric(context.Context, int64, BrandMetricInput) (BrandMetric, error) {
	return BrandMetric{}, a.err
}

func (a *fakeApplication) UpsertBrandCase(context.Context, int64, BrandCaseInput) (BrandCase, error) {
	return BrandCase{}, a.err
}

func contentTestRouter(app Application) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	group := router.Group("/api/v1")
	group.Use(func(c *gin.Context) {
		c.Set(auth.UserIDContextKey, int64(42))
		c.Next()
	})
	NewHTTPHandler(app).RegisterPublic(group)
	NewHTTPHandler(app).RegisterProtected(group)
	NewHTTPHandler(app).RegisterAdmin(group)
	return router
}

func TestListArticlesReturnsEmptyArray(t *testing.T) {
	router := contentTestRouter(&fakeApplication{})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/content/articles", nil)

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if recorder.Body.String() != `{"articles":[]}` {
		t.Fatalf("body = %q, want empty articles array", recorder.Body.String())
	}
}

func TestGetArticleReadsPublicArticleBySlug(t *testing.T) {
	app := &fakeApplication{article: Article{
		ID:          7,
		Slug:        "growth-playbook",
		Title:       "增长手册",
		Summary:     "可执行增长动作",
		Body:        "第一步，明确客户。",
		PublishedAt: time.Date(2026, 6, 24, 9, 0, 0, 0, time.UTC),
	}}
	router := contentTestRouter(app)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/content/articles/growth-playbook", nil)

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.input.Slug != "growth-playbook" {
		t.Fatalf("slug = %q", app.input.Slug)
	}
	if !strings.Contains(recorder.Body.String(), `"title":"增长手册"`) {
		t.Fatalf("body = %s", recorder.Body.String())
	}
}

func TestAdminCreateArticleRejectsNonAdmin(t *testing.T) {
	app := &fakeApplication{err: ErrAdminRequired}
	router := contentTestRouter(app)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/admin/content/articles",
		strings.NewReader(`{"slug":"growth-playbook","title":"增长手册","body":"正文"}`),
	)
	request.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestAdminCreateArticleUsesAuthenticatedUser(t *testing.T) {
	app := &fakeApplication{article: Article{ID: 7, Slug: "growth-playbook", Title: "增长手册", Body: "正文"}}
	router := contentTestRouter(app)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/admin/content/articles",
		strings.NewReader(`{"slug":" growth-playbook ","title":"增长手册","summary":"摘要","body":"正文","status":"published"}`),
	)
	request.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.userID != 42 {
		t.Fatalf("userID = %d, want 42", app.userID)
	}
	if app.input.Slug != " growth-playbook " || app.input.Title != "增长手册" || app.input.Status != "published" {
		t.Fatalf("input = %+v", app.input)
	}
}

func TestListToolsEndpointPassesFilters(t *testing.T) {
	app := &fakeApplication{tools: []Tool{{Slug: "canva-ai", Name: "Canva AI"}}}
	router := contentTestRouter(app)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/content/tools?category=创业获客&q=canva&sort=hot&limit=500", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.toolFilters.Category != "创业获客" || app.toolFilters.Query != "canva" || app.toolFilters.Sort != "hot" || app.toolFilters.Limit != 100 {
		t.Fatalf("filters = %+v", app.toolFilters)
	}
}

func TestGetToolEndpointReadsBySlug(t *testing.T) {
	app := &fakeApplication{tool: Tool{Slug: "canva-ai", Name: "Canva AI"}}
	router := contentTestRouter(app)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/content/tools/canva-ai", nil))

	if recorder.Code != http.StatusOK || app.toolSlug != "canva-ai" {
		t.Fatalf("status/slug/body = %d/%s/%s", recorder.Code, app.toolSlug, recorder.Body.String())
	}
}

func TestFavoriteToolEndpointUsesAuthenticatedUser(t *testing.T) {
	app := &fakeApplication{favorite: FavoriteResult{Slug: "canva-ai", Favorited: true}}
	router := contentTestRouter(app)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/api/v1/content/tools/canva-ai/favorite", nil))

	if recorder.Code != http.StatusOK || app.userID != 42 || app.toolSlug != "canva-ai" {
		t.Fatalf("status/user/slug/body = %d/%d/%s/%s", recorder.Code, app.userID, app.toolSlug, recorder.Body.String())
	}
}

func TestBookmarkArticleEndpointUsesAuthenticatedUser(t *testing.T) {
	app := &fakeApplication{bookmark: BookmarkResult{Slug: "growth-playbook", Bookmarked: true}}
	router := contentTestRouter(app)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/api/v1/content/articles/growth-playbook/bookmark", nil))

	if recorder.Code != http.StatusOK || app.userID != 42 || app.input.Slug != "growth-playbook" {
		t.Fatalf("status/user/slug/body = %d/%d/%s/%s", recorder.Code, app.userID, app.input.Slug, recorder.Body.String())
	}
}

func TestCommunityJoinEndpointUsesAuthenticatedUser(t *testing.T) {
	app := &fakeApplication{joinRequest: CommunityJoinRequest{ID: 9, Status: "submitted"}}
	router := contentTestRouter(app)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/community/join-requests", strings.NewReader(`{"community":"members","contact":"13800138000"}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK || app.userID != 42 || app.joinInput.Community != "members" {
		t.Fatalf("status/user/input/body = %d/%d/%+v/%s", recorder.Code, app.userID, app.joinInput, recorder.Body.String())
	}
}

func TestHelpEndpoints(t *testing.T) {
	app := &fakeApplication{
		helpTopics:   []HelpTopic{{Key: "account", Name: "账号与安全"}},
		helpArticles: []HelpArticle{{Slug: "login-help", Title: "如何登录账号"}},
		helpArticle:  HelpArticle{Title: "如何登录账号", Body: "正文"},
	}
	router := contentTestRouter(app)

	topics := httptest.NewRecorder()
	router.ServeHTTP(topics, httptest.NewRequest(http.MethodGet, "/api/v1/help/topics", nil))
	if topics.Code != http.StatusOK || !strings.Contains(topics.Body.String(), `"topics"`) {
		t.Fatalf("topics status/body = %d/%s", topics.Code, topics.Body.String())
	}

	articles := httptest.NewRecorder()
	router.ServeHTTP(articles, httptest.NewRequest(http.MethodGet, "/api/v1/help/articles?topic=account&q=登录&limit=500", nil))
	if articles.Code != http.StatusOK || app.helpFilters.Limit != 100 {
		t.Fatalf("articles status/filters/body = %d/%+v/%s", articles.Code, app.helpFilters, articles.Body.String())
	}

	detail := httptest.NewRecorder()
	router.ServeHTTP(detail, httptest.NewRequest(http.MethodGet, "/api/v1/help/articles/login-help", nil))
	if detail.Code != http.StatusOK || !strings.Contains(detail.Body.String(), `"slug":"login-help"`) {
		t.Fatalf("detail status/body = %d/%s", detail.Code, detail.Body.String())
	}
}
