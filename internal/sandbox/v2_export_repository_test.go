package sandbox

import (
	"context"
	"regexp"
	"testing"
	"time"

	pgxmock "github.com/pashagolub/pgxmock/v4"
)

func TestPostgresRepositoryStoresPDFPayloadAsBytes(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	now := time.Date(2026, 8, 10, 12, 0, 0, 0, time.UTC)
	expires := now.Add(7 * 24 * time.Hour)
	payload := []byte("%PDF-1.7\nrendered")
	db.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO sandbox_exports (run_id, user_id, format, payload, payload_bytes, expires_at, created_at)
		SELECT $1, $2, $3, $4, $5, $6, COALESCE($7, NOW())
		WHERE EXISTS(SELECT 1 FROM sandbox_sessions WHERE id = $1 AND user_id = $2 AND sandbox_version = 2)
		RETURNING id, created_at
	`)).
		WithArgs(int64(99), int64(42), "pdf", []byte(`{}`), payload, expires, now).
		WillReturnRows(pgxmock.NewRows([]string{"id", "created_at"}).AddRow(int64(700), now))

	repository := NewPostgresRepository(db)
	export, err := repository.CreateV2Export(context.Background(), V2Export{RunID: 99, Format: "pdf", ExpiresAt: expires, CreatedAt: now}, 42, payload)
	if err != nil {
		t.Fatalf("CreateV2Export() error = %v", err)
	}
	if export.ID != 700 {
		t.Fatalf("export = %+v", export)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresRepositoryReadsPDFPayloadBytes(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	now := time.Date(2026, 8, 10, 12, 0, 0, 0, time.UTC)
	payload := []byte("%PDF-1.7\nrendered")
	db.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, run_id, format, payload, payload_bytes, expires_at, created_at
		FROM sandbox_exports WHERE id = $1 AND user_id = $2
	`)).
		WithArgs(int64(700), int64(42)).
		WillReturnRows(pgxmock.NewRows([]string{"id", "run_id", "format", "payload", "payload_bytes", "expires_at", "created_at"}).
			AddRow(int64(700), int64(99), "pdf", []byte(`{}`), payload, now.Add(time.Hour), now))

	repository := NewPostgresRepository(db)
	export, got, err := repository.GetV2Export(context.Background(), 42, 700)
	if err != nil {
		t.Fatalf("GetV2Export() error = %v", err)
	}
	if export.Format != "pdf" || string(got) != string(payload) {
		t.Fatalf("export=%+v payload=%q", export, got)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
