package projects

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zzm/opcv2/internal/auth"
	projectfiles "github.com/zzm/opcv2/internal/projects/files"
)

type Application interface {
	ListOpportunities(ctx context.Context, filters OpportunityFilters) ([]Opportunity, error)
	GetOpportunity(ctx context.Context, slug string) (Opportunity, error)
	ListCases(ctx context.Context, filters CaseFilters) ([]CaseStudy, error)
	GetCase(ctx context.Context, slug string) (CaseStudy, error)
	CreateMatch(ctx context.Context, input MatchInput) (MatchResult, error)
	ListMatches(ctx context.Context, userID int64, limit int) ([]MatchSession, error)
	GetMatch(ctx context.Context, userID, id int64) (MatchSession, error)
	AnswerMatch(ctx context.Context, input AnswerMatchInput) (MatchResult, error)
	CreateComparison(ctx context.Context, input CreateComparisonInput) (Comparison, error)
	GetComparison(ctx context.Context, userID, id int64) (Comparison, error)
	CreateExport(ctx context.Context, input CreateExportInput) (Export, error)
	GetExport(ctx context.Context, userID, id int64) (Export, error)
	FavoriteMatch(ctx context.Context, userID, id int64) (Favorite, error)
	ListFavoriteMatches(ctx context.Context, userID int64, limit int) ([]Favorite, error)
	UnfavoriteMatch(ctx context.Context, userID, id int64) error
}

type MatchWorkflowApplication interface {
	CreateProjectMatch(context.Context, CreateProjectMatchInput) (MatchWorkflowResponse, error)
	AnswerProjectMatch(context.Context, AnswerProjectMatchInput) (MatchWorkflowResponse, error)
	GetProjectMatch(context.Context, int64, int64) (MatchWorkflowResponse, error)
}

type MatchGenerationApplication interface {
	GenerateProjectMatch(context.Context, int64, int64) (MatchGenerationResponse, error)
	CancelProjectMatch(context.Context, int64, int64) (MatchGenerationResponse, error)
	ListProjectMatchProgress(context.Context, int64, int64, int64) ([]MatchProgressEvent, error)
}

type HTTPHandler struct {
	app Application
}

func NewHTTPHandler(app Application) *HTTPHandler {
	return &HTTPHandler{app: app}
}

func (h *HTTPHandler) Register(router *gin.RouterGroup) {
	h.RegisterPublic(router)
	h.RegisterProtected(router)
}

// RegisterPublic mounts catalog and published case reads.
func (h *HTTPHandler) RegisterPublic(router *gin.RouterGroup) {
	router.GET("/config", h.getPublicConfig)
	router.GET("/dicts", h.listDictionaryItems)
	router.GET("/project-categories", h.listProjectCategories)
	router.GET("/projects/home", h.getProjectHome)
	router.GET("/projects/opportunities", h.listOpportunities)
	router.GET("/projects/opportunities/:slug", h.getOpportunity)
	router.GET("/projects/cases", h.listCases)
	router.GET("/projects/cases/:slug", h.getCase)
	router.GET("/project-cases", h.listEvidenceCases)
	router.GET("/project-cases/:slug", h.getEvidenceCase)
	router.POST("/content-corrections", h.submitContentCorrection)
	router.GET("/projects", h.listProjects)
	router.GET("/projects/:id", h.getProject)
}

// RegisterProtected mounts operations that create or read user-owned data.
func (h *HTTPHandler) RegisterProtected(router *gin.RouterGroup) {
	router.POST("/project-match-files", h.uploadProjectMatchFile)
	router.GET("/project-match-files", h.listProjectMatchFiles)
	router.GET("/project-match-files/:id", h.getProjectMatchFile)
	router.POST("/project-match-files/:id/retry", h.retryProjectMatchFile)
	router.DELETE("/project-match-files/:id", h.deleteProjectMatchFile)
	router.POST("/project-matches", h.createProjectMatch)
	router.POST("/project-matches/:id/answer", h.answerProjectMatch)
	router.GET("/project-matches/:id", h.getProjectMatch)
	router.POST("/project-matches/:id/generate", h.generateProjectMatch)
	router.GET("/project-matches/:id/stream", h.streamProjectMatch)
	router.POST("/project-matches/:id/cancel", h.cancelProjectMatch)
	router.POST("/project-matches/:id/export", h.exportProjectMatch)
	router.POST("/projects/matches", h.createMatch)
	router.GET("/projects/matches", h.listMatches)
	router.GET("/projects/matches/:id", h.getMatch)
	router.POST("/projects/matches/:id/answers", h.answerMatch)
	router.POST("/projects/matches/:id/favorite", h.favoriteMatch)
	router.DELETE("/projects/matches/:id/favorite", h.unfavoriteMatch)
	router.GET("/projects/favorites", h.listFavoriteMatches)
	router.POST("/projects/:id/favorite", h.favoriteProject)
	router.DELETE("/projects/:id/favorite", h.unfavoriteProject)
	router.GET("/projects/project-favorites", h.listFavoriteProjects)
	router.POST("/projects/:id/diagnose", h.diagnoseProject)
	router.POST("/projects/comparisons", h.createComparison)
	router.GET("/projects/comparisons/:id", h.getComparison)
	router.GET("/projects/compare", h.listProjectCompare)
	router.POST("/projects/compare-items", h.addProjectCompareItem)
	router.DELETE("/projects/compare-items/:id", h.removeProjectCompareItem)
	router.POST("/projects/exports", h.createExport)
	router.GET("/projects/exports/:id/download", h.downloadExport)
}

func (h *HTTPHandler) projectFileApplication(c *gin.Context) (ProjectFileApplication, bool) {
	app, ok := h.app.(ProjectFileApplication)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "service_not_ready"})
		return nil, false
	}
	return app, true
}

func (h *HTTPHandler) uploadProjectMatchFile(c *gin.Context) {
	app, ok := h.projectFileApplication(c)
	if !ok {
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, projectfiles.DefaultMaxFileSize+(1<<20))
	header, err := c.FormFile("file")
	if err != nil || header.Size <= 0 || header.Size > projectfiles.DefaultMaxFileSize {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "project_file_too_large", "max_bytes": projectfiles.DefaultMaxFileSize})
		return
	}
	reader, err := header.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_project_file"})
		return
	}
	defer reader.Close()
	content, err := io.ReadAll(io.LimitReader(reader, projectfiles.DefaultMaxFileSize+1))
	if err != nil || int64(len(content)) > projectfiles.DefaultMaxFileSize {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "project_file_too_large", "max_bytes": projectfiles.DefaultMaxFileSize})
		return
	}
	file, err := app.UploadProjectMatchFile(c.Request.Context(), UploadProjectMatchFileInput{
		UserID: c.GetInt64(auth.UserIDContextKey), Name: header.Filename,
		MIME: header.Header.Get("Content-Type"), Data: content,
	})
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, file)
}

func (h *HTTPHandler) getProjectMatchFile(c *gin.Context) {
	app, ok := h.projectFileApplication(c)
	if !ok {
		return
	}
	fileID, valid := projectFileID(c)
	if !valid {
		return
	}
	file, err := app.GetProjectMatchFile(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), fileID)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, file)
}

func (h *HTTPHandler) listProjectMatchFiles(c *gin.Context) {
	app, ok := h.projectFileApplication(c)
	if !ok {
		return
	}
	var matchID *int64
	if raw := strings.TrimSpace(c.Query("match_id")); raw != "" {
		value, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || value <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_match_id"})
			return
		}
		matchID = &value
	}
	files, err := app.ListProjectMatchFiles(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), matchID)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"files": files})
}

func (h *HTTPHandler) retryProjectMatchFile(c *gin.Context) {
	app, ok := h.projectFileApplication(c)
	if !ok {
		return
	}
	fileID, valid := projectFileID(c)
	if !valid {
		return
	}
	file, err := app.RetryProjectMatchFile(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), fileID)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, file)
}

func (h *HTTPHandler) deleteProjectMatchFile(c *gin.Context) {
	app, ok := h.projectFileApplication(c)
	if !ok {
		return
	}
	fileID, valid := projectFileID(c)
	if !valid {
		return
	}
	if err := app.DeleteProjectMatchFile(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), fileID); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func projectFileID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_file_id"})
		return 0, false
	}
	return id, true
}

func (h *HTTPHandler) matchGenerationApplication(c *gin.Context) (MatchGenerationApplication, bool) {
	app, ok := h.app.(MatchGenerationApplication)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "service_not_ready"})
		return nil, false
	}
	return app, true
}

func (h *HTTPHandler) generateProjectMatch(c *gin.Context) {
	app, ok := h.matchGenerationApplication(c)
	if !ok {
		return
	}
	id, valid := matchID(c)
	if !valid {
		return
	}
	result, err := app.GenerateProjectMatch(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusAccepted, result)
}

func (h *HTTPHandler) cancelProjectMatch(c *gin.Context) {
	app, ok := h.matchGenerationApplication(c)
	if !ok {
		return
	}
	id, valid := matchID(c)
	if !valid {
		return
	}
	result, err := app.CancelProjectMatch(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *HTTPHandler) streamProjectMatch(c *gin.Context) {
	app, ok := h.matchGenerationApplication(c)
	if !ok {
		return
	}
	id, valid := matchID(c)
	if !valid {
		return
	}
	afterID := int64(0)
	if raw := strings.TrimSpace(c.GetHeader("Last-Event-ID")); raw != "" {
		parsed, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || parsed < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_last_event_id"})
			return
		}
		afterID = parsed
	}
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Status(http.StatusOK)
	flusher, canFlush := c.Writer.(http.Flusher)
	for cycle := 0; cycle < 120; cycle++ {
		events, err := app.ListProjectMatchProgress(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), id, afterID)
		if err != nil {
			return
		}
		for _, event := range events {
			writeSSEEvent(c, event)
			afterID = event.ID
			if canFlush {
				flusher.Flush()
			}
			if event.Event == MatchStepDone || event.Event == MatchStepPartial || event.Event == MatchStepError || event.Event == MatchStepCanceled {
				return
			}
		}
		if cycle == 0 && len(events) == 0 && canFlush {
			flusher.Flush()
		}
		select {
		case <-c.Request.Context().Done():
			return
		case <-time.After(500 * time.Millisecond):
		}
	}
}

func writeSSEEvent(c *gin.Context, event MatchProgressEvent) {
	payload, _ := json.Marshal(event.Payload)
	_, _ = c.Writer.WriteString("id: " + strconv.FormatInt(event.ID, 10) + "\n")
	_, _ = c.Writer.WriteString("event: " + event.Event + "\n")
	_, _ = c.Writer.WriteString("data: " + string(payload) + "\n\n")
}

func (h *HTTPHandler) matchWorkflowApplication(c *gin.Context) (MatchWorkflowApplication, bool) {
	app, ok := h.app.(MatchWorkflowApplication)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "service_not_ready"})
		return nil, false
	}
	return app, true
}

func (h *HTTPHandler) createProjectMatch(c *gin.Context) {
	app, ok := h.matchWorkflowApplication(c)
	if !ok {
		return
	}
	var request CreateProjectMatchInput
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}
	request.UserID = c.GetInt64(auth.UserIDContextKey)
	request.IdempotencyKey = strings.TrimSpace(c.GetHeader("Idempotency-Key"))
	result, err := app.CreateProjectMatch(c.Request.Context(), request)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *HTTPHandler) answerProjectMatch(c *gin.Context) {
	app, ok := h.matchWorkflowApplication(c)
	if !ok {
		return
	}
	id, valid := matchID(c)
	if !valid {
		return
	}
	var request AnswerProjectMatchInput
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}
	request.UserID, request.MatchID = c.GetInt64(auth.UserIDContextKey), id
	request.IdempotencyKey = strings.TrimSpace(c.GetHeader("Idempotency-Key"))
	result, err := app.AnswerProjectMatch(c.Request.Context(), request)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *HTTPHandler) getProjectMatch(c *gin.Context) {
	app, ok := h.matchWorkflowApplication(c)
	if !ok {
		return
	}
	id, valid := matchID(c)
	if !valid {
		return
	}
	result, err := app.GetProjectMatch(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *HTTPHandler) catalogApplication(c *gin.Context) (CatalogApplication, bool) {
	app, ok := h.app.(CatalogApplication)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "service_not_ready"})
		return nil, false
	}
	return app, true
}

func (h *HTTPHandler) evidenceCaseApplication(c *gin.Context) (EvidenceCaseApplication, bool) {
	app, ok := h.app.(EvidenceCaseApplication)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "service_not_ready"})
		return nil, false
	}
	return app, true
}

func (h *HTTPHandler) listEvidenceCases(c *gin.Context) {
	app, ok := h.evidenceCaseApplication(c)
	if !ok {
		return
	}
	page, valid := positiveQueryInt(c, "page", 1)
	if !valid {
		return
	}
	pageSize, valid := positiveQueryInt(c, "page_size", 20)
	if !valid {
		return
	}
	items, err := app.ListEvidenceCases(c.Request.Context(), EvidenceCaseFilters{
		CaseType: c.Query("type"), Industry: c.Query("industry"), Scale: c.Query("scale"), Page: page, PageSize: pageSize,
	})
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, items)
}

func (h *HTTPHandler) submitContentCorrection(c *gin.Context) {
	app, ok := h.app.(ContentCorrectionApplication)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "service_not_ready"})
		return
	}
	var request ContentCorrectionInput
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}
	item, err := app.SubmitContentCorrection(c.Request.Context(), request)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusAccepted, item)
}

func (h *HTTPHandler) getEvidenceCase(c *gin.Context) {
	app, ok := h.evidenceCaseApplication(c)
	if !ok {
		return
	}
	item, err := app.GetEvidenceCase(c.Request.Context(), c.Param("slug"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *HTTPHandler) getPublicConfig(c *gin.Context) {
	app, ok := h.catalogApplication(c)
	if !ok {
		return
	}
	c.JSON(http.StatusOK, app.GetPublicConfig(c.Request.Context()))
}

func (h *HTTPHandler) listDictionaryItems(c *gin.Context) {
	app, ok := h.catalogApplication(c)
	if !ok {
		return
	}
	items, err := app.ListDictionaryItems(c.Request.Context(), c.Query("kind"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *HTTPHandler) listProjectCategories(c *gin.Context) {
	app, ok := h.catalogApplication(c)
	if !ok {
		return
	}
	categories, err := app.ListDictionaryItems(c.Request.Context(), "category")
	if err != nil {
		writeError(c, err)
		return
	}
	tracks, err := app.ListDictionaryItems(c.Request.Context(), "sector")
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"categories": categories, "tracks": tracks})
}

func (h *HTTPHandler) getProjectHome(c *gin.Context) {
	app, ok := h.catalogApplication(c)
	if !ok {
		return
	}
	home, err := app.GetProjectHome(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, home)
}

func (h *HTTPHandler) listProjects(c *gin.Context) {
	app, ok := h.catalogApplication(c)
	if !ok {
		return
	}
	page, valid := positiveQueryInt(c, "page", 1)
	if !valid {
		return
	}
	pageSize, valid := positiveQueryInt(c, "page_size", 12)
	if !valid {
		return
	}
	var featured *bool
	if raw := strings.TrimSpace(c.Query("is_featured")); raw != "" {
		value, err := strconv.ParseBool(raw)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_is_featured"})
			return
		}
		featured = &value
	}
	result, err := app.ListProjects(c.Request.Context(), ProjectFilters{
		Keyword:    c.Query("keyword"),
		Category:   c.Query("category"),
		Track:      c.Query("track"),
		Budget:     c.Query("budget"),
		Difficulty: c.Query("difficulty"),
		Resource:   c.Query("resource"),
		Sort:       c.Query("sort"),
		Featured:   featured,
		Page:       page,
		PageSize:   pageSize,
	})
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *HTTPHandler) getProject(c *gin.Context) {
	app, ok := h.catalogApplication(c)
	if !ok {
		return
	}
	project, err := app.GetProject(c.Request.Context(), c.Param("id"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, project)
}

func positiveQueryInt(c *gin.Context, name string, defaultValue int) (int, bool) {
	raw := strings.TrimSpace(c.Query(name))
	if raw == "" {
		return defaultValue, true
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_" + name})
		return 0, false
	}
	return value, true
}

func (h *HTTPHandler) createExport(c *gin.Context) {
	var request CreateExportInput
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}
	request.UserID = c.GetInt64(auth.UserIDContextKey)
	item, err := h.app.CreateExport(c.Request.Context(), request)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *HTTPHandler) exportProjectMatch(c *gin.Context) {
	id, ok := matchID(c)
	if !ok {
		return
	}
	var request struct {
		Format   string   `json:"format,omitempty"`
		Includes []string `json:"includes,omitempty"`
	}
	if c.Request.Body != nil && c.Request.ContentLength != 0 {
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
			return
		}
	}
	item, err := h.app.CreateExport(c.Request.Context(), CreateExportInput{UserID: c.GetInt64(auth.UserIDContextKey), SourceType: ExportSourceMatch, SourceID: id, Format: request.Format, Includes: request.Includes})
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusAccepted, item)
}

func (h *HTTPHandler) diagnoseProject(c *gin.Context) {
	app, ok := h.app.(ProjectDiagnosisApplication)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "service_not_ready"})
		return
	}
	var request ProjectDiagnosisInput
	if c.Request.Body != nil && c.Request.ContentLength != 0 {
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
			return
		}
	}
	request.UserID = c.GetInt64(auth.UserIDContextKey)
	result, err := app.DiagnoseProject(c.Request.Context(), request, c.Param("id"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *HTTPHandler) favoriteProject(c *gin.Context) {
	app, ok := h.app.(ProjectFavoriteApplication)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "service_not_ready"})
		return
	}
	item, err := app.FavoriteProject(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), c.Param("id"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *HTTPHandler) unfavoriteProject(c *gin.Context) {
	app, ok := h.app.(ProjectFavoriteApplication)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "service_not_ready"})
		return
	}
	if err := app.UnfavoriteProject(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), c.Param("id")); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *HTTPHandler) listFavoriteProjects(c *gin.Context) {
	app, ok := h.app.(ProjectFavoriteApplication)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "service_not_ready"})
		return
	}
	limit := 20
	if raw := strings.TrimSpace(c.Query("limit")); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_limit"})
			return
		}
		limit = parsed
	}
	items, err := app.ListFavoriteProjects(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), limit)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"favorites": items})
}

func (h *HTTPHandler) addProjectCompareItem(c *gin.Context) {
	app, ok := h.app.(ProjectCompareApplication)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "service_not_ready"})
		return
	}
	var request struct {
		ProjectID string `json:"project_id"`
	}
	if err := c.ShouldBindJSON(&request); err != nil || strings.TrimSpace(request.ProjectID) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}
	item, err := app.AddProjectCompareItem(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), request.ProjectID)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *HTTPHandler) removeProjectCompareItem(c *gin.Context) {
	app, ok := h.app.(ProjectCompareApplication)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "service_not_ready"})
		return
	}
	if err := app.RemoveProjectCompareItem(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), c.Param("id")); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *HTTPHandler) listProjectCompare(c *gin.Context) {
	app, ok := h.app.(ProjectCompareApplication)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "service_not_ready"})
		return
	}
	items, err := app.ListProjectCompareItems(c.Request.Context(), c.GetInt64(auth.UserIDContextKey))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}
func (h *HTTPHandler) downloadExport(c *gin.Context) {
	id, ok := matchID(c)
	if !ok {
		return
	}
	item, err := h.app.GetExport(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), id)
	if err != nil {
		writeError(c, err)
		return
	}
	ext := "json"
	if item.Format != "" {
		ext = item.Format
	}
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="project-export-%d.%s"`, item.ID, ext))
	c.Data(http.StatusOK, "application/json; charset=utf-8", item.Payload)
}

func (h *HTTPHandler) createComparison(c *gin.Context) {
	var request CreateComparisonInput
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}
	request.UserID = c.GetInt64(auth.UserIDContextKey)
	result, err := h.app.CreateComparison(c.Request.Context(), request)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}
func (h *HTTPHandler) getComparison(c *gin.Context) {
	id, ok := matchID(c)
	if !ok {
		return
	}
	result, err := h.app.GetComparison(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *HTTPHandler) answerMatch(c *gin.Context) {
	id, ok := matchID(c)
	if !ok {
		return
	}
	var request AnswerMatchInput
	if err := c.ShouldBindJSON(&request); err != nil || len(request.Answers) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}
	request.UserID = c.GetInt64(auth.UserIDContextKey)
	request.SessionID = id
	result, err := h.app.AnswerMatch(c.Request.Context(), request)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *HTTPHandler) listCases(c *gin.Context) {
	items, err := h.app.ListCases(c.Request.Context(), CaseFilters{
		CaseType:        c.Query("type"),
		OpportunitySlug: strings.TrimSpace(c.Query("opportunity_slug")),
		Industry:        strings.TrimSpace(c.Query("industry")),
		Limit:           20,
	})
	if err != nil {
		writeError(c, err)
		return
	}
	if items == nil {
		items = []CaseStudy{}
	}
	c.JSON(http.StatusOK, gin.H{"cases": items})
}
func (h *HTTPHandler) getCase(c *gin.Context) {
	item, err := h.app.GetCase(c.Request.Context(), c.Param("slug"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *HTTPHandler) listOpportunities(c *gin.Context) {
	limit := 20
	if value := c.Query("limit"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_limit"})
			return
		}
		limit = parsed
	}
	items, err := h.app.ListOpportunities(c.Request.Context(), OpportunityFilters{Query: c.Query("q"), Industry: c.Query("industry"), Limit: limit})
	if err != nil {
		writeError(c, err)
		return
	}
	if items == nil {
		items = []Opportunity{}
	}
	c.JSON(http.StatusOK, gin.H{"opportunities": items})
}

func (h *HTTPHandler) getOpportunity(c *gin.Context) {
	item, err := h.app.GetOpportunity(c.Request.Context(), c.Param("slug"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *HTTPHandler) createMatch(c *gin.Context) {
	var request MatchInput
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}
	request.UserID = c.GetInt64(auth.UserIDContextKey)
	result, err := h.app.CreateMatch(c.Request.Context(), request)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *HTTPHandler) listMatches(c *gin.Context) {
	limit := 20
	if value := c.Query("limit"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_limit"})
			return
		}
		limit = parsed
	}
	sessions, err := h.app.ListMatches(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), limit)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"matches": sessions})
}

func (h *HTTPHandler) getMatch(c *gin.Context) {
	id, ok := matchID(c)
	if !ok {
		return
	}
	session, err := h.app.GetMatch(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, session)
}

func (h *HTTPHandler) favoriteMatch(c *gin.Context) {
	id, ok := matchID(c)
	if !ok {
		return
	}
	favorite, err := h.app.FavoriteMatch(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, favorite)
}

func (h *HTTPHandler) listFavoriteMatches(c *gin.Context) {
	limit := 20
	if value := c.Query("limit"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_limit"})
			return
		}
		limit = parsed
	}
	items, err := h.app.ListFavoriteMatches(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), limit)
	if err != nil {
		writeError(c, err)
		return
	}
	if items == nil {
		items = []Favorite{}
	}
	c.JSON(http.StatusOK, gin.H{"favorites": items})
}

func (h *HTTPHandler) unfavoriteMatch(c *gin.Context) {
	id, ok := matchID(c)
	if !ok {
		return
	}
	if err := h.app.UnfavoriteMatch(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), id); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func matchID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_match_id"})
		return 0, false
	}
	return id, true
}

func writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, projectfiles.ErrFileNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "project_file_not_found"})
	case errors.Is(err, projectfiles.ErrFileExpired):
		c.JSON(http.StatusGone, gin.H{"error": "project_file_expired"})
	case errors.Is(err, projectfiles.ErrFileTooLarge):
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "project_file_too_large", "max_bytes": projectfiles.DefaultMaxFileSize})
	case errors.Is(err, projectfiles.ErrFileCountExceeded):
		c.JSON(http.StatusConflict, gin.H{"error": "project_file_count_exceeded", "max": projectfiles.DefaultMaxFileCount})
	case errors.Is(err, projectfiles.ErrTotalSizeExceeded):
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "project_file_total_size_exceeded", "max_bytes": projectfiles.DefaultMaxTotalSize})
	case errors.Is(err, projectfiles.ErrUnsupportedMIME):
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported_project_file_type"})
	case errors.Is(err, projectfiles.ErrMIMEMismatch):
		c.JSON(http.StatusBadRequest, gin.H{"error": "project_file_mime_mismatch"})
	case errors.Is(err, projectfiles.ErrInvalidFileName):
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_project_file_name"})
	case errors.Is(err, projectfiles.ErrUnsafeFile):
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "unsafe_project_file"})
	case errors.Is(err, projectfiles.ErrFileNotReady):
		c.JSON(http.StatusConflict, gin.H{"error": "project_file_not_ready"})
	case errors.Is(err, projectfiles.ErrFileAlreadyAttached):
		c.JSON(http.StatusConflict, gin.H{"error": "project_file_duplicate"})
	case errors.Is(err, ErrSessionNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "match_not_found"})
	case errors.Is(err, ErrOpportunityNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "opportunity_not_found"})
	case errors.Is(err, ErrCaseNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "case_not_found"})
	case errors.Is(err, ErrInvalidMatchAnswers):
		c.JSON(http.StatusConflict, gin.H{"error": "invalid_match_answers"})
	case errors.Is(err, ErrInvalidMatchRequest):
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_match_request"})
	case errors.Is(err, ErrMatchRevisionConflict):
		c.JSON(http.StatusConflict, gin.H{"error": "match_revision_conflict"})
	case errors.Is(err, ErrMatchNotReady):
		c.JSON(http.StatusConflict, gin.H{"error": "match_not_ready"})
	case errors.Is(err, ErrStaleMatchGeneration):
		c.JSON(http.StatusConflict, gin.H{"error": "stale_match_generation"})
	case errors.Is(err, ErrComparisonNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "comparison_not_found"})
	case errors.Is(err, ErrInvalidComparison):
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_comparison"})
	case errors.Is(err, ErrExportNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "export_not_found"})
	case errors.Is(err, ErrInvalidExport):
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_export"})
	case errors.Is(err, ErrInvalidCorrection):
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_content_correction"})
	case errors.Is(err, ErrExportExpired):
		c.JSON(http.StatusGone, gin.H{"error": "export_expired"})
	case errors.Is(err, ErrCompareLimit):
		c.JSON(http.StatusConflict, gin.H{"error": "compare_limit_reached", "max": maxProjectCollectionItems})
	case errors.Is(err, ErrInvalidDictionaryKind):
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_dictionary_kind"})
	case errors.Is(err, ErrServiceNotReady):
		c.JSON(http.StatusInternalServerError, gin.H{"error": "service_not_ready"})
	case errors.Is(err, ErrInvalidAIResult):
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid_ai_result"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
	}
}
