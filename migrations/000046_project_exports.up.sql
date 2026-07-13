CREATE TABLE project_exports (
 id BIGSERIAL PRIMARY KEY, user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
 source_type TEXT NOT NULL CHECK (source_type IN ('match','comparison')), source_id BIGINT NOT NULL CHECK (source_id > 0),
 status TEXT NOT NULL DEFAULT 'ready' CHECK (status IN ('ready','failed')), payload JSONB NOT NULL, created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX project_exports_user_created_idx ON project_exports (user_id, created_at DESC);
