package sandbox

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type fakeProductWorkflowHTTPApp struct {
	*fakeApplication
	analytics SandboxAnalyticsInput
	asked     AskV2RoleInput
	handoff   SandboxTaskHandoffInput
}

func (a *fakeProductWorkflowHTTPApp) RecordSandboxEvent(_ context.Context, input SandboxAnalyticsInput) (SandboxAnalyticsReceipt, error) {
	a.analytics = input
	return SandboxAnalyticsReceipt{EventID: input.EventID, Accepted: true}, nil
}
func (a *fakeProductWorkflowHTTPApp) ListV2FollowUps(context.Context, int64, int64) ([]V2FollowUp, error) {
	return []V2FollowUp{}, nil
}
func (a *fakeProductWorkflowHTTPApp) AskV2Role(_ context.Context, input AskV2RoleInput) (V2FollowUp, error) {
	a.asked = input
	return V2FollowUp{ID: 1, RunID: input.RunID, RoleCode: input.RoleCode, Question: input.Question, Answer: "验证留存"}, nil
}

func (a *fakeProductWorkflowHTTPApp) CreateSandboxTasks(_ context.Context, input SandboxTaskHandoffInput) (SandboxTaskHandoffResult, error) {
	a.handoff = input
	return SandboxTaskHandoffResult{Tasks: []map[string]any{{"id": int64(1), "title": "验证试点", "source_type": "sandbox_session", "source_id": input.RunID}}}, nil
}

func (a *fakeProductWorkflowHTTPApp) CreateSandboxGrowthHandoff(_ context.Context, userID, runID int64) (SandboxGrowthHandoff, error) {
	return SandboxGrowthHandoff{URL: "/growth-calculator?sandbox_run=99", PricingCents: 3900, Channel: "行业伙伴"}, nil
}

func TestSandboxAnalyticsEndpointUsesAuthenticatedUser(t *testing.T) {
	app := &fakeProductWorkflowHTTPApp{fakeApplication: &fakeApplication{}}
	router := sandboxTestRouter(app)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/sandbox/analytics", strings.NewReader(`{"event_id":"sandbox-event-001","event":"sandbox_home_view","visitor_key":"visitor-key-001","route":"/sandbox"}`))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK || app.analytics.UserID != 42 {
		t.Fatalf("status/input/body=%d/%+v/%s", recorder.Code, app.analytics, recorder.Body.String())
	}
}
func TestV2FollowUpEndpointUsesRunAndAuthenticatedUser(t *testing.T) {
	app := &fakeProductWorkflowHTTPApp{fakeApplication: &fakeApplication{}}
	router := sandboxTestRouter(app)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/sandbox-runs/99/follow-ups", strings.NewReader(`{"role_code":"investor","question":"先验证什么？"}`))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK || app.asked.UserID != 42 || app.asked.RunID != 99 {
		t.Fatalf("status/input/body=%d/%+v/%s", recorder.Code, app.asked, recorder.Body.String())
	}
}

func TestV2ReportHandoffEndpointsUseAuthenticatedRun(t *testing.T) {
	app := &fakeProductWorkflowHTTPApp{fakeApplication: &fakeApplication{}}
	router := sandboxTestRouter(app)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/sandbox-runs/99/report/tasks", strings.NewReader(`{"advice_indexes":[0,2]}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusCreated || app.handoff.UserID != 42 || app.handoff.RunID != 99 || len(app.handoff.AdviceIndexes) != 2 {
		t.Fatalf("tasks status/input/body=%d/%+v/%s", recorder.Code, app.handoff, recorder.Body.String())
	}
	recorder = httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/api/v1/sandbox-runs/99/report/growth-handoff", nil))
	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), "sandbox_run=99") {
		t.Fatalf("growth status/body=%d/%s", recorder.Code, recorder.Body.String())
	}
}
