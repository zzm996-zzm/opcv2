ALTER TABLE competitor_scans
    ADD COLUMN IF NOT EXISTS evidence_sources JSONB NOT NULL DEFAULT '[]'::JSONB;
