CREATE TABLE analysis_sessions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    mode TEXT NOT NULL,
    intent TEXT NOT NULL,
    status TEXT NOT NULL,
    questions JSONB NOT NULL DEFAULT '[]'::JSONB,
    result JSONB NOT NULL DEFAULT '{}'::JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX analysis_sessions_user_created_idx
    ON analysis_sessions (user_id, created_at DESC);
