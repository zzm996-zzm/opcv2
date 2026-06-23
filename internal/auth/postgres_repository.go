package auth

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type postgresDB interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

type PostgresUserRepository struct {
	db postgresDB
}

func NewPostgresUserRepository(db postgresDB) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

func (r *PostgresUserRepository) FindOrCreateByPhone(
	ctx context.Context,
	nickname string,
	phone string,
	agreementAcceptedAt time.Time,
) (User, bool, error) {
	const query = `
		INSERT INTO users (nickname, phone, agreement_accepted_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (phone) DO UPDATE SET updated_at = users.updated_at
		RETURNING id, nickname, COALESCE(phone, ''), COALESCE(account, ''), COALESCE(wechat, ''), status, created_at, (xmax = 0) AS created
	`
	var user User
	var created bool
	err := r.db.QueryRow(ctx, query, nickname, phone, agreementAcceptedAt).Scan(
		&user.ID,
		&user.Nickname,
		&user.Phone,
		&user.Account,
		&user.Wechat,
		&user.Status,
		&user.CreatedAt,
		&created,
	)
	return user, created, err
}

func (r *PostgresUserRepository) RegisterAccount(
	ctx context.Context,
	nickname string,
	account string,
	passwordHash string,
	agreementAcceptedAt time.Time,
) (User, error) {
	const query = `
		INSERT INTO users (nickname, account, password_hash, agreement_accepted_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id, nickname, COALESCE(phone, ''), account, COALESCE(wechat, ''), status, created_at
	`
	var user User
	err := r.db.QueryRow(ctx, query, nickname, account, passwordHash, agreementAcceptedAt).Scan(
		&user.ID,
		&user.Nickname,
		&user.Phone,
		&user.Account,
		&user.Wechat,
		&user.Status,
		&user.CreatedAt,
	)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return User{}, ErrAccountExists
	}
	return user, err
}

func (r *PostgresUserRepository) FindCredentialsByAccount(ctx context.Context, account string) (UserCredentials, error) {
	const query = `
		SELECT id, nickname, COALESCE(phone, ''), account, password_hash, COALESCE(wechat, ''), status, created_at
		FROM users
		WHERE account = $1
	`
	var credentials UserCredentials
	err := r.db.QueryRow(ctx, query, account).Scan(
		&credentials.User.ID,
		&credentials.User.Nickname,
		&credentials.User.Phone,
		&credentials.User.Account,
		&credentials.PasswordHash,
		&credentials.User.Wechat,
		&credentials.User.Status,
		&credentials.User.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return UserCredentials{}, ErrUserNotFound
	}
	return credentials, err
}

func (r *PostgresUserRepository) RecordLogin(ctx context.Context, userID int64, meta LoginMeta) error {
	const query = `
		INSERT INTO login_records (user_id, ip, user_agent)
		VALUES ($1, NULLIF($2, '')::INET, $3)
	`
	_, err := r.db.Exec(ctx, query, userID, meta.IP, meta.UserAgent)
	return err
}

func (r *PostgresUserRepository) FindByID(ctx context.Context, userID int64) (User, error) {
	const query = `
		SELECT id, nickname, COALESCE(phone, ''), COALESCE(account, ''), COALESCE(wechat, ''), status, created_at
		FROM users
		WHERE id = $1
	`
	var user User
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&user.ID,
		&user.Nickname,
		&user.Phone,
		&user.Account,
		&user.Wechat,
		&user.Status,
		&user.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrUserNotFound
	}
	return user, err
}
