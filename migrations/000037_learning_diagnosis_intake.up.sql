ALTER TABLE learning_diagnoses
    ADD COLUMN focus_abilities JSONB NOT NULL DEFAULT '[]'::JSONB,
    ADD COLUMN weekly_time TEXT NOT NULL DEFAULT '',
    ADD COLUMN bottleneck TEXT NOT NULL DEFAULT '';
