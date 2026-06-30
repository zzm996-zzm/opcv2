package dashboard

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
	userID  int64
	summary Summary
	err     error
}

func (a *fakeApplication) GetSummary(_ context.Context, userID int64) (Summary, error) {
	a.userID = userID
	return a.summary, a.err
}

func dashboardTestRouter(app Application) *gin.Engine {
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
		Metrics: []Metric{{Label: "新增线索", Value: "328", Change: "+41%"}},
	}}
	router := dashboardTestRouter(app)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard/summary", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.userID != 42 {
		t.Fatalf("userID = %d, want 42", app.userID)
	}
	if !strings.Contains(recorder.Body.String(), `"metrics"`) {
		t.Fatalf("body = %s", recorder.Body.String())
	}
}
