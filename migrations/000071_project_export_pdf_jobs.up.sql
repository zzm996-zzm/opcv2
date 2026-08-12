ALTER TABLE project_exports
    DROP CONSTRAINT IF EXISTS project_exports_status_check;

ALTER TABLE project_exports
    ADD COLUMN IF NOT EXISTS snapshot JSONB NOT NULL DEFAULT '{}'::JSONB,
    ADD COLUMN IF NOT EXISTS payload_bytes BYTEA,
    ADD COLUMN IF NOT EXISTS error_code TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ADD CONSTRAINT project_exports_status_check CHECK (status IN ('queued', 'running', 'ready', 'failed'));

CREATE UNIQUE INDEX IF NOT EXISTS project_exports_active_unique_idx
    ON project_exports (user_id, source_type, source_id, format)
    WHERE status IN ('queued', 'running');
