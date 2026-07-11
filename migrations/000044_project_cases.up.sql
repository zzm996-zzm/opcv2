CREATE TABLE project_cases (
 id BIGSERIAL PRIMARY KEY, slug TEXT NOT NULL UNIQUE CHECK (slug <> ''), opportunity_id BIGINT REFERENCES project_opportunities(id) ON DELETE SET NULL,
 title TEXT NOT NULL CHECK (title <> ''), summary TEXT NOT NULL DEFAULT '', case_type TEXT NOT NULL CHECK (case_type IN ('success','failure')),
 outcome TEXT NOT NULL DEFAULT '', key_actions JSONB NOT NULL DEFAULT '[]'::JSONB, lessons JSONB NOT NULL DEFAULT '[]'::JSONB, pitfalls JSONB NOT NULL DEFAULT '[]'::JSONB,
 source_title TEXT NOT NULL CHECK (source_title <> ''), source_url TEXT NOT NULL CHECK (source_url LIKE 'http%'), captured_at TIMESTAMPTZ NOT NULL,
 status TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft','published')), published_at TIMESTAMPTZ, created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
 CHECK ((status = 'published' AND published_at IS NOT NULL) OR status = 'draft')
);
CREATE INDEX project_cases_published_idx ON project_cases (published_at DESC, id) WHERE status = 'published';
