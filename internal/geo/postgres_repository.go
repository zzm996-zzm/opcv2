package geo

import (
	"context"
	"errors"

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
	engines, err := r.engineCoverages(ctx, userID)
	if err != nil {
		return Overview{}, err
	}
	leadSignals, err := r.leadSignals(ctx, userID)
	if err != nil {
		return Overview{}, err
	}
	keywords, err := r.keywordOpportunities(ctx, userID)
	if err != nil {
		return Overview{}, err
	}
	contentTasks, err := r.contentTasks(ctx, userID)
	if err != nil {
		return Overview{}, err
	}
	return ensureOverviewSlices(Overview{
		Stats:        stats,
		Engines:      engines,
		LeadSignals:  leadSignals,
		Keywords:     keywords,
		ContentTasks: contentTasks,
	}), nil
}

func (r *PostgresRepository) CreateAnalysisRequest(ctx context.Context, userID int64, input AnalysisRequestInput) (AnalysisRequest, error) {
	return scanAnalysisRequest(r.db.QueryRow(ctx, `
		INSERT INTO geo_analysis_requests (user_id, target, status)
		VALUES ($1, $2, $3)
		RETURNING id, user_id, target, status, error_message, created_at, updated_at
	`, userID, input.Target, AnalysisRequestStatusQueued))
}

func (r *PostgresRepository) ListAnalysisRequests(ctx context.Context, userID int64, limit int) ([]AnalysisRequest, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, user_id, target, status, error_message, created_at, updated_at
		FROM geo_analysis_requests
		WHERE user_id = $1
		ORDER BY created_at DESC, id DESC
		LIMIT $2
	`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []AnalysisRequest
	for rows.Next() {
		request, err := scanAnalysisRequest(rows)
		if err != nil {
			return nil, err
		}
		requests = append(requests, request)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if requests == nil {
		return []AnalysisRequest{}, nil
	}
	return requests, nil
}

func (r *PostgresRepository) GetAnalysisRequest(ctx context.Context, userID, id int64) (AnalysisRequest, error) {
	request, err := scanAnalysisRequest(r.db.QueryRow(ctx, `
		SELECT id, user_id, target, status, error_message, created_at, updated_at
		FROM geo_analysis_requests
		WHERE user_id = $1 AND id = $2
	`, userID, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return AnalysisRequest{}, ErrAnalysisRequestNotFound
	}
	return request, err
}

func (r *PostgresRepository) UpdateAnalysisRequestStatus(ctx context.Context, id int64, status string, errorMessage string) (AnalysisRequest, error) {
	request, err := scanAnalysisRequest(r.db.QueryRow(ctx, `
		UPDATE geo_analysis_requests
		SET status = $2, error_message = $3, updated_at = NOW()
		WHERE id = $1
		RETURNING id, user_id, target, status, error_message, created_at, updated_at
	`, id, status, errorMessage))
	if errors.Is(err, pgx.ErrNoRows) {
		return AnalysisRequest{}, ErrAnalysisRequestNotFound
	}
	return request, err
}

func (r *PostgresRepository) metrics(ctx context.Context, userID int64) ([]Metric, error) {
	rows, err := r.db.Query(ctx, `
		SELECT key, label, value
		FROM geo_metrics
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

func (r *PostgresRepository) engineCoverages(ctx context.Context, userID int64) ([]EngineCoverage, error) {
	rows, err := r.db.Query(ctx, `
		SELECT name, coverage_percent, status
		FROM geo_engine_coverages
		WHERE user_id = $1
		ORDER BY sort_order ASC, id ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var engines []EngineCoverage
	for rows.Next() {
		var engine EngineCoverage
		if err := rows.Scan(&engine.Name, &engine.CoveragePercent, &engine.Status); err != nil {
			return nil, err
		}
		engines = append(engines, engine)
	}
	return engines, rows.Err()
}

func (r *PostgresRepository) leadSignals(ctx context.Context, userID int64) ([]LeadSignal, error) {
	rows, err := r.db.Query(ctx, `
		SELECT title, detail
		FROM geo_lead_signals
		WHERE user_id = $1
		ORDER BY sort_order ASC, id ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var signals []LeadSignal
	for rows.Next() {
		var signal LeadSignal
		if err := rows.Scan(&signal.Title, &signal.Detail); err != nil {
			return nil, err
		}
		signals = append(signals, signal)
	}
	return signals, rows.Err()
}

func (r *PostgresRepository) keywordOpportunities(ctx context.Context, userID int64) ([]KeywordOpportunity, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, query, intent, coverage, score, action
		FROM geo_keyword_opportunities
		WHERE user_id = $1
		ORDER BY score DESC, id ASC
		LIMIT 50
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keywords []KeywordOpportunity
	for rows.Next() {
		var keyword KeywordOpportunity
		if err := rows.Scan(&keyword.ID, &keyword.Query, &keyword.Intent, &keyword.Coverage, &keyword.Score, &keyword.Action); err != nil {
			return nil, err
		}
		keywords = append(keywords, keyword)
	}
	return keywords, rows.Err()
}

func (r *PostgresRepository) contentTasks(ctx context.Context, userID int64) ([]ContentTask, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, type, title, priority, due_at
		FROM geo_content_tasks
		WHERE user_id = $1
		ORDER BY sort_order ASC, id ASC
		LIMIT 50
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []ContentTask
	for rows.Next() {
		var task ContentTask
		if err := rows.Scan(&task.ID, &task.Type, &task.Title, &task.Priority, &task.DueAt); err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	return tasks, rows.Err()
}

type scanner interface {
	Scan(dest ...any) error
}

func scanAnalysisRequest(scanner scanner) (AnalysisRequest, error) {
	var request AnalysisRequest
	err := scanner.Scan(
		&request.ID,
		&request.UserID,
		&request.Target,
		&request.Status,
		&request.ErrorMessage,
		&request.CreatedAt,
		&request.UpdatedAt,
	)
	return request, err
}
