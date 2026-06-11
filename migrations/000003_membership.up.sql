CREATE TABLE membership_plans (
    code TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    monthly_analysis_limit INTEGER NOT NULL CHECK (monthly_analysis_limit >= 0),
    lead_export_limit INTEGER NOT NULL CHECK (lead_export_limit >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO membership_plans (code, name, monthly_analysis_limit, lead_export_limit)
VALUES
    ('free', '免费版', 3, 0),
    ('pro', '专业版', 100, 1000)
ON CONFLICT (code) DO NOTHING;

CREATE TABLE user_subscriptions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    plan_code TEXT NOT NULL REFERENCES membership_plans(code),
    status TEXT NOT NULL DEFAULT 'active',
    starts_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ends_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX user_subscriptions_user_unique
    ON user_subscriptions (user_id);

CREATE TABLE credit_accounts (
    user_id BIGINT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    balance INTEGER NOT NULL DEFAULT 0 CHECK (balance >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE credit_transactions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    amount INTEGER NOT NULL,
    reason TEXT NOT NULL,
    reference_type TEXT,
    reference_id BIGINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX credit_transactions_reference_unique
    ON credit_transactions (user_id, reference_type, reference_id)
    WHERE reference_type IS NOT NULL AND reference_id IS NOT NULL;

CREATE TABLE redemption_codes (
    id BIGSERIAL PRIMARY KEY,
    code TEXT NOT NULL UNIQUE,
    plan_code TEXT REFERENCES membership_plans(code),
    credit_amount INTEGER NOT NULL DEFAULT 0 CHECK (credit_amount >= 0),
    max_redemptions INTEGER NOT NULL DEFAULT 1 CHECK (max_redemptions >= 0),
    redeemed_count INTEGER NOT NULL DEFAULT 0 CHECK (redeemed_count >= 0),
    expires_at TIMESTAMPTZ,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (plan_code IS NOT NULL OR credit_amount > 0)
);

CREATE TABLE redemptions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    redemption_code_id BIGINT NOT NULL REFERENCES redemption_codes(id) ON DELETE RESTRICT,
    code TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, redemption_code_id)
);

CREATE INDEX credit_transactions_user_created_idx
    ON credit_transactions (user_id, created_at DESC);

CREATE INDEX redemptions_user_created_idx
    ON redemptions (user_id, created_at DESC);
