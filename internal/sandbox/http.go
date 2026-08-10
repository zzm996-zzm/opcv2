package sandbox

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

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
	ListV2RunHistory(ctx context.Context, input V2RunListInput) (V2RunListResult, error)
	DeleteV2Run(ctx context.Context, userID, runID int64) error
	RenameV2Run(ctx context.Context, input RenameV2RunInput) (V2SandboxRun, error)
}

type V2ExecutionApplication interface {
	StartV2Run(ctx context.Context, userID, runID int64) (V2SandboxRun, error)
	StopV2Run(ctx context.Context, userID, runID int64) (V2SandboxRun, error)
	ListV2Events(ctx context.Context, userID, runID, afterID int64, limit int) ([]V2ProgressEvent, error)
	GetV2Report(ctx context.Context, userID, runID int64) (V2SandboxReport, error)
	GenerateV2Report(ctx context.Context, userID, runID int64) (V2SandboxReport, error)
}

type V2ExportApplication interface {
	CreateV2Export(ctx context.Context, input CreateV2ExportInput) (V2Export, error)
	GetV2Export(ctx context.Context, userID, exportID int64) (V2Export, []byte, error)
}

type V2HomeApplication interface {
	GetV2Home(ctx context.Context, userID int64) (V2SandboxHome, error)
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
	router.GET("/sandbox/home", h.v2Home)
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
	router.PATCH("/sandbox-runs/:id", h.renameV2Run)
	router.POST("/sandbox-runs/:id/start", h.startV2Run)
	router.POST("/sandbox-runs/:id/stop", h.stopV2Run)
	router.GET("/sandbox-runs/:id/events", h.streamV2Events)
	router.GET("/sandbox-runs/:id/stream", h.streamV2Events)
	router.GET("/sandbox-runs/:id/report", h.getV2Report)
	router.POST("/sandbox-runs/:id/report", h.generateV2Report)
	router.POST("/sandbox-runs/:id/exports", h.createV2Export)
	router.POST("/sandbox-runs/:id/report/export", h.createV2Export)
	router.GET("/sandbox-runs/:id/exports/:export_id/download", h.downloadV2Export)
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
	if c.Query("version") != "1" {
		if _, ok := h.app.(V2Application); ok {
			h.listV2Roles(c)
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"roles": httpapi.EnsureSlice(h.app.ListRoles())})
}

func (h *HTTPHandler) v2Home(c *gin.Context) {
	app, ok := h.app.(V2HomeApplication)
	if !ok {
		httpapi.Error(c, http.StatusInternalServerError, "service_not_ready")
		return
	}
	home, err := app.GetV2Home(c.Request.Context(), c.GetInt64(auth.UserIDContextKey))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, home)
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
	page := 1
	if value := strings.TrimSpace(c.Query("page")); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed <= 0 {
			httpapi.BadRequest(c, "invalid_page")
			return
		}
		page = parsed
	}
	result, err := app.ListV2RunHistory(c.Request.Context(), V2RunListInput{
		UserID: c.GetInt64(auth.UserIDContextKey), Page: page, Limit: limit,
		Status: c.Query("status"), Product: c.Query("product"),
	})
	if err != nil {
		writeError(c, err)
		return
	}
	result.Runs = httpapi.EnsureSlice(result.Runs)
	c.JSON(http.StatusOK, result)
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

func (h *HTTPHandler) renameV2Run(c *gin.Context) {
	id, ok := sessionID(c)
	if !ok {
		return
	}
	var request RenameV2RunInput
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
	run, err := app.RenameV2Run(c.Request.Context(), request)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, run)
}

func (h *HTTPHandler) startV2Run(c *gin.Context) {
	id, ok := sessionID(c)
	if !ok {
		return
	}
	app, ok := h.v2ExecutionApplication(c)
	if !ok {
		return
	}
	run, err := app.StartV2Run(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusAccepted, run)
}

func (h *HTTPHandler) stopV2Run(c *gin.Context) {
	id, ok := sessionID(c)
	if !ok {
		return
	}
	app, ok := h.v2ExecutionApplication(c)
	if !ok {
		return
	}
	run, err := app.StopV2Run(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, run)
}

func (h *HTTPHandler) getV2Report(c *gin.Context) {
	id, ok := sessionID(c)
	if !ok {
		return
	}
	app, ok := h.v2ExecutionApplication(c)
	if !ok {
		return
	}
	report, err := app.GetV2Report(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, report)
}

func (h *HTTPHandler) generateV2Report(c *gin.Context) {
	id, ok := sessionID(c)
	if !ok {
		return
	}
	app, ok := h.v2ExecutionApplication(c)
	if !ok {
		return
	}
	report, err := app.GenerateV2Report(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, report)
}

func (h *HTTPHandler) createV2Export(c *gin.Context) {
	id, ok := sessionID(c)
	if !ok {
		return
	}
	var request CreateV2ExportInput
	if err := c.ShouldBindJSON(&request); err != nil {
		httpapi.BadRequest(c, "invalid_request")
		return
	}
	request.UserID = c.GetInt64(auth.UserIDContextKey)
	request.RunID = id
	app, ok := h.v2ExportApplication(c)
	if !ok {
		return
	}
	export, err := app.CreateV2Export(c.Request.Context(), request)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, export)
}

func (h *HTTPHandler) downloadV2Export(c *gin.Context) {
	runID, ok := sessionID(c)
	if !ok {
		return
	}
	exportID, err := strconv.ParseInt(c.Param("export_id"), 10, 64)
	if err != nil || exportID <= 0 {
		httpapi.BadRequest(c, "invalid_export_id")
		return
	}
	app, ok := h.v2ExportApplication(c)
	if !ok {
		return
	}
	export, payload, err := app.GetV2Export(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), exportID)
	if err != nil {
		writeError(c, err)
		return
	}
	if export.RunID != runID {
		httpapi.Error(c, http.StatusNotFound, "sandbox_run_not_found")
		return
	}
	filename := fmt.Sprintf("sandbox-export-%d.%s", export.ID, export.Format)
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Data(http.StatusOK, "application/json; charset=utf-8", payload)
}

func (h *HTTPHandler) streamV2Events(c *gin.Context) {
	id, ok := sessionID(c)
	if !ok {
		return
	}
	workflow, ok := h.v2Application(c)
	if !ok {
		return
	}
	execution, ok := h.v2ExecutionApplication(c)
	if !ok {
		return
	}
	userID := c.GetInt64(auth.UserIDContextKey)
	run, err := workflow.GetV2Run(c.Request.Context(), userID, id)
	if err != nil {
		writeError(c, err)
		return
	}
	afterID := int64(0)
	if value := strings.TrimSpace(c.GetHeader("Last-Event-ID")); value != "" {
		parsed, parseErr := strconv.ParseInt(value, 10, 64)
		if parseErr != nil || parsed < 0 {
			httpapi.BadRequest(c, "invalid_last_event_id")
			return
		}
		afterID = parsed
	}
	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		httpapi.Error(c, http.StatusInternalServerError, "stream_not_supported")
		return
	}
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Status(http.StatusOK)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	keepAlive := time.NewTicker(15 * time.Second)
	defer keepAlive.Stop()
	for {
		events, err := execution.ListV2Events(c.Request.Context(), userID, id, afterID, 200)
		if err != nil {
			return
		}
		for _, event := range events {
			data, _ := json.Marshal(event)
			_, _ = fmt.Fprintf(c.Writer, "id: %d\nevent: %s\ndata: %s\n\n", event.ID, event.Event, data)
			afterID = event.ID
		}
		if len(events) > 0 {
			flusher.Flush()
		}
		if isV2TerminalStatus(run.Status) {
			return
		}
		select {
		case <-c.Request.Context().Done():
			return
		case <-ticker.C:
			run, err = workflow.GetV2Run(c.Request.Context(), userID, id)
			if err != nil {
				return
			}
		case <-keepAlive.C:
			_, _ = c.Writer.Write([]byte(": keep-alive\n\n"))
			flusher.Flush()
		}
	}
}

func isV2TerminalStatus(status string) bool {
	return status == V2StatusDone || status == V2StatusPartial || status == V2StatusFailed || status == V2StatusNoResult
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

func (h *HTTPHandler) v2ExecutionApplication(c *gin.Context) (V2ExecutionApplication, bool) {
	app, ok := h.app.(V2ExecutionApplication)
	if !ok {
		httpapi.Error(c, http.StatusInternalServerError, "service_not_ready")
		return nil, false
	}
	return app, true
}

func (h *HTTPHandler) v2ExportApplication(c *gin.Context) (V2ExportApplication, bool) {
	app, ok := h.app.(V2ExportApplication)
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
