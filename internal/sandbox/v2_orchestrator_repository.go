package sandbox

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
)

func (r *PostgresRepository) ClaimV2Role(ctx context.Context, userID, runID int64, roleCode string) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE sandbox_run_roles rr SET status = 'running', started_at = NOW(), error_code = ''
		WHERE rr.run_id = $1 AND rr.role_code = $2 AND rr.status IN ('pending','queued')
		  AND EXISTS(SELECT 1 FROM sandbox_sessions s WHERE s.id = rr.run_id AND s.user_id = $3 AND s.v2_status = $4)
	`, runID, roleCode, userID, V2StatusRunning)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrStaleRun
	}
	return nil
}

func (r *PostgresRepository) CompleteV2Role(ctx context.Context, userID, runID int64, roleCode string, output V2RoleOutput, inputTokens, outputTokens, latencyMS int) error {
	data, err := json.Marshal(output)
	if err != nil {
		return err
	}
	tag, err := r.db.Exec(ctx, `
		UPDATE sandbox_run_roles rr
		SET status = 'done', stance = $1, output_json = $2, input_tokens = $3,
		    output_tokens = $4, latency_ms = $5, finished_at = NOW(), error_code = ''
		WHERE rr.run_id = $6 AND rr.role_code = $7 AND rr.status = 'running'
		  AND EXISTS(SELECT 1 FROM sandbox_sessions s WHERE s.id = rr.run_id AND s.user_id = $8 AND s.v2_status = $9)
	`, output.Stance, data, inputTokens, outputTokens, latencyMS, runID, roleCode, userID, V2StatusRunning)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrStaleRun
	}
	_, _ = r.db.Exec(ctx, `
		INSERT INTO sandbox_run_messages (run_id, role_code, content, stance)
		VALUES ($1, $2, $3, $4)
	`, runID, roleCode, output.Content, output.Stance)
	return nil
}

func (r *PostgresRepository) FailV2Role(ctx context.Context, userID, runID int64, roleCode, errorCode string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE sandbox_run_roles rr
		SET status = 'failed', retry_count = retry_count + 1, error_code = $1, finished_at = NOW()
		WHERE rr.run_id = $2 AND rr.role_code = $3 AND rr.status IN ('pending','queued','running')
		  AND EXISTS(SELECT 1 FROM sandbox_sessions s WHERE s.id = rr.run_id AND s.user_id = $4 AND s.v2_status = $5)
	`, errorCode, runID, roleCode, userID, V2StatusRunning)
	return err
}

func (r *PostgresRepository) SaveV2Report(ctx context.Context, userID, runID int64, reportSessionID string, report V2SandboxReport, completedRoles, failedRoles []string, inputTokens, outputTokens int, modelGenerated bool) error {
	reportJSON, err := json.Marshal(report)
	if err != nil {
		return err
	}
	completedJSON, _ := json.Marshal(completedRoles)
	failedJSON, _ := json.Marshal(failedRoles)
	tag, err := r.db.Exec(ctx, `
		INSERT INTO sandbox_reports (
			run_id, report_session_id, report, completed_roles, failed_roles,
			prompt_version, input_tokens, output_tokens, is_model_generated
		)
		SELECT $1,$2,$3,$4,$5,'sandbox_report_v1',$6,$7,$8
		WHERE EXISTS(SELECT 1 FROM sandbox_sessions WHERE id = $1 AND user_id = $9 AND v2_status IN ($10, $11, $12))
		ON CONFLICT (run_id) DO UPDATE SET
			report = EXCLUDED.report, completed_roles = EXCLUDED.completed_roles,
			failed_roles = EXCLUDED.failed_roles, input_tokens = EXCLUDED.input_tokens,
			output_tokens = EXCLUDED.output_tokens, is_model_generated = EXCLUDED.is_model_generated
	`, runID, reportSessionID, reportJSON, completedJSON, failedJSON, inputTokens, outputTokens, modelGenerated, userID, V2StatusRunning, V2StatusPartial, V2StatusDone)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrStaleRun
	}
	return nil
}

func (r *PostgresRepository) CompleteV2Run(ctx context.Context, userID, runID int64, status string) error {
	legacyStatus := StatusFailed
	if status == V2StatusDone || status == V2StatusPartial {
		legacyStatus = StatusCompleted
	}
	tag, err := r.db.Exec(ctx, `
		UPDATE sandbox_sessions
		SET status = $1, progress_percent = 100, current_step = $2, v2_status = $2,
		    v2_finished_at = NOW(), v2_report_session_id = (SELECT report_session_id FROM sandbox_reports WHERE run_id = $3),
		    updated_at = NOW()
		WHERE user_id = $4 AND id = $3 AND sandbox_version = 2 AND v2_status = $5
	`, legacyStatus, status, runID, userID, V2StatusRunning)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrStaleRun
	}
	return nil
}

func (r *PostgresRepository) GetV2Report(ctx context.Context, userID, runID int64) (V2SandboxReport, error) {
	var reportJSON []byte
	err := r.db.QueryRow(ctx, `
		SELECT sr.report FROM sandbox_reports sr
		JOIN sandbox_sessions s ON s.id = sr.run_id
		WHERE sr.run_id = $1 AND s.user_id = $2 AND s.sandbox_version = 2
	`, runID, userID).Scan(&reportJSON)
	if errors.Is(err, pgx.ErrNoRows) {
		return V2SandboxReport{}, ErrV2RunNotFound
	}
	if err != nil {
		return V2SandboxReport{}, err
	}
	var report V2SandboxReport
	if err := json.Unmarshal(reportJSON, &report); err != nil {
		return V2SandboxReport{}, err
	}
	return report, nil
}
