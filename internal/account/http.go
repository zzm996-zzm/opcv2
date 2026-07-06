package account

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zzm/opcv2/internal/auth"
	"github.com/zzm/opcv2/internal/platform/httpapi"
)

type Application interface {
	GetProfile(ctx context.Context, userID int64) (ProfilePayload, error)
	GetProfileContext(ctx context.Context, userID int64) (ProfileContext, error)
	UpdateProfile(ctx context.Context, userID int64, update ProfileUpdate) (ProfilePayload, error)
	GetOnboarding(ctx context.Context, userID int64) (OnboardingState, error)
	SaveOnboarding(ctx context.Context, userID int64, state OnboardingState) (OnboardingState, error)
	CompleteOnboarding(ctx context.Context, userID int64) (OnboardingState, error)
	GetPreferences(ctx context.Context, userID int64) (Preferences, error)
	UpdatePreferences(ctx context.Context, userID int64, update PreferencesUpdate) (Preferences, error)
	ListQuotas(ctx context.Context, userID int64) ([]Quota, error)
	ListContent(ctx context.Context, userID int64, limit int) ([]ContentItem, error)
	DeleteAccount(ctx context.Context, userID int64) (DeletionStatus, error)
}

type HTTPHandler struct {
	app Application
}

func NewHTTPHandler(app Application) *HTTPHandler {
	return &HTTPHandler{app: app}
}

func (h *HTTPHandler) Register(router *gin.RouterGroup) {
	router.GET("/account/profile", h.getProfile)
	router.GET("/account/profile-context", h.getProfileContext)
	router.PATCH("/account/profile", h.updateProfile)
	router.GET("/account/onboarding", h.getOnboarding)
	router.PUT("/account/onboarding", h.saveOnboarding)
	router.POST("/account/onboarding/complete", h.completeOnboarding)
	router.GET("/account/preferences", h.getPreferences)
	router.PATCH("/account/preferences", h.updatePreferences)
	router.GET("/account/quotas", h.listQuotas)
	router.GET("/account/content", h.listContent)
	router.DELETE("/account", h.deleteAccount)
}

func (h *HTTPHandler) getProfile(c *gin.Context) {
	payload, err := h.app.GetProfile(c.Request.Context(), c.GetInt64(auth.UserIDContextKey))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, payload)
}

func (h *HTTPHandler) getProfileContext(c *gin.Context) {
	context, err := h.app.GetProfileContext(c.Request.Context(), c.GetInt64(auth.UserIDContextKey))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, context)
}

func (h *HTTPHandler) updateProfile(c *gin.Context) {
	var request ProfileUpdate
	if err := c.ShouldBindJSON(&request); err != nil {
		httpapi.BadRequest(c, "invalid_request")
		return
	}
	payload, err := h.app.UpdateProfile(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), request)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, payload)
}

func (h *HTTPHandler) getOnboarding(c *gin.Context) {
	state, err := h.app.GetOnboarding(c.Request.Context(), c.GetInt64(auth.UserIDContextKey))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, state)
}

func (h *HTTPHandler) saveOnboarding(c *gin.Context) {
	var request OnboardingState
	if err := c.ShouldBindJSON(&request); err != nil {
		httpapi.BadRequest(c, "invalid_request")
		return
	}
	state, err := h.app.SaveOnboarding(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), request)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, state)
}

func (h *HTTPHandler) completeOnboarding(c *gin.Context) {
	state, err := h.app.CompleteOnboarding(c.Request.Context(), c.GetInt64(auth.UserIDContextKey))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, state)
}

func (h *HTTPHandler) getPreferences(c *gin.Context) {
	preferences, err := h.app.GetPreferences(c.Request.Context(), c.GetInt64(auth.UserIDContextKey))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, preferences)
}

func (h *HTTPHandler) updatePreferences(c *gin.Context) {
	var request PreferencesUpdate
	if err := c.ShouldBindJSON(&request); err != nil {
		httpapi.BadRequest(c, "invalid_request")
		return
	}
	preferences, err := h.app.UpdatePreferences(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), request)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, preferences)
}

func (h *HTTPHandler) listQuotas(c *gin.Context) {
	quotas, err := h.app.ListQuotas(c.Request.Context(), c.GetInt64(auth.UserIDContextKey))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"quotas": httpapi.EnsureSlice(quotas)})
}

func (h *HTTPHandler) listContent(c *gin.Context) {
	limit, ok := httpapi.QueryLimit(c, 20, 100)
	if !ok {
		return
	}
	items, err := h.app.ListContent(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), limit)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": httpapi.EnsureSlice(items)})
}

func (h *HTTPHandler) deleteAccount(c *gin.Context) {
	status, err := h.app.DeleteAccount(c.Request.Context(), c.GetInt64(auth.UserIDContextKey))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusAccepted, status)
}

func writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrInvalidEmail), errors.Is(err, ErrInvalidOnboarding), errors.Is(err, ErrUserIDRequired):
		httpapi.BadRequest(c, "invalid_request")
	case errors.Is(err, ErrProfileNotFound):
		httpapi.Error(c, http.StatusNotFound, "profile_not_found")
	case errors.Is(err, ErrServiceNotReady):
		httpapi.Error(c, http.StatusInternalServerError, "service_not_ready")
	default:
		httpapi.Error(c, http.StatusInternalServerError, "internal_error")
	}
}
