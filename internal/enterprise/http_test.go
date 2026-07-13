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
	publicOverview PublicOverview
	publicCases    []PublicCase
	publicCase     PublicCase
	contactConfig  ContactConfig
	diagnosisInput DiagnosisRequestInput
	inquiryInput   InquiryInput
	updateInput    DiagnosisRequestUpdateInput
	diagnosis      DiagnosisRequest
	diagnoses      []DiagnosisRequest
	inquiry        Inquiry
	customer       crm.Customer
	userID         int64
	requestID      int64
	slug           string
	limit          int
	err            error
}

func (a *fakeApplication) Overview(_ context.Context, userID int64) (Overview, error) {
	a.userID = userID
	return a.overview, a.err
}

func (a *fakeApplication) PublicOverview(_ context.Context) (PublicOverview, error) {
	return a.publicOverview, a.err
}

func (a *fakeApplication) ListPublicCases(_ context.Context, limit int) (PublicCasesResponse, error) {
	a.limit = limit
	return PublicCasesResponse{Cases: a.publicCases}, a.err
}

func (a *fakeApplication) GetPublicCase(_ context.Context, slug string) (PublicCase, error) {
	a.slug = slug
	return a.publicCase, a.err
}

func (a *fakeApplication) CreateInquiry(_ context.Context, input InquiryInput) (Inquiry, error) {
	a.inquiryInput = input
	return a.inquiry, a.err
}

func (a *fakeApplication) ContactConfig(_ context.Context) (ContactConfig, error) {
	return a.contactConfig, a.err
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

func enterprisePublicTestRouter(app Application) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	group := router.Group("/api/v1")
	NewHTTPHandler(app).RegisterPublic(group)
	return router
}

func TestPublicOverviewEndpointDoesNotRequireAuthenticatedUser(t *testing.T) {
	app := &fakeApplication{publicOverview: PublicOverview{Headline: "企业AI落地陪跑"}}
	router := enterprisePublicTestRouter(app)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/enterprise/public-overview", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"headline":"企业AI落地陪跑"`) {
		t.Fatalf("body = %s", recorder.Body.String())
	}
}

func TestPublicCasesEndpointUsesLimit(t *testing.T) {
	app := &fakeApplication{publicCases: []PublicCase{{Slug: "ai-sales", Title: "AI销售流程搭建"}}}
	router := enterprisePublicTestRouter(app)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/enterprise/cases?limit=6", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.limit != 6 || !strings.Contains(recorder.Body.String(), `"cases":[`) {
		t.Fatalf("limit/body = %d/%s", app.limit, recorder.Body.String())
	}
}

func TestPublicCaseDetailEndpointUsesSlug(t *testing.T) {
	app := &fakeApplication{publicCase: PublicCase{Slug: "ai-sales", Title: "AI销售流程搭建"}}
	router := enterprisePublicTestRouter(app)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/enterprise/cases/ai-sales", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.slug != "ai-sales" || !strings.Contains(recorder.Body.String(), `"slug":"ai-sales"`) {
		t.Fatalf("slug/body = %s/%s", app.slug, recorder.Body.String())
	}
}

func TestCreateInquiryEndpoint(t *testing.T) {
	app := &fakeApplication{inquiry: Inquiry{ID: 10, Company: "启明星教育", Name: "张总", Need: "AI销售陪跑", Status: "crm_synced", CRMCustomerID: 300}}
	router := enterprisePublicTestRouter(app)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/enterprise/inquiries", strings.NewReader(`{"company":"启明星教育","name":"张总","phone":"13800138000","need":"AI销售陪跑"}`))
	request.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.inquiryInput.Company != "启明星教育" || !strings.Contains(recorder.Body.String(), `"crm_customer_id":300`) {
		t.Fatalf("input/body = %+v/%s", app.inquiryInput, recorder.Body.String())
	}
}

func TestContactConfigEndpoint(t *testing.T) {
	app := &fakeApplication{contactConfig: ContactConfig{ConsultantName: "企业顾问", Wechat: "ai-consultant"}}
	router := enterprisePublicTestRouter(app)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/enterprise/contact-config", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"wechat":"ai-consultant"`) {
		t.Fatalf("body = %s", recorder.Body.String())
	}
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
