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
	input   CreateInput
	userID  int64
	modelID int64
	limit   int
	model   Model
	models  []Model
	err     error
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
