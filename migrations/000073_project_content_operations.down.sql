DROP INDEX IF EXISTS project_operation_audits_target_idx;
DROP INDEX IF EXISTS project_operation_audits_created_idx;
DROP TABLE IF EXISTS project_operation_audits;

ALTER TABLE project_import_batches
    DROP COLUMN IF EXISTS rolled_back_at,
    DROP COLUMN IF EXISTS published_at,
    DROP COLUMN IF EXISTS validation_errors,
    DROP COLUMN IF EXISTS failed_count;
