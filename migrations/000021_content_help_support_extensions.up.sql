ALTER TABLE content_tools
    ADD COLUMN IF NOT EXISTS category TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS tags JSONB NOT NULL DEFAULT '[]'::JSONB,
    ADD COLUMN IF NOT EXISTS sort_weight INTEGER NOT NULL DEFAULT 0;

CREATE INDEX IF NOT EXISTS content_tools_category_status_idx
    ON content_tools (category, status, sort_weight DESC, created_at DESC);

CREATE TABLE content_tool_favorites (
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tool_slug TEXT NOT NULL REFERENCES content_tools(slug) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, tool_slug)
);

CREATE TABLE content_article_bookmarks (
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    article_slug TEXT NOT NULL REFERENCES content_articles(slug) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, article_slug)
);

CREATE TABLE community_join_requests (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    community TEXT NOT NULL,
    contact TEXT NOT NULL,
    note TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'submitted',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX community_join_requests_user_created_idx
    ON community_join_requests (user_id, created_at DESC);

CREATE TABLE help_topics (
    key TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    display_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE help_articles (
    id BIGSERIAL PRIMARY KEY,
    slug TEXT NOT NULL UNIQUE,
    topic TEXT NOT NULL REFERENCES help_topics(key) ON DELETE RESTRICT,
    title TEXT NOT NULL,
    summary TEXT NOT NULL DEFAULT '',
    body TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'published',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX help_articles_topic_status_idx
    ON help_articles (topic, status, created_at DESC);

INSERT INTO help_topics (key, name, display_order)
VALUES
    ('account', '账号与安全', 10),
    ('membership', '套餐与额度', 20),
    ('projects', '项目超市', 30),
    ('sandbox', '商业沙盘', 40),
    ('competitor', '数据破解', 50),
    ('growth', '增长测算', 60)
ON CONFLICT (key) DO NOTHING;

CREATE TABLE support_tickets (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    topic TEXT NOT NULL DEFAULT '',
    title TEXT NOT NULL,
    body TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'open',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX support_tickets_user_created_idx
    ON support_tickets (user_id, created_at DESC);
