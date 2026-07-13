ALTER TABLE sandbox_sessions
    DROP COLUMN IF EXISTS run_attempt,
    DROP COLUMN IF EXISTS error_message,
    DROP COLUMN IF EXISTS current_step,
    DROP COLUMN IF EXISTS progress_percent;
