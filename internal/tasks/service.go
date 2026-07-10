package tasks

import (
	"context"
	"strings"
	"time"
)

type Repository interface {
	CreateTask(ctx context.Context, task Task) (Task, error)
	ListTasks(ctx context.Context, userID int64, filters ListFilters) ([]Task, error)
	CountTasks(ctx context.Context, userID int64, filters ListFilters) (int, error)
	ListTaskProjects(ctx context.Context, userID int64) ([]string, error)
	ListTaskTags(ctx context.Context, userID int64) ([]string, error)
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
		UserID:      input.UserID,
		Title:       strings.TrimSpace(input.Title),
		Description: strings.TrimSpace(input.Description),
		Assignee:    strings.TrimSpace(input.Assignee),
		Project:     strings.TrimSpace(input.Project),
		Status:      StatusTodo,
		Priority:    normalizePriority(input.Priority),
		Tags:        normalizeUniqueStrings(input.Tags),
		DueAt:       input.DueAt,
		Tools:       normalizeStrings(input.Tools),
		Learning:    strings.TrimSpace(input.Learning),
		CreatedAt:   s.now(),
	}
	return s.repository.CreateTask(ctx, task)
}

func (s *Service) ListTasks(ctx context.Context, userID int64, filters ListFilters) ([]Task, error) {
	if s.repository == nil {
		return nil, ErrServiceNotReady
	}
	return s.repository.ListTasks(ctx, userID, normalizeListFilters(filters))
}

func (s *Service) ListTaskPage(ctx context.Context, userID int64, filters ListFilters) (TaskPage, error) {
	if s.repository == nil {
		return TaskPage{}, ErrServiceNotReady
	}
	filters = normalizeListFilters(filters)
	tasks, err := s.repository.ListTasks(ctx, userID, filters)
	if err != nil {
		return TaskPage{}, err
	}
	total, err := s.repository.CountTasks(ctx, userID, filters)
	if err != nil {
		return TaskPage{}, err
	}
	return TaskPage{Tasks: tasks, Total: total, Limit: filters.Limit, Offset: filters.Offset}, nil
}

func normalizeListFilters(filters ListFilters) ListFilters {
	filters.Status = strings.TrimSpace(filters.Status)
	filters.Project = strings.TrimSpace(filters.Project)
	filters.Priority = strings.TrimSpace(filters.Priority)
	filters.Tag = strings.TrimSpace(filters.Tag)
	filters.Query = strings.TrimSpace(filters.Query)
	if filters.Status != "" {
		filters.Status = normalizeStatus(filters.Status)
	}
	if filters.Priority != "" {
		filters.Priority = normalizePriority(filters.Priority)
	}
	if filters.Limit <= 0 {
		filters.Limit = 20
	}
	if filters.Limit > 100 {
		filters.Limit = 100
	}
	if filters.Offset < 0 {
		filters.Offset = 0
	}
	return filters
}

func (s *Service) ListTaskProjects(ctx context.Context, userID int64) ([]string, error) {
	if s.repository == nil {
		return nil, ErrServiceNotReady
	}
	return s.repository.ListTaskProjects(ctx, userID)
}

func (s *Service) ListTaskTags(ctx context.Context, userID int64) ([]string, error) {
	if s.repository == nil {
		return nil, ErrServiceNotReady
	}
	return s.repository.ListTaskTags(ctx, userID)
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
	if update.Description != nil {
		description := strings.TrimSpace(*update.Description)
		update.Description = &description
	}
	if update.Assignee != nil {
		assignee := strings.TrimSpace(*update.Assignee)
		update.Assignee = &assignee
	}
	if update.Project != nil {
		project := strings.TrimSpace(*update.Project)
		update.Project = &project
	}
	if update.Tools != nil {
		tools := normalizeStrings(*update.Tools)
		update.Tools = &tools
	}
	if update.Tags != nil {
		tags := normalizeUniqueStrings(*update.Tags)
		update.Tags = &tags
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

func normalizeUniqueStrings(values []string) []string {
	normalized := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		normalized = append(normalized, value)
	}
	return normalized
}
