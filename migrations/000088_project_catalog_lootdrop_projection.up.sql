-- The crawler mirror is canonical, while this table remains the read model used
-- by the project-market API. Keep the projection in PostgreSQL so imports and
-- translation jobs can refresh it without coupling the web service to MySQL.
ALTER TABLE project_opportunities DROP CONSTRAINT IF EXISTS project_opportunities_check;
ALTER TABLE project_opportunities
    ADD CONSTRAINT project_opportunities_check CHECK (
        (status = 'published' AND published_at IS NOT NULL)
        OR status IN ('draft', 'offline')
    );

CREATE OR REPLACE FUNCTION refresh_lootdrop_project_opportunities(p_source_ids BIGINT[] DEFAULT NULL)
RETURNS INTEGER
LANGUAGE plpgsql
AS $function$
DECLARE
    affected_rows INTEGER := 0;
BEGIN
    WITH source_rows AS (
        SELECT
            r.*,
            COALESCE(s.views, 0) AS source_views,
            COALESCE(rt.translated_payload, '{}'::JSONB) AS rebuild_translation,
            COALESCE(st.translated_payload, '{}'::JSONB) AS startup_translation
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
                NULLIF(rebuild_translation -> 'pivot_idea', 'null'::JSONB),
                NULLIF(rebuild_translation #> '{payload,pivot_idea_zh}', 'null'::JSONB),
                NULLIF(rebuild_translation #> '{payload,pivot_idea}', 'null'::JSONB),
                '{}'::JSONB
            ) AS pivot_translation
        FROM source_rows
    ), normalized_rows AS (
        SELECT
            translated_rows.*,
            COALESCE(
                NULLIF(pivot_translation ->> 'name_zh', ''),
                NULLIF(rebuild_translation ->> 'pivot_idea_name_zh', ''),
                NULLIF(pivot_translation ->> 'name', ''),
                NULLIF(pivot_idea ->> 'name', ''),
                NULLIF(name, ''),
                '未命名重建方向'
            ) AS opportunity_title,
            COALESCE(
                NULLIF(pivot_translation ->> 'concept_zh', ''),
                NULLIF(pivot_translation ->> 'concept', ''),
                '基于 ' || COALESCE(NULLIF(name, ''), '已收录项目') ||
                    ' 失败案例整理的重建方向，当前记录包含执行难度、市场潜力与可扩展性评估。'
            ) AS opportunity_summary,
            COALESCE(
                NULLIF(rebuild_translation ->> 'sector_zh', ''),
                NULLIF(rebuild_translation #>> '{payload,sector_zh}', ''),
                NULLIF(startup_translation ->> 'sector_zh', ''),
                CASE sector
                    WHEN 'Communication Services' THEN '通信服务'
                    WHEN 'Consumer' THEN '消费'
                    WHEN 'Information Technology' THEN '信息技术'
                    WHEN 'Financials' THEN '金融'
                    WHEN 'Industrials' THEN '工业'
                    WHEN 'Health Care' THEN '医疗健康'
                    WHEN 'Real Estate' THEN '房地产'
                    WHEN 'Utilities' THEN '公用事业'
                    WHEN 'Materials' THEN '材料'
                    WHEN 'Energy' THEN '能源'
                    ELSE NULLIF(sector, '')
                END,
                '未分类'
            ) AS opportunity_industry,
            COALESCE(
                NULLIF(rebuild_translation ->> 'product_type_zh', ''),
                NULLIF(rebuild_translation #>> '{payload,product_type_zh}', ''),
                CASE product_type
                    WHEN 'Marketplace' THEN '交易市场'
                    WHEN 'SaaS (B2B)' THEN '企业软件'
                    WHEN 'Social Media' THEN '社交媒体'
                    WHEN 'Mobile App' THEN '移动应用'
                    WHEN 'Financial & Fintech' THEN '金融科技'
                    WHEN 'SaaS (B2C)' THEN '消费软件'
                    WHEN 'EdTech' THEN '教育科技'
                    WHEN 'Consumer Electronics' THEN '消费电子'
                    WHEN 'Developer Tools' THEN '开发者工具'
                    WHEN 'Hardware' THEN '硬件'
                    WHEN 'Medical' THEN '医疗'
                    WHEN 'AI' THEN '人工智能'
                    WHEN 'Robotics' THEN '机器人'
                    WHEN 'IoT' THEN '物联网'
                    WHEN 'Interactive' THEN '互动产品'
                    WHEN 'Blockchain/Crypto' THEN '区块链'
                    WHEN 'Wearables' THEN '可穿戴设备'
                    WHEN 'Biotech' THEN '生物科技'
                    WHEN 'CleanTech' THEN '清洁科技'
                    WHEN 'Aerospace' THEN '航空航天'
                    WHEN 'Browser Extension' THEN '浏览器扩展'
                    WHEN 'Cybersecurity' THEN '网络安全'
                    WHEN 'Personal Care' THEN '个人护理'
                    ELSE NULLIF(product_type, '')
                END,
                '其他'
            ) AS opportunity_product_type,
            COALESCE(
                NULLIF(rebuild_translation ->> 'name_zh', ''),
                NULLIF(rebuild_translation #>> '{payload,name_zh}', ''),
                NULLIF(startup_translation ->> 'name_zh', ''),
                NULLIF(name, ''),
                '原案例'
            ) AS source_company,
            COALESCE(
                NULLIF(rebuild_translation ->> 'market_potential_zh', ''),
                NULLIF(rebuild_translation #>> '{payload,market_potential_zh}', ''),
                CASE LOWER(COALESCE(market_potential, ''))
                    WHEN 'high' THEN '高'
                    WHEN 'medium' THEN '中'
                    WHEN 'mid' THEN '中'
                    WHEN 'low' THEN '低'
                    ELSE '待补充'
                END
            ) AS opportunity_market_potential,
            CASE difficulty
                WHEN 1 THEN '较低'
                WHEN 2 THEN '较低'
                WHEN 3 THEN '中等'
                WHEN 4 THEN '较高'
                WHEN 5 THEN '高'
                ELSE '待补充'
            END AS opportunity_difficulty,
            CASE
                WHEN jsonb_typeof(
                    COALESCE(
                        NULLIF(pivot_translation -> 'mvpSteps_zh', 'null'::JSONB),
                        NULLIF(pivot_translation -> 'mvpSteps', 'null'::JSONB),
                        pivot_idea -> 'mvpSteps',
                        '[]'::JSONB
                    )
                ) = 'array'
                THEN COALESCE(
                    NULLIF(pivot_translation -> 'mvpSteps_zh', 'null'::JSONB),
                    NULLIF(pivot_translation -> 'mvpSteps', 'null'::JSONB),
                    pivot_idea -> 'mvpSteps',
                    '[]'::JSONB
                )
                ELSE '[]'::JSONB
            END AS mvp_steps,
            CASE
                WHEN jsonb_typeof(
                    COALESCE(
                        NULLIF(pivot_translation -> 'techStack_zh', 'null'::JSONB),
                        NULLIF(pivot_translation -> 'techStack', 'null'::JSONB),
                        pivot_idea -> 'techStack',
                        '[]'::JSONB
                    )
                ) = 'array'
                THEN COALESCE(
                    NULLIF(pivot_translation -> 'techStack_zh', 'null'::JSONB),
                    NULLIF(pivot_translation -> 'techStack', 'null'::JSONB),
                    pivot_idea -> 'techStack',
                    '[]'::JSONB
                )
                ELSE '[]'::JSONB
            END AS tech_stack,
            COALESCE(
                NULLIF(pivot_translation ->> 'monetization_zh', ''),
                NULLIF(pivot_translation ->> 'monetization', ''),
                NULLIF(pivot_idea ->> 'monetization', ''),
                '变现方式待补充。'
            ) AS monetization,
            COALESCE(
                NULLIF(rebuild_translation ->> 'primary_cause_of_death_zh', ''),
                NULLIF(rebuild_translation #>> '{payload,primary_cause_of_death_zh}', ''),
                NULLIF(primary_cause_of_death, ''),
                '原案例失败原因待补充。'
            ) AS failure_reason,
            CASE
                WHEN jsonb_typeof(
                    COALESCE(
                        NULLIF(rebuild_translation -> 'the_loot_zh', 'null'::JSONB),
                        NULLIF(rebuild_translation #> '{payload,the_loot_zh}', 'null'::JSONB),
                        the_loot,
                        '[]'::JSONB
                    )
                ) = 'array'
                THEN COALESCE(
                    NULLIF(rebuild_translation -> 'the_loot_zh', 'null'::JSONB),
                    NULLIF(rebuild_translation #> '{payload,the_loot_zh}', 'null'::JSONB),
                    the_loot,
                    '[]'::JSONB
                )
                ELSE '[]'::JSONB
            END AS lessons
        FROM translated_rows
    ), section_rows AS (
        SELECT
            normalized_rows.*,
            COALESCE((
                SELECT jsonb_agg(
                    jsonb_build_object(
                        'title', '执行步骤 ' || step_number,
                        'detail', step_text,
                        'meta', '来自重建方案',
                        'tone', CASE WHEN step_number = 1 THEN 'primary' ELSE 'neutral' END
                    ) ORDER BY step_number
                )
                FROM jsonb_array_elements_text(mvp_steps) WITH ORDINALITY AS steps(step_text, step_number)
            ), '[]'::JSONB) AS step_cards,
            COALESCE((
                SELECT jsonb_agg(
                    jsonb_build_object(
                        'title', '技术与资源 ' || resource_number,
                        'detail', resource_text,
                        'tone', 'neutral'
                    ) ORDER BY resource_number
                )
                FROM jsonb_array_elements_text(tech_stack) WITH ORDINALITY AS resources(resource_text, resource_number)
                WHERE resource_number <= 6
            ), '[]'::JSONB) AS resource_cards,
            COALESCE((
                SELECT jsonb_agg(lesson_text ORDER BY lesson_number)
                FROM jsonb_array_elements_text(lessons) WITH ORDINALITY AS lesson_items(lesson_text, lesson_number)
            ), '[]'::JSONB) AS lesson_items
        FROM normalized_rows
    ), projected_rows AS (
        SELECT
            section_rows.*,
            to_jsonb(array_remove(ARRAY[
                NULLIF(opportunity_industry, ''),
                NULLIF(opportunity_product_type, ''),
                CASE LOWER(COALESCE(market_potential, ''))
                    WHEN 'high' THEN '高潜力'
                    WHEN 'medium' THEN '中等潜力'
                    WHEN 'mid' THEN '中等潜力'
                    WHEN 'low' THEN '低潜力'
                    ELSE NULL
                END,
                CASE WHEN scalability >= 4 THEN '可扩展' ELSE NULL END
            ], NULL)) AS opportunity_tags,
            to_jsonb(array_remove(ARRAY[
                CASE WHEN tech_stack::TEXT ~* '(react|next\.js|node\.js|python|postgres|supabase|firebase|mobile|web|api)' THEN '软件开发' ELSE NULL END,
                CASE WHEN tech_stack::TEXT ~* '(ai|openai|claude|machine learning|pytorch|tensorflow|computer vision|vector)' THEN 'AI模型与数据' ELSE NULL END,
                CASE WHEN tech_stack::TEXT ~* '(hardware|iot|arduino|sensor|device|robot)' THEN '硬件与物联网' ELSE NULL END,
                CASE WHEN tech_stack::TEXT ~* '(aws|gcp|azure|cloudflare|vercel|cloud|kubernetes|docker)' THEN '云基础设施' ELSE NULL END,
                CASE WHEN tech_stack::TEXT ~* '(stripe|payment|billing|payout)' THEN '支付与商业化' ELSE NULL END,
                CASE WHEN jsonb_array_length(tech_stack) > 0 THEN '技术实施' ELSE NULL END
            ], NULL)) AS opportunity_resources,
            jsonb_build_array(
                jsonb_build_object(
                    'key', 'path',
                    'title', '成功路径',
                    'body', opportunity_summary,
                    'items', mvp_steps,
                    'blocks', jsonb_build_array(
                        jsonb_build_object(
                            'type', 'profile',
                            'title', '项目关键指标',
                            'columns', 4,
                            'items', jsonb_build_array(
                                jsonb_build_object('title', '启动预算', 'value', '待补充', 'detail', '原始记录未提供本地启动预算', 'tone', 'neutral'),
                                jsonb_build_object('title', '项目难度', 'value', opportunity_difficulty, 'detail', COALESCE('原始评分 ' || difficulty || '/5', '待补充'), 'tone', 'warning'),
                                jsonb_build_object('title', '可扩展性', 'value', COALESCE(scalability::TEXT || '/5', '待补充'), 'detail', '来自原始重建方案评分', 'tone', 'positive'),
                                jsonb_build_object('title', '参考案例', 'value', source_company, 'detail', '基于该失败案例整理重建方向', 'tone', 'neutral')
                            )
                        ),
                        jsonb_build_object('type', 'steps', 'title', '执行步骤', 'columns', 3, 'items', step_cards),
                        jsonb_build_object('type', 'checklist', 'title', '技术与资源', 'columns', 3, 'items', resource_cards)
                    )
                ),
                jsonb_build_object(
                    'key', 'data',
                    'title', '当前数据',
                    'body', '以下内容来自已同步的原始案例与重建方案，缺失字段保持待补充。',
                    'items', jsonb_build_array(source_company, opportunity_industry, opportunity_product_type),
                    'blocks', jsonb_build_array(jsonb_build_object(
                        'type', 'overview',
                        'title', '原始记录摘要',
                        'columns', 3,
                        'items', jsonb_build_array(
                            jsonb_build_object('title', '原案例', 'value', source_company, 'detail', '公开失败案例记录', 'tone', 'neutral'),
                            jsonb_build_object('title', '市场潜力', 'value', opportunity_market_potential, 'detail', '原始重建方案字段', 'tone', 'positive'),
                            jsonb_build_object('title', '数据更新', 'value', TO_CHAR(updated_at, 'YYYY-MM-DD'), 'detail', '镜像记录更新时间', 'tone', 'neutral')
                        )
                    ))
                ),
                jsonb_build_object(
                    'key', 'swot',
                    'title', '优劣势',
                    'body', '以下判断仅展示原始重建方案中的结构化评分，不替代实际市场验证。',
                    'items', ARRAY[opportunity_market_potential, COALESCE(scalability::TEXT || '/5', '待补充'), opportunity_difficulty],
                    'blocks', jsonb_build_array(jsonb_build_object(
                        'type', 'cards',
                        'title', '结构化评估',
                        'columns', 3,
                        'items', jsonb_build_array(
                            jsonb_build_object('title', '市场潜力', 'value', opportunity_market_potential, 'detail', '原始记录评分', 'tone', 'positive'),
                            jsonb_build_object('title', '可扩展性', 'value', COALESCE(scalability::TEXT || '/5', '待补充'), 'detail', '原始记录评分', 'tone', 'neutral'),
                            jsonb_build_object('title', '执行难度', 'value', opportunity_difficulty, 'detail', '原始记录评分', 'tone', 'warning')
                        )
                    ))
                ),
                jsonb_build_object(
                    'key', 'learning',
                    'title', '可学经验',
                    'body', '经验内容从原始失败案例的复盘字段中整理，建议结合自己的资源逐项验证。',
                    'items', lesson_items,
                    'blocks', jsonb_build_array(jsonb_build_object('type', 'cards', 'title', '原案例留下的经验', 'columns', 2, 'items', (
                        SELECT COALESCE(jsonb_agg(jsonb_build_object('title', '经验 ' || lesson_number, 'detail', lesson_text, 'tone', 'neutral') ORDER BY lesson_number), '[]'::JSONB)
                        FROM jsonb_array_elements_text(lessons) WITH ORDINALITY AS lesson_items(lesson_text, lesson_number)
                    )))
                ),
                jsonb_build_object(
                    'key', 'avoid',
                    'title', '要避免的行为',
                    'body', '原案例的主要失败原因如下，使用时不要把它当作对新项目结果的保证。',
                    'items', to_jsonb(ARRAY[failure_reason]),
                    'blocks', jsonb_build_array(jsonb_build_object(
                        'type', 'cards',
                        'title', '原案例失败原因',
                        'columns', 1,
                        'items', jsonb_build_array(jsonb_build_object('title', '主要原因', 'detail', failure_reason, 'tone', 'danger'))
                    ))
                )
            ) AS opportunity_sections
        FROM section_rows
    )
    INSERT INTO project_opportunities (
        slug, title, summary, industry, tags, budget_band, difficulty,
        resource_requirements, sections, status, sort_order, published_at,
        cover_url, category_code, track_code, invest_cents, revenue_range,
        is_real, primary_source_url, heat, is_featured, detail, updated_at
    )
    SELECT
        'lootdrop-rebuild-' || source_id,
        opportunity_title,
        opportunity_summary,
        opportunity_industry,
        opportunity_tags,
        '',
        opportunity_difficulty,
        opportunity_resources,
        opportunity_sections,
        'published',
        1000000 - COALESCE(source_id::INTEGER, 0),
        COALESCE(scraped_at, imported_at, updated_at, NOW()),
        NULL,
        NULL,
        NULL,
        NULL,
        '',
        FALSE,
        'https://www.loot-drop.io/database-view?id=' || source_id,
        LEAST(GREATEST(COALESCE(source_views, 0), 0), 2147483647)::INTEGER,
        FALSE,
        jsonb_build_object('sections', opportunity_sections, 'source_case', source_company, 'source_record_id', source_id),
        NOW()
    FROM projected_rows
    ON CONFLICT (slug) DO UPDATE SET
        title = EXCLUDED.title,
        summary = EXCLUDED.summary,
        industry = EXCLUDED.industry,
        tags = EXCLUDED.tags,
        budget_band = EXCLUDED.budget_band,
        difficulty = EXCLUDED.difficulty,
        resource_requirements = EXCLUDED.resource_requirements,
        sections = EXCLUDED.sections,
        status = EXCLUDED.status,
        sort_order = EXCLUDED.sort_order,
        published_at = EXCLUDED.published_at,
        cover_url = EXCLUDED.cover_url,
        category_code = EXCLUDED.category_code,
        track_code = EXCLUDED.track_code,
        invest_cents = EXCLUDED.invest_cents,
        revenue_range = EXCLUDED.revenue_range,
        is_real = EXCLUDED.is_real,
        primary_source_url = EXCLUDED.primary_source_url,
        heat = EXCLUDED.heat,
        is_featured = EXCLUDED.is_featured,
        detail = EXCLUDED.detail,
        updated_at = NOW();

    GET DIAGNOSTICS affected_rows = ROW_COUNT;

    IF p_source_ids IS NULL THEN
        UPDATE project_opportunities
        SET status = 'offline', updated_at = NOW()
        WHERE slug LIKE 'lootdrop-rebuild-%'
          AND NOT EXISTS (
              SELECT 1
              FROM lootdrop_rebuild_plans r
              WHERE 'lootdrop-rebuild-' || r.source_id = project_opportunities.slug
          );

        UPDATE project_opportunities
        SET status = 'offline', updated_at = NOW()
        WHERE slug IN (
            'ai-short-video-studio', 'excel-automation-reporting', 'ai-resume-optimization',
            'local-pet-care-subscription', 'knowledge-course-production',
            'local-ai-sales-consulting', 'niche-travel-guide', 'data-dashboard-service'
        );
    END IF;

    RETURN affected_rows;
END;
$function$;

SELECT refresh_lootdrop_project_opportunities();
