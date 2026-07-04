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
	plans    []PlanOption
	usage    []UsageItem
	orders   []Order
	checkout CheckoutResult
	input    CheckoutInput
	userID   int64
	limit    int
	err      error
}

func (a *fakeApplication) CurrentSnapshot(_ context.Context, _ int64) (Snapshot, error) {
	return a.snapshot, a.err
}

func (a *fakeApplication) Redeem(_ context.Context, _ RedeemInput) (RedeemResult, error) {
	return a.result, a.err
}

func (a *fakeApplication) ListPlans(context.Context) ([]PlanOption, error) {
	return a.plans, a.err
}

func (a *fakeApplication) CurrentUsage(_ context.Context, userID int64) ([]UsageItem, error) {
	a.userID = userID
	return a.usage, a.err
}

func (a *fakeApplication) ListOrders(_ context.Context, userID int64, limit int) ([]Order, error) {
	a.userID = userID
	a.limit = limit
	return a.orders, a.err
}

func (a *fakeApplication) CreateCheckout(_ context.Context, input CheckoutInput) (CheckoutResult, error) {
	a.input = input
	return a.checkout, a.err
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

func TestListPlansEndpoint(t *testing.T) {
	router := testRouter(&fakeApplication{plans: []PlanOption{{Code: PlanPro, Name: "会员版", PriceCents: 6900}}})
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/membership/plans", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"plans"`) {
		t.Fatalf("body = %s", recorder.Body.String())
	}
}

func TestUsageAndOrdersEndpointsUseAuthenticatedUser(t *testing.T) {
	app := &fakeApplication{usage: []UsageItem{{Key: "lead_tasks"}}, orders: []Order{{ID: 12}}}
	router := testRouter(app)

	usageRecorder := httptest.NewRecorder()
	router.ServeHTTP(usageRecorder, httptest.NewRequest(http.MethodGet, "/api/v1/membership/usage", nil))
	if usageRecorder.Code != http.StatusOK || app.userID != 42 {
		t.Fatalf("usage status/user/body = %d/%d/%s", usageRecorder.Code, app.userID, usageRecorder.Body.String())
	}

	ordersRecorder := httptest.NewRecorder()
	router.ServeHTTP(ordersRecorder, httptest.NewRequest(http.MethodGet, "/api/v1/membership/orders?limit=500", nil))
	if ordersRecorder.Code != http.StatusOK || app.limit != 100 {
		t.Fatalf("orders status/limit/body = %d/%d/%s", ordersRecorder.Code, app.limit, ordersRecorder.Body.String())
	}
}

func TestCheckoutEndpointUsesAuthenticatedUser(t *testing.T) {
	app := &fakeApplication{checkout: CheckoutResult{
		Order:   Order{ID: 13, OrderNo: "ZS-20260702-0013", Status: OrderPending},
		Payment: PaymentInfo{Mode: "manual"},
	}}
	router := testRouter(app)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/membership/checkout", strings.NewReader(`{"plan_code":"pro","billing_cycle":"month"}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.input.UserID != 42 || app.input.PlanCode != PlanPro || app.input.BillingCycle != "month" {
		t.Fatalf("input = %+v", app.input)
	}
}
