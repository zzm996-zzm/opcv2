package leads

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/zzm/opcv2/internal/jobs"
)

type Repository interface {
	CreateTask(ctx context.Context, task Task) (Task, bool, error)
	GetTask(ctx context.Context, id int64) (Task, error)
	UpdateTaskStatus(ctx context.Context, id int64, status string, errorCode string) error
	StoreResults(ctx context.Context, taskID int64, leads []Lead) error
	ListResults(ctx context.Context, taskID int64, limit int) ([]LeadResult, error)
	ListTasks(ctx context.Context, userID int64, limit int) ([]Task, error)
}

type CreditLedger interface {
	Charge(ctx context.Context, userID int64, amount int, referenceType string, referenceID int64) error
	Refund(ctx context.Context, userID int64, amount int, referenceType string, referenceID int64) error
}

type Queue interface {
	Enqueue(ctx context.Context, job jobs.Job) error
}

type LeadProvider interface {
	SearchLeads(ctx context.Context, input SearchInput) ([]Lead, error)
}

type Service struct {
	repository Repository
	credits    CreditLedger
	queue      Queue
	provider   LeadProvider
	now        func() time.Time
}

func NewService(repository Repository, credits CreditLedger, queue Queue, provider LeadProvider) *Service {
	return &Service{
		repository: repository,
		credits:    credits,
		queue:      queue,
		provider:   provider,
		now:        time.Now,
	}
}

func (s *Service) CreateTask(ctx context.Context, input CreateTaskInput) (Task, error) {
	input.Query = strings.TrimSpace(input.Query)
	if s.repository == nil || s.credits == nil || s.queue == nil {
		return Task{}, ErrServiceNotReady
	}
	if input.UserID <= 0 || input.Query == "" || input.IdempotencyKey == "" {
		return Task{}, ErrInvalidTaskInput
	}
	task, existed, err := s.repository.CreateTask(ctx, Task{
		UserID:         input.UserID,
		Query:          input.Query,
		Status:         StatusQueued,
		IdempotencyKey: input.IdempotencyKey,
		CreditCost:     defaultCreditCost,
		CreatedAt:      s.now(),
	})
	if err != nil {
		return Task{}, err
	}
	if existed {
		return task, nil
	}
	if err := s.credits.Charge(ctx, task.UserID, task.CreditCost, "lead_task", task.ID); err != nil {
		return Task{}, err
	}
	if err := s.queue.Enqueue(ctx, jobs.Job{
		Type:           jobs.TypeLeadSearch,
		IdempotencyKey: input.IdempotencyKey,
		Payload: map[string]any{
			"task_id": task.ID,
		},
		MaxRetry: 3,
		Timeout:  5 * time.Minute,
	}); err != nil {
		return Task{}, err
	}
	return task, nil
}

func (s *Service) ProcessTask(ctx context.Context, taskID int64) error {
	if s.repository == nil || s.credits == nil || s.provider == nil {
		return ErrServiceNotReady
	}
	task, err := s.repository.GetTask(ctx, taskID)
	if err != nil {
		return err
	}
	if err := s.repository.UpdateTaskStatus(ctx, task.ID, StatusRunning, ""); err != nil {
		return err
	}
	leads, err := s.provider.SearchLeads(ctx, SearchInput{Query: task.Query})
	if err != nil {
		return s.failTask(ctx, task, err)
	}
	if err := s.repository.StoreResults(ctx, task.ID, leads); err != nil {
		return err
	}
	return s.repository.UpdateTaskStatus(ctx, task.ID, StatusSucceeded, "")
}

func (s *Service) ListTasks(ctx context.Context, userID int64, limit int) ([]Task, error) {
	if s.repository == nil {
		return nil, ErrServiceNotReady
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return s.repository.ListTasks(ctx, userID, limit)
}

func (s *Service) GetTask(ctx context.Context, userID, id int64) (TaskDetail, error) {
	if s.repository == nil {
		return TaskDetail{}, ErrServiceNotReady
	}
	task, err := s.repository.GetTask(ctx, id)
	if err != nil {
		return TaskDetail{}, err
	}
	if task.UserID != userID {
		return TaskDetail{}, ErrTaskNotFound
	}
	results, err := s.repository.ListResults(ctx, task.ID, 100)
	if err != nil {
		return TaskDetail{}, err
	}
	resultsCount := len(results)
	return TaskDetail{
		Task:            task,
		ProgressPercent: taskProgress(task.Status),
		Message:         taskMessage(task),
		ResultsCount:    resultsCount,
	}, nil
}

func (s *Service) ListResults(ctx context.Context, userID, taskID int64, limit int) ([]LeadResult, error) {
	if s.repository == nil {
		return nil, ErrServiceNotReady
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	task, err := s.repository.GetTask(ctx, taskID)
	if err != nil {
		return nil, err
	}
	if task.UserID != userID {
		return nil, ErrTaskNotFound
	}
	return s.repository.ListResults(ctx, taskID, limit)
}

func (s *Service) failTask(ctx context.Context, task Task, cause error) error {
	code := errorCode(cause)
	if refundable(cause) {
		if err := s.credits.Refund(ctx, task.UserID, task.CreditCost, "lead_task", task.ID); err != nil {
			return err
		}
		_ = s.repository.UpdateTaskStatus(ctx, task.ID, StatusRefunded, code)
		return cause
	}
	_ = s.repository.UpdateTaskStatus(ctx, task.ID, StatusFailed, code)
	return cause
}

func errorCode(err error) string {
	switch {
	case errors.Is(err, ErrProviderTimeout):
		return ErrorProviderTimeout
	case errors.Is(err, ErrProviderRateLimited):
		return ErrorProviderRateLimited
	case errors.Is(err, ErrProviderQuotaExceeded):
		return ErrorProviderQuotaExceeded
	case errors.Is(err, ErrProviderUnavailable):
		return ErrorProviderUnavailable
	default:
		return "internal_error"
	}
}

func refundable(err error) bool {
	return errors.Is(err, ErrProviderUnavailable) || errors.Is(err, ErrProviderTimeout)
}

func taskProgress(status string) int {
	switch status {
	case StatusQueued:
		return 20
	case StatusRunning:
		return 60
	case StatusSucceeded:
		return 100
	case StatusFailed, StatusRefunded, StatusCancelled:
		return 100
	default:
		return 0
	}
}

func taskMessage(task Task) string {
	switch task.Status {
	case StatusQueued:
		return "任务已进入队列，等待开始采集。"
	case StatusRunning:
		return "正在采集公开线索并进行可触达性判断。"
	case StatusSucceeded:
		return "线索采集已完成，可以查看结果并导入 CRM。"
	case StatusRefunded:
		return "供应商暂不可用，本次额度已退回。"
	case StatusFailed:
		if task.ErrorCode != "" {
			return "任务失败：" + task.ErrorCode
		}
		return "任务失败，请稍后重试。"
	case StatusCancelled:
		return "任务已取消。"
	default:
		return "任务状态待确认。"
	}
}
