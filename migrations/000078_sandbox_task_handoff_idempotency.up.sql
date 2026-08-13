ALTER TABLE tasks
    ADD COLUMN idempotency_key TEXT NOT NULL DEFAULT '';

CREATE UNIQUE INDEX tasks_idempotency_unique_idx
    ON tasks (user_id, idempotency_key)
    WHERE idempotency_key <> '';
