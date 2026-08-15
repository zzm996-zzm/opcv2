package tasks

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/zzm/opcv2/internal/auth"
	"github.com/zzm/opcv2/internal/platform/httpapi"
)

func (h *HTTPHandler) listSubtasks(c *gin.Context) {
	taskID, ok := taskID(c)
	if !ok {
		return
	}
	items, err := h.app.ListSubtasks(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), taskID)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"subtasks": httpapi.EnsureSlice(items)})
}

func (h *HTTPHandler) createSubtask(c *gin.Context) {
	taskID, ok := taskID(c)
	if !ok {
		return
	}
	var input CreateSubtaskInput
	if c.ShouldBindJSON(&input) != nil || !validSubtaskTitle(input.Title) || len([]rune(strings.TrimSpace(input.Assignee))) > 100 || (input.ParentSubtaskID != nil && *input.ParentSubtaskID <= 0) {
		httpapi.BadRequest(c, "invalid_request")
		return
	}
	input.UserID, input.TaskID = c.GetInt64(auth.UserIDContextKey), taskID
	item, err := h.app.CreateSubtask(c.Request.Context(), input)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *HTTPHandler) updateSubtask(c *gin.Context) {
	taskID, ok := taskID(c)
	if !ok {
		return
	}
	id, ok := subtaskID(c)
	if !ok {
		return
	}
	var update SubtaskUpdate
	if c.ShouldBindJSON(&update) != nil || (update.Title != nil && !validSubtaskTitle(*update.Title)) || (update.Assignee != nil && len([]rune(strings.TrimSpace(*update.Assignee))) > 100) || (update.ClearDueAt && update.DueAt != nil) {
		httpapi.BadRequest(c, "invalid_request")
		return
	}
	item, err := h.app.UpdateSubtask(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), taskID, id, update)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *HTTPHandler) deleteSubtask(c *gin.Context) {
	taskID, ok := taskID(c)
	if !ok {
		return
	}
	id, ok := subtaskID(c)
	if !ok {
		return
	}
	if err := h.app.DeleteSubtask(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), taskID, id); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func validSubtaskTitle(title string) bool {
	value := strings.TrimSpace(title)
	return value != "" && len([]rune(value)) <= 100
}

func subtaskID(c *gin.Context) (int64, bool) {
	return positivePathID(c, "subtask_id", "invalid_subtask_id")
}
