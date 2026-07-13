CREATE TABLE enterprise_public_overview (
    id BIGSERIAL PRIMARY KEY,
    headline TEXT NOT NULL DEFAULT '',
    subheadline TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    proof_points JSONB NOT NULL DEFAULT '[]'::JSONB,
    stats JSONB NOT NULL DEFAULT '[]'::JSONB,
    service_steps JSONB NOT NULL DEFAULT '[]'::JSONB,
    source_name TEXT NOT NULL DEFAULT '',
    source_url TEXT NOT NULL DEFAULT '',
    source_updated_at TIMESTAMPTZ,
    status TEXT NOT NULL DEFAULT 'draft',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX enterprise_public_overview_status_updated_idx
    ON enterprise_public_overview (status, updated_at DESC, id DESC);

CREATE TABLE enterprise_public_cases (
    id BIGSERIAL PRIMARY KEY,
    slug TEXT NOT NULL UNIQUE,
    company TEXT NOT NULL,
    title TEXT NOT NULL,
    summary TEXT NOT NULL DEFAULT '',
    result TEXT NOT NULL DEFAULT '',
    industry TEXT NOT NULL DEFAULT '',
    services JSONB NOT NULL DEFAULT '[]'::JSONB,
    metrics JSONB NOT NULL DEFAULT '[]'::JSONB,
    body TEXT NOT NULL DEFAULT '',
    source_name TEXT NOT NULL DEFAULT '',
    source_url TEXT NOT NULL DEFAULT '',
    source_updated_at TIMESTAMPTZ,
    status TEXT NOT NULL DEFAULT 'draft',
    sort_order INTEGER NOT NULL DEFAULT 0,
    published_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX enterprise_public_cases_status_sort_idx
    ON enterprise_public_cases (status, sort_order ASC, published_at DESC, id DESC);

CREATE TABLE enterprise_contact_config (
    id BIGSERIAL PRIMARY KEY,
    consultant_name TEXT NOT NULL DEFAULT '',
    title TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    phone TEXT NOT NULL DEFAULT '',
    email TEXT NOT NULL DEFAULT '',
    wechat TEXT NOT NULL DEFAULT '',
    qr_image_url TEXT NOT NULL DEFAULT '',
    contact_url TEXT NOT NULL DEFAULT '',
    source_name TEXT NOT NULL DEFAULT '',
    crm_owner_user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    status TEXT NOT NULL DEFAULT 'draft',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX enterprise_contact_config_status_updated_idx
    ON enterprise_contact_config (status, updated_at DESC, id DESC);

CREATE TABLE enterprise_inquiries (
    id BIGSERIAL PRIMARY KEY,
    company TEXT NOT NULL DEFAULT '',
    name TEXT NOT NULL,
    phone TEXT NOT NULL DEFAULT '',
    email TEXT NOT NULL DEFAULT '',
    wechat TEXT NOT NULL DEFAULT '',
    need TEXT NOT NULL,
    budget TEXT NOT NULL DEFAULT '',
    timeline TEXT NOT NULL DEFAULT '',
    source_page TEXT NOT NULL DEFAULT '/enterprise',
    status TEXT NOT NULL DEFAULT 'submitted',
    crm_customer_id BIGINT REFERENCES crm_customers(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX enterprise_inquiries_created_idx
    ON enterprise_inquiries (created_at DESC, id DESC);

CREATE INDEX enterprise_inquiries_status_created_idx
    ON enterprise_inquiries (status, created_at DESC);
