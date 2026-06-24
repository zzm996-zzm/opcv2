package leads

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
	input  CreateTaskInput
	userID int64
	task   Task
	tasks  []Task
	err    error
}

func (a *fakeApplication) CreateTask(_ context.Context, input CreateTaskInput) (Task, error) {
	a.input = input
	return a.task, a.err
}

func (a *fakeApplication) ListTasks(_ context.Context, userID int64, limit int) ([]Task, error) {
	a.userID = userID
	return a.tasks, a.err
}

func leadTestRouter(app Application) *gin.Engine {
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

func TestCreateTaskEndpointUsesAuthenticatedUser(t *testing.T) {
	app := &fakeApplication{task: Task{ID: 99, UserID: 42, Query: "成都 教培", Status: StatusQueued}}
	router := leadTestRouter(app)
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/leads/tasks",
		strings.NewReader(`{"query":"成都 教培","idempotency_key":"lead-task-42"}`),
	)
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.input.UserID != 42 || app.input.Query != "成都 教培" || app.input.IdempotencyKey != "lead-task-42" {
		t.Fatalf("input = %+v", app.input)
	}
	if !strings.Contains(recorder.Body.String(), `"status":"queued"`) {
		t.Fatalf("body = %s", recorder.Body.String())
	}
}

func TestListTasksEndpointUsesAuthenticatedUser(t *testing.T) {
	app := &fakeApplication{tasks: []Task{{ID: 99, UserID: 42, Query: "成都 教培", Status: StatusSucceeded}}}
	router := leadTestRouter(app)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/leads/tasks", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.userID != 42 {
		t.Fatalf("userID = %d, want 42", app.userID)
	}
	if !strings.Contains(recorder.Body.String(), `"query":"成都 教培"`) {
		t.Fatalf("body = %s", recorder.Body.String())
	}
}
