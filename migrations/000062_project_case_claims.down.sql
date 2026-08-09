DROP TABLE IF EXISTS project_case_claim_sources;
DROP TABLE IF EXISTS project_case_claims;

ALTER TABLE project_cases DROP COLUMN IF EXISTS has_conflict;

ALTER TABLE web_sources DROP CONSTRAINT IF EXISTS web_sources_source_kind_check;
ALTER TABLE web_sources
    ADD CONSTRAINT web_sources_source_kind_check
    CHECK (source_kind IN ('primary', 'authority', 'research', 'media', 'vertical', 'community'));

ALTER TABLE startup_failure_sources DROP CONSTRAINT IF EXISTS startup_failure_sources_source_kind_check;
ALTER TABLE startup_failure_sources
    ADD CONSTRAINT startup_failure_sources_source_kind_check
    CHECK (source_kind IN ('primary', 'authority', 'research', 'media', 'vertical', 'community'));
