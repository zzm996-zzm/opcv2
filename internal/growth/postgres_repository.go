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

func (r *PostgresRepository) CreateDraft(ctx context.Context, draft Draft) (Draft, error) {
	assumptions, err := json.Marshal(draft.Assumptions)
	if err != nil {
		return Draft{}, err
	}
	questions, err := json.Marshal(draft.Questions)
	if err != nil {
		return Draft{}, err
	}
	answers, err := json.Marshal(draft.Answers)
	if err != nil {
		return Draft{}, err
	}
	err = r.db.QueryRow(ctx, `
		INSERT INTO growth_drafts (user_id, input, status, assumptions, questions, answers, model_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $8)
		RETURNING id
	`, draft.UserID, draft.Input, draft.Status, assumptions, questions, answers, nullableInt64(draft.ModelID), draft.CreatedAt).Scan(&draft.ID)
	return draft, err
}

func (r *PostgresRepository) GetDraft(ctx context.Context, userID, id int64) (Draft, error) {
	draft, err := scanDraft(r.db.QueryRow(ctx, `
		SELECT id, user_id, input, status, assumptions, questions, answers, model_id, created_at, updated_at
		FROM growth_drafts
		WHERE user_id = $1 AND id = $2
	`, userID, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return Draft{}, ErrDraftNotFound
	}
	return draft, err
}

func (r *PostgresRepository) UpdateDraft(ctx context.Context, draft Draft) (Draft, error) {
	assumptions, err := json.Marshal(draft.Assumptions)
	if err != nil {
		return Draft{}, err
	}
	questions, err := json.Marshal(draft.Questions)
	if err != nil {
		return Draft{}, err
	}
	answers, err := json.Marshal(draft.Answers)
	if err != nil {
		return Draft{}, err
	}
	err = r.db.QueryRow(ctx, `
		UPDATE growth_drafts
		SET status = $3, assumptions = $4, questions = $5, answers = $6, model_id = $7, updated_at = $8
		WHERE user_id = $1 AND id = $2
		RETURNING id
	`, draft.UserID, draft.ID, draft.Status, assumptions, questions, answers, nullableInt64(draft.ModelID), draft.UpdatedAt).Scan(&draft.ID)
	if errors.Is(err, pgx.ErrNoRows) {
		return Draft{}, ErrDraftNotFound
	}
	return draft, err
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

func scanDraft(scanner modelScanner) (Draft, error) {
	var draft Draft
	var assumptions, questions, answers []byte
	if err := scanner.Scan(&draft.ID, &draft.UserID, &draft.Input, &draft.Status, &assumptions, &questions, &answers, &draft.ModelID, &draft.CreatedAt, &draft.UpdatedAt); err != nil {
		return Draft{}, err
	}
	if err := json.Unmarshal(assumptions, &draft.Assumptions); err != nil {
		return Draft{}, err
	}
	if err := json.Unmarshal(questions, &draft.Questions); err != nil {
		return Draft{}, err
	}
	if err := json.Unmarshal(answers, &draft.Answers); err != nil {
		return Draft{}, err
	}
	return draft, nil
}

func nullableInt64(value *int64) any {
	if value == nil {
		return nil
	}
	return *value
}
