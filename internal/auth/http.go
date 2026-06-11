package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	refreshCookieName = "opcv2_refresh"
	UserIDContextKey  = "auth_user_id"
)

type AuthApplication interface {
	SendCode(ctx context.Context, phone string) error
	Login(ctx context.Context, input LoginInput) (LoginResult, error)
	Refresh(ctx context.Context, refreshToken string) (LoginResult, error)
	Logout(ctx context.Context, refreshToken string) error
	CurrentUser(ctx context.Context, userID int64) (User, error)
}

type HTTPHandler struct {
	app          AuthApplication
	tokens       TokenManager
	secureCookie bool
}

func NewHTTPHandler(app AuthApplication, tokens TokenManager, secureCookie bool) *HTTPHandler {
	return &HTTPHandler{app: app, tokens: tokens, secureCookie: secureCookie}
}

func (h *HTTPHandler) Register(router *gin.RouterGroup) {
	authRoutes := router.Group("/auth")
	authRoutes.POST("/sms/send", h.sendCode)
	authRoutes.POST("/login", h.login)
	authRoutes.POST("/refresh", h.refresh)
	authRoutes.POST("/logout", h.logout)
	router.GET("/me", h.RequireAccessToken(), h.me)
}

func (h *HTTPHandler) sendCode(c *gin.Context) {
	var request struct {
		Phone string `json:"phone"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}
	if err := h.app.SendCode(c.Request.Context(), request.Phone); err != nil {
		writeAuthError(c, err)
		return
	}
	c.JSON(http.StatusAccepted, gin.H{"status": "sent", "cooldown_seconds": 60})
}

func (h *HTTPHandler) login(c *gin.Context) {
	var request struct {
		Nickname          string `json:"nickname"`
		Phone             string `json:"phone"`
		Code              string `json:"code"`
		AgreementAccepted bool   `json:"agreement_accepted"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}
	result, err := h.app.Login(c.Request.Context(), LoginInput{
		Nickname:          request.Nickname,
		Phone:             request.Phone,
		Code:              request.Code,
		AgreementAccepted: request.AgreementAccepted,
		IP:                c.ClientIP(),
		UserAgent:         c.Request.UserAgent(),
	})
	if err != nil {
		writeAuthError(c, err)
		return
	}
	h.setRefreshCookie(c, result.RefreshToken)
	writeLoginResult(c, result)
}

func (h *HTTPHandler) refresh(c *gin.Context) {
	refreshToken, err := c.Cookie(refreshCookieName)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_refresh_token"})
		return
	}
	result, err := h.app.Refresh(c.Request.Context(), refreshToken)
	if err != nil {
		h.clearRefreshCookie(c)
		writeAuthError(c, err)
		return
	}
	h.setRefreshCookie(c, result.RefreshToken)
	writeLoginResult(c, result)
}

func (h *HTTPHandler) logout(c *gin.Context) {
	refreshToken, _ := c.Cookie(refreshCookieName)
	if err := h.app.Logout(c.Request.Context(), refreshToken); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "logout_failed"})
		return
	}
	h.clearRefreshCookie(c)
	c.Status(http.StatusNoContent)
}

func (h *HTTPHandler) me(c *gin.Context) {
	userID := c.GetInt64(UserIDContextKey)
	user, err := h.app.CurrentUser(c.Request.Context(), userID)
	if err != nil {
		writeAuthError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": user})
}

func (h *HTTPHandler) RequireAccessToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := strings.TrimSpace(c.GetHeader("Authorization"))
		if !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid_access_token"})
			return
		}
		userID, err := h.tokens.ParseAccess(strings.TrimSpace(strings.TrimPrefix(header, "Bearer ")))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid_access_token"})
			return
		}
		c.Set(UserIDContextKey, userID)
		c.Next()
	}
}

func (h *HTTPHandler) setRefreshCookie(c *gin.Context, token string) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     refreshCookieName,
		Value:    token,
		Path:     "/api/v1/auth",
		MaxAge:   int((7 * 24 * time.Hour).Seconds()),
		HttpOnly: true,
		Secure:   h.secureCookie,
		SameSite: http.SameSiteLaxMode,
	})
}

func (h *HTTPHandler) clearRefreshCookie(c *gin.Context) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     refreshCookieName,
		Value:    "",
		Path:     "/api/v1/auth",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.secureCookie,
		SameSite: http.SameSiteLaxMode,
	})
}

func writeLoginResult(c *gin.Context, result LoginResult) {
	c.JSON(http.StatusOK, gin.H{
		"user":                    result.User,
		"access_token":            result.AccessToken,
		"access_token_expires_at": result.AccessTokenExpires,
		"is_new_user":             result.IsNewUser,
	})
}

func writeAuthError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrInvalidPhone):
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_phone"})
	case errors.Is(err, ErrInvalidCode):
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_code"})
	case errors.Is(err, ErrCodeRateLimited):
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "code_rate_limited"})
	case errors.Is(err, ErrAgreementRequired):
		c.JSON(http.StatusBadRequest, gin.H{"error": "agreement_required"})
	case errors.Is(err, ErrNicknameRequired):
		c.JSON(http.StatusBadRequest, gin.H{"error": "nickname_required"})
	case errors.Is(err, ErrInvalidRefreshToken), errors.Is(err, ErrInvalidAccessToken):
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_token"})
	case errors.Is(err, ErrUserNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "user_not_found"})
	case errors.Is(err, ErrUserDisabled):
		c.JSON(http.StatusForbidden, gin.H{"error": "user_disabled"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
	}
}
