package sandbox

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
)

func (r *PostgresRepository) PrepareV2Run(ctx context.Context, run V2SandboxRun, roles []V2RunRole, inputHash string, routing map[string]any) (V2SandboxRun, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return V2SandboxRun{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock($1)`, run.UserID); err != nil {
		return V2SandboxRun{}, err
	}
	var active bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM sandbox_sessions WHERE user_id = $1 AND sandbox_version = 2 AND v2_status = $2 AND id <> $3)`, run.UserID, V2StatusRunning, run.ID).Scan(&active); err != nil {
		return V2SandboxRun{}, err
	}
	if active {
		return V2SandboxRun{}, ErrV2ActiveRun
	}
	routingJSON, err := json.Marshal(routing)
	if err != nil {
		return V2SandboxRun{}, err
	}
	updated, err := scanV2Run(tx.QueryRow(ctx, `
		UPDATE sandbox_sessions
		SET status = $1, progress_percent = 0, current_step = $2, error_message = '',
		    v2_status = $2, v2_input_context_hash = $3, v2_model_routing_snapshot = $4,
		    v2_started_at = NOW(), v2_finished_at = NULL, v2_revision = v2_revision + 1,
		    updated_at = NOW()
		WHERE user_id = $5 AND id = $6 AND sandbox_version = 2 AND v2_status = $7 AND v2_revision = $8
		RETURNING `+v2RunColumns,
		StatusRunning, V2StatusRunning, inputHash, routingJSON, run.UserID, run.ID, V2StatusReady, run.Revision,
	))
	if errors.Is(err, pgx.ErrNoRows) {
		return V2SandboxRun{}, ErrV2StartConflict
	}
	if err != nil {
		return V2SandboxRun{}, err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM sandbox_run_roles WHERE run_id = $1`, run.ID); err != nil {
		return V2SandboxRun{}, err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM sandbox_reports WHERE run_id = $1`, run.ID); err != nil {
		return V2SandboxRun{}, err
	}
	for _, role := range roles {
		dimensions, err := json.Marshal(role.Dimensions)
		if err != nil {
			return V2SandboxRun{}, err
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO sandbox_run_roles (
				run_id, role_code, seq, role_session_id, model_route, prompt_version,
				system_prompt, analysis_dimensions, input_context_hash, status
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		`, role.RunID, role.RoleCode, role.Seq, role.RoleSessionID, role.ModelRoute, role.PromptVersion, role.SystemPrompt, dimensions, role.InputHash, role.Status); err != nil {
			return V2SandboxRun{}, err
		}
	}
	if err := appendV2EventTx(ctx, tx, V2ProgressEvent{RunID: run.ID, Event: V2EventRunQueued, Payload: map[string]any{"roles": run.Roles}}, run.UserID); err != nil {
		return V2SandboxRun{}, err
	}
	for _, role := range roles {
		if err := appendV2EventTx(ctx, tx, V2ProgressEvent{RunID: run.ID, RoleCode: role.RoleCode, Event: V2EventRoleQueued, Payload: map[string]any{"seq": role.Seq}}, run.UserID); err != nil {
			return V2SandboxRun{}, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return V2SandboxRun{}, err
	}
	updated.RunRoles = roles
	updated.InputContextHash = inputHash
	updated.ModelRoutingSnapshot = routing
	return updated, nil
}

func (r *PostgresRepository) FailV2RunStart(ctx context.Context, userID, runID int64, errorCode string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE sandbox_sessions s SET status = $1, error_message = $2,
		v2_status = CASE WHEN EXISTS(SELECT 1 FROM sandbox_run_roles rr WHERE rr.run_id=s.id AND rr.status='done') THEN $3 ELSE $4 END,
		current_step = CASE WHEN EXISTS(SELECT 1 FROM sandbox_run_roles rr WHERE rr.run_id=s.id AND rr.status='done') THEN $3 ELSE $4 END,
		v2_finished_at = NOW(), updated_at = NOW()
		WHERE user_id = $5 AND id = $6 AND sandbox_version = 2 AND v2_status = $7
	`, StatusFailed, errorCode, V2StatusPartial, V2StatusFailed, userID, runID, V2StatusRunning)
	return err
}

func (r *PostgresRepository) PrepareV2RoleRetry(ctx context.Context, run V2SandboxRun, role V2RunRole) (V2SandboxRun, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return V2SandboxRun{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock($1)`, run.UserID); err != nil {
		return V2SandboxRun{}, err
	}
	var active bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM sandbox_sessions WHERE user_id=$1 AND sandbox_version=2 AND v2_status=$2 AND id<>$3)`, run.UserID, V2StatusRunning, run.ID).Scan(&active); err != nil {
		return V2SandboxRun{}, err
	}
	if active {
		return V2SandboxRun{}, ErrV2ActiveRun
	}
	updated, err := scanV2Run(tx.QueryRow(ctx, `
		UPDATE sandbox_sessions SET status=$1, progress_percent=0, current_step=$2, error_message='',
		v2_status=$2, v2_started_at=NOW(), v2_finished_at=NULL, v2_report_session_id=NULL,
		v2_revision=v2_revision+1, updated_at=NOW()
		WHERE user_id=$3 AND id=$4 AND sandbox_version=2 AND v2_revision=$5
		  AND v2_status IN ($6,$7,$8)
		RETURNING `+v2RunColumns,
		StatusRunning, V2StatusRunning, run.UserID, run.ID, run.Revision, V2StatusPartial, V2StatusFailed, V2StatusNoResult,
	))
	if errors.Is(err, pgx.ErrNoRows) {
		return V2SandboxRun{}, ErrV2StartConflict
	}
	if err != nil {
		return V2SandboxRun{}, err
	}
	roleUpdate, err := tx.Exec(ctx, `
		UPDATE sandbox_run_roles SET role_session_id=$1, status='queued', stance=NULL, output_json=NULL,
		input_tokens=0, output_tokens=0, latency_ms=NULL, error_code='', started_at=NULL, finished_at=NULL
		WHERE run_id=$2 AND role_code=$3 AND status IN ('failed','cancelled')
	`, role.RoleSessionID, run.ID, role.RoleCode)
	if err != nil {
		return V2SandboxRun{}, err
	}
	if roleUpdate.RowsAffected() == 0 {
		return V2SandboxRun{}, ErrV2InvalidRoles
	}
	if _, err := tx.Exec(ctx, `DELETE FROM sandbox_reports WHERE run_id=$1`, run.ID); err != nil {
		return V2SandboxRun{}, err
	}
	if err := appendV2EventTx(ctx, tx, V2ProgressEvent{RunID: run.ID, RoleCode: role.RoleCode, Event: V2EventRoleQueued, Payload: map[string]any{"retry": true}}, run.UserID); err != nil {
		return V2SandboxRun{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return V2SandboxRun{}, err
	}
	role.RunID = run.ID
	updated.RunRoles = []V2RunRole{role}
	return updated, nil
}

func (r *PostgresRepository) StopV2Run(ctx context.Context, userID, runID int64) (V2SandboxRun, error) {
	run, err := scanV2Run(r.db.QueryRow(ctx, `
		UPDATE sandbox_sessions s
		SET status = $1, progress_percent = 100,
		    v2_status = CASE WHEN EXISTS(SELECT 1 FROM sandbox_run_roles rr WHERE rr.run_id = s.id AND rr.status = 'done') THEN $2 ELSE $3 END,
		    current_step = CASE WHEN EXISTS(SELECT 1 FROM sandbox_run_roles rr WHERE rr.run_id = s.id AND rr.status = 'done') THEN $2 ELSE $3 END,
		    v2_finished_at = NOW(), updated_at = NOW()
		WHERE user_id = $4 AND id = $5 AND sandbox_version = 2 AND v2_status = $6
		RETURNING `+v2RunColumns,
		StatusCanceled, V2StatusPartial, V2StatusFailed, userID, runID, V2StatusRunning,
	))
	if errors.Is(err, pgx.ErrNoRows) {
		return V2SandboxRun{}, ErrV2StartConflict
	}
	if err != nil {
		return V2SandboxRun{}, err
	}
	_, _ = r.db.Exec(ctx, `UPDATE sandbox_run_roles SET status = 'cancelled', finished_at = NOW() WHERE run_id = $1 AND status IN ('pending','queued','running')`, runID)
	return run, nil
}

func (r *PostgresRepository) AppendV2Event(ctx context.Context, event V2ProgressEvent, userID int64) (V2ProgressEvent, error) {
	payload, err := json.Marshal(event.Payload)
	if err != nil {
		return V2ProgressEvent{}, err
	}
	err = r.db.QueryRow(ctx, `
		INSERT INTO sandbox_run_events (run_id, user_id, role_code, event, payload, created_at)
		SELECT $1, $2, $3, $4, $5, COALESCE($6, NOW())
		WHERE EXISTS(SELECT 1 FROM sandbox_sessions WHERE id = $1 AND user_id = $2 AND sandbox_version = 2)
		RETURNING id, created_at
	`, event.RunID, userID, event.RoleCode, event.Event, payload, nullableTime(event.CreatedAt)).Scan(&event.ID, &event.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return V2ProgressEvent{}, ErrV2RunNotFound
	}
	return event, err
}

func (r *PostgresRepository) ListV2Events(ctx context.Context, userID, runID, afterID int64, limit int) ([]V2ProgressEvent, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, run_id, role_code, event, payload, created_at
		FROM sandbox_run_events
		WHERE user_id = $1 AND run_id = $2 AND id > $3
		ORDER BY id ASC LIMIT $4
	`, userID, runID, afterID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	events := make([]V2ProgressEvent, 0)
	for rows.Next() {
		var event V2ProgressEvent
		var payload []byte
		if err := rows.Scan(&event.ID, &event.RunID, &event.RoleCode, &event.Event, &payload, &event.CreatedAt); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(payload, &event.Payload); err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, rows.Err()
}

func appendV2EventTx(ctx context.Context, tx pgx.Tx, event V2ProgressEvent, userID int64) error {
	payload, err := json.Marshal(event.Payload)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `INSERT INTO sandbox_run_events (run_id, user_id, role_code, event, payload) VALUES ($1,$2,$3,$4,$5)`, event.RunID, userID, event.RoleCode, event.Event, payload)
	return err
}

func nullableTime(value interface{ IsZero() bool }) any {
	if value.IsZero() {
		return nil
	}
	return value
}
