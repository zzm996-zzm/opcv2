DROP INDEX IF EXISTS notifications_user_dedupe_idx;
ALTER TABLE notifications
    DROP COLUMN IF EXISTS dedupe_key;
