CREATE TABLE project_match_files (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    match_id BIGINT REFERENCES project_match_sessions(id) ON DELETE SET NULL,
    original_name TEXT NOT NULL CHECK (original_name <> ''),
    mime_type TEXT NOT NULL CHECK (mime_type <> ''),
    detected_mime TEXT NOT NULL CHECK (detected_mime <> ''),
    size_bytes BIGINT NOT NULL CHECK (size_bytes > 0 AND size_bytes <= 20971520),
    storage_key TEXT NOT NULL UNIQUE CHECK (storage_key <> ''),
    sha256 CHAR(64) NOT NULL CHECK (sha256 ~ '^[0-9a-f]{64}$'),
    parse_status TEXT NOT NULL DEFAULT 'uploading'
        CHECK (parse_status IN ('uploading', 'scanning', 'parsing', 'ready', 'failed', 'deleted')),
    extracted_text TEXT,
    extracted_json JSONB NOT NULL DEFAULT '{}'::JSONB CHECK (jsonb_typeof(extracted_json) = 'object'),
    error_code TEXT,
    expires_at TIMESTAMPTZ NOT NULL DEFAULT NOW() + INTERVAL '30 days',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX project_match_files_user_created_idx
    ON project_match_files (user_id, created_at DESC, id DESC)
    WHERE parse_status <> 'deleted';

CREATE INDEX project_match_files_match_idx
    ON project_match_files (match_id, id)
    WHERE match_id IS NOT NULL AND parse_status <> 'deleted';

CREATE UNIQUE INDEX project_match_files_active_checksum_idx
    ON project_match_files (user_id, sha256)
    WHERE parse_status <> 'deleted';
