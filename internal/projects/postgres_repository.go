package projects

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
}

type PostgresRepository struct {
	db postgresDB
}

func NewPostgresRepository(db postgresDB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) ListOpportunities(ctx context.Context, filters OpportunityFilters) ([]Opportunity, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, slug, title, summary, industry, tags, budget_band, difficulty, resource_requirements, sections, status, published_at, updated_at
		FROM project_opportunities
		WHERE status = 'published'
		  AND ($1 = '' OR industry = $1)
		  AND ($2 = '' OR title ILIKE '%' || $2 || '%' OR summary ILIKE '%' || $2 || '%')
		ORDER BY sort_order ASC, published_at DESC, id ASC
		LIMIT $3
	`, filters.Industry, filters.Query, filters.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]Opportunity, 0)
	for rows.Next() {
		item, err := scanOpportunity(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *PostgresRepository) GetOpportunity(ctx context.Context, slug string) (Opportunity, error) {
	item, err := scanOpportunity(r.db.QueryRow(ctx, `
		SELECT id, slug, title, summary, industry, tags, budget_band, difficulty, resource_requirements, sections, status, published_at, updated_at
		FROM project_opportunities
		WHERE slug = $1 AND status = 'published'
	`, slug))
	if errors.Is(err, pgx.ErrNoRows) {
		return Opportunity{}, ErrOpportunityNotFound
	}
	return item, err
}

func (r *PostgresRepository) ListCases(ctx context.Context, filters CaseFilters) ([]CaseStudy, error) {
	rows, err := r.db.Query(ctx, `SELECT id, slug, opportunity_id, title, summary, case_type, outcome, key_actions, lessons, pitfalls, source_title, source_url, captured_at, status, published_at, updated_at FROM project_cases WHERE status = 'published' AND ($1 = '' OR case_type = $1) ORDER BY published_at DESC, id ASC LIMIT $2`, filters.CaseType, filters.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]CaseStudy, 0)
	for rows.Next() {
		item, err := scanCase(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *PostgresRepository) GetCase(ctx context.Context, slug string) (CaseStudy, error) {
	item, err := scanCase(r.db.QueryRow(ctx, `SELECT id, slug, opportunity_id, title, summary, case_type, outcome, key_actions, lessons, pitfalls, source_title, source_url, captured_at, status, published_at, updated_at FROM project_cases WHERE slug = $1 AND status = 'published'`, slug))
	if errors.Is(err, pgx.ErrNoRows) {
		return CaseStudy{}, ErrCaseNotFound
	}
	return item, err
}

func (r *PostgresRepository) CreateComparison(ctx context.Context, comparison Comparison) (Comparison, error) {
	items, err := json.Marshal(comparison.Items)
	if err != nil {
		return Comparison{}, err
	}
	err = r.db.QueryRow(ctx, `INSERT INTO project_comparisons (user_id, items, created_at) VALUES ($1, $2, $3) RETURNING id`, comparison.UserID, items, comparison.CreatedAt).Scan(&comparison.ID)
	return comparison, err
}
func (r *PostgresRepository) GetComparison(ctx context.Context, userID, id int64) (Comparison, error) {
	var comparison Comparison
	var items []byte
	err := r.db.QueryRow(ctx, `SELECT id, user_id, items, created_at FROM project_comparisons WHERE user_id = $1 AND id = $2`, userID, id).Scan(&comparison.ID, &comparison.UserID, &items, &comparison.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Comparison{}, ErrComparisonNotFound
	}
	if err != nil {
		return Comparison{}, err
	}
	if err := json.Unmarshal(items, &comparison.Items); err != nil {
		return Comparison{}, err
	}
	return comparison, nil
}

func (r *PostgresRepository) CreateExport(ctx context.Context, item Export) (Export, error) {
	err := r.db.QueryRow(ctx, `INSERT INTO project_exports (user_id, source_type, source_id, status, payload, created_at) VALUES ($1,$2,$3,$4,$5,$6) RETURNING id`, item.UserID, item.SourceType, item.SourceID, item.Status, item.Payload, item.CreatedAt).Scan(&item.ID)
	return item, err
}
func (r *PostgresRepository) GetExport(ctx context.Context, userID, id int64) (Export, error) {
	var item Export
	err := r.db.QueryRow(ctx, `SELECT id, user_id, source_type, source_id, status, payload, created_at FROM project_exports WHERE user_id = $1 AND id = $2`, userID, id).Scan(&item.ID, &item.UserID, &item.SourceType, &item.SourceID, &item.Status, &item.Payload, &item.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Export{}, ErrExportNotFound
	}
	return item, err
}

func (r *PostgresRepository) CreateSession(ctx context.Context, session MatchSession) (MatchSession, error) {
	questions, err := json.Marshal(session.Questions)
	if err != nil {
		return MatchSession{}, err
	}
	result, err := json.Marshal(session.Result)
	if err != nil {
		return MatchSession{}, err
	}
	err = r.db.QueryRow(ctx, `
		INSERT INTO project_match_sessions (user_id, intent, status, questions, result, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $6)
		RETURNING id
	`,
		session.UserID,
		session.Intent,
		session.Status,
		questions,
		result,
		session.CreatedAt,
	).Scan(&session.ID)
	return session, err
}

func (r *PostgresRepository) ListSessions(ctx context.Context, userID int64, limit int) ([]MatchSession, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, user_id, intent, status, questions, result, created_at, updated_at
		FROM project_match_sessions
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sessions := make([]MatchSession, 0)
	for rows.Next() {
		session, err := scanSession(rows)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return sessions, nil
}

func (r *PostgresRepository) GetSession(ctx context.Context, userID, id int64) (MatchSession, error) {
	session, err := scanSession(r.db.QueryRow(ctx, `
		SELECT id, user_id, intent, status, questions, result, created_at, updated_at
		FROM project_match_sessions
		WHERE user_id = $1 AND id = $2
	`, userID, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return MatchSession{}, ErrSessionNotFound
	}
	return session, err
}

func (r *PostgresRepository) UpdateSession(ctx context.Context, session MatchSession) (MatchSession, error) {
	questions, err := json.Marshal(session.Questions)
	if err != nil {
		return MatchSession{}, err
	}
	result, err := json.Marshal(session.Result)
	if err != nil {
		return MatchSession{}, err
	}
	err = r.db.QueryRow(ctx, `UPDATE project_match_sessions SET status = $3, questions = $4, result = $5, updated_at = $6 WHERE user_id = $1 AND id = $2 RETURNING id`, session.UserID, session.ID, session.Status, questions, result, session.UpdatedAt).Scan(&session.ID)
	if errors.Is(err, pgx.ErrNoRows) {
		return MatchSession{}, ErrSessionNotFound
	}
	return session, err
}

func (r *PostgresRepository) SaveFavorite(ctx context.Context, favorite Favorite) (Favorite, error) {
	err := r.db.QueryRow(ctx, `
		INSERT INTO project_match_favorites (user_id, session_id, created_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, session_id) DO UPDATE SET session_id = EXCLUDED.session_id
		RETURNING id, created_at
	`,
		favorite.UserID,
		favorite.SessionID,
		favorite.CreatedAt,
	).Scan(&favorite.ID, &favorite.CreatedAt)
	return favorite, err
}

func (r *PostgresRepository) ListFavorites(ctx context.Context, userID int64, limit int) ([]Favorite, error) {
	rows, err := r.db.Query(ctx, `
		SELECT f.id, f.user_id, f.session_id, f.created_at,
		       s.id, s.user_id, s.intent, s.status, s.questions, s.result, s.created_at, s.updated_at
		FROM project_match_favorites f
		JOIN project_match_sessions s ON s.id = f.session_id AND s.user_id = f.user_id
		WHERE f.user_id = $1
		ORDER BY f.created_at DESC
		LIMIT $2
	`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]Favorite, 0)
	for rows.Next() {
		var item Favorite
		var session MatchSession
		var questions, result []byte
		if err := rows.Scan(
			&item.ID, &item.UserID, &item.SessionID, &item.CreatedAt,
			&session.ID, &session.UserID, &session.Intent, &session.Status,
			&questions, &result, &session.CreatedAt, &session.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(questions, &session.Questions); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(result, &session.Result); err != nil {
			return nil, err
		}
		item.Session = &session
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *PostgresRepository) DeleteFavorite(ctx context.Context, userID, sessionID int64) error {
	_, err := r.db.Exec(ctx, `DELETE FROM project_match_favorites WHERE user_id = $1 AND session_id = $2`, userID, sessionID)
	return err
}

type sessionScanner interface {
	Scan(dest ...any) error
}

func scanOpportunity(scanner sessionScanner) (Opportunity, error) {
	var item Opportunity
	var tags, resources, sections []byte
	if err := scanner.Scan(&item.ID, &item.Slug, &item.Title, &item.Summary, &item.Industry, &tags, &item.BudgetBand, &item.Difficulty, &resources, &sections, &item.Status, &item.PublishedAt, &item.UpdatedAt); err != nil {
		return Opportunity{}, err
	}
	if err := json.Unmarshal(tags, &item.Tags); err != nil {
		return Opportunity{}, err
	}
	if err := json.Unmarshal(resources, &item.ResourceRequirements); err != nil {
		return Opportunity{}, err
	}
	if err := json.Unmarshal(sections, &item.Sections); err != nil {
		return Opportunity{}, err
	}
	return item, nil
}

func scanCase(scanner sessionScanner) (CaseStudy, error) {
	var item CaseStudy
	var actions, lessons, pitfalls []byte
	if err := scanner.Scan(&item.ID, &item.Slug, &item.OpportunityID, &item.Title, &item.Summary, &item.CaseType, &item.Outcome, &actions, &lessons, &pitfalls, &item.SourceTitle, &item.SourceURL, &item.CapturedAt, &item.Status, &item.PublishedAt, &item.UpdatedAt); err != nil {
		return CaseStudy{}, err
	}
	if err := json.Unmarshal(actions, &item.KeyActions); err != nil {
		return CaseStudy{}, err
	}
	if err := json.Unmarshal(lessons, &item.Lessons); err != nil {
		return CaseStudy{}, err
	}
	if err := json.Unmarshal(pitfalls, &item.Pitfalls); err != nil {
		return CaseStudy{}, err
	}
	return item, nil
}

func scanSession(scanner sessionScanner) (MatchSession, error) {
	var session MatchSession
	var questions []byte
	var result []byte
	if err := scanner.Scan(
		&session.ID,
		&session.UserID,
		&session.Intent,
		&session.Status,
		&questions,
		&result,
		&session.CreatedAt,
		&session.UpdatedAt,
	); err != nil {
		return MatchSession{}, err
	}
	if err := json.Unmarshal(questions, &session.Questions); err != nil {
		return MatchSession{}, err
	}
	if err := json.Unmarshal(result, &session.Result); err != nil {
		return MatchSession{}, err
	}
	return session, nil
}
