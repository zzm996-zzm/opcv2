CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    nickname VARCHAR(50) NOT NULL,
    phone VARCHAR(20) UNIQUE,
    account VARCHAR(32) UNIQUE,
    password_hash TEXT,
    wechat VARCHAR(100),
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    agreement_accepted_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX users_wechat_unique
    ON users (wechat)
    WHERE wechat IS NOT NULL AND wechat <> '';

CREATE TABLE login_records (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    ip INET,
    user_agent TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX login_records_user_created_idx
    ON login_records (user_id, created_at DESC);
