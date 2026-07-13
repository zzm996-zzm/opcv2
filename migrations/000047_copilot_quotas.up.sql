INSERT INTO membership_plan_quotas (plan_code, key, label, limit_value, unit, display_order)
VALUES
    ('free', 'copilot_messages', 'Copilot 对话', 30, '次/月', 50),
    ('free', 'copilot_compare_calls', 'Copilot 多模型对比', 0, '模型次/月', 60),
    ('pro', 'copilot_messages', 'Copilot 对话', 1000, '次/月', 50),
    ('pro', 'copilot_compare_calls', 'Copilot 多模型对比', 300, '模型次/月', 60)
ON CONFLICT (plan_code, key) DO UPDATE SET
    label = EXCLUDED.label,
    limit_value = EXCLUDED.limit_value,
    unit = EXCLUDED.unit,
    display_order = EXCLUDED.display_order,
    updated_at = NOW();
