INSERT INTO membership_plan_quotas (plan_code, key, label, limit_value, unit, display_order)
VALUES
    ('free', 'ai_chat', 'AI 匹配与诊断', 3, '次/月', 80),
    ('pro', 'ai_chat', 'AI 匹配与诊断', 100, '次/月', 80)
ON CONFLICT (plan_code, key) DO UPDATE SET
    label = EXCLUDED.label,
    limit_value = EXCLUDED.limit_value,
    unit = EXCLUDED.unit,
    display_order = EXCLUDED.display_order,
    updated_at = NOW();
