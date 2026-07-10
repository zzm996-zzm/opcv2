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
	ListTasks(ctx context.Context, userID int64, filters ListFilters) ([]Task, error)
	TaskStats(ctx context.Context, userID int64) (Stats, error)
	GetTask(ctx context.Context, userID, id int64) (Task, error)
	UpdateTask(ctx context.Context, userID, id int64, update TaskUpdate) (Task, error)
	DeleteTask(ctx context.Context, userID, id int64) error
}

type HTTPHandler struct {
	app Application
}

func NewHTTPHandler(app Application) *HTTPHandler {
	return &HTTPHandler{app: app}
}

func (h *HTTPHandler) Register(router *gin.RouterGroup) {
	router.POST("/tasks", h.createTask)
	router.GET("/tasks", h.listTasks)
	router.GET("/tasks/stats", h.taskStats)
	router.GET("/tasks/:id", h.getTask)
	router.PATCH("/tasks/:id", h.updateTask)
	router.DELETE("/tasks/:id", h.deleteTask)
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
		validPriority(input.Priority)
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
	tasks, err := h.app.ListTasks(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), ListFilters{
		Status:  status,
		Project: c.Query("project"),
		Query:   c.Query("q"),
		Limit:   limit,
	})
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"tasks": httpapi.EnsureSlice(tasks)})
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
	if update.Title != nil && strings.TrimSpace(*update.Title) == "" {
		return false
	}
	if update.Project != nil && strings.TrimSpace(*update.Project) == "" {
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
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		httpapi.BadRequest(c, "invalid_task_id")
		return 0, false
	}
	return id, true
}

func writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrTaskNotFound):
		httpapi.Error(c, http.StatusNotFound, "task_not_found")
	case errors.Is(err, ErrServiceNotReady):
		httpapi.Error(c, http.StatusInternalServerError, "service_not_ready")
	default:
		httpapi.Error(c, http.StatusInternalServerError, "internal_error")
	}
}
