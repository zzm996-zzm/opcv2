ALTER TABLE task_view_preferences
    ALTER COLUMN columns SET DEFAULT '["title", "project", "assignee", "due_at", "priority", "status", "tags", "progress", "source"]'::jsonb;

UPDATE task_view_preferences
SET columns = '["title", "project", "assignee", "due_at", "priority", "status", "tags", "progress", "source"]'::jsonb,
    updated_at = NOW()
WHERE columns = '["title", "project", "assignee", "due_at", "priority", "status"]'::jsonb;
