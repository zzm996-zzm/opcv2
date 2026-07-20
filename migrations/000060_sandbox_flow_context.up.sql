ALTER TABLE sandbox_sessions
    ADD COLUMN intake JSONB NOT NULL DEFAULT '{}'::jsonb,
    ADD COLUMN settings JSONB NOT NULL DEFAULT '{}'::jsonb;

