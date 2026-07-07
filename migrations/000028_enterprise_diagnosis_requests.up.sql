CREATE TABLE enterprise_diagnosis_requests (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    need TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'submitted',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX enterprise_diagnosis_requests_user_created_idx
    ON enterprise_diagnosis_requests (user_id, created_at DESC);
