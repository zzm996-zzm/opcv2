package learning

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/zzm/opcv2/internal/auth"
	"github.com/zzm/opcv2/internal/platform/httpapi"
)

type Application interface {
	ListCourses(ctx context.Context, filter CourseFilter) ([]Course, error)
	GetCourse(ctx context.Context, slug string) (Course, error)
	ListProgress(ctx context.Context, userID int64) ([]Progress, error)
	CreateDiagnosis(ctx context.Context, input CreateDiagnosisInput) (Diagnosis, error)
	LatestDiagnosis(ctx context.Context, userID int64) (Diagnosis, error)
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
}

func (h *HTTPHandler) RegisterProtected(router *gin.RouterGroup) {
	router.GET("/learning/progress", h.listProgress)
	router.POST("/learning/diagnoses", h.createDiagnosis)
	router.GET("/learning/diagnoses/latest", h.latestDiagnosis)
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
	c.JSON(http.StatusOK, diagnosis)
}

func (h *HTTPHandler) latestDiagnosis(c *gin.Context) {
	diagnosis, err := h.app.LatestDiagnosis(c.Request.Context(), c.GetInt64(auth.UserIDContextKey))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, diagnosis)
}

func writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrCourseNotFound):
		httpapi.Error(c, http.StatusNotFound, "course_not_found")
	case errors.Is(err, ErrDiagnosisNotFound):
		httpapi.Error(c, http.StatusNotFound, "diagnosis_not_found")
	case errors.Is(err, ErrServiceNotReady):
		httpapi.Error(c, http.StatusInternalServerError, "service_not_ready")
	default:
		httpapi.Error(c, http.StatusInternalServerError, "internal_error")
	}
}
