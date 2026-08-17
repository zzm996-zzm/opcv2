DROP FUNCTION IF EXISTS refresh_lootdrop_project_opportunities(BIGINT[]);

DELETE FROM project_opportunities WHERE slug LIKE 'lootdrop-rebuild-%';

UPDATE project_opportunities
SET status = 'published', published_at = COALESCE(published_at, NOW()), updated_at = NOW()
WHERE slug IN (
    'ai-short-video-studio', 'excel-automation-reporting', 'ai-resume-optimization',
    'local-pet-care-subscription', 'knowledge-course-production',
    'local-ai-sales-consulting', 'niche-travel-guide', 'data-dashboard-service'
);

UPDATE project_opportunities SET status = 'draft', updated_at = NOW() WHERE status = 'offline';
ALTER TABLE project_opportunities DROP CONSTRAINT IF EXISTS project_opportunities_check;
ALTER TABLE project_opportunities
    ADD CONSTRAINT project_opportunities_check CHECK (
        (status = 'published' AND published_at IS NOT NULL) OR status = 'draft'
    );
