CREATE TABLE competitor_script_accounts (
    id BIGSERIAL PRIMARY KEY,
    platform TEXT NOT NULL,
    account_label TEXT NOT NULL,
    credential_ref TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'available'
        CHECK (status IN ('available', 'in_use', 'cooldown', 'disabled')),
    cooldown_until TIMESTAMPTZ,
    failure_count INTEGER NOT NULL DEFAULT 0 CHECK (failure_count >= 0),
    max_runs_per_hour INTEGER NOT NULL DEFAULT 10 CHECK (max_runs_per_hour > 0),
    last_used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (platform, account_label)
);

CREATE INDEX competitor_script_accounts_availability_idx
    ON competitor_script_accounts (platform, status, cooldown_until, last_used_at);

CREATE TABLE competitor_script_account_runs (
    id BIGSERIAL PRIMARY KEY,
    account_id BIGINT NOT NULL REFERENCES competitor_script_accounts(id) ON DELETE CASCADE,
    scan_id BIGINT NOT NULL REFERENCES competitor_scans(id) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'running'
        CHECK (status IN ('running', 'succeeded', 'failed')),
    error_code TEXT NOT NULL DEFAULT '',
    started_at TIMESTAMPTZ NOT NULL,
    finished_at TIMESTAMPTZ
);

CREATE INDEX competitor_script_account_runs_frequency_idx
    ON competitor_script_account_runs (account_id, started_at DESC);
