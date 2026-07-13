ALTER TABLE content_tools
    ADD COLUMN provider_name TEXT NOT NULL DEFAULT '',
    ADD COLUMN price_label TEXT NOT NULL DEFAULT '',
    ADD COLUMN platforms JSONB NOT NULL DEFAULT '[]'::JSONB,
    ADD COLUMN features JSONB NOT NULL DEFAULT '[]'::JSONB,
    ADD COLUMN use_cases JSONB NOT NULL DEFAULT '[]'::JSONB,
    ADD COLUMN limitations JSONB NOT NULL DEFAULT '[]'::JSONB,
    ADD COLUMN source_url TEXT NOT NULL DEFAULT '',
    ADD COLUMN source_updated_at TIMESTAMPTZ;

ALTER TABLE content_articles
    ADD COLUMN source_name TEXT NOT NULL DEFAULT '',
    ADD COLUMN source_url TEXT NOT NULL DEFAULT '',
    ADD COLUMN author TEXT NOT NULL DEFAULT '',
    ADD COLUMN category TEXT NOT NULL DEFAULT '',
    ADD COLUMN tags JSONB NOT NULL DEFAULT '[]'::JSONB,
    ADD COLUMN citations JSONB NOT NULL DEFAULT '[]'::JSONB,
    ADD COLUMN source_published_at TIMESTAMPTZ;

CREATE INDEX content_articles_category_published_idx
    ON content_articles (category, status, published_at DESC, created_at DESC);

ALTER TABLE community_config
    ADD COLUMN qr_variants JSONB NOT NULL DEFAULT '[]'::JSONB;
