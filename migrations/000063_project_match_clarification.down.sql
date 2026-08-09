DROP INDEX IF EXISTS project_match_sessions_v2_user_created_idx;

ALTER TABLE project_match_sessions
    DROP COLUMN IF EXISTS skipped_at,
    DROP COLUMN IF EXISTS revision,
    DROP COLUMN IF EXISTS answer_events,
    DROP COLUMN IF EXISTS missing_fields,
    DROP COLUMN IF EXISTS clarification_questions,
    DROP COLUMN IF EXISTS workflow_version;
