ALTER TABLE copilot_files
    ADD COLUMN status TEXT NOT NULL DEFAULT 'ready',
    ADD COLUMN source TEXT NOT NULL DEFAULT 'pasted',
    ADD COLUMN sha256 TEXT NOT NULL DEFAULT '',
    ADD COLUMN extracted_chars INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN error_code TEXT NOT NULL DEFAULT '';

UPDATE copilot_files
SET extracted_chars = CHAR_LENGTH(content)
WHERE extracted_chars = 0;

CREATE INDEX copilot_files_user_status_updated_idx
    ON copilot_files (user_id, status, updated_at DESC);
