package crm

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
	importInput    ImportLeadInput
	listInput      ListCustomersInput
	updateInput    UpdateCustomerInput
	stageInput     UpdateStageInput
	followUpInput  RecordFollowUpInput
	followUpsInput ListFollowUpsInput
	copyInput      FollowUpCopyInput
	dueInput       ListDueInput
	customer       Customer
	customers      []Customer
	followUp       FollowUp
	followUps      []FollowUp
	activities     []Activity
	copy           FollowUpCopy
	stats          PipelineStats
	getUserID      int64
	getCustomerID  int64
	statsUserID    int64
	err            error
}

func (a *fakeApplication) ImportLead(_ context.Context, input ImportLeadInput) (Customer, error) {
	a.importInput = input
	return a.customer, a.err
}

func (a *fakeApplication) ListCustomers(_ context.Context, input ListCustomersInput) ([]Customer, error) {
	a.listInput = input
	return a.customers, a.err
}

func (a *fakeApplication) GetCustomer(_ context.Context, userID, customerID int64) (Customer, error) {
	a.getUserID = userID
	a.getCustomerID = customerID
	return a.customer, a.err
}

func (a *fakeApplication) UpdateCustomer(_ context.Context, input UpdateCustomerInput) (Customer, error) {
	a.updateInput = input
	return a.customer, a.err
}

func (a *fakeApplication) UpdateStage(_ context.Context, input UpdateStageInput) (Customer, error) {
	a.stageInput = input
	return a.customer, a.err
}

func (a *fakeApplication) RecordFollowUp(_ context.Context, input RecordFollowUpInput) (FollowUp, error) {
	a.followUpInput = input
	return a.followUp, a.err
}

func (a *fakeApplication) ListActivities(_ context.Context, userID, customerID int64, limit int) ([]Activity, error) {
	a.getUserID = userID
	a.getCustomerID = customerID
	a.dueInput.Limit = limit
	return a.activities, a.err
}

func (a *fakeApplication) ListFollowUps(_ context.Context, input ListFollowUpsInput) ([]FollowUp, error) {
	a.followUpsInput = input
	return a.followUps, a.err
}

func (a *fakeApplication) ListDueCustomers(_ context.Context, input ListDueInput) ([]Customer, error) {
	a.dueInput = input
	return a.customers, a.err
}

func (a *fakeApplication) PipelineStats(_ context.Context, userID int64) (PipelineStats, error) {
	a.statsUserID = userID
	return a.stats, a.err
}

func (a *fakeApplication) GenerateFollowUpCopy(_ context.Context, input FollowUpCopyInput) (FollowUpCopy, error) {
	a.copyInput = input
	return a.copy, a.err
}

func TestListCustomersEndpointUsesAuthenticatedUserAndFilters(t *testing.T) {
	app := &fakeApplication{customers: []Customer{{ID: 100, UserID: 42, Name: "成都启明星教育", Stage: StageContacted}}}
	router := crmTestRouter(app)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/crm/customers?stage=contacted&q=%E5%90%AF%E6%98%8E%E6%98%9F&limit=10", nil)
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body = %s", recorder.Code, recorder.Body.String())
	}
	if app.listInput.UserID != 42 || app.listInput.Stage != StageContacted || app.listInput.Q != "启明星" || app.listInput.Limit != 10 {
		t.Fatalf("input = %+v", app.listInput)
	}
	if !strings.Contains(recorder.Body.String(), `"customers"`) {
		t.Fatalf("body = %s", recorder.Body.String())
	}
}

func TestListCustomersEndpointReturnsEmptyArray(t *testing.T) {
	app := &fakeApplication{}
	router := crmTestRouter(app)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/crm/customers", nil)
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body = %s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"customers":[]`) {
		t.Fatalf("body = %s", recorder.Body.String())
	}
}

func TestGetCustomerEndpointUsesAuthenticatedUser(t *testing.T) {
	app := &fakeApplication{customer: Customer{ID: 100, UserID: 42, Name: "成都启明星教育"}}
	router := crmTestRouter(app)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/crm/customers/100", nil)
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body = %s", recorder.Code, recorder.Body.String())
	}
	if app.getUserID != 42 || app.getCustomerID != 100 {
		t.Fatalf("get user/customer = %d/%d", app.getUserID, app.getCustomerID)
	}
}

func TestUpdateCustomerEndpointMapsCustomerIDAndUser(t *testing.T) {
	app := &fakeApplication{customer: Customer{ID: 100, UserID: 42, Name: "成都启明星教育"}}
	router := crmTestRouter(app)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPatch, "/api/v1/crm/customers/100", strings.NewReader(`{"name":"成都启明星教育","phone":"028-12345678"}`))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body = %s", recorder.Code, recorder.Body.String())
	}
	if app.updateInput.UserID != 42 || app.updateInput.CustomerID != 100 || app.updateInput.Name == nil || *app.updateInput.Name != "成都启明星教育" {
		t.Fatalf("input = %+v", app.updateInput)
	}
}

func TestListActivitiesEndpointUsesAuthenticatedUserAndCustomer(t *testing.T) {
	now := time.Date(2026, 6, 24, 12, 0, 0, 0, time.UTC)
	app := &fakeApplication{activities: []Activity{{ID: 1, UserID: 42, CustomerID: 100, Type: ActivityCustomerUpdated, Note: "客户资料已更新", CreatedAt: now}}}
	router := crmTestRouter(app)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/crm/customers/100/activities?limit=10", nil)
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body = %s", recorder.Code, recorder.Body.String())
	}
	if app.getUserID != 42 || app.getCustomerID != 100 || app.dueInput.Limit != 10 {
		t.Fatalf("user/customer/limit = %d/%d/%d", app.getUserID, app.getCustomerID, app.dueInput.Limit)
	}
	if !strings.Contains(recorder.Body.String(), `"activities"`) {
		t.Fatalf("body = %s", recorder.Body.String())
	}
}

func TestListActivitiesEndpointReturnsEmptyArray(t *testing.T) {
	app := &fakeApplication{}
	router := crmTestRouter(app)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/crm/customers/100/activities", nil)
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body = %s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"activities":[]`) {
		t.Fatalf("body = %s", recorder.Body.String())
	}
}

func TestImportLeadEndpointUsesAuthenticatedUser(t *testing.T) {
	app := &fakeApplication{customer: Customer{ID: 100, UserID: 42, Name: "成都启明星教育"}}
	router := crmTestRouter(app)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/crm/customers/import-lead", strings.NewReader(`{"lead_result_id":99,"name":"成都启明星教育"}`))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body = %s", recorder.Code, recorder.Body.String())
	}
	if app.importInput.UserID != 42 || app.importInput.LeadResultID != 99 {
		t.Fatalf("input = %+v", app.importInput)
	}
}

func TestUpdateStageEndpointMapsCustomerIDAndUser(t *testing.T) {
	app := &fakeApplication{customer: Customer{ID: 100, UserID: 42, Stage: StageContacted}}
	router := crmTestRouter(app)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/crm/customers/100/stage", strings.NewReader(`{"stage":"contacted","note":"电话已接通"}`))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body = %s", recorder.Code, recorder.Body.String())
	}
	if app.stageInput.UserID != 42 || app.stageInput.CustomerID != 100 || app.stageInput.Stage != StageContacted {
		t.Fatalf("input = %+v", app.stageInput)
	}
}

func TestDueCustomersEndpointUsesAuthenticatedUser(t *testing.T) {
	now := time.Date(2026, 6, 24, 12, 0, 0, 0, time.UTC)
	app := &fakeApplication{customers: []Customer{{ID: 100, UserID: 42, Name: "成都启明星教育", NextFollowUpAt: now}}}
	router := crmTestRouter(app)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/crm/customers/due?limit=10", nil)
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body = %s", recorder.Code, recorder.Body.String())
	}
	if app.dueInput.UserID != 42 || app.dueInput.Limit != 10 {
		t.Fatalf("input = %+v", app.dueInput)
	}
}

func TestDueCustomersEndpointReturnsEmptyArray(t *testing.T) {
	app := &fakeApplication{}
	router := crmTestRouter(app)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/crm/customers/due", nil)
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body = %s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"customers":[]`) {
		t.Fatalf("body = %s", recorder.Body.String())
	}
}

func TestListFollowUpsEndpointUsesAuthenticatedUserAndFilters(t *testing.T) {
	now := time.Date(2026, 6, 24, 12, 0, 0, 0, time.UTC)
	app := &fakeApplication{followUps: []FollowUp{{ID: 1, UserID: 42, CustomerID: 100, Note: "发送方案", NextFollowUpAt: now}}}
	router := crmTestRouter(app)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/crm/follow-ups?customer_id=100&limit=10", nil)
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body = %s", recorder.Code, recorder.Body.String())
	}
	if app.followUpsInput.UserID != 42 || app.followUpsInput.CustomerID != 100 || app.followUpsInput.Limit != 10 {
		t.Fatalf("input = %+v", app.followUpsInput)
	}
	if !strings.Contains(recorder.Body.String(), `"follow_ups"`) {
		t.Fatalf("body = %s", recorder.Body.String())
	}
}

func TestListFollowUpsEndpointReturnsEmptyArray(t *testing.T) {
	app := &fakeApplication{}
	router := crmTestRouter(app)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/crm/follow-ups", nil)
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body = %s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"follow_ups":[]`) {
		t.Fatalf("body = %s", recorder.Body.String())
	}
}

func TestPipelineStatsEndpointUsesAuthenticatedUser(t *testing.T) {
	app := &fakeApplication{stats: PipelineStats{Total: 3, New: 1, Won: 1, DueToday: 1}}
	router := crmTestRouter(app)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/crm/pipeline-stats", nil)
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body = %s", recorder.Code, recorder.Body.String())
	}
	if app.statsUserID != 42 {
		t.Fatalf("stats userID = %d, want 42", app.statsUserID)
	}
	if !strings.Contains(recorder.Body.String(), `"total":3`) {
		t.Fatalf("body = %s", recorder.Body.String())
	}
}

func TestFollowUpCopyEndpointMapsCustomerIDAndUser(t *testing.T) {
	app := &fakeApplication{copy: FollowUpCopy{Subject: "跟进方案", Body: "您好，约个时间沟通方案。", Channel: "wechat"}}
	router := crmTestRouter(app)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/crm/customers/100/follow-up-copy", strings.NewReader(`{"goal":"推进方案会"}`))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body = %s", recorder.Code, recorder.Body.String())
	}
	if app.copyInput.UserID != 42 || app.copyInput.CustomerID != 100 || app.copyInput.Goal != "推进方案会" {
		t.Fatalf("input = %+v", app.copyInput)
	}
}

func crmTestRouter(app Application) *gin.Engine {
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
