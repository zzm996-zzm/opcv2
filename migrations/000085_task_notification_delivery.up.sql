ALTER TABLE notifications
    ADD COLUMN dedupe_key TEXT NOT NULL DEFAULT '';

CREATE UNIQUE INDEX notifications_user_dedupe_idx
    ON notifications (user_id, dedupe_key)
    WHERE dedupe_key <> '';
