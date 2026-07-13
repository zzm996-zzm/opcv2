package projects

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/zzm/opcv2/internal/auth"
)

type fakeApplication struct {
	input         MatchInput
	userID        int64
	matchID       int64
	result        MatchResult
	sessions      []MatchSession
	session       MatchSession
	favorite      Favorite
	err           error
	opportunities []Opportunity
	opportunity   Opportunity
	filters       OpportunityFilters
	cases         []CaseStudy
	caseStudy     CaseStudy
}

func (a *fakeApplication) ListCases(_ context.Context, _ CaseFilters) ([]CaseStudy, error) {
	return a.cases, a.err
}
func (a *fakeApplication) GetCase(_ context.Context, _ string) (CaseStudy, error) {
	return a.caseStudy, a.err
}

func (a *fakeApplication) ListOpportunities(_ context.Context, filters OpportunityFilters) ([]Opportunity, error) {
	a.filters = filters
	return a.opportunities, a.err
}
func (a *fakeApplication) GetOpportunity(_ context.Context, _ string) (Opportunity, error) {
	return a.opportunity, a.err
}

func (a *fakeApplication) CreateMatch(_ context.Context, input MatchInput) (MatchResult, error) {
	a.input = input
	return a.result, a.err
}

func (a *fakeApplication) ListMatches(_ context.Context, userID int64, limit int) ([]MatchSession, error) {
	a.userID = userID
	return a.sessions, a.err
}

func (a *fakeApplication) GetMatch(_ context.Context, userID, id int64) (MatchSession, error) {
	a.userID = userID
	a.matchID = id
	return a.session, a.err
}
func (a *fakeApplication) AnswerMatch(_ context.Context, input AnswerMatchInput) (MatchResult, error) {
	a.input.UserID = input.UserID
	a.matchID = input.SessionID
	return a.result, a.err
}
func (a *fakeApplication) CreateComparison(_ context.Context, input CreateComparisonInput) (Comparison, error) {
	return Comparison{ID: 61, UserID: input.UserID}, a.err
}
func (a *fakeApplication) GetComparison(_ context.Context, userID, id int64) (Comparison, error) {
	return Comparison{ID: id, UserID: userID}, a.err
}
func (a *fakeApplication) CreateExport(_ context.Context, input CreateExportInput) (Export, error) {
	return Export{ID: 71, UserID: input.UserID, Status: "ready"}, a.err
}
func (a *fakeApplication) GetExport(_ context.Context, userID, id int64) (Export, error) {
	return Export{ID: id, UserID: userID, Payload: []byte(`{}`)}, a.err
}

func (a *fakeApplication) FavoriteMatch(_ context.Context, userID, id int64) (Favorite, error) {
	a.userID = userID
	a.matchID = id
	return a.favorite, a.err
}

func projectTestRouter(app Application) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	group := router.Group("/api/v1")
	group.Use(func(c *gin.Context) {
		c.Set(auth.UserIDContextKey, int64(42))
		c.Next()
	})
	NewHTTPHandler(app).Register(group)
	return router
}

func TestCreateMatchEndpointUsesAuthenticatedUser(t *testing.T) {
	app := &fakeApplication{result: MatchResult{
		SessionID: 99,
		Status:    StatusCompleted,
		Projects:  []ProjectMatch{{Title: "AI短视频脚本工作室", Score: 94}},
	}}
	router := projectTestRouter(app)
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/projects/matches",
		strings.NewReader(`{"intent":"我擅长内容创作，预算3万以内，每周20小时"}`),
	)
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.input.UserID != 42 || app.input.Intent == "" {
		t.Fatalf("input = %+v", app.input)
	}
	if !strings.Contains(recorder.Body.String(), `"title":"AI短视频脚本工作室"`) {
		t.Fatalf("body = %s", recorder.Body.String())
	}
}

func TestOpportunityEndpointsReturnPublishedCatalog(t *testing.T) {
	app := &fakeApplication{
		opportunities: []Opportunity{{ID: 42, Slug: "ai-sales", Title: "AI销售顾问"}},
		opportunity:   Opportunity{ID: 42, Slug: "ai-sales", Title: "AI销售顾问"},
	}
	router := projectTestRouter(app)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/projects/opportunities?q=AI&industry=%E4%BC%81%E4%B8%9A%E6%9C%8D%E5%8A%A1", nil))
	if recorder.Code != http.StatusOK || app.filters.Query != "AI" || app.filters.Industry != "企业服务" {
		t.Fatalf("status/filters = %d/%+v body=%s", recorder.Code, app.filters, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"slug":"ai-sales"`) {
		t.Fatalf("body = %s", recorder.Body.String())
	}

	recorder = httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/projects/opportunities/ai-sales", nil))
	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), `"title":"AI销售顾问"`) {
		t.Fatalf("status/body = %d/%s", recorder.Code, recorder.Body.String())
	}
}

func TestListMatchesEndpointUsesAuthenticatedUser(t *testing.T) {
	app := &fakeApplication{sessions: []MatchSession{{ID: 99, UserID: 42, Intent: "我的项目", Status: StatusCompleted}}}
	router := projectTestRouter(app)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/projects/matches", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.userID != 42 {
		t.Fatalf("userID = %d, want 42", app.userID)
	}
	if !strings.Contains(recorder.Body.String(), `"intent":"我的项目"`) {
		t.Fatalf("body = %s", recorder.Body.String())
	}
}

func TestGetMatchEndpointReturnsNotFoundForAnotherUser(t *testing.T) {
	app := &fakeApplication{err: ErrSessionNotFound}
	router := projectTestRouter(app)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/projects/matches/99", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestFavoriteMatchEndpointIsUserScoped(t *testing.T) {
	app := &fakeApplication{favorite: Favorite{ID: 7, UserID: 42, SessionID: 99}}
	router := projectTestRouter(app)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/projects/matches/99/favorite", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.userID != 42 || app.matchID != 99 {
		t.Fatalf("user/match = %d/%d", app.userID, app.matchID)
	}
	if !strings.Contains(recorder.Body.String(), `"session_id":99`) {
		t.Fatalf("body = %s", recorder.Body.String())
	}
}
