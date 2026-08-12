DELETE FROM project_cases
WHERE slug IN (
    'klarna-ai-customer-service-2024',
    'coca-cola-united-process-automation',
    'amazon-recruiting-ai-bias-2018'
);

DELETE FROM web_research_jobs
WHERE trigger_reason IN (
    'seed:project-case:klarna-ai-assistant-2024',
    'seed:project-case:coca-cola-united-power-automate',
    'seed:project-case:amazon-recruiting-ai-bias'
);
