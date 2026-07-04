package leads

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type postgresDB interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Begin(ctx context.Context) (pgx.Tx, error)
}

type PostgresRepository struct {
	db postgresDB
}

func NewPostgresRepository(db postgresDB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) CreateTask(ctx context.Context, task Task) (Task, bool, error) {
	err := r.db.QueryRow(ctx, `
		INSERT INTO lead_tasks (user_id, query, status, idempotency_key, credit_cost, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $6)
		ON CONFLICT (user_id, idempotency_key) DO NOTHING
		RETURNING id, created_at, updated_at
	`,
		task.UserID,
		task.Query,
		task.Status,
		task.IdempotencyKey,
		task.CreditCost,
		task.CreatedAt,
	).Scan(&task.ID, &task.CreatedAt, &task.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		existing, getErr := r.getTaskByIdempotencyKey(ctx, task.UserID, task.IdempotencyKey)
		return existing, true, getErr
	}
	return task, false, err
}

func (r *PostgresRepository) GetTask(ctx context.Context, id int64) (Task, error) {
	task, err := scanTask(r.db.QueryRow(ctx, `
		SELECT id, user_id, query, status, idempotency_key, credit_cost, error_code, created_at, updated_at
		FROM lead_tasks
		WHERE id = $1
	`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return Task{}, ErrTaskNotFound
	}
	return task, err
}

func (r *PostgresRepository) UpdateTaskStatus(ctx context.Context, id int64, status string, errorCode string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE lead_tasks
		SET status = $2, error_code = $3, updated_at = NOW()
		WHERE id = $1
	`, id, status, errorCode)
	return err
}

func (r *PostgresRepository) StoreResults(ctx context.Context, taskID int64, leads []Lead) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, `DELETE FROM lead_results WHERE task_id = $1`, taskID); err != nil {
		return err
	}
	for _, lead := range leads {
		evidence, err := json.Marshal(lead.Evidence)
		if err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO lead_results (task_id, name, phone, email, website, evidence)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, taskID, lead.Name, lead.Phone, lead.Email, lead.Website, evidence); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *PostgresRepository) ListResults(ctx context.Context, taskID int64, limit int) ([]LeadResult, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, task_id, name, phone, email, website, evidence, created_at
		FROM lead_results
		WHERE task_id = $1
		ORDER BY created_at DESC, id DESC
		LIMIT $2
	`, taskID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []LeadResult
	for rows.Next() {
		result, err := scanLeadResult(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return results, nil
}

func (r *PostgresRepository) ListTasks(ctx context.Context, userID int64, limit int) ([]Task, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, user_id, query, status, idempotency_key, credit_cost, error_code, created_at, updated_at
		FROM lead_tasks
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return tasks, nil
}

func (r *PostgresRepository) getTaskByIdempotencyKey(ctx context.Context, userID int64, key string) (Task, error) {
	return scanTask(r.db.QueryRow(ctx, `
		SELECT id, user_id, query, status, idempotency_key, credit_cost, error_code, created_at, updated_at
		FROM lead_tasks
		WHERE user_id = $1 AND idempotency_key = $2
	`, userID, key))
}

type taskScanner interface {
	Scan(dest ...any) error
}

func scanTask(scanner taskScanner) (Task, error) {
	var task Task
	if err := scanner.Scan(
		&task.ID,
		&task.UserID,
		&task.Query,
		&task.Status,
		&task.IdempotencyKey,
		&task.CreditCost,
		&task.ErrorCode,
		&task.CreatedAt,
		&task.UpdatedAt,
	); err != nil {
		return Task{}, err
	}
	return task, nil
}

func scanLeadResult(scanner taskScanner) (LeadResult, error) {
	var result LeadResult
	var evidence []byte
	if err := scanner.Scan(
		&result.ID,
		&result.TaskID,
		&result.Name,
		&result.Phone,
		&result.Email,
		&result.Website,
		&evidence,
		&result.CreatedAt,
	); err != nil {
		return LeadResult{}, err
	}
	if err := json.Unmarshal(evidence, &result.Evidence); err != nil {
		return LeadResult{}, err
	}
	return result, nil
}
