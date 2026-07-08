package membership

import (
	"context"
	"regexp"
	"testing"
	"time"

	pgxmock "github.com/pashagolub/pgxmock/v4"
)

func TestPostgresRepositoryReturnsFreeSnapshotWithCreditBalance(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	now := time.Now()
	db.ExpectExec(regexp.QuoteMeta(`
		INSERT INTO credit_accounts (user_id)
		VALUES ($1)
		ON CONFLICT (user_id) DO NOTHING
	`)).
		WithArgs(int64(42)).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))
	db.ExpectQuery(regexp.QuoteMeta(`
		SELECT balance
		FROM credit_accounts
		WHERE user_id = $1
	`)).
		WithArgs(int64(42)).
		WillReturnRows(pgxmock.NewRows([]string{"balance"}).AddRow(25))
	db.ExpectQuery(regexp.QuoteMeta(`
		SELECT plan_code, status, starts_at, ends_at
		FROM user_subscriptions
		WHERE user_id = $1
		  AND status = 'active'
		  AND (ends_at IS NULL OR ends_at > $2)
		ORDER BY starts_at DESC
		LIMIT 1
	`)).
		WithArgs(int64(42), now).
		WillReturnRows(pgxmock.NewRows([]string{"plan_code", "status", "starts_at", "ends_at"}))

	repository := NewPostgresRepository(db)
	snapshot, err := repository.CurrentSnapshot(context.Background(), 42, now)
	if err != nil {
		t.Fatalf("CurrentSnapshot() error = %v", err)
	}
	if snapshot.Plan.Code != PlanFree || snapshot.CreditBalance != 25 {
		t.Fatalf("snapshot = %+v", snapshot)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresRepositoryListsPlansWithQuotas(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	db.ExpectQuery(regexp.QuoteMeta(`
		SELECT ROW_NUMBER() OVER (ORDER BY display_order, code)::BIGINT AS id,
		       code, name, price_cents, billing_cycle, features, recommended
		FROM membership_plans
		WHERE active = TRUE
		ORDER BY display_order, code
	`)).
		WillReturnRows(pgxmock.NewRows([]string{"id", "code", "name", "price_cents", "billing_cycle", "features", "recommended"}).
			AddRow(int64(1), PlanPro, "会员版", 6900, "month", []byte(`["线索数据实时更新"]`), true))
	db.ExpectQuery(regexp.QuoteMeta(`
		SELECT plan_code, key, label, limit_value, unit
		FROM membership_plan_quotas
		ORDER BY plan_code, display_order, key
	`)).
		WillReturnRows(pgxmock.NewRows([]string{"plan_code", "key", "label", "limit_value", "unit"}).
			AddRow(PlanPro, "lead_tasks", "AI线索任务", 30, "次/月"))

	repository := NewPostgresRepository(db)
	plans, err := repository.ListPlans(context.Background())
	if err != nil {
		t.Fatalf("ListPlans() error = %v", err)
	}
	if len(plans) != 1 || len(plans[0].Quotas) != 1 || plans[0].Features[0] != "线索数据实时更新" {
		t.Fatalf("plans = %+v", plans)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresRepositoryCreatesCheckoutOrder(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 7, 2, 10, 0, 0, 0, time.UTC)
	db.ExpectQuery(regexp.QuoteMeta(`
		SELECT code, name, price_cents, billing_cycle
		FROM membership_plans
		WHERE code = $1 AND active = TRUE
	`)).
		WithArgs(PlanPro).
		WillReturnRows(pgxmock.NewRows([]string{"code", "name", "price_cents", "billing_cycle"}).
			AddRow(PlanPro, "会员版", 6900, "month"))
	db.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO membership_orders (order_no, user_id, plan_code, amount_cents, status, created_at)
		VALUES ($1, $2, $3, $4, 'pending', $5)
		RETURNING id, order_no, plan_code, amount_cents, status, created_at, paid_at
	`)).
		WithArgs("ZS-20260702-000042", int64(42), PlanPro, 6900, now).
		WillReturnRows(pgxmock.NewRows([]string{"id", "order_no", "plan_code", "amount_cents", "status", "created_at", "paid_at"}).
			AddRow(int64(13), "ZS-20260702-000042", PlanPro, 6900, OrderPending, now, nil))

	repository := NewPostgresRepository(db)
	result, err := repository.CreateCheckout(context.Background(), CheckoutInput{
		UserID:       42,
		PlanCode:     PlanPro,
		BillingCycle: "month",
		OrderNo:      "ZS-20260702-000042",
	}, now)
	if err != nil {
		t.Fatalf("CreateCheckout() error = %v", err)
	}
	if result.Order.ID != 13 || result.Payment.Mode != "manual" {
		t.Fatalf("result = %+v", result)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresRepositoryCheckAndConsumeIncrementsUsageOnce(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 7, 6, 10, 0, 0, 0, time.UTC)
	resetAt := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	input := ConsumeInput{
		UserID:         42,
		FeatureKey:     FeatureSandboxRuns,
		Amount:         1,
		IdempotencyKey: "sandbox-run-99",
	}

	db.ExpectBegin()
	db.ExpectQuery(regexp.QuoteMeta(`
		SELECT plan_code, status, starts_at, ends_at
		FROM user_subscriptions
		WHERE user_id = $1
		  AND status = 'active'
		  AND (ends_at IS NULL OR ends_at > $2)
		ORDER BY starts_at DESC
		LIMIT 1
	`)).
		WithArgs(input.UserID, now).
		WillReturnRows(pgxmock.NewRows([]string{"plan_code", "status", "starts_at", "ends_at"}))
	db.ExpectQuery(regexp.QuoteMeta(`
		SELECT label, limit_value, unit
		FROM membership_plan_quotas
		WHERE plan_code = $1 AND key = $2
	`)).
		WithArgs(PlanFree, FeatureSandboxRuns).
		WillReturnRows(pgxmock.NewRows([]string{"label", "limit_value", "unit"}).AddRow("商业沙盘", 1, "次/月"))
	db.ExpectExec(regexp.QuoteMeta(`
		INSERT INTO membership_usage (user_id, key, label, used, limit_value, unit, reset_at, updated_at)
		VALUES ($1, $2, $3, 0, $4, $5, $6, $7)
		ON CONFLICT (user_id, key) DO NOTHING
	`)).
		WithArgs(input.UserID, input.FeatureKey, "商业沙盘", 1, "次/月", resetAt, now).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))
	db.ExpectQuery(regexp.QuoteMeta(`
		SELECT label, used, limit_value, unit, reset_at
		FROM membership_usage
		WHERE user_id = $1 AND key = $2
		FOR UPDATE
	`)).
		WithArgs(input.UserID, input.FeatureKey).
		WillReturnRows(pgxmock.NewRows([]string{"label", "used", "limit_value", "unit", "reset_at"}).
			AddRow("商业沙盘", 0, 1, "次/月", resetAt))
	db.ExpectQuery(regexp.QuoteMeta(`
		SELECT amount, used_after, reset_at
		FROM membership_usage_events
		WHERE user_id = $1 AND key = $2 AND idempotency_key = $3
	`)).
		WithArgs(input.UserID, input.FeatureKey, input.IdempotencyKey).
		WillReturnRows(pgxmock.NewRows([]string{"amount", "used_after", "reset_at"}))
	db.ExpectQuery(regexp.QuoteMeta(`
		UPDATE membership_usage
		SET used = used + $3, label = $4, limit_value = $5, unit = $6, reset_at = $7, updated_at = $8
		WHERE user_id = $1 AND key = $2
		RETURNING used
	`)).
		WithArgs(input.UserID, input.FeatureKey, 1, "商业沙盘", 1, "次/月", resetAt, now).
		WillReturnRows(pgxmock.NewRows([]string{"used"}).AddRow(1))
	db.ExpectExec(regexp.QuoteMeta(`
		INSERT INTO membership_usage_events (user_id, key, idempotency_key, amount, used_after, reset_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`)).
		WithArgs(input.UserID, input.FeatureKey, input.IdempotencyKey, 1, 1, resetAt, now).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))
	db.ExpectCommit()

	repository := NewPostgresRepository(db)
	usage, err := repository.CheckAndConsume(context.Background(), input, now)
	if err != nil {
		t.Fatalf("CheckAndConsume() error = %v", err)
	}
	if usage.Used != 1 || usage.Limit != 1 || usage.ResetAt == nil || !usage.ResetAt.Equal(resetAt) {
		t.Fatalf("usage = %+v", usage)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresRepositoryRefundUsageDecrementsUsageOnce(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 7, 6, 10, 0, 0, 0, time.UTC)
	resetAt := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	input := ConsumeInput{
		UserID:         42,
		FeatureKey:     FeatureLeadTasks,
		Amount:         1,
		IdempotencyKey: "lead-task-42-refund-provider",
	}

	db.ExpectBegin()
	db.ExpectQuery(regexp.QuoteMeta(`
		SELECT plan_code, status, starts_at, ends_at
		FROM user_subscriptions
		WHERE user_id = $1
		  AND status = 'active'
		  AND (ends_at IS NULL OR ends_at > $2)
		ORDER BY starts_at DESC
		LIMIT 1
	`)).
		WithArgs(input.UserID, now).
		WillReturnRows(pgxmock.NewRows([]string{"plan_code", "status", "starts_at", "ends_at"}).AddRow(PlanPro, "active", now.Add(-time.Hour), nil))
	db.ExpectQuery(regexp.QuoteMeta(`
		SELECT label, limit_value, unit
		FROM membership_plan_quotas
		WHERE plan_code = $1 AND key = $2
	`)).
		WithArgs(PlanPro, FeatureLeadTasks).
		WillReturnRows(pgxmock.NewRows([]string{"label", "limit_value", "unit"}).AddRow("AI线索任务", 30, "次/月"))
	db.ExpectExec(regexp.QuoteMeta(`
		INSERT INTO membership_usage (user_id, key, label, used, limit_value, unit, reset_at, updated_at)
		VALUES ($1, $2, $3, 0, $4, $5, $6, $7)
		ON CONFLICT (user_id, key) DO NOTHING
	`)).
		WithArgs(input.UserID, input.FeatureKey, "AI线索任务", 30, "次/月", resetAt, now).
		WillReturnResult(pgxmock.NewResult("INSERT", 0))
	db.ExpectQuery(regexp.QuoteMeta(`
		SELECT label, used, limit_value, unit, reset_at
		FROM membership_usage
		WHERE user_id = $1 AND key = $2
		FOR UPDATE
	`)).
		WithArgs(input.UserID, input.FeatureKey).
		WillReturnRows(pgxmock.NewRows([]string{"label", "used", "limit_value", "unit", "reset_at"}).
			AddRow("AI线索任务", 1, 30, "次/月", resetAt))
	db.ExpectQuery(regexp.QuoteMeta(`
		SELECT amount, used_after, reset_at
		FROM membership_usage_events
		WHERE user_id = $1 AND key = $2 AND idempotency_key = $3
	`)).
		WithArgs(input.UserID, input.FeatureKey, input.IdempotencyKey).
		WillReturnRows(pgxmock.NewRows([]string{"amount", "used_after", "reset_at"}))
	db.ExpectQuery(regexp.QuoteMeta(`
		UPDATE membership_usage
		SET used = $3, label = $4, limit_value = $5, unit = $6, reset_at = $7, updated_at = $8
		WHERE user_id = $1 AND key = $2
		RETURNING used
	`)).
		WithArgs(input.UserID, input.FeatureKey, 0, "AI线索任务", 30, "次/月", resetAt, now).
		WillReturnRows(pgxmock.NewRows([]string{"used"}).AddRow(0))
	db.ExpectExec(regexp.QuoteMeta(`
		INSERT INTO membership_usage_events (user_id, key, idempotency_key, amount, used_after, reset_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`)).
		WithArgs(input.UserID, input.FeatureKey, input.IdempotencyKey, -1, 0, resetAt, now).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))
	db.ExpectCommit()

	repository := NewPostgresRepository(db)
	usage, err := repository.RefundUsage(context.Background(), input, now)
	if err != nil {
		t.Fatalf("RefundUsage() error = %v", err)
	}
	if usage.Used != 0 || usage.Limit != 30 || usage.ResetAt == nil || !usage.ResetAt.Equal(resetAt) {
		t.Fatalf("usage = %+v", usage)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresRepositoryRedeemsCreditCodeInTransaction(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	now := time.Now()
	db.ExpectBegin()
	db.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, code, COALESCE(plan_code, ''), credit_amount, max_redemptions, redeemed_count, expires_at, active
		FROM redemption_codes
		WHERE code = $1
		FOR UPDATE
	`)).
		WithArgs("WELCOME100").
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "code", "plan_code", "credit_amount", "max_redemptions", "redeemed_count", "expires_at", "active",
		}).AddRow(int64(9), "WELCOME100", "", 100, 0, 0, nil, true))
	db.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, created_at
		FROM redemptions
		WHERE user_id = $1 AND redemption_code_id = $2
	`)).
		WithArgs(int64(42), int64(9)).
		WillReturnRows(pgxmock.NewRows([]string{"id", "created_at"}))
	db.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO redemptions (user_id, redemption_code_id, code)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`)).
		WithArgs(int64(42), int64(9), "WELCOME100").
		WillReturnRows(pgxmock.NewRows([]string{"id", "created_at"}).AddRow(int64(11), now))
	db.ExpectExec(regexp.QuoteMeta(`
		INSERT INTO credit_accounts (user_id)
		VALUES ($1)
		ON CONFLICT (user_id) DO NOTHING
	`)).
		WithArgs(int64(42)).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))
	db.ExpectExec(regexp.QuoteMeta(`
		INSERT INTO credit_transactions (user_id, amount, reason, reference_type, reference_id)
		VALUES ($1, $2, 'redemption', 'redemption', $3)
	`)).
		WithArgs(int64(42), 100, int64(11)).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))
	db.ExpectExec(regexp.QuoteMeta(`
		UPDATE credit_accounts
		SET balance = balance + $2, updated_at = NOW()
		WHERE user_id = $1
	`)).
		WithArgs(int64(42), 100).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	db.ExpectExec(regexp.QuoteMeta(`
		UPDATE redemption_codes
		SET redeemed_count = redeemed_count + 1, updated_at = NOW()
		WHERE id = $1
	`)).
		WithArgs(int64(9)).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	db.ExpectCommit()
	db.ExpectExec(regexp.QuoteMeta(`
		INSERT INTO credit_accounts (user_id)
		VALUES ($1)
		ON CONFLICT (user_id) DO NOTHING
	`)).
		WithArgs(int64(42)).
		WillReturnResult(pgxmock.NewResult("INSERT", 0))
	db.ExpectQuery(regexp.QuoteMeta(`
		SELECT balance
		FROM credit_accounts
		WHERE user_id = $1
	`)).
		WithArgs(int64(42)).
		WillReturnRows(pgxmock.NewRows([]string{"balance"}).AddRow(100))
	db.ExpectQuery(regexp.QuoteMeta(`
		SELECT plan_code, status, starts_at, ends_at
		FROM user_subscriptions
		WHERE user_id = $1
		  AND status = 'active'
		  AND (ends_at IS NULL OR ends_at > $2)
		ORDER BY starts_at DESC
		LIMIT 1
	`)).
		WithArgs(int64(42), now).
		WillReturnRows(pgxmock.NewRows([]string{"plan_code", "status", "starts_at", "ends_at"}))

	repository := NewPostgresRepository(db)
	result, err := repository.RedeemCode(
		context.Background(),
		RedeemInput{UserID: 42, Code: "WELCOME100"},
		now,
	)
	if err != nil {
		t.Fatalf("RedeemCode() error = %v", err)
	}
	if result.Snapshot.CreditBalance != 100 || result.AlreadyRedeemed {
		t.Fatalf("result = %+v", result)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
