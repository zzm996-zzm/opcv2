package sandbox

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
)

const v2RunColumns = `
	id, user_id, v2_name, v2_product, v2_context, v2_questions, v2_assumptions,
	v2_roles, v2_orchestration_mode, v2_model_routing_snapshot, v2_evidence_pack,
	COALESCE(v2_input_context_hash, ''), v2_completeness, v2_rounds, v2_revision,
	v2_status, v2_started_at, v2_finished_at, created_at, updated_at`

func (r *PostgresRepository) ListV2RoleConfigs(ctx context.Context) ([]V2RoleConfig, error) {
	rows, err := r.db.Query(ctx, `
		SELECT role_code, display_name, description, analysis_dimensions, default_selected,
		       is_required, default_model_route, prompt_version, system_prompt
		FROM sandbox_role_configs
		WHERE is_active = TRUE
		ORDER BY CASE role_code
			WHEN 'customer' THEN 1 WHEN 'investor' THEN 2 WHEN 'competitor' THEN 3
			WHEN 'channel' THEN 4 WHEN 'supply' THEN 5 WHEN 'expert' THEN 6
			WHEN 'skeptic' THEN 7 WHEN 'partner' THEN 8 ELSE 99 END
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	roles := make([]V2RoleConfig, 0, len(V2RoleCodes))
	for rows.Next() {
		var role V2RoleConfig
		var dimensions []byte
		if err := rows.Scan(&role.Code, &role.DisplayName, &role.Description, &dimensions, &role.DefaultSelected, &role.Required, &role.DefaultModelRoute, &role.PromptVersion, &role.SystemPrompt); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(dimensions, &role.Dimensions); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, rows.Err()
}

func (r *PostgresRepository) CreateV2Run(ctx context.Context, run V2SandboxRun) (V2SandboxRun, error) {
	product, contextValue, questions, assumptions, roles, routing, evidence, err := marshalV2Run(run)
	if err != nil {
		return V2SandboxRun{}, err
	}
	legacyGoal := run.Name
	legacyTarget := run.Context.TargetCustomer
	legacyProduct := run.Product.Name
	err = r.db.QueryRow(ctx, `
		INSERT INTO sandbox_sessions (
			user_id, goal, target_users, product, roles, status, current_step, sandbox_version,
			v2_name, v2_product, v2_context, v2_questions, v2_assumptions, v2_roles,
			v2_orchestration_mode, v2_model_routing_snapshot, v2_evidence_pack,
			v2_completeness, v2_rounds, v2_revision, v2_status, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, 2,
			$8, $9, $10, $11, $12, $13, $14, $15, $16,
			$17, $18, $19, $20, $21, $21
		)
		RETURNING id
	`, run.UserID, legacyGoal, legacyTarget, legacyProduct, roles, StatusDraft, run.Status,
		run.Name, product, contextValue, questions, assumptions, roles, run.OrchestrationMode,
		routing, evidence, run.Completeness, run.Rounds, run.Revision, run.Status, run.CreatedAt,
	).Scan(&run.ID)
	return run, err
}

func (r *PostgresRepository) GetV2Run(ctx context.Context, userID, runID int64) (V2SandboxRun, error) {
	run, err := scanV2Run(r.db.QueryRow(ctx, `SELECT `+v2RunColumns+` FROM sandbox_sessions WHERE user_id = $1 AND id = $2 AND sandbox_version = 2`, userID, runID))
	if errors.Is(err, pgx.ErrNoRows) {
		return V2SandboxRun{}, ErrV2RunNotFound
	}
	if err != nil {
		return V2SandboxRun{}, err
	}
	run.RunRoles, err = r.listV2RunRoles(ctx, run.ID)
	if err != nil {
		return V2SandboxRun{}, err
	}
	report, reportErr := r.getV2Report(ctx, run.ID)
	if reportErr == nil {
		run.Report = &report
	} else if !errors.Is(reportErr, pgx.ErrNoRows) {
		return V2SandboxRun{}, reportErr
	}
	return run, nil
}

func (r *PostgresRepository) UpdateV2RunDraft(ctx context.Context, run V2SandboxRun, expectedRevision int) (V2SandboxRun, error) {
	product, contextValue, questions, assumptions, roles, _, _, err := marshalV2Run(run)
	if err != nil {
		return V2SandboxRun{}, err
	}
	updated, err := scanV2Run(r.db.QueryRow(ctx, `
		UPDATE sandbox_sessions
		SET goal = $1, target_users = $2, product = $3, roles = $4,
		    v2_name = $1, v2_product = $5, v2_context = $6, v2_questions = $7,
		    v2_assumptions = $8, v2_roles = $4, v2_completeness = $9,
		    v2_rounds = $10, v2_status = $11, current_step = $11,
		    v2_revision = v2_revision + 1, updated_at = NOW()
		WHERE user_id = $12 AND id = $13 AND sandbox_version = 2 AND v2_revision = $14
		RETURNING `+v2RunColumns,
		run.Name, run.Context.TargetCustomer, run.Product.Name, roles, product, contextValue,
		questions, assumptions, run.Completeness, run.Rounds, run.Status, run.UserID, run.ID, expectedRevision,
	))
	if errors.Is(err, pgx.ErrNoRows) {
		return V2SandboxRun{}, ErrV2Revision
	}
	return updated, err
}

func (r *PostgresRepository) ListV2Runs(ctx context.Context, userID int64, limit int) ([]V2SandboxRun, error) {
	rows, err := r.db.Query(ctx, `SELECT `+v2RunColumns+` FROM sandbox_sessions WHERE user_id = $1 AND sandbox_version = 2 ORDER BY created_at DESC LIMIT $2`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	runs := make([]V2SandboxRun, 0)
	for rows.Next() {
		run, err := scanV2Run(rows)
		if err != nil {
			return nil, err
		}
		runs = append(runs, run)
	}
	return runs, rows.Err()
}

func (r *PostgresRepository) DeleteV2Run(ctx context.Context, userID, runID int64) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM sandbox_sessions WHERE user_id = $1 AND id = $2 AND sandbox_version = 2 AND v2_status <> $3`, userID, runID, V2StatusRunning)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrV2RunNotFound
	}
	return nil
}

func marshalV2Run(run V2SandboxRun) (product, contextValue, questions, assumptions, roles, routing, evidence []byte, err error) {
	values := []any{run.Product, run.Context, run.Questions, run.Assumptions, run.Roles, run.ModelRoutingSnapshot, run.EvidencePack}
	output := make([][]byte, len(values))
	for i, value := range values {
		output[i], err = json.Marshal(value)
		if err != nil {
			return nil, nil, nil, nil, nil, nil, nil, err
		}
	}
	return output[0], output[1], output[2], output[3], output[4], output[5], output[6], nil
}

func scanV2Run(scanner sessionScanner) (V2SandboxRun, error) {
	var run V2SandboxRun
	var product, contextValue, questions, assumptions, roles, routing, evidence []byte
	if err := scanner.Scan(
		&run.ID, &run.UserID, &run.Name, &product, &contextValue, &questions, &assumptions,
		&roles, &run.OrchestrationMode, &routing, &evidence, &run.InputContextHash,
		&run.Completeness, &run.Rounds, &run.Revision, &run.Status, &run.StartedAt,
		&run.FinishedAt, &run.CreatedAt, &run.UpdatedAt,
	); err != nil {
		return V2SandboxRun{}, err
	}
	for _, target := range []struct {
		data  []byte
		value any
	}{
		{product, &run.Product}, {contextValue, &run.Context}, {questions, &run.Questions},
		{assumptions, &run.Assumptions}, {roles, &run.Roles}, {routing, &run.ModelRoutingSnapshot},
		{evidence, &run.EvidencePack},
	} {
		if len(target.data) > 0 && string(target.data) != "null" {
			if err := json.Unmarshal(target.data, target.value); err != nil {
				return V2SandboxRun{}, err
			}
		}
	}
	if run.Questions == nil {
		run.Questions = []V2Question{}
	}
	if run.Assumptions == nil {
		run.Assumptions = []string{}
	}
	if run.Roles == nil {
		run.Roles = []string{}
	}
	if run.ModelRoutingSnapshot == nil {
		run.ModelRoutingSnapshot = map[string]any{}
	}
	if run.EvidencePack == nil {
		run.EvidencePack = []map[string]any{}
	}
	return run, nil
}

func (r *PostgresRepository) listV2RunRoles(ctx context.Context, runID int64) ([]V2RunRole, error) {
	rows, err := r.db.Query(ctx, `
		SELECT run_id, role_code, seq, role_session_id, model_provider, model_name, model_route,
		       prompt_version, system_prompt, analysis_dimensions, input_context_hash, status, COALESCE(stance, ''),
		       output_json, input_tokens, output_tokens, COALESCE(latency_ms, 0), retry_count,
		       error_code, started_at, finished_at
		FROM sandbox_run_roles WHERE run_id = $1 ORDER BY seq
	`, runID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	roles := make([]V2RunRole, 0)
	for rows.Next() {
		var role V2RunRole
		var dimensions, output []byte
		if err := rows.Scan(&role.RunID, &role.RoleCode, &role.Seq, &role.RoleSessionID, &role.ModelProvider,
			&role.ModelName, &role.ModelRoute, &role.PromptVersion, &role.SystemPrompt, &dimensions, &role.InputHash,
			&role.Status, &role.Stance, &output, &role.InputTokens, &role.OutputTokens, &role.LatencyMS,
			&role.RetryCount, &role.ErrorCode, &role.StartedAt, &role.FinishedAt); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(dimensions, &role.Dimensions); err != nil {
			return nil, err
		}
		if len(output) > 0 && string(output) != "null" {
			var roleOutput V2RoleOutput
			if err := json.Unmarshal(output, &roleOutput); err != nil {
				return nil, err
			}
			role.Output = &roleOutput
		}
		roles = append(roles, role)
	}
	return roles, rows.Err()
}

func (r *PostgresRepository) getV2Report(ctx context.Context, runID int64) (V2SandboxReport, error) {
	var data []byte
	if err := r.db.QueryRow(ctx, `SELECT report FROM sandbox_reports WHERE run_id = $1`, runID).Scan(&data); err != nil {
		return V2SandboxReport{}, err
	}
	var report V2SandboxReport
	if err := json.Unmarshal(data, &report); err != nil {
		return V2SandboxReport{}, err
	}
	return report, nil
}
