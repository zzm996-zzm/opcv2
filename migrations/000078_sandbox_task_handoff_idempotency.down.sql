DROP INDEX IF EXISTS tasks_idempotency_unique_idx;

ALTER TABLE tasks
    DROP COLUMN IF EXISTS idempotency_key;
