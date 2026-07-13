package learning

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
	ListCourses(ctx context.Context, filter CourseFilter) ([]Course, error)
	GetCourse(ctx context.Context, slug string) (Course, error)
	ListCourseMaterials(ctx context.Context, courseSlug string) ([]CourseMaterial, error)
	ListProgress(ctx context.Context, userID int64) ([]Progress, error)
	GetProgress(ctx context.Context, userID int64, courseSlug string) (Progress, error)
	UpdateProgress(ctx context.Context, input UpdateProgressInput) (Progress, error)
	CreateDiagnosis(ctx context.Context, input CreateDiagnosisInput) (Diagnosis, error)
	GetDiagnosis(ctx context.Context, userID, id int64) (Diagnosis, error)
	LatestDiagnosis(ctx context.Context, userID int64) (Diagnosis, error)
	LatestGaps(ctx context.Context, userID int64) (DiagnosisGaps, error)
	LatestRecommendations(ctx context.Context, userID int64) (DiagnosisRecommendations, error)
	LatestPlan(ctx context.Context, userID int64) (DiagnosisPlan, error)
	GetPlan(ctx context.Context, userID, diagnosisID int64) (DiagnosisPlan, error)
	UpdatePlanItem(ctx context.Context, input UpdatePlanItemInput) (PlanItem, error)
	LatestReport(ctx context.Context, userID int64) (DiagnosisReport, error)
}

type HTTPHandler struct {
	app Application
}

func NewHTTPHandler(app Application) *HTTPHandler {
	return &HTTPHandler{app: app}
}

func (h *HTTPHandler) RegisterPublic(router *gin.RouterGroup) {
	router.GET("/learning/courses", h.listCourses)
	router.GET("/learning/courses/:slug", h.getCourse)
	router.GET("/learning/courses/:slug/materials", h.listCourseMaterials)
}

func (h *HTTPHandler) RegisterProtected(router *gin.RouterGroup) {
	router.GET("/learning/progress", h.listProgress)
	router.GET("/learning/progress/:courseSlug", h.getProgress)
	router.PUT("/learning/progress/:courseSlug", h.updateProgress)
	router.POST("/learning/diagnoses", h.createDiagnosis)
	router.POST("/learning/assessments", h.submitAssessment)
	router.GET("/learning/assessments/latest", h.latestDiagnosis)
	router.GET("/learning/assessments/:id", h.getDiagnosis)
	router.GET("/learning/diagnoses/latest", h.latestDiagnosis)
	router.GET("/learning/diagnoses/latest/gaps", h.latestGaps)
	router.GET("/learning/diagnoses/latest/recommendations", h.latestRecommendations)
	router.GET("/learning/diagnoses/latest/plan", h.latestPlan)
	router.GET("/learning/diagnoses/latest/report", h.latestReport)
	router.GET("/learning/diagnoses/:id/plan", h.getPlan)
	router.PUT("/learning/diagnoses/:id/plan/items/:stageNumber", h.updatePlanItem)
	router.GET("/learning/diagnoses/:id", h.getDiagnosis)
}

func (h *HTTPHandler) listCourseMaterials(c *gin.Context) {
	materials, err := h.app.ListCourseMaterials(c.Request.Context(), c.Param("slug"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"materials": httpapi.EnsureSlice(materials)})
}

func (h *HTTPHandler) getProgress(c *gin.Context) {
	progress, err := h.app.GetProgress(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), c.Param("courseSlug"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, progress)
}

func (h *HTTPHandler) updateProgress(c *gin.Context) {
	var request UpdateProgressInput
	if err := c.ShouldBindJSON(&request); err != nil {
		httpapi.BadRequest(c, "invalid_request")
		return
	}
	request.UserID = c.GetInt64(auth.UserIDContextKey)
	request.CourseSlug = c.Param("courseSlug")
	progress, err := h.app.UpdateProgress(c.Request.Context(), request)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, progress)
}

func (h *HTTPHandler) listCourses(c *gin.Context) {
	limit, ok := httpapi.QueryLimit(c, 20, 100)
	if !ok {
		return
	}
	courses, err := h.app.ListCourses(c.Request.Context(), CourseFilter{
		Category: c.Query("category"),
		Limit:    limit,
	})
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"courses": httpapi.EnsureSlice(courses)})
}

func (h *HTTPHandler) getCourse(c *gin.Context) {
	course, err := h.app.GetCourse(c.Request.Context(), c.Param("slug"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, course)
}

func (h *HTTPHandler) listProgress(c *gin.Context) {
	progress, err := h.app.ListProgress(c.Request.Context(), c.GetInt64(auth.UserIDContextKey))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"progress": httpapi.EnsureSlice(progress)})
}

func (h *HTTPHandler) createDiagnosis(c *gin.Context) {
	h.createDiagnosisWithStatus(c, http.StatusOK)
}

func (h *HTTPHandler) submitAssessment(c *gin.Context) {
	h.createDiagnosisWithStatus(c, http.StatusCreated)
}

func (h *HTTPHandler) createDiagnosisWithStatus(c *gin.Context, status int) {
	var request CreateDiagnosisInput
	if err := c.ShouldBindJSON(&request); err != nil {
		httpapi.BadRequest(c, "invalid_request")
		return
	}
	if strings.TrimSpace(request.Goal) == "" || strings.TrimSpace(request.Project) == "" {
		httpapi.BadRequest(c, "invalid_request")
		return
	}
	request.UserID = c.GetInt64(auth.UserIDContextKey)
	diagnosis, err := h.app.CreateDiagnosis(c.Request.Context(), request)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(status, diagnosis)
}

func (h *HTTPHandler) getDiagnosis(c *gin.Context) {
	id, ok := diagnosisID(c)
	if !ok {
		return
	}
	diagnosis, err := h.app.GetDiagnosis(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, diagnosis)
}

func (h *HTTPHandler) getPlan(c *gin.Context) {
	id, ok := diagnosisID(c)
	if !ok {
		return
	}
	plan, err := h.app.GetPlan(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, plan)
}

func (h *HTTPHandler) updatePlanItem(c *gin.Context) {
	id, ok := diagnosisID(c)
	if !ok {
		return
	}
	stageNumber, err := strconv.Atoi(c.Param("stageNumber"))
	if err != nil || stageNumber <= 0 {
		httpapi.BadRequest(c, "invalid_plan_item")
		return
	}
	var request struct {
		Completed *bool `json:"completed"`
	}
	if err := c.ShouldBindJSON(&request); err != nil || request.Completed == nil {
		httpapi.BadRequest(c, "invalid_request")
		return
	}
	item, err := h.app.UpdatePlanItem(c.Request.Context(), UpdatePlanItemInput{
		UserID: c.GetInt64(auth.UserIDContextKey), DiagnosisID: id, StageNumber: stageNumber, Completed: *request.Completed,
	})
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, item)
}

func diagnosisID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		httpapi.BadRequest(c, "invalid_diagnosis_id")
		return 0, false
	}
	return id, true
}

func (h *HTTPHandler) latestDiagnosis(c *gin.Context) {
	diagnosis, err := h.app.LatestDiagnosis(c.Request.Context(), c.GetInt64(auth.UserIDContextKey))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, diagnosis)
}

func (h *HTTPHandler) latestGaps(c *gin.Context) {
	gaps, err := h.app.LatestGaps(c.Request.Context(), c.GetInt64(auth.UserIDContextKey))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gaps)
}

func (h *HTTPHandler) latestRecommendations(c *gin.Context) {
	recommendations, err := h.app.LatestRecommendations(c.Request.Context(), c.GetInt64(auth.UserIDContextKey))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, recommendations)
}

func (h *HTTPHandler) latestPlan(c *gin.Context) {
	plan, err := h.app.LatestPlan(c.Request.Context(), c.GetInt64(auth.UserIDContextKey))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, plan)
}

func (h *HTTPHandler) latestReport(c *gin.Context) {
	report, err := h.app.LatestReport(c.Request.Context(), c.GetInt64(auth.UserIDContextKey))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, report)
}

func writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrCourseNotFound):
		httpapi.Error(c, http.StatusNotFound, "course_not_found")
	case errors.Is(err, ErrProgressNotFound):
		httpapi.Error(c, http.StatusNotFound, "progress_not_found")
	case errors.Is(err, ErrInvalidProgress):
		httpapi.BadRequest(c, "invalid_progress")
	case errors.Is(err, ErrInvalidDiagnosis):
		httpapi.BadRequest(c, "invalid_diagnosis")
	case errors.Is(err, ErrInvalidAIResult):
		httpapi.Error(c, http.StatusInternalServerError, "invalid_ai_result")
	case errors.Is(err, ErrInvalidPlanItem):
		httpapi.BadRequest(c, "invalid_plan_item")
	case errors.Is(err, ErrDiagnosisNotFound):
		httpapi.Error(c, http.StatusNotFound, "diagnosis_not_found")
	case errors.Is(err, ErrServiceNotReady):
		httpapi.Error(c, http.StatusInternalServerError, "service_not_ready")
	default:
		httpapi.Error(c, http.StatusInternalServerError, "internal_error")
	}
}
