package leads

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/zzm/opcv2/internal/auth"
	"github.com/zzm/opcv2/internal/platform/httpapi"
)

type Application interface {
	CreateTask(ctx context.Context, input CreateTaskInput) (Task, error)
	ListTasks(ctx context.Context, userID int64, limit int) ([]Task, error)
	GetTask(ctx context.Context, userID, id int64) (TaskDetail, error)
	ListResults(ctx context.Context, userID, taskID int64, limit int) ([]LeadResult, error)
}

type HTTPHandler struct {
	app Application
}

func NewHTTPHandler(app Application) *HTTPHandler {
	return &HTTPHandler{app: app}
}

func (h *HTTPHandler) Register(router *gin.RouterGroup) {
	router.POST("/leads/tasks", h.createTask)
	router.GET("/leads/tasks", h.listTasks)
	router.GET("/leads/tasks/:id", h.getTask)
	router.GET("/leads/tasks/:id/results", h.listResults)
}

func (h *HTTPHandler) createTask(c *gin.Context) {
	var request CreateTaskInput
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
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

func (h *HTTPHandler) listTasks(c *gin.Context) {
	limit, ok := httpapi.QueryLimit(c, 20, 100)
	if !ok {
		return
	}
	tasks, err := h.app.ListTasks(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), limit)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"tasks": httpapi.EnsureSlice(tasks)})
}

func (h *HTTPHandler) getTask(c *gin.Context) {
	id, ok := taskIDParam(c)
	if !ok {
		return
	}
	detail, err := h.app.GetTask(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, detail)
}

func (h *HTTPHandler) listResults(c *gin.Context) {
	id, ok := taskIDParam(c)
	if !ok {
		return
	}
	limit, ok := httpapi.QueryLimit(c, 20, 100)
	if !ok {
		return
	}
	results, err := h.app.ListResults(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), id, limit)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"results": httpapi.EnsureSlice(results)})
}

func taskIDParam(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		httpapi.BadRequest(c, "invalid_task_id")
		return 0, false
	}
	return id, true
}

func writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrInvalidTaskInput):
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_task_input"})
	case errors.Is(err, ErrTaskNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "task_not_found"})
	case errors.Is(err, ErrServiceNotReady):
		c.JSON(http.StatusInternalServerError, gin.H{"error": "service_not_ready"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
	}
}
