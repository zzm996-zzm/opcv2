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
                CASE WHEN p.detail = '{}'::JSONB THEN jsonb_build_object('sections', p.sections) ELSE p.detail END,
                '{search_content}',
                to_jsonb(CONCAT_WS(' ',
                    NULLIF(r.pivot_idea ->> 'concept', ''),
                    NULLIF((r.pivot_idea -> 'techStack')::TEXT, ''),
                    NULLIF((r.pivot_idea -> 'mvpSteps')::TEXT, ''),
                    NULLIF(r.pivot_idea ->> 'monetization', '')
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
