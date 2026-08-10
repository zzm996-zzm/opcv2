ALTER TABLE project_exports
    ADD COLUMN format TEXT NOT NULL DEFAULT 'json' CHECK (format IN ('json', 'pdf', 'link')),
    ADD COLUMN includes JSONB NOT NULL DEFAULT '[]'::JSONB CHECK (jsonb_typeof(includes) = 'array'),
    ADD COLUMN expires_at TIMESTAMPTZ NOT NULL DEFAULT (NOW() + INTERVAL '7 days');
