DROP TABLE IF EXISTS brand_cases;
DROP TABLE IF EXISTS brand_metrics;
DROP TABLE IF EXISTS community_config;
DROP TABLE IF EXISTS content_tools;
DROP TABLE IF EXISTS content_articles;

DROP INDEX IF EXISTS users_role_idx;

ALTER TABLE users
    DROP COLUMN IF EXISTS role;
