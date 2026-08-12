CREATE TABLE project_kb_state (
    singleton BOOLEAN PRIMARY KEY DEFAULT TRUE CHECK (singleton),
    active_version INTEGER NOT NULL DEFAULT 1 CHECK (active_version > 0),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO project_kb_state (singleton, active_version) VALUES (TRUE, 1)
ON CONFLICT (singleton) DO NOTHING;

CREATE TABLE project_kb_documents (
    version INTEGER NOT NULL CHECK (version > 0),
    document_id TEXT NOT NULL CHECK (document_id <> ''),
    title TEXT NOT NULL CHECK (title <> ''),
    body TEXT NOT NULL,
    metadata JSONB NOT NULL DEFAULT '{}'::JSONB CHECK (jsonb_typeof(metadata) = 'object'),
    search_vector TSVECTOR GENERATED ALWAYS AS (
        to_tsvector('simple', COALESCE(title, '') || ' ' || COALESCE(body, ''))
    ) STORED,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (version, document_id)
);

CREATE INDEX project_kb_documents_search_idx ON project_kb_documents USING GIN (search_vector);

INSERT INTO project_kb_documents (version, document_id, title, body, metadata)
SELECT 1, slug, title,
       concat_ws(' ', summary, industry, tags::TEXT, budget_band, difficulty, resource_requirements::TEXT, detail::TEXT),
       jsonb_build_object('slug', slug, 'kind', 'project', 'knowledge_base_version', '1')
FROM project_opportunities
WHERE status = 'published'
ON CONFLICT (version, document_id) DO NOTHING;

CREATE TABLE project_kb_reindex_jobs (
    id BIGSERIAL PRIMARY KEY,
    task_key TEXT NOT NULL UNIQUE CHECK (task_key <> ''),
    requested_by BIGINT REFERENCES users(id) ON DELETE SET NULL,
    source_version INTEGER NOT NULL CHECK (source_version > 0),
    target_version INTEGER NOT NULL CHECK (target_version > source_version),
    status TEXT NOT NULL DEFAULT 'queued' CHECK (status IN ('queued', 'running', 'ready', 'failed')),
    document_count INTEGER NOT NULL DEFAULT 0 CHECK (document_count >= 0),
    error_code TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    started_at TIMESTAMPTZ,
    finished_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE project_ai_answers (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    entry TEXT NOT NULL CHECK (entry IN ('search','banner_explore','quick_tag','explore','match','diagnose','detail_tab','compare','copilot')),
    raw_input TEXT NOT NULL DEFAULT '',
    parsed_intent JSONB NOT NULL DEFAULT '{}'::JSONB CHECK (jsonb_typeof(parsed_intent) = 'object'),
    cited_chunk_ids JSONB NOT NULL DEFAULT '[]'::JSONB CHECK (jsonb_typeof(cited_chunk_ids) = 'array'),
    cited_web_source_ids JSONB NOT NULL DEFAULT '[]'::JSONB CHECK (jsonb_typeof(cited_web_source_ids) = 'array'),
    research_job_id BIGINT REFERENCES web_research_jobs(id) ON DELETE SET NULL,
    kb_sufficiency NUMERIC(3,2) CHECK (kb_sufficiency BETWEEN 0 AND 1),
    result JSONB NOT NULL DEFAULT '{}'::JSONB,
    model_name TEXT NOT NULL DEFAULT '',
    prompt_version TEXT NOT NULL DEFAULT '',
    kb_version INTEGER NOT NULL CHECK (kb_version > 0),
    cache_key TEXT NOT NULL DEFAULT '',
    latency_ms INTEGER NOT NULL DEFAULT 0 CHECK (latency_ms >= 0),
    status TEXT NOT NULL CHECK (status IN ('ok', 'partial', 'ai_no_result', 'degraded', 'failed')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX project_ai_answers_created_idx ON project_ai_answers (created_at DESC, id DESC);
CREATE INDEX project_ai_answers_cache_idx ON project_ai_answers (cache_key, kb_version) WHERE cache_key <> '';

CREATE TABLE project_ai_response_cache (
    cache_key TEXT NOT NULL,
    kb_version INTEGER NOT NULL,
    response JSONB NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (cache_key, kb_version)
);

CREATE INDEX project_ai_response_cache_expiry_idx ON project_ai_response_cache (expires_at);

CREATE TABLE project_content_production_runs (
    id BIGSERIAL PRIMARY KEY,
    task_key TEXT NOT NULL UNIQUE CHECK (task_key <> ''),
    schedule_date DATE NOT NULL UNIQUE,
    status TEXT NOT NULL DEFAULT 'queued' CHECK (status IN ('queued', 'running', 'ready', 'failed')),
    planned_count INTEGER NOT NULL DEFAULT 30 CHECK (planned_count > 0),
    produced_count INTEGER NOT NULL DEFAULT 0 CHECK (produced_count >= 0),
    error_code TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    started_at TIMESTAMPTZ,
    finished_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
