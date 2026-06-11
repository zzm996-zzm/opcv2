CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TABLE app_metadata (
    key TEXT PRIMARY KEY,
    value JSONB NOT NULL DEFAULT '{}'::JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO app_metadata (key, value)
VALUES ('schema', '{"version": 1}'::JSONB)
ON CONFLICT (key) DO NOTHING;
