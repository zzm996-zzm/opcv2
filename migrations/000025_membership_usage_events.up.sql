INSERT INTO membership_plan_quotas (plan_code, key, label, limit_value, unit, display_order)
VALUES
    ('free', 'sandbox_runs', '商业沙盘', 1, '次/月', 30),
    ('free', 'competitor_scans', '竞品全盘数据破解', 5, '次/月', 40),
    ('pro', 'competitor_scans', '竞品全盘数据破解', 200, '次/月', 40)
ON CONFLICT (plan_code, key) DO UPDATE SET
    label = EXCLUDED.label,
    limit_value = EXCLUDED.limit_value,
    unit = EXCLUDED.unit,
    display_order = EXCLUDED.display_order,
    updated_at = NOW();

CREATE TABLE membership_usage_events (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    key TEXT NOT NULL,
    idempotency_key TEXT NOT NULL,
    amount INTEGER NOT NULL CHECK (amount > 0),
    used_after INTEGER NOT NULL CHECK (used_after >= 0),
    reset_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, key, idempotency_key)
);

CREATE INDEX membership_usage_events_user_created_idx
    ON membership_usage_events (user_id, created_at DESC);
