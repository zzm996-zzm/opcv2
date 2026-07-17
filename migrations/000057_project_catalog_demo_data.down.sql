DELETE FROM project_cases
WHERE slug IN (
    'short-video-first-client',
    'short-video-scope-failure',
    'excel-monthly-report-pilot',
    'resume-privacy-review'
);

DELETE FROM project_opportunities
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
