CREATE TABLE task_attachments (
    id BIGSERIAL PRIMARY KEY,
    task_id BIGINT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    comment_id BIGINT REFERENCES task_comments(id) ON DELETE SET NULL,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL CHECK (name <> ''),
    mime_type TEXT NOT NULL CHECK (mime_type <> ''),
    size_bytes BIGINT NOT NULL CHECK (size_bytes > 0 AND size_bytes <= 10485760),
    storage_key TEXT NOT NULL UNIQUE CHECK (storage_key <> ''),
    sha256 CHAR(64) NOT NULL CHECK (sha256 ~ '^[0-9a-f]{64}$'),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX task_attachments_task_created_idx
    ON task_attachments (task_id, created_at DESC, id DESC)
    WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX task_attachments_active_checksum_idx
    ON task_attachments (task_id, sha256)
    WHERE deleted_at IS NULL;
