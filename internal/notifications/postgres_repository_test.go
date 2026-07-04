package notifications

import (
	"context"
	"regexp"
	"testing"
	"time"

	pgxmock "github.com/pashagolub/pgxmock/v4"
)

func TestPostgresRepositoryListsNotifications(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 7, 2, 10, 0, 0, 0, time.UTC)
	db.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, user_id, type, title, summary, body, source_type, source_id, action_label, action_url, read_at, created_at
		FROM notifications
		WHERE user_id = $1
		  AND ($2 = '' OR type = $2)
		  AND ($3 = 'all' OR ($3 = 'unread' AND read_at IS NULL) OR ($3 = 'read' AND read_at IS NOT NULL))
		ORDER BY created_at DESC
		LIMIT $4
	`)).
		WithArgs(int64(42), TypeTask, StatusUnread, 20).
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "user_id", "type", "title", "summary", "body", "source_type", "source_id", "action_label", "action_url", "read_at", "created_at",
		}).AddRow(
			int64(1), int64(42), TypeTask, "任务", "摘要", "正文", "task", int64(9), "查看", "/tasks", nil, now,
		))

	repository := NewPostgresRepository(db)
	rows, err := repository.ListNotifications(context.Background(), 42, ListFilters{Type: TypeTask, Status: StatusUnread, Limit: 20})
	if err != nil {
		t.Fatalf("ListNotifications() error = %v", err)
	}
	if len(rows) != 1 || rows[0].SourceID == nil || *rows[0].SourceID != 9 {
		t.Fatalf("rows = %+v", rows)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresRepositoryMarksRead(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	readAt := time.Date(2026, 7, 2, 10, 0, 0, 0, time.UTC)
	db.ExpectQuery(regexp.QuoteMeta(`
		UPDATE notifications
		SET read_at = COALESCE(read_at, NOW())
		WHERE user_id = $1 AND id = $2
		RETURNING id, user_id, type, title, summary, body, source_type, source_id, action_label, action_url, read_at, created_at
	`)).
		WithArgs(int64(42), int64(99)).
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "user_id", "type", "title", "summary", "body", "source_type", "source_id", "action_label", "action_url", "read_at", "created_at",
		}).AddRow(
			int64(99), int64(42), TypeTask, "任务", "", "", "", nil, "", "", readAt, readAt,
		))

	repository := NewPostgresRepository(db)
	row, err := repository.MarkRead(context.Background(), 42, 99)
	if err != nil {
		t.Fatalf("MarkRead() error = %v", err)
	}
	if row.ReadAt == nil || !row.ReadAt.Equal(readAt) {
		t.Fatalf("row = %+v", row)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresRepositorySummary(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	db.ExpectQuery(regexp.QuoteMeta(`
		SELECT COUNT(*)
		FROM notifications
		WHERE user_id = $1 AND read_at IS NULL
	`)).
		WithArgs(int64(42)).
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(3))
	db.ExpectQuery(regexp.QuoteMeta(`
		SELECT type, COUNT(*)
		FROM notifications
		WHERE user_id = $1 AND read_at IS NULL
		GROUP BY type
		ORDER BY type
	`)).
		WithArgs(int64(42)).
		WillReturnRows(pgxmock.NewRows([]string{"type", "count"}).AddRow(TypeTask, 2).AddRow(TypeCRM, 1))
	db.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, user_id, type, title, summary, body, source_type, source_id, action_label, action_url, read_at, created_at
		FROM notifications
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT 5
	`)).
		WithArgs(int64(42)).
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "user_id", "type", "title", "summary", "body", "source_type", "source_id", "action_label", "action_url", "read_at", "created_at",
		}))

	repository := NewPostgresRepository(db)
	summary, err := repository.Summary(context.Background(), 42)
	if err != nil {
		t.Fatalf("Summary() error = %v", err)
	}
	if summary.Unread != 3 || len(summary.ByType) != 2 || summary.Latest == nil {
		t.Fatalf("summary = %+v", summary)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
