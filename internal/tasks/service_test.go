package tasks

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/zzm/opcv2/internal/membership"
)

type fakeRepository struct {
	created      Task
	createdTasks []Task
	updated      Task
	deleted      Task
	task         Task
	tasks        []Task
	projects     []string
	tags         []string
	filters      ListFilters
	err          error
	batchUserID  int64
	batchIDs     []int64
	batchStatus  string
	batchCount   int
}

func (r *fakeRepository) BatchUpdateTaskStatus(_ context.Context, userID int64, ids []int64, status string) (int, error) {
	r.batchUserID, r.batchIDs, r.batchStatus = userID, append([]int64(nil), ids...), status
	return r.batchCount, r.err
}

func (r *fakeRepository) BatchDeleteTasks(_ context.Context, userID int64, ids []int64) (int, error) {
	r.batchUserID, r.batchIDs = userID, append([]int64(nil), ids...)
	return r.batchCount, r.err
}

func (r *fakeRepository) CreateTasks(_ context.Context, tasks []Task) ([]Task, error) {
	r.createdTasks = append([]Task(nil), tasks...)
	if r.err != nil {
		return nil, r.err
	}
	created := make([]Task, len(tasks))
	for index, task := range tasks {
		task.ID = int64(100 + index)
		task.UpdatedAt = task.CreatedAt
		created[index] = task
	}
	return created, nil
}

func (r *fakeRepository) CreateTask(_ context.Context, task Task) (Task, error) {
	r.created = task
	task.ID = 99
	task.UpdatedAt = task.CreatedAt
	r.task = task
	return task, r.err
}

func TestServiceCreatesTaskBatchWithOwnedSandboxSource(t *testing.T) {
	repository := &fakeRepository{}
	service := NewService(repository)
	runID := int64(99)
	created, err := service.CreateTasks(context.Background(), BatchCreateInput{UserID: 42, Tasks: []CreateInput{{Title: "访谈十家门店", Description: "验证付费意愿", Project: "AI 运营", Priority: PriorityHigh, SourceType: SourceSandboxSession, SourceID: &runID, SourceTitle: "AI 运营沙盘", SourceURL: "/sandbox-runs/99/report"}}})
	if err != nil || len(created) != 1 {
		t.Fatalf("created/error=%+v/%v", created, err)
	}
	if repository.createdTasks[0].UserID != 42 || repository.createdTasks[0].SourceType != SourceSandboxSession || repository.createdTasks[0].SourceID == nil || *repository.createdTasks[0].SourceID != 99 {
		t.Fatalf("task=%+v", repository.createdTasks[0])
	}
}

func TestServiceRejectsInvalidTaskBatch(t *testing.T) {
	service := NewService(&fakeRepository{})
	for _, input := range []BatchCreateInput{{UserID: 42}, {UserID: 42, Tasks: make([]CreateInput, 21)}, {UserID: 42, Tasks: []CreateInput{{Title: "x", Project: "p", Priority: PriorityHigh, SourceType: SourceSandboxSession}}}} {
		if _, err := service.CreateTasks(context.Background(), input); !errors.Is(err, ErrInvalidTaskBatch) {
			t.Fatalf("input/error=%+v/%v", input, err)
		}
	}
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

func (r *fakeRepository) ListTaskTags(_ context.Context, _ int64) ([]string, error) {
	return r.tags, r.err
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
	if update.Assignee != nil {
		r.task.Assignee = *update.Assignee
	}
	if update.Tags != nil {
		r.task.Tags = *update.Tags
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

func TestServiceBatchUpdatesTaskStatusWithUniqueIDs(t *testing.T) {
	repository := &fakeRepository{batchCount: 2}
	service := NewService(repository)

	count, err := service.BatchUpdateTaskStatus(context.Background(), 42, []int64{9, 7, 9}, StatusCompleted)
	if err != nil {
		t.Fatalf("BatchUpdateTaskStatus() error = %v", err)
	}
	if count != 2 || repository.batchUserID != 42 || repository.batchStatus != StatusCompleted {
		t.Fatalf("count/user/status = %d/%d/%q", count, repository.batchUserID, repository.batchStatus)
	}
	if len(repository.batchIDs) != 2 || repository.batchIDs[0] != 9 || repository.batchIDs[1] != 7 {
		t.Fatalf("batchIDs = %v, want [9 7]", repository.batchIDs)
	}
}

func TestServiceBatchDeletesTasksWithUniqueIDs(t *testing.T) {
	repository := &fakeRepository{batchCount: 2}
	service := NewService(repository)

	count, err := service.BatchDeleteTasks(context.Background(), 42, []int64{7, 9, 7})
	if err != nil || count != 2 {
		t.Fatalf("count/error = %d/%v", count, err)
	}
	if len(repository.batchIDs) != 2 || repository.batchIDs[0] != 7 || repository.batchIDs[1] != 9 {
		t.Fatalf("batchIDs = %v, want [7 9]", repository.batchIDs)
	}
}

func TestServiceRejectsInvalidBatchInputs(t *testing.T) {
	service := NewService(&fakeRepository{})
	tooMany := make([]int64, 101)
	for index := range tooMany {
		tooMany[index] = int64(index + 1)
	}

	tests := []struct {
		name   string
		ids    []int64
		status string
	}{
		{name: "empty", ids: nil, status: StatusTodo},
		{name: "non-positive", ids: []int64{1, 0}, status: StatusTodo},
		{name: "too many", ids: tooMany, status: StatusTodo},
		{name: "invalid status", ids: []int64{1}, status: "done"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := service.BatchUpdateTaskStatus(context.Background(), 42, test.ids, test.status)
			if !errors.Is(err, ErrInvalidTaskBatch) {
				t.Fatalf("err = %v, want ErrInvalidTaskBatch", err)
			}
		})
	}
}

func TestServiceCreatesTaskWithDefaults(t *testing.T) {
	now := time.Date(2026, 6, 30, 11, 0, 0, 0, time.UTC)
	repository := &fakeRepository{}
	service := NewService(repository)
	service.now = func() time.Time { return now }

	task, err := service.CreateTask(context.Background(), CreateInput{
		UserID:      42,
		Title:       "整理首批客户名单",
		Description: " 明确客户范围和访谈目标 ",
		Assignee:    " 李明 ",
		Project:     "AI线索开发",
		Priority:    PriorityHigh,
		Tags:        []string{" 用户研究 ", "访谈", "用户研究", " "},
		Tools:       []string{"CRM", "表格助手"},
		Learning:    "线索评分",
	})

	if err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}
	if task.ID != 99 || task.Status != StatusTodo {
		t.Fatalf("task = %+v", task)
	}
	if repository.created.UserID != 42 || repository.created.Title == "" || repository.created.Description != "明确客户范围和访谈目标" || repository.created.Assignee != "李明" || repository.created.CreatedAt != now {
		t.Fatalf("created = %+v", repository.created)
	}
	if len(repository.created.Tags) != 2 || repository.created.Tags[0] != "用户研究" || repository.created.Tags[1] != "访谈" {
		t.Fatalf("created.Tags = %+v", repository.created.Tags)
	}
}

func TestServiceCreatesTaskWithNormalizedSource(t *testing.T) {
	now := time.Date(2026, 7, 12, 9, 0, 0, 0, time.UTC)
	sourceID := int64(11)
	repository := &fakeRepository{}
	service := NewService(repository)
	service.now = func() time.Time { return now }

	_, err := service.CreateTask(context.Background(), CreateInput{
		UserID: 42, Title: "反击竞品更新", Project: "竞品动态监测", Priority: PriorityHigh,
		SourceType: " competitor_scan ", SourceID: &sourceID, SourceTitle: " 销售自动化提速 ", SourceURL: " /competitor-data ",
	})

	if err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}
	if repository.created.SourceType != SourceCompetitorScan || repository.created.SourceID == nil || *repository.created.SourceID != 11 || repository.created.SourceTitle != "销售自动化提速" || repository.created.SourceURL != "/competitor-data" {
		t.Fatalf("source = %+v", repository.created)
	}
}

func TestServiceAcceptsGrowthModelSource(t *testing.T) {
	sourceID := int64(99)
	repository := &fakeRepository{}
	service := NewService(repository)

	_, err := service.CreateTask(context.Background(), CreateInput{
		UserID: 42, Title: "优化成交转化率", Project: "增长测算", Priority: PriorityHigh,
		SourceType: SourceGrowthModel, SourceID: &sourceID, SourceTitle: "企业培训增长测算", SourceURL: "/growth-calculator",
	})

	if err != nil || repository.created.SourceType != SourceGrowthModel {
		t.Fatalf("err/source = %v/%+v", err, repository.created)
	}
}

func TestServiceRejectsUnsafeTaskSource(t *testing.T) {
	repository := &fakeRepository{}
	service := NewService(repository)

	_, err := service.CreateTask(context.Background(), CreateInput{
		UserID: 42, Title: "反击竞品更新", Project: "竞品动态监测", Priority: PriorityHigh,
		SourceType: SourceCompetitorScan, SourceTitle: "竞品扫描", SourceURL: "//evil.example/steal",
	})

	if !errors.Is(err, ErrInvalidTaskSource) || repository.created.ID != 0 {
		t.Fatalf("err/created = %v/%+v", err, repository.created)
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
		Tag:      " 用户研究 ",
		Query:    " 接口 ",
		Limit:    500,
		Offset:   -10,
	})

	if err != nil {
		t.Fatalf("ListTasks() error = %v", err)
	}
	if repository.filters.Status != StatusInProgress || repository.filters.Project != "商业沙盘" || repository.filters.Priority != PriorityHigh || repository.filters.Tag != "用户研究" || repository.filters.Query != "接口" || repository.filters.Limit != 100 || repository.filters.Offset != 0 {
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

func TestServiceListsTaskTags(t *testing.T) {
	repository := &fakeRepository{tags: []string{"用户研究", "访谈"}}
	service := NewService(repository)

	tags, err := service.ListTaskTags(context.Background(), 42)

	if err != nil {
		t.Fatalf("ListTaskTags() error = %v", err)
	}
	if len(tags) != 2 || tags[1] != "访谈" {
		t.Fatalf("tags = %+v", tags)
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

func TestServiceNormalizesUpdatedTaskFields(t *testing.T) {
	repository := &fakeRepository{task: Task{ID: 99, UserID: 42, Title: "整理客户", Status: StatusTodo}}
	service := NewService(repository)
	assignee := " 李明 "
	tags := []string{" 用户研究 ", "访谈", "用户研究"}

	_, err := service.UpdateTask(context.Background(), 42, 99, TaskUpdate{Assignee: &assignee, Tags: &tags})

	if err != nil {
		t.Fatalf("UpdateTask() error = %v", err)
	}
	if repository.updated.Assignee != "李明" {
		t.Fatalf("updated.Assignee = %q, want 李明", repository.updated.Assignee)
	}
	if len(repository.updated.Tags) != 2 || repository.updated.Tags[0] != "用户研究" {
		t.Fatalf("updated.Tags = %+v", repository.updated.Tags)
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

type fakeSubtaskRepository struct {
	*fakeRepository
	subtasks       []Subtask
	createdSubtask Subtask
	updatedSubtask SubtaskUpdate
	deletedTaskID  int64
	deletedID      int64
}

func (r *fakeSubtaskRepository) ListSubtasks(_ context.Context, userID, taskID int64) ([]Subtask, error) {
	if r.err != nil {
		return nil, r.err
	}
	items := make([]Subtask, 0, len(r.subtasks))
	for _, item := range r.subtasks {
		if item.UserID == userID && item.TaskID == taskID {
			items = append(items, item)
		}
	}
	return items, nil
}

func (r *fakeSubtaskRepository) CreateSubtask(_ context.Context, item Subtask) (Subtask, error) {
	r.createdSubtask = item
	item.ID = 7
	return item, r.err
}

func (r *fakeSubtaskRepository) UpdateSubtask(_ context.Context, userID, taskID, id int64, update SubtaskUpdate) (Subtask, error) {
	r.updatedSubtask = update
	return Subtask{ID: id, TaskID: taskID, UserID: userID}, r.err
}

func (r *fakeSubtaskRepository) DeleteSubtask(_ context.Context, _ int64, taskID, id int64) error {
	r.deletedTaskID, r.deletedID = taskID, id
	return r.err
}

func TestServiceCreatesTrimmedSubtask(t *testing.T) {
	now := time.Date(2026, 7, 10, 10, 0, 0, 0, time.UTC)
	repository := &fakeSubtaskRepository{fakeRepository: &fakeRepository{}}
	service := NewService(repository)
	service.now = func() time.Time { return now }

	item, err := service.CreateSubtask(context.Background(), CreateSubtaskInput{
		UserID: 42, TaskID: 99, Title: "  整理访谈提纲  ", Assignee: " 李明 ",
	})

	if err != nil {
		t.Fatalf("CreateSubtask() error = %v", err)
	}
	if item.ID != 7 || repository.createdSubtask.Title != "整理访谈提纲" || repository.createdSubtask.Assignee != "李明" || !repository.createdSubtask.CreatedAt.Equal(now) {
		t.Fatalf("item/created = %+v/%+v", item, repository.createdSubtask)
	}
}

func TestServiceListsOnlyParentTaskSubtasks(t *testing.T) {
	repository := &fakeSubtaskRepository{
		fakeRepository: &fakeRepository{},
		subtasks: []Subtask{
			{ID: 1, TaskID: 99, UserID: 42},
			{ID: 2, TaskID: 100, UserID: 42},
			{ID: 3, TaskID: 99, UserID: 7},
		},
	}
	service := NewService(repository)

	items, err := service.ListSubtasks(context.Background(), 42, 99)

	if err != nil || len(items) != 1 || items[0].ID != 1 {
		t.Fatalf("items/error = %+v/%v", items, err)
	}
}

func TestServiceNormalizesUpdatedSubtask(t *testing.T) {
	repository := &fakeSubtaskRepository{fakeRepository: &fakeRepository{}}
	service := NewService(repository)
	title, assignee := "  完成访谈提纲  ", " 王芳 "
	completed := true

	_, err := service.UpdateSubtask(context.Background(), 42, 99, 7, SubtaskUpdate{Title: &title, Assignee: &assignee, Completed: &completed})

	if err != nil {
		t.Fatalf("UpdateSubtask() error = %v", err)
	}
	if repository.updatedSubtask.Title == nil || *repository.updatedSubtask.Title != "完成访谈提纲" || repository.updatedSubtask.Assignee == nil || *repository.updatedSubtask.Assignee != "王芳" {
		t.Fatalf("update = %+v", repository.updatedSubtask)
	}
}

func TestServiceDeletesSubtaskFromParentTask(t *testing.T) {
	repository := &fakeSubtaskRepository{fakeRepository: &fakeRepository{}}
	service := NewService(repository)

	err := service.DeleteSubtask(context.Background(), 42, 99, 7)

	if err != nil || repository.deletedTaskID != 99 || repository.deletedID != 7 {
		t.Fatalf("task/id/error = %d/%d/%v", repository.deletedTaskID, repository.deletedID, err)
	}
}

type fakeReminderRepository struct {
	*fakeRepository
	reminder        *TaskReminder
	savedReminder   TaskReminder
	deletedReminder bool
	dispatchAt      time.Time
	dispatchLimit   int
	dispatchCount   int
}

func (r *fakeReminderRepository) GetTaskReminder(_ context.Context, userID, taskID int64) (*TaskReminder, error) {
	if r.err != nil {
		return nil, r.err
	}
	if r.reminder == nil || r.reminder.UserID != userID || r.reminder.TaskID != taskID {
		return nil, nil
	}
	item := *r.reminder
	return &item, nil
}

func (r *fakeReminderRepository) UpsertTaskReminder(_ context.Context, reminder TaskReminder) (TaskReminder, error) {
	r.savedReminder = reminder
	reminder.ID = 8
	return reminder, r.err
}

func (r *fakeReminderRepository) DeleteTaskReminder(_ context.Context, _, _ int64) error {
	r.deletedReminder = true
	return r.err
}

func (r *fakeReminderRepository) DispatchDueTaskReminders(_ context.Context, now time.Time, limit int) (int, error) {
	r.dispatchAt, r.dispatchLimit = now, limit
	return r.dispatchCount, r.err
}

func TestServiceUpsertsFutureTaskReminder(t *testing.T) {
	now := time.Date(2026, 7, 10, 10, 0, 0, 0, time.UTC)
	remindAt := now.Add(2 * time.Hour)
	repository := &fakeReminderRepository{fakeRepository: &fakeRepository{task: Task{ID: 99, UserID: 42}}}
	service := NewService(repository)
	service.now = func() time.Time { return now }

	reminder, err := service.UpsertTaskReminder(context.Background(), UpsertTaskReminderInput{UserID: 42, TaskID: 99, RemindAt: remindAt})

	if err != nil {
		t.Fatalf("UpsertTaskReminder() error = %v", err)
	}
	if reminder.ID != 8 || repository.savedReminder.UserID != 42 || repository.savedReminder.TaskID != 99 || repository.savedReminder.Recurrence != ReminderRecurrenceOnce || !repository.savedReminder.RemindAt.Equal(remindAt) || !repository.savedReminder.CreatedAt.Equal(now) {
		t.Fatalf("reminder/saved = %+v/%+v", reminder, repository.savedReminder)
	}
}

type fakeTaskMembership struct {
	planCode string
	err      error
}

func (m *fakeTaskMembership) CurrentSnapshot(context.Context, int64) (membership.Snapshot, error) {
	return membership.Snapshot{Plan: membership.PlanCatalog[m.planCode]}, m.err
}

func TestServiceRejectsRecurringReminderForFreePlan(t *testing.T) {
	now := time.Date(2026, 7, 11, 10, 0, 0, 0, time.UTC)
	repository := &fakeReminderRepository{fakeRepository: &fakeRepository{task: Task{ID: 99, UserID: 42}}}
	service := NewService(repository, WithMembershipProvider(&fakeTaskMembership{planCode: membership.PlanFree}))
	service.now = func() time.Time { return now }

	_, err := service.UpsertTaskReminder(context.Background(), UpsertTaskReminderInput{
		UserID: 42, TaskID: 99, RemindAt: now.Add(time.Hour), Recurrence: ReminderRecurrenceDaily,
	})

	if !errors.Is(err, ErrRecurringReminderRequiresMembership) || repository.savedReminder.ID != 0 {
		t.Fatalf("err/saved = %v/%+v", err, repository.savedReminder)
	}
}

func TestServiceAllowsRecurringReminderForProPlan(t *testing.T) {
	now := time.Date(2026, 7, 11, 10, 0, 0, 0, time.UTC)
	repository := &fakeReminderRepository{fakeRepository: &fakeRepository{task: Task{ID: 99, UserID: 42}}}
	service := NewService(repository, WithMembershipProvider(&fakeTaskMembership{planCode: membership.PlanPro}))
	service.now = func() time.Time { return now }

	_, err := service.UpsertTaskReminder(context.Background(), UpsertTaskReminderInput{
		UserID: 42, TaskID: 99, RemindAt: now.Add(time.Hour), Recurrence: ReminderRecurrenceWeekly,
	})

	if err != nil || repository.savedReminder.Recurrence != ReminderRecurrenceWeekly {
		t.Fatalf("saved/error = %+v/%v", repository.savedReminder, err)
	}
}

func TestServiceRejectsInvalidReminderRecurrence(t *testing.T) {
	now := time.Date(2026, 7, 11, 10, 0, 0, 0, time.UTC)
	repository := &fakeReminderRepository{fakeRepository: &fakeRepository{task: Task{ID: 99, UserID: 42}}}
	service := NewService(repository)
	service.now = func() time.Time { return now }

	_, err := service.UpsertTaskReminder(context.Background(), UpsertTaskReminderInput{
		UserID: 42, TaskID: 99, RemindAt: now.Add(time.Hour), Recurrence: "monthly",
	})

	if !errors.Is(err, ErrInvalidReminderRecurrence) {
		t.Fatalf("err = %v, want ErrInvalidReminderRecurrence", err)
	}
}

func TestServiceRejectsPastTaskReminder(t *testing.T) {
	now := time.Date(2026, 7, 10, 10, 0, 0, 0, time.UTC)
	repository := &fakeReminderRepository{fakeRepository: &fakeRepository{task: Task{ID: 99, UserID: 42}}}
	service := NewService(repository)
	service.now = func() time.Time { return now }

	_, err := service.UpsertTaskReminder(context.Background(), UpsertTaskReminderInput{UserID: 42, TaskID: 99, RemindAt: now})

	if !errors.Is(err, ErrInvalidReminderTime) || repository.savedReminder.ID != 0 {
		t.Fatalf("err/saved = %v/%+v, want ErrInvalidReminderTime and no save", err, repository.savedReminder)
	}
}

func TestServiceGetsEmptyReminderForOwnedTask(t *testing.T) {
	repository := &fakeReminderRepository{fakeRepository: &fakeRepository{task: Task{ID: 99, UserID: 42}}}
	service := NewService(repository)

	reminder, err := service.GetTaskReminder(context.Background(), 42, 99)

	if err != nil || reminder != nil {
		t.Fatalf("reminder/error = %+v/%v", reminder, err)
	}
}

func TestServiceDeletesOwnedTaskReminder(t *testing.T) {
	repository := &fakeReminderRepository{fakeRepository: &fakeRepository{task: Task{ID: 99, UserID: 42}}}
	service := NewService(repository)

	err := service.DeleteTaskReminder(context.Background(), 42, 99)

	if err != nil || !repository.deletedReminder {
		t.Fatalf("deleted/error = %t/%v", repository.deletedReminder, err)
	}
}

func TestServiceDispatchesDueTaskRemindersAtCurrentTime(t *testing.T) {
	now := time.Date(2026, 7, 10, 10, 0, 0, 0, time.UTC)
	repository := &fakeReminderRepository{fakeRepository: &fakeRepository{}, dispatchCount: 2}
	service := NewService(repository)
	service.now = func() time.Time { return now }

	count, err := service.DispatchDueTaskReminders(context.Background(), 0)

	if err != nil || count != 2 || !repository.dispatchAt.Equal(now) || repository.dispatchLimit != 100 {
		t.Fatalf("count/at/limit/error = %d/%v/%d/%v", count, repository.dispatchAt, repository.dispatchLimit, err)
	}
}
