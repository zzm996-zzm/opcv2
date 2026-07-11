package projects

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
	var items []Opportunity
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

	var sessions []MatchSession
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
