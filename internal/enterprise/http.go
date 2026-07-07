package enterprise

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
	CreateDiagnosisRequest(ctx context.Context, userID int64, input DiagnosisRequestInput) (DiagnosisRequest, error)
	ListDiagnosisRequests(ctx context.Context, userID int64, limit int) (DiagnosisRequestsResponse, error)
}

type HTTPHandler struct {
	app Application
}

func NewHTTPHandler(app Application) *HTTPHandler {
	return &HTTPHandler{app: app}
}

func (h *HTTPHandler) Register(router *gin.RouterGroup) {
	router.GET("/enterprise/overview", h.overview)
	router.GET("/enterprise/diagnosis-requests", h.listDiagnosisRequests)
	router.POST("/enterprise/diagnosis-requests", h.createDiagnosisRequest)
}

func (h *HTTPHandler) overview(c *gin.Context) {
	overview, err := h.app.Overview(c.Request.Context(), c.GetInt64(auth.UserIDContextKey))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, overview)
}

func (h *HTTPHandler) createDiagnosisRequest(c *gin.Context) {
	var request DiagnosisRequestInput
	if err := c.ShouldBindJSON(&request); err != nil {
		httpapi.Error(c, http.StatusBadRequest, "invalid_request")
		return
	}
	result, err := h.app.CreateDiagnosisRequest(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), request)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *HTTPHandler) listDiagnosisRequests(c *gin.Context) {
	limit := 10
	if rawLimit := c.Query("limit"); rawLimit != "" {
		parsed, err := strconv.Atoi(rawLimit)
		if err != nil {
			httpapi.Error(c, http.StatusBadRequest, "invalid_limit")
			return
		}
		limit = parsed
	}
	result, err := h.app.ListDiagnosisRequests(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), limit)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrUserIDRequired):
		httpapi.Error(c, http.StatusUnauthorized, "unauthorized")
	case errors.Is(err, ErrInvalidInput):
		httpapi.Error(c, http.StatusBadRequest, "invalid_input")
	default:
		httpapi.Error(c, http.StatusInternalServerError, "internal_error")
	}
}
