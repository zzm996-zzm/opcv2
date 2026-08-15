package tasks

import (
	"context"
	"testing"
	"time"

	pgxmock "github.com/pashagolub/pgxmock/v4"
)

func TestPostgresRepositoryCreatesAndGetsTaskAIDraft(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 8, 15, 8, 0, 0, 0, time.UTC)
	draftTasks := []Task{{Title: "整理访谈名单", Project: "客户验证", Status: StatusTodo, Priority: PriorityHigh}}
	db.ExpectQuery("INSERT INTO task_ai_drafts").
		WithArgs(int64(42), "验证客户需求", "", pgxmock.AnyArg(), "", "", pgxmock.AnyArg(), TaskAIDraftStatusDraft, []byte(`[]`), now, pgxmock.AnyArg()).
		WillReturnRows(pgxmock.NewRows([]string{"id", "created_at", "updated_at", "adopted_at"}).AddRow(int64(77), now, now, nil))
	db.ExpectQuery("SELECT id, user_id, goal, source_type").
		WithArgs(int64(42), int64(77)).
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "user_id", "goal", "source_type", "source_id", "source_title", "source_url", "tasks", "status", "adopted_task_ids", "created_at", "updated_at", "adopted_at",
		}).AddRow(int64(77), int64(42), "验证客户需求", "", nil, "", "", []byte(`[{"title":"整理访谈名单","project":"客户验证","status":"todo","priority":"high","tools":null,"learning":"","progress":0,"version":0,"source_type":"","source_title":"","source_url":"","created_at":"0001-01-01T00:00:00Z","updated_at":"0001-01-01T00:00:00Z"}]`), TaskAIDraftStatusDraft, []byte(`[]`), now, now, nil))

	repository := NewPostgresRepository(db)
	created, err := repository.CreateTaskAIDraft(context.Background(), TaskAIDraft{
		UserID: 42, Goal: "验证客户需求", Tasks: draftTasks, Status: TaskAIDraftStatusDraft, CreatedAt: now,
	})
	if err != nil || created.ID != 77 || len(created.AdoptedTaskIDs) != 0 {
		t.Fatalf("CreateTaskAIDraft() draft/error = %+v/%v", created, err)
	}
	loaded, err := repository.GetTaskAIDraft(context.Background(), 42, 77)
	if err != nil || len(loaded.Tasks) != 1 || loaded.Tasks[0].Title != "整理访谈名单" {
		t.Fatalf("GetTaskAIDraft() draft/error = %+v/%v", loaded, err)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresRepositoryListsAndCountsTaskActivities(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 8, 15, 8, 30, 0, 0, time.UTC)
	db.ExpectQuery("SELECT activity.id, activity.task_id").
		WithArgs(int64(42), int64(99), 20, 0).
		WillReturnRows(pgxmock.NewRows([]string{"id", "task_id", "user_id", "action", "before_data", "after_data", "metadata", "created_at"}).
			AddRow(int64(1), int64(99), int64(42), "status_changed", []byte(`{"status":"todo"}`), []byte(`{"status":"in_progress"}`), []byte(`{}`), now))
	db.ExpectQuery("SELECT COUNT\\(\\*\\)").
		WithArgs(int64(42), int64(99)).
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(1))

	repository := NewPostgresRepository(db)
	activities, err := repository.ListTaskActivities(context.Background(), 42, 99, 20, 0)
	if err != nil || len(activities) != 1 || activities[0].BeforeData["status"] != StatusTodo || activities[0].AfterData["status"] != StatusInProgress {
		t.Fatalf("ListTaskActivities() activities/error = %+v/%v", activities, err)
	}
	total, err := repository.CountTaskActivities(context.Background(), 42, 99)
	if err != nil || total != 1 {
		t.Fatalf("CountTaskActivities() total/error = %d/%v", total, err)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestTaskOrderClauseUsesWhitelistedStableOrdering(t *testing.T) {
	tests := map[string]string{
		TaskSortCreated:  "created_at DESC, id DESC",
		TaskSortUpdated:  "updated_at DESC, id DESC",
		TaskSortDue:      "due_at ASC NULLS LAST, id DESC",
		TaskSortPriority: "CASE priority WHEN 'high' THEN 1 WHEN 'medium' THEN 2 ELSE 3 END ASC, created_at DESC, id DESC",
		TaskSortProgress: "progress DESC, updated_at DESC, id DESC",
		"unsafe SQL":     "created_at DESC, id DESC",
	}
	for input, expected := range tests {
		if actual := taskOrderClause(input); actual != expected {
			t.Fatalf("taskOrderClause(%q) = %q, want %q", input, actual, expected)
		}
	}
}
