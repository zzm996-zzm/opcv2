package crm

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type postgresDB interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Begin(ctx context.Context) (pgx.Tx, error)
}

type PostgresRepository struct {
	db postgresDB
}

func NewPostgresRepository(db postgresDB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) ImportCustomer(ctx context.Context, customer Customer) (Customer, bool, error) {
	var inserted bool
	var nextFollowUp pgtype.Timestamptz
	err := r.db.QueryRow(ctx, `
		INSERT INTO crm_customers (user_id, import_key, name, phone, email, website, stage, source, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $9)
		ON CONFLICT (user_id, import_key) DO UPDATE SET import_key = EXCLUDED.import_key
		RETURNING id, user_id, import_key, name, phone, email, website, stage, source, next_follow_up_at, created_at, updated_at, xmax = 0 AS inserted
	`,
		customer.UserID,
		customer.ImportKey,
		customer.Name,
		customer.Phone,
		customer.Email,
		customer.Website,
		customer.Stage,
		customer.Source,
		customer.CreatedAt,
	).Scan(
		&customer.ID,
		&customer.UserID,
		&customer.ImportKey,
		&customer.Name,
		&customer.Phone,
		&customer.Email,
		&customer.Website,
		&customer.Stage,
		&customer.Source,
		&nextFollowUp,
		&customer.CreatedAt,
		&customer.UpdatedAt,
		&inserted,
	)
	if nextFollowUp.Valid {
		customer.NextFollowUpAt = nextFollowUp.Time
	}
	return customer, !inserted, err
}

func (r *PostgresRepository) GetCustomer(ctx context.Context, userID, customerID int64) (Customer, error) {
	customer, err := scanCustomer(r.db.QueryRow(ctx, `
		SELECT id, user_id, import_key, name, phone, email, website, stage, source, next_follow_up_at, created_at, updated_at
		FROM crm_customers
		WHERE user_id = $1 AND id = $2
	`, userID, customerID))
	if errors.Is(err, pgx.ErrNoRows) {
		return Customer{}, ErrCustomerNotFound
	}
	return customer, err
}

func (r *PostgresRepository) UpdateStage(ctx context.Context, userID, customerID int64, stage string, activity Activity) (Customer, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return Customer{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	customer, err := scanCustomer(tx.QueryRow(ctx, `
		UPDATE crm_customers
		SET stage = $3, updated_at = $4
		WHERE user_id = $1 AND id = $2
		RETURNING id, user_id, import_key, name, phone, email, website, stage, source, next_follow_up_at, created_at, updated_at
	`, userID, customerID, stage, activity.CreatedAt))
	if errors.Is(err, pgx.ErrNoRows) {
		return Customer{}, ErrCustomerNotFound
	}
	if err != nil {
		return Customer{}, err
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO crm_activities (user_id, customer_id, type, note, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`, activity.UserID, activity.CustomerID, activity.Type, activity.Note, activity.CreatedAt); err != nil {
		return Customer{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Customer{}, err
	}
	return customer, nil
}

func (r *PostgresRepository) RecordFollowUp(ctx context.Context, followUp FollowUp, activity Activity) (FollowUp, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return FollowUp{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	command, err := tx.Exec(ctx, `
		UPDATE crm_customers
		SET next_follow_up_at = $3, updated_at = $4
		WHERE user_id = $1 AND id = $2
	`, followUp.UserID, followUp.CustomerID, followUp.NextFollowUpAt, followUp.CreatedAt)
	if err != nil {
		return FollowUp{}, err
	}
	if command.RowsAffected() == 0 {
		return FollowUp{}, ErrCustomerNotFound
	}
	if err := tx.QueryRow(ctx, `
		INSERT INTO crm_followups (user_id, customer_id, note, next_follow_up_at, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`, followUp.UserID, followUp.CustomerID, followUp.Note, followUp.NextFollowUpAt, followUp.CreatedAt).Scan(&followUp.ID); err != nil {
		return FollowUp{}, err
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO crm_activities (user_id, customer_id, type, note, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`, activity.UserID, activity.CustomerID, activity.Type, activity.Note, activity.CreatedAt); err != nil {
		return FollowUp{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return FollowUp{}, err
	}
	return followUp, nil
}

func (r *PostgresRepository) ListDueCustomers(ctx context.Context, userID int64, dueBefore time.Time, limit int) ([]Customer, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, user_id, import_key, name, phone, email, website, stage, source, next_follow_up_at, created_at, updated_at
		FROM crm_customers
		WHERE user_id = $1 AND next_follow_up_at IS NOT NULL AND next_follow_up_at <= $2
		ORDER BY next_follow_up_at ASC, id ASC
		LIMIT $3
	`, userID, dueBefore, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var customers []Customer
	for rows.Next() {
		customer, err := scanCustomer(rows)
		if err != nil {
			return nil, err
		}
		customers = append(customers, customer)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return customers, nil
}

type customerScanner interface {
	Scan(dest ...any) error
}

func scanCustomer(scanner customerScanner) (Customer, error) {
	var customer Customer
	var nextFollowUp pgtype.Timestamptz
	if err := scanner.Scan(
		&customer.ID,
		&customer.UserID,
		&customer.ImportKey,
		&customer.Name,
		&customer.Phone,
		&customer.Email,
		&customer.Website,
		&customer.Stage,
		&customer.Source,
		&nextFollowUp,
		&customer.CreatedAt,
		&customer.UpdatedAt,
	); err != nil {
		return Customer{}, err
	}
	if nextFollowUp.Valid {
		customer.NextFollowUpAt = nextFollowUp.Time
	}
	return customer, nil
}
