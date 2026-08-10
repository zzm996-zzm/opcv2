DROP TABLE IF EXISTS sandbox_exports;
DROP TABLE IF EXISTS sandbox_reports;
DROP TABLE IF EXISTS sandbox_run_events;
DROP TABLE IF EXISTS sandbox_run_messages;
DROP TABLE IF EXISTS sandbox_run_roles;
DROP TABLE IF EXISTS sandbox_role_configs;

UPDATE membership_plan_quotas SET limit_value = CASE plan_code WHEN 'free' THEN 1 WHEN 'pro' THEN 20 ELSE limit_value END
WHERE key = 'sandbox_runs';

UPDATE membership_usage SET limit_value = 0 WHERE limit_value < 0;

ALTER TABLE membership_plan_quotas
    DROP CONSTRAINT IF EXISTS membership_plan_quotas_limit_value_check,
    ADD CONSTRAINT membership_plan_quotas_limit_value_check CHECK (limit_value >= 0);

ALTER TABLE membership_usage
    DROP CONSTRAINT IF EXISTS membership_usage_limit_value_check,
    ADD CONSTRAINT membership_usage_limit_value_check CHECK (limit_value >= 0);
DROP INDEX IF EXISTS sandbox_sessions_v2_user_idx;
DROP INDEX IF EXISTS sandbox_sessions_v2_one_active_user_idx;
ALTER TABLE sandbox_sessions
    DROP CONSTRAINT IF EXISTS sandbox_sessions_v2_mode_check,
    DROP CONSTRAINT IF EXISTS sandbox_sessions_v2_status_check,
    DROP COLUMN IF EXISTS sandbox_version,
    DROP COLUMN IF EXISTS v2_name,
    DROP COLUMN IF EXISTS v2_product,
    DROP COLUMN IF EXISTS v2_context,
    DROP COLUMN IF EXISTS v2_questions,
    DROP COLUMN IF EXISTS v2_assumptions,
    DROP COLUMN IF EXISTS v2_roles,
    DROP COLUMN IF EXISTS v2_orchestration_mode,
    DROP COLUMN IF EXISTS v2_model_routing_snapshot,
    DROP COLUMN IF EXISTS v2_evidence_pack,
    DROP COLUMN IF EXISTS v2_input_context_hash,
    DROP COLUMN IF EXISTS v2_completeness,
    DROP COLUMN IF EXISTS v2_rounds,
    DROP COLUMN IF EXISTS v2_revision,
    DROP COLUMN IF EXISTS v2_status,
    DROP COLUMN IF EXISTS v2_report_session_id,
    DROP COLUMN IF EXISTS v2_started_at,
    DROP COLUMN IF EXISTS v2_finished_at;
