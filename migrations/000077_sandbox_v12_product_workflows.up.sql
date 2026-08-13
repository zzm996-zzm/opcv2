CREATE TABLE IF NOT EXISTS sandbox_analytics_events (
    id BIGSERIAL PRIMARY KEY,
    event_id TEXT NOT NULL UNIQUE,
    event_name TEXT NOT NULL,
    user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    visitor_hash CHAR(64) NOT NULL,
    run_id BIGINT REFERENCES sandbox_sessions(id) ON DELETE SET NULL,
    route TEXT NOT NULL,
    properties JSONB NOT NULL DEFAULT '{}'::JSONB,
    occurred_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT sandbox_analytics_event_id_check CHECK (char_length(event_id) BETWEEN 8 AND 128),
    CONSTRAINT sandbox_analytics_event_name_check CHECK (char_length(event_name) BETWEEN 3 AND 100),
    CONSTRAINT sandbox_analytics_route_check CHECK (char_length(route) BETWEEN 1 AND 500)
);

CREATE INDEX IF NOT EXISTS sandbox_analytics_user_created_idx
    ON sandbox_analytics_events (user_id, created_at DESC)
    WHERE user_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS sandbox_analytics_run_created_idx
    ON sandbox_analytics_events (run_id, created_at DESC)
    WHERE run_id IS NOT NULL;

CREATE TABLE IF NOT EXISTS sandbox_run_follow_ups (
    id BIGSERIAL PRIMARY KEY,
    run_id BIGINT NOT NULL REFERENCES sandbox_sessions(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_code TEXT NOT NULL REFERENCES sandbox_role_configs(role_code),
    question TEXT NOT NULL,
    answer TEXT NOT NULL,
    input_context_hash CHAR(64) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT sandbox_follow_up_question_check CHECK (char_length(question) BETWEEN 1 AND 2000),
    CONSTRAINT sandbox_follow_up_answer_check CHECK (char_length(answer) BETWEEN 1 AND 10000)
);

CREATE INDEX IF NOT EXISTS sandbox_follow_ups_owner_run_idx
    ON sandbox_run_follow_ups (user_id, run_id, id);

