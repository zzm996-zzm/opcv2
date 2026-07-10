package tasks

import (
	"context"
	"errors"
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
		INSERT INTO tasks (user_id, title, description, assignee, project, status, priority, due_at, tools, learning, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $11)
		RETURNING id
	`)).
		WithArgs(
			int64(42),
			"整理客户名单",
			"完成首批客户画像并安排访谈",
			"李明",
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
		UserID:      42,
		Title:       "整理客户名单",
		Description: "完成首批客户画像并安排访谈",
		Assignee:    "李明",
		Project:     "AI线索开发",
		Status:      StatusTodo,
		Priority:    PriorityHigh,
		DueAt:       &due,
		Tools:       []string{"CRM"},
		Learning:    "线索评分",
		CreatedAt:   now,
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
		SELECT id, user_id, title, description, assignee, project, status, priority, due_at, tools, learning, created_at, updated_at
		FROM tasks
		WHERE user_id = $1
		  AND ($2 = '' OR status = $2)
		  AND ($3 = '' OR project = $3)
		  AND ($4 = '' OR priority = $4)
		  AND ($5 = '' OR title ILIKE '%' || $5 || '%' OR description ILIKE '%' || $5 || '%' OR assignee ILIKE '%' || $5 || '%' OR project ILIKE '%' || $5 || '%' OR learning ILIKE '%' || $5 || '%')
		ORDER BY created_at DESC
		LIMIT $6
		OFFSET $7
	`)).
		WithArgs(int64(42), StatusTodo, "AI线索开发", PriorityHigh, "客户", 20, 10).
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "user_id", "title", "description", "assignee", "project", "status", "priority", "due_at", "tools", "learning", "created_at", "updated_at",
		}).AddRow(
			int64(99),
			int64(42),
			"整理客户名单",
			"完成首批客户画像并安排访谈",
			"李明",
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
	tasks, err := repository.ListTasks(context.Background(), 42, ListFilters{
		Status:   StatusTodo,
		Project:  "AI线索开发",
		Priority: PriorityHigh,
		Query:    "客户",
		Limit:    20,
		Offset:   10,
	})
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

func TestPostgresRepositoryCountsFilteredTasks(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	db.ExpectQuery("(?s)SELECT COUNT\\(\\*\\).*assignee ILIKE").
		WithArgs(int64(42), StatusTodo, "AI线索开发", PriorityHigh, "客户").
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(21))

	repository := NewPostgresRepository(db)
	total, err := repository.CountTasks(context.Background(), 42, ListFilters{
		Status: StatusTodo, Project: "AI线索开发", Priority: PriorityHigh, Query: "客户",
	})
	if err != nil {
		t.Fatalf("CountTasks() error = %v", err)
	}
	if total != 21 {
		t.Fatalf("total = %d, want 21", total)
	}
}

func TestPostgresRepositoryListsTaskProjects(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	db.ExpectQuery("SELECT DISTINCT project").
		WithArgs(int64(42)).
		WillReturnRows(pgxmock.NewRows([]string{"project"}).AddRow("AI线索开发").AddRow("商业沙盘"))

	repository := NewPostgresRepository(db)
	projects, err := repository.ListTaskProjects(context.Background(), 42)
	if err != nil {
		t.Fatalf("ListTaskProjects() error = %v", err)
	}
	if len(projects) != 2 || projects[0] != "AI线索开发" {
		t.Fatalf("projects = %+v", projects)
	}
}

func TestPostgresRepositoryReturnsTaskStats(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 7, 2, 10, 0, 0, 0, time.UTC)
	db.ExpectQuery(regexp.QuoteMeta(`
		SELECT
			COUNT(*)::INT,
			COUNT(*) FILTER (WHERE status = 'todo')::INT,
			COUNT(*) FILTER (WHERE status = 'in_progress')::INT,
			COUNT(*) FILTER (WHERE status = 'completed')::INT,
			COUNT(*) FILTER (WHERE status = 'reminder')::INT,
			COUNT(*) FILTER (WHERE due_at IS NOT NULL AND due_at < $2 AND status <> 'completed')::INT
		FROM tasks
		WHERE user_id = $1
	`)).
		WithArgs(int64(42), now).
		WillReturnRows(pgxmock.NewRows([]string{"total", "todo", "in_progress", "completed", "reminder", "overdue"}).
			AddRow(3, 1, 1, 1, 0, 1))

	repository := NewPostgresRepository(db)
	stats, err := repository.TaskStats(context.Background(), 42, now)
	if err != nil {
		t.Fatalf("TaskStats() error = %v", err)
	}
	if stats.Total != 3 || stats.InProgress != 1 || stats.Overdue != 1 {
		t.Fatalf("stats = %+v", stats)
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
		    description = COALESCE($2, description),
		    assignee = COALESCE($3, assignee),
		    project = COALESCE($4, project),
		    status = COALESCE($5, status),
		    priority = COALESCE($6, priority),
		    due_at = CASE WHEN $8 THEN NULL ELSE COALESCE($7, due_at) END,
		    tools = COALESCE($9, tools),
		    learning = COALESCE($10, learning),
		    updated_at = NOW()
		WHERE user_id = $11 AND id = $12
		RETURNING id, user_id, title, description, assignee, project, status, priority, due_at, tools, learning, created_at, updated_at
	`)).
		WithArgs(
			nil,
			nil,
			nil,
			nil,
			&status,
			nil,
			nil,
			false,
			nil,
			nil,
			int64(42),
			int64(99),
		).
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "user_id", "title", "description", "assignee", "project", "status", "priority", "due_at", "tools", "learning", "created_at", "updated_at",
		}).AddRow(
			int64(99),
			int64(42),
			"整理客户名单",
			"完成首批客户画像并安排访谈",
			"李明",
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

func TestPostgresRepositoryClearsTaskDueAt(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 6, 30, 11, 0, 0, 0, time.UTC)
	db.ExpectQuery("UPDATE tasks").
		WithArgs(
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
			true,
			nil,
			nil,
			int64(42),
			int64(99),
		).
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "user_id", "title", "description", "assignee", "project", "status", "priority", "due_at", "tools", "learning", "created_at", "updated_at",
		}).AddRow(
			int64(99),
			int64(42),
			"整理客户名单",
			"",
			"",
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
	task, err := repository.UpdateTask(context.Background(), 42, 99, TaskUpdate{ClearDueAt: true})
	if err != nil {
		t.Fatalf("UpdateTask() error = %v", err)
	}
	if task.DueAt != nil {
		t.Fatalf("task.DueAt = %v, want nil", task.DueAt)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresRepositoryDeletesOwnedTask(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	db.ExpectExec(regexp.QuoteMeta(`
		DELETE FROM tasks
		WHERE user_id = $1 AND id = $2
	`)).
		WithArgs(int64(42), int64(99)).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	repository := NewPostgresRepository(db)
	if err := repository.DeleteTask(context.Background(), 42, 99); err != nil {
		t.Fatalf("DeleteTask() error = %v", err)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresRepositoryReturnsNotFoundWhenDeleteMisses(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	db.ExpectExec("DELETE FROM tasks").
		WithArgs(int64(42), int64(99)).
		WillReturnResult(pgxmock.NewResult("DELETE", 0))

	repository := NewPostgresRepository(db)
	err = repository.DeleteTask(context.Background(), 42, 99)
	if !errors.Is(err, ErrTaskNotFound) {
		t.Fatalf("err = %v, want ErrTaskNotFound", err)
	}
}
