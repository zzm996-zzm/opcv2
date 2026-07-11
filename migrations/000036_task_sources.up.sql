ALTER TABLE tasks
    ADD COLUMN source_type TEXT NOT NULL DEFAULT '',
    ADD COLUMN source_id BIGINT,
    ADD COLUMN source_title TEXT NOT NULL DEFAULT '',
    ADD COLUMN source_url TEXT NOT NULL DEFAULT '',
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
                'crm_customer'
            ) AND
            (source_id IS NULL OR source_id > 0) AND
            source_title <> '' AND
            source_url LIKE '/%' AND
            source_url NOT LIKE '//%'
        )
    );

CREATE INDEX tasks_user_source_idx
    ON tasks (user_id, source_type, source_id)
    WHERE source_type <> '';
