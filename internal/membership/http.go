package membership

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zzm/opcv2/internal/auth"
	"github.com/zzm/opcv2/internal/platform/httpapi"
)

type Application interface {
	CurrentSnapshot(ctx context.Context, userID int64) (Snapshot, error)
	Redeem(ctx context.Context, input RedeemInput) (RedeemResult, error)
	ListPlans(ctx context.Context) ([]PlanOption, error)
	CurrentUsage(ctx context.Context, userID int64) ([]UsageItem, error)
	FeatureAccess(ctx context.Context, userID int64, keys []string) (FeatureAccessResponse, error)
	ListOrders(ctx context.Context, userID int64, limit int) ([]Order, error)
	CreateCheckout(ctx context.Context, input CheckoutInput) (CheckoutResult, error)
}

type HTTPHandler struct {
	app Application
}

func NewHTTPHandler(app Application) *HTTPHandler {
	return &HTTPHandler{app: app}
}

func (h *HTTPHandler) Register(router *gin.RouterGroup) {
	router.GET("/membership/me", h.me)
	router.GET("/membership/plans", h.listPlans)
	router.GET("/membership/usage", h.usage)
	router.GET("/membership/feature-access", h.featureAccess)
	router.GET("/membership/orders", h.listOrders)
	router.POST("/membership/checkout", h.checkout)
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

func (h *HTTPHandler) listPlans(c *gin.Context) {
	plans, err := h.app.ListPlans(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"plans": httpapi.EnsureSlice(plans)})
}

func (h *HTTPHandler) usage(c *gin.Context) {
	usage, err := h.app.CurrentUsage(c.Request.Context(), c.GetInt64(auth.UserIDContextKey))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"usage": httpapi.EnsureSlice(usage)})
}

func (h *HTTPHandler) featureAccess(c *gin.Context) {
	result, err := h.app.FeatureAccess(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), c.QueryArray("key"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *HTTPHandler) listOrders(c *gin.Context) {
	limit, ok := httpapi.QueryLimit(c, 20, 100)
	if !ok {
		return
	}
	orders, err := h.app.ListOrders(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), limit)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"orders": httpapi.EnsureSlice(orders)})
}

func (h *HTTPHandler) checkout(c *gin.Context) {
	var request CheckoutInput
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}
	request.UserID = c.GetInt64(auth.UserIDContextKey)
	result, err := h.app.CreateCheckout(c.Request.Context(), request)
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
	case errors.Is(err, ErrInvalidCheckout):
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
	case errors.Is(err, ErrCodeNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "code_not_found"})
	case errors.Is(err, ErrPlanNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "plan_not_found"})
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
