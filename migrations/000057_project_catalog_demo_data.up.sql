INSERT INTO project_opportunities (
    slug, title, summary, industry, tags, budget_band, difficulty,
    resource_requirements, sections, status, sort_order, published_at
) VALUES
(
    'ai-short-video-studio',
    'AI短视频脚本工作室',
    '面向企业与个人 IP 的一站式短视频脚本和内容策划服务。',
    '内容服务',
    '["内容创作", "轻资产创业", "个人IP", "AI提效"]'::jsonb,
    '0.8-3万元',
    '中等',
    '["内容策划能力", "客户沟通能力", "AI写作工具", "每周可投入15小时"]'::jsonb,
    '[
      {"title":"成功路径","body":"从一个明确行业切入，用标准化脚本服务完成首批验证，再逐步扩展为月度内容服务。","items":["第1-2周：选择一个熟悉行业，访谈10位潜在客户","第3-4周：制作3套样板脚本并完成首个付费验证","第2个月：沉淀需求表、脚本模板和交付检查表","第3个月起：增加复购套餐和转介绍机制"]},
      {"title":"当前数据","body":"以下为演示目录数据，用于还原页面状态，不代表真实市场预测。","items":["近12个月相关搜索热度演示值：+128%","常见单次报价区间：800-5,000元","演示平均交付周期：1-3天","演示复购率：35.7%"]},
      {"title":"优劣势","body":"该项目启动成本低、交付快，但需要持续建立差异化和客户信任。","items":["优势：轻资产、标准化空间大、可按月复购","优势：AI工具可明显缩短初稿时间","短板：同质化竞争明显，案例质量决定成交率","短板：需求边界不清时容易频繁返工"]},
      {"title":"可学经验","body":"先用小范围真实交付验证方法，再把有效动作固化成模板。","items":["先聚焦一个细分行业，不做全行业通用服务","建立标准需求表和三档服务包","每次交付记录修改原因并更新检查表","用案例前后对比展示价值而非承诺流量"]},
      {"title":"要避免的行为","body":"主要风险来自定位模糊、低价竞争和缺少交付边界。","items":["一开始覆盖过多行业和平台","只用低价吸引客户，忽略服务边界","没有书面确认需求和修改次数","把演示指标当作真实收益承诺"]}
    ]'::jsonb,
    'published', 10, NOW()
),
(
    'excel-automation-reporting',
    'Excel自动化报表定制',
    '为中小企业把重复报表流程改造成可维护的自动化模板。',
    '企业服务',
    '["办公效率", "低成本启动", "可复制"]'::jsonb,
    '0.5-2万元',
    '中等',
    '["Excel函数与透视表", "业务流程梳理", "基础脚本能力"]'::jsonb,
    '[{"title":"成功路径","body":"从高频重复报表切入，先完成单流程验证。","items":["记录人工流程基线","交付可回滚模板","建立维护说明"]}]'::jsonb,
    'published', 20, NOW()
),
(
    'ai-resume-optimization',
    'AI智能简历优化服务',
    '结合岗位要求与求职者经历，提供结构化简历诊断和表达优化。',
    '职业服务',
    '["AI应用", "一人公司", "轻服务"]'::jsonb,
    '0.3-1万元',
    '较低',
    '["招聘或HR经验", "结构化表达", "隐私保护流程"]'::jsonb,
    '[{"title":"成功路径","body":"选择单一岗位族群建立案例和模板。","items":["限定目标岗位","建立信息采集表","人工复核AI输出"]}]'::jsonb,
    'published', 30, NOW()
),
(
    'local-pet-care-subscription',
    '宠物营养定制订阅',
    '为本地养宠家庭提供周期性的营养建议与到家组合服务。',
    '本地生活',
    '["订阅服务", "本地获客", "复购"]'::jsonb,
    '1-5万元',
    '中等',
    '["本地社群", "供应链管理", "宠物营养顾问"]'::jsonb,
    '[{"title":"成功路径","body":"先在单个社区验证需求与履约。","items":["招募20个种子家庭","验证复购周期","控制库存周转"]}]'::jsonb,
    'published', 40, NOW()
),
(
    'knowledge-course-production',
    '知识付费课程制作',
    '帮助专家把线下经验整理为可销售、可持续更新的线上课程。',
    '知识服务',
    '["内容产品", "高毛利", "可复用交付"]'::jsonb,
    '0.8-3万元',
    '中等',
    '["课程设计", "内容编辑", "基础拍摄能力"]'::jsonb,
    '[{"title":"成功路径","body":"从一门最小课程开始验证付费意愿。","items":["访谈目标学员","制作首个章节","小班预售验证"]}]'::jsonb,
    'published', 50, NOW()
),
(
    'local-ai-sales-consulting',
    '本地AI获客顾问',
    '为本地商家搭建轻量的内容、线索和客户跟进流程。',
    '营销服务',
    '["B端服务", "AI应用", "低成本启动"]'::jsonb,
    '1-3万元',
    '中等',
    '["销售沟通", "本地商家资源", "流程搭建能力"]'::jsonb,
    '[{"title":"成功路径","body":"选择一个本地行业完成可量化的流程试点。","items":["梳理现有获客流程","限定一个试点环节","对比人工基线"]}]'::jsonb,
    'published', 60, NOW()
),
(
    'niche-travel-guide',
    '小众旅行攻略平台',
    '围绕特定目的地提供深度攻略、路线设计和本地服务连接。',
    '旅行服务',
    '["内容电商", "小众市场", "社群"]'::jsonb,
    '1-4万元',
    '较高',
    '["目的地经验", "内容运营", "本地合作资源"]'::jsonb,
    '[{"title":"成功路径","body":"先做一个目的地和一条主题路线。","items":["发布10篇深度内容","验证咨询需求","建立本地合作清单"]}]'::jsonb,
    'published', 70, NOW()
),
(
    'data-dashboard-service',
    '经营数据看板服务',
    '为小团队整合销售、投放和交付数据，形成可持续维护的经营看板。',
    '数据服务',
    '["企业服务", "数据分析", "长期服务"]'::jsonb,
    '1-5万元',
    '较高',
    '["数据分析", "指标体系", "客户培训能力"]'::jsonb,
    '[{"title":"成功路径","body":"先统一指标口径，再建设最小可用看板。","items":["确认管理问题","定义核心指标","建立更新机制"]}]'::jsonb,
    'published', 80, NOW()
)
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
    updated_at = NOW();

INSERT INTO project_cases (
    slug, opportunity_id, title, summary, case_type, outcome,
    key_actions, lessons, pitfalls, source_title, source_url,
    captured_at, status, published_at
) VALUES
(
    'short-video-first-client',
    (SELECT id FROM project_opportunities WHERE slug = 'ai-short-video-studio'),
    '从三个样板脚本拿到首个客户',
    '演示案例：聚焦餐饮行业，用三个样板脚本完成首轮需求验证。',
    'success',
    '完成首个付费交付并沉淀需求模板',
    '["限定餐饮行业", "先做样板再报价", "记录每轮修改原因"]'::jsonb,
    '["细分定位比扩大服务范围更重要", "标准化需求表能显著减少返工"]'::jsonb,
    '["不要承诺未验证的播放量"]'::jsonb,
    '演示案例来源', 'https://example.com/opcv2/demo/short-video-first-client',
    NOW(), 'published', NOW()
),
(
    'short-video-scope-failure',
    (SELECT id FROM project_opportunities WHERE slug = 'ai-short-video-studio'),
    '需求边界不清导致连续返工',
    '演示案例：未约定修改次数和验收口径，项目交付周期失控。',
    'failure',
    '交付延期，项目毛利被多轮修改消耗',
    '["复盘沟通记录", "补充书面验收标准"]'::jsonb,
    '["报价前必须明确交付范围"]'::jsonb,
    '["口头确认需求", "无限次修改", "只用低价成交"]'::jsonb,
    '演示案例来源', 'https://example.com/opcv2/demo/short-video-scope-failure',
    NOW(), 'published', NOW()
),
(
    'excel-monthly-report-pilot',
    (SELECT id FROM project_opportunities WHERE slug = 'excel-automation-reporting'),
    '月报流程从两天缩短到两小时',
    '演示案例：先记录人工流程，再逐步替换重复复制和汇总步骤。',
    'success',
    '完成单一月报流程自动化验证',
    '["记录人工基线", "拆分数据清洗步骤", "保留人工复核"]'::jsonb,
    '["先验证一张报表比一次改造全部流程更稳妥"]'::jsonb,
    '["忽略异常数据处理"]'::jsonb,
    '演示案例来源', 'https://example.com/opcv2/demo/excel-monthly-report-pilot',
    NOW(), 'published', NOW()
),
(
    'resume-privacy-review',
    (SELECT id FROM project_opportunities WHERE slug = 'ai-resume-optimization'),
    '简历优化服务的隐私流程改造',
    '演示案例：对敏感信息脱敏，并要求人工复核所有AI建议。',
    'success',
    '建立可复用的隐私和质量检查清单',
    '["采集前告知用途", "敏感信息脱敏", "交付前人工复核"]'::jsonb,
    '["隐私说明和质量边界是服务信任的基础"]'::jsonb,
    '["直接上传完整个人信息到未知工具"]'::jsonb,
    '演示案例来源', 'https://example.com/opcv2/demo/resume-privacy-review',
    NOW(), 'published', NOW()
)
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
    updated_at = NOW();
