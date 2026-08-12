ALTER TABLE project_import_batches
    ADD COLUMN failed_count INTEGER NOT NULL DEFAULT 0 CHECK (failed_count >= 0),
    ADD COLUMN validation_errors JSONB NOT NULL DEFAULT '[]'::JSONB CHECK (jsonb_typeof(validation_errors) = 'array'),
    ADD COLUMN published_at TIMESTAMPTZ,
    ADD COLUMN rolled_back_at TIMESTAMPTZ;

CREATE TABLE project_operation_audits (
    id BIGSERIAL PRIMARY KEY,
    operator_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    action TEXT NOT NULL CHECK (action <> ''),
    target_type TEXT NOT NULL CHECK (target_type <> ''),
    target_id TEXT NOT NULL CHECK (target_id <> ''),
    detail JSONB NOT NULL DEFAULT '{}'::JSONB CHECK (jsonb_typeof(detail) = 'object'),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX project_operation_audits_created_idx
    ON project_operation_audits (created_at DESC, id DESC);
CREATE INDEX project_operation_audits_target_idx
    ON project_operation_audits (target_type, target_id, created_at DESC);
