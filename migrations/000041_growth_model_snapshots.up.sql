CREATE TABLE growth_model_snapshots (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    model_id BIGINT NOT NULL REFERENCES growth_models(id) ON DELETE CASCADE,
    model_name TEXT NOT NULL,
    assumptions JSONB NOT NULL,
    result JSONB NOT NULL,
    scenarios JSONB NOT NULL,
    forecast JSONB NOT NULL,
    recommendations JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX growth_model_snapshots_model_created_idx
    ON growth_model_snapshots (user_id, model_id, created_at DESC);
