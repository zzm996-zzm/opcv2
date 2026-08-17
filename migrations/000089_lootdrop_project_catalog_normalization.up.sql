-- Keep the project catalog searchable even when the optional AI translation
-- queue is unavailable. All labels in this projection are derived from the
-- canonical Loot Drop mirror or from a completed translation row.
CREATE OR REPLACE FUNCTION normalize_lootdrop_project_opportunities(p_source_ids BIGINT[] DEFAULT NULL)
RETURNS INTEGER
LANGUAGE plpgsql
AS $function$
DECLARE
    affected_rows INTEGER := 0;
BEGIN
    WITH source_rows AS (
        SELECT
            r.*,
            COALESCE(s.description, '') AS startup_description,
            COALESCE(rt.translated_payload, '{}'::JSONB) AS rebuild_translation,
            COALESCE(st.translated_payload, '{}'::JSONB) AS startup_translation,
            rt.status AS rebuild_translation_status,
            st.status AS startup_translation_status,
            LOWER(CONCAT_WS(' ',
                r.name,
                r.sector,
                r.product_type,
                r.country,
                r.primary_cause_of_death,
                r.market_potential,
                r.pivot_idea::TEXT
            )) AS source_text
        FROM lootdrop_rebuild_plans r
        LEFT JOIN lootdrop_startups s ON s.source_id = r.source_id
        LEFT JOIN lootdrop_translations rt
            ON rt.dataset = 'rebuild_plan'
           AND rt.record_key = r.source_id::TEXT
           AND rt.language = 'zh-CN'
           AND rt.status = 'completed'
        LEFT JOIN lootdrop_translations st
            ON st.dataset = 'startup'
           AND st.record_key = r.source_id::TEXT
           AND st.language = 'zh-CN'
           AND st.status = 'completed'
        WHERE p_source_ids IS NULL OR r.source_id = ANY (p_source_ids)
    ), translated_rows AS (
        SELECT
            source_rows.*,
            COALESCE(
                NULLIF(rebuild_translation -> 'pivot_idea_zh', 'null'::JSONB),
                NULLIF(rebuild_translation #> '{payload,pivot_idea_zh}', 'null'::JSONB),
                '{}'::JSONB
            ) AS pivot_translation
        FROM source_rows
    ), classified_rows AS (
        SELECT
            translated_rows.*,
            COALESCE(
                NULLIF(pivot_translation ->> 'name_zh', ''),
                NULLIF(pivot_idea ->> 'name', ''),
                NULLIF(name, ''),
                '未命名重建方向'
            ) AS normalized_title,
            COALESCE(
                NULLIF(pivot_translation ->> 'concept_zh', ''),
                NULLIF(pivot_idea ->> 'concept', ''),
                NULLIF(startup_translation ->> 'description_zh', ''),
                NULLIF(startup_description, ''),
                '原始记录未提供项目摘要。'
            ) AS normalized_summary,
            CASE LOWER(TRIM(sector))
                WHEN 'communication services' THEN '通信服务'
                WHEN 'consumer' THEN '消费'
                WHEN 'information technology' THEN '信息技术'
                WHEN 'financials' THEN '金融'
                WHEN 'industrials' THEN '工业'
                WHEN 'health care' THEN '医疗健康'
                WHEN 'real estate' THEN '房地产'
                WHEN 'utilities' THEN '公用事业'
                WHEN 'materials' THEN '材料'
                WHEN 'energy' THEN '能源'
                WHEN 'healthcare' THEN '医疗健康'
                ELSE COALESCE(NULLIF(sector, ''), '未分类')
            END AS normalized_industry,
            CASE LOWER(TRIM(product_type))
                WHEN 'marketplace' THEN '交易市场'
                WHEN 'saas (b2b)' THEN '企业软件'
                WHEN 'saas (b2c)' THEN '消费软件'
                WHEN 'social media' THEN '社交媒体'
                WHEN 'mobile app' THEN '移动应用'
                WHEN 'financial & fintech' THEN '金融科技'
                WHEN 'edtech' THEN '教育科技'
                WHEN 'consumer electronics' THEN '消费电子'
                WHEN 'developer tools' THEN '开发者工具'
                WHEN 'hardware' THEN '硬件'
                WHEN 'medical' THEN '医疗'
                WHEN 'ai' THEN '人工智能'
                WHEN 'robotics' THEN '机器人'
                WHEN 'iot' THEN '物联网'
                WHEN 'interactive' THEN '互动产品'
                WHEN 'blockchain/crypto' THEN '区块链'
                WHEN 'wearables' THEN '可穿戴设备'
                WHEN 'biotech' THEN '生物科技'
                WHEN 'cleantech' THEN '清洁科技'
                WHEN 'aerospace' THEN '航空航天'
                WHEN 'browser extension' THEN '浏览器扩展'
                WHEN 'cybersecurity' THEN '网络安全'
                WHEN 'personal care' THEN '个人护理'
                ELSE COALESCE(NULLIF(product_type, ''), '其他')
            END AS normalized_product_type,
            CASE difficulty
                WHEN 1 THEN '较低'
                WHEN 2 THEN '较低'
                WHEN 3 THEN '中等'
                WHEN 4 THEN '较高'
                WHEN 5 THEN '高'
                ELSE '待补充'
            END AS normalized_difficulty,
            CASE LOWER(TRIM(market_potential))
                WHEN 'high' THEN '高潜力'
                WHEN 'medium' THEN '中等潜力'
                WHEN 'mid' THEN '中等潜力'
                WHEN 'low' THEN '低潜力'
                ELSE NULL
            END AS normalized_potential
        FROM translated_rows
    ), normalized_rows AS (
        SELECT
            classified_rows.*,
            to_jsonb(array_remove(ARRAY[
                CASE WHEN source_text ~* '(\mreact\M|next\.js|node\.js|python|postgres|supabase|firebase|mobile|web|api|saas|software)' THEN '软件开发' END,
                CASE WHEN source_text ~* '(\mai\M|openai|claude|machine learning|pytorch|tensorflow|computer vision|rag|embedding|llm)' THEN 'AI模型与数据' END,
                CASE WHEN source_text ~* '(hardware|iot|arduino|sensor|device|robot|manufactur)' THEN '硬件与物联网' END,
                CASE WHEN source_text ~* '(aws|gcp|azure|cloudflare|vercel|cloud|kubernetes|docker|infrastructure)' THEN '云基础设施' END,
                CASE WHEN source_text ~* '(stripe|payment|billing|payout|subscription|monetiz)' THEN '支付与商业化' END,
                CASE WHEN source_text ~* '(sales|marketing|influencer|partnership|distribution|channel|customer)' THEN '获客与渠道' END,
                CASE WHEN source_text ~* '(medical|health|agri|agricultur|legal|education|finance|fashion|real estate)' THEN '领域专业知识' END
            ]::TEXT[], NULL)) AS normalized_resources,
            to_jsonb(array_remove(ARRAY[
                NULLIF(normalized_industry, ''),
                NULLIF(normalized_product_type, ''),
                normalized_potential,
                CASE WHEN scalability >= 4 THEN '可扩展' END,
                CASE WHEN difficulty <= 2 THEN '低门槛' END,
                CASE WHEN source_text ~* '(\mai\M|openai|claude|machine learning|pytorch|tensorflow|computer vision|llm)' THEN 'AI应用' END,
                CASE WHEN source_text ~* '(subscription|saas|recurring)' THEN '订阅模式' END,
                CASE WHEN source_text ~* '(b2b|enterprise|manufacturer|fleet)' THEN '企业服务' END
            ]::TEXT[], NULL)) AS normalized_tags,
            to_jsonb(array_remove(ARRAY[
                NULLIF(name, ''),
                NULLIF(sector, ''),
                NULLIF(product_type, ''),
                NULLIF(country, ''),
                NULLIF(pivot_idea ->> 'name', ''),
                NULLIF(pivot_idea ->> 'concept', ''),
                NULLIF(primary_cause_of_death, ''),
                NULLIF((pivot_idea -> 'techStack')::TEXT, ''),
                NULLIF((pivot_idea -> 'mvpSteps')::TEXT, '')
            ]::TEXT[], NULL)) AS normalized_search_terms,
            CASE
                WHEN rebuild_translation_status = 'completed' THEN 'translated'
                WHEN startup_translation_status = 'completed' THEN 'partial'
                ELSE 'original'
            END AS normalized_translation_status
        FROM classified_rows
    ), updated_rows AS (
        UPDATE project_opportunities p
        SET title = n.normalized_title,
            summary = n.normalized_summary,
            industry = n.normalized_industry,
            tags = n.normalized_tags,
            difficulty = n.normalized_difficulty,
            resource_requirements = n.normalized_resources,
            detail = jsonb_set(
                jsonb_set(
                    jsonb_set(
                        CASE WHEN p.detail = '{}'::JSONB THEN jsonb_build_object('sections', p.sections) ELSE p.detail END,
                        '{search_terms}', n.normalized_search_terms, TRUE
                    ),
                    '{source_record_id}', to_jsonb(n.source_id), TRUE
                ),
                '{translation_status}', to_jsonb(n.normalized_translation_status), TRUE
            ),
            updated_at = NOW()
        FROM normalized_rows n
        WHERE p.slug = 'lootdrop-rebuild-' || n.source_id
        RETURNING p.id
    )
    SELECT COUNT(*) INTO affected_rows FROM updated_rows;

    RETURN affected_rows;
END;
$function$;

SELECT normalize_lootdrop_project_opportunities();

-- Keep the knowledge-base snapshot aligned with the normalized search terms.
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
