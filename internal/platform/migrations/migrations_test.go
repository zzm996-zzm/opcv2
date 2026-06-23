package migrations

import (
	"context"
	"regexp"
	"testing"
	"testing/fstest"

	pgxmock "github.com/pashagolub/pgxmock/v4"
)

func TestRunFSAppliesPendingMigrations(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	db.ExpectBegin()
	db.ExpectExec(regexp.QuoteMeta("CREATE TABLE IF NOT EXISTS schema_migrations")).
		WillReturnResult(pgxmock.NewResult("CREATE", 0))
	db.ExpectQuery(regexp.QuoteMeta("SELECT version, dirty FROM schema_migrations ORDER BY version DESC LIMIT 1")).
		WillReturnRows(pgxmock.NewRows([]string{"version", "dirty"}).AddRow(int64(2), false))
	db.ExpectExec(regexp.QuoteMeta("DELETE FROM schema_migrations")).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))
	db.ExpectExec(regexp.QuoteMeta("INSERT INTO schema_migrations (version, dirty) VALUES ($1, TRUE)")).
		WithArgs(int64(3)).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))
	db.ExpectExec(regexp.QuoteMeta("CREATE TABLE example (id BIGSERIAL PRIMARY KEY);")).
		WillReturnResult(pgxmock.NewResult("CREATE", 0))
	db.ExpectExec(regexp.QuoteMeta("UPDATE schema_migrations SET dirty = FALSE WHERE version = $1")).
		WithArgs(int64(3)).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	db.ExpectCommit()

	result, err := RunFS(context.Background(), db, fstest.MapFS{
		"000001_first.up.sql": {Data: []byte("CREATE TABLE ignored (id BIGSERIAL PRIMARY KEY);")},
		"000003_next.up.sql":  {Data: []byte("CREATE TABLE example (id BIGSERIAL PRIMARY KEY);")},
	})
	if err != nil {
		t.Fatalf("RunFS() error = %v", err)
	}
	if result.CurrentVersion != 3 || len(result.Applied) != 1 || result.Applied[0] != 3 {
		t.Fatalf("result = %+v, want applied version 3", result)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestRunFSSkipsCurrentMigrations(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	db.ExpectBegin()
	db.ExpectExec(regexp.QuoteMeta("CREATE TABLE IF NOT EXISTS schema_migrations")).
		WillReturnResult(pgxmock.NewResult("CREATE", 0))
	db.ExpectQuery(regexp.QuoteMeta("SELECT version, dirty FROM schema_migrations ORDER BY version DESC LIMIT 1")).
		WillReturnRows(pgxmock.NewRows([]string{"version", "dirty"}).AddRow(int64(5), false))
	db.ExpectCommit()

	result, err := RunFS(context.Background(), db, fstest.MapFS{
		"000005_current.up.sql": {Data: []byte("CREATE TABLE current_version (id BIGSERIAL PRIMARY KEY);")},
	})
	if err != nil {
		t.Fatalf("RunFS() error = %v", err)
	}
	if result.CurrentVersion != 5 || len(result.Applied) != 0 {
		t.Fatalf("result = %+v, want no applied migrations at version 5", result)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestRunFSRejectsDirtySchema(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	db.ExpectBegin()
	db.ExpectExec(regexp.QuoteMeta("CREATE TABLE IF NOT EXISTS schema_migrations")).
		WillReturnResult(pgxmock.NewResult("CREATE", 0))
	db.ExpectQuery(regexp.QuoteMeta("SELECT version, dirty FROM schema_migrations ORDER BY version DESC LIMIT 1")).
		WillReturnRows(pgxmock.NewRows([]string{"version", "dirty"}).AddRow(int64(4), true))
	db.ExpectRollback()

	if _, err := RunFS(context.Background(), db, fstest.MapFS{}); err == nil {
		t.Fatal("RunFS() error = nil, want dirty schema error")
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
