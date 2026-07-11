package tasks

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestPostgresRepositoryBatchOperationsAreAtomicIntegration(t *testing.T) {
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

	userIDs := make([]int64, 2)
	for index := range userIDs {
		if err := db.QueryRow(ctx, `
			INSERT INTO users (nickname, phone, agreement_accepted_at)
			VALUES ('批量任务测试', $1, NOW())
			RETURNING id
		`, fmt.Sprintf("%d%d", time.Now().UnixNano(), index)).Scan(&userIDs[index]); err != nil {
			t.Fatalf("insert user: %v", err)
		}
	}
	t.Cleanup(func() {
		_, _ = db.Exec(context.Background(), `DELETE FROM users WHERE id = ANY($1::bigint[])`, userIDs)
	})

	taskIDs := make([]int64, 3)
	for index, userID := range []int64{userIDs[0], userIDs[0], userIDs[1]} {
		if err := db.QueryRow(ctx, `
			INSERT INTO tasks (user_id, title, project, status, priority)
			VALUES ($1, $2, '批量验证', 'todo', 'medium')
			RETURNING id
		`, userID, fmt.Sprintf("批量任务 %d", index+1)).Scan(&taskIDs[index]); err != nil {
			t.Fatalf("insert task: %v", err)
		}
	}

	repository := NewPostgresRepository(db)
	updated, err := repository.BatchUpdateTaskStatus(ctx, userIDs[0], taskIDs[:2], StatusCompleted)
	if err != nil || updated != 2 {
		t.Fatalf("BatchUpdateTaskStatus() count/error = %d/%v", updated, err)
	}
	if _, err := repository.BatchUpdateTaskStatus(ctx, userIDs[0], []int64{taskIDs[0], taskIDs[2]}, StatusTodo); !errors.Is(err, ErrTaskNotFound) {
		t.Fatalf("mixed-owner update error = %v, want ErrTaskNotFound", err)
	}
	var status string
	if err := db.QueryRow(ctx, `SELECT status FROM tasks WHERE id = $1`, taskIDs[0]).Scan(&status); err != nil || status != StatusCompleted {
		t.Fatalf("task status/error = %q/%v", status, err)
	}

	if _, err := repository.BatchDeleteTasks(ctx, userIDs[0], []int64{taskIDs[0], taskIDs[2]}); !errors.Is(err, ErrTaskNotFound) {
		t.Fatalf("mixed-owner delete error = %v, want ErrTaskNotFound", err)
	}
	var ownedCount int
	if err := db.QueryRow(ctx, `SELECT COUNT(*) FROM tasks WHERE user_id = $1`, userIDs[0]).Scan(&ownedCount); err != nil || ownedCount != 2 {
		t.Fatalf("owned count/error = %d/%v", ownedCount, err)
	}
	deleted, err := repository.BatchDeleteTasks(ctx, userIDs[0], taskIDs[:2])
	if err != nil || deleted != 2 {
		t.Fatalf("BatchDeleteTasks() count/error = %d/%v", deleted, err)
	}
}
