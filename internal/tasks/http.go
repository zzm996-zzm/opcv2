package tasks

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/zzm/opcv2/internal/auth"
	"github.com/zzm/opcv2/internal/platform/httpapi"
)

type Application interface {
	CreateTask(ctx context.Context, input CreateInput) (Task, error)
	GenerateTasks(ctx context.Context, input GenerateTasksInput) (GenerateTasksResult, error)
	ListTaskPage(ctx context.Context, userID int64, filters ListFilters) (TaskPage, error)
	ListTaskProjects(ctx context.Context, userID int64) ([]string, error)
	ListTaskTags(ctx context.Context, userID int64) ([]string, error)
	TaskStats(ctx context.Context, userID int64) (Stats, error)
	GetTask(ctx context.Context, userID, id int64) (Task, error)
	UpdateTask(ctx context.Context, userID, id int64, update TaskUpdate) (Task, error)
	DeleteTask(ctx context.Context, userID, id int64) error
	BatchUpdateTaskStatus(ctx context.Context, userID int64, ids []int64, status string) (int, error)
	BatchDeleteTasks(ctx context.Context, userID int64, ids []int64) (int, error)
	ListSubtasks(ctx context.Context, userID, taskID int64) ([]Subtask, error)
	CreateSubtask(ctx context.Context, input CreateSubtaskInput) (Subtask, error)
	UpdateSubtask(ctx context.Context, userID, taskID, id int64, update SubtaskUpdate) (Subtask, error)
	DeleteSubtask(ctx context.Context, userID, taskID, id int64) error
	GetTaskReminder(ctx context.Context, userID, taskID int64) (*TaskReminder, error)
	UpsertTaskReminder(ctx context.Context, input UpsertTaskReminderInput) (TaskReminder, error)
	DeleteTaskReminder(ctx context.Context, userID, taskID int64) error
}

type HTTPHandler struct {
	app Application
}

func NewHTTPHandler(app Application) *HTTPHandler {
	return &HTTPHandler{app: app}
}

func (h *HTTPHandler) Register(router *gin.RouterGroup) {
	router.POST("/tasks", h.createTask)
	router.POST("/tasks/batch", h.createTaskBatch)
	router.POST("/tasks/generate", h.generateTasks)
	router.GET("/tasks", h.listTasks)
	router.GET("/tasks/stats", h.taskStats)
	router.GET("/tasks/projects", h.listTaskProjects)
	router.GET("/tasks/tags", h.listTaskTags)
	router.PATCH("/tasks/batch", h.batchUpdateTaskStatus)
	router.DELETE("/tasks/batch", h.batchDeleteTasks)
	router.GET("/tasks/:id", h.getTask)
	router.PATCH("/tasks/:id", h.updateTask)
	router.DELETE("/tasks/:id", h.deleteTask)
	router.GET("/tasks/:id/subtasks", h.listSubtasks)
	router.POST("/tasks/:id/subtasks", h.createSubtask)
	router.PATCH("/tasks/:id/subtasks/:subtask_id", h.updateSubtask)
	router.DELETE("/tasks/:id/subtasks/:subtask_id", h.deleteSubtask)
	router.GET("/tasks/:id/reminder", h.getTaskReminder)
	router.PUT("/tasks/:id/reminder", h.upsertTaskReminder)
	router.DELETE("/tasks/:id/reminder", h.deleteTaskReminder)
}

func (h *HTTPHandler) createTaskBatch(c *gin.Context) {
	var request BatchCreateInput
	if err := c.ShouldBindJSON(&request); err != nil || len(request.Tasks) == 0 || len(request.Tasks) > 20 {
		httpapi.BadRequest(c, "invalid_request")
		return
	}
	request.UserID = c.GetInt64(auth.UserIDContextKey)
	if request.UserID <= 0 {
		httpapi.BadRequest(c, "invalid_request")
		return
	}
	for index := range request.Tasks {
		item := &request.Tasks[index]
		item.Title = strings.TrimSpace(item.Title)
		item.Description = strings.TrimSpace(item.Description)
		item.Assignee = strings.TrimSpace(item.Assignee)
		item.Project = strings.TrimSpace(item.Project)
		item.SourceType = strings.TrimSpace(item.SourceType)
		item.SourceTitle = strings.TrimSpace(item.SourceTitle)
		item.SourceURL = strings.TrimSpace(item.SourceURL)
		if !validCreateInput(*item) {
			httpapi.BadRequest(c, "invalid_request")
			return
		}
	}
	creator, ok := h.app.(interface {
		CreateTasks(context.Context, BatchCreateInput) ([]Task, error)
	})
	if !ok {
		httpapi.Error(c, http.StatusInternalServerError, "service_not_ready")
		return
	}
	created, err := creator.CreateTasks(c.Request.Context(), request)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"tasks": httpapi.EnsureSlice(created)})
}

func (h *HTTPHandler) batchUpdateTaskStatus(c *gin.Context) {
	var request BatchTaskStatusInput
	if err := c.ShouldBindJSON(&request); err != nil {
		httpapi.BadRequest(c, "invalid_request")
		return
	}
	ids, ok := normalizeBatchTaskIDs(request.IDs)
	if !ok {
		httpapi.BadRequest(c, "invalid_request")
		return
	}
	if !validStatus(request.Status) {
		httpapi.BadRequest(c, "invalid_status")
		return
	}
	count, err := h.app.BatchUpdateTaskStatus(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), ids, request.Status)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"updated": count})
}

func (h *HTTPHandler) batchDeleteTasks(c *gin.Context) {
	var request BatchTaskIDsInput
	if err := c.ShouldBindJSON(&request); err != nil {
		httpapi.BadRequest(c, "invalid_request")
		return
	}
	ids, ok := normalizeBatchTaskIDs(request.IDs)
	if !ok {
		httpapi.BadRequest(c, "invalid_request")
		return
	}
	count, err := h.app.BatchDeleteTasks(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), ids)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": count})
}

func (h *HTTPHandler) generateTasks(c *gin.Context) {
	var request GenerateTasksInput
	if err := c.ShouldBindJSON(&request); err != nil {
		httpapi.BadRequest(c, "invalid_request")
		return
	}
	request.Goal = strings.TrimSpace(request.Goal)
	if request.Goal == "" || len([]rune(request.Goal)) > 2000 {
		httpapi.BadRequest(c, "invalid_goal")
		return
	}
	request.SourceType = strings.TrimSpace(request.SourceType)
	request.SourceTitle = strings.TrimSpace(request.SourceTitle)
	request.SourceURL = strings.TrimSpace(request.SourceURL)
	if !validTaskSource(request.SourceType, request.SourceID, request.SourceTitle, request.SourceURL) {
		httpapi.BadRequest(c, "invalid_task_source")
		return
	}
	request.UserID = c.GetInt64(auth.UserIDContextKey)
	result, err := h.app.GenerateTasks(c.Request.Context(), request)
	if err != nil {
		writeError(c, err)
		return
	}
	result.Tasks = httpapi.EnsureSlice(result.Tasks)
	c.JSON(http.StatusOK, result)
}

func (h *HTTPHandler) createTask(c *gin.Context) {
	var request CreateInput
	if err := c.ShouldBindJSON(&request); err != nil {
		httpapi.BadRequest(c, "invalid_request")
		return
	}
	if !validCreateInput(request) {
		httpapi.BadRequest(c, "invalid_request")
		return
	}
	request.UserID = c.GetInt64(auth.UserIDContextKey)
	task, err := h.app.CreateTask(c.Request.Context(), request)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, task)
}

func validCreateInput(input CreateInput) bool {
	return strings.TrimSpace(input.Title) != "" &&
		strings.TrimSpace(input.Project) != "" &&
		len([]rune(strings.TrimSpace(input.Title))) <= 100 &&
		len([]rune(strings.TrimSpace(input.Description))) <= 1000 &&
		len([]rune(strings.TrimSpace(input.Assignee))) <= 100 &&
		validTags(input.Tags) &&
		validPriority(input.Priority) &&
		validTaskSource(strings.TrimSpace(input.SourceType), input.SourceID, strings.TrimSpace(input.SourceTitle), strings.TrimSpace(input.SourceURL))
}

func (h *HTTPHandler) listTasks(c *gin.Context) {
	limit, ok := httpapi.QueryLimit(c, 20, 100)
	if !ok {
		return
	}
	status := strings.TrimSpace(c.Query("status"))
	if status != "" && !validStatus(status) {
		httpapi.BadRequest(c, "invalid_status")
		return
	}
	priority := strings.TrimSpace(c.Query("priority"))
	if priority != "" && !validPriority(priority) {
		httpapi.BadRequest(c, "invalid_priority")
		return
	}
	offset, ok := httpapi.QueryOffset(c)
	if !ok {
		return
	}
	page, err := h.app.ListTaskPage(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), ListFilters{
		Status:   status,
		Project:  c.Query("project"),
		Priority: priority,
		Tag:      c.Query("tag"),
		Query:    c.Query("q"),
		Limit:    limit,
		Offset:   offset,
	})
	if err != nil {
		writeError(c, err)
		return
	}
	page.Tasks = httpapi.EnsureSlice(page.Tasks)
	c.JSON(http.StatusOK, page)
}

func (h *HTTPHandler) listTaskTags(c *gin.Context) {
	tags, err := h.app.ListTaskTags(c.Request.Context(), c.GetInt64(auth.UserIDContextKey))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"tags": httpapi.EnsureSlice(tags)})
}

func (h *HTTPHandler) listTaskProjects(c *gin.Context) {
	projects, err := h.app.ListTaskProjects(c.Request.Context(), c.GetInt64(auth.UserIDContextKey))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"projects": httpapi.EnsureSlice(projects)})
}

func (h *HTTPHandler) taskStats(c *gin.Context) {
	stats, err := h.app.TaskStats(c.Request.Context(), c.GetInt64(auth.UserIDContextKey))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, stats)
}

func (h *HTTPHandler) getTask(c *gin.Context) {
	id, ok := taskID(c)
	if !ok {
		return
	}
	task, err := h.app.GetTask(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, task)
}

func (h *HTTPHandler) updateTask(c *gin.Context) {
	id, ok := taskID(c)
	if !ok {
		return
	}
	var request TaskUpdate
	if err := c.ShouldBindJSON(&request); err != nil {
		httpapi.BadRequest(c, "invalid_request")
		return
	}
	if !validTaskUpdate(request) {
		httpapi.BadRequest(c, "invalid_request")
		return
	}
	task, err := h.app.UpdateTask(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), id, request)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, task)
}

func (h *HTTPHandler) deleteTask(c *gin.Context) {
	id, ok := taskID(c)
	if !ok {
		return
	}
	if err := h.app.DeleteTask(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), id); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func validTaskUpdate(update TaskUpdate) bool {
	if update.Title != nil {
		title := strings.TrimSpace(*update.Title)
		if title == "" || len([]rune(title)) > 100 {
			return false
		}
	}
	if update.Project != nil && strings.TrimSpace(*update.Project) == "" {
		return false
	}
	if update.Description != nil && len([]rune(strings.TrimSpace(*update.Description))) > 1000 {
		return false
	}
	if update.Assignee != nil && len([]rune(strings.TrimSpace(*update.Assignee))) > 100 {
		return false
	}
	if update.Tags != nil && !validTags(*update.Tags) {
		return false
	}
	if update.Status != nil && !validStatus(*update.Status) {
		return false
	}
	if update.Priority != nil && !validPriority(*update.Priority) {
		return false
	}
	if update.ClearDueAt && update.DueAt != nil {
		return false
	}
	return true
}

func validTags(tags []string) bool {
	if len(tags) > 10 {
		return false
	}
	for _, tag := range tags {
		if len([]rune(strings.TrimSpace(tag))) > 30 {
			return false
		}
	}
	return true
}

func validStatus(status string) bool {
	switch status {
	case StatusTodo, StatusInProgress, StatusCompleted, StatusReminder:
		return true
	default:
		return false
	}
}

func validPriority(priority string) bool {
	switch priority {
	case PriorityLow, PriorityMedium, PriorityHigh:
		return true
	default:
		return false
	}
}

func taskID(c *gin.Context) (int64, bool) {
	return positivePathID(c, "id", "invalid_task_id")
}

func positivePathID(c *gin.Context, param, errorCode string) (int64, bool) {
	id, err := strconv.ParseInt(c.Param(param), 10, 64)
	if err != nil || id <= 0 {
		httpapi.BadRequest(c, errorCode)
		return 0, false
	}
	return id, true
}

func writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrTaskNotFound):
		httpapi.Error(c, http.StatusNotFound, "task_not_found")
	case errors.Is(err, ErrSubtaskNotFound):
		httpapi.Error(c, http.StatusNotFound, "subtask_not_found")
	case errors.Is(err, ErrReminderNotFound):
		httpapi.Error(c, http.StatusNotFound, "reminder_not_found")
	case errors.Is(err, ErrInvalidReminderTime):
		httpapi.BadRequest(c, "invalid_remind_at")
	case errors.Is(err, ErrInvalidReminderRecurrence):
		httpapi.BadRequest(c, "invalid_recurrence")
	case errors.Is(err, ErrRecurringReminderRequiresMembership):
		httpapi.Error(c, http.StatusPaymentRequired, "membership_required")
	case errors.Is(err, ErrInvalidGeneratedTasks):
		httpapi.Error(c, http.StatusBadGateway, "invalid_ai_result")
	case errors.Is(err, ErrInvalidTaskSource):
		httpapi.BadRequest(c, "invalid_source")
	case errors.Is(err, ErrInvalidTaskBatch):
		httpapi.BadRequest(c, "invalid_request")
	case errors.Is(err, ErrServiceNotReady):
		httpapi.Error(c, http.StatusInternalServerError, "service_not_ready")
	default:
		httpapi.Error(c, http.StatusInternalServerError, "internal_error")
	}
}
