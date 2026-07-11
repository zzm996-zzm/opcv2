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
	sourceID := int64(501)
	db.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO tasks (user_id, title, description, assignee, project, status, priority, tags, due_at, tools, learning, source_type, source_id, source_title, source_url, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $16)
		RETURNING id, created_at, updated_at
	`)).
		WithArgs(
			int64(42),
			"整理客户名单",
			"完成首批客户画像并安排访谈",
			"李明",
			"AI线索开发",
			StatusTodo,
			PriorityHigh,
			[]byte(`["用户研究","访谈"]`),
			&due,
			[]byte(`["CRM"]`),
			"线索评分",
			SourceCRMCustomer,
			&sourceID,
			"重点客户 A",
			"/crm/customers/501",
			now,
		).
		WillReturnRows(pgxmock.NewRows([]string{"id", "created_at", "updated_at"}).AddRow(int64(99), now, now))

	repository := NewPostgresRepository(db)
	task, err := repository.CreateTask(context.Background(), Task{
		UserID:      42,
		Title:       "整理客户名单",
		Description: "完成首批客户画像并安排访谈",
		Assignee:    "李明",
		Project:     "AI线索开发",
		Status:      StatusTodo,
		Priority:    PriorityHigh,
		Tags:        []string{"用户研究", "访谈"},
		DueAt:       &due,
		Tools:       []string{"CRM"},
		Learning:    "线索评分",
		SourceType:  SourceCRMCustomer,
		SourceID:    &sourceID,
		SourceTitle: "重点客户 A",
		SourceURL:   "/crm/customers/501",
		CreatedAt:   now,
	})
	if err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}
	if task.ID != 99 || task.SourceID == nil || *task.SourceID != sourceID || !task.CreatedAt.Equal(now) || !task.UpdatedAt.Equal(now) {
		t.Fatalf("task = %+v", task)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresRepositoryCreatesTasksInTransaction(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 7, 11, 12, 0, 0, 0, time.UTC)
	db.ExpectBegin()
	sourceID := int64(601)
	tasks := []Task{
		{UserID: 42, Title: "整理访谈名单", Project: "客户验证", Status: StatusTodo, Priority: PriorityMedium, Tags: []string{}, Tools: []string{}, SourceType: SourceAnalysisSession, SourceID: &sourceID, SourceTitle: "客户访谈分析", SourceURL: "/analysis/sessions/601", CreatedAt: now},
		{UserID: 42, Title: "完成访谈复盘", Project: "客户验证", Status: StatusTodo, Priority: PriorityMedium, Tags: []string{}, Tools: []string{}, CreatedAt: now},
	}
	for index, task := range tasks {
		db.ExpectQuery("INSERT INTO tasks").
			WithArgs(
				int64(42), task.Title, "", "", "客户验证", StatusTodo, PriorityMedium,
				[]byte(`[]`), (*time.Time)(nil), []byte(`[]`), "",
				task.SourceType, task.SourceID, task.SourceTitle, task.SourceURL, now,
			).
			WillReturnRows(pgxmock.NewRows([]string{"id", "created_at", "updated_at"}).AddRow(int64(101+index), now, now))
	}
	db.ExpectCommit()

	repository := NewPostgresRepository(db)
	created, err := repository.CreateTasks(context.Background(), tasks)

	if err != nil || len(created) != 2 || created[0].ID != 101 || created[1].ID != 102 || created[0].SourceID == nil || created[1].SourceID != nil {
		t.Fatalf("created/error = %+v/%v", created, err)
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
	sourceID := int64(501)
	db.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, user_id, title, description, assignee, project, status, priority, tags, due_at, tools, learning, source_type, source_id, source_title, source_url, created_at, updated_at
		FROM tasks
		WHERE user_id = $1
		  AND ($2 = '' OR status = $2)
		  AND ($3 = '' OR project = $3)
		  AND ($4 = '' OR priority = $4)
		  AND ($5 = '' OR tags ? $5)
		  AND ($6 = '' OR title ILIKE '%' || $6 || '%' OR description ILIKE '%' || $6 || '%' OR assignee ILIKE '%' || $6 || '%' OR project ILIKE '%' || $6 || '%' OR tags::TEXT ILIKE '%' || $6 || '%' OR learning ILIKE '%' || $6 || '%')
		ORDER BY created_at DESC
		LIMIT $7
		OFFSET $8
	`)).
		WithArgs(int64(42), StatusTodo, "AI线索开发", PriorityHigh, "用户研究", "客户", 20, 10).
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "user_id", "title", "description", "assignee", "project", "status", "priority", "tags", "due_at", "tools", "learning", "source_type", "source_id", "source_title", "source_url", "created_at", "updated_at",
		}).AddRow(
			int64(99),
			int64(42),
			"整理客户名单",
			"完成首批客户画像并安排访谈",
			"李明",
			"AI线索开发",
			StatusTodo,
			PriorityHigh,
			[]byte(`["用户研究","访谈"]`),
			nil,
			[]byte(`["CRM"]`),
			"线索评分",
			SourceCRMCustomer,
			&sourceID,
			"重点客户 A",
			"/crm/customers/501",
			now,
			now,
		))

	repository := NewPostgresRepository(db)
	tasks, err := repository.ListTasks(context.Background(), 42, ListFilters{
		Status:   StatusTodo,
		Project:  "AI线索开发",
		Priority: PriorityHigh,
		Tag:      "用户研究",
		Query:    "客户",
		Limit:    20,
		Offset:   10,
	})
	if err != nil {
		t.Fatalf("ListTasks() error = %v", err)
	}
	if len(tasks) != 1 || tasks[0].ID != 99 || tasks[0].Tools[0] != "CRM" || tasks[0].SourceID == nil || *tasks[0].SourceID != sourceID || tasks[0].SourceType != SourceCRMCustomer {
		t.Fatalf("tasks = %+v", tasks)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresRepositoryGetsTaskWithNoSource(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 6, 30, 11, 0, 0, 0, time.UTC)
	db.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, user_id, title, description, assignee, project, status, priority, tags, due_at, tools, learning, source_type, source_id, source_title, source_url, created_at, updated_at
		FROM tasks
		WHERE user_id = $1 AND id = $2
	`)).
		WithArgs(int64(42), int64(99)).
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "user_id", "title", "description", "assignee", "project", "status", "priority", "tags", "due_at", "tools", "learning", "source_type", "source_id", "source_title", "source_url", "created_at", "updated_at",
		}).AddRow(
			int64(99), int64(42), "完成访谈复盘", "", "", "客户验证", StatusTodo, PriorityMedium,
			[]byte(`[]`), nil, []byte(`[]`), "", "", nil, "", "", now, now,
		))

	repository := NewPostgresRepository(db)
	task, err := repository.GetTask(context.Background(), 42, 99)
	if err != nil {
		t.Fatalf("GetTask() error = %v", err)
	}
	if task.ID != 99 || task.SourceType != "" || task.SourceID != nil || task.SourceTitle != "" || task.SourceURL != "" {
		t.Fatalf("task = %+v", task)
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

	db.ExpectQuery("(?s)SELECT COUNT\\(\\*\\).*tags \\? \\$5.*tags::TEXT ILIKE").
		WithArgs(int64(42), StatusTodo, "AI线索开发", PriorityHigh, "用户研究", "客户").
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(21))

	repository := NewPostgresRepository(db)
	total, err := repository.CountTasks(context.Background(), 42, ListFilters{
		Status: StatusTodo, Project: "AI线索开发", Priority: PriorityHigh, Tag: "用户研究", Query: "客户",
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

func TestPostgresRepositoryListsTaskTags(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	db.ExpectQuery("SELECT DISTINCT tag.value").
		WithArgs(int64(42)).
		WillReturnRows(pgxmock.NewRows([]string{"value"}).AddRow("用户研究").AddRow("访谈"))

	repository := NewPostgresRepository(db)
	tags, err := repository.ListTaskTags(context.Background(), 42)
	if err != nil {
		t.Fatalf("ListTaskTags() error = %v", err)
	}
	if len(tags) != 2 || tags[0] != "用户研究" {
		t.Fatalf("tags = %+v", tags)
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
	sourceID := int64(501)
	status := StatusCompleted
	db.ExpectQuery(regexp.QuoteMeta(`
		UPDATE tasks
		SET title = COALESCE($1, title),
		    description = COALESCE($2, description),
		    assignee = COALESCE($3, assignee),
		    project = COALESCE($4, project),
		    status = COALESCE($5, status),
		    priority = COALESCE($6, priority),
		    tags = COALESCE($7, tags),
		    due_at = CASE WHEN $9 THEN NULL ELSE COALESCE($8, due_at) END,
		    tools = COALESCE($10, tools),
		    learning = COALESCE($11, learning),
		    updated_at = NOW()
		WHERE user_id = $12 AND id = $13
		RETURNING id, user_id, title, description, assignee, project, status, priority, tags, due_at, tools, learning, source_type, source_id, source_title, source_url, created_at, updated_at
	`)).
		WithArgs(
			nil,
			nil,
			nil,
			nil,
			&status,
			nil,
			nil,
			nil,
			false,
			nil,
			nil,
			int64(42),
			int64(99),
		).
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "user_id", "title", "description", "assignee", "project", "status", "priority", "tags", "due_at", "tools", "learning", "source_type", "source_id", "source_title", "source_url", "created_at", "updated_at",
		}).AddRow(
			int64(99),
			int64(42),
			"整理客户名单",
			"完成首批客户画像并安排访谈",
			"李明",
			"AI线索开发",
			StatusCompleted,
			PriorityHigh,
			[]byte(`["用户研究"]`),
			nil,
			[]byte(`["CRM"]`),
			"线索评分",
			SourceCRMCustomer,
			&sourceID,
			"重点客户 A",
			"/crm/customers/501",
			now,
			now,
		))

	repository := NewPostgresRepository(db)
	task, err := repository.UpdateTask(context.Background(), 42, 99, TaskUpdate{Status: &status})
	if err != nil {
		t.Fatalf("UpdateTask() error = %v", err)
	}
	if task.Status != StatusCompleted || task.SourceID == nil || *task.SourceID != 501 || task.SourceType != SourceCRMCustomer {
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
			nil,
			true,
			nil,
			nil,
			int64(42),
			int64(99),
		).
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "user_id", "title", "description", "assignee", "project", "status", "priority", "tags", "due_at", "tools", "learning", "source_type", "source_id", "source_title", "source_url", "created_at", "updated_at",
		}).AddRow(
			int64(99),
			int64(42),
			"整理客户名单",
			"",
			"",
			"AI线索开发",
			StatusTodo,
			PriorityHigh,
			[]byte(`[]`),
			nil,
			[]byte(`["CRM"]`),
			"线索评分",
			"",
			nil,
			"",
			"",
			now,
			now,
		))

	repository := NewPostgresRepository(db)
	task, err := repository.UpdateTask(context.Background(), 42, 99, TaskUpdate{ClearDueAt: true})
	if err != nil {
		t.Fatalf("UpdateTask() error = %v", err)
	}
	if task.DueAt != nil || task.SourceType != "" || task.SourceID != nil || task.SourceTitle != "" || task.SourceURL != "" {
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

func TestPostgresRepositoryListsOwnedTaskSubtasks(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 7, 10, 10, 0, 0, 0, time.UTC)
	db.ExpectQuery("SELECT s.id, s.task_id, s.user_id").
		WithArgs(int64(42), int64(99)).
		WillReturnRows(pgxmock.NewRows([]string{"id", "task_id", "user_id", "title", "assignee", "due_at", "completed", "created_at", "updated_at"}).
			AddRow(int64(7), int64(99), int64(42), "整理访谈提纲", "李明", nil, false, now, now))

	repository := NewPostgresRepository(db)
	items, err := repository.ListSubtasks(context.Background(), 42, 99)

	if err != nil || len(items) != 1 || items[0].ID != 7 {
		t.Fatalf("items/error = %+v/%v", items, err)
	}
}

func TestPostgresRepositoryCreatesSubtaskOnlyForOwnedTask(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 7, 10, 10, 0, 0, 0, time.UTC)
	db.ExpectQuery("INSERT INTO task_subtasks").
		WithArgs(int64(99), int64(42), "整理访谈提纲", "李明", nil, now).
		WillReturnRows(pgxmock.NewRows([]string{"id", "task_id", "user_id", "title", "assignee", "due_at", "completed", "created_at", "updated_at"}).
			AddRow(int64(7), int64(99), int64(42), "整理访谈提纲", "李明", nil, false, now, now))

	repository := NewPostgresRepository(db)
	item, err := repository.CreateSubtask(context.Background(), Subtask{TaskID: 99, UserID: 42, Title: "整理访谈提纲", Assignee: "李明", CreatedAt: now})

	if err != nil || item.ID != 7 {
		t.Fatalf("item/error = %+v/%v", item, err)
	}
}

func TestPostgresRepositoryReturnsTaskNotFoundWhenCreatingSubtaskForOtherUser(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	db.ExpectQuery("INSERT INTO task_subtasks").
		WithArgs(int64(99), int64(42), "整理访谈提纲", "", nil, time.Time{}).
		WillReturnRows(pgxmock.NewRows([]string{"id", "task_id", "user_id", "title", "assignee", "due_at", "completed", "created_at", "updated_at"}))

	repository := NewPostgresRepository(db)
	_, err = repository.CreateSubtask(context.Background(), Subtask{TaskID: 99, UserID: 42, Title: "整理访谈提纲"})

	if !errors.Is(err, ErrTaskNotFound) {
		t.Fatalf("err = %v, want ErrTaskNotFound", err)
	}
}

func TestPostgresRepositoryUpdatesOwnedSubtask(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 7, 10, 10, 0, 0, 0, time.UTC)
	completed := true
	db.ExpectQuery("UPDATE task_subtasks").
		WithArgs(nil, nil, nil, false, &completed, int64(42), int64(99), int64(7)).
		WillReturnRows(pgxmock.NewRows([]string{"id", "task_id", "user_id", "title", "assignee", "due_at", "completed", "created_at", "updated_at"}).
			AddRow(int64(7), int64(99), int64(42), "整理访谈提纲", "李明", nil, true, now, now))

	repository := NewPostgresRepository(db)
	item, err := repository.UpdateSubtask(context.Background(), 42, 99, 7, SubtaskUpdate{Completed: &completed})

	if err != nil || !item.Completed {
		t.Fatalf("item/error = %+v/%v", item, err)
	}
}

func TestPostgresRepositoryReturnsNotFoundForOtherUsersSubtask(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	db.ExpectExec("DELETE FROM task_subtasks").
		WithArgs(int64(42), int64(99), int64(7)).
		WillReturnResult(pgxmock.NewResult("DELETE", 0))

	repository := NewPostgresRepository(db)
	err = repository.DeleteSubtask(context.Background(), 42, 99, 7)

	if !errors.Is(err, ErrSubtaskNotFound) {
		t.Fatalf("err = %v, want ErrSubtaskNotFound", err)
	}
}

func TestPostgresRepositoryGetsTaskReminder(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 7, 10, 10, 0, 0, 0, time.UTC)
	remindAt := now.Add(2 * time.Hour)
	db.ExpectQuery("SELECT id, task_id, user_id, remind_at, recurrence, sent_at, created_at, updated_at").
		WithArgs(int64(42), int64(99)).
		WillReturnRows(pgxmock.NewRows([]string{"id", "task_id", "user_id", "remind_at", "recurrence", "sent_at", "created_at", "updated_at"}).
			AddRow(int64(8), int64(99), int64(42), remindAt, ReminderRecurrenceOnce, nil, now, now))

	repository := NewPostgresRepository(db)
	reminder, err := repository.GetTaskReminder(context.Background(), 42, 99)

	if err != nil || reminder == nil || reminder.ID != 8 || !reminder.RemindAt.Equal(remindAt) {
		t.Fatalf("reminder/error = %+v/%v", reminder, err)
	}
}

func TestPostgresRepositoryReturnsNilWhenTaskReminderIsMissing(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	db.ExpectQuery("SELECT id, task_id, user_id, remind_at, recurrence, sent_at, created_at, updated_at").
		WithArgs(int64(42), int64(99)).
		WillReturnRows(pgxmock.NewRows([]string{"id", "task_id", "user_id", "remind_at", "recurrence", "sent_at", "created_at", "updated_at"}))

	repository := NewPostgresRepository(db)
	reminder, err := repository.GetTaskReminder(context.Background(), 42, 99)

	if err != nil || reminder != nil {
		t.Fatalf("reminder/error = %+v/%v", reminder, err)
	}
}

func TestPostgresRepositoryUpsertsReminderOnlyForOwnedTask(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 7, 10, 10, 0, 0, 0, time.UTC)
	remindAt := now.Add(2 * time.Hour)
	db.ExpectQuery("INSERT INTO task_reminders").
		WithArgs(int64(99), int64(42), remindAt, ReminderRecurrenceWeekly, now).
		WillReturnRows(pgxmock.NewRows([]string{"id", "task_id", "user_id", "remind_at", "recurrence", "sent_at", "created_at", "updated_at"}).
			AddRow(int64(8), int64(99), int64(42), remindAt, ReminderRecurrenceWeekly, nil, now, now))

	repository := NewPostgresRepository(db)
	reminder, err := repository.UpsertTaskReminder(context.Background(), TaskReminder{TaskID: 99, UserID: 42, RemindAt: remindAt, Recurrence: ReminderRecurrenceWeekly, CreatedAt: now})

	if err != nil || reminder.ID != 8 {
		t.Fatalf("reminder/error = %+v/%v", reminder, err)
	}
}

func TestPostgresRepositoryDeletesOwnedTaskReminder(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	db.ExpectExec("DELETE FROM task_reminders").
		WithArgs(int64(42), int64(99)).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	repository := NewPostgresRepository(db)
	if err := repository.DeleteTaskReminder(context.Background(), 42, 99); err != nil {
		t.Fatalf("DeleteTaskReminder() error = %v", err)
	}
}

func TestPostgresRepositoryDispatchesDueTaskReminders(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 7, 10, 10, 0, 0, 0, time.UTC)
	db.ExpectQuery("(?s)WITH due AS.*recurrence.*INTERVAL '1 day'.*INTERVAL '7 days'.*status <> 'completed'.*notifications_enabled").
		WithArgs(now, 100).
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(2))

	repository := NewPostgresRepository(db)
	count, err := repository.DispatchDueTaskReminders(context.Background(), now, 100)

	if err != nil || count != 2 {
		t.Fatalf("count/error = %d/%v", count, err)
	}
}
