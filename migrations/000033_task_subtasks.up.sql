CREATE TABLE task_subtasks (
    id BIGSERIAL PRIMARY KEY,
    task_id BIGINT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    assignee TEXT NOT NULL DEFAULT '',
    due_at TIMESTAMPTZ,
    completed BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX task_subtasks_task_created_idx ON task_subtasks (task_id, created_at, id);
CREATE INDEX task_subtasks_user_idx ON task_subtasks (user_id);
