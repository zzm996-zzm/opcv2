CREATE TABLE project_dict_items (
    code TEXT PRIMARY KEY,
    kind TEXT NOT NULL CHECK (kind IN ('category', 'sector', 'product_type', 'antipattern', 'country', 'cn_channel')),
    name_zh TEXT NOT NULL CHECK (name_zh <> ''),
    name_en TEXT,
    sort_order INTEGER NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX project_dict_items_kind_sort_idx
    ON project_dict_items (kind, sort_order, code)
    WHERE is_active = TRUE;

CREATE TABLE project_import_batches (
    id BIGSERIAL PRIMARY KEY,
    batch_no TEXT NOT NULL UNIQUE CHECK (batch_no <> ''),
    planned_count INTEGER NOT NULL CHECK (planned_count >= 0),
    succeeded_count INTEGER NOT NULL DEFAULT 0 CHECK (succeeded_count >= 0),
    status TEXT NOT NULL DEFAULT 'collecting'
        CHECK (status IN ('collecting', 'ai_enriching', 'reviewing', 'published', 'rolled_back')),
    operator_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE startup_failures (
    id BIGSERIAL PRIMARY KEY,
    slug TEXT NOT NULL UNIQUE CHECK (slug <> ''),
    name TEXT NOT NULL CHECK (name <> ''),
    name_zh TEXT,
    country_code TEXT REFERENCES project_dict_items(code),
    sector_code TEXT REFERENCES project_dict_items(code),
    product_type_code TEXT REFERENCES project_dict_items(code),
    burned_cents BIGINT CHECK (burned_cents IS NULL OR burned_cents >= 0),
    currency CHAR(3) NOT NULL DEFAULT 'USD',
    founded_year INTEGER,
    died_year INTEGER,
    value_prop_zh TEXT,
    death_cause_zh TEXT,
    failure_analysis_zh TEXT,
    learnings_zh JSONB NOT NULL DEFAULT '[]'::JSONB CHECK (jsonb_typeof(learnings_zh) = 'array'),
    is_model_generated BOOLEAN NOT NULL DEFAULT TRUE,
    review_status TEXT NOT NULL DEFAULT 'draft'
        CHECK (review_status IN ('draft', 'ai_enriched', 'human_reviewed', 'published', 'rejected')),
    batch_id BIGINT REFERENCES project_import_batches(id) ON DELETE SET NULL,
    reviewed_by BIGINT REFERENCES users(id) ON DELETE SET NULL,
    reviewed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (founded_year IS NULL OR died_year IS NULL OR died_year >= founded_year)
);

CREATE INDEX startup_failures_filter_idx
    ON startup_failures (sector_code, product_type_code, died_year, review_status);

CREATE TABLE startup_failure_sources (
    id BIGSERIAL PRIMARY KEY,
    failure_id BIGINT NOT NULL REFERENCES startup_failures(id) ON DELETE CASCADE,
    field_name TEXT NOT NULL CHECK (field_name <> ''),
    source_url TEXT NOT NULL CHECK (source_url ~* '^https?://'),
    source_name TEXT,
    source_kind TEXT NOT NULL DEFAULT 'secondary'
        CHECK (source_kind IN ('primary', 'authority', 'research', 'media', 'vertical', 'community')),
    published_at TIMESTAMPTZ,
    fetched_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    evidence_excerpt TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX startup_failure_sources_failure_field_idx
    ON startup_failure_sources (failure_id, field_name);

CREATE TABLE startup_failure_antipatterns (
    failure_id BIGINT NOT NULL REFERENCES startup_failures(id) ON DELETE CASCADE,
    antipattern_code TEXT NOT NULL REFERENCES project_dict_items(code),
    weight NUMERIC(3,2) NOT NULL DEFAULT 1.0 CHECK (weight >= 0 AND weight <= 1),
    PRIMARY KEY (failure_id, antipattern_code)
);

CREATE TABLE rebuild_plans (
    id BIGSERIAL PRIMARY KEY,
    failure_id BIGINT NOT NULL UNIQUE REFERENCES startup_failures(id) ON DELETE CASCADE,
    concept_zh TEXT NOT NULL CHECK (concept_zh <> ''),
    market_potential TEXT CHECK (market_potential IN ('high', 'mid', 'low')),
    difficulty INTEGER CHECK (difficulty BETWEEN 1 AND 5),
    scalability INTEGER CHECK (scalability BETWEEN 1 AND 5),
    rebuild_score INTEGER CHECK (rebuild_score BETWEEN 0 AND 100),
    execution_phases JSONB NOT NULL DEFAULT '[]'::JSONB CHECK (jsonb_typeof(execution_phases) = 'array'),
    tech_stack_zh JSONB NOT NULL DEFAULT '[]'::JSONB CHECK (jsonb_typeof(tech_stack_zh) = 'array'),
    monetization_zh TEXT,
    model_name TEXT,
    prompt_version TEXT,
    is_model_generated BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE rebuild_localizations (
    plan_id BIGINT PRIMARY KEY REFERENCES rebuild_plans(id) ON DELETE CASCADE,
    solo_version TEXT NOT NULL CHECK (solo_version <> ''),
    team_size_min INTEGER NOT NULL DEFAULT 1 CHECK (team_size_min >= 1),
    cn_budget_cents BIGINT NOT NULL CHECK (cn_budget_cents >= 0),
    cn_budget_band TEXT NOT NULL CHECK (cn_budget_band IN ('0-5k', '5k-2w', '2w-10w', '10w+')),
    cn_channels JSONB NOT NULL DEFAULT '[]'::JSONB CHECK (jsonb_typeof(cn_channels) = 'array'),
    cn_compliance TEXT,
    time_to_first_revenue TEXT,
    is_model_generated BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE web_research_jobs (
    id BIGSERIAL PRIMARY KEY,
    scene TEXT NOT NULL CHECK (scene IN ('match', 'explore', 'case', 'current_data')),
    ref_type TEXT,
    ref_id BIGINT,
    trigger_reason TEXT,
    query_plan JSONB NOT NULL DEFAULT '[]'::JSONB CHECK (jsonb_typeof(query_plan) = 'array'),
    missing_evidence JSONB NOT NULL DEFAULT '[]'::JSONB CHECK (jsonb_typeof(missing_evidence) = 'array'),
    status TEXT NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'running', 'partial', 'done', 'failed', 'cancelled')),
    error_code TEXT,
    started_at TIMESTAMPTZ,
    finished_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX web_research_jobs_scene_status_idx
    ON web_research_jobs (scene, status, created_at DESC);

CREATE TABLE web_sources (
    id BIGSERIAL PRIMARY KEY,
    job_id BIGINT NOT NULL REFERENCES web_research_jobs(id) ON DELETE CASCADE,
    url TEXT NOT NULL CHECK (url ~* '^https?://'),
    canonical_url TEXT NOT NULL CHECK (canonical_url ~* '^https?://'),
    title TEXT,
    publisher TEXT,
    source_kind TEXT NOT NULL DEFAULT 'secondary'
        CHECK (source_kind IN ('primary', 'authority', 'research', 'media', 'vertical', 'community')),
    quality_score NUMERIC(3,2) CHECK (quality_score BETWEEN 0 AND 1),
    published_at TIMESTAMPTZ,
    fetched_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    evidence_excerpt TEXT,
    extracted_facts JSONB NOT NULL DEFAULT '[]'::JSONB CHECK (jsonb_typeof(extracted_facts) = 'array'),
    content_hash CHAR(64),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (job_id, canonical_url)
);

CREATE INDEX web_sources_job_quality_idx
    ON web_sources (job_id, quality_score DESC, id);

ALTER TABLE project_opportunities
    ADD COLUMN cover_url TEXT,
    ADD COLUMN category_code TEXT REFERENCES project_dict_items(code),
    ADD COLUMN track_code TEXT REFERENCES project_dict_items(code),
    ADD COLUMN invest_cents BIGINT CHECK (invest_cents IS NULL OR invest_cents >= 0),
    ADD COLUMN revenue_range TEXT,
    ADD COLUMN is_real BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN primary_source_url TEXT,
    ADD COLUMN heat INTEGER NOT NULL DEFAULT 0 CHECK (heat >= 0),
    ADD COLUMN is_featured BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN source_failure_id BIGINT REFERENCES startup_failures(id) ON DELETE SET NULL,
    ADD COLUMN rebuild_plan_id BIGINT REFERENCES rebuild_plans(id) ON DELETE SET NULL,
    ADD COLUMN detail JSONB NOT NULL DEFAULT '{}'::JSONB CHECK (jsonb_typeof(detail) = 'object');

ALTER TABLE project_opportunities DROP CONSTRAINT IF EXISTS project_opportunities_status_check;
ALTER TABLE project_opportunities
    ADD CONSTRAINT project_opportunities_status_check CHECK (status IN ('draft', 'published', 'offline'));

CREATE INDEX project_opportunities_catalog_idx
    ON project_opportunities (is_featured DESC, heat DESC, published_at DESC, id)
    WHERE status = 'published';

CREATE TABLE opportunity_items (
    id BIGSERIAL PRIMARY KEY,
    project_id BIGINT REFERENCES project_opportunities(id) ON DELETE SET NULL,
    seed_failure_id BIGINT REFERENCES startup_failures(id) ON DELETE SET NULL,
    group_codes JSONB NOT NULL DEFAULT '[]'::JSONB CHECK (jsonb_typeof(group_codes) = 'array'),
    title TEXT NOT NULL CHECK (title <> ''),
    summary TEXT NOT NULL,
    opportunity_reason TEXT NOT NULL,
    why_now TEXT,
    target_user TEXT,
    budget_band TEXT,
    difficulty TEXT,
    risks JSONB NOT NULL DEFAULT '[]'::JSONB CHECK (jsonb_typeof(risks) = 'array'),
    status TEXT NOT NULL DEFAULT 'draft'
        CHECK (status IN ('draft', 'researching', 'reviewing', 'published', 'rejected', 'stale')),
    verified_at TIMESTAMPTZ,
    published_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX opportunity_items_published_idx
    ON opportunity_items (verified_at DESC, id)
    WHERE status = 'published';

CREATE TABLE opportunity_item_sources (
    opportunity_id BIGINT NOT NULL REFERENCES opportunity_items(id) ON DELETE CASCADE,
    web_source_id BIGINT NOT NULL REFERENCES web_sources(id) ON DELETE CASCADE,
    claim_key TEXT NOT NULL CHECK (claim_key <> ''),
    PRIMARY KEY (opportunity_id, web_source_id, claim_key)
);

ALTER TABLE project_cases
    ADD COLUMN cover_url TEXT,
    ADD COLUMN scale TEXT,
    ADD COLUMN result_summary TEXT,
    ADD COLUMN content_md TEXT,
    ADD COLUMN evidence_status TEXT NOT NULL DEFAULT 'draft'
        CHECK (evidence_status IN ('draft', 'researching', 'needs_review', 'verified', 'stale', 'rejected')),
    ADD COLUMN last_verified_at TIMESTAMPTZ,
    ADD COLUMN is_locked BOOLEAN NOT NULL DEFAULT FALSE;

UPDATE project_cases SET result_summary = outcome WHERE result_summary IS NULL;
ALTER TABLE project_cases ALTER COLUMN result_summary SET NOT NULL;

CREATE TABLE project_case_sources (
    case_id BIGINT NOT NULL REFERENCES project_cases(id) ON DELETE CASCADE,
    web_source_id BIGINT NOT NULL REFERENCES web_sources(id) ON DELETE CASCADE,
    field_name TEXT NOT NULL CHECK (field_name <> ''),
    is_primary BOOLEAN NOT NULL DEFAULT FALSE,
    PRIMARY KEY (case_id, web_source_id, field_name)
);

CREATE TABLE project_favorites (
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    project_id BIGINT NOT NULL REFERENCES project_opportunities(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, project_id)
);

CREATE INDEX project_favorites_project_idx ON project_favorites (project_id, created_at DESC);

CREATE TABLE project_views (
    id BIGSERIAL PRIMARY KEY,
    project_id BIGINT NOT NULL REFERENCES project_opportunities(id) ON DELETE CASCADE,
    user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    visitor_key TEXT,
    viewed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX project_views_project_viewed_idx ON project_views (project_id, viewed_at DESC);

CREATE TABLE project_compare_items (
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    project_id BIGINT NOT NULL REFERENCES project_opportunities(id) ON DELETE CASCADE,
    added_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, project_id)
);

ALTER TABLE project_comparisons DROP CONSTRAINT IF EXISTS project_comparisons_items_check;
ALTER TABLE project_comparisons
    ADD CONSTRAINT project_comparisons_items_check
    CHECK (jsonb_typeof(items) = 'array' AND jsonb_array_length(items) BETWEEN 2 AND 5);

CREATE TABLE content_corrections (
    id BIGSERIAL PRIMARY KEY,
    target_type TEXT NOT NULL CHECK (target_type IN ('failure', 'plan', 'case', 'project')),
    target_id BIGINT NOT NULL,
    reason TEXT NOT NULL CHECK (reason <> ''),
    evidence_url TEXT CHECK (evidence_url IS NULL OR evidence_url ~* '^https?://'),
    contact TEXT,
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'accepted', 'rejected')),
    handled_by BIGINT REFERENCES users(id) ON DELETE SET NULL,
    handled_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX content_corrections_target_idx ON content_corrections (target_type, target_id, created_at DESC);

ALTER TABLE project_match_sessions
    ADD COLUMN name TEXT,
    ADD COLUMN input_snapshot JSONB NOT NULL DEFAULT '{}'::JSONB CHECK (jsonb_typeof(input_snapshot) = 'object'),
    ADD COLUMN parsed_profile JSONB NOT NULL DEFAULT '{}'::JSONB CHECK (jsonb_typeof(parsed_profile) = 'object'),
    ADD COLUMN field_sources JSONB NOT NULL DEFAULT '{}'::JSONB CHECK (jsonb_typeof(field_sources) = 'object'),
    ADD COLUMN analysis_summary TEXT,
    ADD COLUMN assumptions JSONB NOT NULL DEFAULT '[]'::JSONB CHECK (jsonb_typeof(assumptions) = 'array'),
    ADD COLUMN completeness NUMERIC(3,2) NOT NULL DEFAULT 0 CHECK (completeness BETWEEN 0 AND 1),
    ADD COLUMN rounds INTEGER NOT NULL DEFAULT 0 CHECK (rounds BETWEEN 0 AND 3),
    ADD COLUMN question_count INTEGER NOT NULL DEFAULT 0 CHECK (question_count BETWEEN 0 AND 8),
    ADD COLUMN kb_sufficiency NUMERIC(3,2) CHECK (kb_sufficiency BETWEEN 0 AND 1),
    ADD COLUMN research_job_id BIGINT REFERENCES web_research_jobs(id) ON DELETE SET NULL,
    ADD COLUMN idempotency_key TEXT,
    ADD COLUMN cancelled_at TIMESTAMPTZ,
    ADD COLUMN error_code TEXT;

CREATE UNIQUE INDEX project_match_sessions_user_idempotency_idx
    ON project_match_sessions (user_id, idempotency_key)
    WHERE idempotency_key IS NOT NULL AND idempotency_key <> '';
