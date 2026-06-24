package leads

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/zzm/opcv2/internal/jobs"
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

type fakeCredits struct {
	charged  []int64
	refunded []int64
}

func (c *fakeCredits) Charge(_ context.Context, userID int64, amount int, referenceType string, referenceID int64) error {
	c.charged = append(c.charged, referenceID)
	return nil
}

func (c *fakeCredits) Refund(_ context.Context, userID int64, amount int, referenceType string, referenceID int64) error {
	c.refunded = append(c.refunded, referenceID)
	return nil
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

func TestServiceCreatesTaskChargesCreditsAndEnqueues(t *testing.T) {
	repository := &memoryRepository{}
	credits := &fakeCredits{}
	queue := &fakeQueue{}
	service := NewService(repository, credits, queue, nil)
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
	if len(credits.charged) != 1 || credits.charged[0] != task.ID {
		t.Fatalf("charged = %+v", credits.charged)
	}
	if len(queue.jobs) != 1 || queue.jobs[0].Type != jobs.TypeLeadSearch || queue.jobs[0].IdempotencyKey != "lead-task-42" {
		t.Fatalf("jobs = %+v", queue.jobs)
	}
}

func TestServiceCreateTaskIsIdempotent(t *testing.T) {
	repository := &memoryRepository{}
	credits := &fakeCredits{}
	queue := &fakeQueue{}
	service := NewService(repository, credits, queue, nil)

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
	if len(credits.charged) != 1 || len(queue.jobs) != 1 {
		t.Fatalf("charged=%+v jobs=%+v", credits.charged, queue.jobs)
	}
}

func TestServiceWorkerStoresResultsOnSuccess(t *testing.T) {
	repository := &memoryRepository{}
	provider := &fakeLeadProvider{leads: []Lead{{Name: "成都启明星教育", Phone: "028-12345678", Website: "https://example.com"}}}
	service := NewService(repository, &fakeCredits{}, &fakeQueue{}, provider)
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
	credits := &fakeCredits{}
	provider := &fakeLeadProvider{err: ErrProviderUnavailable}
	service := NewService(repository, credits, &fakeQueue{}, provider)
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
	if len(credits.refunded) != 1 || credits.refunded[0] != task.ID {
		t.Fatalf("refunded = %+v", credits.refunded)
	}
}

func TestServiceWorkerDoesNotRefundProviderQuotaFailure(t *testing.T) {
	repository := &memoryRepository{}
	credits := &fakeCredits{}
	provider := &fakeLeadProvider{err: ErrProviderQuotaExceeded}
	service := NewService(repository, credits, &fakeQueue{}, provider)
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
	if len(credits.refunded) != 0 {
		t.Fatalf("refunded = %+v", credits.refunded)
	}
}
