ALTER TABLE community_config
    DROP COLUMN IF EXISTS qr_variants;

DROP INDEX IF EXISTS content_articles_category_published_idx;

ALTER TABLE content_articles
    DROP COLUMN IF EXISTS source_published_at,
    DROP COLUMN IF EXISTS citations,
    DROP COLUMN IF EXISTS tags,
    DROP COLUMN IF EXISTS category,
    DROP COLUMN IF EXISTS author,
    DROP COLUMN IF EXISTS source_url,
    DROP COLUMN IF EXISTS source_name;

ALTER TABLE content_tools
    DROP COLUMN IF EXISTS source_updated_at,
    DROP COLUMN IF EXISTS source_url,
    DROP COLUMN IF EXISTS limitations,
    DROP COLUMN IF EXISTS use_cases,
    DROP COLUMN IF EXISTS features,
    DROP COLUMN IF EXISTS platforms,
    DROP COLUMN IF EXISTS price_label,
    DROP COLUMN IF EXISTS provider_name;
