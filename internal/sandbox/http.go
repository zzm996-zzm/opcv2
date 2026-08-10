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
	AskRole(ctx context.Context, input AskRoleInput) (Message, error)
	ListMessages(ctx context.Context, userID, sessionID int64) ([]Message, error)
	RunSession(ctx context.Context, userID, id int64) (Session, error)
	RetrySession(ctx context.Context, userID, id int64) (Session, error)
	CancelSession(ctx context.Context, userID, id int64) (Session, error)
	ListSessions(ctx context.Context, userID int64, limit int) ([]Session, error)
	GetSession(ctx context.Context, userID, id int64) (Session, error)
}

type FlowApplication interface {
	Options() Options
	ListExamples() []Session
	CreateIntake(ctx context.Context, input IntakeCreateInput) (Session, error)
	AnswerIntake(ctx context.Context, input IntakeAnswerInput) (Session, error)
	CompleteIntake(ctx context.Context, userID, id int64) (Session, error)
	UpdateSessionSettings(ctx context.Context, userID, id int64, settings RunSettings) (Session, error)
}

type V2Application interface {
	ListV2Roles(ctx context.Context) ([]V2RoleConfig, error)
	CreateV2Run(ctx context.Context, input CreateV2RunInput) (V2SandboxRun, error)
	AnswerV2Run(ctx context.Context, input AnswerV2RunInput) (V2SandboxRun, error)
	SetV2Roles(ctx context.Context, input SetV2RolesInput) (V2SandboxRun, error)
	GetV2Run(ctx context.Context, userID, runID int64) (V2SandboxRun, error)
	ListV2Runs(ctx context.Context, userID int64, limit int) ([]V2SandboxRun, error)
	DeleteV2Run(ctx context.Context, userID, runID int64) error
}

type HTTPHandler struct {
	app Application
}

func NewHTTPHandler(app Application) *HTTPHandler {
	return &HTTPHandler{app: app}
}

func (h *HTTPHandler) Register(router *gin.RouterGroup) {
	router.GET("/sandbox/options", h.options)
	router.GET("/sandbox/examples", h.listExamples)
	router.GET("/sandbox/roles", h.listRoles)
	router.POST("/sandbox/sessions/intake", h.createIntake)
	router.POST("/sandbox/intake", h.createIntake)
	router.POST("/sandbox/sessions", h.createSession)
	router.PATCH("/sandbox/sessions/:id/draft", h.updateSessionDraft)
	router.PATCH("/sandbox/sessions/:id/settings", h.updateSessionSettings)
	router.PUT("/sandbox/sessions/:id/intake/questions/:key", h.answerIntake)
	router.PUT("/sandbox/sessions/:id/intake/answers/:key", h.answerIntake)
	router.POST("/sandbox/sessions/:id/intake/complete", h.completeIntake)
	router.POST("/sandbox/sessions/:id/run", h.runSession)
	router.POST("/sandbox/sessions/:id/retry", h.retrySession)
	router.POST("/sandbox/sessions/:id/cancel", h.cancelSession)
	router.GET("/sandbox/sessions/:id/status", h.getSession)
	router.GET("/sandbox/sessions/:id/messages", h.listMessages)
	router.POST("/sandbox/sessions/:id/messages", h.askRole)
	router.GET("/sandbox/sessions", h.listSessions)
	router.GET("/sandbox/sessions/:id", h.getSession)
	router.GET("/sandbox-runs/roles", h.listV2Roles)
	router.POST("/sandbox-runs", h.createV2Run)
	router.POST("/sandbox-runs/:id/answer", h.answerV2Run)
	router.POST("/sandbox-runs/:id/roles", h.setV2Roles)
	router.GET("/sandbox-runs", h.listV2Runs)
	router.GET("/sandbox-runs/:id", h.getV2Run)
	router.DELETE("/sandbox-runs/:id", h.deleteV2Run)
}

func (h *HTTPHandler) listExamples(c *gin.Context) {
	app, ok := h.flowApplication(c)
	if !ok {
		return
	}
	c.JSON(http.StatusOK, gin.H{"sessions": httpapi.EnsureSlice(app.ListExamples())})
}

func (h *HTTPHandler) options(c *gin.Context) {
	app, ok := h.flowApplication(c)
	if !ok {
		return
	}
	c.JSON(http.StatusOK, app.Options())
}

func (h *HTTPHandler) createIntake(c *gin.Context) {
	var request IntakeCreateInput
	if err := c.ShouldBindJSON(&request); err != nil || strings.TrimSpace(request.InitialIdea) == "" || len([]rune(strings.TrimSpace(request.InitialIdea))) > 5000 {
		httpapi.BadRequest(c, "invalid_request")
		return
	}
	request.UserID = c.GetInt64(auth.UserIDContextKey)
	app, ok := h.flowApplication(c)
	if !ok {
		return
	}
	session, err := app.CreateIntake(c.Request.Context(), request)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, session)
}

func (h *HTTPHandler) answerIntake(c *gin.Context) {
	id, ok := sessionID(c)
	if !ok {
		return
	}
	var request IntakeAnswerInput
	if err := c.ShouldBindJSON(&request); err != nil {
		httpapi.BadRequest(c, "invalid_request")
		return
	}
	request.UserID = c.GetInt64(auth.UserIDContextKey)
	request.SessionID = id
	request.QuestionKey = strings.TrimSpace(c.Param("key"))
	if request.QuestionKey == "" || (strings.TrimSpace(request.Answer) == "" && !request.Skipped) {
		httpapi.BadRequest(c, "invalid_request")
		return
	}
	app, ok := h.flowApplication(c)
	if !ok {
		return
	}
	session, err := app.AnswerIntake(c.Request.Context(), request)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, session)
}

func (h *HTTPHandler) completeIntake(c *gin.Context) {
	id, ok := sessionID(c)
	if !ok {
		return
	}
	app, ok := h.flowApplication(c)
	if !ok {
		return
	}
	session, err := app.CompleteIntake(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, session)
}

func (h *HTTPHandler) listMessages(c *gin.Context) {
	id, ok := sessionID(c)
	if !ok {
		return
	}
	messages, err := h.app.ListMessages(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"messages": httpapi.EnsureSlice(messages)})
}

func (h *HTTPHandler) askRole(c *gin.Context) {
	id, ok := sessionID(c)
	if !ok {
		return
	}
	var request AskRoleInput
	if err := c.ShouldBindJSON(&request); err != nil {
		httpapi.BadRequest(c, "invalid_request")
		return
	}
	request.UserID = c.GetInt64(auth.UserIDContextKey)
	request.SessionID = id
	message, err := h.app.AskRole(c.Request.Context(), request)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, message)
}

func (h *HTTPHandler) listRoles(c *gin.Context) {
	if c.Query("version") == "2" {
		h.listV2Roles(c)
		return
	}
	c.JSON(http.StatusOK, gin.H{"roles": httpapi.EnsureSlice(h.app.ListRoles())})
}

func (h *HTTPHandler) listV2Roles(c *gin.Context) {
	app, ok := h.v2Application(c)
	if !ok {
		return
	}
	roles, err := app.ListV2Roles(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"roles": httpapi.EnsureSlice(roles)})
}

func (h *HTTPHandler) createV2Run(c *gin.Context) {
	var request CreateV2RunInput
	if err := c.ShouldBindJSON(&request); err != nil {
		httpapi.BadRequest(c, "invalid_request")
		return
	}
	request.UserID = c.GetInt64(auth.UserIDContextKey)
	app, ok := h.v2Application(c)
	if !ok {
		return
	}
	run, err := app.CreateV2Run(c.Request.Context(), request)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, run)
}

func (h *HTTPHandler) answerV2Run(c *gin.Context) {
	id, ok := sessionID(c)
	if !ok {
		return
	}
	var request AnswerV2RunInput
	if err := c.ShouldBindJSON(&request); err != nil {
		httpapi.BadRequest(c, "invalid_request")
		return
	}
	request.UserID = c.GetInt64(auth.UserIDContextKey)
	request.RunID = id
	app, ok := h.v2Application(c)
	if !ok {
		return
	}
	run, err := app.AnswerV2Run(c.Request.Context(), request)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, run)
}

func (h *HTTPHandler) setV2Roles(c *gin.Context) {
	id, ok := sessionID(c)
	if !ok {
		return
	}
	var request SetV2RolesInput
	if err := c.ShouldBindJSON(&request); err != nil {
		httpapi.BadRequest(c, "invalid_request")
		return
	}
	request.UserID = c.GetInt64(auth.UserIDContextKey)
	request.RunID = id
	app, ok := h.v2Application(c)
	if !ok {
		return
	}
	run, err := app.SetV2Roles(c.Request.Context(), request)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, run)
}

func (h *HTTPHandler) getV2Run(c *gin.Context) {
	id, ok := sessionID(c)
	if !ok {
		return
	}
	app, ok := h.v2Application(c)
	if !ok {
		return
	}
	run, err := app.GetV2Run(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, run)
}

func (h *HTTPHandler) listV2Runs(c *gin.Context) {
	limit, ok := httpapi.QueryLimit(c, 20, 100)
	if !ok {
		return
	}
	app, ok := h.v2Application(c)
	if !ok {
		return
	}
	runs, err := app.ListV2Runs(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), limit)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"runs": httpapi.EnsureSlice(runs)})
}

func (h *HTTPHandler) deleteV2Run(c *gin.Context) {
	id, ok := sessionID(c)
	if !ok {
		return
	}
	app, ok := h.v2Application(c)
	if !ok {
		return
	}
	if err := app.DeleteV2Run(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), id); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
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

func (h *HTTPHandler) retrySession(c *gin.Context) {
	id, ok := sessionID(c)
	if !ok {
		return
	}
	session, err := h.app.RetrySession(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusAccepted, session)
}

func (h *HTTPHandler) cancelSession(c *gin.Context) {
	id, ok := sessionID(c)
	if !ok {
		return
	}
	session, err := h.app.CancelSession(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), id)
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
	if err := c.ShouldBindJSON(&request); err != nil || (request.Goal == nil && request.TargetUsers == nil && request.Product == nil && request.Roles == nil && request.Settings == nil) {
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

func (h *HTTPHandler) updateSessionSettings(c *gin.Context) {
	id, ok := sessionID(c)
	if !ok {
		return
	}
	var request RunSettings
	if err := c.ShouldBindJSON(&request); err != nil {
		httpapi.BadRequest(c, "invalid_request")
		return
	}
	app, ok := h.flowApplication(c)
	if !ok {
		return
	}
	session, err := app.UpdateSessionSettings(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), id, request)
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
	c.JSON(http.StatusAccepted, session)
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

func (h *HTTPHandler) flowApplication(c *gin.Context) (FlowApplication, bool) {
	app, ok := h.app.(FlowApplication)
	if !ok {
		httpapi.Error(c, http.StatusInternalServerError, "service_not_ready")
		return nil, false
	}
	return app, true
}

func (h *HTTPHandler) v2Application(c *gin.Context) (V2Application, bool) {
	app, ok := h.app.(V2Application)
	if !ok {
		httpapi.Error(c, http.StatusInternalServerError, "service_not_ready")
		return nil, false
	}
	return app, true
}

func writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrV2RunNotFound):
		httpapi.Error(c, http.StatusNotFound, "sandbox_run_not_found")
	case errors.Is(err, ErrV2InvalidRequest):
		httpapi.BadRequest(c, "invalid_request")
	case errors.Is(err, ErrV2InvalidRoles):
		httpapi.BadRequest(c, "invalid_roles")
	case errors.Is(err, ErrV2ActiveRun):
		httpapi.Error(c, http.StatusConflict, "active_run_exists")
	case errors.Is(err, ErrV2StartConflict):
		httpapi.Error(c, http.StatusConflict, "run_start_conflict")
	case errors.Is(err, ErrV2Revision):
		httpapi.Error(c, http.StatusConflict, "revision_conflict")
	case errors.Is(err, ErrV2ExportExpired):
		httpapi.Error(c, http.StatusGone, "export_expired")
	case errors.Is(err, ErrSessionNotFound):
		httpapi.Error(c, http.StatusNotFound, "session_not_found")
	case errors.Is(err, ErrServiceNotReady):
		httpapi.Error(c, http.StatusInternalServerError, "service_not_ready")
	case errors.Is(err, ErrInvalidAIResult):
		httpapi.Error(c, http.StatusInternalServerError, "invalid_ai_result")
	case errors.Is(err, ErrInvalidSession):
		httpapi.BadRequest(c, "invalid_session")
	case errors.Is(err, ErrInvalidIntake):
		httpapi.BadRequest(c, "invalid_intake")
	case errors.Is(err, ErrIntakeIncomplete):
		httpapi.BadRequest(c, "intake_incomplete")
	case errors.Is(err, ErrStaleRun):
		httpapi.Error(c, http.StatusConflict, "stale_run")
	case errors.Is(err, membership.ErrQuotaExceeded):
		httpapi.Error(c, http.StatusPaymentRequired, "quota_exceeded")
	case errors.Is(err, membership.ErrQuotaNotFound):
		httpapi.Error(c, http.StatusInternalServerError, "quota_not_configured")
	default:
		httpapi.Error(c, http.StatusInternalServerError, "internal_error")
	}
}
