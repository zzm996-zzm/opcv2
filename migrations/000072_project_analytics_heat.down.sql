DROP INDEX IF EXISTS project_unlocks_project_unlocked_idx;
DROP TABLE IF EXISTS project_unlocks;
ALTER TABLE project_views DROP COLUMN IF EXISTS analytics_event_id;
DROP INDEX IF EXISTS project_analytics_events_project_created_idx;
DROP INDEX IF EXISTS project_analytics_events_name_created_idx;
DROP TABLE IF EXISTS project_analytics_events;
