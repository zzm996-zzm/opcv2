DROP INDEX IF EXISTS project_match_sessions_user_idempotency_idx;

ALTER TABLE project_match_sessions
    DROP COLUMN IF EXISTS error_code,
    DROP COLUMN IF EXISTS cancelled_at,
    DROP COLUMN IF EXISTS idempotency_key,
    DROP COLUMN IF EXISTS research_job_id,
    DROP COLUMN IF EXISTS kb_sufficiency,
    DROP COLUMN IF EXISTS question_count,
    DROP COLUMN IF EXISTS rounds,
    DROP COLUMN IF EXISTS completeness,
    DROP COLUMN IF EXISTS assumptions,
    DROP COLUMN IF EXISTS analysis_summary,
    DROP COLUMN IF EXISTS field_sources,
    DROP COLUMN IF EXISTS parsed_profile,
    DROP COLUMN IF EXISTS input_snapshot,
    DROP COLUMN IF EXISTS name;

DROP TABLE IF EXISTS content_corrections;

ALTER TABLE project_comparisons DROP CONSTRAINT IF EXISTS project_comparisons_items_check;
ALTER TABLE project_comparisons
    ADD CONSTRAINT project_comparisons_items_check
    CHECK (jsonb_typeof(items) = 'array' AND jsonb_array_length(items) BETWEEN 2 AND 4);

DROP TABLE IF EXISTS project_compare_items;
DROP TABLE IF EXISTS project_views;
DROP TABLE IF EXISTS project_favorites;
DROP TABLE IF EXISTS project_case_sources;

ALTER TABLE project_cases
    DROP COLUMN IF EXISTS is_locked,
    DROP COLUMN IF EXISTS last_verified_at,
    DROP COLUMN IF EXISTS evidence_status,
    DROP COLUMN IF EXISTS content_md,
    DROP COLUMN IF EXISTS result_summary,
    DROP COLUMN IF EXISTS scale,
    DROP COLUMN IF EXISTS cover_url;

DROP TABLE IF EXISTS opportunity_item_sources;
DROP TABLE IF EXISTS opportunity_items;

DROP INDEX IF EXISTS project_opportunities_catalog_idx;

ALTER TABLE project_opportunities DROP CONSTRAINT IF EXISTS project_opportunities_status_check;
ALTER TABLE project_opportunities
    ADD CONSTRAINT project_opportunities_status_check CHECK (status IN ('draft', 'published'));

ALTER TABLE project_opportunities
    DROP COLUMN IF EXISTS detail,
    DROP COLUMN IF EXISTS rebuild_plan_id,
    DROP COLUMN IF EXISTS source_failure_id,
    DROP COLUMN IF EXISTS is_featured,
    DROP COLUMN IF EXISTS heat,
    DROP COLUMN IF EXISTS primary_source_url,
    DROP COLUMN IF EXISTS is_real,
    DROP COLUMN IF EXISTS revenue_range,
    DROP COLUMN IF EXISTS invest_cents,
    DROP COLUMN IF EXISTS track_code,
    DROP COLUMN IF EXISTS category_code,
    DROP COLUMN IF EXISTS cover_url;

DROP TABLE IF EXISTS web_sources;
DROP TABLE IF EXISTS web_research_jobs;
DROP TABLE IF EXISTS rebuild_localizations;
DROP TABLE IF EXISTS rebuild_plans;
DROP TABLE IF EXISTS startup_failure_antipatterns;
DROP TABLE IF EXISTS startup_failure_sources;
DROP TABLE IF EXISTS startup_failures;
DROP TABLE IF EXISTS project_import_batches;
DROP TABLE IF EXISTS project_dict_items;
