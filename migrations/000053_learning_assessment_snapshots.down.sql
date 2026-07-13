ALTER TABLE learning_diagnoses
    DROP COLUMN IF EXISTS report_snapshot,
    DROP COLUMN IF EXISTS plan_snapshot,
    DROP COLUMN IF EXISTS recommendations_snapshot,
    DROP COLUMN IF EXISTS gaps_snapshot,
    DROP COLUMN IF EXISTS evidence_sources,
    DROP COLUMN IF EXISTS assumptions,
    DROP COLUMN IF EXISTS disclaimer,
    DROP COLUMN IF EXISTS basis,
    DROP COLUMN IF EXISTS answers;
