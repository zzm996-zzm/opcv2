package analysis

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5"
)

type postgresDB interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type PostgresRepository struct {
	db postgresDB
}

func NewPostgresRepository(db postgresDB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) CreateSession(ctx context.Context, session Session) (Session, error) {
	questions, err := json.Marshal(session.Questions)
	if err != nil {
		return Session{}, err
	}
	result, err := json.Marshal(session.Result)
	if err != nil {
		return Session{}, err
	}
	err = r.db.QueryRow(ctx, `
		INSERT INTO analysis_sessions (user_id, mode, intent, status, questions, result, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`,
		session.UserID,
		session.Mode,
		session.Intent,
		session.Status,
		string(questions),
		result,
		session.CreatedAt,
	).Scan(&session.ID)
	return session, err
}
