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
	importInput   ImportLeadInput
	stageInput    UpdateStageInput
	followUpInput RecordFollowUpInput
	copyInput     FollowUpCopyInput
	dueInput      ListDueInput
	customer      Customer
	customers     []Customer
	followUp      FollowUp
	copy          FollowUpCopy
	err           error
}

func (a *fakeApplication) ImportLead(_ context.Context, input ImportLeadInput) (Customer, error) {
	a.importInput = input
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

func (a *fakeApplication) ListDueCustomers(_ context.Context, input ListDueInput) ([]Customer, error) {
	a.dueInput = input
	return a.customers, a.err
}

func (a *fakeApplication) GenerateFollowUpCopy(_ context.Context, input FollowUpCopyInput) (FollowUpCopy, error) {
	a.copyInput = input
	return a.copy, a.err
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
