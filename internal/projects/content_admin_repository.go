package projects

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func (r *PostgresRepository) IsProjectAdmin(ctx context.Context, userID int64) (bool, error) {
	var isAdmin bool
	err := r.db.QueryRow(ctx, `SELECT role = 'admin' FROM users WHERE id = $1 AND status = 'active'`, userID).Scan(&isAdmin)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	return isAdmin, err
}

func (r *PostgresRepository) CreateImportBatch(ctx context.Context, batch ImportBatch, items []ImportFailureInput) (result ImportBatch, err error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return ImportBatch{}, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()
	validationJSON, err := json.Marshal(batch.ValidationErrors)
	if err != nil {
		return ImportBatch{}, err
	}
	err = scanImportBatch(tx.QueryRow(ctx, `
		INSERT INTO project_import_batches
			(batch_no, planned_count, succeeded_count, failed_count, status, operator_id,
			 validation_errors, created_at, updated_at)
		VALUES ($1, $2, 0, $3, $4, $5, $6, $7, $7)
		RETURNING id, batch_no, planned_count, succeeded_count, failed_count, status, operator_id,
		          validation_errors, created_at, updated_at, published_at, rolled_back_at
	`, batch.BatchNo, batch.PlannedCount, batch.FailedCount, batch.Status, batch.OperatorID, validationJSON, batch.CreatedAt), &result)
	if err != nil {
		return ImportBatch{}, err
	}
	for index, item := range items {
		failureID, insertErr := insertImportFailure(ctx, tx, result.ID, item)
		if errors.Is(insertErr, pgx.ErrNoRows) {
			result.ValidationErrors = append(result.ValidationErrors, ImportValidationError{Index: index, Slug: item.Slug, Field: "slug", Code: "slug_already_exists"})
			result.FailedCount++
			continue
		}
		if insertErr != nil {
			return ImportBatch{}, insertErr
		}
		for _, source := range item.Sources {
			if _, err = tx.Exec(ctx, `
				INSERT INTO startup_failure_sources
					(failure_id, field_name, source_url, source_name, source_kind, evidence_excerpt, fetched_at)
				VALUES ($1, $2, $3, NULLIF($4, ''), $5, NULLIF($6, ''), $7)
			`, failureID, source.FieldName, source.URL, source.Name, source.Kind, source.Excerpt, batch.CreatedAt); err != nil {
				return ImportBatch{}, err
			}
		}
		result.SucceededCount++
	}
	result.UpdatedAt = batch.CreatedAt
	validationJSON, err = json.Marshal(result.ValidationErrors)
	if err != nil {
		return ImportBatch{}, err
	}
	err = scanImportBatch(tx.QueryRow(ctx, `
		UPDATE project_import_batches
		SET succeeded_count = $2, failed_count = $3, validation_errors = $4, updated_at = $5
		WHERE id = $1
		RETURNING id, batch_no, planned_count, succeeded_count, failed_count, status, operator_id,
		          validation_errors, created_at, updated_at, published_at, rolled_back_at
	`, result.ID, result.SucceededCount, result.FailedCount, validationJSON, result.UpdatedAt), &result)
	if err != nil {
		return ImportBatch{}, err
	}
	detail, _ := json.Marshal(map[string]any{"succeeded_count": result.SucceededCount, "failed_count": result.FailedCount})
	if _, err = tx.Exec(ctx, `
		INSERT INTO project_operation_audits (operator_id, action, target_type, target_id, detail, created_at)
		VALUES ($1, 'import_batch_created', 'import_batch', $2, $3, $4)
	`, batch.OperatorID, fmt.Sprint(result.ID), detail, batch.CreatedAt); err != nil {
		return ImportBatch{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return ImportBatch{}, err
	}
	return result, nil
}

func insertImportFailure(ctx context.Context, tx pgx.Tx, batchID int64, item ImportFailureInput) (int64, error) {
	learnings, err := json.Marshal(item.LearningsZH)
	if err != nil {
		return 0, err
	}
	var id int64
	err = tx.QueryRow(ctx, `
		INSERT INTO startup_failures
			(slug, name, name_zh, country_code, sector_code, product_type_code, burned_cents,
			 currency, founded_year, died_year, value_prop_zh, death_cause_zh, failure_analysis_zh,
			 learnings_zh, is_model_generated, review_status, batch_id, created_at, updated_at)
		VALUES ($1, $2, NULLIF($3, ''), NULLIF($4, ''), NULLIF($5, ''), NULLIF($6, ''), $7,
		        $8, $9, $10, $11, $12, $13, $14, FALSE, 'human_reviewed', $15, NOW(), NOW())
		ON CONFLICT (slug) DO NOTHING
		RETURNING id
	`, item.Slug, item.Name, item.NameZH, item.CountryCode, item.SectorCode, item.ProductTypeCode,
		item.BurnedCents, item.Currency, item.FoundedYear, item.DiedYear, item.ValuePropZH,
		item.DeathCauseZH, item.FailureAnalysisZH, learnings, batchID).Scan(&id)
	return id, err
}

func (r *PostgresRepository) ListImportBatches(ctx context.Context, limit int) ([]ImportBatch, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, batch_no, planned_count, succeeded_count, failed_count, status, operator_id,
		       validation_errors, created_at, updated_at, published_at, rolled_back_at
		FROM project_import_batches ORDER BY created_at DESC, id DESC LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]ImportBatch, 0)
	for rows.Next() {
		var item ImportBatch
		if err := scanImportBatch(rows, &item); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *PostgresRepository) GetImportBatch(ctx context.Context, id int64) (ImportBatch, error) {
	var item ImportBatch
	err := scanImportBatch(r.db.QueryRow(ctx, `
		SELECT id, batch_no, planned_count, succeeded_count, failed_count, status, operator_id,
		       validation_errors, created_at, updated_at, published_at, rolled_back_at
		FROM project_import_batches WHERE id = $1
	`, id), &item)
	if errors.Is(err, pgx.ErrNoRows) {
		return ImportBatch{}, ErrImportBatchNotFound
	}
	return item, err
}

func (r *PostgresRepository) PublishImportBatch(ctx context.Context, id, operatorID int64, now time.Time) (ImportBatch, error) {
	var item ImportBatch
	err := scanImportBatch(r.db.QueryRow(ctx, `
		WITH publishable AS (
			SELECT b.id FROM project_import_batches b
			WHERE b.id = $1 AND b.status = 'reviewing' AND b.succeeded_count > 0
			  AND NOT EXISTS (
				SELECT 1 FROM startup_failures f
				WHERE f.batch_id = b.id AND (
					f.review_status <> 'human_reviewed'
					OR NOT EXISTS (
						SELECT 1 FROM startup_failure_sources s
						WHERE s.failure_id = f.id AND s.source_kind IN ('primary', 'authority')
					)
					AND (SELECT COUNT(DISTINCT s.source_url) FROM startup_failure_sources s
					     WHERE s.failure_id = f.id AND s.source_kind IN ('research', 'media', 'vertical')) < 2
				)
			)
		), updated_batch AS (
			UPDATE project_import_batches b SET status = 'published', published_at = $3, updated_at = $3
			FROM publishable p WHERE b.id = p.id RETURNING b.*
		), updated_failures AS (
			UPDATE startup_failures f SET review_status = 'published', reviewed_by = $2,
			    reviewed_at = $3, updated_at = $3
			FROM updated_batch b WHERE f.batch_id = b.id RETURNING f.id
		), audit AS (
			INSERT INTO project_operation_audits (operator_id, action, target_type, target_id, detail, created_at)
			SELECT $2, 'import_batch_published', 'import_batch', id::TEXT,
			       jsonb_build_object('published_items', (SELECT COUNT(*) FROM updated_failures)), $3
			FROM updated_batch
		)
		SELECT id, batch_no, planned_count, succeeded_count, failed_count, status, operator_id,
		       validation_errors, created_at, updated_at, published_at, rolled_back_at
		FROM updated_batch
	`, id, operatorID, now), &item)
	if errors.Is(err, pgx.ErrNoRows) {
		if _, getErr := r.GetImportBatch(ctx, id); getErr != nil {
			return ImportBatch{}, getErr
		}
		return ImportBatch{}, ErrPublicationGate
	}
	return item, err
}

func (r *PostgresRepository) RollbackImportBatch(ctx context.Context, id, operatorID int64, now time.Time) (ImportBatch, error) {
	var item ImportBatch
	err := scanImportBatch(r.db.QueryRow(ctx, `
		WITH updated_batch AS (
			UPDATE project_import_batches SET status = 'rolled_back', rolled_back_at = $3, updated_at = $3
			WHERE id = $1 AND status IN ('reviewing', 'published') RETURNING *
		), rejected_failures AS (
			UPDATE startup_failures f SET review_status = 'rejected', reviewed_by = $2,
			    reviewed_at = $3, updated_at = $3
			FROM updated_batch b WHERE f.batch_id = b.id RETURNING f.id
		), offline_projects AS (
			UPDATE project_opportunities p SET status = 'offline', updated_at = $3
			WHERE p.source_failure_id IN (SELECT id FROM rejected_failures) RETURNING p.id
		), stale_opportunities AS (
			UPDATE opportunity_items o SET status = 'stale', updated_at = $3
			WHERE o.seed_failure_id IN (SELECT id FROM rejected_failures) RETURNING o.id
		), audit AS (
			INSERT INTO project_operation_audits (operator_id, action, target_type, target_id, detail, created_at)
			SELECT $2, 'import_batch_rolled_back', 'import_batch', id::TEXT,
			       jsonb_build_object('offline_projects', (SELECT COUNT(*) FROM offline_projects),
			                          'stale_opportunities', (SELECT COUNT(*) FROM stale_opportunities)), $3
			FROM updated_batch
		)
		SELECT id, batch_no, planned_count, succeeded_count, failed_count, status, operator_id,
		       validation_errors, created_at, updated_at, published_at, rolled_back_at
		FROM updated_batch
	`, id, operatorID, now), &item)
	if errors.Is(err, pgx.ErrNoRows) {
		if _, getErr := r.GetImportBatch(ctx, id); getErr != nil {
			return ImportBatch{}, getErr
		}
		return ImportBatch{}, ErrImportBatchState
	}
	return item, err
}

func (r *PostgresRepository) ListProjectOperationAudits(ctx context.Context, limit int) ([]OperationAudit, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, operator_id, action, target_type, target_id, detail, created_at
		FROM project_operation_audits ORDER BY created_at DESC, id DESC LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]OperationAudit, 0)
	for rows.Next() {
		var item OperationAudit
		var detail []byte
		if err := rows.Scan(&item.ID, &item.OperatorID, &item.Action, &item.TargetType, &item.TargetID, &detail, &item.CreatedAt); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(detail, &item.Detail); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

type importBatchScanner interface {
	Scan(dest ...any) error
}

func scanImportBatch(scanner importBatchScanner, item *ImportBatch) error {
	var validation []byte
	var operatorID pgtype.Int8
	var publishedAt, rolledBackAt pgtype.Timestamptz
	if err := scanner.Scan(&item.ID, &item.BatchNo, &item.PlannedCount, &item.SucceededCount,
		&item.FailedCount, &item.Status, &operatorID, &validation, &item.CreatedAt, &item.UpdatedAt,
		&publishedAt, &rolledBackAt); err != nil {
		return err
	}
	if operatorID.Valid {
		item.OperatorID = &operatorID.Int64
	}
	if publishedAt.Valid {
		item.PublishedAt = &publishedAt.Time
	}
	if rolledBackAt.Valid {
		item.RolledBackAt = &rolledBackAt.Time
	}
	if err := json.Unmarshal(validation, &item.ValidationErrors); err != nil {
		return err
	}
	if item.ValidationErrors == nil {
		item.ValidationErrors = []ImportValidationError{}
	}
	return nil
}
