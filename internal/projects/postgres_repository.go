package projects

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	projectfiles "github.com/zzm/opcv2/internal/projects/files"
	projectretrieval "github.com/zzm/opcv2/internal/projects/retrieval"
)

type postgresDB interface {
	Begin(ctx context.Context) (pgx.Tx, error)
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

func (r *PostgresRepository) RecordProjectEvent(ctx context.Context, event AnalyticsEvent) (bool, error) {
	properties, err := json.Marshal(event.Properties)
	if err != nil {
		return false, err
	}
	var recorded bool
	err = r.db.QueryRow(ctx, `
		WITH inserted AS (
			INSERT INTO project_analytics_events
				(event_id, event_name, user_id, visitor_hash, route, ref_module, project_id,
				 properties, occurred_at, created_at)
			VALUES ($1, $2, $3, $4, $5, NULLIF($6, ''), $7, $8, $9, $10)
			ON CONFLICT (event_id) DO NOTHING
			RETURNING id, event_name, project_id, user_id, visitor_hash, occurred_at
		), inserted_view AS (
			INSERT INTO project_views (project_id, user_id, visitor_key, viewed_at, analytics_event_id)
			SELECT project_id, user_id, visitor_hash, occurred_at, id
			FROM inserted
			WHERE event_name = 'project_detail_view' AND project_id IS NOT NULL
			RETURNING id
		)
		SELECT EXISTS(SELECT 1 FROM inserted)
	`, event.EventID, event.EventName, event.UserID, event.VisitorHash, event.Route, event.RefModule,
		event.ProjectID, properties, event.OccurredAt, event.CreatedAt).Scan(&recorded)
	return recorded, err
}

func (r *PostgresRepository) RecalculateProjectHeat(ctx context.Context, projectID int64, cutoff time.Time) error {
	command, err := r.db.Exec(ctx, `
		UPDATE project_opportunities p
		SET heat = metrics.views + metrics.favorites * 3 + metrics.unlocks * 10,
		    updated_at = NOW()
		FROM (
			SELECT
				(SELECT COUNT(DISTINCT visitor_key)::INTEGER FROM project_views WHERE project_id = $1 AND viewed_at >= $2) AS views,
				(SELECT COUNT(*)::INTEGER FROM project_favorites WHERE project_id = $1 AND created_at >= $2) AS favorites,
				(SELECT COUNT(*)::INTEGER FROM project_unlocks WHERE project_id = $1 AND unlocked_at >= $2) AS unlocks
		) metrics
		WHERE p.id = $1
	`, projectID, cutoff)
	if err != nil {
		return err
	}
	if command.RowsAffected() == 0 {
		return ErrOpportunityNotFound
	}
	return nil
}

func (r *PostgresRepository) Search(ctx context.Context, request projectretrieval.SearchRequest) ([]projectretrieval.Document, error) {
	rows, err := r.db.Query(ctx, `
		SELECT document_id, title, body, metadata
		FROM project_kb_documents
		WHERE version = COALESCE(NULLIF($1, '')::INTEGER,
		                      (SELECT active_version FROM project_kb_state WHERE singleton = TRUE))
		ORDER BY document_id ASC
		LIMIT 100
	`, request.KnowledgeBaseVersion)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	documents := make([]projectretrieval.Document, 0)
	for rows.Next() {
		var id, title, body string
		var metadataBytes []byte
		if err := rows.Scan(&id, &title, &body, &metadataBytes); err != nil {
			return nil, err
		}
		metadata := map[string]string{}
		if err := json.Unmarshal(metadataBytes, &metadata); err != nil {
			return nil, err
		}
		documents = append(documents, projectretrieval.Document{
			ID: id, Title: title, Text: body, Metadata: metadata,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return projectretrieval.RankDocuments(request.Query, documents, request.Limit), nil
}

func (r *PostgresRepository) SaveMatchEvidence(ctx context.Context, userID, matchID int64, attempt int, sufficiency float64, evidence []MatchEvidence) error {
	command, err := r.db.Exec(ctx, `
		UPDATE project_match_sessions
		SET kb_sufficiency = $4, updated_at = NOW()
		WHERE user_id = $1 AND id = $2 AND workflow_version = 2 AND generation_attempt = $3
		  AND status IN ('queued', 'running')
	`, userID, matchID, attempt, sufficiency)
	if err != nil {
		return err
	}
	if command.RowsAffected() == 0 {
		return ErrStaleMatchGeneration
	}
	if _, err := r.db.Exec(ctx, `DELETE FROM project_match_evidence WHERE user_id = $1 AND match_id = $2 AND attempt = $3`, userID, matchID, attempt); err != nil {
		return err
	}
	for _, item := range evidence {
		if _, err := r.db.Exec(ctx, `
			INSERT INTO project_match_evidence
				(match_id, user_id, attempt, source_type, source_id, source_url, title, publisher,
				 excerpt, quality_score, untrusted_content)
			VALUES ($1, $2, $3, $4, NULLIF($5, ''), $6, $7, $8, $9, $10, $11)
		`, matchID, userID, attempt, item.SourceType, item.SourceID, item.URL, item.Title, item.Publisher, item.Excerpt, item.Quality, item.UntrustedContent); err != nil {
			return err
		}
	}
	return nil
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
		       OR industry ILIKE '%' || $1 || '%' OR tags::TEXT ILIKE '%' || $1 || '%'
		       OR COALESCE(primary_source_url, '') ILIKE '%' || $1 || '%'
		       OR detail::TEXT ILIKE '%' || $1 || '%')
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
		       resource_requirements,
		       CASE
		           WHEN detail = '{}'::jsonb THEN jsonb_build_object('sections', sections)
		           ELSE detail
		       END AS detail,
		       published_at, updated_at
		FROM project_opportunities
		WHERE status = 'published' AND (slug = $1 OR id::TEXT = $1)
	`, ref))
	if errors.Is(err, pgx.ErrNoRows) {
		return Project{}, ErrOpportunityNotFound
	}
	return item, err
}

func (r *PostgresRepository) ListEvidenceCases(ctx context.Context, filters EvidenceCaseFilters) (EvidenceCasePage, error) {
	query := `
		SELECT c.id, c.title, COALESCE(c.cover_url, ''), COALESCE(c.result_summary, c.outcome),
		       COALESCE(o.industry, ''), COALESCE(c.scale, ''),
		       CASE WHEN c.case_type = 'failure' THEN 'fail' ELSE c.case_type END,
		       COALESCE(MAX(ws.url) FILTER (WHERE cs.is_primary), ''), COUNT(DISTINCT ws.id),
		       c.last_verified_at, c.published_at, c.opportunity_id, c.evidence_status, c.has_conflict,
		       COUNT(DISTINCT ws.id) FILTER (WHERE ws.source_kind IN ('primary', 'authority')),
		       COUNT(DISTINCT ws.id) FILTER (WHERE ws.source_kind IN ('research', 'media', 'vertical')),
		       COUNT(*) OVER()
		FROM project_cases c
		LEFT JOIN project_opportunities o ON o.id = c.opportunity_id
		LEFT JOIN project_case_sources cs ON cs.case_id = c.id
		LEFT JOIN web_sources ws ON ws.id = cs.web_source_id
		WHERE c.status = 'published' AND c.evidence_status = 'verified'
		  AND ($1 = '' OR c.case_type = $1)
		  AND ($2 = '' OR o.industry = $2)
		  AND ($3 = '' OR c.scale = $3)
		GROUP BY c.id, c.title, c.cover_url, c.result_summary, c.outcome, o.industry, c.scale,
		         c.case_type, c.last_verified_at, c.published_at, c.opportunity_id, c.evidence_status,
		         c.has_conflict
		HAVING c.has_conflict = FALSE
		   AND (COUNT(DISTINCT ws.id) FILTER (WHERE ws.source_kind IN ('primary', 'authority')) > 0
		        OR COUNT(DISTINCT ws.id) FILTER (WHERE ws.source_kind IN ('research', 'media', 'vertical')) >= 2)
		ORDER BY c.published_at DESC NULLS LAST, c.id ASC
		LIMIT $4 OFFSET $5
	`
	rows, err := r.db.Query(ctx, query, filters.CaseType, filters.Industry, filters.Scale, filters.PageSize, (filters.Page-1)*filters.PageSize)
	if err != nil {
		return EvidenceCasePage{}, err
	}
	defer rows.Close()
	page := EvidenceCasePage{Items: make([]EvidenceCaseItem, 0), Page: filters.Page, PageSize: filters.PageSize}
	for rows.Next() {
		item, total, err := scanEvidenceCaseItem(rows)
		if err != nil {
			return EvidenceCasePage{}, err
		}
		page.Total = total
		page.Items = append(page.Items, item)
	}
	return page, rows.Err()
}

func (r *PostgresRepository) GetEvidenceCase(ctx context.Context, ref string) (EvidenceCaseDetail, error) {
	var detail EvidenceCaseDetail
	item, _, err := scanEvidenceCaseItem(r.db.QueryRow(ctx, `
		SELECT c.id, c.title, COALESCE(c.cover_url, ''), COALESCE(c.result_summary, c.outcome),
		       COALESCE(o.industry, ''), COALESCE(c.scale, ''),
		       CASE WHEN c.case_type = 'failure' THEN 'fail' ELSE c.case_type END,
		       COALESCE(MAX(ws.url) FILTER (WHERE cs.is_primary), ''), COUNT(DISTINCT ws.id),
		       c.last_verified_at, c.published_at, c.opportunity_id, c.evidence_status, c.has_conflict,
		       COUNT(DISTINCT ws.id) FILTER (WHERE ws.source_kind IN ('primary', 'authority')),
		       COUNT(DISTINCT ws.id) FILTER (WHERE ws.source_kind IN ('research', 'media', 'vertical')),
		       1
		FROM project_cases c
		LEFT JOIN project_opportunities o ON o.id = c.opportunity_id
		LEFT JOIN project_case_sources cs ON cs.case_id = c.id
		LEFT JOIN web_sources ws ON ws.id = cs.web_source_id
		WHERE c.status = 'published' AND c.evidence_status = 'verified'
		  AND (c.slug = $1 OR c.id::TEXT = $1)
		GROUP BY c.id, c.title, c.cover_url, c.result_summary, c.outcome, o.industry, c.scale,
		         c.case_type, c.last_verified_at, c.published_at, c.opportunity_id, c.evidence_status,
		         c.has_conflict
		HAVING c.has_conflict = FALSE
		   AND (COUNT(DISTINCT ws.id) FILTER (WHERE ws.source_kind IN ('primary', 'authority')) > 0
		        OR COUNT(DISTINCT ws.id) FILTER (WHERE ws.source_kind IN ('research', 'media', 'vertical')) >= 2)
	`, ref))
	if errors.Is(err, pgx.ErrNoRows) {
		return EvidenceCaseDetail{}, ErrCaseNotFound
	}
	if err != nil {
		return EvidenceCaseDetail{}, err
	}
	detail.EvidenceCaseItem = item
	if err := r.db.QueryRow(ctx, `SELECT COALESCE(content_md, '') FROM project_cases WHERE id = $1`, item.ID).Scan(&detail.ContentMD); err != nil {
		return EvidenceCaseDetail{}, err
	}

	claimsRows, err := r.db.Query(ctx, `
		SELECT id, claim_type, field_name, value_text, COALESCE(detail, ''), is_model_generated
		FROM project_case_claims
		WHERE case_id = $1
		ORDER BY claim_type ASC, sort_order ASC, id ASC
	`, item.ID)
	if err != nil {
		return EvidenceCaseDetail{}, err
	}
	defer claimsRows.Close()
	for claimsRows.Next() {
		var id int64
		var claimType, fieldName, value, claimDetail string
		var modelGenerated bool
		if err := claimsRows.Scan(&id, &claimType, &fieldName, &value, &claimDetail, &modelGenerated); err != nil {
			return EvidenceCaseDetail{}, err
		}
		var refs []int64
		refRows, err := r.db.Query(ctx, `SELECT web_source_id FROM project_case_claim_sources WHERE claim_id = $1 ORDER BY web_source_id`, id)
		if err != nil {
			return EvidenceCaseDetail{}, err
		}
		for refRows.Next() {
			var sourceID int64
			if err := refRows.Scan(&sourceID); err != nil {
				refRows.Close()
				return EvidenceCaseDetail{}, err
			}
			refs = append(refs, sourceID)
		}
		refErr := refRows.Err()
		refRows.Close()
		if refErr != nil {
			return EvidenceCaseDetail{}, refErr
		}
		if claimType == "fact" {
			detail.Facts = append(detail.Facts, EvidenceFact{Field: fieldName, Value: value, SourceRefs: refs})
		} else {
			detail.Analyses = append(detail.Analyses, EvidenceAnalysis{Point: value, Detail: claimDetail, IsModelGenerated: modelGenerated, SourceRefs: refs})
		}
	}
	if err := claimsRows.Err(); err != nil {
		return EvidenceCaseDetail{}, err
	}

	sourceRows, err := r.db.Query(ctx, `
		SELECT ws.id, COALESCE(ws.title, ''), COALESCE(ws.publisher, ''), ws.url,
		       ws.published_at, ws.fetched_at, ws.quality_score, ws.source_kind,
		       COALESCE(ARRAY(SELECT DISTINCT c.field_name FROM project_case_sources c WHERE c.case_id = $1 AND c.web_source_id = ws.id), '{}'),
		       COALESCE((SELECT BOOL_OR(c.is_primary) FROM project_case_sources c WHERE c.case_id = $1 AND c.web_source_id = ws.id), FALSE)
		FROM web_sources ws
		WHERE ws.id IN (
			SELECT web_source_id FROM project_case_sources WHERE case_id = $1
			UNION
			SELECT cs.web_source_id FROM project_case_claim_sources cs
			JOIN project_case_claims c ON c.id = cs.claim_id
			WHERE c.case_id = $1
		)
		ORDER BY ws.quality_score DESC NULLS LAST, ws.id ASC
	`, item.ID)
	if err != nil {
		return EvidenceCaseDetail{}, err
	}
	defer sourceRows.Close()
	for sourceRows.Next() {
		var source EvidenceSource
		var quality pgtype.Float8
		if err := sourceRows.Scan(&source.ID, &source.Title, &source.Publisher, &source.URL, &source.PublishedAt, &source.FetchedAt, &quality, &source.Kind, &source.ClaimFields, &source.IsPrimary); err != nil {
			return EvidenceCaseDetail{}, err
		}
		if quality.Valid {
			value := quality.Float64
			source.Quality = &value
		}
		detail.Sources = append(detail.Sources, source)
	}
	return detail, sourceRows.Err()
}

func scanEvidenceCaseItem(scanner sessionScanner) (EvidenceCaseItem, int, error) {
	var item EvidenceCaseItem
	var projectID pgtype.Int8
	var total int
	if err := scanner.Scan(
		&item.ID, &item.Title, &item.CoverURL, &item.ResultSummary, &item.Industry, &item.Scale,
		&item.Type, &item.PrimarySourceURL, &item.SourceCount, &item.VerifiedAt, &item.PublishedAt,
		&projectID, &item.EvidenceStatus, &item.HasConflict, &item.PrimarySourceCount,
		&item.ReliableSecondaryCount, &total,
	); err != nil {
		return EvidenceCaseItem{}, 0, err
	}
	if projectID.Valid {
		value := projectID.Int64
		item.ProjectID = &value
	}
	return item, total, nil
}

func (r *PostgresRepository) ListOpportunities(ctx context.Context, filters OpportunityFilters) ([]Opportunity, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, slug, title, summary, industry, tags, budget_band, difficulty, resource_requirements, sections, status, published_at, updated_at
		FROM project_opportunities
		WHERE status = 'published'
		  AND ($1 = '' OR industry = $1)
		  AND ($2 = ''
		    OR title ILIKE '%' || $2 || '%'
		    OR summary ILIKE '%' || $2 || '%'
		    OR industry ILIKE '%' || $2 || '%'
		    OR tags::TEXT ILIKE '%' || $2 || '%'
		    OR resource_requirements::TEXT ILIKE '%' || $2 || '%'
		    OR sections::TEXT ILIKE '%' || $2 || '%'
		    OR detail::TEXT ILIKE '%' || $2 || '%')
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
	if item.ExpiresAt.IsZero() {
		item.ExpiresAt = item.CreatedAt.Add(7 * 24 * time.Hour)
	}
	if item.Format == "" {
		item.Format = "json"
	}
	includes, err := json.Marshal(nonNilStrings(item.Includes))
	if err != nil {
		return Export{}, err
	}
	if len(item.Snapshot) == 0 {
		item.Snapshot = []byte(`{}`)
	}
	if item.UpdatedAt.IsZero() {
		item.UpdatedAt = item.CreatedAt
	}
	err = r.db.QueryRow(ctx, `INSERT INTO project_exports (user_id, source_type, source_id, status, payload, snapshot, payload_bytes, format, includes, expires_at, created_at, updated_at, error_code) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13) RETURNING id`, item.UserID, item.SourceType, item.SourceID, item.Status, []byte(`{}`), item.Snapshot, nil, item.Format, includes, item.ExpiresAt, item.CreatedAt, item.UpdatedAt, item.ErrorCode).Scan(&item.ID)
	return item, err
}
func (r *PostgresRepository) GetExport(ctx context.Context, userID, id int64) (Export, error) {
	var item Export
	var includes []byte
	err := r.db.QueryRow(ctx, `SELECT id, user_id, source_type, source_id, status, snapshot, COALESCE(payload_bytes, CASE WHEN status = 'ready' THEN payload::text::bytea END), format, includes, expires_at, created_at, updated_at, error_code FROM project_exports WHERE user_id = $1 AND id = $2`, userID, id).Scan(&item.ID, &item.UserID, &item.SourceType, &item.SourceID, &item.Status, &item.Snapshot, &item.Payload, &item.Format, &includes, &item.ExpiresAt, &item.CreatedAt, &item.UpdatedAt, &item.ErrorCode)
	if errors.Is(err, pgx.ErrNoRows) {
		return Export{}, ErrExportNotFound
	}
	if err == nil && len(includes) > 0 {
		if unmarshalErr := json.Unmarshal(includes, &item.Includes); unmarshalErr != nil {
			return Export{}, unmarshalErr
		}
	}
	return item, err
}

func (r *PostgresRepository) FindReusableExport(ctx context.Context, userID int64, sourceType string, sourceID int64, format string, now time.Time) (Export, error) {
	var id int64
	err := r.db.QueryRow(ctx, `SELECT id FROM project_exports WHERE user_id=$1 AND source_type=$2 AND source_id=$3 AND format=$4 AND status IN ('queued','running','ready') AND expires_at>$5 ORDER BY id DESC LIMIT 1`, userID, sourceType, sourceID, format, now).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return Export{}, ErrExportNotFound
	}
	if err != nil {
		return Export{}, err
	}
	return r.GetExport(ctx, userID, id)
}

func (r *PostgresRepository) UpdateExport(ctx context.Context, item Export) (Export, error) {
	payloadJSON := []byte(`{}`)
	var payloadBytes []byte
	if item.Format == "pdf" {
		payloadBytes = item.Payload
	} else if len(item.Payload) > 0 {
		payloadJSON = item.Payload
	}
	err := r.db.QueryRow(ctx, `UPDATE project_exports SET status=$3, payload=$4, payload_bytes=$5, error_code=$6, updated_at=$7 WHERE id=$1 AND user_id=$2 RETURNING updated_at`, item.ID, item.UserID, item.Status, payloadJSON, payloadBytes, item.ErrorCode, item.UpdatedAt).Scan(&item.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Export{}, ErrExportNotFound
	}
	return item, err
}

func (r *PostgresRepository) SaveProjectFavorite(ctx context.Context, favorite ProjectFavorite) (ProjectFavorite, error) {
	err := r.db.QueryRow(ctx, `
		INSERT INTO project_favorites (user_id, project_id, created_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, project_id) DO UPDATE SET project_id = EXCLUDED.project_id
		RETURNING created_at
	`, favorite.UserID, favorite.ProjectID, favorite.CreatedAt).Scan(&favorite.CreatedAt)
	if err != nil {
		return ProjectFavorite{}, err
	}
	return r.getProjectFavorite(ctx, favorite.UserID, favorite.ProjectID)
}

func (r *PostgresRepository) getProjectFavorite(ctx context.Context, userID, projectID int64) (ProjectFavorite, error) {
	var favorite ProjectFavorite
	err := r.db.QueryRow(ctx, `
		SELECT f.user_id, f.project_id, p.slug, p.title, f.created_at
		FROM project_favorites f JOIN project_opportunities p ON p.id = f.project_id
		WHERE f.user_id = $1 AND f.project_id = $2
	`, userID, projectID).Scan(&favorite.UserID, &favorite.ProjectID, &favorite.Slug, &favorite.Title, &favorite.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return ProjectFavorite{}, ErrOpportunityNotFound
	}
	return favorite, err
}

func (r *PostgresRepository) ListProjectFavorites(ctx context.Context, userID int64, limit int) ([]ProjectFavorite, error) {
	rows, err := r.db.Query(ctx, `
		SELECT f.user_id, f.project_id, p.slug, p.title, f.created_at
		FROM project_favorites f JOIN project_opportunities p ON p.id = f.project_id
		WHERE f.user_id = $1 AND p.status = 'published'
		ORDER BY f.created_at DESC, f.project_id DESC LIMIT $2
	`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]ProjectFavorite, 0)
	for rows.Next() {
		var item ProjectFavorite
		if err := rows.Scan(&item.UserID, &item.ProjectID, &item.Slug, &item.Title, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *PostgresRepository) DeleteProjectFavorite(ctx context.Context, userID, projectID int64) error {
	_, err := r.db.Exec(ctx, `DELETE FROM project_favorites WHERE user_id = $1 AND project_id = $2`, userID, projectID)
	return err
}

func (r *PostgresRepository) AddProjectCompareItem(ctx context.Context, item ProjectCompareItem, userID int64) (ProjectCompareItem, error) {
	err := r.db.QueryRow(ctx, `
		WITH user_lock AS (
			SELECT pg_advisory_xact_lock($1)
		)
		INSERT INTO project_compare_items (user_id, project_id, added_at)
		SELECT $1, $2, $3
		FROM user_lock
		WHERE (SELECT COUNT(*) FROM project_compare_items WHERE user_id = $1) < 5
		   OR EXISTS (SELECT 1 FROM project_compare_items WHERE user_id = $1 AND project_id = $2)
		ON CONFLICT (user_id, project_id) DO UPDATE SET added_at = project_compare_items.added_at
		RETURNING added_at
	`, userID, item.ProjectID, item.AddedAt).Scan(&item.AddedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return ProjectCompareItem{}, ErrCompareLimit
	}
	if err != nil {
		return ProjectCompareItem{}, err
	}
	return r.getProjectCompareItem(ctx, userID, item.ProjectID)
}

func (r *PostgresRepository) getProjectCompareItem(ctx context.Context, userID, projectID int64) (ProjectCompareItem, error) {
	var item ProjectCompareItem
	err := r.db.QueryRow(ctx, `
		SELECT c.project_id, p.slug, p.title, c.added_at
		FROM project_compare_items c JOIN project_opportunities p ON p.id = c.project_id
		WHERE c.user_id = $1 AND c.project_id = $2
	`, userID, projectID).Scan(&item.ProjectID, &item.Slug, &item.Title, &item.AddedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return ProjectCompareItem{}, ErrOpportunityNotFound
	}
	return item, err
}

func (r *PostgresRepository) ListProjectCompareItems(ctx context.Context, userID int64) ([]ProjectCompareItem, error) {
	rows, err := r.db.Query(ctx, `
		SELECT c.project_id, p.slug, p.title, c.added_at
		FROM project_compare_items c JOIN project_opportunities p ON p.id = c.project_id
		WHERE c.user_id = $1 AND p.status = 'published'
		ORDER BY c.added_at ASC, c.project_id ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]ProjectCompareItem, 0)
	for rows.Next() {
		var item ProjectCompareItem
		if err := rows.Scan(&item.ProjectID, &item.Slug, &item.Title, &item.AddedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *PostgresRepository) DeleteProjectCompareItem(ctx context.Context, userID, projectID int64) error {
	_, err := r.db.Exec(ctx, `DELETE FROM project_compare_items WHERE user_id = $1 AND project_id = $2`, userID, projectID)
	return err
}

func (r *PostgresRepository) CreateContentCorrection(ctx context.Context, item ContentCorrection) (ContentCorrection, error) {
	err := r.db.QueryRow(ctx, `
		INSERT INTO content_corrections (target_type, target_id, reason, evidence_url, contact, status, created_at)
		VALUES ($1, $2, $3, NULLIF($4, ''), NULLIF($5, ''), $6, $7)
		RETURNING id, status, created_at
	`, item.TargetType, item.TargetID, item.Reason, item.EvidenceURL, item.Contact, item.Status, item.CreatedAt).Scan(&item.ID, &item.Status, &item.CreatedAt)
	return item, err
}

func (r *PostgresRepository) CreateFile(ctx context.Context, file projectfiles.File) (projectfiles.File, error) {
	extractedJSON, err := json.Marshal(nonNilMap(file.ExtractedJSON))
	if err != nil {
		return projectfiles.File{}, err
	}
	return scanProjectFile(r.db.QueryRow(ctx, `
		INSERT INTO project_match_files (
			user_id, match_id, original_name, mime_type, detected_mime, size_bytes,
			storage_key, sha256, parse_status, extracted_text, extracted_json, error_code,
			expires_at, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NULLIF($10, ''), $11, NULLIF($12, ''), $13, $14, $15)
		RETURNING id, user_id, match_id, original_name, mime_type, detected_mime, size_bytes,
		          storage_key, sha256, parse_status, COALESCE(extracted_text, ''), extracted_json,
		          COALESCE(error_code, ''), expires_at, created_at, updated_at
	`, file.UserID, file.MatchID, file.Name, file.MIME, file.DetectedMIME, file.Size, file.ObjectKey,
		file.SHA256, file.ParseStatus, file.ExtractedText, extractedJSON, file.ErrorCode,
		file.ExpiresAt, file.CreatedAt, file.UpdatedAt))
}

func (r *PostgresRepository) GetFile(ctx context.Context, userID, fileID int64) (projectfiles.File, error) {
	file, err := scanProjectFile(r.db.QueryRow(ctx, `
		SELECT id, user_id, match_id, original_name, mime_type, detected_mime, size_bytes,
		       storage_key, sha256, parse_status, COALESCE(extracted_text, ''), extracted_json,
		       COALESCE(error_code, ''), expires_at, created_at, updated_at
		FROM project_match_files
		WHERE user_id = $1 AND id = $2 AND parse_status <> 'deleted'
	`, userID, fileID))
	if errors.Is(err, pgx.ErrNoRows) {
		return projectfiles.File{}, projectfiles.ErrFileNotFound
	}
	return file, err
}

func (r *PostgresRepository) ListFiles(ctx context.Context, userID int64, matchID *int64) ([]projectfiles.File, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, user_id, match_id, original_name, mime_type, detected_mime, size_bytes,
		       storage_key, sha256, parse_status, COALESCE(extracted_text, ''), extracted_json,
		       COALESCE(error_code, ''), expires_at, created_at, updated_at
		FROM project_match_files
		WHERE user_id = $1 AND parse_status <> 'deleted'
		  AND ($2::BIGINT IS NULL OR match_id = $2)
		ORDER BY created_at DESC, id DESC
	`, userID, matchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]projectfiles.File, 0)
	for rows.Next() {
		file, scanErr := scanProjectFile(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		items = append(items, file)
	}
	return items, rows.Err()
}

func (r *PostgresRepository) UpdateFile(ctx context.Context, file projectfiles.File) (projectfiles.File, error) {
	extractedJSON, err := json.Marshal(nonNilMap(file.ExtractedJSON))
	if err != nil {
		return projectfiles.File{}, err
	}
	updated, err := scanProjectFile(r.db.QueryRow(ctx, `
		UPDATE project_match_files
		SET match_id = $3, parse_status = $4, extracted_text = NULLIF($5, ''),
		    extracted_json = $6, error_code = NULLIF($7, ''), updated_at = $8
		WHERE id = $1 AND user_id = $2
		RETURNING id, user_id, match_id, original_name, mime_type, detected_mime, size_bytes,
		          storage_key, sha256, parse_status, COALESCE(extracted_text, ''), extracted_json,
		          COALESCE(error_code, ''), expires_at, created_at, updated_at
	`, file.ID, file.UserID, file.MatchID, file.ParseStatus, file.ExtractedText, extractedJSON, file.ErrorCode, file.UpdatedAt))
	if errors.Is(err, pgx.ErrNoRows) {
		return projectfiles.File{}, projectfiles.ErrFileNotFound
	}
	return updated, err
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
		WHERE user_id = $1 AND COALESCE(workflow_version, 1) = 1
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
		WHERE user_id = $1 AND id = $2 AND COALESCE(workflow_version, 1) = 1
	`, userID, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return MatchSession{}, ErrSessionNotFound
	}
	return session, err
}

func (r *PostgresRepository) FindMatchRunByIdempotency(ctx context.Context, userID int64, key string) (MatchRun, error) {
	if key == "" {
		return MatchRun{}, ErrSessionNotFound
	}
	run, err := scanMatchRun(r.db.QueryRow(ctx, `
		SELECT id, user_id, workflow_version, COALESCE(name, ''), intent, input_snapshot, parsed_profile,
		       field_sources, COALESCE(analysis_summary, ''), completeness, missing_fields,
		       clarification_questions, question_count, rounds, answer_events, assumptions,
		       revision, skipped_at, generation_attempt, progress_percent, current_step, result,
		       COALESCE(error_code, ''), cancelled_at, status, idempotency_key, created_at, updated_at
		FROM project_match_sessions
		WHERE user_id = $1 AND workflow_version = 2 AND idempotency_key = $2
	`, userID, key))
	if errors.Is(err, pgx.ErrNoRows) {
		return MatchRun{}, ErrSessionNotFound
	}
	return run, err
}

func (r *PostgresRepository) CreateMatchRun(ctx context.Context, run MatchRun) (MatchRun, bool, error) {
	inputSnapshot, err := json.Marshal(run.InputSnapshot)
	if err != nil {
		return MatchRun{}, false, err
	}
	profile, err := json.Marshal(nonNilMap(run.ParsedProfile))
	if err != nil {
		return MatchRun{}, false, err
	}
	fieldSources, err := json.Marshal(nonNilFieldSources(run.FieldSources))
	if err != nil {
		return MatchRun{}, false, err
	}
	questions, err := json.Marshal(nonNilQuestions(run.Questions))
	if err != nil {
		return MatchRun{}, false, err
	}
	missing, err := json.Marshal(nonNilStrings(run.MissingFields))
	if err != nil {
		return MatchRun{}, false, err
	}
	answerEvents, err := json.Marshal(nonNilAnswerEvents(run.AnswerEvents))
	if err != nil {
		return MatchRun{}, false, err
	}
	assumptions, err := json.Marshal(nonNilStrings(run.Assumptions))
	if err != nil {
		return MatchRun{}, false, err
	}
	err = r.db.QueryRow(ctx, `
		INSERT INTO project_match_sessions
			(user_id, workflow_version, intent, input_snapshot, parsed_profile, field_sources,
			 analysis_summary, completeness, missing_fields, clarification_questions,
			 question_count, rounds, answer_events, assumptions, revision, skipped_at,
			 status, idempotency_key, created_at, updated_at)
		VALUES ($1, 2, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $18)
		ON CONFLICT (user_id, idempotency_key) WHERE idempotency_key IS NOT NULL AND idempotency_key <> ''
		DO NOTHING
		RETURNING id
	`, run.UserID, run.Need, inputSnapshot, profile, fieldSources, run.AnalysisSummary, run.Completeness,
		missing, questions, run.QuestionCount, run.Rounds, answerEvents, assumptions, run.Revision, run.SkippedAt,
		run.Status, nullableString(run.IdempotencyKey), run.CreatedAt).Scan(&run.ID)
	if errors.Is(err, pgx.ErrNoRows) {
		existing, findErr := r.FindMatchRunByIdempotency(ctx, run.UserID, run.IdempotencyKey)
		if findErr != nil {
			return MatchRun{}, false, findErr
		}
		return existing, false, nil
	}
	return run, true, err
}

func (r *PostgresRepository) GetMatchRun(ctx context.Context, userID, id int64) (MatchRun, error) {
	run, err := scanMatchRun(r.db.QueryRow(ctx, `
		SELECT id, user_id, workflow_version, COALESCE(name, ''), intent, input_snapshot, parsed_profile,
		       field_sources, COALESCE(analysis_summary, ''), completeness, missing_fields,
		       clarification_questions, question_count, rounds, answer_events, assumptions,
		       revision, skipped_at, generation_attempt, progress_percent, current_step, result,
		       COALESCE(error_code, ''), cancelled_at, status, idempotency_key, created_at, updated_at
		FROM project_match_sessions
		WHERE user_id = $1 AND id = $2 AND workflow_version = 2
	`, userID, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return MatchRun{}, ErrSessionNotFound
	}
	return run, err
}

func (r *PostgresRepository) ListMatchRuns(ctx context.Context, userID int64, limit int) ([]MatchRun, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, user_id, workflow_version, COALESCE(name, ''), intent, input_snapshot, parsed_profile,
		       field_sources, COALESCE(analysis_summary, ''), completeness, missing_fields,
		       clarification_questions, question_count, rounds, answer_events, assumptions,
		       revision, skipped_at, generation_attempt, progress_percent, current_step, result,
		       COALESCE(error_code, ''), cancelled_at, status, idempotency_key, created_at, updated_at
		FROM project_match_sessions
		WHERE user_id = $1 AND workflow_version = 2
		ORDER BY created_at DESC, id DESC
		LIMIT $2
	`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	runs := make([]MatchRun, 0)
	for rows.Next() {
		run, scanErr := scanMatchRun(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		runs = append(runs, run)
	}
	return runs, rows.Err()
}

func (r *PostgresRepository) UpdateMatchRun(ctx context.Context, run MatchRun, expectedRevision int) (MatchRun, error) {
	inputSnapshot, err := json.Marshal(run.InputSnapshot)
	if err != nil {
		return MatchRun{}, err
	}
	profile, err := json.Marshal(nonNilMap(run.ParsedProfile))
	if err != nil {
		return MatchRun{}, err
	}
	fieldSources, err := json.Marshal(nonNilFieldSources(run.FieldSources))
	if err != nil {
		return MatchRun{}, err
	}
	questions, err := json.Marshal(nonNilQuestions(run.Questions))
	if err != nil {
		return MatchRun{}, err
	}
	missing, err := json.Marshal(nonNilStrings(run.MissingFields))
	if err != nil {
		return MatchRun{}, err
	}
	answerEvents, err := json.Marshal(nonNilAnswerEvents(run.AnswerEvents))
	if err != nil {
		return MatchRun{}, err
	}
	assumptions, err := json.Marshal(nonNilStrings(run.Assumptions))
	if err != nil {
		return MatchRun{}, err
	}
	updated, err := scanMatchRun(r.db.QueryRow(ctx, `
		UPDATE project_match_sessions
		SET input_snapshot = $3, parsed_profile = $4, field_sources = $5, analysis_summary = $6,
		    completeness = $7, missing_fields = $8, clarification_questions = $9,
		    question_count = $10, rounds = $11, answer_events = $12, assumptions = $13,
		    revision = revision + 1, skipped_at = $14, status = $15, updated_at = $16
		WHERE user_id = $1 AND id = $2 AND workflow_version = 2 AND revision = $17
		RETURNING id, user_id, workflow_version, COALESCE(name, ''), intent, input_snapshot, parsed_profile,
		          field_sources, COALESCE(analysis_summary, ''), completeness, missing_fields,
		          clarification_questions, question_count, rounds, answer_events, assumptions,
		          revision, skipped_at, generation_attempt, progress_percent, current_step, result,
		          COALESCE(error_code, ''), cancelled_at, status, idempotency_key, created_at, updated_at
	`, run.UserID, run.ID, inputSnapshot, profile, fieldSources, run.AnalysisSummary, run.Completeness,
		missing, questions, run.QuestionCount, run.Rounds, answerEvents, assumptions, run.SkippedAt,
		run.Status, run.UpdatedAt, expectedRevision))
	if errors.Is(err, pgx.ErrNoRows) {
		return MatchRun{}, ErrMatchRevisionConflict
	}
	return updated, err
}

func scanMatchRun(scanner sessionScanner) (MatchRun, error) {
	var run MatchRun
	var inputSnapshot, profile, fieldSources, missing, questions, answerEvents, assumptions, result []byte
	var key pgtype.Text
	if err := scanner.Scan(&run.ID, &run.UserID, &run.WorkflowVersion, &run.Name, &run.Need, &inputSnapshot, &profile,
		&fieldSources, &run.AnalysisSummary, &run.Completeness, &missing, &questions, &run.QuestionCount, &run.Rounds,
		&answerEvents, &assumptions, &run.Revision, &run.SkippedAt, &run.GenerationAttempt, &run.ProgressPercent,
		&run.CurrentStep, &result, &run.ErrorCode, &run.CanceledAt, &run.Status, &key, &run.CreatedAt, &run.UpdatedAt); err != nil {
		return MatchRun{}, err
	}
	if key.Valid {
		run.IdempotencyKey = key.String
	}
	if err := json.Unmarshal(inputSnapshot, &run.InputSnapshot); err != nil {
		return MatchRun{}, err
	}
	if err := json.Unmarshal(profile, &run.ParsedProfile); err != nil {
		return MatchRun{}, err
	}
	if err := json.Unmarshal(fieldSources, &run.FieldSources); err != nil {
		return MatchRun{}, err
	}
	if err := json.Unmarshal(missing, &run.MissingFields); err != nil {
		return MatchRun{}, err
	}
	if err := json.Unmarshal(questions, &run.Questions); err != nil {
		return MatchRun{}, err
	}
	if err := json.Unmarshal(answerEvents, &run.AnswerEvents); err != nil {
		return MatchRun{}, err
	}
	if err := json.Unmarshal(assumptions, &run.Assumptions); err != nil {
		return MatchRun{}, err
	}
	if err := json.Unmarshal(result, &run.Result); err != nil {
		return MatchRun{}, err
	}
	return run, nil
}

func nonNilMap(value map[string]any) map[string]any {
	if value == nil {
		return map[string]any{}
	}
	return value
}
func nonNilFieldSources(value map[string][]MatchFieldSource) map[string][]MatchFieldSource {
	if value == nil {
		return map[string][]MatchFieldSource{}
	}
	return value
}
func nonNilQuestions(value []ClarificationQuestion) []ClarificationQuestion {
	if value == nil {
		return []ClarificationQuestion{}
	}
	return value
}
func nonNilAnswerEvents(value []MatchAnswerEvent) []MatchAnswerEvent {
	if value == nil {
		return []MatchAnswerEvent{}
	}
	return value
}
func nonNilStrings(value []string) []string {
	if value == nil {
		return []string{}
	}
	return value
}
func nullableString(value string) any {
	if value == "" {
		return nil
	}
	return value
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

func scanProjectFile(scanner sessionScanner) (projectfiles.File, error) {
	var file projectfiles.File
	var matchID pgtype.Int8
	var extractedJSON []byte
	if err := scanner.Scan(
		&file.ID, &file.UserID, &matchID, &file.Name, &file.MIME, &file.DetectedMIME, &file.Size,
		&file.ObjectKey, &file.SHA256, &file.ParseStatus, &file.ExtractedText, &extractedJSON,
		&file.ErrorCode, &file.ExpiresAt, &file.CreatedAt, &file.UpdatedAt,
	); err != nil {
		return projectfiles.File{}, err
	}
	if matchID.Valid {
		value := matchID.Int64
		file.MatchID = &value
	}
	if len(extractedJSON) > 0 {
		if err := json.Unmarshal(extractedJSON, &file.ExtractedJSON); err != nil {
			return projectfiles.File{}, err
		}
	}
	return file, nil
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
