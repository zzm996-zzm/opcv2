CREATE TABLE user_projects (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'draft',
    source_type TEXT NOT NULL DEFAULT '',
    source_id BIGINT,
    idempotency_key TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT user_projects_status_check CHECK (status IN ('draft', 'active', 'archived'))
);

CREATE INDEX user_projects_user_created_idx ON user_projects (user_id, created_at DESC, id DESC);
CREATE UNIQUE INDEX user_projects_user_idempotency_idx
    ON user_projects (user_id, idempotency_key)
    WHERE idempotency_key <> '';
