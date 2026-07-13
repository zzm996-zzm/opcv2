DROP INDEX IF EXISTS copilot_files_user_status_updated_idx;

ALTER TABLE copilot_files
    DROP COLUMN IF EXISTS error_code,
    DROP COLUMN IF EXISTS extracted_chars,
    DROP COLUMN IF EXISTS sha256,
    DROP COLUMN IF EXISTS source,
    DROP COLUMN IF EXISTS status;
