package auth

import (
	"context"
	"regexp"
	"testing"
	"time"

	pgxmock "github.com/pashagolub/pgxmock/v4"
)

func TestPostgresUserRepositoryFindOrCreateByPhone(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	now := time.Now()
	db.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO users (nickname, phone, agreement_accepted_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (phone) DO UPDATE SET updated_at = users.updated_at
		RETURNING id, nickname, phone, COALESCE(wechat, ''), status, created_at, (xmax = 0) AS created
	`)).
		WithArgs("张晨", "13800138000", now).
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "nickname", "phone", "wechat", "status", "created_at", "created",
		}).AddRow(int64(42), "张晨", "13800138000", "", "active", now, true))

	repository := NewPostgresUserRepository(db)
	user, created, err := repository.FindOrCreateByPhone(
		context.Background(),
		"张晨",
		"13800138000",
		now,
	)
	if err != nil {
		t.Fatalf("FindOrCreateByPhone() error = %v", err)
	}
	if user.ID != 42 || !created {
		t.Fatalf("user/created = %+v/%v", user, created)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresUserRepositoryRecordsLogin(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	db.ExpectExec(regexp.QuoteMeta(`
		INSERT INTO login_records (user_id, ip, user_agent)
		VALUES ($1, NULLIF($2, '')::INET, $3)
	`)).
		WithArgs(int64(42), "127.0.0.1", "test-agent").
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	repository := NewPostgresUserRepository(db)
	err = repository.RecordLogin(context.Background(), 42, LoginMeta{
		IP:        "127.0.0.1",
		UserAgent: "test-agent",
	})
	if err != nil {
		t.Fatalf("RecordLogin() error = %v", err)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
