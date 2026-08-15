CREATE TABLE task_ai_drafts (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    goal TEXT NOT NULL,
    source_type TEXT NOT NULL DEFAULT '',
    source_id BIGINT,
    source_title TEXT NOT NULL DEFAULT '',
    source_url TEXT NOT NULL DEFAULT '',
    tasks JSONB NOT NULL DEFAULT '[]'::JSONB,
    status TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'adopted')),
    adopted_task_ids JSONB NOT NULL DEFAULT '[]'::JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    adopted_at TIMESTAMPTZ
);

CREATE INDEX task_ai_drafts_user_created_idx
    ON task_ai_drafts (user_id, created_at DESC, id DESC);

CREATE INDEX task_ai_drafts_user_status_idx
    ON task_ai_drafts (user_id, status, created_at DESC);

CREATE TABLE task_activities (
    id BIGSERIAL PRIMARY KEY,
    task_id BIGINT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    action TEXT NOT NULL,
    before_data JSONB NOT NULL DEFAULT '{}'::JSONB,
    after_data JSONB NOT NULL DEFAULT '{}'::JSONB,
    metadata JSONB NOT NULL DEFAULT '{}'::JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX task_activities_task_created_idx
    ON task_activities (task_id, created_at DESC, id DESC);

CREATE INDEX task_activities_user_created_idx
    ON task_activities (user_id, created_at DESC, id DESC);

CREATE OR REPLACE FUNCTION record_task_activity() RETURNS TRIGGER AS $$
DECLARE
    activity_action TEXT;
    before_payload JSONB := '{}'::JSONB;
    after_payload JSONB := '{}'::JSONB;
BEGIN
    IF TG_OP = 'INSERT' THEN
        activity_action := 'created';
    ELSE
        IF OLD.title IS NOT DISTINCT FROM NEW.title
            AND OLD.assignee IS NOT DISTINCT FROM NEW.assignee
            AND OLD.project IS NOT DISTINCT FROM NEW.project
            AND OLD.status IS NOT DISTINCT FROM NEW.status
            AND OLD.priority IS NOT DISTINCT FROM NEW.priority
            AND OLD.tags IS NOT DISTINCT FROM NEW.tags
            AND OLD.due_at IS NOT DISTINCT FROM NEW.due_at
            AND OLD.progress IS NOT DISTINCT FROM NEW.progress
            AND OLD.deleted_at IS NOT DISTINCT FROM NEW.deleted_at THEN
            RETURN NEW;
        END IF;
        IF OLD.deleted_at IS NULL AND NEW.deleted_at IS NOT NULL THEN
            activity_action := 'deleted';
        ELSIF OLD.deleted_at IS NOT NULL AND NEW.deleted_at IS NULL THEN
            activity_action := 'restored';
        ELSIF OLD.status IS DISTINCT FROM NEW.status THEN
            activity_action := 'status_changed';
        ELSE
            activity_action := 'updated';
        END IF;
        before_payload := jsonb_strip_nulls(jsonb_build_object(
            'title', OLD.title,
            'assignee', OLD.assignee,
            'project', OLD.project,
            'status', OLD.status,
            'priority', OLD.priority,
            'tags', OLD.tags,
            'due_at', OLD.due_at,
            'progress', OLD.progress,
            'version', OLD.version
        ));
    END IF;
    after_payload := jsonb_strip_nulls(jsonb_build_object(
        'title', NEW.title,
        'assignee', NEW.assignee,
        'project', NEW.project,
        'status', NEW.status,
        'priority', NEW.priority,
        'tags', NEW.tags,
        'due_at', NEW.due_at,
        'progress', NEW.progress,
        'version', NEW.version,
        'source_type', NEW.source_type
    ));
    INSERT INTO task_activities (task_id, user_id, action, before_data, after_data, metadata, created_at)
    VALUES (NEW.id, NEW.user_id, activity_action, before_payload, after_payload, '{}'::JSONB, NOW());
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER tasks_record_activity
    AFTER INSERT OR UPDATE ON tasks
    FOR EACH ROW EXECUTE FUNCTION record_task_activity();
