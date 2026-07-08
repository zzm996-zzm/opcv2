package crm

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
	CreateCustomer(ctx context.Context, input CreateCustomerInput) (Customer, error)
	ImportLead(ctx context.Context, input ImportLeadInput) (Customer, error)
	ListCustomers(ctx context.Context, input ListCustomersInput) ([]Customer, error)
	GetCustomer(ctx context.Context, userID, customerID int64) (Customer, error)
	UpdateCustomer(ctx context.Context, input UpdateCustomerInput) (Customer, error)
	UpdateStage(ctx context.Context, input UpdateStageInput) (Customer, error)
	RecordFollowUp(ctx context.Context, input RecordFollowUpInput) (FollowUp, error)
	ListActivities(ctx context.Context, userID, customerID int64, limit int) ([]Activity, error)
	ListFollowUps(ctx context.Context, input ListFollowUpsInput) ([]FollowUp, error)
	ListDueCustomers(ctx context.Context, input ListDueInput) ([]Customer, error)
	PipelineStats(ctx context.Context, userID int64) (PipelineStats, error)
	GenerateFollowUpCopy(ctx context.Context, input FollowUpCopyInput) (FollowUpCopy, error)
}

type HTTPHandler struct {
	app Application
}

func NewHTTPHandler(app Application) *HTTPHandler {
	return &HTTPHandler{app: app}
}

func (h *HTTPHandler) Register(router *gin.RouterGroup) {
	router.POST("/crm/customers", h.createCustomer)
	router.POST("/crm/customers/import-lead", h.importLead)
	router.GET("/crm/customers", h.listCustomers)
	router.GET("/crm/customers/due", h.listDueCustomers)
	router.GET("/crm/customers/:id", h.getCustomer)
	router.PATCH("/crm/customers/:id", h.updateCustomer)
	router.GET("/crm/customers/:id/activities", h.listActivities)
	router.POST("/crm/customers/:id/stage", h.updateStage)
	router.POST("/crm/customers/:id/follow-ups", h.recordFollowUp)
	router.POST("/crm/customers/:id/follow-up-copy", h.generateFollowUpCopy)
	router.GET("/crm/follow-ups", h.listFollowUps)
	router.GET("/crm/pipeline-stats", h.pipelineStats)
}

func (h *HTTPHandler) createCustomer(c *gin.Context) {
	var request CreateCustomerInput
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}
	request.UserID = c.GetInt64(auth.UserIDContextKey)
	customer, err := h.app.CreateCustomer(c.Request.Context(), request)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, customer)
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

func (h *HTTPHandler) listCustomers(c *gin.Context) {
	limit, ok := queryLimit(c)
	if !ok {
		return
	}
	customers, err := h.app.ListCustomers(c.Request.Context(), ListCustomersInput{
		UserID: c.GetInt64(auth.UserIDContextKey),
		Stage:  c.Query("stage"),
		Source: c.Query("source"),
		Q:      c.Query("q"),
		Limit:  limit,
	})
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"customers": httpapi.EnsureSlice(customers)})
}

func (h *HTTPHandler) getCustomer(c *gin.Context) {
	id, ok := customerID(c)
	if !ok {
		return
	}
	customer, err := h.app.GetCustomer(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, customer)
}

func (h *HTTPHandler) updateCustomer(c *gin.Context) {
	id, ok := customerID(c)
	if !ok {
		return
	}
	var request UpdateCustomerInput
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}
	request.UserID = c.GetInt64(auth.UserIDContextKey)
	request.CustomerID = id
	customer, err := h.app.UpdateCustomer(c.Request.Context(), request)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, customer)
}

func (h *HTTPHandler) listActivities(c *gin.Context) {
	id, ok := customerID(c)
	if !ok {
		return
	}
	limit, ok := queryLimit(c)
	if !ok {
		return
	}
	activities, err := h.app.ListActivities(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), id, limit)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"activities": httpapi.EnsureSlice(activities)})
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

func (h *HTTPHandler) listFollowUps(c *gin.Context) {
	limit, ok := queryLimit(c)
	if !ok {
		return
	}
	var customerID int64
	if value := c.Query("customer_id"); value != "" {
		parsed, err := strconv.ParseInt(value, 10, 64)
		if err != nil || parsed <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_customer_id"})
			return
		}
		customerID = parsed
	}
	followUps, err := h.app.ListFollowUps(c.Request.Context(), ListFollowUpsInput{
		UserID:     c.GetInt64(auth.UserIDContextKey),
		CustomerID: customerID,
		Q:          c.Query("q"),
		Due:        c.Query("due"),
		Limit:      limit,
	})
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"follow_ups": httpapi.EnsureSlice(followUps)})
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
	limit, ok := queryLimit(c)
	if !ok {
		return
	}
	customers, err := h.app.ListDueCustomers(c.Request.Context(), ListDueInput{
		UserID: c.GetInt64(auth.UserIDContextKey),
		Limit:  limit,
	})
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"customers": httpapi.EnsureSlice(customers)})
}

func (h *HTTPHandler) pipelineStats(c *gin.Context) {
	stats, err := h.app.PipelineStats(c.Request.Context(), c.GetInt64(auth.UserIDContextKey))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, stats)
}

func queryLimit(c *gin.Context) (int, bool) {
	limit := defaultListLimit
	if value := c.Query("limit"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_limit"})
			return 0, false
		}
		limit = parsed
	}
	return limit, true
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
