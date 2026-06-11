package membership

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type postgresDB interface {
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
