CREATE TABLE project_match_evidence (
    id BIGSERIAL PRIMARY KEY,
    match_id BIGINT NOT NULL REFERENCES project_match_sessions(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    attempt INTEGER NOT NULL CHECK (attempt > 0),
    source_type TEXT NOT NULL CHECK (source_type IN ('knowledge_base', 'web')),
    source_id TEXT,
    source_url TEXT NOT NULL DEFAULT '',
    title TEXT NOT NULL CHECK (title <> ''),
    publisher TEXT NOT NULL DEFAULT '',
    excerpt TEXT NOT NULL,
    quality_score NUMERIC(4,3) NOT NULL CHECK (quality_score BETWEEN 0 AND 1),
    untrusted_content BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (match_id, attempt, source_type, source_url, source_id)
);

CREATE INDEX project_match_evidence_match_attempt_idx
    ON project_match_evidence (match_id, user_id, attempt, quality_score DESC, id);
