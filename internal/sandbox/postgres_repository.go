package sandbox

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
	roles, err := json.Marshal(session.Roles)
	if err != nil {
		return Session{}, err
	}
	report, err := marshalReport(session.Report)
	if err != nil {
		return Session{}, err
	}
	err = r.db.QueryRow(ctx, `
		INSERT INTO sandbox_sessions (user_id, goal, target_users, product, roles, status, report, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $8)
		RETURNING id
	`,
		session.UserID,
		session.Goal,
		session.TargetUsers,
		session.Product,
		roles,
		session.Status,
		report,
		session.CreatedAt,
	).Scan(&session.ID)
	return session, err
}

func (r *PostgresRepository) UpdateSessionDraft(ctx context.Context, userID, id int64, update DraftUpdate) (Session, error) {
	roles, err := json.Marshal(*update.Roles)
	if err != nil {
		return Session{}, err
	}
	session, err := scanSession(r.db.QueryRow(ctx, `
		UPDATE sandbox_sessions
		SET goal = $1, target_users = $2, product = $3, roles = $4, updated_at = NOW()
		WHERE user_id = $5 AND id = $6 AND status = $7
		RETURNING id, user_id, goal, target_users, product, roles, status, report, created_at, updated_at
	`, *update.Goal, *update.TargetUsers, *update.Product, roles, userID, id, StatusDraft))
	if errors.Is(err, pgx.ErrNoRows) {
		return Session{}, ErrSessionNotFound
	}
	return session, err
}

func (r *PostgresRepository) CreateMessage(ctx context.Context, message Message) (Message, error) {
	err := r.db.QueryRow(ctx, `
		INSERT INTO sandbox_messages (session_id, user_id, role, question, answer, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`, message.SessionID, message.UserID, message.Role, message.Question, message.Answer, message.CreatedAt).Scan(&message.ID)
	return message, err
}

func (r *PostgresRepository) ListMessages(ctx context.Context, userID, sessionID int64) ([]Message, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, session_id, user_id, role, question, answer, created_at
		FROM sandbox_messages
		WHERE user_id = $1 AND session_id = $2
		ORDER BY created_at ASC, id ASC
	`, userID, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	messages := make([]Message, 0)
	for rows.Next() {
		var message Message
		if err := rows.Scan(&message.ID, &message.SessionID, &message.UserID, &message.Role, &message.Question, &message.Answer, &message.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, message)
	}
	return messages, rows.Err()
}

func (r *PostgresRepository) UpdateSessionResult(ctx context.Context, userID, id int64, result Report) (Session, error) {
	report, err := marshalReport(result)
	if err != nil {
		return Session{}, err
	}
	session, err := scanSession(r.db.QueryRow(ctx, `
		UPDATE sandbox_sessions
		SET status = $1, report = $2, updated_at = NOW()
		WHERE user_id = $3 AND id = $4
		RETURNING id, user_id, goal, target_users, product, roles, status, report, created_at, updated_at
	`, StatusCompleted, report, userID, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return Session{}, ErrSessionNotFound
	}
	return session, err
}

func marshalReport(report Report) ([]byte, error) {
	if report.Score == 0 &&
		report.Summary == "" &&
		len(report.Metrics) == 0 &&
		len(report.RoleSummaries) == 0 &&
		len(report.Risks) == 0 &&
		len(report.NextActions) == 0 {
		return []byte(`{}`), nil
	}
	return json.Marshal(report)
}

func (r *PostgresRepository) ListSessions(ctx context.Context, userID int64, limit int) ([]Session, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, user_id, goal, target_users, product, roles, status, report, created_at, updated_at
		FROM sandbox_sessions
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
		SELECT id, user_id, goal, target_users, product, roles, status, report, created_at, updated_at
		FROM sandbox_sessions
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
	var roles []byte
	var report []byte
	if err := scanner.Scan(
		&session.ID,
		&session.UserID,
		&session.Goal,
		&session.TargetUsers,
		&session.Product,
		&roles,
		&session.Status,
		&report,
		&session.CreatedAt,
		&session.UpdatedAt,
	); err != nil {
		return Session{}, err
	}
	if err := json.Unmarshal(roles, &session.Roles); err != nil {
		return Session{}, err
	}
	if err := json.Unmarshal(report, &session.Report); err != nil {
		return Session{}, err
	}
	return session, nil
}
