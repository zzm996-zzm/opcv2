UPDATE competitor_scans
SET status = 'completed'
WHERE status = 'succeeded';

ALTER TABLE competitor_scans
    DROP COLUMN IF EXISTS error_message,
    DROP COLUMN IF EXISTS current_step,
    DROP COLUMN IF EXISTS progress_percent;
