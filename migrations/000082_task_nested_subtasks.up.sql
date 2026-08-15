ALTER TABLE task_subtasks
    ADD COLUMN parent_subtask_id BIGINT REFERENCES task_subtasks(id) ON DELETE CASCADE,
    ADD CONSTRAINT task_subtasks_not_self_parent CHECK (parent_subtask_id IS NULL OR parent_subtask_id <> id);

CREATE INDEX task_subtasks_parent_idx ON task_subtasks (parent_subtask_id);
CREATE INDEX task_subtasks_task_parent_created_idx ON task_subtasks (task_id, parent_subtask_id, created_at, id);
