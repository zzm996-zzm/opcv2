UPDATE project_opportunities
SET is_featured = false,
    heat = CASE slug
        WHEN 'excel-automation-reporting' THEN 3
        WHEN 'ai-short-video-studio' THEN 1
        ELSE 0
    END,
    updated_at = NOW()
WHERE slug IN (
    'excel-automation-reporting',
    'ai-short-video-studio',
    'local-ai-sales-consulting',
    'data-dashboard-service'
);
