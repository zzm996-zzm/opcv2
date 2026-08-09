ALTER TABLE project_match_sessions
    ADD COLUMN workflow_version SMALLINT NOT NULL DEFAULT 1 CHECK (workflow_version IN (1, 2)),
    ADD COLUMN clarification_questions JSONB NOT NULL DEFAULT '[]'::JSONB
        CHECK (jsonb_typeof(clarification_questions) = 'array'),
    ADD COLUMN missing_fields JSONB NOT NULL DEFAULT '[]'::JSONB
        CHECK (jsonb_typeof(missing_fields) = 'array'),
    ADD COLUMN answer_events JSONB NOT NULL DEFAULT '[]'::JSONB
        CHECK (jsonb_typeof(answer_events) = 'array'),
    ADD COLUMN revision INTEGER NOT NULL DEFAULT 1 CHECK (revision > 0),
    ADD COLUMN skipped_at TIMESTAMPTZ;

CREATE INDEX project_match_sessions_v2_user_created_idx
    ON project_match_sessions (user_id, created_at DESC, id)
    WHERE workflow_version = 2;
