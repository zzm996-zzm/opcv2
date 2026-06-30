CREATE TABLE competitor_scans (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    targets JSONB NOT NULL DEFAULT '[]'::JSONB,
    focus TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL,
    competitors JSONB NOT NULL DEFAULT '[]'::JSONB,
    conclusions JSONB NOT NULL DEFAULT '[]'::JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX competitor_scans_user_created_idx
    ON competitor_scans (user_id, created_at DESC);

CREATE TABLE competitor_watchlist (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    category TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT '',
    threat TEXT NOT NULL DEFAULT '',
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    channels JSONB NOT NULL DEFAULT '[]'::JSONB,
    signal TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX competitor_watchlist_user_seen_idx
    ON competitor_watchlist (user_id, last_seen_at DESC);

CREATE TABLE competitor_events (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    company TEXT NOT NULL,
    title TEXT NOT NULL,
    detail TEXT NOT NULL DEFAULT '',
    level TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX competitor_events_user_occurred_idx
    ON competitor_events (user_id, occurred_at DESC);
