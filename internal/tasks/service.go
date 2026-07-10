package tasks

import (
	"context"
	"strings"
	"time"
)

type Repository interface {
	CreateTask(ctx context.Context, task Task) (Task, error)
	ListTasks(ctx context.Context, userID int64, filters ListFilters) ([]Task, error)
	TaskStats(ctx context.Context, userID int64, now time.Time) (Stats, error)
	GetTask(ctx context.Context, userID, id int64) (Task, error)
	UpdateTask(ctx context.Context, userID, id int64, update TaskUpdate) (Task, error)
	DeleteTask(ctx context.Context, userID, id int64) error
}

type Service struct {
	repository Repository
	now        func() time.Time
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository, now: time.Now}
}

func (s *Service) CreateTask(ctx context.Context, input CreateInput) (Task, error) {
	if s.repository == nil {
		return Task{}, ErrServiceNotReady
	}
	task := Task{
		UserID:    input.UserID,
		Title:     strings.TrimSpace(input.Title),
		Project:   strings.TrimSpace(input.Project),
		Status:    StatusTodo,
		Priority:  normalizePriority(input.Priority),
		DueAt:     input.DueAt,
		Tools:     normalizeStrings(input.Tools),
		Learning:  strings.TrimSpace(input.Learning),
		CreatedAt: s.now(),
	}
	return s.repository.CreateTask(ctx, task)
}

func (s *Service) ListTasks(ctx context.Context, userID int64, filters ListFilters) ([]Task, error) {
	if s.repository == nil {
		return nil, ErrServiceNotReady
	}
	filters.Status = strings.TrimSpace(filters.Status)
	filters.Project = strings.TrimSpace(filters.Project)
	filters.Query = strings.TrimSpace(filters.Query)
	if filters.Status != "" {
		filters.Status = normalizeStatus(filters.Status)
	}
	if filters.Limit <= 0 {
		filters.Limit = 20
	}
	if filters.Limit > 100 {
		filters.Limit = 100
	}
	return s.repository.ListTasks(ctx, userID, filters)
}

func (s *Service) TaskStats(ctx context.Context, userID int64) (Stats, error) {
	if s.repository == nil {
		return Stats{}, ErrServiceNotReady
	}
	return s.repository.TaskStats(ctx, userID, s.now())
}

func (s *Service) GetTask(ctx context.Context, userID, id int64) (Task, error) {
	if s.repository == nil {
		return Task{}, ErrServiceNotReady
	}
	return s.repository.GetTask(ctx, userID, id)
}

func (s *Service) UpdateTask(ctx context.Context, userID, id int64, update TaskUpdate) (Task, error) {
	if s.repository == nil {
		return Task{}, ErrServiceNotReady
	}
	if update.Status != nil {
		status := normalizeStatus(*update.Status)
		update.Status = &status
	}
	if update.Priority != nil {
		priority := normalizePriority(*update.Priority)
		update.Priority = &priority
	}
	if update.Title != nil {
		title := strings.TrimSpace(*update.Title)
		update.Title = &title
	}
	if update.Project != nil {
		project := strings.TrimSpace(*update.Project)
		update.Project = &project
	}
	if update.Tools != nil {
		tools := normalizeStrings(*update.Tools)
		update.Tools = &tools
	}
	if update.Learning != nil {
		learning := strings.TrimSpace(*update.Learning)
		update.Learning = &learning
	}
	return s.repository.UpdateTask(ctx, userID, id, update)
}

func (s *Service) DeleteTask(ctx context.Context, userID, id int64) error {
	if s.repository == nil {
		return ErrServiceNotReady
	}
	return s.repository.DeleteTask(ctx, userID, id)
}

func normalizeStatus(status string) string {
	switch status {
	case StatusInProgress, StatusCompleted, StatusReminder:
		return status
	default:
		return StatusTodo
	}
}

func normalizePriority(priority string) string {
	switch priority {
	case PriorityLow, PriorityHigh:
		return priority
	default:
		return PriorityMedium
	}
}

func normalizeStrings(values []string) []string {
	normalized := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			normalized = append(normalized, value)
		}
	}
	return normalized
}
