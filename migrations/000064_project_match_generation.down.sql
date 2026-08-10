DROP TABLE IF EXISTS project_match_progress_events;

ALTER TABLE project_match_sessions
    DROP COLUMN IF EXISTS current_step,
    DROP COLUMN IF EXISTS progress_percent,
    DROP COLUMN IF EXISTS generation_attempt;
