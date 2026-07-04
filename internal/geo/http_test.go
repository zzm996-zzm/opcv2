package geo

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
	overview        Overview
	analysisRequest AnalysisRequest
	requests        []AnalysisRequest
	userID          int64
	requestID       int64
	input           AnalysisRequestInput
	limit           int
	err             error
}

func (a *fakeApplication) Overview(_ context.Context, userID int64) (Overview, error) {
	a.userID = userID
	return a.overview, a.err
}

func (a *fakeApplication) CreateAnalysisRequest(_ context.Context, userID int64, input AnalysisRequestInput) (AnalysisRequest, error) {
	a.userID = userID
	a.input = input
	return a.analysisRequest, a.err
}

func (a *fakeApplication) ListAnalysisRequests(_ context.Context, userID int64, limit int) ([]AnalysisRequest, error) {
	a.userID = userID
	a.limit = limit
	return a.requests, a.err
}

func (a *fakeApplication) GetAnalysisRequest(_ context.Context, userID, id int64) (AnalysisRequest, error) {
	a.userID = userID
	a.requestID = id
	return a.analysisRequest, a.err
}

func geoTestRouter(app Application) *gin.Engine {
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

func TestOverviewEndpointUsesAuthenticatedUser(t *testing.T) {
	app := &fakeApplication{overview: Overview{
		Stats: []Metric{{Key: "coverage", Label: "AI引用覆盖", Value: "12%"}},
	}}
	router := geoTestRouter(app)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/geo/overview", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.userID != 42 || !strings.Contains(recorder.Body.String(), `"stats":[`) {
		t.Fatalf("user/body = %d/%s", app.userID, recorder.Body.String())
	}
}

func TestCreateAnalysisRequestEndpointUsesAuthenticatedUser(t *testing.T) {
	app := &fakeApplication{analysisRequest: AnalysisRequest{
		ID:     9,
		UserID: 42,
		Target: "面向制造业的 AI 质检工具",
		Status: AnalysisRequestStatusQueued,
	}}
	router := geoTestRouter(app)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/api/v1/geo/analysis-requests", strings.NewReader(`{"target":"面向制造业的 AI 质检工具"}`)))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.userID != 42 || app.input.Target != "面向制造业的 AI 质检工具" {
		t.Fatalf("user/input = %d/%+v", app.userID, app.input)
	}
	if !strings.Contains(recorder.Body.String(), `"status":"queued"`) {
		t.Fatalf("body = %s", recorder.Body.String())
	}
}

func TestCreateAnalysisRequestEndpointRejectsInvalidJSON(t *testing.T) {
	router := geoTestRouter(&fakeApplication{})
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/api/v1/geo/analysis-requests", strings.NewReader(`{`)))

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestListAnalysisRequestsEndpointUsesAuthenticatedUser(t *testing.T) {
	app := &fakeApplication{requests: []AnalysisRequest{{
		ID:     7,
		UserID: 42,
		Target: "面向制造业的 AI 质检工具",
		Status: AnalysisRequestStatusQueued,
	}}}
	router := geoTestRouter(app)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/geo/analysis-requests?limit=10", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.userID != 42 || app.limit != 10 {
		t.Fatalf("user/limit = %d/%d", app.userID, app.limit)
	}
	if !strings.Contains(recorder.Body.String(), `"requests":[`) || !strings.Contains(recorder.Body.String(), `"status":"queued"`) {
		t.Fatalf("body = %s", recorder.Body.String())
	}
}

func TestGetAnalysisRequestEndpointUsesAuthenticatedUser(t *testing.T) {
	app := &fakeApplication{analysisRequest: AnalysisRequest{
		ID:     7,
		UserID: 42,
		Target: "面向制造业的 AI 质检工具",
		Status: AnalysisRequestStatusQueued,
	}}
	router := geoTestRouter(app)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/geo/analysis-requests/7", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.userID != 42 || app.requestID != 7 {
		t.Fatalf("user/request = %d/%d", app.userID, app.requestID)
	}
	if !strings.Contains(recorder.Body.String(), `"target":"面向制造业的 AI 质检工具"`) {
		t.Fatalf("body = %s", recorder.Body.String())
	}
}

func TestGetAnalysisRequestEndpointRejectsInvalidID(t *testing.T) {
	router := geoTestRouter(&fakeApplication{})
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/geo/analysis-requests/not-a-number", nil))

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
}
