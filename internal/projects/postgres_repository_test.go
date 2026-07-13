package projects

import (
	"context"
	"regexp"
	"testing"
	"time"

	pgxmock "github.com/pashagolub/pgxmock/v4"
)

func TestPostgresRepositoryCreatesMatchSession(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 6, 24, 9, 0, 0, 0, time.UTC)
	db.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO project_match_sessions (user_id, intent, status, questions, result, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $6)
		RETURNING id
	`)).
		WithArgs(
			int64(42),
			"我擅长内容创作，预算3万以内，每周20小时",
			StatusCompleted,
			[]byte(`[]`),
			pgxmock.AnyArg(),
			now,
		).
		WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(int64(99)))

	repository := NewPostgresRepository(db)
	session, err := repository.CreateSession(context.Background(), MatchSession{
		UserID:    42,
		Intent:    "我擅长内容创作，预算3万以内，每周20小时",
		Status:    StatusCompleted,
		Questions: []Question{},
		Result:    MatchResult{Status: StatusCompleted},
		CreatedAt: now,
	})
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}
	if session.ID != 99 {
		t.Fatalf("session.ID = %d, want 99", session.ID)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresRepositoryListsMatchSessionsForUser(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 6, 24, 9, 0, 0, 0, time.UTC)
	db.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, user_id, intent, status, questions, result, created_at, updated_at
		FROM project_match_sessions
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`)).
		WithArgs(int64(42), 20).
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "user_id", "intent", "status", "questions", "result", "created_at", "updated_at",
		}).AddRow(
			int64(99),
			int64(42),
			"我的项目",
			StatusCompleted,
			[]byte(`[]`),
			[]byte(`{"status":"completed","projects":[{"rank":1,"title":"AI短视频脚本工作室","score":94}]}`),
			now,
			now,
		))

	repository := NewPostgresRepository(db)
	sessions, err := repository.ListSessions(context.Background(), 42, 20)
	if err != nil {
		t.Fatalf("ListSessions() error = %v", err)
	}
	if len(sessions) != 1 || sessions[0].ID != 99 || sessions[0].Result.Projects[0].Title != "AI短视频脚本工作室" {
		t.Fatalf("sessions = %+v", sessions)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresRepositoryGetsMatchSessionForUser(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 6, 24, 9, 0, 0, 0, time.UTC)
	db.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, user_id, intent, status, questions, result, created_at, updated_at
		FROM project_match_sessions
		WHERE user_id = $1 AND id = $2
	`)).
		WithArgs(int64(42), int64(99)).
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "user_id", "intent", "status", "questions", "result", "created_at", "updated_at",
		}).AddRow(
			int64(99),
			int64(42),
			"我的项目",
			StatusCompleted,
			[]byte(`[]`),
			[]byte(`{"status":"completed","projects":[{"rank":1,"title":"AI短视频脚本工作室","score":94}]}`),
			now,
			now,
		))

	repository := NewPostgresRepository(db)
	session, err := repository.GetSession(context.Background(), 42, 99)
	if err != nil {
		t.Fatalf("GetSession() error = %v", err)
	}
	if session.ID != 99 || session.Result.Projects[0].Title != "AI短视频脚本工作室" {
		t.Fatalf("session = %+v", session)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresRepositorySavesFavoriteIdempotently(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 6, 24, 9, 0, 0, 0, time.UTC)
	db.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO project_match_favorites (user_id, session_id, created_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, session_id) DO UPDATE SET session_id = EXCLUDED.session_id
		RETURNING id, created_at
	`)).
		WithArgs(int64(42), int64(99), now).
		WillReturnRows(pgxmock.NewRows([]string{"id", "created_at"}).AddRow(int64(7), now))

	repository := NewPostgresRepository(db)
	favorite, err := repository.SaveFavorite(context.Background(), Favorite{
		UserID:    42,
		SessionID: 99,
		CreatedAt: now,
	})
	if err != nil {
		t.Fatalf("SaveFavorite() error = %v", err)
	}
	if favorite.ID != 7 || !favorite.CreatedAt.Equal(now) {
		t.Fatalf("favorite = %+v", favorite)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresRepositoryListsAndDeletesFavorites(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 6, 24, 9, 0, 0, 0, time.UTC)
	db.ExpectQuery("SELECT f\\.id, f\\.user_id, f\\.session_id, f\\.created_at,").
		WithArgs(int64(42), 20).
		WillReturnRows(pgxmock.NewRows([]string{
			"favorite_id", "favorite_user_id", "session_id", "favorite_created_at",
			"id", "user_id", "intent", "status", "questions", "result", "created_at", "updated_at",
		}).AddRow(
			int64(7), int64(42), int64(99), now,
			int64(99), int64(42), "线上轻资产", StatusCompleted, []byte(`[]`), []byte(`{"status":"completed","projects":[]}`), now, now,
		))
	db.ExpectExec(regexp.QuoteMeta(`DELETE FROM project_match_favorites WHERE user_id = $1 AND session_id = $2`)).
		WithArgs(int64(42), int64(99)).WillReturnResult(pgxmock.NewResult("DELETE", 1))

	repository := NewPostgresRepository(db)
	items, err := repository.ListFavorites(context.Background(), 42, 20)
	if err != nil || len(items) != 1 || items[0].Session == nil || items[0].Session.Intent != "线上轻资产" {
		t.Fatalf("favorites = %+v err=%v", items, err)
	}
	if err := repository.DeleteFavorite(context.Background(), 42, 99); err != nil {
		t.Fatalf("DeleteFavorite() error = %v", err)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
