package enterprise

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
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

func (r *PostgresRepository) PublicOverview(ctx context.Context) (PublicOverview, error) {
	var overview PublicOverview
	var proofPoints []byte
	var stats []byte
	var serviceSteps []byte
	var sourceUpdatedAt pgtype.Timestamptz
	var updatedAt time.Time
	err := r.db.QueryRow(ctx, `
		SELECT headline, subheadline, description, proof_points, stats, service_steps, source_name, source_url, source_updated_at, updated_at
		FROM enterprise_public_overview
		WHERE status = 'published'
		ORDER BY updated_at DESC, id DESC
		LIMIT 1
	`).Scan(
		&overview.Headline,
		&overview.Subheadline,
		&overview.Description,
		&proofPoints,
		&stats,
		&serviceSteps,
		&overview.SourceName,
		&overview.SourceURL,
		&sourceUpdatedAt,
		&updatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return emptyPublicOverview(), nil
	}
	if err != nil {
		return PublicOverview{}, err
	}
	if err := decodeJSON(proofPoints, &overview.ProofPoints); err != nil {
		return PublicOverview{}, err
	}
	if err := decodeJSON(stats, &overview.Stats); err != nil {
		return PublicOverview{}, err
	}
	if err := decodeJSON(serviceSteps, &overview.ServiceSteps); err != nil {
		return PublicOverview{}, err
	}
	if sourceUpdatedAt.Valid {
		overview.SourceUpdatedAt = sourceUpdatedAt.Time.Format(time.RFC3339)
	}
	overview.UpdatedAt = updatedAt.Format(time.RFC3339)
	return ensurePublicOverviewSlices(overview), nil
}

func (r *PostgresRepository) ListPublicCases(ctx context.Context, limit int) ([]PublicCase, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, slug, company, title, summary, result, industry, services, metrics, body, source_name, source_url, source_updated_at, updated_at
		FROM enterprise_public_cases
		WHERE status = 'published'
		ORDER BY sort_order ASC, published_at DESC, id DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cases := []PublicCase{}
	for rows.Next() {
		item, err := scanPublicCase(rows)
		if err != nil {
			return nil, err
		}
		cases = append(cases, item)
	}
	return cases, rows.Err()
}

func (r *PostgresRepository) GetPublicCase(ctx context.Context, slug string) (PublicCase, error) {
	item, err := scanPublicCase(r.db.QueryRow(ctx, `
		SELECT id, slug, company, title, summary, result, industry, services, metrics, body, source_name, source_url, source_updated_at, updated_at
		FROM enterprise_public_cases
		WHERE slug = $1 AND status = 'published'
	`, slug))
	if errors.Is(err, pgx.ErrNoRows) {
		return PublicCase{}, ErrPublicCaseNotFound
	}
	return item, err
}

func (r *PostgresRepository) ContactConfig(ctx context.Context) (ContactConfig, int64, error) {
	var config ContactConfig
	var ownerUserID pgtype.Int8
	var updatedAt time.Time
	err := r.db.QueryRow(ctx, `
		SELECT consultant_name, title, description, phone, email, wechat, qr_image_url, contact_url, source_name, crm_owner_user_id, updated_at
		FROM enterprise_contact_config
		WHERE status = 'published'
		ORDER BY updated_at DESC, id DESC
		LIMIT 1
	`).Scan(
		&config.ConsultantName,
		&config.Title,
		&config.Description,
		&config.Phone,
		&config.Email,
		&config.Wechat,
		&config.QRImageURL,
		&config.ContactURL,
		&config.SourceName,
		&ownerUserID,
		&updatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return ContactConfig{}, 0, nil
	}
	if err != nil {
		return ContactConfig{}, 0, err
	}
	config.UpdatedAt = updatedAt.Format(time.RFC3339)
	return config, ownerUserID.Int64, nil
}

func (r *PostgresRepository) CreateInquiry(ctx context.Context, input InquiryInput) (Inquiry, error) {
	var inquiry Inquiry
	var crmCustomerID pgtype.Int8
	var createdAt time.Time
	var updatedAt time.Time
	err := r.db.QueryRow(ctx, `
		INSERT INTO enterprise_inquiries (company, name, phone, email, wechat, need, budget, timeline, source_page)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, company, name, phone, email, wechat, need, budget, timeline, source_page, status, crm_customer_id, created_at, updated_at
	`,
		input.Company,
		input.Name,
		input.Phone,
		input.Email,
		input.Wechat,
		input.Need,
		input.Budget,
		input.Timeline,
		input.SourcePage,
	).Scan(
		&inquiry.ID,
		&inquiry.Company,
		&inquiry.Name,
		&inquiry.Phone,
		&inquiry.Email,
		&inquiry.Wechat,
		&inquiry.Need,
		&inquiry.Budget,
		&inquiry.Timeline,
		&inquiry.SourcePage,
		&inquiry.Status,
		&crmCustomerID,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return Inquiry{}, err
	}
	if crmCustomerID.Valid {
		inquiry.CRMCustomerID = crmCustomerID.Int64
	}
	inquiry.CreatedAt = createdAt.Format(time.RFC3339)
	inquiry.UpdatedAt = updatedAt.Format(time.RFC3339)
	return inquiry, nil
}

func (r *PostgresRepository) UpdateInquiryCRMCustomer(ctx context.Context, inquiryID int64, customerID int64) (Inquiry, error) {
	var inquiry Inquiry
	var crmCustomerID pgtype.Int8
	var createdAt time.Time
	var updatedAt time.Time
	err := r.db.QueryRow(ctx, `
		UPDATE enterprise_inquiries
		SET crm_customer_id = $2, status = 'crm_synced', updated_at = now()
		WHERE id = $1
		RETURNING id, company, name, phone, email, wechat, need, budget, timeline, source_page, status, crm_customer_id, created_at, updated_at
	`, inquiryID, customerID).Scan(
		&inquiry.ID,
		&inquiry.Company,
		&inquiry.Name,
		&inquiry.Phone,
		&inquiry.Email,
		&inquiry.Wechat,
		&inquiry.Need,
		&inquiry.Budget,
		&inquiry.Timeline,
		&inquiry.SourcePage,
		&inquiry.Status,
		&crmCustomerID,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return Inquiry{}, err
	}
	if crmCustomerID.Valid {
		inquiry.CRMCustomerID = crmCustomerID.Int64
	}
	inquiry.CreatedAt = createdAt.Format(time.RFC3339)
	inquiry.UpdatedAt = updatedAt.Format(time.RFC3339)
	return inquiry, nil
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

type scanner interface {
	Scan(dest ...any) error
}

func scanPublicCase(row scanner) (PublicCase, error) {
	var item PublicCase
	var services []byte
	var metrics []byte
	var sourceUpdatedAt pgtype.Timestamptz
	var updatedAt time.Time
	err := row.Scan(
		&item.ID,
		&item.Slug,
		&item.Company,
		&item.Title,
		&item.Summary,
		&item.Result,
		&item.Industry,
		&services,
		&metrics,
		&item.Body,
		&item.SourceName,
		&item.SourceURL,
		&sourceUpdatedAt,
		&updatedAt,
	)
	if err != nil {
		return PublicCase{}, err
	}
	if err := decodeJSON(services, &item.Services); err != nil {
		return PublicCase{}, err
	}
	if err := decodeJSON(metrics, &item.Metrics); err != nil {
		return PublicCase{}, err
	}
	if sourceUpdatedAt.Valid {
		item.SourceUpdatedAt = sourceUpdatedAt.Time.Format(time.RFC3339)
	}
	item.UpdatedAt = updatedAt.Format(time.RFC3339)
	return ensurePublicCaseSlices(item), nil
}

func decodeJSON(data []byte, target any) error {
	if len(data) == 0 {
		data = []byte("[]")
	}
	return json.Unmarshal(data, target)
}
