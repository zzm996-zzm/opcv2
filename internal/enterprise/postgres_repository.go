package enterprise

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

type postgresDB interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type PostgresRepository struct {
	db postgresDB
}

func NewPostgresRepository(db postgresDB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Overview(ctx context.Context, userID int64) (Overview, error) {
	stats, err := r.metrics(ctx, userID)
	if err != nil {
		return Overview{}, err
	}
	plans, err := r.plans(ctx, userID)
	if err != nil {
		return Overview{}, err
	}
	deliveryBoard, err := r.deliveryBoard(ctx, userID)
	if err != nil {
		return Overview{}, err
	}
	milestones, err := r.milestones(ctx, userID)
	if err != nil {
		return Overview{}, err
	}
	cases, err := r.cases(ctx, userID)
	if err != nil {
		return Overview{}, err
	}
	return ensureOverviewSlices(Overview{
		Stats:         stats,
		Plans:         plans,
		DeliveryBoard: deliveryBoard,
		Milestones:    milestones,
		Cases:         cases,
	}), nil
}

func (r *PostgresRepository) CreateDiagnosisRequest(ctx context.Context, userID int64, input DiagnosisRequestInput) (DiagnosisRequest, error) {
	var request DiagnosisRequest
	var createdAt time.Time
	var updatedAt time.Time
	err := r.db.QueryRow(ctx, `
		INSERT INTO enterprise_diagnosis_requests (user_id, need)
		VALUES ($1, $2)
		RETURNING id, user_id, need, status, created_at, updated_at
	`, userID, input.Need).Scan(
		&request.ID,
		&request.UserID,
		&request.Need,
		&request.Status,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return DiagnosisRequest{}, err
	}
	request.CreatedAt = createdAt.Format(time.RFC3339)
	request.UpdatedAt = updatedAt.Format(time.RFC3339)
	return request, nil
}

func (r *PostgresRepository) metrics(ctx context.Context, userID int64) ([]Metric, error) {
	rows, err := r.db.Query(ctx, `
		SELECT key, label, value
		FROM enterprise_metrics
		WHERE user_id = $1
		ORDER BY sort_order ASC, id ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metrics []Metric
	for rows.Next() {
		var metric Metric
		if err := rows.Scan(&metric.Key, &metric.Label, &metric.Value); err != nil {
			return nil, err
		}
		metrics = append(metrics, metric)
	}
	return metrics, rows.Err()
}

func (r *PostgresRepository) plans(ctx context.Context, userID int64) ([]Plan, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, title, audience, price_label, focus, result
		FROM enterprise_plans
		WHERE user_id = $1
		ORDER BY sort_order ASC, id ASC
		LIMIT 50
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var plans []Plan
	for rows.Next() {
		var plan Plan
		if err := rows.Scan(&plan.ID, &plan.Title, &plan.Audience, &plan.PriceLabel, &plan.Focus, &plan.Result); err != nil {
			return nil, err
		}
		plans = append(plans, plan)
	}
	return plans, rows.Err()
}

func (r *PostgresRepository) deliveryBoard(ctx context.Context, userID int64) ([]DeliveryItem, error) {
	rows, err := r.db.Query(ctx, `
		SELECT stage, count, detail
		FROM enterprise_delivery_board
		WHERE user_id = $1
		ORDER BY sort_order ASC, id ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []DeliveryItem
	for rows.Next() {
		var item DeliveryItem
		if err := rows.Scan(&item.Stage, &item.Count, &item.Detail); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *PostgresRepository) milestones(ctx context.Context, userID int64) ([]Milestone, error) {
	rows, err := r.db.Query(ctx, `
		SELECT time_label, title, detail
		FROM enterprise_milestones
		WHERE user_id = $1
		ORDER BY sort_order ASC, id ASC
		LIMIT 50
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var milestones []Milestone
	for rows.Next() {
		var milestone Milestone
		if err := rows.Scan(&milestone.TimeLabel, &milestone.Title, &milestone.Detail); err != nil {
			return nil, err
		}
		milestones = append(milestones, milestone)
	}
	return milestones, rows.Err()
}

func (r *PostgresRepository) cases(ctx context.Context, userID int64) ([]Case, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, company, result
		FROM enterprise_cases
		WHERE user_id = $1
		ORDER BY sort_order ASC, id ASC
		LIMIT 50
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cases []Case
	for rows.Next() {
		var item Case
		if err := rows.Scan(&item.ID, &item.Company, &item.Result); err != nil {
			return nil, err
		}
		cases = append(cases, item)
	}
	return cases, rows.Err()
}
