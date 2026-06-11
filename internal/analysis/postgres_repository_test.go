package analysis

import (
	"context"
	"regexp"
	"testing"
	"time"

	pgxmock "github.com/pashagolub/pgxmock/v4"
)

func TestPostgresRepositoryCreatesSession(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	now := time.Now()
	db.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO analysis_sessions (user_id, mode, intent, status, questions, result, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`)).
		WithArgs(
			int64(42),
			ModeDirection,
			"我有10年教培经验，5万本金，每周20小时",
			StatusCompleted,
			"[]",
			pgxmock.AnyArg(),
			now,
		).
		WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(int64(99)))

	repository := NewPostgresRepository(db)
	session, err := repository.CreateSession(context.Background(), Session{
		UserID:    42,
		Mode:      ModeDirection,
		Intent:    "我有10年教培经验，5万本金，每周20小时",
		Status:    StatusCompleted,
		Questions: []Question{},
		Result:    DirectionResult{Status: StatusCompleted},
		CreatedAt: now,
	})
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}
	if session.ID != 99 {
		t.Fatalf("session ID = %d, want 99", session.ID)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
