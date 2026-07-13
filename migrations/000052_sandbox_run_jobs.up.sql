ALTER TABLE sandbox_sessions
    ADD COLUMN progress_percent INTEGER NOT NULL DEFAULT 0
        CHECK (progress_percent >= 0 AND progress_percent <= 100),
    ADD COLUMN current_step TEXT NOT NULL DEFAULT '',
    ADD COLUMN error_message TEXT NOT NULL DEFAULT '',
    ADD COLUMN run_attempt INTEGER NOT NULL DEFAULT 0 CHECK (run_attempt >= 0);

UPDATE sandbox_sessions
SET progress_percent = CASE WHEN status = 'completed' THEN 100 ELSE 0 END,
    current_step = CASE WHEN status = 'completed' THEN 'completed' ELSE status END;

UPDATE sandbox_sessions
SET report = report || jsonb_build_object(
    'basis', 'model_simulation',
    'disclaimer', '本报告由 AI 基于用户输入进行情景推演，不代表真实市场统计、收益承诺或已验证事实。',
    'assumptions', jsonb_build_array('历史报告未记录结构化假设，使用前需重新核验输入条件。'),
    'evidence_sources', '[]'::JSONB
)
WHERE status = 'completed' AND report <> '{}'::JSONB;
