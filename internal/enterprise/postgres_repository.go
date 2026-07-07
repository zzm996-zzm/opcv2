package enterprise

import (
	"context"
	"errors"
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

func (r *PostgresRepository) ListDiagnosisRequests(ctx context.Context, userID int64, limit int) ([]DiagnosisRequest, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, user_id, need, status, created_at, updated_at
		FROM enterprise_diagnosis_requests
		WHERE user_id = $1
		ORDER BY created_at DESC, id DESC
		LIMIT $2
	`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	requests := []DiagnosisRequest{}
	for rows.Next() {
		var request DiagnosisRequest
		var createdAt time.Time
		var updatedAt time.Time
		if err := rows.Scan(&request.ID, &request.UserID, &request.Need, &request.Status, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		request.CreatedAt = createdAt.Format(time.RFC3339)
		request.UpdatedAt = updatedAt.Format(time.RFC3339)
		requests = append(requests, request)
	}
	return requests, rows.Err()
}

func (r *PostgresRepository) GetDiagnosisRequest(ctx context.Context, userID int64, requestID int64) (DiagnosisRequest, error) {
	var request DiagnosisRequest
	var createdAt time.Time
	var updatedAt time.Time
	err := r.db.QueryRow(ctx, `
		SELECT id, user_id, need, status, created_at, updated_at
		FROM enterprise_diagnosis_requests
		WHERE id = $1 AND user_id = $2
	`, requestID, userID).Scan(
		&request.ID,
		&request.UserID,
		&request.Need,
		&request.Status,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return DiagnosisRequest{}, ErrDiagnosisRequestNotFound
		}
		return DiagnosisRequest{}, err
	}
	request.CreatedAt = createdAt.Format(time.RFC3339)
	request.UpdatedAt = updatedAt.Format(time.RFC3339)
	return request, nil
}

func (r *PostgresRepository) UpdateDiagnosisRequest(ctx context.Context, userID int64, requestID int64, input DiagnosisRequestUpdateInput) (DiagnosisRequest, error) {
	var request DiagnosisRequest
	var createdAt time.Time
	var updatedAt time.Time
	err := r.db.QueryRow(ctx, `
		UPDATE enterprise_diagnosis_requests
		SET status = $3, updated_at = now()
		WHERE id = $1 AND user_id = $2
		RETURNING id, user_id, need, status, created_at, updated_at
	`, requestID, userID, input.Status).Scan(
		&request.ID,
		&request.UserID,
		&request.Need,
		&request.Status,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return DiagnosisRequest{}, ErrDiagnosisRequestNotFound
		}
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
		FROM (
			SELECT sort_order, id, stage, count, detail
			FROM enterprise_delivery_board
			WHERE user_id = $1
			UNION ALL
			SELECT
				900 + CASE status
					WHEN 'submitted' THEN 1
					WHEN 'follow_up_created' THEN 2
					WHEN 'in_delivery' THEN 3
					WHEN 'completed' THEN 4
					ELSE 9
				END AS sort_order,
				0 AS id,
				CASE status
					WHEN 'submitted' THEN '待承接预约'
					WHEN 'follow_up_created' THEN '已生成跟进'
					WHEN 'in_delivery' THEN '交付中预约'
					WHEN 'completed' THEN '已完成交付'
					ELSE '其他预约'
				END AS stage,
				COUNT(*)::int AS count,
				CASE status
					WHEN 'submitted' THEN '等待生成跟进任务'
					WHEN 'follow_up_created' THEN '已生成任务，等待进入交付'
					WHEN 'in_delivery' THEN '已进入企业陪跑交付'
					WHEN 'completed' THEN '已完成交付并沉淀案例'
					ELSE '其他诊断预约状态'
				END AS detail
			FROM enterprise_diagnosis_requests
			WHERE user_id = $1
			GROUP BY status
		) board
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
		FROM (
			SELECT sort_order, id, time_label, title, detail
			FROM enterprise_milestones
			WHERE user_id = $1
			UNION ALL
			SELECT
				900 AS sort_order,
				id,
				to_char(updated_at AT TIME ZONE 'Asia/Shanghai', 'MM-DD') AS time_label,
				'交付启动' AS title,
				need AS detail
			FROM enterprise_diagnosis_requests
			WHERE user_id = $1 AND status = 'in_delivery'
		) milestones
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
		FROM (
			SELECT sort_order, id, company, result
			FROM enterprise_cases
			WHERE user_id = $1
			UNION ALL
			SELECT
				900 AS sort_order,
				-id AS id,
				'企业诊断交付' AS company,
				need AS result
			FROM enterprise_diagnosis_requests
			WHERE user_id = $1 AND status = 'completed'
		) cases
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
