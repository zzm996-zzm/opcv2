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
		RETURNING id, nickname, COALESCE(phone, ''), COALESCE(account, ''), COALESCE(wechat, ''), status, created_at, (xmax = 0) AS created
	`)).
		WithArgs("张晨", "13800138000", now).
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "nickname", "phone", "account", "wechat", "status", "created_at", "created",
		}).AddRow(int64(42), "张晨", "13800138000", "", "", "active", now, true))

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

func TestPostgresUserRepositoryRegisterAccount(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	now := time.Now()
	db.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO users (nickname, account, password_hash, agreement_accepted_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id, nickname, COALESCE(phone, ''), account, COALESCE(wechat, ''), status, created_at
	`)).
		WithArgs("部署测试", "deploy_user", "hashed-password", now).
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "nickname", "phone", "account", "wechat", "status", "created_at",
		}).AddRow(int64(42), "部署测试", "", "deploy_user", "", "active", now))

	repository := NewPostgresUserRepository(db)
	user, err := repository.RegisterAccount(
		context.Background(),
		"部署测试",
		"deploy_user",
		"hashed-password",
		now,
	)
	if err != nil {
		t.Fatalf("RegisterAccount() error = %v", err)
	}
	if user.ID != 42 || user.Account != "deploy_user" || user.Phone != "" {
		t.Fatalf("user = %+v", user)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresUserRepositoryFindCredentialsByAccount(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	now := time.Now()
	db.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, nickname, COALESCE(phone, ''), account, password_hash, COALESCE(wechat, ''), status, created_at
		FROM users
		WHERE account = $1
	`)).
		WithArgs("deploy_user").
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "nickname", "phone", "account", "password_hash", "wechat", "status", "created_at",
		}).AddRow(int64(42), "部署测试", "", "deploy_user", "hashed-password", "", "active", now))

	repository := NewPostgresUserRepository(db)
	credentials, err := repository.FindCredentialsByAccount(context.Background(), "deploy_user")
	if err != nil {
		t.Fatalf("FindCredentialsByAccount() error = %v", err)
	}
	if credentials.User.ID != 42 || credentials.User.Account != "deploy_user" || credentials.PasswordHash != "hashed-password" {
		t.Fatalf("credentials = %+v", credentials)
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
