ALTER TABLE tasks
    ADD COLUMN progress SMALLINT NOT NULL DEFAULT 0,
    ADD COLUMN completed_at TIMESTAMPTZ,
    ADD COLUMN version BIGINT NOT NULL DEFAULT 1,
    ADD COLUMN deleted_at TIMESTAMPTZ;

ALTER TABLE tasks
    ADD CONSTRAINT tasks_progress_check CHECK (progress BETWEEN 0 AND 100),
    ADD CONSTRAINT tasks_version_check CHECK (version > 0),
    ADD CONSTRAINT tasks_status_check CHECK (
        status IN ('todo', 'in_progress', 'review', 'blocked', 'completed', 'cancelled', 'reminder')
    );

UPDATE tasks
SET completed_at = COALESCE(completed_at, updated_at),
    progress = CASE WHEN progress = 0 THEN 100 ELSE progress END
WHERE status = 'completed';

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
                'growth_model',
                'copilot_message'
            ) AND
            (source_id IS NULL OR source_id > 0) AND
            source_title <> '' AND
            source_url LIKE '/%' AND
            source_url NOT LIKE '//%'
        )
    );

DROP INDEX IF EXISTS tasks_idempotency_unique_idx;

CREATE UNIQUE INDEX tasks_idempotency_unique_idx
    ON tasks (user_id, idempotency_key)
    WHERE idempotency_key <> '' AND deleted_at IS NULL;

CREATE INDEX tasks_active_user_created_idx
    ON tasks (user_id, created_at DESC, id DESC)
    WHERE deleted_at IS NULL;

CREATE INDEX tasks_active_user_status_due_idx
    ON tasks (user_id, status, due_at)
    WHERE deleted_at IS NULL;
