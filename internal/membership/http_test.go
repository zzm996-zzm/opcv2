package membership

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

type fakeApplication struct {
	snapshot Snapshot
	result   RedeemResult
	err      error
}

func (a *fakeApplication) CurrentSnapshot(_ context.Context, _ int64) (Snapshot, error) {
	return a.snapshot, a.err
}

func (a *fakeApplication) Redeem(_ context.Context, _ RedeemInput) (RedeemResult, error) {
	return a.result, a.err
}

func testRouter(app Application) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := NewHTTPHandler(app)
	group := router.Group("/api/v1")
	group.Use(func(c *gin.Context) {
		c.Set("auth_user_id", int64(42))
		c.Next()
	})
	handler.Register(group)
	return router
}

func TestCurrentMembershipEndpoint(t *testing.T) {
	router := testRouter(&fakeApplication{snapshot: Snapshot{
		Plan:          PlanCatalog[PlanFree],
		CreditBalance: 25,
	}})
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/membership/me", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"credit_balance":25`) {
		t.Fatalf("body = %s", recorder.Body.String())
	}
}

func TestRedeemEndpoint(t *testing.T) {
	now := time.Now()
	router := testRouter(&fakeApplication{result: RedeemResult{
		Snapshot: Snapshot{Plan: PlanCatalog[PlanPro], CreditBalance: 100},
		Redemption: Redemption{
			UserID:    42,
			Code:      "PRO100",
			CreatedAt: now,
		},
	}})
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/redemptions/redeem",
		strings.NewReader(`{"code":"PRO100"}`),
	)
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"plan"`) {
		t.Fatalf("body = %s", recorder.Body.String())
	}
}
