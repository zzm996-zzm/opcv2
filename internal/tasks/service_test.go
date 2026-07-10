package tasks

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeRepository struct {
	created  Task
	updated  Task
	deleted  Task
	task     Task
	tasks    []Task
	projects []string
	filters  ListFilters
	err      error
}

func (r *fakeRepository) CreateTask(_ context.Context, task Task) (Task, error) {
	r.created = task
	task.ID = 99
	task.UpdatedAt = task.CreatedAt
	r.task = task
	return task, r.err
}

func (r *fakeRepository) ListTasks(_ context.Context, userID int64, filters ListFilters) ([]Task, error) {
	r.filters = filters
	if r.err != nil {
		return nil, r.err
	}
	rows := make([]Task, 0, len(r.tasks))
	for _, task := range r.tasks {
		if task.UserID == userID {
			rows = append(rows, task)
		}
	}
	start := min(len(rows), filters.Offset)
	end := min(len(rows), start+filters.Limit)
	return rows[start:end], nil
}

func (r *fakeRepository) CountTasks(_ context.Context, userID int64, _ ListFilters) (int, error) {
	if r.err != nil {
		return 0, r.err
	}
	total := 0
	for _, task := range r.tasks {
		if task.UserID == userID {
			total++
		}
	}
	return total, nil
}

func (r *fakeRepository) ListTaskProjects(_ context.Context, _ int64) ([]string, error) {
	return r.projects, r.err
}

func (r *fakeRepository) TaskStats(_ context.Context, userID int64, now time.Time) (Stats, error) {
	if r.err != nil {
		return Stats{}, r.err
	}
	stats := Stats{}
	for _, task := range r.tasks {
		if task.UserID != userID {
			continue
		}
		stats.Total++
		switch task.Status {
		case StatusTodo:
			stats.Todo++
		case StatusInProgress:
			stats.InProgress++
		case StatusCompleted:
			stats.Completed++
		case StatusReminder:
			stats.Reminder++
		}
		if task.DueAt != nil && task.DueAt.Before(now) && task.Status != StatusCompleted {
			stats.Overdue++
		}
	}
	return stats, nil
}

func (r *fakeRepository) GetTask(_ context.Context, userID, id int64) (Task, error) {
	if r.err != nil {
		return Task{}, r.err
	}
	if r.task.UserID != userID || r.task.ID != id {
		return Task{}, ErrTaskNotFound
	}
	return r.task, nil
}

func (r *fakeRepository) UpdateTask(_ context.Context, userID, id int64, update TaskUpdate) (Task, error) {
	if r.err != nil {
		return Task{}, r.err
	}
	if r.task.UserID != userID || r.task.ID != id {
		return Task{}, ErrTaskNotFound
	}
	if update.Status != nil {
		r.task.Status = *update.Status
	}
	if update.Priority != nil {
		r.task.Priority = *update.Priority
	}
	if update.Title != nil {
		r.task.Title = *update.Title
	}
	r.updated = r.task
	return r.task, nil
}

func (r *fakeRepository) DeleteTask(_ context.Context, userID, id int64) error {
	if r.err != nil {
		return r.err
	}
	if r.task.UserID != userID || r.task.ID != id {
		return ErrTaskNotFound
	}
	r.deleted = r.task
	return nil
}

func TestServiceCreatesTaskWithDefaults(t *testing.T) {
	now := time.Date(2026, 6, 30, 11, 0, 0, 0, time.UTC)
	repository := &fakeRepository{}
	service := NewService(repository)
	service.now = func() time.Time { return now }

	task, err := service.CreateTask(context.Background(), CreateInput{
		UserID:   42,
		Title:    "整理首批客户名单",
		Project:  "AI线索开发",
		Priority: PriorityHigh,
		Tools:    []string{"CRM", "表格助手"},
		Learning: "线索评分",
	})

	if err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}
	if task.ID != 99 || task.Status != StatusTodo {
		t.Fatalf("task = %+v", task)
	}
	if repository.created.UserID != 42 || repository.created.Title == "" || repository.created.CreatedAt != now {
		t.Fatalf("created = %+v", repository.created)
	}
}

func TestServiceListsOnlyUserTasks(t *testing.T) {
	repository := &fakeRepository{tasks: []Task{
		{ID: 1, UserID: 42, Title: "我的任务"},
		{ID: 2, UserID: 7, Title: "别人的任务"},
	}}
	service := NewService(repository)

	page, err := service.ListTaskPage(context.Background(), 42, ListFilters{Limit: 20})

	if err != nil {
		t.Fatalf("ListTasks() error = %v", err)
	}
	if len(page.Tasks) != 1 || page.Tasks[0].Title != "我的任务" || page.Total != 1 {
		t.Fatalf("page = %+v", page)
	}
}

func TestServiceNormalizesTaskFilters(t *testing.T) {
	repository := &fakeRepository{tasks: []Task{{ID: 1, UserID: 42, Title: "我的任务"}}}
	service := NewService(repository)

	_, err := service.ListTaskPage(context.Background(), 42, ListFilters{
		Status:   " in_progress ",
		Project:  " 商业沙盘 ",
		Priority: " high ",
		Query:    " 接口 ",
		Limit:    500,
		Offset:   -10,
	})

	if err != nil {
		t.Fatalf("ListTasks() error = %v", err)
	}
	if repository.filters.Status != StatusInProgress || repository.filters.Project != "商业沙盘" || repository.filters.Priority != PriorityHigh || repository.filters.Query != "接口" || repository.filters.Limit != 100 || repository.filters.Offset != 0 {
		t.Fatalf("filters = %+v", repository.filters)
	}
}

func TestServiceListsTaskProjects(t *testing.T) {
	repository := &fakeRepository{projects: []string{"AI线索开发", "商业沙盘"}}
	service := NewService(repository)

	projects, err := service.ListTaskProjects(context.Background(), 42)

	if err != nil {
		t.Fatalf("ListTaskProjects() error = %v", err)
	}
	if len(projects) != 2 || projects[1] != "商业沙盘" {
		t.Fatalf("projects = %+v", projects)
	}
}

func TestServiceReturnsTaskStats(t *testing.T) {
	now := time.Date(2026, 7, 2, 10, 0, 0, 0, time.UTC)
	due := now.Add(-time.Hour)
	repository := &fakeRepository{tasks: []Task{
		{ID: 1, UserID: 42, Status: StatusTodo, DueAt: &due},
		{ID: 2, UserID: 42, Status: StatusInProgress},
		{ID: 3, UserID: 7, Status: StatusCompleted},
	}}
	service := NewService(repository)
	service.now = func() time.Time { return now }

	stats, err := service.TaskStats(context.Background(), 42)

	if err != nil {
		t.Fatalf("TaskStats() error = %v", err)
	}
	if stats.Total != 2 || stats.Todo != 1 || stats.InProgress != 1 || stats.Overdue != 1 {
		t.Fatalf("stats = %+v", stats)
	}
}

func TestServiceUpdatesOwnedTask(t *testing.T) {
	repository := &fakeRepository{task: Task{ID: 99, UserID: 42, Title: "整理客户", Status: StatusTodo}}
	service := NewService(repository)
	status := StatusCompleted

	task, err := service.UpdateTask(context.Background(), 42, 99, TaskUpdate{Status: &status})

	if err != nil {
		t.Fatalf("UpdateTask() error = %v", err)
	}
	if task.Status != StatusCompleted || repository.updated.Status != StatusCompleted {
		t.Fatalf("task = %+v", task)
	}
}

func TestServiceRejectsOtherUsersTask(t *testing.T) {
	repository := &fakeRepository{task: Task{ID: 99, UserID: 7, Title: "别人的任务"}}
	service := NewService(repository)
	status := StatusCompleted

	_, err := service.UpdateTask(context.Background(), 42, 99, TaskUpdate{Status: &status})

	if !errors.Is(err, ErrTaskNotFound) {
		t.Fatalf("err = %v, want ErrTaskNotFound", err)
	}
}

func TestServiceDeletesOwnedTask(t *testing.T) {
	repository := &fakeRepository{task: Task{ID: 99, UserID: 42, Title: "整理客户"}}
	service := NewService(repository)

	err := service.DeleteTask(context.Background(), 42, 99)

	if err != nil {
		t.Fatalf("DeleteTask() error = %v", err)
	}
	if repository.deleted.ID != 99 {
		t.Fatalf("deleted = %+v", repository.deleted)
	}
}

func TestServiceRejectsDeletingOtherUsersTask(t *testing.T) {
	repository := &fakeRepository{task: Task{ID: 99, UserID: 7, Title: "别人的任务"}}
	service := NewService(repository)

	err := service.DeleteTask(context.Background(), 42, 99)

	if !errors.Is(err, ErrTaskNotFound) {
		t.Fatalf("err = %v, want ErrTaskNotFound", err)
	}
}
