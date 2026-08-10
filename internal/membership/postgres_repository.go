package membership

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type postgresDB interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Begin(ctx context.Context) (pgx.Tx, error)
}

type PostgresRepository struct {
	db postgresDB
}

func NewPostgresRepository(db postgresDB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) CurrentSnapshot(ctx context.Context, userID int64, now time.Time) (Snapshot, error) {
	if err := r.ensureCreditAccount(ctx, r.db, userID); err != nil {
		return Snapshot{}, err
	}
	balance, err := r.creditBalance(ctx, r.db, userID)
	if err != nil {
		return Snapshot{}, err
	}
	subscription, err := r.activeSubscription(ctx, r.db, userID, now)
	if err != nil {
		return Snapshot{}, err
	}
	planCode := PlanFree
	if subscription.PlanCode != "" {
		planCode = subscription.PlanCode
	}
	plan, ok := PlanCatalog[planCode]
	if !ok {
		plan = PlanCatalog[PlanFree]
	}
	return Snapshot{
		Plan:          plan,
		Subscription:  subscription,
		CreditBalance: balance,
	}, nil
}

func (r *PostgresRepository) ListPlans(ctx context.Context) ([]PlanOption, error) {
	rows, err := r.db.Query(ctx, `
		SELECT ROW_NUMBER() OVER (ORDER BY display_order, code)::BIGINT AS id,
		       code, name, price_cents, billing_cycle, features, recommended
		FROM membership_plans
		WHERE active = TRUE
		ORDER BY display_order, code
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var plans []PlanOption
	byCode := map[string]int{}
	for rows.Next() {
		var plan PlanOption
		var features []byte
		if err := rows.Scan(&plan.ID, &plan.Code, &plan.Name, &plan.PriceCents, &plan.BillingCycle, &features, &plan.Recommended); err != nil {
			return nil, err
		}
		if len(features) > 0 {
			if err := json.Unmarshal(features, &plan.Features); err != nil {
				return nil, err
			}
		}
		byCode[plan.Code] = len(plans)
		plans = append(plans, plan)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	quotaRows, err := r.db.Query(ctx, `
		SELECT plan_code, key, label, limit_value, unit
		FROM membership_plan_quotas
		ORDER BY plan_code, display_order, key
	`)
	if err != nil {
		return nil, err
	}
	defer quotaRows.Close()
	for quotaRows.Next() {
		var planCode string
		var quota PlanQuota
		if err := quotaRows.Scan(&planCode, &quota.Key, &quota.Label, &quota.Limit, &quota.Unit); err != nil {
			return nil, err
		}
		if index, ok := byCode[planCode]; ok {
			plans[index].Quotas = append(plans[index].Quotas, quota)
		}
	}
	if err := quotaRows.Err(); err != nil {
		return nil, err
	}
	if plans == nil {
		return []PlanOption{}, nil
	}
	return plans, nil
}

func (r *PostgresRepository) CurrentUsage(ctx context.Context, userID int64, now time.Time) ([]UsageItem, error) {
	planCode := PlanFree
	subscription, err := r.activeSubscription(ctx, r.db, userID, now)
	if err != nil {
		return nil, err
	}
	if subscription.PlanCode != "" {
		planCode = subscription.PlanCode
	}
	resetAt := nextMonthlyReset(now)
	rows, err := r.db.Query(ctx, `
		SELECT q.key,
		       q.label,
		       CASE
		           WHEN u.reset_at IS NULL OR u.reset_at <= $3 THEN 0
		           ELSE COALESCE(u.used, 0)
		       END AS used,
		       q.limit_value,
		       q.unit,
		       CASE
		           WHEN u.reset_at IS NULL OR u.reset_at <= $3 THEN $4::TIMESTAMPTZ
		           ELSE u.reset_at
		       END AS reset_at
		FROM membership_plan_quotas q
		LEFT JOIN membership_usage u
		  ON u.user_id = $1 AND u.key = q.key
		WHERE q.plan_code = $2
		ORDER BY q.display_order, q.key
	`, userID, planCode, now, resetAt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var usage []UsageItem
	for rows.Next() {
		var item UsageItem
		var resetAt sql.NullTime
		if err := rows.Scan(&item.Key, &item.Label, &item.Used, &item.Limit, &item.Unit, &resetAt); err != nil {
			return nil, err
		}
		item.ResetAt = nullableTime(resetAt)
		usage = append(usage, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if usage == nil {
		return []UsageItem{}, nil
	}
	return usage, nil
}

func (r *PostgresRepository) CheckAndConsume(ctx context.Context, input ConsumeInput, now time.Time) (UsageItem, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return UsageItem{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	planCode := PlanFree
	subscription, err := r.activeSubscription(ctx, tx, input.UserID, now)
	if err != nil {
		return UsageItem{}, err
	}
	if subscription.PlanCode != "" {
		planCode = subscription.PlanCode
	}

	quota, err := r.planQuota(ctx, tx, planCode, input.FeatureKey)
	if err != nil {
		return UsageItem{}, err
	}
	resetAt := nextMonthlyReset(now)
	if _, err := tx.Exec(ctx, `
		INSERT INTO membership_usage (user_id, key, label, used, limit_value, unit, reset_at, updated_at)
		VALUES ($1, $2, $3, 0, $4, $5, $6, $7)
		ON CONFLICT (user_id, key) DO NOTHING
	`, input.UserID, input.FeatureKey, quota.Label, quota.Limit, quota.Unit, resetAt, now); err != nil {
		return UsageItem{}, err
	}

	usage, err := r.lockUsage(ctx, tx, input.UserID, input.FeatureKey)
	if err != nil {
		return UsageItem{}, err
	}
	if usage.ResetAt == nil || !usage.ResetAt.After(now) {
		usage.Used = 0
		usage.ResetAt = &resetAt
	}
	usage.Key = input.FeatureKey
	usage.Label = quota.Label
	usage.Limit = quota.Limit
	usage.Unit = quota.Unit

	if existing, ok, err := r.existingUsageEvent(ctx, tx, input); err != nil {
		return UsageItem{}, err
	} else if ok {
		usage.Used = existing.Used
		usage.ResetAt = existing.ResetAt
		if err := tx.Commit(ctx); err != nil {
			return UsageItem{}, err
		}
		return usage, nil
	}

	if quota.Limit >= 0 && usage.Used+input.Amount > quota.Limit {
		return UsageItem{}, ErrQuotaExceeded
	}
	usedAfter := usage.Used + input.Amount
	resetAtValue := resetAt
	if usage.ResetAt != nil {
		resetAtValue = *usage.ResetAt
	}
	if err := tx.QueryRow(ctx, `
		UPDATE membership_usage
		SET used = used + $3, label = $4, limit_value = $5, unit = $6, reset_at = $7, updated_at = $8
		WHERE user_id = $1 AND key = $2
		RETURNING used
	`, input.UserID, input.FeatureKey, input.Amount, quota.Label, quota.Limit, quota.Unit, resetAtValue, now).Scan(&usage.Used); err != nil {
		return UsageItem{}, err
	}
	if usage.Used != usedAfter {
		usedAfter = usage.Used
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO membership_usage_events (user_id, key, idempotency_key, amount, used_after, reset_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, input.UserID, input.FeatureKey, input.IdempotencyKey, input.Amount, usedAfter, resetAtValue, now); err != nil {
		return UsageItem{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return UsageItem{}, err
	}
	return usage, nil
}

func (r *PostgresRepository) RefundUsage(ctx context.Context, input ConsumeInput, now time.Time) (UsageItem, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return UsageItem{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	planCode := PlanFree
	subscription, err := r.activeSubscription(ctx, tx, input.UserID, now)
	if err != nil {
		return UsageItem{}, err
	}
	if subscription.PlanCode != "" {
		planCode = subscription.PlanCode
	}

	quota, err := r.planQuota(ctx, tx, planCode, input.FeatureKey)
	if err != nil {
		return UsageItem{}, err
	}
	resetAt := nextMonthlyReset(now)
	if _, err := tx.Exec(ctx, `
		INSERT INTO membership_usage (user_id, key, label, used, limit_value, unit, reset_at, updated_at)
		VALUES ($1, $2, $3, 0, $4, $5, $6, $7)
		ON CONFLICT (user_id, key) DO NOTHING
	`, input.UserID, input.FeatureKey, quota.Label, quota.Limit, quota.Unit, resetAt, now); err != nil {
		return UsageItem{}, err
	}

	usage, err := r.lockUsage(ctx, tx, input.UserID, input.FeatureKey)
	if err != nil {
		return UsageItem{}, err
	}
	if usage.ResetAt == nil || !usage.ResetAt.After(now) {
		usage.Used = 0
		usage.ResetAt = &resetAt
	}
	usage.Key = input.FeatureKey
	usage.Label = quota.Label
	usage.Limit = quota.Limit
	usage.Unit = quota.Unit

	if existing, ok, err := r.existingUsageEvent(ctx, tx, input); err != nil {
		return UsageItem{}, err
	} else if ok {
		usage.Used = existing.Used
		usage.ResetAt = existing.ResetAt
		if err := tx.Commit(ctx); err != nil {
			return UsageItem{}, err
		}
		return usage, nil
	}

	refundAmount := input.Amount
	if refundAmount > usage.Used {
		refundAmount = usage.Used
	}
	if refundAmount == 0 {
		if err := tx.Commit(ctx); err != nil {
			return UsageItem{}, err
		}
		return usage, nil
	}
	usedAfter := usage.Used - refundAmount
	resetAtValue := resetAt
	if usage.ResetAt != nil {
		resetAtValue = *usage.ResetAt
	}
	if err := tx.QueryRow(ctx, `
		UPDATE membership_usage
		SET used = $3, label = $4, limit_value = $5, unit = $6, reset_at = $7, updated_at = $8
		WHERE user_id = $1 AND key = $2
		RETURNING used
	`, input.UserID, input.FeatureKey, usedAfter, quota.Label, quota.Limit, quota.Unit, resetAtValue, now).Scan(&usage.Used); err != nil {
		return UsageItem{}, err
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO membership_usage_events (user_id, key, idempotency_key, amount, used_after, reset_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, input.UserID, input.FeatureKey, input.IdempotencyKey, -refundAmount, usage.Used, resetAtValue, now); err != nil {
		return UsageItem{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return UsageItem{}, err
	}
	return usage, nil
}

func (r *PostgresRepository) ListOrders(ctx context.Context, userID int64, limit int) ([]Order, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, order_no, plan_code, amount_cents, status, created_at, paid_at
		FROM membership_orders
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var orders []Order
	for rows.Next() {
		var order Order
		if err := rows.Scan(&order.ID, &order.OrderNo, &order.PlanCode, &order.AmountCents, &order.Status, &order.CreatedAt, &order.PaidAt); err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if orders == nil {
		return []Order{}, nil
	}
	return orders, nil
}

func (r *PostgresRepository) CreateCheckout(ctx context.Context, input CheckoutInput, now time.Time) (CheckoutResult, error) {
	var plan struct {
		Code         string
		Name         string
		PriceCents   int
		BillingCycle string
	}
	err := r.db.QueryRow(ctx, `
		SELECT code, name, price_cents, billing_cycle
		FROM membership_plans
		WHERE code = $1 AND active = TRUE
	`, input.PlanCode).Scan(&plan.Code, &plan.Name, &plan.PriceCents, &plan.BillingCycle)
	if errors.Is(err, pgx.ErrNoRows) {
		return CheckoutResult{}, ErrPlanNotFound
	}
	if err != nil {
		return CheckoutResult{}, err
	}
	if plan.BillingCycle != input.BillingCycle {
		return CheckoutResult{}, ErrPlanNotFound
	}

	var order Order
	err = r.db.QueryRow(ctx, `
		INSERT INTO membership_orders (order_no, user_id, plan_code, amount_cents, status, created_at)
		VALUES ($1, $2, $3, $4, 'pending', $5)
		RETURNING id, order_no, plan_code, amount_cents, status, created_at, paid_at
	`, input.OrderNo, input.UserID, plan.Code, plan.PriceCents, now).Scan(
		&order.ID,
		&order.OrderNo,
		&order.PlanCode,
		&order.AmountCents,
		&order.Status,
		&order.CreatedAt,
		&order.PaidAt,
	)
	if err != nil {
		return CheckoutResult{}, err
	}
	return CheckoutResult{
		Order:   order,
		Payment: PaymentInfo{Mode: "manual", Message: "请联系顾问完成开通"},
	}, nil
}

func (r *PostgresRepository) RedeemCode(ctx context.Context, input RedeemInput, now time.Time) (RedeemResult, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return RedeemResult{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	code, err := r.lockRedemptionCode(ctx, tx, input.Code)

	if err != nil {
		return RedeemResult{}, err
	}
	if !code.Active {
		return RedeemResult{}, ErrCodeNotFound
	}
	if code.ExpiresAt != nil && !code.ExpiresAt.After(now) {
		return RedeemResult{}, ErrCodeExpired
	}

	redemption, exists, err := r.existingRedemption(ctx, tx, input.UserID, code.ID)
	if err != nil {
		return RedeemResult{}, err
	}
	if exists {
		if err := tx.Commit(ctx); err != nil {
			return RedeemResult{}, err
		}
		snapshot, err := r.CurrentSnapshot(ctx, input.UserID, now)
		if err != nil {
			return RedeemResult{}, err
		}
		redemption.Code = input.Code
		return RedeemResult{Snapshot: snapshot, Redemption: redemption, AlreadyRedeemed: true}, nil
	}
	if code.MaxRedemptions > 0 && code.RedeemedCount >= code.MaxRedemptions {
		return RedeemResult{}, ErrCodeExhausted
	}

	redemption, err = r.insertRedemption(ctx, tx, input.UserID, code.ID, input.Code)
	if err != nil {
		return RedeemResult{}, err
	}
	if code.CreditAmount > 0 {
		if err := r.ensureCreditAccount(ctx, tx, input.UserID); err != nil {
			return RedeemResult{}, err
		}
		if err := r.recordCredit(ctx, tx, input.UserID, code.CreditAmount, redemption.ID); err != nil {
			return RedeemResult{}, err
		}
	}
	if code.PlanCode != "" {
		if err := r.upsertSubscription(ctx, tx, input.UserID, code.PlanCode); err != nil {
			return RedeemResult{}, err
		}
	}
	if _, err := tx.Exec(ctx, `
		UPDATE redemption_codes
		SET redeemed_count = redeemed_count + 1, updated_at = NOW()
		WHERE id = $1
	`, code.ID); err != nil {
		return RedeemResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return RedeemResult{}, err
	}

	snapshot, err := r.CurrentSnapshot(ctx, input.UserID, now)
	if err != nil {
		return RedeemResult{}, err
	}
	redemption.Code = input.Code
	return RedeemResult{Snapshot: snapshot, Redemption: redemption}, nil
}

func (r *PostgresRepository) ensureCreditAccount(ctx context.Context, runner interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
}, userID int64) error {
	_, err := runner.Exec(ctx, `
		INSERT INTO credit_accounts (user_id)
		VALUES ($1)
		ON CONFLICT (user_id) DO NOTHING
	`, userID)
	return err
}

func (r *PostgresRepository) creditBalance(ctx context.Context, runner interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}, userID int64) (int, error) {
	var balance int
	err := runner.QueryRow(ctx, `
		SELECT balance
		FROM credit_accounts
		WHERE user_id = $1
	`, userID).Scan(&balance)
	return balance, err
}

func (r *PostgresRepository) activeSubscription(ctx context.Context, runner interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}, userID int64, now time.Time) (Subscription, error) {
	var subscription Subscription
	err := runner.QueryRow(ctx, `
		SELECT plan_code, status, starts_at, ends_at
		FROM user_subscriptions
		WHERE user_id = $1
		  AND status = 'active'
		  AND (ends_at IS NULL OR ends_at > $2)
		ORDER BY starts_at DESC
		LIMIT 1
	`, userID, now).Scan(
		&subscription.PlanCode,
		&subscription.Status,
		&subscription.StartsAt,
		&subscription.EndsAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Subscription{}, nil
	}
	if err != nil {
		return Subscription{}, err
	}
	subscription.UserID = userID
	return subscription, nil
}

type lockedCode struct {
	ID             int64
	Code           string
	PlanCode       string
	CreditAmount   int
	MaxRedemptions int
	RedeemedCount  int
	ExpiresAt      *time.Time
	Active         bool
}

func (r *PostgresRepository) lockRedemptionCode(ctx context.Context, tx pgx.Tx, code string) (lockedCode, error) {
	var result lockedCode
	err := tx.QueryRow(ctx, `
		SELECT id, code, COALESCE(plan_code, ''), credit_amount, max_redemptions, redeemed_count, expires_at, active
		FROM redemption_codes
		WHERE code = $1
		FOR UPDATE
	`, code).Scan(
		&result.ID,
		&result.Code,
		&result.PlanCode,
		&result.CreditAmount,
		&result.MaxRedemptions,
		&result.RedeemedCount,
		&result.ExpiresAt,
		&result.Active,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return lockedCode{}, ErrCodeNotFound
	}
	return result, err
}

func (r *PostgresRepository) planQuota(ctx context.Context, runner interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}, planCode, featureKey string) (PlanQuota, error) {
	var quota PlanQuota
	err := runner.QueryRow(ctx, `
		SELECT label, limit_value, unit
		FROM membership_plan_quotas
		WHERE plan_code = $1 AND key = $2
	`, planCode, featureKey).Scan(&quota.Label, &quota.Limit, &quota.Unit)
	if errors.Is(err, pgx.ErrNoRows) {
		return PlanQuota{}, ErrQuotaNotFound
	}
	if err != nil {
		return PlanQuota{}, err
	}
	quota.Key = featureKey
	return quota, nil
}

func (r *PostgresRepository) lockUsage(ctx context.Context, tx pgx.Tx, userID int64, featureKey string) (UsageItem, error) {
	var usage UsageItem
	var resetAt sql.NullTime
	err := tx.QueryRow(ctx, `
		SELECT label, used, limit_value, unit, reset_at
		FROM membership_usage
		WHERE user_id = $1 AND key = $2
		FOR UPDATE
	`, userID, featureKey).Scan(&usage.Label, &usage.Used, &usage.Limit, &usage.Unit, &resetAt)
	usage.Key = featureKey
	usage.ResetAt = nullableTime(resetAt)
	return usage, err
}

func (r *PostgresRepository) existingUsageEvent(ctx context.Context, tx pgx.Tx, input ConsumeInput) (UsageItem, bool, error) {
	var usage UsageItem
	var amount int
	var resetAt sql.NullTime
	err := tx.QueryRow(ctx, `
		SELECT amount, used_after, reset_at
		FROM membership_usage_events
		WHERE user_id = $1 AND key = $2 AND idempotency_key = $3
	`, input.UserID, input.FeatureKey, input.IdempotencyKey).Scan(&amount, &usage.Used, &resetAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return UsageItem{}, false, nil
	}
	if err != nil {
		return UsageItem{}, false, err
	}
	usage.Key = input.FeatureKey
	usage.ResetAt = nullableTime(resetAt)
	return usage, true, nil
}

func nextMonthlyReset(now time.Time) time.Time {
	return time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, now.Location())
}

func nullableTime(value sql.NullTime) *time.Time {
	if !value.Valid {
		return nil
	}
	return &value.Time
}

func (r *PostgresRepository) existingRedemption(ctx context.Context, tx pgx.Tx, userID, codeID int64) (Redemption, bool, error) {
	var redemption Redemption
	err := tx.QueryRow(ctx, `
		SELECT id, created_at
		FROM redemptions
		WHERE user_id = $1 AND redemption_code_id = $2
	`, userID, codeID).Scan(&redemption.ID, &redemption.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Redemption{}, false, nil
	}
	if err != nil {
		return Redemption{}, false, err
	}
	redemption.UserID = userID
	return redemption, true, nil
}

func (r *PostgresRepository) insertRedemption(ctx context.Context, tx pgx.Tx, userID, codeID int64, code string) (Redemption, error) {
	var redemption Redemption
	err := tx.QueryRow(ctx, `
		INSERT INTO redemptions (user_id, redemption_code_id, code)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`, userID, codeID, code).Scan(&redemption.ID, &redemption.CreatedAt)
	if err != nil {
		return Redemption{}, err
	}
	redemption.UserID = userID
	redemption.Code = code
	return redemption, nil
}

func (r *PostgresRepository) recordCredit(ctx context.Context, tx pgx.Tx, userID int64, amount int, redemptionID int64) error {
	if _, err := tx.Exec(ctx, `
		INSERT INTO credit_transactions (user_id, amount, reason, reference_type, reference_id)
		VALUES ($1, $2, 'redemption', 'redemption', $3)
	`, userID, amount, redemptionID); err != nil {
		return err
	}
	_, err := tx.Exec(ctx, `
		UPDATE credit_accounts
		SET balance = balance + $2, updated_at = NOW()
		WHERE user_id = $1
	`, userID, amount)
	return err
}

func (r *PostgresRepository) upsertSubscription(ctx context.Context, tx pgx.Tx, userID int64, planCode string) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO user_subscriptions (user_id, plan_code, status, starts_at)
		VALUES ($1, $2, 'active', NOW())
		ON CONFLICT (user_id) DO UPDATE SET
			plan_code = EXCLUDED.plan_code,
			status = 'active',
			starts_at = NOW(),
			ends_at = NULL,
			updated_at = NOW()
	`, userID, planCode)
	return err
}
