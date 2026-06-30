CREATE TABLE tasks (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    project TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL,
    priority TEXT NOT NULL,
    due_at TIMESTAMPTZ,
    tools JSONB NOT NULL DEFAULT '[]'::JSONB,
    learning TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX tasks_user_created_idx
    ON tasks (user_id, created_at DESC);

CREATE INDEX tasks_user_status_due_idx
    ON tasks (user_id, status, due_at);
