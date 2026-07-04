package home

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zzm/opcv2/internal/auth"
	"github.com/zzm/opcv2/internal/platform/httpapi"
)

type Application interface {
	Summary(ctx context.Context, userID int64) (Summary, error)
}

type HTTPHandler struct {
	app Application
}

func NewHTTPHandler(app Application) *HTTPHandler {
	return &HTTPHandler{app: app}
}

func (h *HTTPHandler) Register(router *gin.RouterGroup) {
	router.GET("/home/summary", h.summary)
}

func (h *HTTPHandler) summary(c *gin.Context) {
	summary, err := h.app.Summary(c.Request.Context(), c.GetInt64(auth.UserIDContextKey))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, summary)
}

func writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrUserIDRequired):
		httpapi.Error(c, http.StatusUnauthorized, "unauthorized")
	default:
		httpapi.Error(c, http.StatusInternalServerError, "internal_error")
	}
}
