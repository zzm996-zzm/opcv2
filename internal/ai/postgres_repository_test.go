package ai

import (
	"context"
	"regexp"
	"testing"
	"time"

	pgxmock "github.com/pashagolub/pgxmock/v4"
)

func TestPostgresRepositoryCreatesRun(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	createdAt := time.Date(2026, 6, 23, 9, 30, 0, 0, time.UTC)
	db.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO ai_runs (
			user_id,
			feature,
			prompt_version,
			provider,
			model,
			status,
			request,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $8)
		RETURNING id
	`)).
		WithArgs(
			int64(42),
			"analysis.direction",
			"direction_v1",
			"development",
			"dev-model",
			StatusPending,
			[]byte(`{"intent":"开一个咖啡店"}`),
			createdAt,
		).
		WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(int64(11)))

	repository := NewPostgresRepository(db)
	run, err := repository.CreateRun(context.Background(), Run{
		UserID:        42,
		Feature:       "analysis.direction",
		PromptVersion: "direction_v1",
		Provider:      "development",
		Model:         "dev-model",
		Status:        StatusPending,
		Request:       []byte(`{"intent":"开一个咖啡店"}`),
		CreatedAt:     createdAt,
	})
	if err != nil {
		t.Fatalf("CreateRun() error = %v", err)
	}
	if run.ID != 11 {
		t.Fatalf("run.ID = %d, want 11", run.ID)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresRepositoryCompletesRun(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	db.ExpectExec(regexp.QuoteMeta(`
		UPDATE ai_runs
		SET status = $2,
		    response = $3,
		    input_tokens = $4,
		    output_tokens = $5,
		    latency_ms = $6,
		    error_code = '',
		    error_message = '',
		    updated_at = NOW()
		WHERE id = $1
	`)).
		WithArgs(
			int64(11),
			StatusCompleted,
			[]byte(`{"cards":[{"name":"AI短视频脚本工作室"}]}`),
			123,
			456,
			789,
		).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	repository := NewPostgresRepository(db)
	err = repository.CompleteRun(context.Background(), 11, RunResult{
		Response:     []byte(`{"cards":[{"name":"AI短视频脚本工作室"}]}`),
		InputTokens:  123,
		OutputTokens: 456,
		LatencyMS:    789,
	})
	if err != nil {
		t.Fatalf("CompleteRun() error = %v", err)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresRepositoryFailsRun(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	db.ExpectExec(regexp.QuoteMeta(`
		UPDATE ai_runs
		SET status = $2,
		    error_code = $3,
		    error_message = $4,
		    latency_ms = $5,
		    updated_at = NOW()
		WHERE id = $1
	`)).
		WithArgs(
			int64(11),
			StatusFailed,
			"invalid_model_json",
			"AI response did not match the expected schema",
			321,
		).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	repository := NewPostgresRepository(db)
	err = repository.FailRun(context.Background(), 11, RunFailure{
		Code:      "invalid_model_json",
		Message:   "AI response did not match the expected schema",
		LatencyMS: 321,
	})
	if err != nil {
		t.Fatalf("FailRun() error = %v", err)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
