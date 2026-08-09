package projects

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
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

func (r *PostgresRepository) ListDictionaryItems(ctx context.Context, kind string) ([]DictionaryItem, error) {
	rows, err := r.db.Query(ctx, `
		SELECT code, kind, name_zh, COALESCE(name_en, ''), sort_order
		FROM project_dict_items
		WHERE kind = $1 AND is_active = TRUE
		ORDER BY sort_order ASC, code ASC
	`, kind)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]DictionaryItem, 0)
	for rows.Next() {
		var item DictionaryItem
		if err := rows.Scan(&item.Code, &item.Kind, &item.NameZH, &item.NameEN, &item.Sort); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *PostgresRepository) ListProjects(ctx context.Context, filters ProjectFilters) (ProjectPage, error) {
	orderBy := "heat DESC, published_at DESC, id ASC"
	if filters.Sort == "latest" {
		orderBy = "published_at DESC, id ASC"
	}
	query := `
		SELECT id, slug, title, COALESCE(cover_url, ''),
		       COALESCE(category_code, ''), COALESCE(track_code, industry, ''), difficulty,
		       invest_cents, budget_band, COALESCE(revenue_range, ''), is_real,
		       COALESCE(primary_source_url, ''), summary, tags, heat, is_featured,
		       resource_requirements, detail, published_at, updated_at,
		       COUNT(*) OVER()
		FROM project_opportunities
		WHERE status = 'published'
		  AND ($1 = '' OR title ILIKE '%' || $1 || '%' OR summary ILIKE '%' || $1 || '%'
		       OR industry ILIKE '%' || $1 || '%' OR tags::TEXT ILIKE '%' || $1 || '%')
		  AND ($2 = '' OR category_code = $2)
		  AND ($3 = '' OR track_code = $3 OR industry = $3)
		  AND ($4 = '' OR budget_band = $4)
		  AND ($5 = '' OR difficulty = $5)
		  AND ($6 = '' OR resource_requirements ? $6)
		  AND ($7 = FALSE OR is_featured = $8)
		ORDER BY ` + orderBy + `
		LIMIT $9 OFFSET $10
	`
	featuredSet := filters.Featured != nil
	featured := false
	if filters.Featured != nil {
		featured = *filters.Featured
	}
	rows, err := r.db.Query(ctx, query,
		filters.Keyword,
		filters.Category,
		filters.Track,
		filters.Budget,
		filters.Difficulty,
		filters.Resource,
		featuredSet,
		featured,
		filters.PageSize,
		(filters.Page-1)*filters.PageSize,
	)
	if err != nil {
		return ProjectPage{}, err
	}
	defer rows.Close()
	page := ProjectPage{Items: make([]Project, 0), Page: filters.Page, PageSize: filters.PageSize}
	for rows.Next() {
		item, total, err := scanCatalogProjectPageRow(rows)
		if err != nil {
			return ProjectPage{}, err
		}
		page.Total = total
		page.Items = append(page.Items, item)
	}
	return page, rows.Err()
}

func (r *PostgresRepository) GetProject(ctx context.Context, ref string) (Project, error) {
	item, err := scanCatalogProject(r.db.QueryRow(ctx, `
		SELECT id, slug, title, COALESCE(cover_url, ''),
		       COALESCE(category_code, ''), COALESCE(track_code, industry, ''), difficulty,
		       invest_cents, budget_band, COALESCE(revenue_range, ''), is_real,
		       COALESCE(primary_source_url, ''), summary, tags, heat, is_featured,
		       resource_requirements, detail, published_at, updated_at
		FROM project_opportunities
		WHERE status = 'published' AND (slug = $1 OR id::TEXT = $1)
	`, ref))
	if errors.Is(err, pgx.ErrNoRows) {
		return Project{}, ErrOpportunityNotFound
	}
	return item, err
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
	rows, err := r.db.Query(ctx, `
		SELECT c.id, c.slug, c.opportunity_id,
		       COALESCE(o.slug, ''), COALESCE(o.title, ''), COALESCE(o.industry, ''),
		       c.title, c.summary, c.case_type, c.outcome, c.key_actions, c.lessons, c.pitfalls,
		       c.source_title, c.source_url, c.captured_at, c.status, c.published_at, c.updated_at
		FROM project_cases c
		LEFT JOIN project_opportunities o ON o.id = c.opportunity_id
		WHERE c.status = 'published'
		  AND ($1 = '' OR c.case_type = $1)
		  AND ($2 = '' OR c.opportunity_id = (
			  SELECT id
			  FROM project_opportunities
			  WHERE slug = $2 AND status = 'published'
		  ))
		  AND ($3 = '' OR o.industry = $3)
		ORDER BY c.published_at DESC, c.id ASC
		LIMIT $4
	`, filters.CaseType, filters.OpportunitySlug, filters.Industry, filters.Limit)
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
	item, err := scanCase(r.db.QueryRow(ctx, `
		SELECT c.id, c.slug, c.opportunity_id,
		       COALESCE(o.slug, ''), COALESCE(o.title, ''), COALESCE(o.industry, ''),
		       c.title, c.summary, c.case_type, c.outcome, c.key_actions, c.lessons, c.pitfalls,
		       c.source_title, c.source_url, c.captured_at, c.status, c.published_at, c.updated_at
		FROM project_cases c
		LEFT JOIN project_opportunities o ON o.id = c.opportunity_id
		WHERE c.slug = $1 AND c.status = 'published'
	`, slug))
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
	answersToStore := session.Answers
	if answersToStore == nil {
		answersToStore = []Answer{}
	}
	answers, err := json.Marshal(answersToStore)
	if err != nil {
		return MatchSession{}, err
	}
	questions, err := json.Marshal(session.Questions)
	if err != nil {
		return MatchSession{}, err
	}
	result, err := json.Marshal(session.Result)
	if err != nil {
		return MatchSession{}, err
	}
	err = r.db.QueryRow(ctx, `
		INSERT INTO project_match_sessions (user_id, intent, answers, status, questions, result, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $7)
		RETURNING id
	`,
		session.UserID,
		session.Intent,
		answers,
		session.Status,
		questions,
		result,
		session.CreatedAt,
	).Scan(&session.ID)
	return session, err
}

func (r *PostgresRepository) ListSessions(ctx context.Context, userID int64, limit int) ([]MatchSession, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, user_id, intent, answers, status, questions, result, created_at, updated_at
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
		SELECT id, user_id, intent, answers, status, questions, result, created_at, updated_at
		FROM project_match_sessions
		WHERE user_id = $1 AND id = $2
	`, userID, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return MatchSession{}, ErrSessionNotFound
	}
	return session, err
}

func (r *PostgresRepository) UpdateSession(ctx context.Context, session MatchSession) (MatchSession, error) {
	answersToStore := session.Answers
	if answersToStore == nil {
		answersToStore = []Answer{}
	}
	answers, err := json.Marshal(answersToStore)
	if err != nil {
		return MatchSession{}, err
	}
	questions, err := json.Marshal(session.Questions)
	if err != nil {
		return MatchSession{}, err
	}
	result, err := json.Marshal(session.Result)
	if err != nil {
		return MatchSession{}, err
	}
	err = r.db.QueryRow(ctx, `UPDATE project_match_sessions SET answers = $3, status = $4, questions = $5, result = $6, updated_at = $7 WHERE user_id = $1 AND id = $2 RETURNING id`, session.UserID, session.ID, answers, session.Status, questions, result, session.UpdatedAt).Scan(&session.ID)
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

func scanCatalogProject(scanner sessionScanner) (Project, error) {
	var item Project
	var tags, resources, detail []byte
	var investCents pgtype.Int8
	if err := scanner.Scan(
		&item.ID, &item.Slug, &item.Title, &item.CoverURL,
		&item.Category, &item.Track, &item.Difficulty,
		&investCents, &item.BudgetBand, &item.RevenueRange, &item.IsReal,
		&item.SourceURL, &item.Summary, &tags, &item.Heat, &item.IsFeatured,
		&resources, &detail, &item.PublishedAt, &item.UpdatedAt,
	); err != nil {
		return Project{}, err
	}
	if err := json.Unmarshal(tags, &item.Tags); err != nil {
		return Project{}, err
	}
	if err := json.Unmarshal(resources, &item.ResourceRequirements); err != nil {
		return Project{}, err
	}
	if investCents.Valid {
		value := investCents.Int64
		item.InvestCents = &value
	}
	item.Detail = append(item.Detail[:0], detail...)
	return item, nil
}

func scanCatalogProjectPageRow(scanner sessionScanner) (Project, int, error) {
	var item Project
	var tags, resources, detail []byte
	var investCents pgtype.Int8
	var total int
	if err := scanner.Scan(
		&item.ID, &item.Slug, &item.Title, &item.CoverURL,
		&item.Category, &item.Track, &item.Difficulty,
		&investCents, &item.BudgetBand, &item.RevenueRange, &item.IsReal,
		&item.SourceURL, &item.Summary, &tags, &item.Heat, &item.IsFeatured,
		&resources, &detail, &item.PublishedAt, &item.UpdatedAt, &total,
	); err != nil {
		return Project{}, 0, err
	}
	if err := json.Unmarshal(tags, &item.Tags); err != nil {
		return Project{}, 0, err
	}
	if err := json.Unmarshal(resources, &item.ResourceRequirements); err != nil {
		return Project{}, 0, err
	}
	if investCents.Valid {
		value := investCents.Int64
		item.InvestCents = &value
	}
	item.Detail = append(item.Detail[:0], detail...)
	return item, total, nil
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
	if err := scanner.Scan(
		&item.ID, &item.Slug, &item.OpportunityID,
		&item.OpportunitySlug, &item.OpportunityTitle, &item.Industry,
		&item.Title, &item.Summary, &item.CaseType, &item.Outcome,
		&actions, &lessons, &pitfalls,
		&item.SourceTitle, &item.SourceURL, &item.CapturedAt,
		&item.Status, &item.PublishedAt, &item.UpdatedAt,
	); err != nil {
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
	var answers []byte
	var questions []byte
	var result []byte
	if err := scanner.Scan(
		&session.ID,
		&session.UserID,
		&session.Intent,
		&answers,
		&session.Status,
		&questions,
		&result,
		&session.CreatedAt,
		&session.UpdatedAt,
	); err != nil {
		return MatchSession{}, err
	}
	if err := json.Unmarshal(answers, &session.Answers); err != nil {
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
