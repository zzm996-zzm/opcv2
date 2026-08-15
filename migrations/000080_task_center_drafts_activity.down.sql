DROP TRIGGER IF EXISTS tasks_record_activity ON tasks;
DROP FUNCTION IF EXISTS record_task_activity();

DROP INDEX IF EXISTS task_activities_user_created_idx;
DROP INDEX IF EXISTS task_activities_task_created_idx;
DROP TABLE IF EXISTS task_activities;

DROP INDEX IF EXISTS task_ai_drafts_user_status_idx;
DROP INDEX IF EXISTS task_ai_drafts_user_created_idx;
DROP TABLE IF EXISTS task_ai_drafts;
