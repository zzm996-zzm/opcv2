ALTER TABLE copilot_memories
    ADD COLUMN status TEXT NOT NULL DEFAULT 'active';

ALTER TABLE copilot_memories
    ADD CONSTRAINT copilot_memories_status_check
    CHECK (status IN ('pending', 'active', 'inactive'));
