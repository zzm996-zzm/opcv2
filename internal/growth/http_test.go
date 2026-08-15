package growth

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
	input           CreateInput
	userID          int64
	modelID         int64
	limit           int
	model           Model
	models          []Model
	scenarios       GrowthScenarios
	forecast        GrowthForecast
	recommendations GrowthRecommendations
	risks           GrowthRisks
	actionPlan      GrowthActionPlan
	inputs          GrowthInputs
	err             error
	draft           Draft
	calculation     DraftCalculation
	draftInput      CreateDraftInput
	answerInput     AnswerDraftInput
	calculateInput  CalculateDraftInput
	snapshots       []ModelSnapshot
	recalculate     RecalculateResult
	export          GrowthReportExport
	exportInput     ExportModelInput
	comparison      GrowthComparison
	comparisonInput CompareModelsInput
}

type fakePagedApplication struct {
	*fakeApplication
	pageInput ListModelsInput
	page      ModelPage
}

func (a *fakePagedApplication) ListModelPage(_ context.Context, input ListModelsInput) (ModelPage, error) {
	a.pageInput = input
	return a.page, a.err
}

func (a *fakeApplication) ListSnapshots(_ context.Context, userID, modelID int64, limit int) ([]ModelSnapshot, error) {
	a.userID, a.modelID, a.limit = userID, modelID, limit
	return a.snapshots, a.err
}

func (a *fakeApplication) CreateDraft(_ context.Context, input CreateDraftInput) (Draft, error) {
	a.draftInput = input
	return a.draft, a.err
}

func (a *fakeApplication) GetDraft(_ context.Context, userID, id int64) (Draft, error) {
	a.userID, a.modelID = userID, id
	return a.draft, a.err
}

func (a *fakeApplication) AnswerDraft(_ context.Context, input AnswerDraftInput) (Draft, error) {
	a.answerInput = input
	return a.draft, a.err
}

func (a *fakeApplication) CalculateDraft(_ context.Context, input CalculateDraftInput) (DraftCalculation, error) {
	a.calculateInput = input
	return a.calculation, a.err
}

func (a *fakeApplication) RecalculateModel(_ context.Context, userID, id int64) (RecalculateResult, error) {
	a.userID, a.modelID = userID, id
	return a.recalculate, a.err
}

func (a *fakeApplication) ExportModel(_ context.Context, input ExportModelInput) (GrowthReportExport, error) {
	a.exportInput = input
	a.userID, a.modelID = input.UserID, input.ModelID
	return a.export, a.err
}

func (a *fakeApplication) CompareModels(_ context.Context, input CompareModelsInput) (GrowthComparison, error) {
	a.comparisonInput = input
	a.userID = input.UserID
	return a.comparison, a.err
}

func (a *fakeApplication) CreateModel(_ context.Context, input CreateInput) (Model, error) {
	a.input = input
	return a.model, a.err
}

func (a *fakeApplication) ListModels(_ context.Context, userID int64, limit int) ([]Model, error) {
	a.userID = userID
	a.limit = limit
	return a.models, a.err
}

func (a *fakeApplication) GetModel(_ context.Context, userID, id int64) (Model, error) {
	a.userID = userID
	a.modelID = id
	return a.model, a.err
}

func (a *fakeApplication) ModelScenarios(_ context.Context, userID, id int64) (GrowthScenarios, error) {
	a.userID = userID
	a.modelID = id
	return a.scenarios, a.err
}

func (a *fakeApplication) ModelForecast(_ context.Context, userID, id int64) (GrowthForecast, error) {
	a.userID = userID
	a.modelID = id
	return a.forecast, a.err
}

func (a *fakeApplication) ModelRecommendations(_ context.Context, userID, id int64) (GrowthRecommendations, error) {
	a.userID = userID
	a.modelID = id
	return a.recommendations, a.err
}

func (a *fakeApplication) ModelRisks(_ context.Context, userID, id int64) (GrowthRisks, error) {
	a.userID = userID
	a.modelID = id
	return a.risks, a.err
}

func (a *fakeApplication) ModelActionPlan(_ context.Context, userID, id int64) (GrowthActionPlan, error) {
	a.userID = userID
	a.modelID = id
	return a.actionPlan, a.err
}

func (a *fakeApplication) ModelInputs(_ context.Context, userID, id int64) (GrowthInputs, error) {
	a.userID = userID
	a.modelID = id
	return a.inputs, a.err
}

func growthTestRouter(app Application) *gin.Engine {
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

func TestCreateModelEndpointUsesAuthenticatedUser(t *testing.T) {
	app := &fakeApplication{model: Model{ID: 99, UserID: 42, Name: "标准方案", Result: Result{MonthlyRevenue: 186000}}}
	router := growthTestRouter(app)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/growth/models", strings.NewReader(`{
		"name":"标准方案",
		"monthly_visits":24000,
		"lead_rate":0.068,
		"deal_rate":0.14,
		"average_order":820,
		"acquisition_cost":42,
		"delivery_cost":51000
	}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.input.UserID != 42 || app.input.Name != "标准方案" {
		t.Fatalf("input = %+v", app.input)
	}
}

func TestDraftEndpointsUseAuthenticatedUser(t *testing.T) {
	app := &fakeApplication{
		draft:       Draft{ID: 71, UserID: 42, Status: DraftStatusNeedsInput},
		calculation: DraftCalculation{Draft: Draft{ID: 71, UserID: 42, Status: DraftStatusCalculated}, Model: Model{ID: 99}},
	}
	router := growthTestRouter(app)

	request := httptest.NewRequest(http.MethodPost, "/api/v1/growth/drafts", strings.NewReader(`{"input":"企业培训服务增长测算"}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK || app.draftInput.UserID != 42 {
		t.Fatalf("create status/input = %d/%+v body=%s", recorder.Code, app.draftInput, recorder.Body.String())
	}

	request = httptest.NewRequest(http.MethodPost, "/api/v1/growth/drafts/71/answers", strings.NewReader(`{"answers":{"acquisition_cost":80}}`))
	request.Header.Set("Content-Type", "application/json")
	recorder = httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK || app.answerInput.UserID != 42 || app.answerInput.DraftID != 71 {
		t.Fatalf("answer status/input = %d/%+v body=%s", recorder.Code, app.answerInput, recorder.Body.String())
	}

	request = httptest.NewRequest(http.MethodPost, "/api/v1/growth/drafts/71/calculate", strings.NewReader(`{}`))
	request.Header.Set("Content-Type", "application/json")
	recorder = httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK || app.calculateInput.UserID != 42 || app.calculateInput.DraftID != 71 {
		t.Fatalf("calculate status/input = %d/%+v body=%s", recorder.Code, app.calculateInput, recorder.Body.String())
	}
}

func TestCreateDraftEndpointRejectsInputOver2000Characters(t *testing.T) {
	app := &fakeApplication{}
	router := growthTestRouter(app)
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/growth/drafts",
		strings.NewReader(`{"input":"`+strings.Repeat("业", 2001)+`"}`),
	)
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestListSnapshotsEndpointUsesOwnedModel(t *testing.T) {
	app := &fakeApplication{snapshots: []ModelSnapshot{{ID: 501, UserID: 42, ModelID: 99}}}
	router := growthTestRouter(app)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/growth/models/99/snapshots", nil))
	if recorder.Code != http.StatusOK || app.userID != 42 || app.modelID != 99 || app.limit != 20 {
		t.Fatalf("status/user/model/limit = %d/%d/%d/%d body=%s", recorder.Code, app.userID, app.modelID, app.limit, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"snapshots"`) {
		t.Fatalf("body = %s", recorder.Body.String())
	}
}

func TestRecalculateModelEndpointUsesAuthenticatedUser(t *testing.T) {
	app := &fakeApplication{recalculate: RecalculateResult{Model: Model{ID: 100, UserID: 42, Name: "标准方案 · 再测算"}}}
	router := growthTestRouter(app)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/api/v1/growth/models/99/recalculate", nil))

	if recorder.Code != http.StatusOK || app.userID != 42 || app.modelID != 99 {
		t.Fatalf("status/user/model = %d/%d/%d body=%s", recorder.Code, app.userID, app.modelID, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"model"`) {
		t.Fatalf("body = %s", recorder.Body.String())
	}
}

func TestExportModelEndpointReturnsAttachmentForAuthenticatedUser(t *testing.T) {
	app := &fakeApplication{export: GrowthReportExport{
		Model:      Model{ID: 99, UserID: 42, Name: "标准方案"},
		Disclaimer: "仅供决策参考",
	}}
	router := growthTestRouter(app)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/growth/models/99/export", strings.NewReader(`{"format":"json"}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK || app.exportInput.UserID != 42 || app.exportInput.ModelID != 99 || app.exportInput.Format != "json" {
		t.Fatalf("status/input = %d/%+v body=%s", recorder.Code, app.exportInput, recorder.Body.String())
	}
	if got := recorder.Header().Get("Content-Disposition"); got != `attachment; filename="growth-report-99.json"` {
		t.Fatalf("content disposition = %q", got)
	}
	if !strings.Contains(recorder.Body.String(), `"disclaimer":"仅供决策参考"`) {
		t.Fatalf("body = %s", recorder.Body.String())
	}
}

func TestExportModelEndpointRejectsUnsupportedFormat(t *testing.T) {
	app := &fakeApplication{}
	router := growthTestRouter(app)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/growth/models/99/export", strings.NewReader(`{"format":"pdf"}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest || !strings.Contains(recorder.Body.String(), `"error":"invalid_export_format"`) {
		t.Fatalf("status/body = %d/%s", recorder.Code, recorder.Body.String())
	}
}

func TestCompareModelsEndpointUsesAuthenticatedUser(t *testing.T) {
	app := &fakeApplication{comparison: GrowthComparison{Models: []Model{{ID: 99}, {ID: 100}}}}
	router := growthTestRouter(app)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/growth/models/compare", strings.NewReader(`{"model_ids":[99,100]}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK || app.comparisonInput.UserID != 42 || len(app.comparisonInput.ModelIDs) != 2 || app.comparisonInput.ModelIDs[1] != 100 {
		t.Fatalf("status/input = %d/%+v body=%s", recorder.Code, app.comparisonInput, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"models"`) {
		t.Fatalf("body = %s", recorder.Body.String())
	}
}

func TestCompareModelsEndpointRejectsInvalidSelection(t *testing.T) {
	app := &fakeApplication{err: ErrInvalidComparison}
	router := growthTestRouter(app)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/growth/models/compare", strings.NewReader(`{"model_ids":[99]}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest || !strings.Contains(recorder.Body.String(), `"error":"invalid_comparison"`) {
		t.Fatalf("status/body = %d/%s", recorder.Code, recorder.Body.String())
	}
}

func TestCreateModelEndpointRejectsInvalidAssumptions(t *testing.T) {
	app := &fakeApplication{model: Model{ID: 99, UserID: 42, Name: "标准方案"}}
	router := growthTestRouter(app)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/growth/models", strings.NewReader(`{
		"name":" ",
		"monthly_visits":0,
		"lead_rate":1.2,
		"deal_rate":0.14,
		"average_order":820,
		"acquisition_cost":42,
		"delivery_cost":51000
	}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.input.UserID != 0 {
		t.Fatalf("CreateModel should not be called, input = %+v", app.input)
	}
}

func TestListModelsEndpointUsesAuthenticatedUser(t *testing.T) {
	app := &fakeApplication{models: []Model{{ID: 99, UserID: 42, Name: "标准方案"}}}
	router := growthTestRouter(app)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/growth/models", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.userID != 42 {
		t.Fatalf("userID = %d, want 42", app.userID)
	}
	if !strings.Contains(recorder.Body.String(), `"models"`) {
		t.Fatalf("body = %s", recorder.Body.String())
	}
}

func TestListModelsEndpointReturnsEmptyArrayAndCapsLimit(t *testing.T) {
	app := &fakeApplication{}
	router := growthTestRouter(app)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/growth/models?limit=500", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.userID != 42 || app.limit != 100 {
		t.Fatalf("user/limit = %d/%d", app.userID, app.limit)
	}
	if !strings.Contains(recorder.Body.String(), `"models":[]`) {
		t.Fatalf("body = %s", recorder.Body.String())
	}
}

func TestListModelsEndpointSupportsSearchAndPagination(t *testing.T) {
	app := &fakePagedApplication{
		fakeApplication: &fakeApplication{},
		page:            ModelPage{Models: []Model{{ID: 99, UserID: 42, Name: "SaaS 增长"}}, Total: 31, Limit: 10, Offset: 20},
	}
	router := growthTestRouter(app)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/growth/models?q=SaaS&limit=10&offset=20", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.pageInput.UserID != 42 || app.pageInput.Query != "SaaS" || app.pageInput.Limit != 10 || app.pageInput.Offset != 20 {
		t.Fatalf("input = %+v", app.pageInput)
	}
	if !strings.Contains(recorder.Body.String(), `"total":31`) || !strings.Contains(recorder.Body.String(), `"models"`) {
		t.Fatalf("body = %s", recorder.Body.String())
	}
}

func TestGetModelEndpointReturnsNotFoundForOtherUser(t *testing.T) {
	app := &fakeApplication{err: ErrModelNotFound}
	router := growthTestRouter(app)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/growth/models/99", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestDerivedGrowthEndpointsUseAuthenticatedUserAndModelID(t *testing.T) {
	tests := []struct {
		name string
		path string
		app  *fakeApplication
		want string
	}{
		{
			name: "scenarios",
			path: "/api/v1/growth/models/99/scenarios",
			app:  &fakeApplication{scenarios: GrowthScenarios{ModelID: 99, Scenarios: []GrowthScenario{{Name: "标准方案"}}}},
			want: `"scenarios"`,
		},
		{
			name: "forecast",
			path: "/api/v1/growth/models/99/forecast",
			app:  &fakeApplication{forecast: GrowthForecast{ModelID: 99, Months: []ForecastMonth{{Month: "第3月", Revenue: 186960}}}},
			want: `"months"`,
		},
		{
			name: "recommendations",
			path: "/api/v1/growth/models/99/recommendations",
			app:  &fakeApplication{recommendations: GrowthRecommendations{ModelID: 99, ActionItems: []string{"优先优化成交率"}}},
			want: `"action_items"`,
		},
		{
			name: "risks",
			path: "/api/v1/growth/models/99/risks",
			app:  &fakeApplication{risks: GrowthRisks{ModelID: 99, OverallLevel: "medium", Risks: []GrowthRisk{{Name: "净利润率"}}}},
			want: `"risks"`,
		},
		{
			name: "action plan",
			path: "/api/v1/growth/models/99/action-plan",
			app:  &fakeApplication{actionPlan: GrowthActionPlan{ModelID: 99, Phases: []GrowthActionPhase{{Name: "0–30 天"}}}},
			want: `"phases"`,
		},
		{
			name: "inputs",
			path: "/api/v1/growth/models/99/inputs",
			app:  &fakeApplication{inputs: GrowthInputs{ModelID: 99, CompletenessPercent: 100, Fields: []GrowthInputField{{Key: "deal_rate"}}}},
			want: `"fields"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := growthTestRouter(tt.app)
			recorder := httptest.NewRecorder()

			router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, tt.path, nil))

			if recorder.Code != http.StatusOK {
				t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
			}
			if tt.app.userID != 42 || tt.app.modelID != 99 {
				t.Fatalf("user/model = %d/%d, want 42/99", tt.app.userID, tt.app.modelID)
			}
			if !strings.Contains(recorder.Body.String(), tt.want) {
				t.Fatalf("body = %s", recorder.Body.String())
			}
		})
	}
}
