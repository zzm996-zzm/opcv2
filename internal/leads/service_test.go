package leads

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/zzm/opcv2/internal/jobs"
	"github.com/zzm/opcv2/internal/membership"
)

type memoryRepository struct {
	tasks   []Task
	results map[int64][]Lead
	nextID  int64
}

func (r *memoryRepository) CreateTask(_ context.Context, task Task) (Task, bool, error) {
	for _, existing := range r.tasks {
		if existing.UserID == task.UserID && existing.IdempotencyKey == task.IdempotencyKey {
			return existing, true, nil
		}
	}
	if r.nextID == 0 {
		r.nextID = 100
	}
	task.ID = r.nextID
	r.nextID++
	r.tasks = append(r.tasks, task)
	return task, false, nil
}

func (r *memoryRepository) GetTask(_ context.Context, id int64) (Task, error) {
	for _, task := range r.tasks {
		if task.ID == id {
			return task, nil
		}
	}
	return Task{}, ErrTaskNotFound
}

func (r *memoryRepository) UpdateTaskStatus(_ context.Context, id int64, status string, errorCode string) error {
	for index := range r.tasks {
		if r.tasks[index].ID == id {
			r.tasks[index].Status = status
			r.tasks[index].ErrorCode = errorCode
			return nil
		}
	}
	return ErrTaskNotFound
}

func (r *memoryRepository) StoreResults(_ context.Context, taskID int64, leads []Lead) error {
	if r.results == nil {
		r.results = map[int64][]Lead{}
	}
	r.results[taskID] = leads
	return nil
}

func (r *memoryRepository) ListResults(_ context.Context, taskID int64, limit int) ([]LeadResult, error) {
	leads := r.results[taskID]
	results := make([]LeadResult, 0, len(leads))
	for index, lead := range leads {
		results = append(results, LeadResult{
			ID:       int64(index + 1),
			TaskID:   taskID,
			Name:     lead.Name,
			Phone:    lead.Phone,
			Email:    lead.Email,
			Website:  lead.Website,
			Evidence: lead.Evidence,
		})
	}
	if limit > 0 && len(results) > limit {
		return results[:limit], nil
	}
	return results, nil
}

func (r *memoryRepository) ListTasks(_ context.Context, userID int64, limit int) ([]Task, error) {
	var tasks []Task
	for _, task := range r.tasks {
		if task.UserID == userID {
			tasks = append(tasks, task)
		}
	}
	if limit > 0 && len(tasks) > limit {
		return tasks[:limit], nil
	}
	return tasks, nil
}

type fakeQuota struct {
	consumed []membership.ConsumeInput
	refunded []membership.ConsumeInput
	err      error
}

func (q *fakeQuota) CheckAndConsume(_ context.Context, input membership.ConsumeInput) (membership.UsageItem, error) {
	q.consumed = append(q.consumed, input)
	if q.err != nil {
		return membership.UsageItem{}, q.err
	}
	return membership.UsageItem{Key: input.FeatureKey, Used: len(q.consumed), Limit: 30, Unit: "次/月"}, nil
}

func (q *fakeQuota) RefundUsage(_ context.Context, input membership.ConsumeInput) (membership.UsageItem, error) {
	q.refunded = append(q.refunded, input)
	return membership.UsageItem{Key: input.FeatureKey, Used: 0, Limit: 30, Unit: "次/月"}, nil
}

type fakeQueue struct {
	jobs []jobs.Job
}

func (q *fakeQueue) Enqueue(_ context.Context, job jobs.Job) error {
	q.jobs = append(q.jobs, job)
	return nil
}

type fakeLeadProvider struct {
	leads []Lead
	err   error
}

func (p *fakeLeadProvider) SearchLeads(_ context.Context, input SearchInput) ([]Lead, error) {
	return p.leads, p.err
}

func TestServiceCreatesTaskConsumesQuotaAndEnqueues(t *testing.T) {
	repository := &memoryRepository{}
	quota := &fakeQuota{}
	queue := &fakeQueue{}
	service := NewService(repository, quota, queue, nil)
	service.now = func() time.Time { return time.Date(2026, 6, 24, 10, 0, 0, 0, time.UTC) }

	task, err := service.CreateTask(context.Background(), CreateTaskInput{
		UserID:         42,
		Query:          "成都 教培",
		IdempotencyKey: "lead-task-42",
	})
	if err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}
	if task.Status != StatusQueued || task.ID == 0 {
		t.Fatalf("task = %+v", task)
	}
	if len(quota.consumed) != 1 || quota.consumed[0].FeatureKey != membership.FeatureLeadTasks || quota.consumed[0].IdempotencyKey != "lead-task-42" {
		t.Fatalf("consumed = %+v", quota.consumed)
	}
	if len(queue.jobs) != 1 || queue.jobs[0].Type != jobs.TypeLeadSearch || queue.jobs[0].IdempotencyKey != "lead-task-42" {
		t.Fatalf("jobs = %+v", queue.jobs)
	}
}

func TestServiceCreateTaskIsIdempotent(t *testing.T) {
	repository := &memoryRepository{}
	quota := &fakeQuota{}
	queue := &fakeQueue{}
	service := NewService(repository, quota, queue, nil)

	first, err := service.CreateTask(context.Background(), CreateTaskInput{UserID: 42, Query: "成都 教培", IdempotencyKey: "same-key"})
	if err != nil {
		t.Fatalf("CreateTask() first error = %v", err)
	}
	second, err := service.CreateTask(context.Background(), CreateTaskInput{UserID: 42, Query: "成都 教培", IdempotencyKey: "same-key"})
	if err != nil {
		t.Fatalf("CreateTask() second error = %v", err)
	}
	if first.ID != second.ID {
		t.Fatalf("first=%+v second=%+v", first, second)
	}
	if len(quota.consumed) != 2 || len(queue.jobs) != 1 {
		t.Fatalf("consumed=%+v jobs=%+v", quota.consumed, queue.jobs)
	}
}

func TestServiceCreateTaskStopsWhenQuotaExceeded(t *testing.T) {
	repository := &memoryRepository{}
	quota := &fakeQuota{err: membership.ErrQuotaExceeded}
	queue := &fakeQueue{}
	service := NewService(repository, quota, queue, nil)

	_, err := service.CreateTask(context.Background(), CreateTaskInput{UserID: 42, Query: "成都 教培", IdempotencyKey: "quota-key"})

	if !errors.Is(err, membership.ErrQuotaExceeded) {
		t.Fatalf("CreateTask() error = %v, want ErrQuotaExceeded", err)
	}
	if len(repository.tasks) != 0 || len(queue.jobs) != 0 {
		t.Fatalf("tasks=%+v jobs=%+v", repository.tasks, queue.jobs)
	}
}

func TestServiceWorkerStoresResultsOnSuccess(t *testing.T) {
	repository := &memoryRepository{}
	provider := &fakeLeadProvider{leads: []Lead{{Name: "成都启明星教育", Phone: "028-12345678", Website: "https://example.com"}}}
	service := NewService(repository, &fakeQuota{}, &fakeQueue{}, provider)
	task, _, err := repository.CreateTask(context.Background(), Task{UserID: 42, Query: "成都 教培", Status: StatusQueued, IdempotencyKey: "lead-task-42"})
	if err != nil {
		t.Fatalf("CreateTask fixture error = %v", err)
	}

	err = service.ProcessTask(context.Background(), task.ID)
	if err != nil {
		t.Fatalf("ProcessTask() error = %v", err)
	}
	updated, err := repository.GetTask(context.Background(), task.ID)
	if err != nil {
		t.Fatalf("GetTask() error = %v", err)
	}
	if updated.Status != StatusSucceeded {
		t.Fatalf("updated task = %+v", updated)
	}
	if len(repository.results[task.ID]) != 1 || repository.results[task.ID][0].Name != "成都启明星教育" {
		t.Fatalf("results = %+v", repository.results)
	}
}

func TestServiceWorkerRefundsOnSystemFailure(t *testing.T) {
	repository := &memoryRepository{}
	quota := &fakeQuota{}
	provider := &fakeLeadProvider{err: ErrProviderUnavailable}
	service := NewService(repository, quota, &fakeQueue{}, provider)
	task, _, err := repository.CreateTask(context.Background(), Task{UserID: 42, Query: "成都 教培", Status: StatusQueued, IdempotencyKey: "lead-task-42", CreditCost: 1})
	if err != nil {
		t.Fatalf("CreateTask fixture error = %v", err)
	}

	err = service.ProcessTask(context.Background(), task.ID)
	if !errors.Is(err, ErrProviderUnavailable) {
		t.Fatalf("ProcessTask() error = %v, want ErrProviderUnavailable", err)
	}
	updated, _ := repository.GetTask(context.Background(), task.ID)
	if updated.Status != StatusRefunded {
		t.Fatalf("updated task = %+v", updated)
	}
	if len(quota.refunded) != 1 || quota.refunded[0].FeatureKey != membership.FeatureLeadTasks {
		t.Fatalf("refunded = %+v", quota.refunded)
	}
}

func TestServiceWorkerDoesNotRefundProviderQuotaFailure(t *testing.T) {
	repository := &memoryRepository{}
	quota := &fakeQuota{}
	provider := &fakeLeadProvider{err: ErrProviderQuotaExceeded}
	service := NewService(repository, quota, &fakeQueue{}, provider)
	task, _, err := repository.CreateTask(context.Background(), Task{UserID: 42, Query: "成都 教培", Status: StatusQueued, IdempotencyKey: "lead-task-42", CreditCost: 1})
	if err != nil {
		t.Fatalf("CreateTask fixture error = %v", err)
	}

	err = service.ProcessTask(context.Background(), task.ID)
	if !errors.Is(err, ErrProviderQuotaExceeded) {
		t.Fatalf("ProcessTask() error = %v, want ErrProviderQuotaExceeded", err)
	}
	updated, _ := repository.GetTask(context.Background(), task.ID)
	if updated.Status != StatusFailed || updated.ErrorCode != ErrorProviderQuotaExceeded {
		t.Fatalf("updated task = %+v", updated)
	}
	if len(quota.refunded) != 0 {
		t.Fatalf("refunded = %+v", quota.refunded)
	}
}

func TestServiceGetsTaskDetailForOwner(t *testing.T) {
	repository := &memoryRepository{results: map[int64][]Lead{
		100: {{Name: "成都启明星教育"}},
	}}
	task, _, err := repository.CreateTask(context.Background(), Task{UserID: 42, Query: "成都 教培", Status: StatusSucceeded, IdempotencyKey: "lead-task-42"})
	if err != nil {
		t.Fatalf("CreateTask fixture error = %v", err)
	}
	service := NewService(repository, &fakeQuota{}, &fakeQueue{}, nil)

	detail, err := service.GetTask(context.Background(), 42, task.ID)

	if err != nil {
		t.Fatalf("GetTask() error = %v", err)
	}
	if detail.Task.ID != task.ID || detail.ProgressPercent != 100 || detail.ResultsCount != 1 {
		t.Fatalf("detail = %+v", detail)
	}
}

func TestServiceRejectsOtherUsersLeadTask(t *testing.T) {
	repository := &memoryRepository{}
	task, _, err := repository.CreateTask(context.Background(), Task{UserID: 7, Query: "成都 教培", Status: StatusSucceeded, IdempotencyKey: "lead-task-7"})
	if err != nil {
		t.Fatalf("CreateTask fixture error = %v", err)
	}
	service := NewService(repository, &fakeQuota{}, &fakeQueue{}, nil)

	_, err = service.GetTask(context.Background(), 42, task.ID)

	if !errors.Is(err, ErrTaskNotFound) {
		t.Fatalf("err = %v, want ErrTaskNotFound", err)
	}
}

func TestServiceListsTaskResultsForOwner(t *testing.T) {
	repository := &memoryRepository{results: map[int64][]Lead{
		100: {
			{Name: "成都启明星教育", Phone: "028-12345678", Evidence: []Evidence{{Type: "website", Title: "官网", URL: "https://example.com"}}},
			{Name: "橙果职业培训"},
		},
	}}
	task, _, err := repository.CreateTask(context.Background(), Task{UserID: 42, Query: "成都 教培", Status: StatusSucceeded, IdempotencyKey: "lead-task-42"})
	if err != nil {
		t.Fatalf("CreateTask fixture error = %v", err)
	}
	service := NewService(repository, &fakeQuota{}, &fakeQueue{}, nil)

	results, err := service.ListResults(context.Background(), 42, task.ID, 1)

	if err != nil {
		t.Fatalf("ListResults() error = %v", err)
	}
	if len(results) != 1 || results[0].Name != "成都启明星教育" || len(results[0].Evidence) != 1 {
		t.Fatalf("results = %+v", results)
	}
}
