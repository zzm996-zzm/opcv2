CREATE TABLE task_comments (
    id BIGSERIAL PRIMARY KEY,
    task_id BIGINT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    parent_comment_id BIGINT REFERENCES task_comments(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT task_comments_content_check CHECK (char_length(btrim(content)) BETWEEN 1 AND 2000)
);

CREATE INDEX task_comments_task_created_idx
    ON task_comments (task_id, created_at DESC, id DESC)
    WHERE deleted_at IS NULL;

CREATE INDEX task_comments_user_created_idx
    ON task_comments (user_id, created_at DESC, id DESC)
    WHERE deleted_at IS NULL;
