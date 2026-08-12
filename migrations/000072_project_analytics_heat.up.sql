CREATE TABLE project_analytics_events (
    id BIGSERIAL PRIMARY KEY,
    event_id TEXT NOT NULL UNIQUE CHECK (event_id <> ''),
    event_name TEXT NOT NULL CHECK (event_name <> ''),
    user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    visitor_hash CHAR(64) NOT NULL CHECK (visitor_hash ~ '^[0-9a-f]{64}$'),
    route TEXT NOT NULL CHECK (route <> ''),
    ref_module TEXT,
    project_id BIGINT REFERENCES project_opportunities(id) ON DELETE CASCADE,
    properties JSONB NOT NULL DEFAULT '{}'::JSONB CHECK (jsonb_typeof(properties) = 'object'),
    occurred_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX project_analytics_events_name_created_idx
    ON project_analytics_events (event_name, created_at DESC);
CREATE INDEX project_analytics_events_project_created_idx
    ON project_analytics_events (project_id, created_at DESC)
    WHERE project_id IS NOT NULL;

ALTER TABLE project_views
    ADD COLUMN analytics_event_id BIGINT UNIQUE REFERENCES project_analytics_events(id) ON DELETE SET NULL;

CREATE TABLE project_unlocks (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    project_id BIGINT NOT NULL REFERENCES project_opportunities(id) ON DELETE CASCADE,
    order_no TEXT NOT NULL UNIQUE CHECK (order_no <> ''),
    unlocked_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, project_id)
);

CREATE INDEX project_unlocks_project_unlocked_idx
    ON project_unlocks (project_id, unlocked_at DESC);
