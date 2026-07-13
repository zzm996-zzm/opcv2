CREATE TABLE learning_course_materials (
    id BIGSERIAL PRIMARY KEY,
    course_slug TEXT NOT NULL REFERENCES learning_courses(slug) ON DELETE CASCADE,
    title TEXT NOT NULL,
    material_type TEXT NOT NULL,
    content_url TEXT NOT NULL DEFAULT '',
    position INTEGER NOT NULL DEFAULT 0 CHECK (position >= 0),
    downloadable BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX learning_course_materials_course_position_idx
    ON learning_course_materials (course_slug, position, id);

CREATE TABLE learning_plan_items (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    diagnosis_id BIGINT NOT NULL REFERENCES learning_diagnoses(id) ON DELETE CASCADE,
    stage_number INTEGER NOT NULL CHECK (stage_number > 0),
    title TEXT NOT NULL,
    completed BOOLEAN NOT NULL DEFAULT FALSE,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, diagnosis_id, stage_number)
);

CREATE INDEX learning_plan_items_user_diagnosis_idx
    ON learning_plan_items (user_id, diagnosis_id, stage_number);
