ALTER TABLE competitor_scans
    ADD COLUMN IF NOT EXISTS progress_percent INTEGER NOT NULL DEFAULT 0 CHECK (progress_percent >= 0 AND progress_percent <= 100),
    ADD COLUMN IF NOT EXISTS current_step TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS error_message TEXT NOT NULL DEFAULT '';

UPDATE competitor_scans
SET status = 'succeeded',
    progress_percent = 100,
    current_step = 'succeeded',
    error_message = ''
WHERE status = 'completed';
