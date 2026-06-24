package leads

import (
	"context"
	"regexp"
	"testing"
	"time"

	pgxmock "github.com/pashagolub/pgxmock/v4"
)

func TestPostgresRepositoryCreatesTask(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 6, 24, 10, 0, 0, 0, time.UTC)
	db.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO lead_tasks (user_id, query, status, idempotency_key, credit_cost, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $6)
		ON CONFLICT (user_id, idempotency_key) DO NOTHING
		RETURNING id, created_at, updated_at
	`)).
		WithArgs(int64(42), "成都 教培", StatusQueued, "lead-task-42", 1, now).
		WillReturnRows(pgxmock.NewRows([]string{"id", "created_at", "updated_at"}).AddRow(int64(99), now, now))

	repository := NewPostgresRepository(db)
	task, existed, err := repository.CreateTask(context.Background(), Task{
		UserID:         42,
		Query:          "成都 教培",
		Status:         StatusQueued,
		IdempotencyKey: "lead-task-42",
		CreditCost:     1,
		CreatedAt:      now,
	})
	if err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}
	if existed || task.ID != 99 {
		t.Fatalf("task=%+v existed=%v", task, existed)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresRepositoryReturnsExistingTaskOnIdempotencyConflict(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 6, 24, 10, 0, 0, 0, time.UTC)
	db.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO lead_tasks (user_id, query, status, idempotency_key, credit_cost, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $6)
		ON CONFLICT (user_id, idempotency_key) DO NOTHING
		RETURNING id, created_at, updated_at
	`)).
		WithArgs(int64(42), "成都 教培", StatusQueued, "lead-task-42", 1, now).
		WillReturnRows(pgxmock.NewRows([]string{"id", "created_at", "updated_at"}))
	db.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, user_id, query, status, idempotency_key, credit_cost, error_code, created_at, updated_at
		FROM lead_tasks
		WHERE user_id = $1 AND idempotency_key = $2
	`)).
		WithArgs(int64(42), "lead-task-42").
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "user_id", "query", "status", "idempotency_key", "credit_cost", "error_code", "created_at", "updated_at",
		}).AddRow(int64(99), int64(42), "成都 教培", StatusQueued, "lead-task-42", 1, "", now, now))

	repository := NewPostgresRepository(db)
	task, existed, err := repository.CreateTask(context.Background(), Task{
		UserID:         42,
		Query:          "成都 教培",
		Status:         StatusQueued,
		IdempotencyKey: "lead-task-42",
		CreditCost:     1,
		CreatedAt:      now,
	})
	if err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}
	if !existed || task.ID != 99 {
		t.Fatalf("task=%+v existed=%v", task, existed)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresRepositoryStoresLeadResults(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	db.ExpectBegin()
	db.ExpectExec(regexp.QuoteMeta(`DELETE FROM lead_results WHERE task_id = $1`)).
		WithArgs(int64(99)).
		WillReturnResult(pgxmock.NewResult("DELETE", 0))
	db.ExpectExec(regexp.QuoteMeta(`
		INSERT INTO lead_results (task_id, name, phone, email, website, evidence)
		VALUES ($1, $2, $3, $4, $5, $6)
	`)).
		WithArgs(int64(99), "成都启明星教育", "028-12345678", "", "https://example.com", pgxmock.AnyArg()).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))
	db.ExpectCommit()

	repository := NewPostgresRepository(db)
	err = repository.StoreResults(context.Background(), 99, []Lead{{
		Name:    "成都启明星教育",
		Phone:   "028-12345678",
		Website: "https://example.com",
		Evidence: []Evidence{{
			Type:  "website",
			Title: "官网",
			URL:   "https://example.com",
		}},
	}})
	if err != nil {
		t.Fatalf("StoreResults() error = %v", err)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
