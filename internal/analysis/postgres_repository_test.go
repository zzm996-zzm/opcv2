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

func TestPostgresRepositoryListsSessionsForUser(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	now := time.Now()
	db.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, user_id, mode, intent, status, questions, result, created_at, updated_at
		FROM analysis_sessions
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`)).
		WithArgs(int64(42), 20).
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "user_id", "mode", "intent", "status", "questions", "result", "created_at", "updated_at",
		}).AddRow(
			int64(99),
			int64(42),
			ModeDirection,
			"我有10年教培经验，5万本金，每周20小时",
			StatusCompleted,
			[]byte(`[]`),
			[]byte(`{"status":"completed","cards":[{"name":"本地教培小班陪跑","score":91}]}`),
			now,
			now,
		))

	repository := NewPostgresRepository(db)
	sessions, err := repository.ListSessions(context.Background(), 42, 20)
	if err != nil {
		t.Fatalf("ListSessions() error = %v", err)
	}
	if len(sessions) != 1 || sessions[0].ID != 99 || sessions[0].Result.Cards[0].Name != "本地教培小班陪跑" {
		t.Fatalf("sessions = %+v", sessions)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresRepositoryGetsSessionForUser(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	now := time.Now()
	db.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, user_id, mode, intent, status, questions, result, created_at, updated_at
		FROM analysis_sessions
		WHERE user_id = $1 AND id = $2
	`)).
		WithArgs(int64(42), int64(99)).
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "user_id", "mode", "intent", "status", "questions", "result", "created_at", "updated_at",
		}).AddRow(
			int64(99),
			int64(42),
			ModeDirection,
			"我有10年教培经验，5万本金，每周20小时",
			StatusCompleted,
			[]byte(`[]`),
			[]byte(`{"status":"completed","cards":[{"name":"本地教培小班陪跑","score":91}]}`),
			now,
			now,
		))

	repository := NewPostgresRepository(db)
	session, err := repository.GetSession(context.Background(), 42, 99)
	if err != nil {
		t.Fatalf("GetSession() error = %v", err)
	}
	if session.ID != 99 || session.UserID != 42 || session.Result.Cards[0].Name != "本地教培小班陪跑" {
		t.Fatalf("session = %+v", session)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresRepositoryCreatesAndListsActionItems(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	now := time.Now()
	db.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO analysis_action_items (user_id, session_id, day_index, title, detail, completed, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, user_id, session_id, day_index, title, detail, completed, created_at, updated_at
	`)).
		WithArgs(int64(42), int64(99), 1, "整理资源清单", "写成一页表格", false, now, now).
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "user_id", "session_id", "day_index", "title", "detail", "completed", "created_at", "updated_at",
		}).AddRow(int64(7), int64(42), int64(99), 1, "整理资源清单", "写成一页表格", false, now, now))
	db.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, user_id, session_id, day_index, title, detail, completed, created_at, updated_at
		FROM analysis_action_items
		WHERE user_id = $1 AND session_id = $2
		ORDER BY day_index ASC, id ASC
	`)).
		WithArgs(int64(42), int64(99)).
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "user_id", "session_id", "day_index", "title", "detail", "completed", "created_at", "updated_at",
		}).AddRow(int64(7), int64(42), int64(99), 1, "整理资源清单", "写成一页表格", false, now, now))

	repository := NewPostgresRepository(db)
	created, err := repository.CreateActionItems(context.Background(), []ActionItem{{
		UserID: 42, SessionID: 99, DayIndex: 1, Title: "整理资源清单", Detail: "写成一页表格", CreatedAt: now, UpdatedAt: now,
	}})
	if err != nil {
		t.Fatalf("CreateActionItems() error = %v", err)
	}
	items, err := repository.ListActionItems(context.Background(), 42, 99)
	if err != nil {
		t.Fatalf("ListActionItems() error = %v", err)
	}
	if len(created) != 1 || len(items) != 1 || items[0].Title != "整理资源清单" {
		t.Fatalf("created/items = %+v/%+v", created, items)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresRepositoryUpdatesActionItem(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	now := time.Now()
	db.ExpectQuery(regexp.QuoteMeta(`
		UPDATE analysis_action_items
		SET completed = $4, updated_at = NOW()
		WHERE user_id = $1 AND session_id = $2 AND id = $3
		RETURNING id, user_id, session_id, day_index, title, detail, completed, created_at, updated_at
	`)).
		WithArgs(int64(42), int64(99), int64(7), true).
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "user_id", "session_id", "day_index", "title", "detail", "completed", "created_at", "updated_at",
		}).AddRow(int64(7), int64(42), int64(99), 1, "整理资源清单", "写成一页表格", true, now, now))

	repository := NewPostgresRepository(db)
	item, err := repository.UpdateActionItem(context.Background(), 42, 99, 7, true)
	if err != nil {
		t.Fatalf("UpdateActionItem() error = %v", err)
	}
	if !item.Completed {
		t.Fatalf("item = %+v, want completed", item)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
