package tasks

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

func (r *PostgresRepository) GetTaskReminder(ctx context.Context, userID, taskID int64) (*TaskReminder, error) {
	reminder, err := scanTaskReminder(r.db.QueryRow(ctx, `
		SELECT id, task_id, user_id, remind_at, sent_at, created_at, updated_at
		FROM task_reminders
		WHERE user_id = $1 AND task_id = $2
	`, userID, taskID))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &reminder, nil
}

func (r *PostgresRepository) UpsertTaskReminder(ctx context.Context, reminder TaskReminder) (TaskReminder, error) {
	item, err := scanTaskReminder(r.db.QueryRow(ctx, `
		INSERT INTO task_reminders (task_id, user_id, remind_at, created_at, updated_at)
		SELECT t.id, t.user_id, $3, $4, $4
		FROM tasks t
		WHERE t.id = $1 AND t.user_id = $2
		ON CONFLICT (task_id) DO UPDATE
		SET remind_at = EXCLUDED.remind_at,
		    sent_at = NULL,
		    updated_at = EXCLUDED.updated_at
		RETURNING id, task_id, user_id, remind_at, sent_at, created_at, updated_at
	`, reminder.TaskID, reminder.UserID, reminder.RemindAt, reminder.CreatedAt))
	if errors.Is(err, pgx.ErrNoRows) {
		return TaskReminder{}, ErrTaskNotFound
	}
	return item, err
}

func (r *PostgresRepository) DeleteTaskReminder(ctx context.Context, userID, taskID int64) error {
	tag, err := r.db.Exec(ctx, `
		DELETE FROM task_reminders
		WHERE user_id = $1 AND task_id = $2
	`, userID, taskID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrReminderNotFound
	}
	return nil
}

func (r *PostgresRepository) DispatchDueTaskReminders(ctx context.Context, now time.Time, limit int) (int, error) {
	var count int
	err := r.db.QueryRow(ctx, `
		WITH due AS (
			UPDATE task_reminders AS reminder
			SET sent_at = $1,
			    updated_at = $1
			WHERE reminder.id IN (
				SELECT candidate.id
				FROM task_reminders AS candidate
				WHERE candidate.sent_at IS NULL
				  AND candidate.remind_at <= $1
				ORDER BY candidate.remind_at, candidate.id
				LIMIT $2
				FOR UPDATE SKIP LOCKED
			)
			RETURNING reminder.user_id, reminder.task_id
		), deliveries AS (
			SELECT due.user_id,
			       due.task_id,
			       task.title,
			       task.status,
			       COALESCE(preference.notifications_enabled, TRUE) AS notifications_enabled
			FROM due
			JOIN tasks AS task ON task.id = due.task_id
			LEFT JOIN user_preferences AS preference ON preference.user_id = due.user_id
		), inserted AS (
			INSERT INTO notifications (
				user_id, type, title, summary, body, source_type, source_id, action_label, action_url, created_at
			)
			SELECT user_id,
			       'task',
			       '任务提醒：' || title,
			       '任务已到提醒时间',
			       '任务「' || title || '」已到设定的提醒时间。',
			       'task',
			       task_id,
			       '查看任务',
			       '/tasks',
			       $1
			FROM deliveries
			WHERE status <> 'completed'
			  AND notifications_enabled
			RETURNING id
		)
		SELECT COUNT(*)::INT FROM inserted
	`, now, limit).Scan(&count)
	return count, err
}

func scanTaskReminder(scanner taskScanner) (TaskReminder, error) {
	var reminder TaskReminder
	var sentAt sql.NullTime
	err := scanner.Scan(
		&reminder.ID,
		&reminder.TaskID,
		&reminder.UserID,
		&reminder.RemindAt,
		&sentAt,
		&reminder.CreatedAt,
		&reminder.UpdatedAt,
	)
	if sentAt.Valid {
		reminder.SentAt = &sentAt.Time
	}
	return reminder, err
}
