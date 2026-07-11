DROP INDEX task_reminders_due_idx;

ALTER TABLE task_reminders
    DROP COLUMN recurrence;

CREATE INDEX task_reminders_due_idx
    ON task_reminders (remind_at, id)
    WHERE sent_at IS NULL;
