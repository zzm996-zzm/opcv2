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
	userID   int64
	input    ArticleInput
	article  Article
	articles []Article
	err      error
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

func (a *fakeApplication) ListTools(context.Context) ([]Tool, error) {
	return []Tool{}, a.err
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
