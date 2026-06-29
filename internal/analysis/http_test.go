package analysis

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type fakeApplication struct {
	input    DirectionInput
	result   DirectionResult
	sessions []Session
	session  Session
	items    []ActionItem
	item     ActionItem
	err      error
}

func (a *fakeApplication) StartDirection(_ context.Context, input DirectionInput) (DirectionResult, error) {
	a.input = input
	return a.result, a.err
}

func (a *fakeApplication) ListSessions(_ context.Context, userID int64, limit int) ([]Session, error) {
	a.input.UserID = userID
	return a.sessions, a.err
}

func (a *fakeApplication) GetSession(_ context.Context, userID, id int64) (Session, error) {
	a.input.UserID = userID
	a.session.ID = id
	return a.session, a.err
}

func (a *fakeApplication) ListActionItems(_ context.Context, userID, sessionID int64) ([]ActionItem, error) {
	a.input.UserID = userID
	a.session.ID = sessionID
	return a.items, a.err
}

func (a *fakeApplication) UpdateActionItem(_ context.Context, userID, sessionID, itemID int64, completed bool) (ActionItem, error) {
	a.input.UserID = userID
	a.session.ID = sessionID
	a.item.ID = itemID
	a.item.Completed = completed
	return a.item, a.err
}

func TestDirectionEndpointUsesAuthenticatedUser(t *testing.T) {
	app := &fakeApplication{result: DirectionResult{
		SessionID: 99,
		Status:    StatusNeedsInput,
		Questions: []Question{{Key: "budget", Text: "启动资金大概多少？"}},
	}}
	router := testRouter(app)
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/analysis/direction",
		strings.NewReader(`{"intent":"想创业"}`),
	)
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.input.UserID != 42 || app.input.Intent != "想创业" {
		t.Fatalf("input = %+v", app.input)
	}
	if !strings.Contains(recorder.Body.String(), `"status":"needs_input"`) {
		t.Fatalf("body = %s", recorder.Body.String())
	}
}

func TestListSessionsEndpointUsesAuthenticatedUser(t *testing.T) {
	app := &fakeApplication{sessions: []Session{
		{ID: 99, UserID: 42, Mode: ModeDirection, Intent: "我的项目", Status: StatusCompleted},
	}}
	router := testRouter(app)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/analysis/sessions", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.input.UserID != 42 {
		t.Fatalf("user id = %d, want 42", app.input.UserID)
	}
	if !strings.Contains(recorder.Body.String(), `"intent":"我的项目"`) {
		t.Fatalf("body = %s", recorder.Body.String())
	}
}

func TestGetSessionEndpointReturnsStoredReport(t *testing.T) {
	app := &fakeApplication{session: Session{
		ID:     99,
		UserID: 42,
		Mode:   ModeDirection,
		Intent: "我的项目",
		Status: StatusCompleted,
		Result: DirectionResult{Cards: []DirectionCard{{Name: "本地教培小班陪跑", Score: 91}}},
	}}
	router := testRouter(app)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/analysis/sessions/99", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.input.UserID != 42 || app.session.ID != 99 {
		t.Fatalf("input user/session = %d/%d", app.input.UserID, app.session.ID)
	}
	if !strings.Contains(recorder.Body.String(), `"name":"本地教培小班陪跑"`) {
		t.Fatalf("body = %s", recorder.Body.String())
	}
}

func TestGetSessionEndpointReturnsNotFoundForMissingSession(t *testing.T) {
	app := &fakeApplication{err: ErrSessionNotFound}
	router := testRouter(app)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/analysis/sessions/99", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestListActionItemsEndpointUsesAuthenticatedUser(t *testing.T) {
	app := &fakeApplication{items: []ActionItem{
		{ID: 7, UserID: 42, SessionID: 99, DayIndex: 1, Title: "整理资源清单"},
	}}
	router := testRouter(app)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/analysis/sessions/99/action-items", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.input.UserID != 42 || app.session.ID != 99 {
		t.Fatalf("input user/session = %d/%d", app.input.UserID, app.session.ID)
	}
	if !strings.Contains(recorder.Body.String(), `"title":"整理资源清单"`) {
		t.Fatalf("body = %s", recorder.Body.String())
	}
}

func TestUpdateActionItemEndpointUsesAuthenticatedUser(t *testing.T) {
	app := &fakeApplication{item: ActionItem{ID: 7, UserID: 42, SessionID: 99, DayIndex: 1, Title: "整理资源清单"}}
	router := testRouter(app)
	request := httptest.NewRequest(
		http.MethodPatch,
		"/api/v1/analysis/sessions/99/action-items/7",
		strings.NewReader(`{"completed":true}`),
	)
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.input.UserID != 42 || app.session.ID != 99 || app.item.ID != 7 || !app.item.Completed {
		t.Fatalf("input user/session/item = %d/%d/%+v", app.input.UserID, app.session.ID, app.item)
	}
	if !strings.Contains(recorder.Body.String(), `"completed":true`) {
		t.Fatalf("body = %s", recorder.Body.String())
	}
}
