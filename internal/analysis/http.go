package analysis

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/zzm/opcv2/internal/auth"
)

type Application interface {
	StartDirection(ctx context.Context, input DirectionInput) (DirectionResult, error)
	ListSessions(ctx context.Context, userID int64, limit int) ([]Session, error)
	GetSession(ctx context.Context, userID, id int64) (Session, error)
	ListActionItems(ctx context.Context, userID, sessionID int64) ([]ActionItem, error)
	UpdateActionItem(ctx context.Context, userID, sessionID, itemID int64, completed bool) (ActionItem, error)
}

type HTTPHandler struct {
	app Application
}

func NewHTTPHandler(app Application) *HTTPHandler {
	return &HTTPHandler{app: app}
}

func (h *HTTPHandler) Register(router *gin.RouterGroup) {
	router.POST("/analysis/direction", h.direction)
	router.GET("/analysis/sessions", h.listSessions)
	router.GET("/analysis/sessions/:id", h.getSession)
	router.GET("/analysis/sessions/:id/action-items", h.listActionItems)
	router.PATCH("/analysis/sessions/:id/action-items/:item_id", h.updateActionItem)
}

func (h *HTTPHandler) direction(c *gin.Context) {
	var request DirectionInput
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}
	request.UserID = c.GetInt64(auth.UserIDContextKey)
	result, err := h.app.StartDirection(c.Request.Context(), request)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *HTTPHandler) listSessions(c *gin.Context) {
	limit := 20
	if value := c.Query("limit"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_limit"})
			return
		}
		limit = parsed
	}
	sessions, err := h.app.ListSessions(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), limit)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"sessions": sessions})
}

func (h *HTTPHandler) getSession(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_session_id"})
		return
	}
	session, err := h.app.GetSession(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, session)
}

func (h *HTTPHandler) listActionItems(c *gin.Context) {
	sessionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || sessionID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_session_id"})
		return
	}
	items, err := h.app.ListActionItems(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), sessionID)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *HTTPHandler) updateActionItem(c *gin.Context) {
	sessionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || sessionID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_session_id"})
		return
	}
	itemID, err := strconv.ParseInt(c.Param("item_id"), 10, 64)
	if err != nil || itemID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_action_item_id"})
		return
	}
	var request struct {
		Completed bool `json:"completed"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}
	item, err := h.app.UpdateActionItem(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), sessionID, itemID, request.Completed)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, item)
}

func writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrServiceNotReady):
		c.JSON(http.StatusInternalServerError, gin.H{"error": "service_not_ready"})
	case errors.Is(err, ErrSessionNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "session_not_found"})
	case errors.Is(err, ErrActionItemNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "action_item_not_found"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
	}
}
