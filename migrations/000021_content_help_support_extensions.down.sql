DROP TABLE IF EXISTS support_tickets;
DROP TABLE IF EXISTS help_articles;
DROP TABLE IF EXISTS help_topics;
DROP TABLE IF EXISTS community_join_requests;
DROP TABLE IF EXISTS content_article_bookmarks;
DROP TABLE IF EXISTS content_tool_favorites;

DROP INDEX IF EXISTS content_tools_category_status_idx;

ALTER TABLE content_tools
    DROP COLUMN IF EXISTS sort_weight,
    DROP COLUMN IF EXISTS tags,
    DROP COLUMN IF EXISTS category;
