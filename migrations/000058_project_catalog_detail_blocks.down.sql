UPDATE project_opportunities
SET sections = CASE slug
        WHEN 'ai-short-video-studio' THEN '[
          {"title":"成功路径","body":"从一个明确行业切入，用标准化脚本服务完成首批验证，再逐步扩展为月度内容服务。","items":["第1-2周：选择一个熟悉行业，访谈10位潜在客户","第3-4周：制作3套样板脚本并完成首个付费验证","第2个月：沉淀需求表、脚本模板和交付检查表","第3个月起：增加复购套餐和转介绍机制"]},
          {"title":"当前数据","body":"以下为演示目录数据，用于还原页面状态，不代表真实市场预测。","items":["近12个月相关搜索热度演示值：+128%","常见单次报价区间：800-5,000元","演示平均交付周期：1-3天","演示复购率：35.7%"]},
          {"title":"优劣势","body":"该项目启动成本低、交付快，但需要持续建立差异化和客户信任。","items":["优势：轻资产、标准化空间大、可按月复购","优势：AI工具可明显缩短初稿时间","短板：同质化竞争明显，案例质量决定成交率","短板：需求边界不清时容易频繁返工"]},
          {"title":"可学经验","body":"先用小范围真实交付验证方法，再把有效动作固化成模板。","items":["先聚焦一个细分行业，不做全行业通用服务","建立标准需求表和三档服务包","每次交付记录修改原因并更新检查表","用案例前后对比展示价值而非承诺流量"]},
          {"title":"要避免的行为","body":"主要风险来自定位模糊、低价竞争和缺少交付边界。","items":["一开始覆盖过多行业和平台","只用低价吸引客户，忽略服务边界","没有书面确认需求和修改次数","把演示指标当作真实收益承诺"]}
        ]'::jsonb
        WHEN 'excel-automation-reporting' THEN '[{"title":"成功路径","body":"从高频重复报表切入，先完成单流程验证。","items":["记录人工流程基线","交付可回滚模板","建立维护说明"]}]'::jsonb
        WHEN 'ai-resume-optimization' THEN '[{"title":"成功路径","body":"选择单一岗位族群建立案例和模板。","items":["限定目标岗位","建立信息采集表","人工复核AI输出"]}]'::jsonb
        WHEN 'local-pet-care-subscription' THEN '[{"title":"成功路径","body":"先在单个社区验证需求与履约。","items":["招募20个种子家庭","验证复购周期","控制库存周转"]}]'::jsonb
        WHEN 'knowledge-course-production' THEN '[{"title":"成功路径","body":"从一门最小课程开始验证付费意愿。","items":["访谈目标学员","制作首个章节","小班预售验证"]}]'::jsonb
        WHEN 'local-ai-sales-consulting' THEN '[{"title":"成功路径","body":"选择一个本地行业完成可量化的流程试点。","items":["梳理现有获客流程","限定一个试点环节","对比人工基线"]}]'::jsonb
        WHEN 'niche-travel-guide' THEN '[{"title":"成功路径","body":"先做一个目的地和一条主题路线。","items":["发布10篇深度内容","验证咨询需求","建立本地合作清单"]}]'::jsonb
        WHEN 'data-dashboard-service' THEN '[{"title":"成功路径","body":"先统一指标口径，再建设最小可用看板。","items":["确认管理问题","定义核心指标","建立更新机制"]}]'::jsonb
        ELSE sections
    END,
    updated_at = NOW()
WHERE slug IN (
    'ai-short-video-studio',
    'excel-automation-reporting',
    'ai-resume-optimization',
    'local-pet-care-subscription',
    'knowledge-course-production',
    'local-ai-sales-consulting',
    'niche-travel-guide',
    'data-dashboard-service'
);
