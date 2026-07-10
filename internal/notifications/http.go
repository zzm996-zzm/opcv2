package notifications

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
	ListNotifications(ctx context.Context, userID int64, filters ListFilters) ([]Notification, error)
	GetNotification(ctx context.Context, userID, id int64) (Notification, error)
	MarkRead(ctx context.Context, userID, id int64) (Notification, error)
	MarkAllRead(ctx context.Context, userID int64) (int, error)
	DeleteNotification(ctx context.Context, userID, id int64) error
	Summary(ctx context.Context, userID int64) (Summary, error)
}

type HTTPHandler struct {
	app Application
}

func NewHTTPHandler(app Application) *HTTPHandler {
	return &HTTPHandler{app: app}
}

func (h *HTTPHandler) Register(router *gin.RouterGroup) {
	router.GET("/notifications", h.listNotifications)
	router.GET("/notifications/:id", h.getNotification)
	router.PATCH("/notifications/:id/read", h.markRead)
	router.POST("/notifications/read-all", h.markAllRead)
	router.DELETE("/notifications/:id", h.deleteNotification)
}

func (h *HTTPHandler) listNotifications(c *gin.Context) {
	limit, ok := httpapi.QueryLimit(c, 20, 100)
	if !ok {
		return
	}
	rows, err := h.app.ListNotifications(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), ListFilters{
		Type:   c.Query("type"),
		Status: c.Query("status"),
		Limit:  limit,
	})
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"notifications": httpapi.EnsureSlice(rows)})
}

func (h *HTTPHandler) getNotification(c *gin.Context) {
	if c.Param("id") == "summary" {
		h.summary(c)
		return
	}
	id, ok := notificationID(c)
	if !ok {
		return
	}
	row, err := h.app.GetNotification(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, row)
}

func (h *HTTPHandler) markRead(c *gin.Context) {
	id, ok := notificationID(c)
	if !ok {
		return
	}
	row, err := h.app.MarkRead(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, row)
}

func (h *HTTPHandler) markAllRead(c *gin.Context) {
	updated, err := h.app.MarkAllRead(c.Request.Context(), c.GetInt64(auth.UserIDContextKey))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"updated": updated})
}

func (h *HTTPHandler) deleteNotification(c *gin.Context) {
	id, ok := notificationID(c)
	if !ok {
		return
	}
	if err := h.app.DeleteNotification(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), id); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *HTTPHandler) summary(c *gin.Context) {
	summary, err := h.app.Summary(c.Request.Context(), c.GetInt64(auth.UserIDContextKey))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, summary)
}

func notificationID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		httpapi.BadRequest(c, "invalid_notification_id")
		return 0, false
	}
	return id, true
}

func writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrInvalidFilter), errors.Is(err, ErrInvalidNotificationID), errors.Is(err, ErrUserIDRequired):
		httpapi.BadRequest(c, "invalid_request")
	case errors.Is(err, ErrNotificationNotFound):
		httpapi.Error(c, http.StatusNotFound, "notification_not_found")
	case errors.Is(err, ErrServiceNotReady):
		httpapi.Error(c, http.StatusInternalServerError, "service_not_ready")
	default:
		httpapi.Error(c, http.StatusInternalServerError, "internal_error")
	}
}
