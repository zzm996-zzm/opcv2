INSERT INTO membership_plan_quotas (plan_code, key, label, limit_value, unit, display_order)
VALUES
    ('free', 'copilot_file_analysis', 'Copilot 文件分析', 3, '个/月', 70),
    ('pro', 'copilot_file_analysis', 'Copilot 文件分析', 100, '个/月', 70)
ON CONFLICT (plan_code, key) DO UPDATE SET
    label = EXCLUDED.label,
    limit_value = EXCLUDED.limit_value,
    unit = EXCLUDED.unit,
    display_order = EXCLUDED.display_order,
    updated_at = NOW();
