package membership

import (
	"context"
	"errors"
	"testing"
	"time"
)

type memoryRepository struct {
	account       CreditAccount
	transactions  []CreditTransaction
	codes         map[string]*RedemptionCode
	redemptions   map[string]Redemption
	subscriptions map[int64]Subscription
	plans         []PlanOption
	usage         []UsageItem
	consumed      []ConsumeInput
	refunded      []ConsumeInput
	consumeResult UsageItem
	consumeErr    error
	refundResult  UsageItem
	refundErr     error
	orders        []Order
	checkout      CheckoutInput
}

func newMemoryRepository() *memoryRepository {
	return &memoryRepository{
		codes:         map[string]*RedemptionCode{},
		redemptions:   map[string]Redemption{},
		subscriptions: map[int64]Subscription{},
	}
}

func (r *memoryRepository) CurrentSnapshot(_ context.Context, userID int64, now time.Time) (Snapshot, error) {
	subscription := r.subscriptions[userID]
	planCode := PlanFree
	if subscription.Status == "active" && (subscription.EndsAt == nil || subscription.EndsAt.After(now)) {
		planCode = subscription.PlanCode
	}
	return Snapshot{
		Plan:          PlanCatalog[planCode],
		Subscription:  subscription,
		CreditBalance: r.account.Balance,
	}, nil
}

func (r *memoryRepository) RedeemCode(ctx context.Context, input RedeemInput, now time.Time) (RedeemResult, error) {
	code, ok := r.codes[input.Code]
	if !ok || !code.Active {
		return RedeemResult{}, ErrCodeNotFound
	}
	if code.ExpiresAt != nil && !code.ExpiresAt.After(now) {
		return RedeemResult{}, ErrCodeExpired
	}
	key := input.Code + ":" + string(rune(input.UserID))
	if redemption, ok := r.redemptions[key]; ok {
		snapshot, err := r.CurrentSnapshot(ctx, input.UserID, now)
		if err != nil {
			return RedeemResult{}, err
		}
		return RedeemResult{Snapshot: snapshot, Redemption: redemption, AlreadyRedeemed: true}, nil
	}
	if code.MaxRedemptions > 0 && code.RedeemedCount >= code.MaxRedemptions {
		return RedeemResult{}, ErrCodeExhausted
	}

	if code.CreditAmount > 0 {
		r.account.UserID = input.UserID
		r.account.Balance += code.CreditAmount
		r.transactions = append(r.transactions, CreditTransaction{
			UserID: input.UserID,
			Amount: code.CreditAmount,
			Reason: "redemption",
		})
	}
	if code.PlanCode != "" {
		r.subscriptions[input.UserID] = Subscription{
			UserID:   input.UserID,
			PlanCode: code.PlanCode,
			Status:   "active",
			StartsAt: now,
		}
	}
	code.RedeemedCount++
	redemption := Redemption{UserID: input.UserID, Code: input.Code, CreatedAt: now}
	r.redemptions[key] = redemption
	snapshot, err := r.CurrentSnapshot(ctx, input.UserID, now)
	if err != nil {
		return RedeemResult{}, err
	}
	return RedeemResult{Snapshot: snapshot, Redemption: redemption}, nil
}

func (r *memoryRepository) ListPlans(context.Context) ([]PlanOption, error) {
	return r.plans, nil
}

func (r *memoryRepository) CurrentUsage(context.Context, int64, time.Time) ([]UsageItem, error) {
	return r.usage, nil
}

func (r *memoryRepository) CheckAndConsume(_ context.Context, input ConsumeInput, _ time.Time) (UsageItem, error) {
	r.consumed = append(r.consumed, input)
	if r.consumeErr != nil {
		return UsageItem{}, r.consumeErr
	}
	return r.consumeResult, nil
}

func (r *memoryRepository) RefundUsage(_ context.Context, input ConsumeInput, _ time.Time) (UsageItem, error) {
	r.refunded = append(r.refunded, input)
	if r.refundErr != nil {
		return UsageItem{}, r.refundErr
	}
	return r.refundResult, nil
}

func (r *memoryRepository) ListOrders(_ context.Context, _ int64, limit int) ([]Order, error) {
	if limit > len(r.orders) {
		limit = len(r.orders)
	}
	return r.orders[:limit], nil
}

func (r *memoryRepository) CreateCheckout(_ context.Context, input CheckoutInput, _ time.Time) (CheckoutResult, error) {
	r.checkout = input
	for _, plan := range r.plans {
		if plan.Code == input.PlanCode && plan.BillingCycle == input.BillingCycle {
			order := Order{ID: 13, OrderNo: input.OrderNo, PlanCode: plan.Code, AmountCents: plan.PriceCents, Status: OrderPending}
			return CheckoutResult{
				Order:   order,
				Payment: PaymentInfo{Mode: "manual", Message: "请联系顾问完成开通"},
			}, nil
		}
	}
	return CheckoutResult{}, ErrPlanNotFound
}

func TestServiceReturnsFreeSnapshotForNewUser(t *testing.T) {
	service := NewService(newMemoryRepository())

	snapshot, err := service.CurrentSnapshot(context.Background(), 42)
	if err != nil {
		t.Fatalf("CurrentSnapshot() error = %v", err)
	}
	if snapshot.Plan.Code != PlanFree || snapshot.CreditBalance != 0 {
		t.Fatalf("snapshot = %+v", snapshot)
	}
}

func TestServiceRedeemsCreditCodeOnce(t *testing.T) {
	repository := newMemoryRepository()
	repository.codes["WELCOME100"] = &RedemptionCode{
		Code:         "WELCOME100",
		CreditAmount: 100,
		Active:       true,
	}
	service := NewService(repository)

	first, err := service.Redeem(context.Background(), RedeemInput{UserID: 42, Code: " welcome100 "})
	if err != nil {
		t.Fatalf("first Redeem() error = %v", err)
	}
	second, err := service.Redeem(context.Background(), RedeemInput{UserID: 42, Code: "WELCOME100"})
	if err != nil {
		t.Fatalf("second Redeem() error = %v", err)
	}

	if first.Snapshot.CreditBalance != 100 || second.Snapshot.CreditBalance != 100 {
		t.Fatalf("balances = %d/%d, want 100/100", first.Snapshot.CreditBalance, second.Snapshot.CreditBalance)
	}
	if !second.AlreadyRedeemed {
		t.Fatal("second redeem should be idempotent")
	}
	if len(repository.transactions) != 1 {
		t.Fatalf("transactions = %d, want 1", len(repository.transactions))
	}
}

func TestServiceRedeemsPlanCode(t *testing.T) {
	repository := newMemoryRepository()
	repository.codes["PRO30"] = &RedemptionCode{
		Code:     "PRO30",
		PlanCode: PlanPro,
		Active:   true,
	}
	service := NewService(repository)

	result, err := service.Redeem(context.Background(), RedeemInput{UserID: 42, Code: "PRO30"})
	if err != nil {
		t.Fatalf("Redeem() error = %v", err)
	}
	if result.Snapshot.Plan.Code != PlanPro {
		t.Fatalf("plan = %q, want %q", result.Snapshot.Plan.Code, PlanPro)
	}
}

func TestServiceRejectsInvalidCode(t *testing.T) {
	service := NewService(newMemoryRepository())

	_, err := service.Redeem(context.Background(), RedeemInput{UserID: 42, Code: "NOPE"})
	if !errors.Is(err, ErrCodeNotFound) {
		t.Fatalf("Redeem() error = %v, want ErrCodeNotFound", err)
	}
}

func TestServiceListsPlans(t *testing.T) {
	repository := newMemoryRepository()
	repository.plans = []PlanOption{{
		Code:         PlanPro,
		Name:         "会员版",
		PriceCents:   6900,
		BillingCycle: "month",
		Recommended:  true,
	}}
	service := NewService(repository)

	plans, err := service.ListPlans(context.Background())

	if err != nil {
		t.Fatalf("ListPlans() error = %v", err)
	}
	if len(plans) != 1 || plans[0].Code != PlanPro || !plans[0].Recommended {
		t.Fatalf("plans = %+v", plans)
	}
}

func TestServiceListsUsageAndOrdersWithCappedLimit(t *testing.T) {
	repository := newMemoryRepository()
	repository.usage = []UsageItem{{Key: "lead_tasks", Used: 8, Limit: 30, Unit: "次/月"}}
	repository.orders = []Order{{ID: 1}, {ID: 2}}
	service := NewService(repository)

	usage, err := service.CurrentUsage(context.Background(), 42)
	if err != nil {
		t.Fatalf("CurrentUsage() error = %v", err)
	}
	orders, err := service.ListOrders(context.Background(), 42, 500)
	if err != nil {
		t.Fatalf("ListOrders() error = %v", err)
	}

	if len(usage) != 1 || len(orders) != 2 {
		t.Fatalf("usage/orders = %+v/%+v", usage, orders)
	}
}

func TestServiceReturnsDefaultFeatureAccess(t *testing.T) {
	service := NewService(newMemoryRepository())

	response, err := service.FeatureAccess(context.Background(), 42, nil)

	if err != nil {
		t.Fatalf("FeatureAccess() error = %v", err)
	}
	if len(response.Features) != 4 {
		t.Fatalf("features = %+v", response.Features)
	}
	for _, feature := range response.Features {
		if feature.Status != "locked" || feature.AllowWorkflow {
			t.Fatalf("feature should be locked without workflow: %+v", feature)
		}
	}
}

func TestServiceFiltersFeatureAccessByKey(t *testing.T) {
	service := NewService(newMemoryRepository())

	response, err := service.FeatureAccess(context.Background(), 42, []string{FeatureCRM})

	if err != nil {
		t.Fatalf("FeatureAccess() error = %v", err)
	}
	if len(response.Features) != 1 || response.Features[0].Key != FeatureCRM {
		t.Fatalf("features = %+v", response.Features)
	}
}

func TestServiceCreatesManualCheckoutOrder(t *testing.T) {
	repository := newMemoryRepository()
	repository.plans = []PlanOption{{Code: PlanPro, BillingCycle: "month", PriceCents: 6900}}
	service := NewService(repository)
	service.now = func() time.Time { return time.Date(2026, 7, 2, 10, 0, 0, 0, time.UTC) }

	result, err := service.CreateCheckout(context.Background(), CheckoutInput{
		UserID:       42,
		PlanCode:     " pro ",
		BillingCycle: " month ",
	})

	if err != nil {
		t.Fatalf("CreateCheckout() error = %v", err)
	}
	if result.Order.Status != OrderPending || result.Payment.Mode != "manual" {
		t.Fatalf("result = %+v", result)
	}
	if repository.checkout.PlanCode != PlanPro || repository.checkout.BillingCycle != "month" || repository.checkout.OrderNo == "" {
		t.Fatalf("checkout = %+v", repository.checkout)
	}
}

func TestServiceCheckAndConsumeNormalizesInput(t *testing.T) {
	repository := newMemoryRepository()
	repository.consumeResult = UsageItem{Key: FeatureSandboxRuns, Used: 1, Limit: 20, Unit: "次/月"}
	service := NewService(repository)
	service.now = func() time.Time { return time.Date(2026, 7, 6, 10, 0, 0, 0, time.UTC) }

	usage, err := service.CheckAndConsume(context.Background(), ConsumeInput{
		UserID:         42,
		FeatureKey:     " sandbox_runs ",
		Amount:         0,
		IdempotencyKey: " sandbox-run-99 ",
	})

	if err != nil {
		t.Fatalf("CheckAndConsume() error = %v", err)
	}
	if usage.Used != 1 || len(repository.consumed) != 1 {
		t.Fatalf("usage/consumed = %+v/%+v", usage, repository.consumed)
	}
	consumed := repository.consumed[0]
	if consumed.FeatureKey != FeatureSandboxRuns || consumed.Amount != 1 || consumed.IdempotencyKey != "sandbox-run-99" {
		t.Fatalf("consumed = %+v", consumed)
	}
}

func TestServiceCheckAndConsumeReturnsQuotaExceeded(t *testing.T) {
	repository := newMemoryRepository()
	repository.consumeErr = ErrQuotaExceeded
	service := NewService(repository)

	_, err := service.CheckAndConsume(context.Background(), ConsumeInput{
		UserID:         42,
		FeatureKey:     FeatureCompetitorScans,
		Amount:         1,
		IdempotencyKey: "competitor-scan-1",
	})

	if !errors.Is(err, ErrQuotaExceeded) {
		t.Fatalf("err = %v, want ErrQuotaExceeded", err)
	}
}

func TestServiceRefundUsageNormalizesInput(t *testing.T) {
	repository := newMemoryRepository()
	repository.refundResult = UsageItem{Key: FeatureLeadTasks, Used: 0, Limit: 30, Unit: "次/月"}
	service := NewService(repository)

	usage, err := service.RefundUsage(context.Background(), ConsumeInput{
		UserID:         42,
		FeatureKey:     " lead_tasks ",
		Amount:         0,
		IdempotencyKey: " lead-task-42-refund ",
	})

	if err != nil {
		t.Fatalf("RefundUsage() error = %v", err)
	}
	if usage.Used != 0 || len(repository.refunded) != 1 {
		t.Fatalf("usage/refunded = %+v/%+v", usage, repository.refunded)
	}
	refunded := repository.refunded[0]
	if refunded.FeatureKey != FeatureLeadTasks || refunded.Amount != 1 || refunded.IdempotencyKey != "lead-task-42-refund" {
		t.Fatalf("refunded = %+v", refunded)
	}
}
