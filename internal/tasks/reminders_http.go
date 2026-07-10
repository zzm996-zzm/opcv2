package tasks

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zzm/opcv2/internal/auth"
	"github.com/zzm/opcv2/internal/platform/httpapi"
)

func (h *HTTPHandler) getTaskReminder(c *gin.Context) {
	taskID, ok := taskID(c)
	if !ok {
		return
	}
	reminder, err := h.app.GetTaskReminder(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), taskID)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"reminder": reminder})
}

func (h *HTTPHandler) upsertTaskReminder(c *gin.Context) {
	taskID, ok := taskID(c)
	if !ok {
		return
	}
	var input UpsertTaskReminderInput
	if c.ShouldBindJSON(&input) != nil || input.RemindAt.IsZero() {
		httpapi.BadRequest(c, "invalid_request")
		return
	}
	input.UserID = c.GetInt64(auth.UserIDContextKey)
	input.TaskID = taskID
	reminder, err := h.app.UpsertTaskReminder(c.Request.Context(), input)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, reminder)
}

func (h *HTTPHandler) deleteTaskReminder(c *gin.Context) {
	taskID, ok := taskID(c)
	if !ok {
		return
	}
	if err := h.app.DeleteTaskReminder(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), taskID); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
