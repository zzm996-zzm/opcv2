package auth

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	refreshCookieName = "opcv2_refresh"
	UserIDContextKey  = "auth_user_id"
	maxAuthBodySize   = 16 << 10
)

type AuthApplication interface {
	SendCode(ctx context.Context, phone string) error
	Login(ctx context.Context, input LoginInput) (LoginResult, error)
	Register(ctx context.Context, input RegisterInput) (LoginResult, error)
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
	authRoutes.POST("/register", h.register)
	authRoutes.POST("/refresh", h.refresh)
	authRoutes.POST("/logout", h.logout)
	router.GET("/me", h.RequireAccessToken(), h.me)
}

func (h *HTTPHandler) sendCode(c *gin.Context) {
	var request struct {
		Phone string `json:"phone"`
	}
	if err := decodeAuthJSON(c, &request); err != nil {
		writeAuthResponse(c, http.StatusBadRequest, "invalid_request", "请求内容有误，请检查后重试")
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
		Account           string `json:"account"`
		Password          string `json:"password"`
		AgreementAccepted bool   `json:"agreement_accepted"`
	}
	if err := decodeAuthJSON(c, &request); err != nil {
		writeAuthResponse(c, http.StatusBadRequest, "invalid_request", "登录信息格式有误，请检查后重试")
		return
	}
	result, err := h.app.Login(c.Request.Context(), LoginInput{
		Nickname:          request.Nickname,
		Phone:             request.Phone,
		Code:              request.Code,
		Account:           request.Account,
		Password:          request.Password,
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

func (h *HTTPHandler) register(c *gin.Context) {
	var request struct {
		Nickname          string `json:"nickname"`
		Account           string `json:"account"`
		Password          string `json:"password"`
		AgreementAccepted bool   `json:"agreement_accepted"`
	}
	if err := decodeAuthJSON(c, &request); err != nil {
		writeAuthResponse(c, http.StatusBadRequest, "invalid_request", "注册信息格式有误，请检查后重试")
		return
	}
	result, err := h.app.Register(c.Request.Context(), RegisterInput{
		Nickname:          request.Nickname,
		Account:           request.Account,
		Password:          request.Password,
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
		writeAuthResponse(c, http.StatusUnauthorized, "invalid_refresh_token", "登录状态已过期，请重新登录")
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
		writeAuthResponse(c, http.StatusInternalServerError, "logout_failed", "暂时无法退出登录，请稍后再试")
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
			c.Abort()
			writeAuthResponse(c, http.StatusUnauthorized, "invalid_access_token", "请先登录后再继续操作")
			return
		}
		userID, err := h.tokens.ParseAccess(strings.TrimSpace(strings.TrimPrefix(header, "Bearer ")))
		if err != nil {
			c.Abort()
			writeAuthResponse(c, http.StatusUnauthorized, "invalid_access_token", "登录状态已过期，请重新登录")
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

func decodeAuthJSON(c *gin.Context, dst any) error {
	if c.Request.Body == nil {
		return errors.New("empty request body")
	}
	decoder := json.NewDecoder(http.MaxBytesReader(c.Writer, c.Request.Body, maxAuthBodySize))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); err != nil {
		if errors.Is(err, io.EOF) {
			return nil
		}
		return errors.New("request must contain a single JSON value")
	}
	return errors.New("request must contain a single JSON value")
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
		writeAuthResponse(c, http.StatusBadRequest, "invalid_phone", "请输入正确的中国大陆手机号")
	case errors.Is(err, ErrInvalidCode):
		writeAuthResponse(c, http.StatusUnauthorized, "invalid_code", "验证码错误或已过期，请重新获取")
	case errors.Is(err, ErrCodeRateLimited):
		writeAuthResponse(c, http.StatusTooManyRequests, "code_rate_limited", "验证码发送太频繁，请稍后再试")
	case errors.Is(err, ErrSMSUnavailable):
		writeAuthResponse(c, http.StatusServiceUnavailable, "sms_unavailable", "短信服务暂时不可用，请稍后再试")
	case errors.Is(err, ErrAgreementRequired):
		writeAuthResponse(c, http.StatusBadRequest, "agreement_required", "请先阅读并同意用户协议和隐私政策")
	case errors.Is(err, ErrNicknameRequired):
		writeAuthResponse(c, http.StatusBadRequest, "nickname_required", "请输入昵称")
	case errors.Is(err, ErrInvalidAccount):
		writeAuthResponse(c, http.StatusBadRequest, "invalid_account", "账号需为4-32位字母、数字或下划线")
	case errors.Is(err, ErrInvalidPassword):
		writeAuthResponse(c, http.StatusBadRequest, "invalid_password", "密码需为6-72位")
	case errors.Is(err, ErrInvalidCredentials):
		writeAuthResponse(c, http.StatusUnauthorized, "invalid_credentials", "账号或密码不正确，请重新输入")
	case errors.Is(err, ErrAccountExists):
		writeAuthResponse(c, http.StatusConflict, "account_exists", "该账号已注册，请直接登录")
	case errors.Is(err, ErrInvalidRefreshToken), errors.Is(err, ErrInvalidAccessToken):
		writeAuthResponse(c, http.StatusUnauthorized, "invalid_token", "登录状态已过期，请重新登录")
	case errors.Is(err, ErrUserNotFound):
		writeAuthResponse(c, http.StatusNotFound, "user_not_found", "账号不存在或已失效")
	case errors.Is(err, ErrUserDisabled):
		writeAuthResponse(c, http.StatusForbidden, "user_disabled", "该账号已被停用，如有疑问请联系客服")
	case errors.Is(err, ErrAuthNotConfigured):
		writeAuthResponse(c, http.StatusInternalServerError, "service_not_ready", "登录服务暂时不可用，请稍后再试")
	default:
		writeAuthResponse(c, http.StatusInternalServerError, "internal_error", "服务开小差了，请稍后再试")
	}
}

func writeAuthResponse(c *gin.Context, status int, code, message string) {
	c.JSON(status, gin.H{"error": code, "message": message})
}
