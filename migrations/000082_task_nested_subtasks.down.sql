DROP INDEX IF EXISTS task_subtasks_task_parent_created_idx;
DROP INDEX IF EXISTS task_subtasks_parent_idx;

ALTER TABLE task_subtasks
    DROP CONSTRAINT IF EXISTS task_subtasks_not_self_parent,
    DROP COLUMN IF EXISTS parent_subtask_id;
