package tasks

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
	input   CreateInput
	userID  int64
	taskID  int64
	deleted bool
	filters ListFilters
	update  TaskUpdate
	task    Task
	tasks   []Task
	stats   Stats
	err     error
}

func (a *fakeApplication) CreateTask(_ context.Context, input CreateInput) (Task, error) {
	a.input = input
	return a.task, a.err
}

func (a *fakeApplication) ListTasks(_ context.Context, userID int64, filters ListFilters) ([]Task, error) {
	a.userID = userID
	a.filters = filters
	return a.tasks, a.err
}

func (a *fakeApplication) TaskStats(_ context.Context, userID int64) (Stats, error) {
	a.userID = userID
	return a.stats, a.err
}

func (a *fakeApplication) GetTask(_ context.Context, userID, id int64) (Task, error) {
	a.userID = userID
	a.taskID = id
	return a.task, a.err
}

func (a *fakeApplication) UpdateTask(_ context.Context, userID, id int64, update TaskUpdate) (Task, error) {
	a.userID = userID
	a.taskID = id
	a.update = update
	return a.task, a.err
}

func (a *fakeApplication) DeleteTask(_ context.Context, userID, id int64) error {
	a.userID = userID
	a.taskID = id
	a.deleted = true
	return a.err
}

func tasksTestRouter(app Application) *gin.Engine {
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
	app := &fakeApplication{task: Task{ID: 99, UserID: 42, Title: "整理客户名单", Status: StatusTodo}}
	router := tasksTestRouter(app)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", strings.NewReader(`{
		"title":"整理客户名单",
		"project":"AI线索开发",
		"priority":"high",
		"tools":["CRM"],
		"learning":"线索评分"
	}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.input.UserID != 42 || app.input.Title == "" {
		t.Fatalf("input = %+v", app.input)
	}
	if !strings.Contains(recorder.Body.String(), `"status":"todo"`) {
		t.Fatalf("body = %s", recorder.Body.String())
	}
}

func TestCreateTaskEndpointRejectsMissingTitle(t *testing.T) {
	app := &fakeApplication{task: Task{ID: 99, UserID: 42, Title: "整理客户名单", Status: StatusTodo}}
	router := tasksTestRouter(app)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", strings.NewReader(`{
		"title":" ",
		"project":"AI线索开发",
		"priority":"high",
		"tools":["CRM"],
		"learning":"线索评分"
	}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.input.UserID != 0 {
		t.Fatalf("CreateTask should not be called, input = %+v", app.input)
	}
}

func TestCreateTaskEndpointRejectsInvalidPriority(t *testing.T) {
	app := &fakeApplication{task: Task{ID: 99, UserID: 42, Title: "整理客户名单", Status: StatusTodo}}
	router := tasksTestRouter(app)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", strings.NewReader(`{
		"title":"整理客户名单",
		"project":"AI线索开发",
		"priority":"urgent",
		"tools":["CRM"],
		"learning":"线索评分"
	}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestListTasksEndpointUsesAuthenticatedUser(t *testing.T) {
	app := &fakeApplication{tasks: []Task{{ID: 99, UserID: 42, Title: "我的任务"}}}
	router := tasksTestRouter(app)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/tasks", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.userID != 42 {
		t.Fatalf("userID = %d, want 42", app.userID)
	}
	if !strings.Contains(recorder.Body.String(), `"tasks"`) {
		t.Fatalf("body = %s", recorder.Body.String())
	}
}

func TestListTasksEndpointPassesFilters(t *testing.T) {
	app := &fakeApplication{tasks: []Task{{ID: 99, UserID: 42, Title: "我的任务"}}}
	router := tasksTestRouter(app)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/tasks?status=in_progress&project=商业沙盘&q=接口&limit=10", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.filters.Status != StatusInProgress || app.filters.Project != "商业沙盘" || app.filters.Query != "接口" || app.filters.Limit != 10 {
		t.Fatalf("filters = %+v", app.filters)
	}
}

func TestListTasksEndpointReturnsEmptyArrayAndCapsLimit(t *testing.T) {
	app := &fakeApplication{}
	router := tasksTestRouter(app)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/tasks?limit=500", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.userID != 42 || app.filters.Limit != 100 {
		t.Fatalf("user/limit = %d/%d", app.userID, app.filters.Limit)
	}
	if !strings.Contains(recorder.Body.String(), `"tasks":[]`) {
		t.Fatalf("body = %s", recorder.Body.String())
	}
}

func TestTaskStatsEndpointUsesAuthenticatedUser(t *testing.T) {
	app := &fakeApplication{stats: Stats{Total: 3, Todo: 1, InProgress: 1, Completed: 1}}
	router := tasksTestRouter(app)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/tasks/stats", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.userID != 42 || !strings.Contains(recorder.Body.String(), `"total":3`) {
		t.Fatalf("user/body = %d/%s", app.userID, recorder.Body.String())
	}
}

func TestUpdateTaskEndpointUsesAuthenticatedUser(t *testing.T) {
	app := &fakeApplication{task: Task{ID: 99, UserID: 42, Title: "整理客户名单", Status: StatusCompleted}}
	router := tasksTestRouter(app)
	request := httptest.NewRequest(http.MethodPatch, "/api/v1/tasks/99", strings.NewReader(`{"status":"completed"}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.userID != 42 || app.taskID != 99 || app.update.Status == nil || *app.update.Status != StatusCompleted {
		t.Fatalf("user/task/update = %d/%d/%+v", app.userID, app.taskID, app.update)
	}
}

func TestUpdateTaskEndpointRejectsInvalidStatus(t *testing.T) {
	app := &fakeApplication{task: Task{ID: 99, UserID: 42, Title: "整理客户名单", Status: StatusTodo}}
	router := tasksTestRouter(app)
	request := httptest.NewRequest(http.MethodPatch, "/api/v1/tasks/99", strings.NewReader(`{"status":"done"}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.update.Status != nil {
		t.Fatalf("UpdateTask should not be called, update = %+v", app.update)
	}
}

func TestGetTaskEndpointReturnsNotFoundForOtherUser(t *testing.T) {
	app := &fakeApplication{err: ErrTaskNotFound}
	router := tasksTestRouter(app)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/tasks/99", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestDeleteTaskEndpointUsesAuthenticatedUser(t *testing.T) {
	app := &fakeApplication{}
	router := tasksTestRouter(app)
	request := httptest.NewRequest(http.MethodDelete, "/api/v1/tasks/99", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if !app.deleted || app.userID != 42 || app.taskID != 99 {
		t.Fatalf("deleted/user/task = %t/%d/%d", app.deleted, app.userID, app.taskID)
	}
}

func TestDeleteTaskEndpointReturnsNotFoundForOtherUser(t *testing.T) {
	app := &fakeApplication{err: ErrTaskNotFound}
	router := tasksTestRouter(app)
	request := httptest.NewRequest(http.MethodDelete, "/api/v1/tasks/99", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
}
