ALTER TABLE membership_plans
    ADD COLUMN IF NOT EXISTS price_cents INTEGER NOT NULL DEFAULT 0 CHECK (price_cents >= 0),
    ADD COLUMN IF NOT EXISTS billing_cycle TEXT NOT NULL DEFAULT 'month',
    ADD COLUMN IF NOT EXISTS features JSONB NOT NULL DEFAULT '[]'::JSONB,
    ADD COLUMN IF NOT EXISTS recommended BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS active BOOLEAN NOT NULL DEFAULT TRUE,
    ADD COLUMN IF NOT EXISTS display_order INTEGER NOT NULL DEFAULT 0;

UPDATE membership_plans
SET price_cents = 0,
    billing_cycle = 'month',
    features = '["免费分析", "基础项目建议"]'::JSONB,
    recommended = FALSE,
    active = TRUE,
    display_order = 10
WHERE code = 'free';

UPDATE membership_plans
SET name = '会员版',
    price_cents = 6900,
    billing_cycle = 'month',
    features = '["线索数据实时更新", "去水印导出结果", "优先处理与客服支持"]'::JSONB,
    recommended = TRUE,
    active = TRUE,
    display_order = 20
WHERE code = 'pro';

CREATE TABLE membership_plan_quotas (
    id BIGSERIAL PRIMARY KEY,
    plan_code TEXT NOT NULL REFERENCES membership_plans(code) ON DELETE CASCADE,
    key TEXT NOT NULL,
    label TEXT NOT NULL,
    limit_value INTEGER NOT NULL DEFAULT 0 CHECK (limit_value >= 0),
    unit TEXT NOT NULL DEFAULT '',
    display_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (plan_code, key)
);

INSERT INTO membership_plan_quotas (plan_code, key, label, limit_value, unit, display_order)
VALUES
    ('free', 'analysis_reports', '免费分析', 3, '次/月', 10),
    ('free', 'lead_tasks', 'AI线索任务', 0, '次/月', 20),
    ('pro', 'analysis_reports', '免费分析', 100, '次/月', 10),
    ('pro', 'lead_tasks', 'AI线索任务', 30, '次/月', 20),
    ('pro', 'sandbox_runs', '商业沙盘', 20, '次/月', 30)
ON CONFLICT (plan_code, key) DO NOTHING;

CREATE TABLE membership_usage (
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    key TEXT NOT NULL,
    label TEXT NOT NULL,
    used INTEGER NOT NULL DEFAULT 0 CHECK (used >= 0),
    limit_value INTEGER NOT NULL DEFAULT 0 CHECK (limit_value >= 0),
    unit TEXT NOT NULL DEFAULT '',
    reset_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, key)
);

CREATE TABLE membership_orders (
    id BIGSERIAL PRIMARY KEY,
    order_no TEXT NOT NULL UNIQUE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    plan_code TEXT NOT NULL REFERENCES membership_plans(code),
    amount_cents INTEGER NOT NULL CHECK (amount_cents >= 0),
    status TEXT NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    paid_at TIMESTAMPTZ
);

CREATE INDEX membership_orders_user_created_idx
    ON membership_orders (user_id, created_at DESC);
