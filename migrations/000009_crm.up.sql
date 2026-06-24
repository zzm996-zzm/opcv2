CREATE TABLE crm_customers (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    import_key TEXT NOT NULL,
    name TEXT NOT NULL,
    phone TEXT NOT NULL DEFAULT '',
    email TEXT NOT NULL DEFAULT '',
    website TEXT NOT NULL DEFAULT '',
    stage TEXT NOT NULL DEFAULT 'new',
    source TEXT NOT NULL DEFAULT 'lead',
    next_follow_up_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, import_key)
);

CREATE INDEX crm_customers_user_stage_idx
    ON crm_customers (user_id, stage);

CREATE INDEX crm_customers_user_next_follow_up_idx
    ON crm_customers (user_id, next_follow_up_at)
    WHERE next_follow_up_at IS NOT NULL;

CREATE TABLE crm_activities (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    customer_id BIGINT NOT NULL REFERENCES crm_customers(id) ON DELETE CASCADE,
    type TEXT NOT NULL,
    note TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX crm_activities_customer_created_idx
    ON crm_activities (customer_id, created_at DESC);

CREATE TABLE crm_followups (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    customer_id BIGINT NOT NULL REFERENCES crm_customers(id) ON DELETE CASCADE,
    note TEXT NOT NULL,
    next_follow_up_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX crm_followups_customer_created_idx
    ON crm_followups (customer_id, created_at DESC);
