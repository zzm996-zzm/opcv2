CREATE TABLE sandbox_messages (
    id BIGSERIAL PRIMARY KEY,
    session_id BIGINT NOT NULL REFERENCES sandbox_sessions(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role TEXT NOT NULL,
    question TEXT NOT NULL,
    answer TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX sandbox_messages_session_created_idx
    ON sandbox_messages (session_id, created_at ASC);

CREATE INDEX sandbox_messages_user_created_idx
    ON sandbox_messages (user_id, created_at DESC);
