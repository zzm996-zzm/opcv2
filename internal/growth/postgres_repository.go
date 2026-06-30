package growth

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
)

type postgresDB interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

type PostgresRepository struct {
	db postgresDB
}

func NewPostgresRepository(db postgresDB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) CreateModel(ctx context.Context, model Model) (Model, error) {
	assumptions, err := json.Marshal(model.Assumptions)
	if err != nil {
		return Model{}, err
	}
	result, err := json.Marshal(model.Result)
	if err != nil {
		return Model{}, err
	}
	err = r.db.QueryRow(ctx, `
		INSERT INTO growth_models (user_id, name, assumptions, result, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $5)
		RETURNING id
	`, model.UserID, model.Name, assumptions, result, model.CreatedAt).Scan(&model.ID)
	return model, err
}

func (r *PostgresRepository) ListModels(ctx context.Context, userID int64, limit int) ([]Model, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, user_id, name, assumptions, result, created_at, updated_at
		FROM growth_models
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var models []Model
	for rows.Next() {
		model, err := scanModel(rows)
		if err != nil {
			return nil, err
		}
		models = append(models, model)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return models, nil
}

func (r *PostgresRepository) GetModel(ctx context.Context, userID, id int64) (Model, error) {
	model, err := scanModel(r.db.QueryRow(ctx, `
		SELECT id, user_id, name, assumptions, result, created_at, updated_at
		FROM growth_models
		WHERE user_id = $1 AND id = $2
	`, userID, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return Model{}, ErrModelNotFound
	}
	return model, err
}

type modelScanner interface {
	Scan(dest ...any) error
}

func scanModel(scanner modelScanner) (Model, error) {
	var model Model
	var assumptions []byte
	var result []byte
	if err := scanner.Scan(
		&model.ID,
		&model.UserID,
		&model.Name,
		&assumptions,
		&result,
		&model.CreatedAt,
		&model.UpdatedAt,
	); err != nil {
		return Model{}, err
	}
	if err := json.Unmarshal(assumptions, &model.Assumptions); err != nil {
		return Model{}, err
	}
	if err := json.Unmarshal(result, &model.Result); err != nil {
		return Model{}, err
	}
	return model, nil
}
