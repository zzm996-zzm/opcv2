package sandbox

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
	input     CreateInput
	userID    int64
	sessionID int64
	limit     int
	session   Session
	sessions  []Session
	err       error
}

func (a *fakeApplication) CreateSession(_ context.Context, input CreateInput) (Session, error) {
	a.input = input
	return a.session, a.err
}

func (a *fakeApplication) RunSession(_ context.Context, userID, id int64) (Session, error) {
	a.userID = userID
	a.sessionID = id
	return a.session, a.err
}

func (a *fakeApplication) ListSessions(_ context.Context, userID int64, limit int) ([]Session, error) {
	a.userID = userID
	a.limit = limit
	return a.sessions, a.err
}

func (a *fakeApplication) GetSession(_ context.Context, userID, id int64) (Session, error) {
	a.userID = userID
	a.sessionID = id
	return a.session, a.err
}

func sandboxTestRouter(app Application) *gin.Engine {
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

func TestCreateSessionEndpointUsesAuthenticatedUser(t *testing.T) {
	app := &fakeApplication{session: Session{ID: 99, UserID: 42, Status: StatusDraft}}
	router := sandboxTestRouter(app)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/sandbox/sessions", strings.NewReader(`{
		"goal":"验证 AI 客服项目",
		"target_users":"本地教培机构",
		"product":"AI 客服工具",
		"roles":["用户","投资人"]
	}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.input.UserID != 42 || app.input.Goal == "" || len(app.input.Roles) != 2 {
		t.Fatalf("input = %+v", app.input)
	}
}

func TestCreateSessionEndpointRejectsMissingRequiredFields(t *testing.T) {
	app := &fakeApplication{session: Session{ID: 99, UserID: 42, Status: StatusDraft}}
	router := sandboxTestRouter(app)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/sandbox/sessions", strings.NewReader(`{
		"goal":" ",
		"target_users":"本地教培机构",
		"product":"AI 客服工具",
		"roles":["用户"]
	}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.input.Goal != "" {
		t.Fatalf("CreateSession should not be called, input = %+v", app.input)
	}
}

func TestCreateSessionEndpointRejectsBlankRoles(t *testing.T) {
	app := &fakeApplication{session: Session{ID: 99, UserID: 42, Status: StatusDraft}}
	router := sandboxTestRouter(app)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/sandbox/sessions", strings.NewReader(`{
		"goal":"验证 AI 客服项目",
		"target_users":"本地教培机构",
		"product":"AI 客服工具",
		"roles":[" ",""]
	}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestRunSessionEndpointUsesAuthenticatedUser(t *testing.T) {
	app := &fakeApplication{session: Session{
		ID:     99,
		UserID: 42,
		Status: StatusCompleted,
		Report: Report{Score: 83, Summary: "可以验证"},
	}}
	router := sandboxTestRouter(app)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/sandbox/sessions/99/run", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.userID != 42 || app.sessionID != 99 {
		t.Fatalf("user/session = %d/%d", app.userID, app.sessionID)
	}
	if !strings.Contains(recorder.Body.String(), `"score":83`) {
		t.Fatalf("body = %s", recorder.Body.String())
	}
}

func TestListSessionsEndpointReturnsEmptyArrayAndCapsLimit(t *testing.T) {
	app := &fakeApplication{}
	router := sandboxTestRouter(app)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/sandbox/sessions?limit=500", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.userID != 42 || app.limit != 100 {
		t.Fatalf("user/limit = %d/%d", app.userID, app.limit)
	}
	if !strings.Contains(recorder.Body.String(), `"sessions":[]`) {
		t.Fatalf("body = %s", recorder.Body.String())
	}
}

func TestGetSessionEndpointReturnsNotFoundForOtherUser(t *testing.T) {
	app := &fakeApplication{err: ErrSessionNotFound}
	router := sandboxTestRouter(app)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/sandbox/sessions/99", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
}
