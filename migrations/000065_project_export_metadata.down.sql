ALTER TABLE project_exports
    DROP COLUMN IF EXISTS expires_at,
    DROP COLUMN IF EXISTS includes,
    DROP COLUMN IF EXISTS format;
