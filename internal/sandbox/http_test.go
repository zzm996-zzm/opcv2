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
	messages  []Message
	askInput  AskRoleInput
}

func (a *fakeApplication) ListRoles() []Role { return DefaultRoles() }
func (a *fakeApplication) UpdateSessionDraft(_ context.Context, userID, id int64, update DraftUpdate) (Session, error) {
	a.userID = userID
	a.sessionID = id
	a.session.Roles = nil
	if update.Roles != nil {
		a.session.Roles = *update.Roles
	}
	return a.session, a.err
}
func (a *fakeApplication) AskRole(_ context.Context, input AskRoleInput) (Message, error) {
	a.askInput = input
	return Message{ID: 1, SessionID: input.SessionID, UserID: input.UserID, Role: input.Role, Question: input.Question, Answer: "关注留存"}, a.err
}
func (a *fakeApplication) ListMessages(_ context.Context, userID, sessionID int64) ([]Message, error) {
	a.userID = userID
	a.sessionID = sessionID
	return a.messages, a.err
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
func (a *fakeApplication) RetrySession(ctx context.Context, userID, id int64) (Session, error) {
	return a.RunSession(ctx, userID, id)
}
func (a *fakeApplication) CancelSession(_ context.Context, userID, id int64) (Session, error) {
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

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestRolesAndDraftUpdateEndpoints(t *testing.T) {
	app := &fakeApplication{session: Session{ID: 99, UserID: 42, Status: StatusDraft}}
	router := sandboxTestRouter(app)

	rolesRecorder := httptest.NewRecorder()
	router.ServeHTTP(rolesRecorder, httptest.NewRequest(http.MethodGet, "/api/v1/sandbox/roles", nil))
	if rolesRecorder.Code != http.StatusOK || !strings.Contains(rolesRecorder.Body.String(), `"key":"user"`) {
		t.Fatalf("roles status/body = %d/%s", rolesRecorder.Code, rolesRecorder.Body.String())
	}

	updateRecorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPatch, "/api/v1/sandbox/sessions/99/draft", strings.NewReader(`{"roles":["用户视角","投资人视角"]}`))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(updateRecorder, request)
	if updateRecorder.Code != http.StatusOK || app.userID != 42 || app.sessionID != 99 || len(app.session.Roles) != 2 {
		t.Fatalf("update status/app/body = %d/%+v/%s", updateRecorder.Code, app, updateRecorder.Body.String())
	}
}

func TestRunSessionEndpointUsesAuthenticatedUser(t *testing.T) {
	app := &fakeApplication{session: Session{
		ID:         99,
		UserID:     42,
		Status:     StatusQueued,
		RunAttempt: 1,
	}}
	router := sandboxTestRouter(app)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/sandbox/sessions/99/run", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusAccepted {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.userID != 42 || app.sessionID != 99 {
		t.Fatalf("user/session = %d/%d", app.userID, app.sessionID)
	}
	if !strings.Contains(recorder.Body.String(), `"status":"queued"`) {
		t.Fatalf("body = %s", recorder.Body.String())
	}
}

func TestRetryCancelAndStatusEndpoints(t *testing.T) {
	app := &fakeApplication{session: Session{ID: 99, UserID: 42, Status: StatusCanceled, RunAttempt: 2}}
	router := sandboxTestRouter(app)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/api/v1/sandbox/sessions/99/retry", nil))
	if recorder.Code != http.StatusAccepted || app.sessionID != 99 {
		t.Fatalf("retry status/body = %d/%s", recorder.Code, recorder.Body.String())
	}

	recorder = httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/api/v1/sandbox/sessions/99/cancel", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("cancel status/body = %d/%s", recorder.Code, recorder.Body.String())
	}

	recorder = httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/sandbox/sessions/99/status", nil))
	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), `"run_attempt":2`) {
		t.Fatalf("status status/body = %d/%s", recorder.Code, recorder.Body.String())
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

func TestSandboxRoleMessageEndpointsUseAuthenticatedUser(t *testing.T) {
	app := &fakeApplication{messages: []Message{{ID: 1, SessionID: 99, UserID: 42, Role: "用户视角", Question: "会买吗", Answer: "会先试用"}}}
	router := sandboxTestRouter(app)

	listRecorder := httptest.NewRecorder()
	router.ServeHTTP(listRecorder, httptest.NewRequest(http.MethodGet, "/api/v1/sandbox/sessions/99/messages", nil))
	if listRecorder.Code != http.StatusOK || !strings.Contains(listRecorder.Body.String(), `"messages"`) {
		t.Fatalf("list status/body = %d/%s", listRecorder.Code, listRecorder.Body.String())
	}

	createRecorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/sandbox/sessions/99/messages", strings.NewReader(`{"role":"投资人视角","question":"最关注什么？"}`))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(createRecorder, request)
	if createRecorder.Code != http.StatusOK || app.askInput.UserID != 42 || app.askInput.SessionID != 99 || app.askInput.Role != "投资人视角" {
		t.Fatalf("create status/input/body = %d/%+v/%s", createRecorder.Code, app.askInput, createRecorder.Body.String())
	}
}
