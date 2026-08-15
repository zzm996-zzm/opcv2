CREATE TABLE task_view_preferences (
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    view_type TEXT NOT NULL DEFAULT 'list' CHECK (view_type = 'list'),
    columns JSONB NOT NULL DEFAULT '["title", "project", "assignee", "due_at", "priority", "status", "tags", "progress", "source"]'::jsonb,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, view_type)
);
