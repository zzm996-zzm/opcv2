UPDATE project_opportunities
SET is_featured = seed.is_featured,
    heat = GREATEST(project_opportunities.heat, seed.heat),
    updated_at = NOW()
FROM (
    VALUES
        ('excel-automation-reporting', true, 96),
        ('ai-short-video-studio', true, 92),
        ('local-ai-sales-consulting', true, 88),
        ('data-dashboard-service', true, 84)
) AS seed(slug, is_featured, heat)
WHERE project_opportunities.slug = seed.slug;
