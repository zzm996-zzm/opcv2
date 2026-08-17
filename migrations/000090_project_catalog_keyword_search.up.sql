CREATE OR REPLACE FUNCTION project_catalog_keyword_matches(p_value TEXT, p_query TEXT)
RETURNS BOOLEAN
LANGUAGE SQL
IMMUTABLE
PARALLEL SAFE
AS $function$
    SELECT CASE
        WHEN TRIM(COALESCE(p_query, '')) = '' THEN TRUE
        WHEN TRIM(p_query) ~ '^[A-Za-z0-9 _.-]+$'
         AND REGEXP_REPLACE(LOWER(TRIM(p_query)), '[^a-z0-9]+', '', 'g') <> ''
        THEN NOT EXISTS (
            SELECT 1
            FROM REGEXP_SPLIT_TO_TABLE(LOWER(TRIM(p_query)), '[^a-z0-9]+') AS words(token)
            WHERE token <> ''
              AND NOT (
                  TO_TSVECTOR('simple', COALESCE(p_value, ''))
                  @@ TO_TSQUERY('simple', token || ':*')
              )
        )
        ELSE COALESCE(p_value, '') ILIKE '%' || TRIM(p_query) || '%'
    END
$function$;

CREATE OR REPLACE FUNCTION index_lootdrop_project_search_content(p_source_ids BIGINT[] DEFAULT NULL)
RETURNS INTEGER
LANGUAGE plpgsql
AS $function$
DECLARE
    affected_rows INTEGER := 0;
BEGIN
    WITH updated_rows AS (
        UPDATE project_opportunities p
        SET detail = jsonb_set(
                jsonb_set(
                    CASE WHEN p.detail = '{}'::JSONB THEN jsonb_build_object('sections', p.sections) ELSE p.detail END,
                    '{search_terms}',
                    to_jsonb(array_remove(ARRAY[
                        NULLIF(r.name, ''),
                        NULLIF(r.sector, ''),
                        NULLIF(r.product_type, ''),
                        NULLIF(r.country, ''),
                        NULLIF(r.pivot_idea ->> 'name', ''),
                        NULLIF(r.primary_cause_of_death, '')
                    ]::TEXT[], NULL)),
                    TRUE
                ),
                '{search_content}',
                to_jsonb(CONCAT_WS(' ',
                    NULLIF(r.pivot_idea ->> 'concept', ''),
                    NULLIF((r.pivot_idea -> 'techStack')::TEXT, ''),
                    NULLIF((r.pivot_idea -> 'mvpSteps')::TEXT, ''),
                    NULLIF(r.pivot_idea ->> 'monetization', ''),
                    NULLIF(r.the_loot::TEXT, '')
                )),
                TRUE
            ),
            updated_at = NOW()
        FROM lootdrop_rebuild_plans r
        WHERE p.slug = 'lootdrop-rebuild-' || r.source_id
          AND (p_source_ids IS NULL OR r.source_id = ANY (p_source_ids))
        RETURNING p.id
    )
    SELECT COUNT(*) INTO affected_rows FROM updated_rows;

    RETURN affected_rows;
END;
$function$;

SELECT index_lootdrop_project_search_content();

INSERT INTO project_kb_documents (version, document_id, title, body, metadata)
SELECT state.active_version, p.slug, p.title,
       concat_ws(' ', p.summary, p.industry, p.tags::TEXT, p.budget_band,
                 p.difficulty, p.resource_requirements::TEXT, p.detail::TEXT),
       jsonb_build_object('slug', p.slug, 'kind', 'project', 'knowledge_base_version', state.active_version)
FROM project_opportunities p
CROSS JOIN project_kb_state state
WHERE p.status = 'published'
ON CONFLICT (version, document_id) DO UPDATE SET
    title = EXCLUDED.title,
    body = EXCLUDED.body,
    metadata = EXCLUDED.metadata;
