ALTER TABLE copilot_memories
    DROP CONSTRAINT IF EXISTS copilot_memories_status_check;

ALTER TABLE copilot_memories
    DROP COLUMN IF EXISTS status;
