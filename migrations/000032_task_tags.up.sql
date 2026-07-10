ALTER TABLE tasks
    ADD COLUMN tags JSONB NOT NULL DEFAULT '[]'::JSONB;

CREATE INDEX tasks_tags_idx
    ON tasks USING GIN (tags);
