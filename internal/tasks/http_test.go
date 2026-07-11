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
	input           CreateInput
	generateInput   GenerateTasksInput
	generated       GenerateTasksResult
	userID          int64
	taskID          int64
	deleted         bool
	filters         ListFilters
	update          TaskUpdate
	task            Task
	tasks           []Task
	total           int
	projects        []string
	tags            []string
	stats           Stats
	err             error
	subtasks        []Subtask
	subtaskInput    CreateSubtaskInput
	subtaskID       int64
	subtaskUpdate   SubtaskUpdate
	subtaskDeleted  bool
	reminder        *TaskReminder
	reminderInput   UpsertTaskReminderInput
	reminderDeleted bool
	batchIDs        []int64
	batchStatus     string
	batchCount      int
}

func (a *fakeApplication) CreateTask(_ context.Context, input CreateInput) (Task, error) {
	a.input = input
	return a.task, a.err
}

func (a *fakeApplication) GenerateTasks(_ context.Context, input GenerateTasksInput) (GenerateTasksResult, error) {
	a.generateInput = input
	return a.generated, a.err
}

func (a *fakeApplication) ListTaskPage(_ context.Context, userID int64, filters ListFilters) (TaskPage, error) {
	a.userID = userID
	a.filters = filters
	return TaskPage{Tasks: a.tasks, Total: a.total, Limit: filters.Limit, Offset: filters.Offset}, a.err
}

func (a *fakeApplication) ListTaskProjects(_ context.Context, userID int64) ([]string, error) {
	a.userID = userID
	return a.projects, a.err
}

func (a *fakeApplication) ListTaskTags(_ context.Context, userID int64) ([]string, error) {
	a.userID = userID
	return a.tags, a.err
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

func (a *fakeApplication) BatchUpdateTaskStatus(_ context.Context, userID int64, ids []int64, status string) (int, error) {
	a.userID, a.batchIDs, a.batchStatus = userID, append([]int64(nil), ids...), status
	return a.batchCount, a.err
}

func (a *fakeApplication) BatchDeleteTasks(_ context.Context, userID int64, ids []int64) (int, error) {
	a.userID, a.batchIDs = userID, append([]int64(nil), ids...)
	return a.batchCount, a.err
}

func (a *fakeApplication) ListSubtasks(_ context.Context, userID, taskID int64) ([]Subtask, error) {
	a.userID, a.taskID = userID, taskID
	return a.subtasks, a.err
}
func (a *fakeApplication) CreateSubtask(_ context.Context, input CreateSubtaskInput) (Subtask, error) {
	a.subtaskInput = input
	return Subtask{ID: 1, TaskID: input.TaskID, UserID: input.UserID, Title: input.Title}, a.err
}
func (a *fakeApplication) UpdateSubtask(_ context.Context, userID, taskID, id int64, update SubtaskUpdate) (Subtask, error) {
	a.userID, a.taskID, a.subtaskID, a.subtaskUpdate = userID, taskID, id, update
	return Subtask{ID: id, TaskID: taskID, UserID: userID, Completed: update.Completed != nil && *update.Completed}, a.err
}
func (a *fakeApplication) DeleteSubtask(_ context.Context, userID, taskID, id int64) error {
	a.userID, a.taskID, a.subtaskID, a.subtaskDeleted = userID, taskID, id, true
	return a.err
}

func (a *fakeApplication) GetTaskReminder(_ context.Context, userID, taskID int64) (*TaskReminder, error) {
	a.userID, a.taskID = userID, taskID
	return a.reminder, a.err
}

func (a *fakeApplication) UpsertTaskReminder(_ context.Context, input UpsertTaskReminderInput) (TaskReminder, error) {
	a.reminderInput = input
	return TaskReminder{ID: 8, UserID: input.UserID, TaskID: input.TaskID, RemindAt: input.RemindAt}, a.err
}

func (a *fakeApplication) DeleteTaskReminder(_ context.Context, userID, taskID int64) error {
	a.userID, a.taskID, a.reminderDeleted = userID, taskID, true
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
		"description":"完成首批客户画像并安排访谈",
		"assignee":"李明",
		"project":"AI线索开发",
		"priority":"high",
		"tags":["用户研究","访谈"],
		"tools":["CRM"],
		"learning":"线索评分",
		"source_type":"competitor_scan",
		"source_id":11,
		"source_title":"销售自动化提速",
		"source_url":"/competitor-data"
	}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.input.UserID != 42 || app.input.Title == "" || app.input.Description != "完成首批客户画像并安排访谈" || app.input.Assignee != "李明" || len(app.input.Tags) != 2 || app.input.SourceType != SourceCompetitorScan || app.input.SourceID == nil || *app.input.SourceID != 11 {
		t.Fatalf("input = %+v", app.input)
	}
	if !strings.Contains(recorder.Body.String(), `"status":"todo"`) {
		t.Fatalf("body = %s", recorder.Body.String())
	}
}

func TestCreateTaskEndpointRejectsUnsafeSourceURL(t *testing.T) {
	app := &fakeApplication{}
	router := tasksTestRouter(app)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", strings.NewReader(`{
		"title":"反击竞品更新",
		"project":"竞品动态监测",
		"priority":"high",
		"source_type":"competitor_scan",
		"source_title":"竞品扫描",
		"source_url":"//evil.example/steal"
	}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest || app.input.UserID != 0 {
		t.Fatalf("status/input/body = %d/%+v/%s", recorder.Code, app.input, recorder.Body.String())
	}
}

func TestGenerateTasksEndpointUsesAuthenticatedUser(t *testing.T) {
	app := &fakeApplication{generated: GenerateTasksResult{Tasks: []Task{{ID: 101, UserID: 42, Title: "整理访谈名单", Status: StatusTodo}}}}
	router := tasksTestRouter(app)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/tasks/generate", strings.NewReader(`{"goal":"验证教培客户需求"}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK || app.generateInput.UserID != 42 || app.generateInput.Goal != "验证教培客户需求" || !strings.Contains(recorder.Body.String(), `"tasks"`) {
		t.Fatalf("status/input/body = %d/%+v/%s", recorder.Code, app.generateInput, recorder.Body.String())
	}
}

func TestGenerateTasksEndpointRejectsMissingGoal(t *testing.T) {
	app := &fakeApplication{}
	router := tasksTestRouter(app)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/tasks/generate", strings.NewReader(`{"goal":" "}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest || app.generateInput.UserID != 0 {
		t.Fatalf("status/input/body = %d/%+v/%s", recorder.Code, app.generateInput, recorder.Body.String())
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

func TestCreateTaskEndpointRejectsOversizedAssignee(t *testing.T) {
	app := &fakeApplication{task: Task{ID: 99, UserID: 42, Title: "整理客户名单", Status: StatusTodo}}
	router := tasksTestRouter(app)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", strings.NewReader(`{
		"title":"整理客户名单",
		"project":"AI线索开发",
		"assignee":"`+strings.Repeat("任", 101)+`",
		"priority":"high"
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

func TestCreateTaskEndpointRejectsTooManyTags(t *testing.T) {
	app := &fakeApplication{task: Task{ID: 99, UserID: 42, Title: "整理客户名单", Status: StatusTodo}}
	router := tasksTestRouter(app)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", strings.NewReader(`{
		"title":"整理客户名单",
		"project":"AI线索开发",
		"priority":"high",
		"tags":["1","2","3","4","5","6","7","8","9","10","11"]
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

func TestListTaskTagsEndpointUsesAuthenticatedUser(t *testing.T) {
	app := &fakeApplication{tags: []string{"用户研究", "访谈"}}
	router := tasksTestRouter(app)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/tasks/tags", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK || app.userID != 42 {
		t.Fatalf("status = %d userID = %d body=%s", recorder.Code, app.userID, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"tags":["用户研究","访谈"]`) {
		t.Fatalf("body = %s", recorder.Body.String())
	}
}

func TestListTasksEndpointPassesFilters(t *testing.T) {
	app := &fakeApplication{tasks: []Task{{ID: 99, UserID: 42, Title: "我的任务"}}, total: 21}
	router := tasksTestRouter(app)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/tasks?status=in_progress&project=商业沙盘&priority=high&tag=用户研究&q=接口&limit=10&offset=20", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.filters.Status != StatusInProgress || app.filters.Project != "商业沙盘" || app.filters.Priority != PriorityHigh || app.filters.Tag != "用户研究" || app.filters.Query != "接口" || app.filters.Limit != 10 || app.filters.Offset != 20 {
		t.Fatalf("filters = %+v", app.filters)
	}
	if !strings.Contains(recorder.Body.String(), `"total":21`) || !strings.Contains(recorder.Body.String(), `"offset":20`) {
		t.Fatalf("body = %s", recorder.Body.String())
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

func TestListTaskProjectsEndpointReturnsUserOptions(t *testing.T) {
	app := &fakeApplication{projects: []string{"AI线索开发", "商业沙盘"}}
	router := tasksTestRouter(app)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/tasks/projects", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.userID != 42 || !strings.Contains(recorder.Body.String(), `"projects":["AI线索开发","商业沙盘"]`) {
		t.Fatalf("user/body = %d/%s", app.userID, recorder.Body.String())
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

func TestValidTaskUpdateRejectsOversizedTitle(t *testing.T) {
	title := strings.Repeat("任", 101)
	if validTaskUpdate(TaskUpdate{Title: &title}) {
		t.Fatal("validTaskUpdate() accepted a title longer than 100 characters")
	}
}

func TestValidTaskUpdateRejectsOversizedTag(t *testing.T) {
	tags := []string{strings.Repeat("标", 31)}
	if validTaskUpdate(TaskUpdate{Tags: &tags}) {
		t.Fatal("validTaskUpdate() accepted a tag longer than 30 characters")
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

func TestListSubtasksEndpointReturnsTaskChildren(t *testing.T) {
	app := &fakeApplication{subtasks: []Subtask{{ID: 7, TaskID: 99, UserID: 42, Title: "整理访谈提纲"}}}
	router := tasksTestRouter(app)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/tasks/99/subtasks", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.userID != 42 || app.taskID != 99 || !strings.Contains(recorder.Body.String(), `"title":"整理访谈提纲"`) {
		t.Fatalf("user/task/body = %d/%d/%s", app.userID, app.taskID, recorder.Body.String())
	}
}

func TestCreateSubtaskEndpointUsesParentTaskAndAuthenticatedUser(t *testing.T) {
	app := &fakeApplication{}
	router := tasksTestRouter(app)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/tasks/99/subtasks", strings.NewReader(`{"title":"整理访谈提纲","assignee":"李明"}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.subtaskInput.UserID != 42 || app.subtaskInput.TaskID != 99 || app.subtaskInput.Title != "整理访谈提纲" {
		t.Fatalf("input = %+v", app.subtaskInput)
	}
}

func TestCreateSubtaskEndpointRejectsBlankTitle(t *testing.T) {
	app := &fakeApplication{}
	router := tasksTestRouter(app)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/tasks/99/subtasks", strings.NewReader(`{"title":" "}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest || app.subtaskInput.UserID != 0 {
		t.Fatalf("status = %d input=%+v body=%s", recorder.Code, app.subtaskInput, recorder.Body.String())
	}
}

func TestUpdateSubtaskEndpointUsesSubtaskPathID(t *testing.T) {
	app := &fakeApplication{}
	router := tasksTestRouter(app)
	request := httptest.NewRequest(http.MethodPatch, "/api/v1/tasks/99/subtasks/7", strings.NewReader(`{"completed":true}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.userID != 42 || app.taskID != 99 || app.subtaskID != 7 || app.subtaskUpdate.Completed == nil || !*app.subtaskUpdate.Completed {
		t.Fatalf("user/task/subtask/update = %d/%d/%d/%+v", app.userID, app.taskID, app.subtaskID, app.subtaskUpdate)
	}
}

func TestDeleteSubtaskEndpointUsesSubtaskPathID(t *testing.T) {
	app := &fakeApplication{}
	router := tasksTestRouter(app)
	request := httptest.NewRequest(http.MethodDelete, "/api/v1/tasks/99/subtasks/7", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent || !app.subtaskDeleted || app.userID != 42 || app.taskID != 99 || app.subtaskID != 7 {
		t.Fatalf("status/deleted/user/task/subtask = %d/%t/%d/%d/%d body=%s", recorder.Code, app.subtaskDeleted, app.userID, app.taskID, app.subtaskID, recorder.Body.String())
	}
}

func TestGetTaskReminderEndpointReturnsNullableReminder(t *testing.T) {
	app := &fakeApplication{}
	router := tasksTestRouter(app)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/tasks/99/reminder", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK || app.userID != 42 || app.taskID != 99 || !strings.Contains(recorder.Body.String(), `"reminder":null`) {
		t.Fatalf("status/user/task/body = %d/%d/%d/%s", recorder.Code, app.userID, app.taskID, recorder.Body.String())
	}
}

func TestUpsertTaskReminderEndpointUsesAuthenticatedUser(t *testing.T) {
	app := &fakeApplication{}
	router := tasksTestRouter(app)
	request := httptest.NewRequest(http.MethodPut, "/api/v1/tasks/99/reminder", strings.NewReader(`{"remind_at":"2026-07-18T10:00:00Z"}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK || app.reminderInput.UserID != 42 || app.reminderInput.TaskID != 99 || app.reminderInput.RemindAt.IsZero() {
		t.Fatalf("status/input/body = %d/%+v/%s", recorder.Code, app.reminderInput, recorder.Body.String())
	}
}

func TestUpsertRecurringTaskReminderEndpointReturnsMembershipRequired(t *testing.T) {
	app := &fakeApplication{err: ErrRecurringReminderRequiresMembership}
	router := tasksTestRouter(app)
	request := httptest.NewRequest(http.MethodPut, "/api/v1/tasks/99/reminder", strings.NewReader(`{"remind_at":"2026-07-18T10:00:00Z","recurrence":"daily"}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusPaymentRequired || !strings.Contains(recorder.Body.String(), `"error":"membership_required"`) || app.reminderInput.Recurrence != ReminderRecurrenceDaily {
		t.Fatalf("status/input/body = %d/%+v/%s", recorder.Code, app.reminderInput, recorder.Body.String())
	}
}

func TestUpsertTaskReminderEndpointRejectsMissingTime(t *testing.T) {
	app := &fakeApplication{}
	router := tasksTestRouter(app)
	request := httptest.NewRequest(http.MethodPut, "/api/v1/tasks/99/reminder", strings.NewReader(`{}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest || app.reminderInput.UserID != 0 {
		t.Fatalf("status/input/body = %d/%+v/%s", recorder.Code, app.reminderInput, recorder.Body.String())
	}
}

func TestDeleteTaskReminderEndpointUsesAuthenticatedUser(t *testing.T) {
	app := &fakeApplication{}
	router := tasksTestRouter(app)
	request := httptest.NewRequest(http.MethodDelete, "/api/v1/tasks/99/reminder", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent || !app.reminderDeleted || app.userID != 42 || app.taskID != 99 {
		t.Fatalf("status/deleted/user/task/body = %d/%t/%d/%d/%s", recorder.Code, app.reminderDeleted, app.userID, app.taskID, recorder.Body.String())
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

func TestBatchUpdateTaskStatusEndpointUsesAuthenticatedUser(t *testing.T) {
	app := &fakeApplication{batchCount: 2}
	request := httptest.NewRequest(http.MethodPatch, "/api/v1/tasks/batch", strings.NewReader(`{"ids":[9,7,9],"status":"completed"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	tasksTestRouter(app).ServeHTTP(response, request)

	if response.Code != http.StatusOK || response.Body.String() != `{"updated":2}` {
		t.Fatalf("status/body = %d/%s", response.Code, response.Body.String())
	}
	if app.userID != 42 || app.batchStatus != StatusCompleted || len(app.batchIDs) != 2 {
		t.Fatalf("user/status/ids = %d/%q/%v", app.userID, app.batchStatus, app.batchIDs)
	}
}

func TestBatchDeleteTasksEndpointUsesAuthenticatedUser(t *testing.T) {
	app := &fakeApplication{batchCount: 2}
	request := httptest.NewRequest(http.MethodDelete, "/api/v1/tasks/batch", strings.NewReader(`{"ids":[7,9]}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	tasksTestRouter(app).ServeHTTP(response, request)

	if response.Code != http.StatusOK || response.Body.String() != `{"deleted":2}` {
		t.Fatalf("status/body = %d/%s", response.Code, response.Body.String())
	}
	if app.userID != 42 || len(app.batchIDs) != 2 {
		t.Fatalf("user/ids = %d/%v", app.userID, app.batchIDs)
	}
}

func TestBatchTaskEndpointsRejectInvalidInput(t *testing.T) {
	tests := []struct {
		method string
		body   string
	}{
		{method: http.MethodPatch, body: `{"ids":[],"status":"todo"}`},
		{method: http.MethodPatch, body: `{"ids":[1,0],"status":"todo"}`},
		{method: http.MethodPatch, body: `{"ids":[1],"status":"done"}`},
		{method: http.MethodDelete, body: `{"ids":[]}`},
	}
	for _, test := range tests {
		app := &fakeApplication{}
		request := httptest.NewRequest(test.method, "/api/v1/tasks/batch", strings.NewReader(test.body))
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()

		tasksTestRouter(app).ServeHTTP(response, request)

		if response.Code != http.StatusBadRequest {
			t.Fatalf("%s %s status = %d body=%s", test.method, test.body, response.Code, response.Body.String())
		}
	}
}

func TestBatchUpdateTaskStatusEndpointRejectsInvalidStatus(t *testing.T) {
	app := &fakeApplication{}
	request := httptest.NewRequest(http.MethodPatch, "/api/v1/tasks/batch", strings.NewReader(`{"ids":[1],"status":"done"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	tasksTestRouter(app).ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest || !strings.Contains(response.Body.String(), `"error":"invalid_status"`) {
		t.Fatalf("status/body = %d/%s", response.Code, response.Body.String())
	}
}
