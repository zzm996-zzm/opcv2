package growth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

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
	if err == nil {
		model.UpdatedAt = model.CreatedAt
	}
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

func (r *PostgresRepository) CreateSnapshot(ctx context.Context, snapshot ModelSnapshot) (ModelSnapshot, error) {
	assumptions, err := json.Marshal(snapshot.Assumptions)
	if err != nil {
		return ModelSnapshot{}, err
	}
	result, err := json.Marshal(snapshot.Result)
	if err != nil {
		return ModelSnapshot{}, err
	}
	scenarios, err := json.Marshal(snapshot.Scenarios)
	if err != nil {
		return ModelSnapshot{}, err
	}
	forecast, err := json.Marshal(snapshot.Forecast)
	if err != nil {
		return ModelSnapshot{}, err
	}
	recommendations, err := json.Marshal(snapshot.Recommendations)
	if err != nil {
		return ModelSnapshot{}, err
	}
	err = r.db.QueryRow(ctx, `
		INSERT INTO growth_model_snapshots (user_id, model_id, model_name, assumptions, result, scenarios, forecast, recommendations, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`, snapshot.UserID, snapshot.ModelID, snapshot.ModelName, assumptions, result, scenarios, forecast, recommendations, snapshot.CreatedAt).Scan(&snapshot.ID)
	return snapshot, err
}

func (r *PostgresRepository) ListSnapshots(ctx context.Context, userID, modelID int64, limit int) ([]ModelSnapshot, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, user_id, model_id, model_name, assumptions, result, scenarios, forecast, recommendations, created_at
		FROM growth_model_snapshots
		WHERE user_id = $1 AND model_id = $2
		ORDER BY created_at DESC
		LIMIT $3
	`, userID, modelID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var snapshots []ModelSnapshot
	for rows.Next() {
		snapshot, err := scanSnapshot(rows)
		if err != nil {
			return nil, err
		}
		snapshots = append(snapshots, snapshot)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return snapshots, nil
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

func (r *PostgresRepository) ListModelPage(ctx context.Context, input ListModelsInput) (ModelPage, error) {
	businessTypeExpr := `CASE
		WHEN name ILIKE '%saas%' OR name ILIKE '%软件%' OR name ILIKE '%系统%' THEN 'saas'
		WHEN name ILIKE '%电商%' OR name ILIKE '%商城%' OR name ILIKE '%零售%' THEN 'ecommerce'
		WHEN name ILIKE '%教培%' OR name ILIKE '%培训%' OR name ILIKE '%课程%' THEN 'education'
		WHEN name ILIKE '%咨询%' OR name ILIKE '%服务%' OR name ILIKE '%顾问%' THEN 'service'
		WHEN name ILIKE '%内容%' OR name ILIKE '%自媒体%' OR name ILIKE '%社群%' THEN 'content'
		ELSE 'other'
	END`
	riskExpression := `CASE
		WHEN COALESCE(NULLIF(result->>'net_margin', '')::numeric, 0) < 0
			OR COALESCE(NULLIF(result->>'payback_days', '')::numeric, 0) > 90
			OR COALESCE(NULLIF(assumptions->>'deal_rate', '')::numeric, 0) < 0.08
			OR COALESCE(NULLIF(assumptions->>'acquisition_cost', '')::numeric, 0) /
				NULLIF(COALESCE(NULLIF(assumptions->>'average_order', '')::numeric, 0) * COALESCE(NULLIF(assumptions->>'deal_rate', '')::numeric, 0), 0) > 0.30 THEN 'high'
		WHEN COALESCE(NULLIF(result->>'net_margin', '')::numeric, 0) < 0.20
			OR COALESCE(NULLIF(result->>'payback_days', '')::numeric, 0) > 45
			OR COALESCE(NULLIF(assumptions->>'deal_rate', '')::numeric, 0) < 0.15
			OR COALESCE(NULLIF(assumptions->>'acquisition_cost', '')::numeric, 0) /
				NULLIF(COALESCE(NULLIF(assumptions->>'average_order', '')::numeric, 0) * COALESCE(NULLIF(assumptions->>'deal_rate', '')::numeric, 0), 0) > 0.15 THEN 'medium'
		ELSE 'low'
	END`
	whereClause := fmt.Sprintf(`
		WHERE user_id = $1
		  AND ($2 = '' OR name ILIKE '%%' || $2 || '%%')
		  AND ($3::timestamptz IS NULL OR created_at >= $3)
		  AND ($4::timestamptz IS NULL OR created_at < $4)
		  AND ($5 = '' OR (%s) = $5)
		  AND ($6 = '' OR $6 = 'completed')
		  AND ($7 = '' OR (%s) = $7)
	`, businessTypeExpr, riskExpression)
	var total int
	if err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM growth_models "+whereClause,
		input.UserID, input.Query, input.From, input.To, input.BusinessType, input.Status, input.RiskLevel).Scan(&total); err != nil {
		return ModelPage{}, err
	}
	rows, err := r.db.Query(ctx, fmt.Sprintf(`
		SELECT id, user_id, name, assumptions, result, created_at, updated_at
		FROM growth_models
		%s
		ORDER BY
		  CASE WHEN $8 = 'revenue_desc' THEN COALESCE(NULLIF(result->>'monthly_revenue', '')::numeric, 0) END DESC NULLS LAST,
		  CASE WHEN $8 = 'revenue_asc' THEN COALESCE(NULLIF(result->>'monthly_revenue', '')::numeric, 0) END ASC NULLS LAST,
		  CASE WHEN $8 = 'margin_desc' THEN COALESCE(NULLIF(result->>'net_margin', '')::numeric, 0) END DESC NULLS LAST,
		  CASE WHEN $8 = 'margin_asc' THEN COALESCE(NULLIF(result->>'net_margin', '')::numeric, 0) END ASC NULLS LAST,
		  CASE WHEN $8 = 'created_asc' THEN created_at END ASC NULLS LAST,
		  created_at DESC, id DESC
		LIMIT $9 OFFSET $10
	`, whereClause), input.UserID, input.Query, input.From, input.To, input.BusinessType, input.Status, input.RiskLevel, input.Sort, input.Limit, input.Offset)
	if err != nil {
		return ModelPage{}, err
	}
	defer rows.Close()
	models := make([]Model, 0)
	for rows.Next() {
		model, err := scanModel(rows)
		if err != nil {
			return ModelPage{}, err
		}
		models = append(models, model)
	}
	if err := rows.Err(); err != nil {
		return ModelPage{}, err
	}
	return ModelPage{Models: models, Total: total, Limit: input.Limit, Offset: input.Offset}, nil
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

func scanSnapshot(scanner modelScanner) (ModelSnapshot, error) {
	var snapshot ModelSnapshot
	var assumptions, result, scenarios, forecast, recommendations []byte
	if err := scanner.Scan(&snapshot.ID, &snapshot.UserID, &snapshot.ModelID, &snapshot.ModelName, &assumptions, &result, &scenarios, &forecast, &recommendations, &snapshot.CreatedAt); err != nil {
		return ModelSnapshot{}, err
	}
	if err := json.Unmarshal(assumptions, &snapshot.Assumptions); err != nil {
		return ModelSnapshot{}, err
	}
	if err := json.Unmarshal(result, &snapshot.Result); err != nil {
		return ModelSnapshot{}, err
	}
	if err := json.Unmarshal(scenarios, &snapshot.Scenarios); err != nil {
		return ModelSnapshot{}, err
	}
	if err := json.Unmarshal(forecast, &snapshot.Forecast); err != nil {
		return ModelSnapshot{}, err
	}
	if err := json.Unmarshal(recommendations, &snapshot.Recommendations); err != nil {
		return ModelSnapshot{}, err
	}
	return snapshot, nil
}
