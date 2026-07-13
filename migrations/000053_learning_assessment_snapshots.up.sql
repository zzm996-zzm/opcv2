ALTER TABLE learning_diagnoses
    ADD COLUMN answers JSONB NOT NULL DEFAULT '[]'::JSONB,
    ADD COLUMN basis TEXT NOT NULL DEFAULT 'legacy_estimate',
    ADD COLUMN disclaimer TEXT NOT NULL DEFAULT '',
    ADD COLUMN assumptions JSONB NOT NULL DEFAULT '[]'::JSONB,
    ADD COLUMN evidence_sources JSONB NOT NULL DEFAULT '[]'::JSONB,
    ADD COLUMN gaps_snapshot JSONB NOT NULL DEFAULT '{}'::JSONB,
    ADD COLUMN recommendations_snapshot JSONB NOT NULL DEFAULT '{}'::JSONB,
    ADD COLUMN plan_snapshot JSONB NOT NULL DEFAULT '{}'::JSONB,
    ADD COLUMN report_snapshot JSONB NOT NULL DEFAULT '{}'::JSONB;

UPDATE learning_diagnoses
SET disclaimer = '该历史诊断由旧版固定估算逻辑生成，不代表标准化考试成绩或客观能力认证。',
    assumptions = jsonb_build_array('历史诊断未保存结构化评估假设，使用前需重新诊断。'),
    evidence_sources = '[]'::JSONB
WHERE status = 'completed';
