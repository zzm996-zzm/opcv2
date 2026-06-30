package tasks

import (
	"context"
	"regexp"
	"testing"
	"time"

	pgxmock "github.com/pashagolub/pgxmock/v4"
)

func TestPostgresRepositoryCreatesTask(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 6, 30, 11, 0, 0, 0, time.UTC)
	due := now.Add(2 * time.Hour)
	db.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO tasks (user_id, title, project, status, priority, due_at, tools, learning, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $9)
		RETURNING id
	`)).
		WithArgs(
			int64(42),
			"整理客户名单",
			"AI线索开发",
			StatusTodo,
			PriorityHigh,
			&due,
			[]byte(`["CRM"]`),
			"线索评分",
			now,
		).
		WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(int64(99)))

	repository := NewPostgresRepository(db)
	task, err := repository.CreateTask(context.Background(), Task{
		UserID:    42,
		Title:     "整理客户名单",
		Project:   "AI线索开发",
		Status:    StatusTodo,
		Priority:  PriorityHigh,
		DueAt:     &due,
		Tools:     []string{"CRM"},
		Learning:  "线索评分",
		CreatedAt: now,
	})
	if err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}
	if task.ID != 99 {
		t.Fatalf("task.ID = %d, want 99", task.ID)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresRepositoryListsTasksForUser(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 6, 30, 11, 0, 0, 0, time.UTC)
	db.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, user_id, title, project, status, priority, due_at, tools, learning, created_at, updated_at
		FROM tasks
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`)).
		WithArgs(int64(42), 20).
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "user_id", "title", "project", "status", "priority", "due_at", "tools", "learning", "created_at", "updated_at",
		}).AddRow(
			int64(99),
			int64(42),
			"整理客户名单",
			"AI线索开发",
			StatusTodo,
			PriorityHigh,
			nil,
			[]byte(`["CRM"]`),
			"线索评分",
			now,
			now,
		))

	repository := NewPostgresRepository(db)
	tasks, err := repository.ListTasks(context.Background(), 42, 20)
	if err != nil {
		t.Fatalf("ListTasks() error = %v", err)
	}
	if len(tasks) != 1 || tasks[0].ID != 99 || tasks[0].Tools[0] != "CRM" {
		t.Fatalf("tasks = %+v", tasks)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresRepositoryUpdatesOwnedTask(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 6, 30, 11, 0, 0, 0, time.UTC)
	status := StatusCompleted
	db.ExpectQuery(regexp.QuoteMeta(`
		UPDATE tasks
		SET title = COALESCE($1, title),
		    project = COALESCE($2, project),
		    status = COALESCE($3, status),
		    priority = COALESCE($4, priority),
		    due_at = COALESCE($5, due_at),
		    tools = COALESCE($6, tools),
		    learning = COALESCE($7, learning),
		    updated_at = NOW()
		WHERE user_id = $8 AND id = $9
		RETURNING id, user_id, title, project, status, priority, due_at, tools, learning, created_at, updated_at
	`)).
		WithArgs(
			nil,
			nil,
			&status,
			nil,
			nil,
			nil,
			nil,
			int64(42),
			int64(99),
		).
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "user_id", "title", "project", "status", "priority", "due_at", "tools", "learning", "created_at", "updated_at",
		}).AddRow(
			int64(99),
			int64(42),
			"整理客户名单",
			"AI线索开发",
			StatusCompleted,
			PriorityHigh,
			nil,
			[]byte(`["CRM"]`),
			"线索评分",
			now,
			now,
		))

	repository := NewPostgresRepository(db)
	task, err := repository.UpdateTask(context.Background(), 42, 99, TaskUpdate{Status: &status})
	if err != nil {
		t.Fatalf("UpdateTask() error = %v", err)
	}
	if task.Status != StatusCompleted {
		t.Fatalf("task = %+v", task)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
