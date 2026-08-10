ALTER TABLE project_match_sessions
    ADD COLUMN generation_attempt INTEGER NOT NULL DEFAULT 0 CHECK (generation_attempt >= 0),
    ADD COLUMN progress_percent INTEGER NOT NULL DEFAULT 0 CHECK (progress_percent BETWEEN 0 AND 100),
    ADD COLUMN current_step TEXT NOT NULL DEFAULT '';

CREATE TABLE project_match_progress_events (
    id BIGSERIAL PRIMARY KEY,
    match_id BIGINT NOT NULL REFERENCES project_match_sessions(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    attempt INTEGER NOT NULL CHECK (attempt > 0),
    event TEXT NOT NULL CHECK (event IN (
        'queued', 'analyzing', 'retrieving_kb', 'researching_web', 'merging',
        'generating', 'done', 'partial', 'error', 'canceled'
    )),
    progress_percent INTEGER NOT NULL CHECK (progress_percent BETWEEN 0 AND 100),
    payload JSONB NOT NULL DEFAULT '{}'::JSONB CHECK (jsonb_typeof(payload) = 'object'),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX project_match_progress_events_replay_idx
    ON project_match_progress_events (match_id, user_id, id);
