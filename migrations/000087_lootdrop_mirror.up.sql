CREATE TABLE lootdrop_sync_runs (
    id BIGSERIAL PRIMARY KEY,
    source_run_id BIGINT,
    source_mode TEXT NOT NULL DEFAULT '',
    source_status TEXT NOT NULL DEFAULT '',
    source_site_url TEXT NOT NULL,
    source_started_at TIMESTAMPTZ,
    source_finished_at TIMESTAMPTZ,
    source_counts JSONB NOT NULL DEFAULT '{}'::JSONB CHECK (jsonb_typeof(source_counts) = 'object'),
    imported_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    status TEXT NOT NULL DEFAULT 'imported' CHECK (status IN ('importing', 'imported', 'failed')),
    error_message TEXT NOT NULL DEFAULT ''
);

CREATE INDEX lootdrop_sync_runs_imported_idx ON lootdrop_sync_runs (imported_at DESC, id DESC);

CREATE TABLE lootdrop_startups (
    source_id BIGINT PRIMARY KEY,
    name TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    sector TEXT NOT NULL DEFAULT '',
    country TEXT NOT NULL DEFAULT '',
    founders JSONB NOT NULL DEFAULT '[]'::JSONB,
    investors JSONB NOT NULL DEFAULT '[]'::JSONB,
    start_year INTEGER,
    end_year INTEGER,
    total_funding NUMERIC(24, 2),
    difficulty INTEGER,
    difficulty_reason TEXT NOT NULL DEFAULT '',
    scalability INTEGER,
    scalability_reason TEXT NOT NULL DEFAULT '',
    market_potential TEXT NOT NULL DEFAULT '',
    market_potential_reason TEXT NOT NULL DEFAULT '',
    cause_of_death TEXT NOT NULL DEFAULT '',
    primary_cause_of_death TEXT NOT NULL DEFAULT '',
    the_loot JSONB NOT NULL DEFAULT '[]'::JSONB,
    market_analysis TEXT NOT NULL DEFAULT '',
    pivot_idea JSONB NOT NULL DEFAULT '{}'::JSONB,
    product_type TEXT NOT NULL DEFAULT '',
    source_status TEXT NOT NULL DEFAULT '',
    views BIGINT,
    condensed_value_prop TEXT NOT NULL DEFAULT '',
    condensed_cause_of_death TEXT NOT NULL DEFAULT '',
    source_url TEXT NOT NULL DEFAULT 'https://www.loot-drop.io/database-view',
    source_payload JSONB NOT NULL DEFAULT '{}'::JSONB,
    source_hash CHAR(64) NOT NULL,
    scraped_at TIMESTAMPTZ,
    last_import_run_id BIGINT REFERENCES lootdrop_sync_runs(id) ON DELETE SET NULL,
    imported_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX lootdrop_startups_sector_idx ON lootdrop_startups (sector);
CREATE INDEX lootdrop_startups_cause_idx ON lootdrop_startups (primary_cause_of_death);
CREATE INDEX lootdrop_startups_product_type_idx ON lootdrop_startups (product_type);
CREATE INDEX lootdrop_startups_updated_idx ON lootdrop_startups (updated_at DESC, source_id DESC);

CREATE TABLE lootdrop_rebuild_plans (
    source_id BIGINT PRIMARY KEY,
    name TEXT NOT NULL DEFAULT '',
    sector TEXT NOT NULL DEFAULT '',
    product_type TEXT NOT NULL DEFAULT '',
    country TEXT NOT NULL DEFAULT '',
    total_funding NUMERIC(24, 2),
    primary_cause_of_death TEXT NOT NULL DEFAULT '',
    market_potential TEXT NOT NULL DEFAULT '',
    scalability INTEGER,
    difficulty INTEGER,
    pivot_idea JSONB NOT NULL DEFAULT '{}'::JSONB,
    the_loot JSONB NOT NULL DEFAULT '[]'::JSONB,
    source_page INTEGER,
    source_payload JSONB NOT NULL DEFAULT '{}'::JSONB,
    source_hash CHAR(64) NOT NULL,
    scraped_at TIMESTAMPTZ,
    last_import_run_id BIGINT REFERENCES lootdrop_sync_runs(id) ON DELETE SET NULL,
    imported_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX lootdrop_rebuild_plans_market_idx ON lootdrop_rebuild_plans (market_potential, scalability DESC);

CREATE TABLE lootdrop_ideas (
    source_id BIGINT PRIMARY KEY,
    title TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    model TEXT NOT NULL DEFAULT '',
    effort TEXT NOT NULL DEFAULT '',
    speed TEXT NOT NULL DEFAULT '',
    category TEXT NOT NULL DEFAULT '',
    tags JSONB NOT NULL DEFAULT '[]'::JSONB,
    persona TEXT NOT NULL DEFAULT '',
    source_payload JSONB NOT NULL DEFAULT '{}'::JSONB,
    source_hash CHAR(64) NOT NULL,
    scraped_at TIMESTAMPTZ,
    last_import_run_id BIGINT REFERENCES lootdrop_sync_runs(id) ON DELETE SET NULL,
    imported_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX lootdrop_ideas_category_idx ON lootdrop_ideas (category);

CREATE TABLE lootdrop_dataset_rows (
    dataset TEXT NOT NULL,
    record_key TEXT NOT NULL,
    payload JSONB NOT NULL DEFAULT '{}'::JSONB,
    raw_text TEXT NOT NULL DEFAULT '',
    source_url TEXT NOT NULL DEFAULT '',
    content_type TEXT NOT NULL DEFAULT '',
    http_status INTEGER,
    source_hash CHAR(64),
    source_fetched_at TIMESTAMPTZ,
    source_row_count INTEGER,
    scraped_at TIMESTAMPTZ,
    last_import_run_id BIGINT REFERENCES lootdrop_sync_runs(id) ON DELETE SET NULL,
    imported_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (dataset, record_key)
);

CREATE INDEX lootdrop_dataset_rows_dataset_idx ON lootdrop_dataset_rows (dataset, updated_at DESC);

CREATE TABLE lootdrop_translations (
    dataset TEXT NOT NULL,
    record_key TEXT NOT NULL,
    language TEXT NOT NULL DEFAULT 'zh-CN',
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'completed', 'failed')),
    translated_payload JSONB NOT NULL DEFAULT '{}'::JSONB,
    model TEXT NOT NULL DEFAULT '',
    source_hash CHAR(64),
    error_message TEXT NOT NULL DEFAULT '',
    translated_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (dataset, record_key, language)
);

CREATE INDEX lootdrop_translations_pending_idx ON lootdrop_translations (dataset, language, status, updated_at);

CREATE TABLE lootdrop_import_state (
    singleton BOOLEAN PRIMARY KEY DEFAULT TRUE CHECK (singleton),
    active_run_id BIGINT REFERENCES lootdrop_sync_runs(id) ON DELETE SET NULL,
    last_source_run_id BIGINT,
    last_source_finished_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO lootdrop_import_state (singleton) VALUES (TRUE)
ON CONFLICT (singleton) DO NOTHING;
