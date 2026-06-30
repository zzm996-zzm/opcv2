CREATE TABLE sandbox_sessions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    goal TEXT NOT NULL,
    target_users TEXT NOT NULL,
    product TEXT NOT NULL,
    roles JSONB NOT NULL DEFAULT '[]'::JSONB,
    status TEXT NOT NULL,
    report JSONB NOT NULL DEFAULT '{}'::JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX sandbox_sessions_user_created_idx
    ON sandbox_sessions (user_id, created_at DESC);

CREATE INDEX sandbox_sessions_status_created_idx
    ON sandbox_sessions (status, created_at DESC);
