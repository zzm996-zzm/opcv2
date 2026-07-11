DROP INDEX tasks_user_source_idx;

ALTER TABLE tasks
    DROP COLUMN source_url,
    DROP COLUMN source_title,
    DROP COLUMN source_id,
    DROP COLUMN source_type;
