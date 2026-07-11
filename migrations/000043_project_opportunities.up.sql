CREATE TABLE project_opportunities (
    id BIGSERIAL PRIMARY KEY,
    slug TEXT NOT NULL UNIQUE CHECK (slug <> ''),
    title TEXT NOT NULL CHECK (title <> ''),
    summary TEXT NOT NULL DEFAULT '',
    industry TEXT NOT NULL DEFAULT '',
    tags JSONB NOT NULL DEFAULT '[]'::JSONB CHECK (jsonb_typeof(tags) = 'array'),
    budget_band TEXT NOT NULL DEFAULT '',
    difficulty TEXT NOT NULL DEFAULT '',
    resource_requirements JSONB NOT NULL DEFAULT '[]'::JSONB CHECK (jsonb_typeof(resource_requirements) = 'array'),
    sections JSONB NOT NULL DEFAULT '[]'::JSONB CHECK (jsonb_typeof(sections) = 'array'),
    status TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'published')),
    sort_order INTEGER NOT NULL DEFAULT 0,
    published_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK ((status = 'published' AND published_at IS NOT NULL) OR status = 'draft')
);

CREATE INDEX project_opportunities_published_idx ON project_opportunities (sort_order, published_at DESC, id) WHERE status = 'published';
