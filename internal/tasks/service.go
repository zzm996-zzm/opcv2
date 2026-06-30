package tasks

import (
	"context"
	"strings"
	"time"
)

type Repository interface {
	CreateTask(ctx context.Context, task Task) (Task, error)
	ListTasks(ctx context.Context, userID int64, limit int) ([]Task, error)
	GetTask(ctx context.Context, userID, id int64) (Task, error)
	UpdateTask(ctx context.Context, userID, id int64, update TaskUpdate) (Task, error)
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

func (s *Service) ListTasks(ctx context.Context, userID int64, limit int) ([]Task, error) {
	if s.repository == nil {
		return nil, ErrServiceNotReady
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return s.repository.ListTasks(ctx, userID, limit)
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
