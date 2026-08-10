ALTER TABLE sandbox_sessions
    ADD COLUMN IF NOT EXISTS sandbox_version INTEGER NOT NULL DEFAULT 1,
    ADD COLUMN IF NOT EXISTS v2_name TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS v2_product JSONB NOT NULL DEFAULT '{}'::JSONB,
    ADD COLUMN IF NOT EXISTS v2_context JSONB NOT NULL DEFAULT '{}'::JSONB,
    ADD COLUMN IF NOT EXISTS v2_questions JSONB NOT NULL DEFAULT '[]'::JSONB,
    ADD COLUMN IF NOT EXISTS v2_assumptions JSONB NOT NULL DEFAULT '[]'::JSONB,
    ADD COLUMN IF NOT EXISTS v2_roles JSONB NOT NULL DEFAULT '[]'::JSONB,
    ADD COLUMN IF NOT EXISTS v2_orchestration_mode TEXT NOT NULL DEFAULT 'isolated_sessions',
    ADD COLUMN IF NOT EXISTS v2_model_routing_snapshot JSONB NOT NULL DEFAULT '{}'::JSONB,
    ADD COLUMN IF NOT EXISTS v2_evidence_pack JSONB NOT NULL DEFAULT '[]'::JSONB,
    ADD COLUMN IF NOT EXISTS v2_input_context_hash CHAR(64),
    ADD COLUMN IF NOT EXISTS v2_completeness NUMERIC(3,2) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS v2_rounds INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS v2_revision INTEGER NOT NULL DEFAULT 1,
    ADD COLUMN IF NOT EXISTS v2_status TEXT NOT NULL DEFAULT 'draft',
    ADD COLUMN IF NOT EXISTS v2_report_session_id TEXT,
    ADD COLUMN IF NOT EXISTS v2_started_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS v2_finished_at TIMESTAMPTZ;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'sandbox_sessions_v2_mode_check') THEN
        ALTER TABLE sandbox_sessions ADD CONSTRAINT sandbox_sessions_v2_mode_check CHECK (v2_orchestration_mode IN ('isolated_sessions'));
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'sandbox_sessions_v2_status_check') THEN
        ALTER TABLE sandbox_sessions ADD CONSTRAINT sandbox_sessions_v2_status_check CHECK (v2_status IN ('draft', 'clarifying', 'ready', 'running', 'partial', 'done', 'failed', 'ai_no_result'));
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS sandbox_sessions_v2_user_idx
    ON sandbox_sessions (user_id, created_at DESC)
    WHERE sandbox_version = 2;

CREATE TABLE IF NOT EXISTS sandbox_role_configs (
    role_code TEXT PRIMARY KEY,
    display_name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    system_prompt TEXT NOT NULL,
    analysis_dimensions JSONB NOT NULL DEFAULT '[]'::JSONB,
    output_schema JSONB NOT NULL DEFAULT '{}'::JSONB,
    default_selected BOOLEAN NOT NULL DEFAULT TRUE,
    is_required BOOLEAN NOT NULL DEFAULT FALSE,
    default_model_route TEXT NOT NULL DEFAULT 'role_default',
    prompt_version TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS sandbox_run_roles (
    run_id BIGINT NOT NULL REFERENCES sandbox_sessions(id) ON DELETE CASCADE,
    role_code TEXT NOT NULL REFERENCES sandbox_role_configs(role_code),
    seq INTEGER NOT NULL CHECK (seq > 0),
    role_session_id TEXT NOT NULL UNIQUE,
    model_provider TEXT NOT NULL DEFAULT '',
    model_name TEXT NOT NULL DEFAULT '',
    model_route TEXT NOT NULL DEFAULT 'role_default',
    prompt_version TEXT NOT NULL,
    analysis_dimensions JSONB NOT NULL DEFAULT '[]'::JSONB,
    input_context_hash CHAR(64) NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'queued', 'running', 'done', 'failed', 'cancelled')),
    stance TEXT CHECK (stance IN ('support', 'neutral', 'oppose')),
    output_json JSONB,
    input_tokens INTEGER NOT NULL DEFAULT 0,
    output_tokens INTEGER NOT NULL DEFAULT 0,
    latency_ms INTEGER,
    retry_count INTEGER NOT NULL DEFAULT 0,
    error_code TEXT NOT NULL DEFAULT '',
    started_at TIMESTAMPTZ,
    finished_at TIMESTAMPTZ,
    PRIMARY KEY (run_id, role_code)
);

CREATE INDEX IF NOT EXISTS sandbox_run_roles_status_idx
    ON sandbox_run_roles (run_id, status, seq);

CREATE TABLE IF NOT EXISTS sandbox_run_messages (
    id BIGSERIAL PRIMARY KEY,
    run_id BIGINT NOT NULL REFERENCES sandbox_sessions(id) ON DELETE CASCADE,
    role_code TEXT NOT NULL,
    content TEXT NOT NULL,
    stance TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS sandbox_run_messages_run_idx ON sandbox_run_messages (run_id, id);

CREATE TABLE IF NOT EXISTS sandbox_run_events (
    id BIGSERIAL PRIMARY KEY,
    run_id BIGINT NOT NULL REFERENCES sandbox_sessions(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_code TEXT NOT NULL DEFAULT '',
    event TEXT NOT NULL,
    payload JSONB NOT NULL DEFAULT '{}'::JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS sandbox_run_events_replay_idx ON sandbox_run_events (run_id, user_id, id);

CREATE TABLE IF NOT EXISTS sandbox_reports (
    run_id BIGINT PRIMARY KEY REFERENCES sandbox_sessions(id) ON DELETE CASCADE,
    report_session_id TEXT NOT NULL UNIQUE,
    report JSONB NOT NULL,
    completed_roles JSONB NOT NULL DEFAULT '[]'::JSONB,
    failed_roles JSONB NOT NULL DEFAULT '[]'::JSONB,
    model_provider TEXT NOT NULL DEFAULT '',
    model_name TEXT NOT NULL DEFAULT '',
    prompt_version TEXT NOT NULL DEFAULT '',
    input_tokens INTEGER NOT NULL DEFAULT 0,
    output_tokens INTEGER NOT NULL DEFAULT 0,
    is_model_generated BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS sandbox_exports (
    id BIGSERIAL PRIMARY KEY,
    run_id BIGINT NOT NULL REFERENCES sandbox_sessions(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    format TEXT NOT NULL CHECK (format IN ('pdf', 'link', 'json')),
    payload JSONB NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO sandbox_role_configs (role_code, display_name, description, system_prompt, analysis_dimensions, default_selected, is_required, prompt_version)
VALUES
 ('customer', '目标客户', '判断真实购买者会不会买', '从目标客户和付费者角度判断购买意愿。', '["pain","usage","buyer","alternatives","trigger","price","decision","trust","friction","retention"]', TRUE, FALSE, 'sandbox_role_v1'),
 ('investor', '投资人', '判断投入与规模化价值', '从投资人与经营者角度判断需求、经济模型和回报路径。', '["demand","market","revenue","unit_economics","scale","moat","team","capital","cash_burn","return"]', TRUE, FALSE, 'sandbox_role_v1'),
 ('competitor', '竞争对手', '寻找竞争软肋和反击路径', '站在既有玩家角度设计反击并寻找项目软肋。', '["substitutes","differentiation","copyability","price_war","channel","brand","switching_cost","speed","weakness","defense"]', TRUE, FALSE, 'sandbox_role_v1'),
 ('channel', '渠道方', '判断是否愿意销售', '从渠道伙伴角度判断客群、利润、动销和合作条件。', '["fit","margin","sales_cycle","training","inventory","payment","after_sales","conflict","promotion","priority"]', TRUE, FALSE, 'sandbox_role_v1'),
 ('supply', '供应链运营', '判断能否稳定交付', '从供应链和运营角度判断成本、交付、质量和扩量瓶颈。', '["supply","cost","capacity","delivery","quality","inventory","logistics","service","sop","cash","bottleneck"]', TRUE, FALSE, 'sandbox_role_v1'),
 ('expert', '行业专家', '用行业规律校正判断', '从行业规律、政策、周期和准入角度校正局部判断。', '["lifecycle","drivers","trend","maturity","policy","compliance","benchmark","seasonality","window","risk"]', TRUE, FALSE, 'sandbox_role_v1'),
 ('skeptic', '悲观者', '主动推翻乐观假设', '主动寻找致命假设、证据缺口和停止条件。', '["fatal_assumption","evidence_gap","refusal","worst_case","cash_break","execution","bias","compliance","abuse","kill_criteria"]', TRUE, TRUE, 'sandbox_role_v1'),
 ('partner', '合伙人', '判断分工与合作风险', '从合伙人角度判断能力互补、职责、决策和退出安排。', '["founder_fit","complement","resources","roles","decision","equity","commitment","conflict","milestone","exit"]', FALSE, FALSE, 'sandbox_role_v1')
ON CONFLICT (role_code) DO UPDATE SET
    display_name = EXCLUDED.display_name,
    description = EXCLUDED.description,
    system_prompt = EXCLUDED.system_prompt,
    analysis_dimensions = EXCLUDED.analysis_dimensions,
    default_selected = EXCLUDED.default_selected,
    is_required = EXCLUDED.is_required,
    prompt_version = EXCLUDED.prompt_version,
    updated_at = NOW();
