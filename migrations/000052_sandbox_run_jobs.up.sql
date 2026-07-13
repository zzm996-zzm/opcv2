ALTER TABLE sandbox_sessions
    ADD COLUMN progress_percent INTEGER NOT NULL DEFAULT 0
        CHECK (progress_percent >= 0 AND progress_percent <= 100),
    ADD COLUMN current_step TEXT NOT NULL DEFAULT '',
    ADD COLUMN error_message TEXT NOT NULL DEFAULT '',
    ADD COLUMN run_attempt INTEGER NOT NULL DEFAULT 0 CHECK (run_attempt >= 0);

UPDATE sandbox_sessions
SET progress_percent = CASE WHEN status = 'completed' THEN 100 ELSE 0 END,
    current_step = CASE WHEN status = 'completed' THEN 'completed' ELSE status END;
