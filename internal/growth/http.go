package growth

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/zzm/opcv2/internal/auth"
	"github.com/zzm/opcv2/internal/platform/httpapi"
)

type Application interface {
	CreateDraft(ctx context.Context, input CreateDraftInput) (Draft, error)
	GetDraft(ctx context.Context, userID, id int64) (Draft, error)
	AnswerDraft(ctx context.Context, input AnswerDraftInput) (Draft, error)
	CalculateDraft(ctx context.Context, input CalculateDraftInput) (DraftCalculation, error)
	CreateModel(ctx context.Context, input CreateInput) (Model, error)
	ListModels(ctx context.Context, userID int64, limit int) ([]Model, error)
	GetModel(ctx context.Context, userID, id int64) (Model, error)
	ModelScenarios(ctx context.Context, userID, id int64) (GrowthScenarios, error)
	ModelForecast(ctx context.Context, userID, id int64) (GrowthForecast, error)
	ModelRecommendations(ctx context.Context, userID, id int64) (GrowthRecommendations, error)
	ListSnapshots(ctx context.Context, userID, modelID int64, limit int) ([]ModelSnapshot, error)
}

type HTTPHandler struct {
	app Application
}

type modelPageApplication interface {
	ListModelPage(context.Context, ListModelsInput) (ModelPage, error)
}

type recalculateApplication interface {
	RecalculateModel(context.Context, int64, int64) (RecalculateResult, error)
}

type exportApplication interface {
	ExportModel(context.Context, ExportModelInput) (GrowthReportExport, error)
}

type compareApplication interface {
	CompareModels(context.Context, CompareModelsInput) (GrowthComparison, error)
}

type risksApplication interface {
	ModelRisks(context.Context, int64, int64) (GrowthRisks, error)
}

type actionPlanApplication interface {
	ModelActionPlan(context.Context, int64, int64) (GrowthActionPlan, error)
}

type inputsApplication interface {
	ModelInputs(context.Context, int64, int64) (GrowthInputs, error)
}

func NewHTTPHandler(app Application) *HTTPHandler {
	return &HTTPHandler{app: app}
}

func (h *HTTPHandler) Register(router *gin.RouterGroup) {
	router.POST("/growth/drafts", h.createDraft)
	router.GET("/growth/drafts/:id", h.getDraft)
	router.POST("/growth/drafts/:id/answers", h.answerDraft)
	router.POST("/growth/drafts/:id/calculate", h.calculateDraft)
	router.POST("/growth/models", h.createModel)
	router.POST("/growth/models/compare", h.compareModels)
	router.GET("/growth/models", h.listModels)
	router.GET("/growth/models/:id", h.getModel)
	router.GET("/growth/models/:id/scenarios", h.modelScenarios)
	router.GET("/growth/models/:id/forecast", h.modelForecast)
	router.GET("/growth/models/:id/recommendations", h.modelRecommendations)
	router.GET("/growth/models/:id/risks", h.modelRisks)
	router.GET("/growth/models/:id/action-plan", h.modelActionPlan)
	router.GET("/growth/models/:id/inputs", h.modelInputs)
	router.GET("/growth/models/:id/snapshots", h.listSnapshots)
	router.POST("/growth/models/:id/recalculate", h.recalculateModel)
	router.POST("/growth/models/:id/export", h.exportModel)
}

func (h *HTTPHandler) createDraft(c *gin.Context) {
	var request CreateDraftInput
	if err := c.ShouldBindJSON(&request); err != nil || strings.TrimSpace(request.Input) == "" || len([]rune(request.Input)) > 2000 {
		httpapi.BadRequest(c, "invalid_request")
		return
	}
	request.UserID = c.GetInt64(auth.UserIDContextKey)
	draft, err := h.app.CreateDraft(c.Request.Context(), request)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, draft)
}

func (h *HTTPHandler) getDraft(c *gin.Context) {
	id, ok := resourceIDParam(c, "invalid_draft_id")
	if !ok {
		return
	}
	draft, err := h.app.GetDraft(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, draft)
}

func (h *HTTPHandler) answerDraft(c *gin.Context) {
	id, ok := resourceIDParam(c, "invalid_draft_id")
	if !ok {
		return
	}
	var request AnswerDraftInput
	if err := c.ShouldBindJSON(&request); err != nil || len(request.Answers) == 0 {
		httpapi.BadRequest(c, "invalid_request")
		return
	}
	request.UserID, request.DraftID = c.GetInt64(auth.UserIDContextKey), id
	draft, err := h.app.AnswerDraft(c.Request.Context(), request)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, draft)
}

func (h *HTTPHandler) calculateDraft(c *gin.Context) {
	id, ok := resourceIDParam(c, "invalid_draft_id")
	if !ok {
		return
	}
	var request CalculateDraftInput
	if err := c.ShouldBindJSON(&request); err != nil {
		httpapi.BadRequest(c, "invalid_request")
		return
	}
	request.UserID, request.DraftID = c.GetInt64(auth.UserIDContextKey), id
	calculation, err := h.app.CalculateDraft(c.Request.Context(), request)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, calculation)
}

func (h *HTTPHandler) createModel(c *gin.Context) {
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
	model, err := h.app.CreateModel(c.Request.Context(), request)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, model)
}

func validCreateInput(input CreateInput) bool {
	return strings.TrimSpace(input.Name) != "" &&
		input.MonthlyVisits > 0 &&
		input.LeadRate >= 0 &&
		input.LeadRate <= 1 &&
		input.DealRate >= 0 &&
		input.DealRate <= 1 &&
		input.AverageOrder > 0 &&
		input.AcquisitionCost >= 0 &&
		input.DeliveryCost >= 0
}

func (h *HTTPHandler) listModels(c *gin.Context) {
	limit, ok := httpapi.QueryLimit(c, 20, 100)
	if !ok {
		return
	}
	offset, ok := httpapi.QueryOffset(c)
	if !ok {
		return
	}
	query := strings.TrimSpace(c.Query("q"))
	if len([]rune(query)) > 100 {
		httpapi.BadRequest(c, "invalid_query")
		return
	}
	if app, supportsPaging := h.app.(modelPageApplication); supportsPaging {
		page, err := app.ListModelPage(c.Request.Context(), ListModelsInput{
			UserID: c.GetInt64(auth.UserIDContextKey), Query: query, Limit: limit, Offset: offset,
		})
		if err != nil {
			writeError(c, err)
			return
		}
		page.Models = httpapi.EnsureSlice(page.Models)
		c.JSON(http.StatusOK, page)
		return
	}
	models, err := h.app.ListModels(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), limit)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"models": httpapi.EnsureSlice(models), "total": len(models), "limit": limit, "offset": offset})
}

func (h *HTTPHandler) getModel(c *gin.Context) {
	id, ok := modelIDParam(c)
	if !ok {
		return
	}
	model, err := h.app.GetModel(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, model)
}

func (h *HTTPHandler) modelScenarios(c *gin.Context) {
	id, ok := modelIDParam(c)
	if !ok {
		return
	}
	scenarios, err := h.app.ModelScenarios(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, scenarios)
}

func (h *HTTPHandler) modelForecast(c *gin.Context) {
	id, ok := modelIDParam(c)
	if !ok {
		return
	}
	forecast, err := h.app.ModelForecast(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, forecast)
}

func (h *HTTPHandler) modelRecommendations(c *gin.Context) {
	id, ok := modelIDParam(c)
	if !ok {
		return
	}
	recommendations, err := h.app.ModelRecommendations(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, recommendations)
}

func (h *HTTPHandler) modelRisks(c *gin.Context) {
	id, ok := modelIDParam(c)
	if !ok {
		return
	}
	app, ok := h.app.(risksApplication)
	if !ok {
		httpapi.Error(c, http.StatusInternalServerError, "service_not_ready")
		return
	}
	risks, err := app.ModelRisks(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), id)
	if err != nil {
		writeError(c, err)
		return
	}
	risks.Risks = httpapi.EnsureSlice(risks.Risks)
	c.JSON(http.StatusOK, risks)
}

func (h *HTTPHandler) modelActionPlan(c *gin.Context) {
	id, ok := modelIDParam(c)
	if !ok {
		return
	}
	app, ok := h.app.(actionPlanApplication)
	if !ok {
		httpapi.Error(c, http.StatusInternalServerError, "service_not_ready")
		return
	}
	plan, err := app.ModelActionPlan(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), id)
	if err != nil {
		writeError(c, err)
		return
	}
	plan.Phases = httpapi.EnsureSlice(plan.Phases)
	for index := range plan.Phases {
		plan.Phases[index].Items = httpapi.EnsureSlice(plan.Phases[index].Items)
	}
	c.JSON(http.StatusOK, plan)
}

func (h *HTTPHandler) modelInputs(c *gin.Context) {
	id, ok := modelIDParam(c)
	if !ok {
		return
	}
	app, ok := h.app.(inputsApplication)
	if !ok {
		httpapi.Error(c, http.StatusInternalServerError, "service_not_ready")
		return
	}
	inputs, err := app.ModelInputs(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), id)
	if err != nil {
		writeError(c, err)
		return
	}
	inputs.Fields = httpapi.EnsureSlice(inputs.Fields)
	c.JSON(http.StatusOK, inputs)
}

func (h *HTTPHandler) listSnapshots(c *gin.Context) {
	id, ok := modelIDParam(c)
	if !ok {
		return
	}
	limit, ok := httpapi.QueryLimit(c, 20, 100)
	if !ok {
		return
	}
	snapshots, err := h.app.ListSnapshots(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), id, limit)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"snapshots": httpapi.EnsureSlice(snapshots)})
}

func (h *HTTPHandler) recalculateModel(c *gin.Context) {
	id, ok := modelIDParam(c)
	if !ok {
		return
	}
	app, ok := h.app.(recalculateApplication)
	if !ok {
		httpapi.Error(c, http.StatusInternalServerError, "service_not_ready")
		return
	}
	result, err := app.RecalculateModel(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *HTTPHandler) exportModel(c *gin.Context) {
	id, ok := modelIDParam(c)
	if !ok {
		return
	}
	app, ok := h.app.(exportApplication)
	if !ok {
		httpapi.Error(c, http.StatusInternalServerError, "service_not_ready")
		return
	}
	var request ExportModelInput
	if err := c.ShouldBindJSON(&request); err != nil {
		httpapi.BadRequest(c, "invalid_request")
		return
	}
	request.UserID = c.GetInt64(auth.UserIDContextKey)
	request.ModelID = id
	if strings.ToLower(strings.TrimSpace(request.Format)) != "json" {
		httpapi.BadRequest(c, "invalid_export_format")
		return
	}
	export, err := app.ExportModel(c.Request.Context(), request)
	if err != nil {
		writeError(c, err)
		return
	}
	c.Header("Content-Type", "application/json; charset=utf-8")
	c.Header("Content-Disposition", `attachment; filename="growth-report-`+strconv.FormatInt(id, 10)+`.json"`)
	c.JSON(http.StatusOK, export)
}

func (h *HTTPHandler) compareModels(c *gin.Context) {
	app, ok := h.app.(compareApplication)
	if !ok {
		httpapi.Error(c, http.StatusInternalServerError, "service_not_ready")
		return
	}
	var request CompareModelsInput
	if err := c.ShouldBindJSON(&request); err != nil {
		httpapi.BadRequest(c, "invalid_request")
		return
	}
	request.UserID = c.GetInt64(auth.UserIDContextKey)
	comparison, err := app.CompareModels(c.Request.Context(), request)
	if err != nil {
		writeError(c, err)
		return
	}
	comparison.Models = httpapi.EnsureSlice(comparison.Models)
	c.JSON(http.StatusOK, comparison)
}

func modelIDParam(c *gin.Context) (int64, bool) {
	return resourceIDParam(c, "invalid_model_id")
}

func resourceIDParam(c *gin.Context, code string) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		httpapi.BadRequest(c, code)
		return 0, false
	}
	return id, true
}

func writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrModelNotFound):
		httpapi.Error(c, http.StatusNotFound, "model_not_found")
	case errors.Is(err, ErrDraftNotFound):
		httpapi.Error(c, http.StatusNotFound, "draft_not_found")
	case errors.Is(err, ErrDraftNotReady):
		httpapi.Error(c, http.StatusConflict, "draft_not_ready")
	case errors.Is(err, ErrInvalidAnswers):
		httpapi.BadRequest(c, "invalid_answers")
	case errors.Is(err, ErrInvalidExportFormat):
		httpapi.BadRequest(c, "invalid_export_format")
	case errors.Is(err, ErrInvalidComparison):
		httpapi.BadRequest(c, "invalid_comparison")
	case errors.Is(err, ErrServiceNotReady):
		httpapi.Error(c, http.StatusInternalServerError, "service_not_ready")
	default:
		httpapi.Error(c, http.StatusInternalServerError, "internal_error")
	}
}
