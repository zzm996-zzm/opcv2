ALTER TABLE project_match_sessions
ADD COLUMN answers JSONB NOT NULL DEFAULT '[]'::JSONB;
