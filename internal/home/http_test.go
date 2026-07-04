package home

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
	summary Summary
	userID  int64
	err     error
}

func (a *fakeApplication) Summary(_ context.Context, userID int64) (Summary, error) {
	a.userID = userID
	return a.summary, a.err
}

func homeTestRouter(app Application) *gin.Engine {
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

func TestSummaryEndpointUsesAuthenticatedUser(t *testing.T) {
	app := &fakeApplication{summary: Summary{
		AccountSummary: AccountSummary{PlanName: "会员版"},
	}}
	router := homeTestRouter(app)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/home/summary", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.userID != 42 || !strings.Contains(recorder.Body.String(), `"plan_name":"会员版"`) {
		t.Fatalf("user/body = %d/%s", app.userID, recorder.Body.String())
	}
}
