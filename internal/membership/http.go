package membership

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zzm/opcv2/internal/auth"
)

type Application interface {
	CurrentSnapshot(ctx context.Context, userID int64) (Snapshot, error)
	Redeem(ctx context.Context, input RedeemInput) (RedeemResult, error)
}

type HTTPHandler struct {
	app Application
}

func NewHTTPHandler(app Application) *HTTPHandler {
	return &HTTPHandler{app: app}
}

func (h *HTTPHandler) Register(router *gin.RouterGroup) {
	router.GET("/membership/me", h.me)
	router.POST("/redemptions/redeem", h.redeem)
}

func (h *HTTPHandler) me(c *gin.Context) {
	result, err := h.app.CurrentSnapshot(c.Request.Context(), c.GetInt64(auth.UserIDContextKey))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *HTTPHandler) redeem(c *gin.Context) {
	var request struct {
		Code string `json:"code"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}
	result, err := h.app.Redeem(c.Request.Context(), RedeemInput{
		UserID: c.GetInt64(auth.UserIDContextKey),
		Code:   request.Code,
	})
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrInvalidCode):
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_code"})
	case errors.Is(err, ErrCodeNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "code_not_found"})
	case errors.Is(err, ErrCodeExpired):
		c.JSON(http.StatusGone, gin.H{"error": "code_expired"})
	case errors.Is(err, ErrCodeExhausted):
		c.JSON(http.StatusConflict, gin.H{"error": "code_exhausted"})
	case errors.Is(err, ErrUserIDRequired):
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
	}
}
