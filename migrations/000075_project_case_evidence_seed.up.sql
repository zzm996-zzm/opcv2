INSERT INTO web_research_jobs (
    scene, ref_type, ref_id, trigger_reason, query_plan, missing_evidence,
    status, started_at, finished_at
)
SELECT
    'case', 'project_opportunity', o.id, seed.trigger_reason,
    seed.query_plan::jsonb, '[]'::jsonb, 'done', seed.finished_at, seed.finished_at
FROM (
    VALUES
        (
            'local-ai-sales-consulting',
            'seed:project-case:klarna-ai-assistant-2024',
            '["Klarna AI assistant customer service results 2024 official"]',
            '2024-02-27T12:00:00Z'::timestamptz
        ),
        (
            'excel-automation-reporting',
            'seed:project-case:coca-cola-united-power-automate',
            '["Coca-Cola United Power Automate SAP invoicing official customer story"]',
            '2021-02-04T12:00:00Z'::timestamptz
        ),
        (
            'ai-resume-optimization',
            'seed:project-case:amazon-recruiting-ai-bias',
            '["Amazon recruiting AI bias Reuters BBC"]',
            '2018-10-10T12:00:00Z'::timestamptz
        )
) AS seed(project_slug, trigger_reason, query_plan, finished_at)
JOIN project_opportunities o ON o.slug = seed.project_slug
WHERE NOT EXISTS (
    SELECT 1 FROM web_research_jobs existing
    WHERE existing.trigger_reason = seed.trigger_reason
);

INSERT INTO web_sources (
    job_id, url, canonical_url, title, publisher, source_kind,
    quality_score, published_at, fetched_at, evidence_excerpt, extracted_facts
)
SELECT
    job.id, source.url, source.url, source.title, source.publisher,
    source.source_kind, source.quality_score, source.published_at,
    '2026-08-12T00:00:00+08:00'::timestamptz,
    source.evidence_excerpt, source.extracted_facts::jsonb
FROM (
    VALUES
        (
            'seed:project-case:klarna-ai-assistant-2024',
            'https://www.klarna.com/international/press/klarna-ai-assistant-handles-two-thirds-of-customer-service-chats-in-its-first-month/',
            'Klarna AI assistant handles two-thirds of customer service chats in its first month',
            'Klarna', 'primary', 0.98::numeric,
            '2024-02-27T00:00:00Z'::timestamptz,
            'Klarna reported 2.3 million conversations, two-thirds of its customer-service chats, resolution under two minutes versus 11 minutes, and an estimated USD 40 million profit improvement in 2024.',
            '["2.3 million conversations in the first month","two-thirds of customer-service chats","work equivalent to 700 full-time agents","customer satisfaction on par with human agents","25% reduction in repeat inquiries","resolution time under 2 minutes versus 11 minutes","available in 23 markets and more than 35 languages","estimated USD 40 million profit improvement in 2024"]'
        ),
        (
            'seed:project-case:coca-cola-united-power-automate',
            'https://www.microsoft.com/en/customers/story/845187-coca-cola-bottling-company-united-consumer-goods-power-automate',
            'Coca-Cola Bottling Company United uses Microsoft Power Automate to streamline processes',
            'Microsoft', 'primary', 0.96::numeric,
            '2020-01-01T00:00:00Z'::timestamptz,
            'Microsoft reports that Coca-Cola Bottling Company United automated shipment requests and reduced order validation from half a day to seconds.',
            '["automated on-demand shipment requests","reduced order validation from half a day to seconds"]'
        ),
        (
            'seed:project-case:coca-cola-united-power-automate',
            'https://www.microsoft.com/en-us/power-platform/blog/power-apps/coca-cola-united-automates-multi-step-sap-invoicing-process-power-platform/',
            'Coca-Cola UNITED automates a multi-step SAP invoicing process with Microsoft Power Platform',
            'Microsoft Power Platform', 'primary', 0.95::numeric,
            '2021-02-04T00:00:00Z'::timestamptz,
            'Microsoft describes a multi-step SAP invoicing automation used for more than 50,000 orders in a process previously managed by one employee.',
            '["automated a multi-step SAP invoicing process","processed more than 50,000 orders","the prior process was managed by one employee"]'
        ),
        (
            'seed:project-case:amazon-recruiting-ai-bias',
            'https://www.reuters.com/article/world/insight-amazon-scraps-secret-ai-recruiting-tool-that-showed-bias-against-women-idUSKCN1MK0AG/',
            'Amazon scraps secret AI recruiting tool that showed bias against women',
            'Reuters', 'media', 0.95::numeric,
            '2018-10-10T00:00:00Z'::timestamptz,
            'Reuters reported that Amazon developed an experimental recruiting tool that scored applicants, showed bias against women, and was abandoned.',
            '["Amazon developed an experimental AI recruiting tool","the tool scored job applicants","the tool showed bias against women","Amazon abandoned the tool"]'
        ),
        (
            'seed:project-case:amazon-recruiting-ai-bias',
            'https://www.bbc.com/news/technology-45809919',
            'Amazon ditched AI recruitment tool that favoured men for technical jobs',
            'BBC News', 'media', 0.92::numeric,
            '2018-10-10T00:00:00Z'::timestamptz,
            'BBC independently reported that Amazon stopped using an experimental recruiting system after it was found to discriminate against women.',
            '["Amazon stopped using an experimental recruiting system","the system was found to discriminate against women"]'
        )
) AS source(
    trigger_reason, url, title, publisher, source_kind, quality_score,
    published_at, evidence_excerpt, extracted_facts
)
JOIN web_research_jobs job ON job.trigger_reason = source.trigger_reason
ON CONFLICT (job_id, canonical_url) DO UPDATE SET
    url = EXCLUDED.url,
    title = EXCLUDED.title,
    publisher = EXCLUDED.publisher,
    source_kind = EXCLUDED.source_kind,
    quality_score = EXCLUDED.quality_score,
    published_at = EXCLUDED.published_at,
    fetched_at = EXCLUDED.fetched_at,
    evidence_excerpt = EXCLUDED.evidence_excerpt,
    extracted_facts = EXCLUDED.extracted_facts;

INSERT INTO project_cases (
    slug, opportunity_id, title, summary, case_type, outcome,
    key_actions, lessons, pitfalls, source_title, source_url, captured_at,
    status, published_at, scale, result_summary, content_md,
    evidence_status, last_verified_at, is_locked, has_conflict
)
SELECT
    seed.slug, o.id, seed.title, seed.summary, seed.case_type, seed.outcome,
    seed.key_actions::jsonb, seed.lessons::jsonb, seed.pitfalls::jsonb,
    seed.source_title, seed.source_url, seed.captured_at,
    'published', seed.published_at, seed.scale, seed.result_summary,
    seed.content_md, 'verified', '2026-08-12T00:00:00+08:00'::timestamptz,
    false, false
FROM (
    VALUES
        (
            'klarna-ai-customer-service-2024', 'local-ai-sales-consulting',
            'Klarna：AI 客服首月处理三分之二咨询',
            'Klarna 官方披露其 AI 助手上线首月的客服运营数据。',
            'success',
            '首月处理 230 万次对话，平均解决时间由 11 分钟降至不足 2 分钟。',
            '["在客服场景上线 AI 助手","覆盖多市场与多语言","持续跟踪解决时间、重复咨询和满意度"]',
            '["应先选择指标清晰、可与人工基线比较的流程","效率指标需要与客户满意度同时观察"]',
            '["不能把单家企业披露的数据直接外推为所有行业的收益承诺"]',
            'Klarna 官方新闻稿',
            'https://www.klarna.com/international/press/klarna-ai-assistant-handles-two-thirds-of-customer-service-chats-in-its-first-month/',
            '2026-08-12T00:00:00+08:00'::timestamptz,
            '2024-02-27T00:00:00Z'::timestamptz,
            '大型企业',
            '官方披露：首月 230 万次对话、覆盖三分之二客服咨询，解决时间不足 2 分钟。',
            E'## 案例背景\n\nKlarna 在客户服务场景上线 AI 助手，并在首月公开运营结果。\n\n## 可借鉴边界\n\n本案例说明明确流程和指标有助于验证自动化价值；实际效果仍取决于业务数据、流程设计与人工兜底。'
        ),
        (
            'coca-cola-united-process-automation', 'excel-automation-reporting',
            'Coca-Cola United：订单与 SAP 发票流程自动化',
            'Microsoft 客户案例披露 Coca-Cola United 对订单校验及 SAP 发票流程的自动化实践。',
            'success',
            '订单校验由半天缩短到数秒，自动化发票流程处理超过 5 万笔订单。',
            '["从具体订单与发票流程切入","连接既有 SAP 流程","保留明确的业务负责人"]',
            '["先记录人工处理基线，再评估自动化收益","单一高频流程适合作为企业自动化试点"]',
            '["案例使用企业级平台，不能直接等同于简单 Excel 模板的实施成本"]',
            'Microsoft 客户案例',
            'https://www.microsoft.com/en/customers/story/845187-coca-cola-bottling-company-united-consumer-goods-power-automate',
            '2026-08-12T00:00:00+08:00'::timestamptz,
            '2021-02-04T00:00:00Z'::timestamptz,
            '大型企业',
            '官方案例披露：订单校验由半天降至数秒，SAP 发票自动化流程处理超过 5 万笔订单。',
            E'## 案例背景\n\nCoca-Cola Bottling Company United 使用 Microsoft Power Platform 自动化订单和发票相关流程。\n\n## 可借鉴边界\n\n该案例可用于理解流程梳理和基线测量的重要性，但具体技术栈、组织成本和收益不能直接套用到中小企业。'
        ),
        (
            'amazon-recruiting-ai-bias-2018', 'ai-resume-optimization',
            'Amazon：招聘 AI 工具因性别偏见被放弃',
            'Reuters 与 BBC 均报道 Amazon 的实验性招聘工具对女性候选人产生不利偏见，项目最终被放弃。',
            'failure',
            '实验性招聘评分工具因偏见问题未投入正式使用。',
            '["开发自动评分工具","发现对女性候选人的不利偏差","放弃该实验性系统"]',
            '["招聘辅助工具必须进行分群公平性评估","高风险决策不能只依赖历史数据训练出的自动评分"]',
            '["把历史招聘偏差当作有效模式学习","缺少充分的公平性验证与人工问责"]',
            'Reuters 调查报道',
            'https://www.reuters.com/article/world/insight-amazon-scraps-secret-ai-recruiting-tool-that-showed-bias-against-women-idUSKCN1MK0AG/',
            '2026-08-12T00:00:00+08:00'::timestamptz,
            '2018-10-10T00:00:00Z'::timestamptz,
            '大型企业',
            '两家独立媒体报道：实验性工具在候选人评分中表现出性别偏见，Amazon 随后放弃该工具。',
            E'## 案例背景\n\nAmazon 曾开发实验性 AI 招聘评分工具。Reuters 与 BBC 报道该工具表现出对女性候选人的不利偏见，并被放弃。\n\n## 可借鉴边界\n\n本案例适用于提醒简历辅助服务进行公平性审查和人工复核，不代表所有 AI 招聘工具都会产生相同结果。'
        )
) AS seed(
    slug, project_slug, title, summary, case_type, outcome, key_actions,
    lessons, pitfalls, source_title, source_url, captured_at, published_at,
    scale, result_summary, content_md
)
JOIN project_opportunities o ON o.slug = seed.project_slug
ON CONFLICT (slug) DO UPDATE SET
    opportunity_id = EXCLUDED.opportunity_id,
    title = EXCLUDED.title,
    summary = EXCLUDED.summary,
    case_type = EXCLUDED.case_type,
    outcome = EXCLUDED.outcome,
    key_actions = EXCLUDED.key_actions,
    lessons = EXCLUDED.lessons,
    pitfalls = EXCLUDED.pitfalls,
    source_title = EXCLUDED.source_title,
    source_url = EXCLUDED.source_url,
    captured_at = EXCLUDED.captured_at,
    status = EXCLUDED.status,
    published_at = EXCLUDED.published_at,
    scale = EXCLUDED.scale,
    result_summary = EXCLUDED.result_summary,
    content_md = EXCLUDED.content_md,
    evidence_status = EXCLUDED.evidence_status,
    last_verified_at = EXCLUDED.last_verified_at,
    is_locked = EXCLUDED.is_locked,
    has_conflict = EXCLUDED.has_conflict,
    updated_at = NOW();

INSERT INTO project_case_sources (case_id, web_source_id, field_name, is_primary)
SELECT c.id, ws.id, source.field_name, source.is_primary
FROM (
    VALUES
        ('klarna-ai-customer-service-2024', 'seed:project-case:klarna-ai-assistant-2024',
         'https://www.klarna.com/international/press/klarna-ai-assistant-handles-two-thirds-of-customer-service-chats-in-its-first-month/', 'case_outcome', true),
        ('coca-cola-united-process-automation', 'seed:project-case:coca-cola-united-power-automate',
         'https://www.microsoft.com/en/customers/story/845187-coca-cola-bottling-company-united-consumer-goods-power-automate', 'order_validation', true),
        ('coca-cola-united-process-automation', 'seed:project-case:coca-cola-united-power-automate',
         'https://www.microsoft.com/en-us/power-platform/blog/power-apps/coca-cola-united-automates-multi-step-sap-invoicing-process-power-platform/', 'invoice_automation', false),
        ('amazon-recruiting-ai-bias-2018', 'seed:project-case:amazon-recruiting-ai-bias',
         'https://www.reuters.com/article/world/insight-amazon-scraps-secret-ai-recruiting-tool-that-showed-bias-against-women-idUSKCN1MK0AG/', 'case_outcome', true),
        ('amazon-recruiting-ai-bias-2018', 'seed:project-case:amazon-recruiting-ai-bias',
         'https://www.bbc.com/news/technology-45809919', 'case_outcome', false)
) AS source(case_slug, trigger_reason, url, field_name, is_primary)
JOIN project_cases c ON c.slug = source.case_slug
JOIN web_research_jobs job ON job.trigger_reason = source.trigger_reason
JOIN web_sources ws ON ws.job_id = job.id AND ws.canonical_url = source.url
ON CONFLICT (case_id, web_source_id, field_name) DO UPDATE SET
    is_primary = EXCLUDED.is_primary;

DELETE FROM project_case_claims
WHERE case_id IN (
    SELECT id FROM project_cases
    WHERE slug IN (
        'klarna-ai-customer-service-2024',
        'coca-cola-united-process-automation',
        'amazon-recruiting-ai-bias-2018'
    )
);

WITH claim_seed AS (
    SELECT * FROM (
        VALUES
            ('klarna-ai-customer-service-2024', 'fact', 'conversation_volume', '首月处理 230 万次对话', '', false, 10,
             'seed:project-case:klarna-ai-assistant-2024', 'https://www.klarna.com/international/press/klarna-ai-assistant-handles-two-thirds-of-customer-service-chats-in-its-first-month/'),
            ('klarna-ai-customer-service-2024', 'fact', 'service_share', '覆盖三分之二的客服咨询', '', false, 20,
             'seed:project-case:klarna-ai-assistant-2024', 'https://www.klarna.com/international/press/klarna-ai-assistant-handles-two-thirds-of-customer-service-chats-in-its-first-month/'),
            ('klarna-ai-customer-service-2024', 'fact', 'resolution_time', '平均解决时间不足 2 分钟，此前为 11 分钟', '', false, 30,
             'seed:project-case:klarna-ai-assistant-2024', 'https://www.klarna.com/international/press/klarna-ai-assistant-handles-two-thirds-of-customer-service-chats-in-its-first-month/'),
            ('klarna-ai-customer-service-2024', 'fact', 'repeat_inquiries', '重复咨询减少 25%', '', false, 40,
             'seed:project-case:klarna-ai-assistant-2024', 'https://www.klarna.com/international/press/klarna-ai-assistant-handles-two-thirds-of-customer-service-chats-in-its-first-month/'),
            ('klarna-ai-customer-service-2024', 'analysis', 'transferable_lesson', '先用解决时间、重复咨询率和满意度共同验证客服自动化', '这是基于官方披露指标形成的模型分析，不是 Klarna 的原文结论。', true, 100,
             'seed:project-case:klarna-ai-assistant-2024', 'https://www.klarna.com/international/press/klarna-ai-assistant-handles-two-thirds-of-customer-service-chats-in-its-first-month/'),

            ('coca-cola-united-process-automation', 'fact', 'order_validation', '订单校验由半天缩短到数秒', '', false, 10,
             'seed:project-case:coca-cola-united-power-automate', 'https://www.microsoft.com/en/customers/story/845187-coca-cola-bottling-company-united-consumer-goods-power-automate'),
            ('coca-cola-united-process-automation', 'fact', 'invoice_volume', '自动化 SAP 发票流程处理超过 5 万笔订单', '', false, 20,
             'seed:project-case:coca-cola-united-power-automate', 'https://www.microsoft.com/en-us/power-platform/blog/power-apps/coca-cola-united-automates-multi-step-sap-invoicing-process-power-platform/'),
            ('coca-cola-united-process-automation', 'fact', 'prior_ownership', '该流程此前由一名员工负责', '', false, 30,
             'seed:project-case:coca-cola-united-power-automate', 'https://www.microsoft.com/en-us/power-platform/blog/power-apps/coca-cola-united-automates-multi-step-sap-invoicing-process-power-platform/'),
            ('coca-cola-united-process-automation', 'analysis', 'transferable_lesson', '高频且基线明确的单一流程适合作为自动化试点', '这是依据案例流程特征形成的模型分析，不能视为对所有企业的效果承诺。', true, 100,
             'seed:project-case:coca-cola-united-power-automate', 'https://www.microsoft.com/en/customers/story/845187-coca-cola-bottling-company-united-consumer-goods-power-automate'),

            ('amazon-recruiting-ai-bias-2018', 'fact', 'tool_purpose', 'Amazon 开发了用于候选人评分的实验性 AI 招聘工具', '', false, 10,
             'seed:project-case:amazon-recruiting-ai-bias', 'https://www.reuters.com/article/world/insight-amazon-scraps-secret-ai-recruiting-tool-that-showed-bias-against-women-idUSKCN1MK0AG/'),
            ('amazon-recruiting-ai-bias-2018', 'fact', 'bias', '工具表现出对女性候选人的不利偏见', '', false, 20,
             'seed:project-case:amazon-recruiting-ai-bias', 'https://www.reuters.com/article/world/insight-amazon-scraps-secret-ai-recruiting-tool-that-showed-bias-against-women-idUSKCN1MK0AG/'),
            ('amazon-recruiting-ai-bias-2018', 'fact', 'outcome', 'Amazon 放弃了该实验性工具', '', false, 30,
             'seed:project-case:amazon-recruiting-ai-bias', 'https://www.bbc.com/news/technology-45809919'),
            ('amazon-recruiting-ai-bias-2018', 'analysis', 'transferable_lesson', '简历辅助系统需要分群公平性测试、人工复核和可追责的最终决策流程', '这是基于失败案例风险形成的模型分析。', true, 100,
             'seed:project-case:amazon-recruiting-ai-bias', 'https://www.reuters.com/article/world/insight-amazon-scraps-secret-ai-recruiting-tool-that-showed-bias-against-women-idUSKCN1MK0AG/')
    ) AS values_table(
        case_slug, claim_type, field_name, value_text, detail,
        is_model_generated, sort_order, trigger_reason, source_url
    )
), inserted_claims AS (
    INSERT INTO project_case_claims (
        case_id, claim_type, field_name, value_text, detail,
        is_model_generated, sort_order
    )
    SELECT
        c.id, seed.claim_type, seed.field_name, seed.value_text,
        seed.detail, seed.is_model_generated, seed.sort_order
    FROM claim_seed seed
    JOIN project_cases c ON c.slug = seed.case_slug
    RETURNING id, case_id, claim_type, field_name, sort_order
)
INSERT INTO project_case_claim_sources (claim_id, web_source_id)
SELECT claim.id, ws.id
FROM inserted_claims claim
JOIN project_cases c ON c.id = claim.case_id
JOIN claim_seed seed
  ON seed.case_slug = c.slug
 AND seed.claim_type = claim.claim_type
 AND seed.field_name = claim.field_name
 AND seed.sort_order = claim.sort_order
JOIN web_research_jobs job ON job.trigger_reason = seed.trigger_reason
JOIN web_sources ws ON ws.job_id = job.id AND ws.canonical_url = seed.source_url
WHERE seed.source_url IS NOT NULL
ON CONFLICT DO NOTHING;
