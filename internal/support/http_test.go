package support

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/zzm/opcv2/internal/auth"
)

type fakeApplication struct {
	input  TicketInput
	ticket Ticket
	rows   []Ticket
	userID int64
	limit  int
	err    error
}

func (a *fakeApplication) CreateTicket(_ context.Context, userID int64, input TicketInput) (Ticket, error) {
	a.userID = userID
	a.input = input
	return a.ticket, a.err
}

func (a *fakeApplication) ListTickets(_ context.Context, userID int64, limit int) ([]Ticket, error) {
	a.userID = userID
	a.limit = limit
	return a.rows, a.err
}

func supportTestRouter(app Application) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	group := router.Group("/api/v1")
	group.Use(func(c *gin.Context) {
		c.Set(auth.UserIDContextKey, int64(42))
		c.Next()
	})
	NewHTTPHandler(app).Register(group)
	return router
}

func TestCreateTicketEndpointUsesAuthenticatedUser(t *testing.T) {
	app := &fakeApplication{ticket: Ticket{ID: 22, Status: StatusOpen}}
	router := supportTestRouter(app)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/support/tickets", strings.NewReader(`{"topic":"套餐与额度","title":"额度没有更新","body":"正文"}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.userID != 42 || app.input.Title != "额度没有更新" {
		t.Fatalf("user/input = %d/%+v", app.userID, app.input)
	}
}

func TestListTicketsEndpointReturnsEmptyArrayAndCapsLimit(t *testing.T) {
	app := &fakeApplication{}
	router := supportTestRouter(app)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/support/tickets?limit=500", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.userID != 42 || app.limit != 100 || !strings.Contains(recorder.Body.String(), `"tickets":[]`) {
		t.Fatalf("user/limit/body = %d/%d/%s", app.userID, app.limit, recorder.Body.String())
	}
}
