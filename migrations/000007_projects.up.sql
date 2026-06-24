CREATE TABLE project_match_sessions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    intent TEXT NOT NULL,
    status TEXT NOT NULL,
    questions JSONB NOT NULL DEFAULT '[]'::JSONB,
    result JSONB NOT NULL DEFAULT '{}'::JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX project_match_sessions_user_created_idx
    ON project_match_sessions (user_id, created_at DESC);

CREATE TABLE project_match_favorites (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    session_id BIGINT NOT NULL REFERENCES project_match_sessions(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, session_id)
);

CREATE INDEX project_match_favorites_user_created_idx
    ON project_match_favorites (user_id, created_at DESC);
