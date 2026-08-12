package projects

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestPostgresRepositoryRejectsPartialImportBatchPublicationIntegration(t *testing.T) {
	databaseURL := os.Getenv("OPCV2_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("OPCV2_TEST_DATABASE_URL is not set")
	}
	ctx := context.Background()
	db, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("pgxpool.New() error = %v", err)
	}
	defer db.Close()

	now := time.Now().UTC()
	batchNo := fmt.Sprintf("integration-partial-%d", now.UnixNano())
	validation, err := json.Marshal([]ImportValidationError{{Index: 1, Slug: "invalid slug", Field: "slug", Code: "invalid_format"}})
	if err != nil {
		t.Fatalf("marshal validation: %v", err)
	}
	var batchID int64
	err = db.QueryRow(ctx, `
		INSERT INTO project_import_batches
			(batch_no, planned_count, succeeded_count, failed_count, status, validation_errors, created_at, updated_at)
		VALUES ($1, 2, 1, 1, 'reviewing', $2, $3, $3)
		RETURNING id
	`, batchNo, validation, now).Scan(&batchID)
	if err != nil {
		t.Fatalf("insert batch: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.Exec(context.Background(), `DELETE FROM project_import_batches WHERE id = $1`, batchID)
	})

	repository := NewPostgresRepository(db)
	if _, err := repository.PublishImportBatch(ctx, batchID, 1, now.Add(time.Minute)); err != ErrPublicationGate {
		t.Fatalf("PublishImportBatch() error = %v, want ErrPublicationGate", err)
	}
	var status string
	if err := db.QueryRow(ctx, `SELECT status FROM project_import_batches WHERE id = $1`, batchID).Scan(&status); err != nil {
		t.Fatalf("read batch status: %v", err)
	}
	if status != ImportBatchReviewing {
		t.Fatalf("batch status = %q, want %q", status, ImportBatchReviewing)
	}
}

func TestPostgresRepositoryRebuildsProjectKBIntegration(t *testing.T) {
	databaseURL := os.Getenv("OPCV2_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("OPCV2_TEST_DATABASE_URL is not set")
	}
	ctx := context.Background()
	db, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("pgxpool.New() error = %v", err)
	}
	defer db.Close()

	var sourceVersion int
	if err := db.QueryRow(ctx, `SELECT active_version FROM project_kb_state WHERE singleton = TRUE`).Scan(&sourceVersion); err != nil {
		t.Fatalf("read active version: %v", err)
	}
	now := time.Now().UTC()
	taskKey := fmt.Sprintf("integration-kb-reindex-%d", now.UnixNano())
	var jobID int64
	err = db.QueryRow(ctx, `
		INSERT INTO project_kb_reindex_jobs
			(task_key, source_version, target_version, status, created_at, updated_at)
		VALUES ($1, $2, $2 + 1, 'queued', $3, $3)
		RETURNING id
	`, taskKey, sourceVersion, now).Scan(&jobID)
	if err != nil {
		t.Fatalf("insert reindex job: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.Exec(context.Background(), `DELETE FROM project_operation_audits WHERE target_type = 'kb_reindex' AND target_id = $1`, taskKey)
		_, _ = db.Exec(context.Background(), `DELETE FROM project_kb_documents WHERE version = $1`, sourceVersion+1)
		_, _ = db.Exec(context.Background(), `UPDATE project_kb_state SET active_version = $1 WHERE singleton = TRUE`, sourceVersion)
		_, _ = db.Exec(context.Background(), `DELETE FROM project_kb_reindex_jobs WHERE id = $1`, jobID)
	})

	repository := NewPostgresRepository(db)
	job, err := repository.RebuildProjectKB(ctx, jobID, now.Add(time.Minute))
	if err != nil {
		t.Fatalf("RebuildProjectKB() error = %v", err)
	}
	if job.Status != OperationsStatusReady || job.TargetVersion != sourceVersion+1 || job.DocumentCount <= 0 {
		t.Fatalf("reindex job = %+v", job)
	}
	var activeVersion, documentCount int
	if err := db.QueryRow(ctx, `
		SELECT active_version, (SELECT COUNT(*) FROM project_kb_documents WHERE version = active_version)
		FROM project_kb_state WHERE singleton = TRUE
	`).Scan(&activeVersion, &documentCount); err != nil {
		t.Fatalf("read rebuilt state: %v", err)
	}
	if activeVersion != sourceVersion+1 || documentCount != job.DocumentCount {
		t.Fatalf("active version/documents = %d/%d, job = %+v", activeVersion, documentCount, job)
	}
}
