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
		RETURNING id, nickname, phone, COALESCE(wechat, ''), status, created_at, (xmax = 0) AS created
	`
	var user User
	var created bool
	err := r.db.QueryRow(ctx, query, nickname, phone, agreementAcceptedAt).Scan(
		&user.ID,
		&user.Nickname,
		&user.Phone,
		&user.Wechat,
		&user.Status,
		&user.CreatedAt,
		&created,
	)
	return user, created, err
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
		SELECT id, nickname, phone, COALESCE(wechat, ''), status, created_at
		FROM users
		WHERE id = $1
	`
	var user User
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&user.ID,
		&user.Nickname,
		&user.Phone,
		&user.Wechat,
		&user.Status,
		&user.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrUserNotFound
	}
	return user, err
}
