CREATE TABLE geo_metrics (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    key TEXT NOT NULL,
    label TEXT NOT NULL,
    value TEXT NOT NULL,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, key)
);

CREATE INDEX geo_metrics_user_sort_idx
    ON geo_metrics (user_id, sort_order ASC, id ASC);

CREATE TABLE geo_engine_coverages (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    coverage_percent INTEGER NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT '',
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, name)
);

CREATE INDEX geo_engine_coverages_user_sort_idx
    ON geo_engine_coverages (user_id, sort_order ASC, id ASC);

CREATE TABLE geo_lead_signals (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    detail TEXT NOT NULL DEFAULT '',
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX geo_lead_signals_user_sort_idx
    ON geo_lead_signals (user_id, sort_order ASC, id ASC);

CREATE TABLE geo_keyword_opportunities (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    query TEXT NOT NULL,
    intent TEXT NOT NULL DEFAULT '',
    coverage TEXT NOT NULL DEFAULT '',
    score INTEGER NOT NULL DEFAULT 0,
    action TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX geo_keyword_opportunities_user_score_idx
    ON geo_keyword_opportunities (user_id, score DESC, id ASC);

CREATE TABLE geo_content_tasks (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type TEXT NOT NULL DEFAULT '',
    title TEXT NOT NULL,
    priority TEXT NOT NULL DEFAULT '',
    due_at TEXT NOT NULL DEFAULT '',
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX geo_content_tasks_user_sort_idx
    ON geo_content_tasks (user_id, sort_order ASC, id ASC);
