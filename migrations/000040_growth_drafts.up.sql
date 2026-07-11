CREATE TABLE growth_drafts (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    input TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('needs_input', 'ready', 'calculated')),
    assumptions JSONB NOT NULL DEFAULT '{}'::JSONB,
    questions JSONB NOT NULL DEFAULT '[]'::JSONB,
    answers JSONB NOT NULL DEFAULT '{}'::JSONB,
    model_id BIGINT REFERENCES growth_models(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX growth_drafts_user_updated_idx ON growth_drafts (user_id, updated_at DESC);
