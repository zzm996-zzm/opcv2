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
		SELECT reminder.id, reminder.task_id, reminder.user_id, reminder.remind_at,
		       reminder.recurrence, reminder.sent_at, reminder.created_at, reminder.updated_at
		FROM task_reminders AS reminder
		JOIN tasks AS task ON task.id = reminder.task_id AND task.deleted_at IS NULL
		WHERE reminder.user_id = $1 AND reminder.task_id = $2
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
		INSERT INTO task_reminders (task_id, user_id, remind_at, recurrence, created_at, updated_at)
		SELECT t.id, t.user_id, $3, $4, $5, $5
		FROM tasks t
		WHERE t.id = $1 AND t.user_id = $2 AND t.deleted_at IS NULL
		ON CONFLICT (task_id) DO UPDATE
		SET remind_at = EXCLUDED.remind_at,
		    recurrence = EXCLUDED.recurrence,
		    sent_at = NULL,
		    updated_at = EXCLUDED.updated_at
		RETURNING id, task_id, user_id, remind_at, recurrence, sent_at, created_at, updated_at
	`, reminder.TaskID, reminder.UserID, reminder.RemindAt, reminder.Recurrence, reminder.CreatedAt))
	if errors.Is(err, pgx.ErrNoRows) {
		return TaskReminder{}, ErrTaskNotFound
	}
	return item, err
}

func (r *PostgresRepository) DeleteTaskReminder(ctx context.Context, userID, taskID int64) error {
	tag, err := r.db.Exec(ctx, `
		DELETE FROM task_reminders AS reminder
		USING tasks AS task
		WHERE reminder.user_id = $1 AND reminder.task_id = $2
		  AND task.id = reminder.task_id AND task.deleted_at IS NULL
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
			    remind_at = CASE reminder.recurrence
			        WHEN 'daily' THEN reminder.remind_at + (FLOOR(EXTRACT(EPOCH FROM ($1 - reminder.remind_at)) / 86400) + 1) * INTERVAL '1 day'
			        WHEN 'weekly' THEN reminder.remind_at + (FLOOR(EXTRACT(EPOCH FROM ($1 - reminder.remind_at)) / 604800) + 1) * INTERVAL '7 days'
			        ELSE reminder.remind_at
			    END,
			    updated_at = $1
			WHERE reminder.id IN (
				SELECT candidate.id
					FROM task_reminders AS candidate
					JOIN tasks AS task ON task.id = candidate.task_id
					WHERE (candidate.recurrence <> 'once' OR candidate.sent_at IS NULL)
					  AND candidate.remind_at <= $1
					  AND task.deleted_at IS NULL
					  AND task.status NOT IN ('completed', 'cancelled')
				ORDER BY candidate.remind_at, candidate.id
				LIMIT $2
				FOR UPDATE SKIP LOCKED
			)
			RETURNING reminder.id AS reminder_id, reminder.user_id, reminder.task_id, reminder.remind_at AS next_remind_at
		), deliveries AS (
			SELECT due.reminder_id,
			       due.user_id,
			       due.task_id,
			       due.next_remind_at,
			       task.title,
			       task.status,
			       COALESCE(preference.notifications_enabled, TRUE) AS notifications_enabled
			FROM due
			JOIN tasks AS task ON task.id = due.task_id
			LEFT JOIN user_preferences AS preference ON preference.user_id = due.user_id
		), inserted AS (
			INSERT INTO notifications (
				user_id, type, title, summary, body, source_type, source_id, action_label, action_url, dedupe_key, created_at
			)
			SELECT user_id,
			       'task',
			       '任务提醒：' || title,
			       '任务已到提醒时间',
			       '任务「' || title || '」已到设定的提醒时间。',
			       'task',
			       task_id,
			       '查看任务',
			       '/tasks?task_id=' || task_id::TEXT,
			       'task.reminder.' || reminder_id::TEXT || '.' || EXTRACT(EPOCH FROM next_remind_at)::BIGINT::TEXT,
			       $1
			FROM deliveries
			WHERE status <> 'completed'
			  AND notifications_enabled
			ON CONFLICT (user_id, dedupe_key) WHERE dedupe_key <> '' DO NOTHING
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
		&reminder.Recurrence,
		&sentAt,
		&reminder.CreatedAt,
		&reminder.UpdatedAt,
	)
	if sentAt.Valid {
		reminder.SentAt = &sentAt.Time
	}
	return reminder, err
}
