CREATE TABLE competitor_raw_snapshots (
    id BIGSERIAL PRIMARY KEY,
    scan_id BIGINT NOT NULL REFERENCES competitor_scans(id) ON DELETE CASCADE,
    platform TEXT NOT NULL,
    raw_payload JSONB NOT NULL DEFAULT '{}'::JSONB,
    object_key TEXT NOT NULL DEFAULT '',
    captured_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX competitor_raw_snapshots_scan_captured_idx
    ON competitor_raw_snapshots (scan_id, captured_at DESC);

CREATE TABLE competitor_evidence_sources (
    id BIGSERIAL PRIMARY KEY,
    scan_id BIGINT NOT NULL REFERENCES competitor_scans(id) ON DELETE CASCADE,
    source_type TEXT NOT NULL,
    platform TEXT NOT NULL DEFAULT '',
    title TEXT NOT NULL,
    source_url TEXT NOT NULL DEFAULT '',
    summary TEXT NOT NULL DEFAULT '',
    screenshot_object_key TEXT NOT NULL DEFAULT '',
    captured_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX competitor_evidence_sources_scan_captured_idx
    ON competitor_evidence_sources (scan_id, captured_at DESC);
