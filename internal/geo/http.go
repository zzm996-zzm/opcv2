package geo

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/zzm/opcv2/internal/auth"
	"github.com/zzm/opcv2/internal/platform/httpapi"
)

type Application interface {
	Overview(ctx context.Context, userID int64) (Overview, error)
	CreateAnalysisRequest(ctx context.Context, userID int64, input AnalysisRequestInput) (AnalysisRequest, error)
	ListAnalysisRequests(ctx context.Context, userID int64, limit int) ([]AnalysisRequest, error)
	GetAnalysisRequest(ctx context.Context, userID, id int64) (AnalysisRequest, error)
}

type HTTPHandler struct {
	app Application
}

func NewHTTPHandler(app Application) *HTTPHandler {
	return &HTTPHandler{app: app}
}

func (h *HTTPHandler) Register(router *gin.RouterGroup) {
	router.GET("/geo/overview", h.overview)
	router.POST("/geo/analysis-requests", h.createAnalysisRequest)
	router.GET("/geo/analysis-requests", h.listAnalysisRequests)
	router.GET("/geo/analysis-requests/:id", h.getAnalysisRequest)
}

func (h *HTTPHandler) overview(c *gin.Context) {
	overview, err := h.app.Overview(c.Request.Context(), c.GetInt64(auth.UserIDContextKey))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, overview)
}

func (h *HTTPHandler) createAnalysisRequest(c *gin.Context) {
	var request AnalysisRequestInput
	if err := c.ShouldBindJSON(&request); err != nil {
		httpapi.BadRequest(c, "invalid_request")
		return
	}
	analysisRequest, err := h.app.CreateAnalysisRequest(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), request)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, analysisRequest)
}

func (h *HTTPHandler) listAnalysisRequests(c *gin.Context) {
	limit, ok := httpapi.QueryLimit(c, 20, 100)
	if !ok {
		return
	}
	requests, err := h.app.ListAnalysisRequests(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), limit)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"requests": httpapi.EnsureSlice(requests)})
}

func (h *HTTPHandler) getAnalysisRequest(c *gin.Context) {
	id, ok := analysisRequestID(c)
	if !ok {
		return
	}
	request, err := h.app.GetAnalysisRequest(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, request)
}

func analysisRequestID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		httpapi.BadRequest(c, "invalid_analysis_request_id")
		return 0, false
	}
	return id, true
}

func writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrAnalysisRequestNotFound):
		httpapi.Error(c, http.StatusNotFound, "analysis_request_not_found")
	case errors.Is(err, ErrInvalidAnalysisRequest):
		httpapi.BadRequest(c, "invalid_request")
	case errors.Is(err, ErrInvalidAnalysisRequestID):
		httpapi.BadRequest(c, "invalid_analysis_request_id")
	case errors.Is(err, ErrUserIDRequired):
		httpapi.Error(c, http.StatusUnauthorized, "unauthorized")
	case errors.Is(err, ErrServiceNotReady):
		httpapi.Error(c, http.StatusInternalServerError, "service_not_ready")
	default:
		httpapi.Error(c, http.StatusInternalServerError, "internal_error")
	}
}
