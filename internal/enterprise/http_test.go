package enterprise

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
	overview Overview
	userID   int64
	err      error
}

func (a *fakeApplication) Overview(_ context.Context, userID int64) (Overview, error) {
	a.userID = userID
	return a.overview, a.err
}

func enterpriseTestRouter(app Application) *gin.Engine {
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

func TestOverviewEndpointUsesAuthenticatedUser(t *testing.T) {
	app := &fakeApplication{overview: Overview{
		Stats: []Metric{{Key: "companies", Label: "服务企业数", Value: "2"}},
	}}
	router := enterpriseTestRouter(app)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/enterprise/overview", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.userID != 42 || !strings.Contains(recorder.Body.String(), `"stats":[`) {
		t.Fatalf("user/body = %d/%s", app.userID, recorder.Body.String())
	}
}
