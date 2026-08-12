DROP INDEX IF EXISTS project_exports_active_unique_idx;

ALTER TABLE project_exports
    DROP CONSTRAINT IF EXISTS project_exports_status_check,
    DROP COLUMN IF EXISTS updated_at,
    DROP COLUMN IF EXISTS error_code,
    DROP COLUMN IF EXISTS payload_bytes,
    DROP COLUMN IF EXISTS snapshot,
    ADD CONSTRAINT project_exports_status_check CHECK (status IN ('ready', 'failed'));
