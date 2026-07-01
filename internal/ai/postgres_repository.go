package ai

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type postgresDB interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

type PostgresRepository struct {
	db postgresDB
}

func NewPostgresRepository(db postgresDB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) CreateRun(ctx context.Context, run Run) (Run, error) {
	err := r.db.QueryRow(ctx, `
		INSERT INTO ai_runs (
			user_id,
			feature,
			prompt_version,
			provider,
			model,
			status,
			request,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $8)
		RETURNING id
	`,
		run.UserID,
		run.Feature,
		run.PromptVersion,
		run.Provider,
		run.Model,
		run.Status,
		run.Request,
		run.CreatedAt,
	).Scan(&run.ID)
	return run, err
}

func (r *PostgresRepository) CompleteRun(ctx context.Context, id int64, result RunResult) error {
	_, err := r.db.Exec(ctx, `
		UPDATE ai_runs
		SET status = $2,
		    response = $3,
		    input_tokens = $4,
		    output_tokens = $5,
		    latency_ms = $6,
		    error_code = '',
		    error_message = '',
		    updated_at = NOW()
		WHERE id = $1
	`,
		id,
		StatusCompleted,
		result.Response,
		result.InputTokens,
		result.OutputTokens,
		result.LatencyMS,
	)
	return err
}

func (r *PostgresRepository) FailRun(ctx context.Context, id int64, failure RunFailure) error {
	_, err := r.db.Exec(ctx, `
		UPDATE ai_runs
		SET status = $2,
		    error_code = $3,
		    error_message = $4,
		    latency_ms = $5,
		    updated_at = NOW()
		WHERE id = $1
	`,
		id,
		StatusFailed,
		failure.Code,
		failure.Message,
		failure.LatencyMS,
	)
	return err
}

func (r *PostgresRepository) ListRuns(ctx context.Context, userID int64, featurePrefix string, limit int) ([]Run, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			id,
			user_id,
			feature,
			prompt_version,
			provider,
			model,
			status,
			error_code,
			error_message,
			input_tokens,
			output_tokens,
			latency_ms,
			created_at,
			updated_at
		FROM ai_runs
		WHERE user_id = $1
		  AND feature LIKE $2
		ORDER BY created_at DESC
		LIMIT $3
	`, userID, featurePrefix+"%", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	runs := []Run{}
	for rows.Next() {
		var run Run
		if err := rows.Scan(
			&run.ID,
			&run.UserID,
			&run.Feature,
			&run.PromptVersion,
			&run.Provider,
			&run.Model,
			&run.Status,
			&run.ErrorCode,
			&run.ErrorMessage,
			&run.InputTokens,
			&run.OutputTokens,
			&run.LatencyMS,
			&run.CreatedAt,
			&run.UpdatedAt,
		); err != nil {
			return nil, err
		}
		runs = append(runs, run)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return runs, nil
}
