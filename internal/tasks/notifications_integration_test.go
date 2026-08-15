package tasks

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestPostgresRepositoryDispatchesTaskLifecycleAndDeadlineNotificationsIntegration(t *testing.T) {
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
		VALUES ('任务通知集成测试', $1, NOW())
		RETURNING id
	`, fmt.Sprintf("%d", time.Now().UnixNano())).Scan(&userID)
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.Exec(context.Background(), `DELETE FROM users WHERE id = $1`, userID)
	})

	now := time.Now().UTC().Truncate(time.Second)
	var lifecycleTaskID int64
	err = db.QueryRow(ctx, `
		INSERT INTO tasks (user_id, title, project, status, priority)
		VALUES ($1, '确认交付范围', '任务通知测试', 'todo', 'high')
		RETURNING id
	`, userID).Scan(&lifecycleTaskID)
	if err != nil {
		t.Fatalf("insert lifecycle task: %v", err)
	}
	if _, err := db.Exec(ctx, `UPDATE tasks SET status = 'in_progress', updated_at = NOW() WHERE id = $1`, lifecycleTaskID); err != nil {
		t.Fatalf("update lifecycle task: %v", err)
	}
	for title, dueAt := range map[string]time.Time{
		"准备临期材料": now.Add(24 * time.Hour),
		"处理逾期事项": now.Add(-49 * time.Hour),
	} {
		if _, err := db.Exec(ctx, `
			INSERT INTO tasks (user_id, title, project, status, priority, due_at)
			VALUES ($1, $2, '任务通知测试', 'todo', 'medium', $3)
		`, userID, title, dueAt); err != nil {
			t.Fatalf("insert deadline task %q: %v", title, err)
		}
	}

	repository := NewPostgresRepository(db)
	count, err := repository.DispatchTaskNotifications(ctx, now, 100)
	if err != nil {
		t.Fatalf("DispatchTaskNotifications() error = %v", err)
	}
	if count < 3 {
		t.Fatalf("notification count = %d, want at least 3", count)
	}
	var userNotificationCount int
	if err := db.QueryRow(ctx, `
		SELECT COUNT(*) FROM notifications
		WHERE user_id = $1 AND type = 'task' AND dedupe_key <> ''
	`, userID).Scan(&userNotificationCount); err != nil {
		t.Fatalf("count user notifications: %v", err)
	}
	if userNotificationCount != 3 {
		t.Fatalf("user notification count = %d, want 3", userNotificationCount)
	}

	secondCount, err := repository.DispatchTaskNotifications(ctx, now.Add(time.Second), 100)
	if err != nil {
		t.Fatalf("second DispatchTaskNotifications() error = %v", err)
	}
	if secondCount != 0 {
		t.Fatalf("second notification count = %d, want 0", secondCount)
	}
}
