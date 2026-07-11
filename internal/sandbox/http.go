package sandbox

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/zzm/opcv2/internal/auth"
	"github.com/zzm/opcv2/internal/membership"
	"github.com/zzm/opcv2/internal/platform/httpapi"
)

type Application interface {
	ListRoles() []Role
	CreateSession(ctx context.Context, input CreateInput) (Session, error)
	UpdateSessionDraft(ctx context.Context, userID, id int64, update DraftUpdate) (Session, error)
	RunSession(ctx context.Context, userID, id int64) (Session, error)
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
	router.GET("/sandbox/roles", h.listRoles)
	router.POST("/sandbox/sessions", h.createSession)
	router.PATCH("/sandbox/sessions/:id/draft", h.updateSessionDraft)
	router.POST("/sandbox/sessions/:id/run", h.runSession)
	router.GET("/sandbox/sessions", h.listSessions)
	router.GET("/sandbox/sessions/:id", h.getSession)
}

func (h *HTTPHandler) listRoles(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"roles": httpapi.EnsureSlice(h.app.ListRoles())})
}

func (h *HTTPHandler) createSession(c *gin.Context) {
	var request CreateInput
	if err := c.ShouldBindJSON(&request); err != nil {
		httpapi.BadRequest(c, "invalid_request")
		return
	}
	if !validCreateInput(request) {
		httpapi.BadRequest(c, "invalid_request")
		return
	}
	request.UserID = c.GetInt64(auth.UserIDContextKey)
	session, err := h.app.CreateSession(c.Request.Context(), request)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, session)
}

func validCreateInput(input CreateInput) bool {
	return strings.TrimSpace(input.Goal) != "" &&
		strings.TrimSpace(input.TargetUsers) != "" &&
		strings.TrimSpace(input.Product) != ""
}

func (h *HTTPHandler) updateSessionDraft(c *gin.Context) {
	id, ok := sessionID(c)
	if !ok {
		return
	}
	var request DraftUpdate
	if err := c.ShouldBindJSON(&request); err != nil || (request.Goal == nil && request.TargetUsers == nil && request.Product == nil && request.Roles == nil) {
		httpapi.BadRequest(c, "invalid_request")
		return
	}
	session, err := h.app.UpdateSessionDraft(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), id, request)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, session)
}

func (h *HTTPHandler) runSession(c *gin.Context) {
	id, ok := sessionID(c)
	if !ok {
		return
	}
	session, err := h.app.RunSession(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, session)
}

func (h *HTTPHandler) listSessions(c *gin.Context) {
	limit, ok := httpapi.QueryLimit(c, 20, 100)
	if !ok {
		return
	}
	sessions, err := h.app.ListSessions(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), limit)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"sessions": httpapi.EnsureSlice(sessions)})
}

func (h *HTTPHandler) getSession(c *gin.Context) {
	id, ok := sessionID(c)
	if !ok {
		return
	}
	session, err := h.app.GetSession(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, session)
}

func sessionID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		httpapi.BadRequest(c, "invalid_session_id")
		return 0, false
	}
	return id, true
}

func writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrSessionNotFound):
		httpapi.Error(c, http.StatusNotFound, "session_not_found")
	case errors.Is(err, ErrServiceNotReady):
		httpapi.Error(c, http.StatusInternalServerError, "service_not_ready")
	case errors.Is(err, ErrInvalidAIResult):
		httpapi.Error(c, http.StatusInternalServerError, "invalid_ai_result")
	case errors.Is(err, ErrInvalidSession):
		httpapi.BadRequest(c, "invalid_session")
	case errors.Is(err, membership.ErrQuotaExceeded):
		httpapi.Error(c, http.StatusPaymentRequired, "quota_exceeded")
	case errors.Is(err, membership.ErrQuotaNotFound):
		httpapi.Error(c, http.StatusInternalServerError, "quota_not_configured")
	default:
		httpapi.Error(c, http.StatusInternalServerError, "internal_error")
	}
}
