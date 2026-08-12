package projects

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func (r *PostgresRepository) GetActiveProjectKBVersion(ctx context.Context) (int, error) {
	var version int
	err := r.db.QueryRow(ctx, `SELECT active_version FROM project_kb_state WHERE singleton = TRUE`).Scan(&version)
	return version, err
}

func (r *PostgresRepository) CreateProjectKBReindex(ctx context.Context, userID int64, taskKey string, now time.Time) (KBReindexJob, error) {
	var item KBReindexJob
	err := scanKBReindexJob(r.db.QueryRow(ctx, `
		WITH inserted AS (
			INSERT INTO project_kb_reindex_jobs
				(task_key, requested_by, source_version, target_version, status, created_at, updated_at)
			SELECT $2, $1, active_version, active_version + 1, 'queued', $3, $3
			FROM project_kb_state WHERE singleton = TRUE
			ON CONFLICT (task_key) DO NOTHING
			RETURNING *, TRUE AS created
		)
		SELECT id, task_key, requested_by, source_version, target_version, status,
		       document_count, error_code, created_at, started_at, finished_at, updated_at, created
		FROM inserted
		UNION ALL
		SELECT id, task_key, requested_by, source_version, target_version, status,
		       document_count, error_code, created_at, started_at, finished_at, updated_at, FALSE
		FROM project_kb_reindex_jobs WHERE task_key = $2 AND NOT EXISTS (SELECT 1 FROM inserted)
	`, userID, taskKey, now), &item)
	return item, err
}

func (r *PostgresRepository) GetProjectKBReindex(ctx context.Context, id int64) (KBReindexJob, error) {
	var item KBReindexJob
	err := scanKBReindexJob(r.db.QueryRow(ctx, `
		SELECT id, task_key, requested_by, source_version, target_version, status,
		       document_count, error_code, created_at, started_at, finished_at, updated_at, FALSE
		FROM project_kb_reindex_jobs WHERE id = $1
	`, id), &item)
	if errors.Is(err, pgx.ErrNoRows) {
		return KBReindexJob{}, ErrKBReindexNotFound
	}
	return item, err
}

func (r *PostgresRepository) MarkProjectKBReindexFailed(ctx context.Context, id int64, code string, now time.Time) error {
	command, err := r.db.Exec(ctx, `
		UPDATE project_kb_reindex_jobs
		SET status = 'failed', error_code = $2, finished_at = $3, updated_at = $3
		WHERE id = $1 AND status IN ('queued', 'running')
	`, id, code, now)
	if err != nil {
		return err
	}
	if command.RowsAffected() == 0 {
		if _, getErr := r.GetProjectKBReindex(ctx, id); getErr != nil {
			return getErr
		}
	}
	return nil
}

func (r *PostgresRepository) RebuildProjectKB(ctx context.Context, id int64, now time.Time) (result KBReindexJob, err error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return KBReindexJob{}, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()
	var activeVersion int
	if err = tx.QueryRow(ctx, `SELECT active_version FROM project_kb_state WHERE singleton = TRUE FOR UPDATE`).Scan(&activeVersion); err != nil {
		return KBReindexJob{}, err
	}
	err = scanKBReindexJob(tx.QueryRow(ctx, `
		SELECT id, task_key, requested_by, source_version, target_version, status,
		       document_count, error_code, created_at, started_at, finished_at, updated_at, FALSE
		FROM project_kb_reindex_jobs WHERE id = $1 FOR UPDATE
	`, id), &result)
	if errors.Is(err, pgx.ErrNoRows) {
		return KBReindexJob{}, ErrKBReindexNotFound
	}
	if err != nil {
		return KBReindexJob{}, err
	}
	if result.Status == OperationsStatusReady {
		if err = tx.Commit(ctx); err != nil {
			return KBReindexJob{}, err
		}
		return result, nil
	}
	if result.Status != OperationsStatusQueued || result.SourceVersion != activeVersion || result.TargetVersion != activeVersion+1 {
		return KBReindexJob{}, ErrInvalidKBReindex
	}
	if _, err = tx.Exec(ctx, `UPDATE project_kb_reindex_jobs SET status = 'running', started_at = $2, updated_at = $2 WHERE id = $1`, id, now); err != nil {
		return KBReindexJob{}, err
	}
	if _, err = tx.Exec(ctx, `DELETE FROM project_kb_documents WHERE version = $1`, result.TargetVersion); err != nil {
		return KBReindexJob{}, err
	}
	command, err := tx.Exec(ctx, `
		INSERT INTO project_kb_documents (version, document_id, title, body, metadata, created_at)
		SELECT $1::INTEGER, slug, title,
		       concat_ws(' ', summary, industry, tags::TEXT, budget_band, difficulty,
		                 resource_requirements::TEXT, detail::TEXT),
		       jsonb_build_object('slug', slug, 'kind', 'project', 'knowledge_base_version', ($1::INTEGER)::TEXT), $2
		FROM project_opportunities WHERE status = 'published'
	`, result.TargetVersion, now)
	if err != nil {
		return KBReindexJob{}, err
	}
	result.DocumentCount = int(command.RowsAffected())
	if _, err = tx.Exec(ctx, `UPDATE project_kb_state SET active_version = $1, updated_at = $2 WHERE singleton = TRUE`, result.TargetVersion, now); err != nil {
		return KBReindexJob{}, err
	}
	if _, err = tx.Exec(ctx, `DELETE FROM project_ai_response_cache WHERE kb_version <> $1`, result.TargetVersion); err != nil {
		return KBReindexJob{}, err
	}
	err = scanKBReindexJob(tx.QueryRow(ctx, `
		UPDATE project_kb_reindex_jobs
		SET status = 'ready', document_count = $2, error_code = '', finished_at = $3, updated_at = $3
		WHERE id = $1
		RETURNING id, task_key, requested_by, source_version, target_version, status,
		          document_count, error_code, created_at, started_at, finished_at, updated_at, FALSE
	`, id, result.DocumentCount, now), &result)
	if err != nil {
		return KBReindexJob{}, err
	}
	detail, _ := json.Marshal(map[string]any{"source_version": result.SourceVersion, "target_version": result.TargetVersion, "document_count": result.DocumentCount})
	if _, err = tx.Exec(ctx, `
		INSERT INTO project_operation_audits (operator_id, action, target_type, target_id, detail, created_at)
		VALUES ($1, 'kb_reindexed', 'kb_reindex', $2, $3, $4)
	`, result.RequestedBy, result.TaskKey, detail, now); err != nil {
		return KBReindexJob{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return KBReindexJob{}, err
	}
	return result, nil
}

func (r *PostgresRepository) SaveProjectAIAnswer(ctx context.Context, item ProjectAIAnswer) error {
	parsed, err := json.Marshal(item.ParsedIntent)
	if err != nil {
		return err
	}
	chunks, err := json.Marshal(item.CitedChunkIDs)
	if err != nil {
		return err
	}
	web, err := json.Marshal(item.CitedWebSourceIDs)
	if err != nil {
		return err
	}
	result, err := json.Marshal(item.Result)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(ctx, `
		INSERT INTO project_ai_answers
			(user_id, entry, raw_input, parsed_intent, cited_chunk_ids, cited_web_source_ids,
			 kb_sufficiency, result, model_name, prompt_version, kb_version, cache_key,
			 latency_ms, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`, item.UserID, item.Entry, item.RawInput, parsed, chunks, web, item.KBSufficiency, result,
		item.ModelName, item.PromptVersion, item.KBVersion, item.CacheKey, item.LatencyMS,
		normalizeAIAnswerStatus(item.Status), item.CreatedAt)
	return err
}

func (r *PostgresRepository) ListProjectAIAnswers(ctx context.Context, limit int) ([]ProjectAIAnswer, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, user_id, entry, raw_input, parsed_intent, cited_chunk_ids, cited_web_source_ids,
		       kb_sufficiency, result, model_name, prompt_version, kb_version, cache_key,
		       latency_ms, status, created_at
		FROM project_ai_answers ORDER BY created_at DESC, id DESC LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]ProjectAIAnswer, 0)
	for rows.Next() {
		var item ProjectAIAnswer
		var parsed, chunks, web, result []byte
		var userID pgtype.Int8
		var sufficiency pgtype.Float8
		if err := rows.Scan(&item.ID, &userID, &item.Entry, &item.RawInput, &parsed, &chunks, &web,
			&sufficiency, &result, &item.ModelName, &item.PromptVersion, &item.KBVersion,
			&item.CacheKey, &item.LatencyMS, &item.Status, &item.CreatedAt); err != nil {
			return nil, err
		}
		if userID.Valid {
			item.UserID = &userID.Int64
		}
		if sufficiency.Valid {
			item.KBSufficiency = &sufficiency.Float64
		}
		if err := json.Unmarshal(parsed, &item.ParsedIntent); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(chunks, &item.CitedChunkIDs); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(web, &item.CitedWebSourceIDs); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(result, &item.Result); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *PostgresRepository) CreateContentProductionRun(ctx context.Context, taskKey, date string, planned int, now time.Time) (ContentProductionRun, error) {
	var item ContentProductionRun
	err := scanContentProductionRun(r.db.QueryRow(ctx, `
		WITH inserted AS (
			INSERT INTO project_content_production_runs
				(task_key, schedule_date, status, planned_count, created_at, updated_at)
			VALUES ($1, $2::DATE, 'queued', $3, $4, $4)
			ON CONFLICT (schedule_date) DO NOTHING
			RETURNING *, TRUE AS created
		)
		SELECT id, task_key, schedule_date::TEXT, status, planned_count, produced_count,
		       error_code, created_at, started_at, finished_at, updated_at, created
		FROM inserted
		UNION ALL
		SELECT id, task_key, schedule_date::TEXT, status, planned_count, produced_count,
		       error_code, created_at, started_at, finished_at, updated_at, FALSE
		FROM project_content_production_runs WHERE schedule_date = $2::DATE AND NOT EXISTS (SELECT 1 FROM inserted)
	`, taskKey, date, planned, now), &item)
	return item, err
}

func (r *PostgresRepository) ProduceProjectContent(ctx context.Context, id int64, now time.Time) (ContentProductionRun, error) {
	var item ContentProductionRun
	err := scanContentProductionRun(r.db.QueryRow(ctx, `
		WITH claimed AS (
			UPDATE project_content_production_runs SET status = 'running', started_at = $2, updated_at = $2
			WHERE id = $1 AND status = 'queued' RETURNING *
		), candidates AS (
			SELECT f.id, f.name, COALESCE(f.name_zh, f.name) AS name_zh, f.value_prop_zh,
			       f.death_cause_zh, p.concept_zh, p.market_potential, p.difficulty,
			       p.execution_phases, p.monetization_zh, l.cn_budget_band
			FROM startup_failures f
			JOIN claimed c ON TRUE
			JOIN rebuild_plans p ON p.failure_id = f.id
			JOIN rebuild_localizations l ON l.plan_id = p.id
			WHERE f.review_status = 'published'
			  AND NOT EXISTS (SELECT 1 FROM opportunity_items o WHERE o.seed_failure_id = f.id)
			ORDER BY f.updated_at ASC, f.id ASC
			LIMIT (SELECT planned_count FROM claimed)
		), produced AS (
			INSERT INTO opportunity_items
				(seed_failure_id, group_codes, title, summary, opportunity_reason, why_now,
				 target_user, budget_band, difficulty, risks, status, created_at, updated_at)
			SELECT id, '["failure_lessons"]'::JSONB, concept_zh,
			       COALESCE(value_prop_zh, name_zh),
			       concat('基于 ', name_zh, ' 的失败教训：', COALESCE(death_cause_zh, '待复核')),
			       COALESCE(monetization_zh, ''), '待运营复核', cn_budget_band,
			       COALESCE(difficulty::TEXT, ''), jsonb_build_array(COALESCE(death_cause_zh, '待补充风险')),
			       'reviewing', $2, $2
			FROM candidates RETURNING id
		), completed AS (
			UPDATE project_content_production_runs r
			SET status = 'ready', produced_count = (SELECT COUNT(*) FROM produced),
			    finished_at = $2, updated_at = $2
			FROM claimed c WHERE r.id = c.id RETURNING r.*
		)
		SELECT id, task_key, schedule_date::TEXT, status, planned_count, produced_count,
		       error_code, created_at, started_at, finished_at, updated_at, FALSE
		FROM completed
	`, id, now), &item)
	if errors.Is(err, pgx.ErrNoRows) {
		return ContentProductionRun{}, ErrInvalidKBReindex
	}
	return item, err
}

func (r *PostgresRepository) MarkContentProductionFailed(ctx context.Context, id int64, code string, now time.Time) error {
	command, err := r.db.Exec(ctx, `
		UPDATE project_content_production_runs
		SET status = 'failed', error_code = $2, finished_at = $3, updated_at = $3
		WHERE id = $1 AND status IN ('queued', 'running')
	`, id, code, now)
	if err != nil {
		return err
	}
	if command.RowsAffected() == 0 {
		return ErrInvalidKBReindex
	}
	return nil
}

type kbOperationScanner interface{ Scan(dest ...any) error }

func scanKBReindexJob(scanner kbOperationScanner, item *KBReindexJob) error {
	var requestedBy pgtype.Int8
	var startedAt, finishedAt pgtype.Timestamptz
	if err := scanner.Scan(&item.ID, &item.TaskKey, &requestedBy, &item.SourceVersion, &item.TargetVersion,
		&item.Status, &item.DocumentCount, &item.ErrorCode, &item.CreatedAt, &startedAt, &finishedAt,
		&item.UpdatedAt, &item.created); err != nil {
		return err
	}
	if requestedBy.Valid {
		item.RequestedBy = &requestedBy.Int64
	}
	if startedAt.Valid {
		item.StartedAt = &startedAt.Time
	}
	if finishedAt.Valid {
		item.FinishedAt = &finishedAt.Time
	}
	return nil
}

func scanContentProductionRun(scanner kbOperationScanner, item *ContentProductionRun) error {
	var startedAt, finishedAt pgtype.Timestamptz
	if err := scanner.Scan(&item.ID, &item.TaskKey, &item.ScheduleDate, &item.Status,
		&item.PlannedCount, &item.ProducedCount, &item.ErrorCode, &item.CreatedAt,
		&startedAt, &finishedAt, &item.UpdatedAt, &item.created); err != nil {
		return err
	}
	if startedAt.Valid {
		item.StartedAt = &startedAt.Time
	}
	if finishedAt.Valid {
		item.FinishedAt = &finishedAt.Time
	}
	return nil
}
