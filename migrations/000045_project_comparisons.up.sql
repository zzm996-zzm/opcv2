CREATE TABLE project_comparisons (
 id BIGSERIAL PRIMARY KEY, user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
 items JSONB NOT NULL CHECK (jsonb_typeof(items) = 'array' AND jsonb_array_length(items) BETWEEN 2 AND 4),
 created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX project_comparisons_user_created_idx ON project_comparisons (user_id, created_at DESC);
