package support

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zzm/opcv2/internal/auth"
	"github.com/zzm/opcv2/internal/platform/httpapi"
)

type Application interface {
	CreateTicket(ctx context.Context, userID int64, input TicketInput) (Ticket, error)
	ListTickets(ctx context.Context, userID int64, limit int) ([]Ticket, error)
}

type HTTPHandler struct {
	app Application
}

func NewHTTPHandler(app Application) *HTTPHandler {
	return &HTTPHandler{app: app}
}

func (h *HTTPHandler) Register(router *gin.RouterGroup) {
	router.POST("/support/tickets", h.createTicket)
	router.GET("/support/tickets", h.listTickets)
}

func (h *HTTPHandler) createTicket(c *gin.Context) {
	var request TicketInput
	if err := c.ShouldBindJSON(&request); err != nil {
		httpapi.BadRequest(c, "invalid_request")
		return
	}
	ticket, err := h.app.CreateTicket(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), request)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, ticket)
}

func (h *HTTPHandler) listTickets(c *gin.Context) {
	limit, ok := httpapi.QueryLimit(c, 20, 100)
	if !ok {
		return
	}
	tickets, err := h.app.ListTickets(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), limit)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"tickets": httpapi.EnsureSlice(tickets)})
}

func writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrInvalidInput):
		httpapi.BadRequest(c, "invalid_request")
	case errors.Is(err, ErrUserIDRequired):
		httpapi.Error(c, http.StatusUnauthorized, "unauthorized")
	case errors.Is(err, ErrServiceNotReady):
		httpapi.Error(c, http.StatusInternalServerError, "service_not_ready")
	default:
		httpapi.Error(c, http.StatusInternalServerError, "internal_error")
	}
}
