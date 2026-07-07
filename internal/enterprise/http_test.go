package enterprise

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/zzm/opcv2/internal/auth"
	"github.com/zzm/opcv2/internal/crm"
)

type fakeApplication struct {
	overview       Overview
	diagnosisInput DiagnosisRequestInput
	updateInput    DiagnosisRequestUpdateInput
	diagnosis      DiagnosisRequest
	diagnoses      []DiagnosisRequest
	customer       crm.Customer
	userID         int64
	requestID      int64
	limit          int
	err            error
}

func (a *fakeApplication) Overview(_ context.Context, userID int64) (Overview, error) {
	a.userID = userID
	return a.overview, a.err
}

func (a *fakeApplication) CreateDiagnosisRequest(_ context.Context, userID int64, input DiagnosisRequestInput) (DiagnosisRequest, error) {
	a.userID = userID
	a.diagnosisInput = input
	return a.diagnosis, a.err
}

func (a *fakeApplication) ListDiagnosisRequests(_ context.Context, userID int64, limit int) (DiagnosisRequestsResponse, error) {
	a.userID = userID
	a.limit = limit
	return DiagnosisRequestsResponse{Requests: a.diagnoses}, a.err
}

func (a *fakeApplication) UpdateDiagnosisRequest(_ context.Context, userID int64, requestID int64, input DiagnosisRequestUpdateInput) (DiagnosisRequest, error) {
	a.userID = userID
	a.requestID = requestID
	a.updateInput = input
	return a.diagnosis, a.err
}

func (a *fakeApplication) ImportDiagnosisRequestCustomer(_ context.Context, userID int64, requestID int64) (crm.Customer, error) {
	a.userID = userID
	a.requestID = requestID
	return a.customer, a.err
}

func enterpriseTestRouter(app Application) *gin.Engine {
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
		Stats: []Metric{{Key: "companies", Label: "服务企业数", Value: "2"}},
	}}
	router := enterpriseTestRouter(app)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/enterprise/overview", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.userID != 42 || !strings.Contains(recorder.Body.String(), `"stats":[`) {
		t.Fatalf("user/body = %d/%s", app.userID, recorder.Body.String())
	}
}

func TestCreateDiagnosisRequestEndpointUsesAuthenticatedUser(t *testing.T) {
	app := &fakeApplication{diagnosis: DiagnosisRequest{ID: 11, UserID: 42, Need: "30人销售团队需要AI获客陪跑", Status: "submitted"}}
	router := enterpriseTestRouter(app)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/enterprise/diagnosis-requests", strings.NewReader(`{"need":"30人销售团队需要AI获客陪跑"}`))
	request.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.userID != 42 || app.diagnosisInput.Need != "30人销售团队需要AI获客陪跑" || !strings.Contains(recorder.Body.String(), `"status":"submitted"`) {
		t.Fatalf("user/input/body = %d/%+v/%s", app.userID, app.diagnosisInput, recorder.Body.String())
	}
}

func TestListDiagnosisRequestsEndpointUsesAuthenticatedUser(t *testing.T) {
	app := &fakeApplication{diagnoses: []DiagnosisRequest{{ID: 11, UserID: 42, Need: "30人销售团队需要AI获客陪跑", Status: "submitted"}}}
	router := enterpriseTestRouter(app)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/enterprise/diagnosis-requests?limit=5", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.userID != 42 || app.limit != 5 || !strings.Contains(recorder.Body.String(), `"requests":[`) {
		t.Fatalf("user/limit/body = %d/%d/%s", app.userID, app.limit, recorder.Body.String())
	}
}

func TestUpdateDiagnosisRequestEndpointUsesAuthenticatedUser(t *testing.T) {
	app := &fakeApplication{diagnosis: DiagnosisRequest{ID: 11, UserID: 42, Need: "30人销售团队需要AI获客陪跑", Status: "follow_up_created"}}
	router := enterpriseTestRouter(app)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPatch, "/api/v1/enterprise/diagnosis-requests/11", strings.NewReader(`{"status":"follow_up_created"}`))
	request.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.userID != 42 || app.requestID != 11 || app.updateInput.Status != "follow_up_created" || !strings.Contains(recorder.Body.String(), `"status":"follow_up_created"`) {
		t.Fatalf("user/request/input/body = %d/%d/%+v/%s", app.userID, app.requestID, app.updateInput, recorder.Body.String())
	}
}

func TestImportDiagnosisRequestCustomerEndpointUsesAuthenticatedUser(t *testing.T) {
	app := &fakeApplication{customer: crm.Customer{ID: 100, UserID: 42, ImportKey: "enterprise_diagnosis_request:11", Name: "30人销售团队需要AI获客陪跑", Stage: crm.StageWon, Source: crm.SourceEnterprise}}
	router := enterpriseTestRouter(app)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/api/v1/enterprise/diagnosis-requests/11/crm-customer", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.userID != 42 || app.requestID != 11 || !strings.Contains(recorder.Body.String(), `"source":"enterprise"`) {
		t.Fatalf("user/request/body = %d/%d/%s", app.userID, app.requestID, recorder.Body.String())
	}
}
