package crm

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/zzm/opcv2/internal/auth"
)

type Application interface {
	ImportLead(ctx context.Context, input ImportLeadInput) (Customer, error)
	UpdateStage(ctx context.Context, input UpdateStageInput) (Customer, error)
	RecordFollowUp(ctx context.Context, input RecordFollowUpInput) (FollowUp, error)
	ListDueCustomers(ctx context.Context, input ListDueInput) ([]Customer, error)
	GenerateFollowUpCopy(ctx context.Context, input FollowUpCopyInput) (FollowUpCopy, error)
}

type HTTPHandler struct {
	app Application
}

func NewHTTPHandler(app Application) *HTTPHandler {
	return &HTTPHandler{app: app}
}

func (h *HTTPHandler) Register(router *gin.RouterGroup) {
	router.POST("/crm/customers/import-lead", h.importLead)
	router.POST("/crm/customers/:id/stage", h.updateStage)
	router.POST("/crm/customers/:id/follow-ups", h.recordFollowUp)
	router.POST("/crm/customers/:id/follow-up-copy", h.generateFollowUpCopy)
	router.GET("/crm/customers/due", h.listDueCustomers)
}

func (h *HTTPHandler) importLead(c *gin.Context) {
	var request ImportLeadInput
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}
	request.UserID = c.GetInt64(auth.UserIDContextKey)
	customer, err := h.app.ImportLead(c.Request.Context(), request)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, customer)
}

func (h *HTTPHandler) updateStage(c *gin.Context) {
	id, ok := customerID(c)
	if !ok {
		return
	}
	var request UpdateStageInput
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}
	request.UserID = c.GetInt64(auth.UserIDContextKey)
	request.CustomerID = id
	customer, err := h.app.UpdateStage(c.Request.Context(), request)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, customer)
}

func (h *HTTPHandler) recordFollowUp(c *gin.Context) {
	id, ok := customerID(c)
	if !ok {
		return
	}
	var request RecordFollowUpInput
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}
	request.UserID = c.GetInt64(auth.UserIDContextKey)
	request.CustomerID = id
	followUp, err := h.app.RecordFollowUp(c.Request.Context(), request)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, followUp)
}

func (h *HTTPHandler) generateFollowUpCopy(c *gin.Context) {
	id, ok := customerID(c)
	if !ok {
		return
	}
	var request FollowUpCopyInput
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}
	request.UserID = c.GetInt64(auth.UserIDContextKey)
	request.CustomerID = id
	copy, err := h.app.GenerateFollowUpCopy(c.Request.Context(), request)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, copy)
}

func (h *HTTPHandler) listDueCustomers(c *gin.Context) {
	limit := defaultListLimit
	if value := c.Query("limit"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_limit"})
			return
		}
		limit = parsed
	}
	customers, err := h.app.ListDueCustomers(c.Request.Context(), ListDueInput{
		UserID: c.GetInt64(auth.UserIDContextKey),
		Limit:  limit,
	})
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"customers": customers})
}

func customerID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_customer_id"})
		return 0, false
	}
	return id, true
}

func writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrInvalidInput):
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_crm_input"})
	case errors.Is(err, ErrCustomerNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "customer_not_found"})
	case errors.Is(err, ErrServiceNotReady):
		c.JSON(http.StatusInternalServerError, gin.H{"error": "service_not_ready"})
	case errors.Is(err, ErrInvalidAIResult):
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid_ai_result"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
	}
}
