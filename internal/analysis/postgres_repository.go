package analysis

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
)

type postgresDB interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
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

func (r *PostgresRepository) ListSessions(ctx context.Context, userID int64, limit int) ([]Session, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, user_id, mode, intent, status, questions, result, created_at, updated_at
		FROM analysis_sessions
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []Session
	for rows.Next() {
		session, err := scanSession(rows)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return sessions, nil
}

func (r *PostgresRepository) GetSession(ctx context.Context, userID, id int64) (Session, error) {
	session, err := scanSession(r.db.QueryRow(ctx, `
		SELECT id, user_id, mode, intent, status, questions, result, created_at, updated_at
		FROM analysis_sessions
		WHERE user_id = $1 AND id = $2
	`, userID, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return Session{}, ErrSessionNotFound
	}
	return session, err
}

type sessionScanner interface {
	Scan(dest ...any) error
}

func scanSession(scanner sessionScanner) (Session, error) {
	var session Session
	var questions []byte
	var result []byte
	if err := scanner.Scan(
		&session.ID,
		&session.UserID,
		&session.Mode,
		&session.Intent,
		&session.Status,
		&questions,
		&result,
		&session.CreatedAt,
		&session.UpdatedAt,
	); err != nil {
		return Session{}, err
	}
	if err := json.Unmarshal(questions, &session.Questions); err != nil {
		return Session{}, err
	}
	if err := json.Unmarshal(result, &session.Result); err != nil {
		return Session{}, err
	}
	return session, nil
}
