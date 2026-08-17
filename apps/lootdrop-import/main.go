package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type sourceRun struct {
	ID         int64
	Mode       string
	Status     string
	SourceURL  string
	StartedAt  time.Time
	FinishedAt *time.Time
	Counts     []byte
}

type rowMap map[string]any

var genericDatasets = []string{
	"crawl_runs", "dashboard_analytics", "database_explore_rows", "idea_vote_counts",
	"product_type_articles", "top10_lists", "top10_items", "failure_frameworks",
	"failure_patterns", "faq_items", "site_stats_snapshots", "sitemap_urls", "source_documents",
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	sourceDSN := flag.String("source-dsn", envOrDefault("LOOTDROP_MYSQL_DSN", "lootdrop:lootdrop@tcp(127.0.0.1:3307)/lootdrop?parseTime=true&charset=utf8mb4"), "Loot Drop MySQL DSN")
	targetURL := flag.String("database-url", envOrDefault("OPCV2_DATABASE_URL", "postgres://opcv2:opcv2@127.0.0.1:5432/opcv2?sslmode=disable"), "OPC PostgreSQL URL")
	flag.Parse()

	ctx := context.Background()
	source, err := sql.Open("mysql", *sourceDSN)
	if err != nil {
		logger.Error("open source mysql", "error", err)
		os.Exit(1)
	}
	defer source.Close()
	if err := source.PingContext(ctx); err != nil {
		logger.Error("ping source mysql", "error", err)
		os.Exit(1)
	}
	target, err := pgxpool.New(ctx, *targetURL)
	if err != nil {
		logger.Error("open target postgres", "error", err)
		os.Exit(1)
	}
	defer target.Close()
	if err := target.Ping(ctx); err != nil {
		logger.Error("ping target postgres", "error", err)
		os.Exit(1)
	}

	if err := importAll(ctx, source, target); err != nil {
		logger.Error("lootdrop import failed", "error", err)
		os.Exit(1)
	}
	logger.Info("lootdrop import completed")
}

func importAll(ctx context.Context, source *sql.DB, target *pgxpool.Pool) error {
	run, err := latestSourceRun(ctx, source)
	if err != nil {
		return err
	}
	tx, err := target.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	importRunID, err := createImportRun(ctx, tx, run)
	if err != nil {
		return err
	}
	if err := importStartups(ctx, source, tx, importRunID); err != nil {
		return fmt.Errorf("import startups: %w", err)
	}
	if err := importRebuildPlans(ctx, source, tx, importRunID); err != nil {
		return fmt.Errorf("import rebuild plans: %w", err)
	}
	if err := importIdeas(ctx, source, tx, importRunID); err != nil {
		return fmt.Errorf("import ideas: %w", err)
	}
	if _, err := tx.Exec(ctx, `SELECT refresh_lootdrop_project_opportunities()`); err != nil {
		return fmt.Errorf("sync project opportunity projection: %w", err)
	}
	if _, err := tx.Exec(ctx, `SELECT normalize_lootdrop_project_opportunities()`); err != nil {
		return fmt.Errorf("normalize project opportunity catalog: %w", err)
	}
	if _, err := tx.Exec(ctx, `SELECT index_lootdrop_project_search_content()`); err != nil {
		return fmt.Errorf("index project opportunity search content: %w", err)
	}
	if err := syncProjectCaseProjection(ctx, source, tx); err != nil {
		return fmt.Errorf("sync project case projection: %w", err)
	}
	for _, dataset := range genericDatasets {
		if err := importGenericDataset(ctx, source, tx, importRunID, dataset); err != nil {
			return fmt.Errorf("import %s: %w", dataset, err)
		}
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO lootdrop_import_state (singleton, active_run_id, last_source_run_id, last_source_finished_at, updated_at)
		VALUES (TRUE, $1, $2, $3, NOW())
		ON CONFLICT (singleton) DO UPDATE SET active_run_id = EXCLUDED.active_run_id,
			last_source_run_id = EXCLUDED.last_source_run_id,
			last_source_finished_at = EXCLUDED.last_source_finished_at,
			updated_at = NOW()
	`, importRunID, nullableInt64(run.ID), nullableTime(run.FinishedAt)); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `UPDATE lootdrop_sync_runs SET status = 'imported' WHERE id = $1`, importRunID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func latestSourceRun(ctx context.Context, db *sql.DB) (sourceRun, error) {
	var run sourceRun
	var finished sql.NullTime
	if err := db.QueryRowContext(ctx, `
		SELECT id, mode, status, source_site_url, started_at, finished_at, counts
		FROM crawl_runs ORDER BY id DESC LIMIT 1
	`).Scan(&run.ID, &run.Mode, &run.Status, &run.SourceURL, &run.StartedAt, &finished, &run.Counts); err != nil {
		return sourceRun{}, err
	}
	if finished.Valid {
		run.FinishedAt = &finished.Time
	}
	if !json.Valid(run.Counts) {
		run.Counts = []byte(`{}`)
	}
	return run, nil
}

func createImportRun(ctx context.Context, tx pgx.Tx, run sourceRun) (int64, error) {
	var id int64
	err := tx.QueryRow(ctx, `
		INSERT INTO lootdrop_sync_runs (
			source_run_id, source_mode, source_status, source_site_url,
			source_started_at, source_finished_at, source_counts, status
		) VALUES ($1, $2, $3, $4, $5, $6, $7, 'importing')
		RETURNING id
	`, run.ID, run.Mode, run.Status, run.SourceURL, run.StartedAt, nullableTime(run.FinishedAt), run.Counts).Scan(&id)
	return id, err
}

func importStartups(ctx context.Context, source *sql.DB, tx pgx.Tx, runID int64) error {
	rows, err := queryMaps(ctx, source, "startups")
	if err != nil {
		return err
	}
	ids := make([]int64, 0, len(rows))
	for _, row := range rows {
		payload, hash, err := payloadAndHash(row)
		if err != nil {
			return err
		}
		id, ok := int64Value(row["id"])
		if !ok || id <= 0 {
			continue
		}
		ids = append(ids, id)
		_, err = tx.Exec(ctx, `
			INSERT INTO lootdrop_startups (
				source_id, name, description, sector, country, founders, investors,
				start_year, end_year, total_funding, difficulty, difficulty_reason,
				scalability, scalability_reason, market_potential, market_potential_reason,
				cause_of_death, primary_cause_of_death, the_loot, market_analysis, pivot_idea,
				product_type, source_status, views, condensed_value_prop, condensed_cause_of_death,
				source_payload, source_hash, scraped_at, last_import_run_id, imported_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15,
				$16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30, NOW(), NOW())
			ON CONFLICT (source_id) DO UPDATE SET
				name = EXCLUDED.name, description = EXCLUDED.description, sector = EXCLUDED.sector,
				country = EXCLUDED.country, founders = EXCLUDED.founders, investors = EXCLUDED.investors,
				start_year = EXCLUDED.start_year, end_year = EXCLUDED.end_year, total_funding = EXCLUDED.total_funding,
				difficulty = EXCLUDED.difficulty, difficulty_reason = EXCLUDED.difficulty_reason,
				scalability = EXCLUDED.scalability, scalability_reason = EXCLUDED.scalability_reason,
				market_potential = EXCLUDED.market_potential, market_potential_reason = EXCLUDED.market_potential_reason,
				cause_of_death = EXCLUDED.cause_of_death, primary_cause_of_death = EXCLUDED.primary_cause_of_death,
				the_loot = EXCLUDED.the_loot, market_analysis = EXCLUDED.market_analysis, pivot_idea = EXCLUDED.pivot_idea,
				product_type = EXCLUDED.product_type, source_status = EXCLUDED.source_status, views = EXCLUDED.views,
				condensed_value_prop = EXCLUDED.condensed_value_prop, condensed_cause_of_death = EXCLUDED.condensed_cause_of_death,
				source_payload = EXCLUDED.source_payload, source_hash = EXCLUDED.source_hash, scraped_at = EXCLUDED.scraped_at,
				last_import_run_id = EXCLUDED.last_import_run_id, updated_at = NOW()
		`, id, stringValue(row["name"]), stringValue(row["description"]), stringValue(row["sector"]), stringValue(row["country"]),
			jsonColumn(row["founders"], `[]`), jsonColumn(row["investors"], `[]`), nullableInt(row["start_year"]), nullableInt(row["end_year"]),
			numericValue(row["total_funding"]), nullableInt(row["difficulty"]), stringValue(row["difficulty_reason"]), nullableInt(row["scalability"]),
			stringValue(row["scalability_reason"]), stringValue(row["market_potential"]), stringValue(row["market_potential_reason"]),
			stringValue(row["cause_of_death"]), stringValue(row["primary_cause_of_death"]), jsonColumn(row["the_loot"], `[]`),
			stringValue(row["market_analysis"]), jsonColumn(row["pivot_idea"], `{}`), stringValue(row["product_type"]), stringValue(row["status"]),
			nullableInt64(row["views"]), stringValue(row["condensed_value_prop"]), stringValue(row["condensed_cause_of_death"]),
			payload, hash, nullableTimeValue(row["scraped_at"]), runID)
		if err != nil {
			return err
		}
		if err := ensureTranslation(ctx, tx, "startup", strconv.FormatInt(id, 10), hash); err != nil {
			return err
		}
	}
	return deleteMissingIDs(ctx, tx, "lootdrop_startups", ids)
}

func importRebuildPlans(ctx context.Context, source *sql.DB, tx pgx.Tx, runID int64) error {
	rows, err := queryMaps(ctx, source, "rebuild_plans")
	if err != nil {
		return err
	}
	ids := make([]int64, 0, len(rows))
	for _, row := range rows {
		payload, hash, err := payloadAndHash(row)
		if err != nil {
			return err
		}
		id, ok := int64Value(row["startup_id"])
		if !ok || id <= 0 {
			continue
		}
		ids = append(ids, id)
		_, err = tx.Exec(ctx, `
			INSERT INTO lootdrop_rebuild_plans (
				source_id, name, sector, product_type, country, total_funding, primary_cause_of_death,
				market_potential, scalability, difficulty, pivot_idea, the_loot, source_page,
				source_payload, source_hash, scraped_at, last_import_run_id, imported_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, NOW(), NOW())
			ON CONFLICT (source_id) DO UPDATE SET
				name = EXCLUDED.name, sector = EXCLUDED.sector, product_type = EXCLUDED.product_type,
				country = EXCLUDED.country, total_funding = EXCLUDED.total_funding,
				primary_cause_of_death = EXCLUDED.primary_cause_of_death, market_potential = EXCLUDED.market_potential,
				scalability = EXCLUDED.scalability, difficulty = EXCLUDED.difficulty, pivot_idea = EXCLUDED.pivot_idea,
				the_loot = EXCLUDED.the_loot, source_page = EXCLUDED.source_page, source_payload = EXCLUDED.source_payload,
				source_hash = EXCLUDED.source_hash, scraped_at = EXCLUDED.scraped_at, last_import_run_id = EXCLUDED.last_import_run_id,
				updated_at = NOW()
		`, id, stringValue(row["name"]), stringValue(row["sector"]), stringValue(row["product_type"]), stringValue(row["country"]),
			numericValue(row["total_funding"]), stringValue(row["primary_cause_of_death"]), stringValue(row["market_potential"]),
			nullableInt(row["scalability"]), nullableInt(row["difficulty"]), jsonColumn(row["pivot_idea"], `{}`), jsonColumn(row["the_loot"], `[]`),
			nullableInt(row["source_page"]), payload, hash, nullableTimeValue(row["scraped_at"]), runID)
		if err != nil {
			return err
		}
		if err := ensureTranslation(ctx, tx, "rebuild_plan", strconv.FormatInt(id, 10), hash); err != nil {
			return err
		}
	}
	return deleteMissingIDs(ctx, tx, "lootdrop_rebuild_plans", ids)
}

func importIdeas(ctx context.Context, source *sql.DB, tx pgx.Tx, runID int64) error {
	rows, err := queryMaps(ctx, source, "ideas")
	if err != nil {
		return err
	}
	ids := make([]int64, 0, len(rows))
	for _, row := range rows {
		payload, hash, err := payloadAndHash(row)
		if err != nil {
			return err
		}
		id, ok := int64Value(row["idea_id"])
		if !ok || id <= 0 {
			continue
		}
		ids = append(ids, id)
		_, err = tx.Exec(ctx, `
			INSERT INTO lootdrop_ideas (
				source_id, title, description, model, effort, speed, category, tags, persona,
				source_payload, source_hash, scraped_at, last_import_run_id, imported_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, NOW(), NOW())
			ON CONFLICT (source_id) DO UPDATE SET
				title = EXCLUDED.title, description = EXCLUDED.description, model = EXCLUDED.model,
				effort = EXCLUDED.effort, speed = EXCLUDED.speed, category = EXCLUDED.category,
				tags = EXCLUDED.tags, persona = EXCLUDED.persona, source_payload = EXCLUDED.source_payload,
				source_hash = EXCLUDED.source_hash, scraped_at = EXCLUDED.scraped_at, last_import_run_id = EXCLUDED.last_import_run_id,
				updated_at = NOW()
		`, id, stringValue(row["title"]), stringValue(row["description"]), stringValue(row["model"]), stringValue(row["effort"]),
			stringValue(row["speed"]), stringValue(row["category"]), jsonColumn(row["tags"], `[]`), stringValue(row["persona"]),
			payload, hash, nullableTimeValue(row["scraped_at"]), runID)
		if err != nil {
			return err
		}
		if err := ensureTranslation(ctx, tx, "idea", strconv.FormatInt(id, 10), hash); err != nil {
			return err
		}
	}
	return deleteMissingIDs(ctx, tx, "lootdrop_ideas", ids)
}

func importGenericDataset(ctx context.Context, source *sql.DB, tx pgx.Tx, runID int64, dataset string) error {
	rows, err := queryMaps(ctx, source, dataset)
	if err != nil {
		return err
	}
	keys := make([]string, 0, len(rows))
	for _, row := range rows {
		key := recordKey(dataset, row)
		if key == "" {
			continue
		}
		keys = append(keys, key)
		payload, hash, err := payloadAndHash(row)
		if err != nil {
			return err
		}
		_, err = tx.Exec(ctx, `
			INSERT INTO lootdrop_dataset_rows (
				dataset, record_key, payload, raw_text, source_url, content_type, http_status,
				source_hash, source_fetched_at, source_row_count, scraped_at, last_import_run_id,
				imported_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, NOW(), NOW())
			ON CONFLICT (dataset, record_key) DO UPDATE SET
				payload = EXCLUDED.payload, raw_text = EXCLUDED.raw_text, source_url = EXCLUDED.source_url,
				content_type = EXCLUDED.content_type, http_status = EXCLUDED.http_status, source_hash = EXCLUDED.source_hash,
				source_fetched_at = EXCLUDED.source_fetched_at, source_row_count = EXCLUDED.source_row_count,
				scraped_at = EXCLUDED.scraped_at, last_import_run_id = EXCLUDED.last_import_run_id, updated_at = NOW()
		`, dataset, key, payload, stringValue(row["raw_text"]), firstString(row, "url", "source_url", "loc"), stringValue(row["content_type"]),
			nullableInt(row["http_status"]), hash, firstTime(row, "fetched_at", "captured_at", "last_seen_at"), nullableInt(row["row_count"]),
			firstTime(row, "scraped_at", "last_seen_at", "captured_at", "fetched_at"), runID)
		if err != nil {
			return err
		}
	}
	if len(keys) == 0 {
		_, err = tx.Exec(ctx, `DELETE FROM lootdrop_dataset_rows WHERE dataset = $1`, dataset)
		return err
	}
	_, err = tx.Exec(ctx, `DELETE FROM lootdrop_dataset_rows WHERE dataset = $1 AND NOT (record_key = ANY($2::text[]))`, dataset, keys)
	return err
}

func syncProjectCaseClaims(ctx context.Context, tx pgx.Tx, caseID, sourceID int64, row rowMap, translated map[string]any) error {
	if _, err := tx.Exec(ctx, `DELETE FROM project_case_claims WHERE case_id = $1`, caseID); err != nil {
		return err
	}
	type claim struct {
		kind, field, value, detail string
		modelGenerated             bool
	}
	claims := []claim{
		{kind: "fact", field: "company_name", value: translatedString(translated, "name_zh", stringValue(row["name"]))},
		{kind: "fact", field: "description", value: translatedString(translated, "description_zh", stringValue(row["description"]))},
		{kind: "fact", field: "sector", value: translatedString(translated, "sector_zh", stringValue(row["sector"]))},
		{kind: "fact", field: "country", value: translatedString(translated, "country_zh", stringValue(row["country"]))},
		{kind: "fact", field: "start_year", value: stringValue(row["start_year"])},
		{kind: "fact", field: "end_year", value: stringValue(row["end_year"])},
		{kind: "fact", field: "total_funding", value: stringValue(row["total_funding"])},
		{kind: "fact", field: "primary_cause_of_death", value: translatedString(translated, "primary_cause_of_death_zh", stringValue(row["primary_cause_of_death"]))},
		{kind: "fact", field: "market_potential", value: stringValue(row["market_potential"])},
		{kind: "analysis", field: "cause_of_death", value: translatedString(translated, "cause_of_death_zh", stringValue(row["cause_of_death"])), modelGenerated: true},
		{kind: "analysis", field: "market_analysis", value: translatedString(translated, "market_analysis_zh", stringValue(row["market_analysis"])), modelGenerated: true},
		{kind: "analysis", field: "pivot_idea", value: string(jsonColumn(row["pivot_idea"], `{}`)), modelGenerated: true},
	}
	sortOrder := 0
	for _, item := range claims {
		if strings.TrimSpace(item.value) == "" {
			continue
		}
		var claimID int64
		if err := tx.QueryRow(ctx, `
			INSERT INTO project_case_claims (case_id, claim_type, field_name, value_text, detail, is_model_generated, sort_order)
			VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id
		`, caseID, item.kind, item.field, item.value, item.detail, item.modelGenerated, sortOrder).Scan(&claimID); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO project_case_claim_sources (claim_id, web_source_id)
			VALUES ($1, $2) ON CONFLICT DO NOTHING
		`, claimID, sourceID); err != nil {
			return err
		}
		sortOrder++
	}
	return nil
}

func startupTranslation(ctx context.Context, tx pgx.Tx, id int64) map[string]any {
	var payload []byte
	if err := tx.QueryRow(ctx, `
		SELECT translated_payload FROM lootdrop_translations
		WHERE dataset = 'startup' AND record_key = $1 AND language = 'zh-CN' AND status = 'completed'
	`, strconv.FormatInt(id, 10)).Scan(&payload); err != nil || !json.Valid(payload) {
		return nil
	}
	var translated map[string]any
	if json.Unmarshal(payload, &translated) != nil {
		return nil
	}
	return translated
}

func translatedString(payload map[string]any, key, fallback string) string {
	if payload != nil {
		if value, ok := payload[key].(string); ok && strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return fallback
}

func deleteMissingIDs(ctx context.Context, tx pgx.Tx, table string, ids []int64) error {
	if len(ids) == 0 {
		_, err := tx.Exec(ctx, `DELETE FROM `+table)
		return err
	}
	_, err := tx.Exec(ctx, `DELETE FROM `+table+` WHERE NOT (source_id = ANY($1::bigint[]))`, ids)
	return err
}

// project_cases is a read projection for the existing project-market case UI.
// lootdrop_* remains the canonical mirror and is re-projected on every import.
func syncProjectCaseProjection(ctx context.Context, source *sql.DB, tx pgx.Tx) error {
	rows, err := queryMaps(ctx, source, "startups")
	if err != nil {
		return err
	}
	var jobID int64
	err = tx.QueryRow(ctx, `
		INSERT INTO web_research_jobs (scene, ref_type, trigger_reason, status, finished_at, updated_at)
		VALUES ('current_data', 'lootdrop_startup', 'lootdrop_sync', 'done', NOW(), NOW())
		RETURNING id
	`).Scan(&jobID)
	if err != nil {
		return err
	}
	ids := make([]int64, 0, len(rows))
	for _, row := range rows {
		id, ok := int64Value(row["id"])
		if !ok || id <= 0 {
			continue
		}
		ids = append(ids, id)
		name := stringValue(row["name"])
		description := stringValue(row["description"])
		cause := stringValue(row["cause_of_death"])
		if cause == "" {
			cause = stringValue(row["primary_cause_of_death"])
		}
		translated := startupTranslation(ctx, tx, id)
		name = translatedString(translated, "name_zh", name)
		description = translatedString(translated, "description_zh", description)
		cause = translatedString(translated, "cause_of_death_zh", cause)
		slug := "lootdrop-startup-" + strconv.FormatInt(id, 10)
		sourceURL := "https://www.loot-drop.io/database-view?id=" + strconv.FormatInt(id, 10)
		capturedAt := nullableTimeValue(row["scraped_at"])
		if capturedAt == nil {
			capturedAt = time.Now().UTC()
		}
		content := strings.TrimSpace(strings.Join([]string{name, description, cause}, "\n\n"))
		_, err = tx.Exec(ctx, `
			INSERT INTO project_cases (
				slug, title, summary, result_summary, case_type, outcome, key_actions, lessons, pitfalls,
				source_title, source_url, captured_at, status, published_at,
				evidence_status, last_verified_at, is_locked, has_conflict, content_md, updated_at
			) VALUES ($1, $2, $3, $4, 'failure', $4, $5, $6, $7, 'Loot Drop', $8, $9, 'published', NOW(), 'verified', $9, TRUE, FALSE, $10, NOW())
			ON CONFLICT (slug) DO UPDATE SET
				title = EXCLUDED.title, summary = EXCLUDED.summary, outcome = EXCLUDED.outcome,
				key_actions = EXCLUDED.key_actions, lessons = EXCLUDED.lessons, pitfalls = EXCLUDED.pitfalls,
				source_title = EXCLUDED.source_title, source_url = EXCLUDED.source_url,
				captured_at = EXCLUDED.captured_at, status = EXCLUDED.status, published_at = EXCLUDED.published_at,
				evidence_status = EXCLUDED.evidence_status, last_verified_at = EXCLUDED.last_verified_at,
				is_locked = EXCLUDED.is_locked, has_conflict = EXCLUDED.has_conflict, content_md = EXCLUDED.content_md,
				updated_at = NOW()
		`, slug, name, description, cause, jsonColumn(row["the_loot"], `[]`), jsonColumn(row["the_loot"], `[]`),
			jsonColumn(row["primary_cause_of_death"], `[]`), sourceURL, capturedAt, content)
		if err != nil {
			return err
		}
		var caseID int64
		if err := tx.QueryRow(ctx, `SELECT id FROM project_cases WHERE slug = $1`, slug).Scan(&caseID); err != nil {
			return err
		}
		var sourceID int64
		err = tx.QueryRow(ctx, `
			INSERT INTO web_sources (job_id, url, canonical_url, title, publisher, source_kind, quality_score, fetched_at, evidence_excerpt)
			VALUES ($1, $2, $2, $3, 'Loot Drop', 'primary', 0.95, $4, $5)
			ON CONFLICT (job_id, canonical_url) DO UPDATE SET title = EXCLUDED.title, fetched_at = EXCLUDED.fetched_at,
				evidence_excerpt = EXCLUDED.evidence_excerpt
			RETURNING id
		`, jobID, sourceURL, name, capturedAt, cause).Scan(&sourceID)
		if err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO project_case_sources (case_id, web_source_id, field_name, is_primary)
			VALUES ($1, $2, 'source', TRUE)
			ON CONFLICT (case_id, web_source_id, field_name) DO UPDATE SET is_primary = TRUE
			`, caseID, sourceID); err != nil {
			return err
		}
		if err := syncProjectCaseClaims(ctx, tx, caseID, sourceID, row, translated); err != nil {
			return err
		}
	}
	if len(ids) == 0 {
		_, err = tx.Exec(ctx, `DELETE FROM project_cases WHERE slug LIKE 'lootdrop-startup-%'`)
		return err
	}
	_, err = tx.Exec(ctx, `DELETE FROM project_cases WHERE slug LIKE 'lootdrop-startup-%' AND substring(slug from '[0-9]+$')::bigint <> ALL($1::bigint[])`, ids)
	return err
}

func ensureTranslation(ctx context.Context, tx pgx.Tx, dataset, key, hash string) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO lootdrop_translations (dataset, record_key, language, status, source_hash, updated_at)
		VALUES ($1, $2, 'zh-CN', 'pending', $3, NOW())
		ON CONFLICT (dataset, record_key, language) DO UPDATE SET
			status = CASE WHEN lootdrop_translations.source_hash IS DISTINCT FROM EXCLUDED.source_hash THEN 'pending' ELSE lootdrop_translations.status END,
			source_hash = EXCLUDED.source_hash,
			error_message = CASE WHEN lootdrop_translations.source_hash IS DISTINCT FROM EXCLUDED.source_hash THEN '' ELSE lootdrop_translations.error_message END,
			updated_at = NOW()
	`, dataset, key, hash)
	return err
}

func queryMaps(ctx context.Context, db *sql.DB, table string) ([]rowMap, error) {
	rows, err := db.QueryContext(ctx, "SELECT * FROM `"+table+"`")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	result := make([]rowMap, 0)
	for rows.Next() {
		values := make([]any, len(columns))
		pointers := make([]any, len(columns))
		for index := range values {
			pointers[index] = &values[index]
		}
		if err := rows.Scan(pointers...); err != nil {
			return nil, err
		}
		row := rowMap{}
		for index, column := range columns {
			row[column] = normalizeValue(values[index])
		}
		result = append(result, row)
	}
	return result, rows.Err()
}

func recordKey(dataset string, row rowMap) string {
	switch dataset {
	case "crawl_runs", "site_stats_snapshots", "top10_lists", "failure_frameworks", "failure_patterns", "faq_items", "product_type_articles", "sitemap_urls", "source_documents":
		for _, key := range []string{"id", "list_key", "framework_key", "pattern_id", "faq_key", "slug", "loc", "source_key"} {
			if value := stringValue(row[key]); value != "" {
				return value
			}
		}
	case "top10_items":
		rank, _ := int64Value(row["rank_no"])
		return firstString(row, "list_key") + ":" + strconv.FormatInt(rank, 10)
	default:
		for _, key := range []string{"startup_id", "idea_id", "id"} {
			if value := stringValue(row[key]); value != "" {
				return value
			}
		}
	}
	return ""
}

func payloadAndHash(row rowMap) ([]byte, string, error) {
	payload, err := json.Marshal(row)
	if err != nil {
		return nil, "", err
	}
	digest := sha256.Sum256(payload)
	return payload, hex.EncodeToString(digest[:]), nil
}

func normalizeValue(value any) any {
	switch typed := value.(type) {
	case []byte:
		trimmed := strings.TrimSpace(string(typed))
		if json.Valid(typed) {
			var decoded any
			if json.Unmarshal(typed, &decoded) == nil {
				return decoded
			}
		}
		return trimmed
	default:
		return value
	}
}

func jsonColumn(value any, fallback string) []byte {
	if value == nil {
		return []byte(fallback)
	}
	if raw, ok := value.(json.RawMessage); ok && json.Valid(raw) {
		return raw
	}
	encoded, err := json.Marshal(value)
	if err != nil || !json.Valid(encoded) {
		return []byte(fallback)
	}
	return encoded
}

func stringValue(value any) string {
	switch typed := value.(type) {
	case nil:
		return ""
	case string:
		return typed
	case []byte:
		return string(typed)
	case json.Number:
		return typed.String()
	default:
		return fmt.Sprint(typed)
	}
}

func int64Value(value any) (int64, bool) {
	text := strings.TrimSpace(stringValue(value))
	if text == "" {
		return 0, false
	}
	parsed, err := strconv.ParseInt(text, 10, 64)
	return parsed, err == nil
}

func nullableInt(value any) any {
	parsed, ok := int64Value(value)
	if !ok {
		return nil
	}
	return parsed
}

func nullableInt64(value any) any { return nullableInt(value) }

func numericValue(value any) any {
	text := strings.TrimSpace(stringValue(value))
	if text == "" {
		return nil
	}
	return text
}

func nullableTimeValue(value any) any {
	switch typed := value.(type) {
	case nil:
		return nil
	case time.Time:
		return typed
	case string:
		for _, layout := range []string{"2006-01-02 15:04:05.999999", "2006-01-02 15:04:05", time.RFC3339Nano} {
			if parsed, err := time.ParseInLocation(layout, typed, time.UTC); err == nil {
				return parsed
			}
		}
	}
	return nil
}

func nullableTime(value *time.Time) any {
	if value == nil {
		return nil
	}
	return *value
}

func firstString(row rowMap, keys ...string) string {
	for _, key := range keys {
		if value := stringValue(row[key]); value != "" {
			return value
		}
	}
	return ""
}

func firstTime(row rowMap, keys ...string) any {
	for _, key := range keys {
		if value := nullableTimeValue(row[key]); value != nil {
			return value
		}
	}
	return nil
}

func envOrDefault(name, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		return value
	}
	return fallback
}
