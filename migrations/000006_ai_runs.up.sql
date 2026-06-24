CREATE TABLE ai_runs (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    feature TEXT NOT NULL,
    prompt_version TEXT NOT NULL,
    provider TEXT NOT NULL,
    model TEXT NOT NULL,
    status TEXT NOT NULL,
    request JSONB NOT NULL DEFAULT '{}'::JSONB,
    response JSONB NOT NULL DEFAULT '{}'::JSONB,
    error_code TEXT NOT NULL DEFAULT '',
    error_message TEXT NOT NULL DEFAULT '',
    input_tokens INTEGER NOT NULL DEFAULT 0,
    output_tokens INTEGER NOT NULL DEFAULT 0,
    latency_ms INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX ai_runs_user_created_idx
    ON ai_runs (user_id, created_at DESC);

CREATE INDEX ai_runs_feature_created_idx
    ON ai_runs (feature, created_at DESC);

CREATE INDEX ai_runs_status_created_idx
    ON ai_runs (status, created_at DESC);
