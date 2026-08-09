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
	input            MatchInput
	userID           int64
	matchID          int64
	result           MatchResult
	sessions         []MatchSession
	session          MatchSession
	favorite         Favorite
	favorites        []Favorite
	err              error
	opportunities    []Opportunity
	opportunity      Opportunity
	filters          OpportunityFilters
	cases            []CaseStudy
	caseStudy        CaseStudy
	caseFilters      CaseFilters
	publicConfig     PublicConfig
	dictionaries     []DictionaryItem
	home             ProjectHome
	projectPage      ProjectPage
	catalogProject   Project
	projectFilters   ProjectFilters
	evidenceCasePage EvidenceCasePage
	evidenceCase     EvidenceCaseDetail
	evidenceFilters  EvidenceCaseFilters
	workflowCreate   CreateProjectMatchInput
	workflowAnswer   AnswerProjectMatchInput
	workflowResponse MatchWorkflowResponse
}

func (a *fakeApplication) GetPublicConfig(context.Context) PublicConfig { return a.publicConfig }
func (a *fakeApplication) ListDictionaryItems(_ context.Context, _ string) ([]DictionaryItem, error) {
	return a.dictionaries, a.err
}
func (a *fakeApplication) GetProjectHome(context.Context) (ProjectHome, error) { return a.home, a.err }
func (a *fakeApplication) ListProjects(_ context.Context, filters ProjectFilters) (ProjectPage, error) {
	a.projectFilters = filters
	return a.projectPage, a.err
}
func (a *fakeApplication) GetProject(context.Context, string) (Project, error) {
	return a.catalogProject, a.err
}
func (a *fakeApplication) ListEvidenceCases(_ context.Context, filters EvidenceCaseFilters) (EvidenceCasePage, error) {
	a.evidenceFilters = filters
	return a.evidenceCasePage, a.err
}
func (a *fakeApplication) GetEvidenceCase(context.Context, string) (EvidenceCaseDetail, error) {
	return a.evidenceCase, a.err
}

func (a *fakeApplication) ListCases(_ context.Context, filters CaseFilters) ([]CaseStudy, error) {
	a.caseFilters = filters
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

func (a *fakeApplication) ListFavoriteMatches(_ context.Context, userID int64, _ int) ([]Favorite, error) {
	a.userID = userID
	return a.favorites, a.err
}

func (a *fakeApplication) UnfavoriteMatch(_ context.Context, userID, id int64) error {
	a.userID = userID
	a.matchID = id
	return a.err
}

func (a *fakeApplication) CreateProjectMatch(_ context.Context, input CreateProjectMatchInput) (MatchWorkflowResponse, error) {
	a.workflowCreate = input
	return a.workflowResponse, a.err
}

func (a *fakeApplication) AnswerProjectMatch(_ context.Context, input AnswerProjectMatchInput) (MatchWorkflowResponse, error) {
	a.workflowAnswer = input
	return a.workflowResponse, a.err
}

func (a *fakeApplication) GetProjectMatch(_ context.Context, userID, id int64) (MatchWorkflowResponse, error) {
	a.userID, a.matchID = userID, id
	return a.workflowResponse, a.err
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

func TestProjectMatchWorkflowEndpointsUseAuthenticatedUserAndIdempotency(t *testing.T) {
	app := &fakeApplication{workflowResponse: MatchWorkflowResponse{MatchID: 200, Status: MatchStatusClarifying, Revision: 1}}
	router := projectTestRouter(app)

	create := httptest.NewRequest(http.MethodPost, "/api/v1/project-matches", strings.NewReader(`{"need":"做内容项目","profile_patch":{"team_size":1}}`))
	create.Header.Set("Content-Type", "application/json")
	create.Header.Set("Idempotency-Key", "create-42-1")
	createRecorder := httptest.NewRecorder()
	router.ServeHTTP(createRecorder, create)
	if createRecorder.Code != http.StatusOK || app.workflowCreate.UserID != 42 || app.workflowCreate.IdempotencyKey != "create-42-1" {
		t.Fatalf("create status/input = %d/%+v body=%s", createRecorder.Code, app.workflowCreate, createRecorder.Body.String())
	}

	answer := httptest.NewRequest(http.MethodPost, "/api/v1/project-matches/200/answer", strings.NewReader(`{"revision":1,"answers":[{"question_id":"budget","field":"budget_band","value":"0-5k"}]}`))
	answer.Header.Set("Content-Type", "application/json")
	answer.Header.Set("Idempotency-Key", "answer-200-1")
	answerRecorder := httptest.NewRecorder()
	router.ServeHTTP(answerRecorder, answer)
	if answerRecorder.Code != http.StatusOK || app.workflowAnswer.UserID != 42 || app.workflowAnswer.MatchID != 200 || app.workflowAnswer.Revision != 1 || app.workflowAnswer.IdempotencyKey != "answer-200-1" {
		t.Fatalf("answer status/input = %d/%+v body=%s", answerRecorder.Code, app.workflowAnswer, answerRecorder.Body.String())
	}

	getRecorder := httptest.NewRecorder()
	router.ServeHTTP(getRecorder, httptest.NewRequest(http.MethodGet, "/api/v1/project-matches/200", nil))
	if getRecorder.Code != http.StatusOK || app.userID != 42 || app.matchID != 200 {
		t.Fatalf("get status/user/match = %d/%d/%d", getRecorder.Code, app.userID, app.matchID)
	}
}

func TestProjectMatchWorkflowReturnsRevisionConflict(t *testing.T) {
	app := &fakeApplication{err: ErrMatchRevisionConflict}
	router := projectTestRouter(app)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/project-matches/200/answer", strings.NewReader(`{"revision":1,"skip":true}`))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusConflict || !strings.Contains(recorder.Body.String(), "match_revision_conflict") {
		t.Fatalf("status/body = %d/%s", recorder.Code, recorder.Body.String())
	}
}

func TestOpportunityEndpointsReturnPublishedCatalog(t *testing.T) {
	app := &fakeApplication{
		opportunities: []Opportunity{{ID: 42, Slug: "ai-sales", Title: "AI销售顾问"}},
		opportunity: Opportunity{
			ID: 42, Slug: "ai-sales", Title: "AI销售顾问",
			Sections: []OpportunitySection{{
				Key: "data", Title: "当前数据", Body: "接口结构化数据", Items: []string{"预算区间"},
				Blocks: []OpportunitySectionBlock{{
					Type: "metrics", Title: "关键指标", Columns: 2,
					Items: []OpportunitySectionItem{{Title: "启动预算", Value: "1万元", Tone: "positive", Progress: 72, Tags: []string{"接口数据"}}},
				}},
			}},
		},
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
	if !strings.Contains(recorder.Body.String(), `"key":"data"`) ||
		!strings.Contains(recorder.Body.String(), `"type":"metrics"`) ||
		!strings.Contains(recorder.Body.String(), `"progress":72`) {
		t.Fatalf("structured blocks missing from body = %s", recorder.Body.String())
	}
}

func TestProjectMarketPublicEndpoints(t *testing.T) {
	app := &fakeApplication{
		publicConfig:   PublicConfig{FeaturePaywallEnabled: false},
		dictionaries:   []DictionaryItem{{Code: "ai", Kind: "sector", NameZH: "人工智能"}},
		home:           ProjectHome{Hero: ProjectHero{Title: "项目超市"}, Featured: []Project{}},
		projectPage:    ProjectPage{Items: []Project{{ID: 42, Slug: "ai-sales", Title: "AI销售顾问"}}, Page: 2, PageSize: 6, Total: 1},
		catalogProject: Project{ID: 42, Slug: "ai-sales", Title: "AI销售顾问", LockedBlocks: []string{}, IsUnlocked: true},
	}
	router := projectTestRouter(app)

	checks := []struct {
		path string
		body string
	}{
		{path: "/api/v1/config", body: `"feature_paywall_enabled":false`},
		{path: "/api/v1/dicts?kind=sector", body: `"name_zh":"人工智能"`},
		{path: "/api/v1/projects/home", body: `"title":"项目超市"`},
		{path: "/api/v1/projects/42", body: `"slug":"ai-sales"`},
	}
	for _, check := range checks {
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, check.path, nil))
		if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), check.body) {
			t.Fatalf("GET %s status/body = %d/%s", check.path, recorder.Code, recorder.Body.String())
		}
	}

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/projects?keyword=AI&category=service&track=ai&budget=0-5k&difficulty=low&resource=solo&sort=latest&is_featured=true&page=2&page_size=6", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("list status/body = %d/%s", recorder.Code, recorder.Body.String())
	}
	if app.projectFilters.Keyword != "AI" || app.projectFilters.Page != 2 || app.projectFilters.PageSize != 6 || app.projectFilters.Featured == nil || !*app.projectFilters.Featured {
		t.Fatalf("filters = %+v", app.projectFilters)
	}
}

func TestListCasesEndpointFiltersByOpportunitySlug(t *testing.T) {
	app := &fakeApplication{cases: []CaseStudy{{
		ID: 81, Slug: "short-video-first-client", OpportunitySlug: "ai-short-video-studio",
		OpportunityTitle: "AI短视频脚本工作室", Industry: "内容服务", Title: "首个客户", CaseType: "success",
	}}}
	router := projectTestRouter(app)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, httptest.NewRequest(
		http.MethodGet,
		"/api/v1/projects/cases?type=success&opportunity_slug=%20ai-short-video-studio%20&industry=%20%E5%86%85%E5%AE%B9%E6%9C%8D%E5%8A%A1%20",
		nil,
	))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.caseFilters.CaseType != "success" || app.caseFilters.OpportunitySlug != "ai-short-video-studio" || app.caseFilters.Industry != "内容服务" || app.caseFilters.Limit != 20 {
		t.Fatalf("filters = %+v", app.caseFilters)
	}
	if !strings.Contains(recorder.Body.String(), `"slug":"short-video-first-client"`) ||
		!strings.Contains(recorder.Body.String(), `"opportunity_title":"AI短视频脚本工作室"`) ||
		!strings.Contains(recorder.Body.String(), `"industry":"内容服务"`) {
		t.Fatalf("body = %s", recorder.Body.String())
	}
}

func TestEvidenceCaseEndpointsUseIndependentPRDContract(t *testing.T) {
	app := &fakeApplication{
		evidenceCasePage: EvidenceCasePage{Items: []EvidenceCaseItem{{ID: 81, Title: "失败案例", Type: "fail"}}, Page: 2, PageSize: 10, Total: 1},
		evidenceCase: EvidenceCaseDetail{
			EvidenceCaseItem: EvidenceCaseItem{ID: 81, Title: "失败案例", Type: "fail"},
			Facts:            []EvidenceFact{{Field: "died_year", Value: "2024", SourceRefs: []int64{91}}},
			Sources:          []EvidenceSource{{ID: 91, URL: "https://example.com/case"}},
		},
	}
	router := projectTestRouter(app)

	list := httptest.NewRecorder()
	router.ServeHTTP(list, httptest.NewRequest(http.MethodGet, "/api/v1/project-cases?type=fail&industry=AI&scale=solo&page=2&page_size=10", nil))
	if list.Code != http.StatusOK || !strings.Contains(list.Body.String(), `"items"`) || !strings.Contains(list.Body.String(), `"title":"失败案例"`) {
		t.Fatalf("list status/body = %d/%s", list.Code, list.Body.String())
	}
	if app.evidenceFilters.CaseType != "fail" || app.evidenceFilters.Page != 2 || app.evidenceFilters.PageSize != 10 {
		t.Fatalf("filters = %+v", app.evidenceFilters)
	}

	detail := httptest.NewRecorder()
	router.ServeHTTP(detail, httptest.NewRequest(http.MethodGet, "/api/v1/project-cases/81", nil))
	if detail.Code != http.StatusOK || !strings.Contains(detail.Body.String(), `"facts"`) || !strings.Contains(detail.Body.String(), `"source_refs":[91]`) {
		t.Fatalf("detail status/body = %d/%s", detail.Code, detail.Body.String())
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

func TestListAndDeleteFavoriteMatchEndpoints(t *testing.T) {
	app := &fakeApplication{favorites: []Favorite{{ID: 7, UserID: 42, SessionID: 99, Session: &MatchSession{ID: 99, Intent: "线上轻资产"}}}}
	router := projectTestRouter(app)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/projects/favorites", nil))
	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), `"intent":"线上轻资产"`) {
		t.Fatalf("list status/body = %d/%s", recorder.Code, recorder.Body.String())
	}

	recorder = httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodDelete, "/api/v1/projects/matches/99/favorite", nil))
	if recorder.Code != http.StatusNoContent || app.userID != 42 || app.matchID != 99 {
		t.Fatalf("delete status/user/match = %d/%d/%d", recorder.Code, app.userID, app.matchID)
	}
}
