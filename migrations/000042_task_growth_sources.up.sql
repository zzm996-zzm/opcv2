ALTER TABLE tasks
    DROP CONSTRAINT tasks_source_fields_check;

ALTER TABLE tasks
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
    );
