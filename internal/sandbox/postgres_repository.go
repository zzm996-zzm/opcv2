package sandbox

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type postgresDB interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
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
	intake, err := marshalIntake(session.Intake)
	if err != nil {
		return Session{}, err
	}
	settings, err := marshalSettings(session.Settings)
	if err != nil {
		return Session{}, err
	}
	report, err := marshalReport(session.Report)
	if err != nil {
		return Session{}, err
	}
	err = r.db.QueryRow(ctx, `
		INSERT INTO sandbox_sessions (user_id, goal, target_users, product, roles, status, progress_percent, current_step, error_message, run_attempt, intake, settings, report, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $14)
		RETURNING id
	`,
		session.UserID,
		session.Goal,
		session.TargetUsers,
		session.Product,
		roles,
		session.Status,
		session.ProgressPercent,
		session.CurrentStep,
		session.ErrorMessage,
		session.RunAttempt,
		intake,
		settings,
		report,
		session.CreatedAt,
	).Scan(&session.ID)
	return session, err
}

func (r *PostgresRepository) UpdateSessionDraft(ctx context.Context, userID, id int64, update DraftUpdate) (Session, error) {
	if update.Goal == nil || update.TargetUsers == nil || update.Product == nil || update.Roles == nil {
		return Session{}, ErrInvalidSession
	}
	roles, err := json.Marshal(*update.Roles)
	if err != nil {
		return Session{}, err
	}
	if update.Settings != nil {
		settings, err := marshalSettings(*update.Settings)
		if err != nil {
			return Session{}, err
		}
		session, err := scanSession(r.db.QueryRow(ctx, `
			UPDATE sandbox_sessions
			SET goal = $1, target_users = $2, product = $3, roles = $4, settings = $5, updated_at = NOW()
			WHERE user_id = $6 AND id = $7 AND status = $8
			RETURNING id, user_id, goal, target_users, product, roles, status, progress_percent, current_step, error_message, run_attempt, intake, settings, report, created_at, updated_at
		`, *update.Goal, *update.TargetUsers, *update.Product, roles, settings, userID, id, StatusDraft))
		if errors.Is(err, pgx.ErrNoRows) {
			return Session{}, ErrSessionNotFound
		}
		return session, err
	}
	session, err := scanSession(r.db.QueryRow(ctx, `
		UPDATE sandbox_sessions
		SET goal = $1, target_users = $2, product = $3, roles = $4, updated_at = NOW()
		WHERE user_id = $5 AND id = $6 AND status = $7
		RETURNING id, user_id, goal, target_users, product, roles, status, progress_percent, current_step, error_message, run_attempt, intake, settings, report, created_at, updated_at
	`, *update.Goal, *update.TargetUsers, *update.Product, roles, userID, id, StatusDraft))
	if errors.Is(err, pgx.ErrNoRows) {
		return Session{}, ErrSessionNotFound
	}
	return session, err
}

func (r *PostgresRepository) UpdateSessionIntake(ctx context.Context, userID, id int64, intakeValue Intake) (Session, error) {
	intake, err := marshalIntake(intakeValue)
	if err != nil {
		return Session{}, err
	}
	session, err := scanSession(r.db.QueryRow(ctx, `
		UPDATE sandbox_sessions
		SET intake = $1, current_step = $2, updated_at = NOW()
		WHERE user_id = $3 AND id = $4 AND status = $5
		RETURNING id, user_id, goal, target_users, product, roles, status, progress_percent, current_step, error_message, run_attempt, intake, settings, report, created_at, updated_at
	`, intake, intakeValue.Status, userID, id, StatusDraft))
	if errors.Is(err, pgx.ErrNoRows) {
		return Session{}, ErrSessionNotFound
	}
	return session, err
}

func (r *PostgresRepository) UpdateSessionSettings(ctx context.Context, userID, id int64, settingsValue RunSettings) (Session, error) {
	settings, err := marshalSettings(settingsValue)
	if err != nil {
		return Session{}, err
	}
	session, err := scanSession(r.db.QueryRow(ctx, `
		UPDATE sandbox_sessions
		SET settings = $1, updated_at = NOW()
		WHERE user_id = $2 AND id = $3 AND status = $4
		RETURNING id, user_id, goal, target_users, product, roles, status, progress_percent, current_step, error_message, run_attempt, intake, settings, report, created_at, updated_at
	`, settings, userID, id, StatusDraft))
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

func (r *PostgresRepository) PrepareSessionRun(ctx context.Context, userID, id int64) (Session, error) {
	session, err := scanSession(r.db.QueryRow(ctx, `
		UPDATE sandbox_sessions
		SET status = $1, progress_percent = 0, current_step = $1, error_message = '',
		    run_attempt = run_attempt + 1, updated_at = NOW()
		WHERE user_id = $2 AND id = $3 AND status IN ($4, $5, $6)
		RETURNING id, user_id, goal, target_users, product, roles, status, progress_percent, current_step, error_message, run_attempt, intake, settings, report, created_at, updated_at
	`, StatusQueued, userID, id, StatusDraft, StatusFailed, StatusCanceled))
	if errors.Is(err, pgx.ErrNoRows) {
		return Session{}, ErrInvalidSession
	}
	return session, err
}

func (r *PostgresRepository) UpdateSessionProgress(ctx context.Context, userID, id int64, attempt int, status string, progress int, step, errorMessage string) (Session, error) {
	session, err := scanSession(r.db.QueryRow(ctx, `
		UPDATE sandbox_sessions
		SET status = $1, progress_percent = $2, current_step = $3, error_message = $4, updated_at = NOW()
		WHERE user_id = $5 AND id = $6 AND run_attempt = $7
		RETURNING id, user_id, goal, target_users, product, roles, status, progress_percent, current_step, error_message, run_attempt, intake, settings, report, created_at, updated_at
	`, status, progress, step, errorMessage, userID, id, attempt))
	if errors.Is(err, pgx.ErrNoRows) {
		return Session{}, ErrStaleRun
	}
	return session, err
}

func (r *PostgresRepository) CancelSession(ctx context.Context, userID, id int64) (Session, error) {
	session, err := scanSession(r.db.QueryRow(ctx, `
		UPDATE sandbox_sessions
		SET status = $1, progress_percent = 100, current_step = $1, error_message = '', updated_at = NOW()
		WHERE user_id = $2 AND id = $3 AND status IN ($4, $5)
		RETURNING id, user_id, goal, target_users, product, roles, status, progress_percent, current_step, error_message, run_attempt, intake, settings, report, created_at, updated_at
	`, StatusCanceled, userID, id, StatusQueued, StatusRunning))
	if errors.Is(err, pgx.ErrNoRows) {
		return Session{}, ErrInvalidSession
	}
	return session, err
}

func (r *PostgresRepository) UpdateSessionResult(ctx context.Context, userID, id int64, attempt int, result Report) (Session, error) {
	report, err := marshalReport(result)
	if err != nil {
		return Session{}, err
	}
	session, err := scanSession(r.db.QueryRow(ctx, `
		UPDATE sandbox_sessions
		SET status = $1, progress_percent = 100, current_step = $1, error_message = '', report = $2, updated_at = NOW()
		WHERE user_id = $3 AND id = $4 AND run_attempt = $5 AND status = $6
		RETURNING id, user_id, goal, target_users, product, roles, status, progress_percent, current_step, error_message, run_attempt, intake, settings, report, created_at, updated_at
	`, StatusCompleted, report, userID, id, attempt, StatusRunning))
	if errors.Is(err, pgx.ErrNoRows) {
		return Session{}, ErrStaleRun
	}
	return session, err
}

func marshalReport(report Report) ([]byte, error) {
	if report.Score == 0 &&
		report.Summary == "" &&
		report.Basis == "" &&
		report.Disclaimer == "" &&
		len(report.Assumptions) == 0 &&
		len(report.EvidenceSources) == 0 &&
		len(report.Metrics) == 0 &&
		len(report.RoleSummaries) == 0 &&
		len(report.Risks) == 0 &&
		len(report.NextActions) == 0 &&
		report.ReportVersion == "" &&
		report.ConsumerProbability == 0 &&
		report.RiskLevel == "" &&
		report.RecommendationGrade == "" &&
		len(report.CoreConclusions) == 0 &&
		len(report.OpportunityAnalysis) == 0 &&
		len(report.RiskAnalysis) == 0 &&
		len(report.ActionPlan) == 0 &&
		len(report.GrowthPath) == 0 &&
		len(report.ValidationMetrics) == 0 &&
		len(report.Timeline) == 0 {
		return []byte(`{}`), nil
	}
	return json.Marshal(report)
}

func marshalIntake(intake Intake) ([]byte, error) {
	if intake.Status == "" && intake.InitialIdea == "" && len(intake.RecognizedFields) == 0 && len(intake.Questions) == 0 {
		return []byte(`{}`), nil
	}
	return json.Marshal(intake)
}

func marshalSettings(settings RunSettings) ([]byte, error) {
	if settings.Depth == "" && settings.OutputStyle == "" && !settings.GenerateOutline && len(settings.Variables) == 0 {
		return json.Marshal(DefaultRunSettings())
	}
	if settings.Variables == nil {
		settings.Variables = map[string]string{}
	}
	return json.Marshal(settings)
}

func (r *PostgresRepository) ListSessions(ctx context.Context, userID int64, limit int) ([]Session, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, user_id, goal, target_users, product, roles, status, progress_percent, current_step, error_message, run_attempt, intake, settings, report, created_at, updated_at
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
		SELECT id, user_id, goal, target_users, product, roles, status, progress_percent, current_step, error_message, run_attempt, intake, settings, report, created_at, updated_at
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
	var intake []byte
	var settings []byte
	var report []byte
	if err := scanner.Scan(
		&session.ID,
		&session.UserID,
		&session.Goal,
		&session.TargetUsers,
		&session.Product,
		&roles,
		&session.Status,
		&session.ProgressPercent,
		&session.CurrentStep,
		&session.ErrorMessage,
		&session.RunAttempt,
		&intake,
		&settings,
		&report,
		&session.CreatedAt,
		&session.UpdatedAt,
	); err != nil {
		return Session{}, err
	}
	if len(roles) > 0 && string(roles) != "null" {
		if err := json.Unmarshal(roles, &session.Roles); err != nil {
			return Session{}, err
		}
	}
	if len(intake) > 0 && string(intake) != "null" {
		if err := json.Unmarshal(intake, &session.Intake); err != nil {
			return Session{}, err
		}
	}
	if session.Intake.Status == "" {
		session.Intake = DefaultReadyIntake()
	}
	if session.Intake.RecognizedFields == nil {
		session.Intake.RecognizedFields = []RecognizedField{}
	}
	if session.Intake.Questions == nil {
		session.Intake.Questions = []IntakeQuestion{}
	}
	if len(settings) > 0 && string(settings) != "null" {
		if err := json.Unmarshal(settings, &session.Settings); err != nil {
			return Session{}, err
		}
	}
	if session.Settings.Depth == "" {
		session.Settings = DefaultRunSettings()
	}
	if session.Settings.Variables == nil {
		session.Settings.Variables = map[string]string{}
	}
	if err := json.Unmarshal(report, &session.Report); err != nil {
		return Session{}, err
	}
	return session, nil
}
