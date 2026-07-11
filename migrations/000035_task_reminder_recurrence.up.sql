ALTER TABLE task_reminders
    ADD COLUMN recurrence TEXT NOT NULL DEFAULT 'once'
        CHECK (recurrence IN ('once', 'daily', 'weekly'));

DROP INDEX task_reminders_due_idx;

CREATE INDEX task_reminders_due_idx
    ON task_reminders (remind_at, id)
    WHERE recurrence <> 'once' OR sent_at IS NULL;
