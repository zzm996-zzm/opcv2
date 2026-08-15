package tasks

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestPostgresRepositoryDispatchesReminderIntegration(t *testing.T) {
	databaseURL := os.Getenv("OPCV2_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("OPCV2_TEST_DATABASE_URL is not set")
	}
	ctx := context.Background()
	db, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("pgxpool.New() error = %v", err)
	}
	defer db.Close()

	var userID int64
	err = db.QueryRow(ctx, `
		INSERT INTO users (nickname, phone, agreement_accepted_at)
		VALUES ('提醒集成测试', $1, NOW())
		RETURNING id
	`, fmt.Sprintf("%d", time.Now().UnixNano())).Scan(&userID)
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.Exec(context.Background(), `DELETE FROM users WHERE id = $1`, userID)
	})

	var taskID int64
	err = db.QueryRow(ctx, `
		INSERT INTO tasks (user_id, title, project, status, priority)
		VALUES ($1, '提交客户访谈报告', '客户验证', 'todo', 'high')
		RETURNING id
	`, userID).Scan(&taskID)
	if err != nil {
		t.Fatalf("insert task: %v", err)
	}

	now := time.Now().UTC().Truncate(time.Second)
	repository := NewPostgresRepository(db)
	_, err = repository.UpsertTaskReminder(ctx, TaskReminder{
		TaskID: taskID, UserID: userID, RemindAt: now.Add(-time.Minute), Recurrence: ReminderRecurrenceOnce, CreatedAt: now.Add(-2 * time.Minute),
	})
	if err != nil {
		t.Fatalf("UpsertTaskReminder() error = %v", err)
	}

	count, err := repository.DispatchDueTaskReminders(ctx, now, 100)
	if err != nil {
		t.Fatalf("DispatchDueTaskReminders() error = %v", err)
	}
	if count != 1 {
		t.Fatalf("count = %d, want 1", count)
	}

	var notificationType, sourceType, actionURL string
	var sourceID int64
	err = db.QueryRow(ctx, `
		SELECT type, source_type, source_id, action_url
		FROM notifications
		WHERE user_id = $1
	`, userID).Scan(&notificationType, &sourceType, &sourceID, &actionURL)
	if err != nil {
		t.Fatalf("read notification: %v", err)
	}
	if notificationType != "task" || sourceType != "task" || sourceID != taskID || actionURL != "/tasks?task_id="+fmt.Sprint(taskID) {
		t.Fatalf("notification = %q/%q/%d/%q", notificationType, sourceType, sourceID, actionURL)
	}

	count, err = repository.DispatchDueTaskReminders(ctx, now.Add(time.Minute), 100)
	if err != nil || count != 0 {
		t.Fatalf("second dispatch count/error = %d/%v, want 0/nil", count, err)
	}

	_, err = repository.UpsertTaskReminder(ctx, TaskReminder{
		TaskID: taskID, UserID: userID, RemindAt: now.Add(-49 * time.Hour), Recurrence: ReminderRecurrenceDaily, CreatedAt: now,
	})
	if err != nil {
		t.Fatalf("UpsertTaskReminder(daily) error = %v", err)
	}
	count, err = repository.DispatchDueTaskReminders(ctx, now, 100)
	if err != nil || count != 1 {
		t.Fatalf("daily dispatch count/error = %d/%v, want 1/nil", count, err)
	}
	reminder, err := repository.GetTaskReminder(ctx, userID, taskID)
	if err != nil || reminder == nil {
		t.Fatalf("GetTaskReminder() reminder/error = %+v/%v", reminder, err)
	}
	if reminder.Recurrence != ReminderRecurrenceDaily || reminder.SentAt == nil || !reminder.RemindAt.After(now) || reminder.RemindAt.After(now.Add(24*time.Hour)) {
		t.Fatalf("advanced daily reminder = %+v", reminder)
	}
	count, err = repository.DispatchDueTaskReminders(ctx, now, 100)
	if err != nil || count != 0 {
		t.Fatalf("repeated daily dispatch count/error = %d/%v, want 0/nil", count, err)
	}
	var notificationCount int
	if err := db.QueryRow(ctx, `SELECT COUNT(*) FROM notifications WHERE user_id = $1`, userID).Scan(&notificationCount); err != nil {
		t.Fatalf("count notifications: %v", err)
	}
	if notificationCount != 2 {
		t.Fatalf("notification count = %d, want 2", notificationCount)
	}
}
