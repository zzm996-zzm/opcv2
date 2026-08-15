DROP INDEX IF EXISTS tasks_active_user_status_due_idx;
DROP INDEX IF EXISTS tasks_active_user_created_idx;

DROP INDEX IF EXISTS tasks_idempotency_unique_idx;

CREATE UNIQUE INDEX tasks_idempotency_unique_idx
    ON tasks (user_id, idempotency_key)
    WHERE idempotency_key <> '';

ALTER TABLE tasks
    DROP CONSTRAINT IF EXISTS tasks_source_fields_check,
    ADD CONSTRAINT tasks_source_fields_check CHECK (
        (
            source_type = '' AND
            source_id IS NULL AND
            source_title = '' AND
            source_url = ''
        ) OR (
            source_type IN (
                'analysis_session',
                'project_match',
                'sandbox_session',
                'competitor_scan',
                'competitor_monitoring',
                'enterprise_diagnosis',
                'lead_task',
                'crm_customer',
                'learning_diagnosis',
                'growth_model'
            ) AND
            (source_id IS NULL OR source_id > 0) AND
            source_title <> '' AND
            source_url LIKE '/%' AND
            source_url NOT LIKE '//%'
        )
    ),
    DROP CONSTRAINT IF EXISTS tasks_progress_check,
    DROP CONSTRAINT IF EXISTS tasks_version_check,
    DROP CONSTRAINT IF EXISTS tasks_status_check,
    DROP COLUMN IF EXISTS deleted_at,
    DROP COLUMN IF EXISTS version,
    DROP COLUMN IF EXISTS completed_at,
    DROP COLUMN IF EXISTS progress;
