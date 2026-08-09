ALTER TABLE startup_failure_sources DROP CONSTRAINT IF EXISTS startup_failure_sources_source_kind_check;
ALTER TABLE startup_failure_sources
    ADD CONSTRAINT startup_failure_sources_source_kind_check
    CHECK (source_kind IN ('primary', 'authority', 'research', 'media', 'vertical', 'community', 'secondary'));

ALTER TABLE web_sources DROP CONSTRAINT IF EXISTS web_sources_source_kind_check;
ALTER TABLE web_sources
    ADD CONSTRAINT web_sources_source_kind_check
    CHECK (source_kind IN ('primary', 'authority', 'research', 'media', 'vertical', 'community', 'secondary'));

ALTER TABLE project_cases
    ADD COLUMN has_conflict BOOLEAN NOT NULL DEFAULT FALSE;

CREATE TABLE project_case_claims (
    id BIGSERIAL PRIMARY KEY,
    case_id BIGINT NOT NULL REFERENCES project_cases(id) ON DELETE CASCADE,
    claim_type TEXT NOT NULL CHECK (claim_type IN ('fact', 'analysis')),
    field_name TEXT NOT NULL CHECK (field_name <> ''),
    value_text TEXT NOT NULL,
    detail TEXT,
    is_model_generated BOOLEAN NOT NULL DEFAULT FALSE,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX project_case_claims_case_sort_idx
    ON project_case_claims (case_id, claim_type, sort_order, id);

CREATE TABLE project_case_claim_sources (
    claim_id BIGINT NOT NULL REFERENCES project_case_claims(id) ON DELETE CASCADE,
    web_source_id BIGINT NOT NULL REFERENCES web_sources(id) ON DELETE CASCADE,
    PRIMARY KEY (claim_id, web_source_id)
);
