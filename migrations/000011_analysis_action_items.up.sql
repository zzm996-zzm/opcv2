CREATE TABLE analysis_action_items (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    session_id BIGINT NOT NULL REFERENCES analysis_sessions(id) ON DELETE CASCADE,
    day_index INTEGER NOT NULL CHECK (day_index > 0),
    title TEXT NOT NULL,
    detail TEXT NOT NULL,
    completed BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX analysis_action_items_session_idx
    ON analysis_action_items (user_id, session_id, day_index);
