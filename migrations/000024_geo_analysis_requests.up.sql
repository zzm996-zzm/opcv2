CREATE TABLE geo_analysis_requests (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    target TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'queued',
    error_message TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX geo_analysis_requests_user_created_idx
    ON geo_analysis_requests (user_id, created_at DESC);

CREATE INDEX geo_analysis_requests_status_created_idx
    ON geo_analysis_requests (status, created_at ASC);
