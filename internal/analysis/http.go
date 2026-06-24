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

func writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrServiceNotReady):
		c.JSON(http.StatusInternalServerError, gin.H{"error": "service_not_ready"})
	case errors.Is(err, ErrSessionNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "session_not_found"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
	}
}
