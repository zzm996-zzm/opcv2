package projects

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
)

func (r *PostgresRepository) PrepareMatchGeneration(ctx context.Context, userID, matchID int64) (MatchRun, error) {
	var attempt int
	err := r.db.QueryRow(ctx, `
		WITH updated AS (
			UPDATE project_match_sessions
			SET status = 'queued', generation_attempt = generation_attempt + 1,
			    progress_percent = 0, current_step = 'queued', error_code = NULL,
			    cancelled_at = NULL, result = '{}'::JSONB, updated_at = NOW()
			WHERE user_id = $1 AND id = $2 AND workflow_version = 2
			  AND status IN ('ready', 'failed', 'canceled', 'partial')
			RETURNING id, user_id, generation_attempt
		)
		INSERT INTO project_match_progress_events
			(match_id, user_id, attempt, event, progress_percent, payload)
		SELECT id, user_id, generation_attempt, 'queued', 0, '{}'::JSONB FROM updated
		RETURNING attempt
	`, userID, matchID).Scan(&attempt)
	if errors.Is(err, pgx.ErrNoRows) {
		return MatchRun{}, ErrMatchNotReady
	}
	if err != nil {
		return MatchRun{}, err
	}
	return r.GetMatchRun(ctx, userID, matchID)
}

func (r *PostgresRepository) UpdateMatchGeneration(ctx context.Context, userID, matchID int64, attempt int, status string, progress int, step, errorCode string, result *MatchResult) (MatchRun, error) {
	resultJSON := []byte(`{}`)
	hasResult := result != nil
	if result != nil {
		var err error
		resultJSON, err = json.Marshal(result)
		if err != nil {
			return MatchRun{}, err
		}
	}
	var eventID int64
	err := r.db.QueryRow(ctx, `
		WITH updated AS (
			UPDATE project_match_sessions
			SET status = $4, progress_percent = $5, current_step = $6,
			    error_code = NULLIF($7, ''),
			    result = CASE WHEN $9 THEN $8::JSONB ELSE result END,
			    updated_at = NOW()
			WHERE user_id = $1 AND id = $2 AND workflow_version = 2
			  AND generation_attempt = $3 AND status IN ('queued', 'running')
			RETURNING id, user_id
		)
		INSERT INTO project_match_progress_events
			(match_id, user_id, attempt, event, progress_percent, payload)
		SELECT id, user_id, $3, $6, $5, '{}'::JSONB FROM updated
		RETURNING id
	`, userID, matchID, attempt, status, progress, step, errorCode, resultJSON, hasResult).Scan(&eventID)
	if errors.Is(err, pgx.ErrNoRows) {
		return MatchRun{}, ErrStaleMatchGeneration
	}
	if err != nil {
		return MatchRun{}, err
	}
	return r.GetMatchRun(ctx, userID, matchID)
}

func (r *PostgresRepository) CancelMatchGeneration(ctx context.Context, userID, matchID int64) (MatchRun, error) {
	var eventID int64
	err := r.db.QueryRow(ctx, `
		WITH updated AS (
			UPDATE project_match_sessions
			SET status = 'canceled', progress_percent = 100, current_step = 'canceled',
			    cancelled_at = NOW(), updated_at = NOW()
			WHERE user_id = $1 AND id = $2 AND workflow_version = 2
			  AND status IN ('queued', 'running')
			RETURNING id, user_id, generation_attempt
		)
		INSERT INTO project_match_progress_events
			(match_id, user_id, attempt, event, progress_percent, payload)
		SELECT id, user_id, generation_attempt, 'canceled', 100, '{}'::JSONB FROM updated
		RETURNING id
	`, userID, matchID).Scan(&eventID)
	if errors.Is(err, pgx.ErrNoRows) {
		return MatchRun{}, ErrMatchNotReady
	}
	if err != nil {
		return MatchRun{}, err
	}
	return r.GetMatchRun(ctx, userID, matchID)
}

func (r *PostgresRepository) ListMatchProgressEvents(ctx context.Context, userID, matchID, afterID int64) ([]MatchProgressEvent, error) {
	rows, err := r.db.Query(ctx, `
		SELECT e.id, e.match_id, e.attempt, e.event, e.progress_percent, e.payload, e.created_at
		FROM project_match_progress_events e
		JOIN project_match_sessions s ON s.id = e.match_id AND s.user_id = e.user_id
		WHERE e.user_id = $1 AND e.match_id = $2 AND e.id > $3 AND s.workflow_version = 2
		ORDER BY e.id ASC
	`, userID, matchID, afterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	events := make([]MatchProgressEvent, 0)
	for rows.Next() {
		var event MatchProgressEvent
		var payload []byte
		if err := rows.Scan(&event.ID, &event.MatchID, &event.Attempt, &event.Event, &event.ProgressPercent, &payload, &event.CreatedAt); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(payload, &event.Payload); err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, rows.Err()
}
