package enterprise

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/zzm/opcv2/internal/auth"
	"github.com/zzm/opcv2/internal/crm"
	"github.com/zzm/opcv2/internal/platform/httpapi"
)

type Application interface {
	Overview(ctx context.Context, userID int64) (Overview, error)
	PublicOverview(ctx context.Context) (PublicOverview, error)
	ListPublicCases(ctx context.Context, limit int) (PublicCasesResponse, error)
	GetPublicCase(ctx context.Context, slug string) (PublicCase, error)
	CreateInquiry(ctx context.Context, input InquiryInput) (Inquiry, error)
	ContactConfig(ctx context.Context) (ContactConfig, error)
	CreateDiagnosisRequest(ctx context.Context, userID int64, input DiagnosisRequestInput) (DiagnosisRequest, error)
	ListDiagnosisRequests(ctx context.Context, userID int64, limit int) (DiagnosisRequestsResponse, error)
	UpdateDiagnosisRequest(ctx context.Context, userID int64, requestID int64, input DiagnosisRequestUpdateInput) (DiagnosisRequest, error)
	ImportDiagnosisRequestCustomer(ctx context.Context, userID int64, requestID int64) (crm.Customer, error)
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
	router.PATCH("/enterprise/diagnosis-requests/:id", h.updateDiagnosisRequest)
	router.POST("/enterprise/diagnosis-requests/:id/crm-customer", h.importDiagnosisRequestCustomer)
}

func (h *HTTPHandler) RegisterPublic(router *gin.RouterGroup) {
	router.GET("/enterprise/public-overview", h.publicOverview)
	router.GET("/enterprise/cases", h.listPublicCases)
	router.GET("/enterprise/cases/:slug", h.getPublicCase)
	router.POST("/enterprise/inquiries", h.createInquiry)
	router.GET("/enterprise/contact-config", h.contactConfig)
}

func (h *HTTPHandler) overview(c *gin.Context) {
	overview, err := h.app.Overview(c.Request.Context(), c.GetInt64(auth.UserIDContextKey))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, overview)
}

func (h *HTTPHandler) publicOverview(c *gin.Context) {
	overview, err := h.app.PublicOverview(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, overview)
}

func (h *HTTPHandler) listPublicCases(c *gin.Context) {
	limit := 20
	if rawLimit := c.Query("limit"); rawLimit != "" {
		parsed, err := strconv.Atoi(rawLimit)
		if err != nil {
			httpapi.Error(c, http.StatusBadRequest, "invalid_limit")
			return
		}
		limit = parsed
	}
	result, err := h.app.ListPublicCases(c.Request.Context(), limit)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *HTTPHandler) getPublicCase(c *gin.Context) {
	result, err := h.app.GetPublicCase(c.Request.Context(), c.Param("slug"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *HTTPHandler) createInquiry(c *gin.Context) {
	var request InquiryInput
	if err := c.ShouldBindJSON(&request); err != nil {
		httpapi.Error(c, http.StatusBadRequest, "invalid_request")
		return
	}
	result, err := h.app.CreateInquiry(c.Request.Context(), request)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *HTTPHandler) contactConfig(c *gin.Context) {
	result, err := h.app.ContactConfig(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
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

func (h *HTTPHandler) updateDiagnosisRequest(c *gin.Context) {
	requestID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		httpapi.Error(c, http.StatusBadRequest, "invalid_request_id")
		return
	}
	var request DiagnosisRequestUpdateInput
	if err := c.ShouldBindJSON(&request); err != nil {
		httpapi.Error(c, http.StatusBadRequest, "invalid_request")
		return
	}
	result, err := h.app.UpdateDiagnosisRequest(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), requestID, request)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *HTTPHandler) importDiagnosisRequestCustomer(c *gin.Context) {
	requestID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		httpapi.Error(c, http.StatusBadRequest, "invalid_request_id")
		return
	}
	result, err := h.app.ImportDiagnosisRequestCustomer(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), requestID)
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
	case errors.Is(err, ErrInvalidInput), errors.Is(err, crm.ErrInvalidInput):
		httpapi.Error(c, http.StatusBadRequest, "invalid_input")
	case errors.Is(err, ErrDiagnosisRequestNotFound):
		httpapi.Error(c, http.StatusNotFound, "diagnosis_request_not_found")
	case errors.Is(err, ErrPublicCaseNotFound):
		httpapi.Error(c, http.StatusNotFound, "enterprise_case_not_found")
	case errors.Is(err, crm.ErrServiceNotReady):
		httpapi.Error(c, http.StatusInternalServerError, "service_not_ready")
	default:
		httpapi.Error(c, http.StatusInternalServerError, "internal_error")
	}
}
