package tasks

import (
	"context"
	"strings"
	"time"

	"github.com/zzm/opcv2/internal/ai"
	"github.com/zzm/opcv2/internal/membership"
)

type Repository interface {
	CreateTask(ctx context.Context, task Task) (Task, error)
	CreateTasks(ctx context.Context, tasks []Task) ([]Task, error)
	ListTasks(ctx context.Context, userID int64, filters ListFilters) ([]Task, error)
	CountTasks(ctx context.Context, userID int64, filters ListFilters) (int, error)
	ListTaskProjects(ctx context.Context, userID int64) ([]string, error)
	ListTaskTags(ctx context.Context, userID int64) ([]string, error)
	TaskStats(ctx context.Context, userID int64, now time.Time) (Stats, error)
	GetTask(ctx context.Context, userID, id int64) (Task, error)
	UpdateTask(ctx context.Context, userID, id int64, update TaskUpdate) (Task, error)
	DeleteTask(ctx context.Context, userID, id int64) error
	RestoreTask(ctx context.Context, userID, id int64) error
	BatchUpdateTaskStatus(ctx context.Context, userID int64, ids []int64, status string) (int, error)
	BatchDeleteTasks(ctx context.Context, userID int64, ids []int64) (int, error)
	CreateTaskAIDraft(ctx context.Context, draft TaskAIDraft) (TaskAIDraft, error)
	GetTaskAIDraft(ctx context.Context, userID, id int64) (TaskAIDraft, error)
	AdoptTaskAIDraft(ctx context.Context, userID, id int64, tasks []Task) ([]Task, error)
	ListTaskActivities(ctx context.Context, userID, taskID int64, limit, offset int) ([]TaskActivity, error)
	CountTaskActivities(ctx context.Context, userID, taskID int64) (int, error)
}

type MembershipProvider interface {
	CurrentSnapshot(context.Context, int64) (membership.Snapshot, error)
}

type TaskGenerator interface {
	GenerateJSON(context.Context, ai.GenerateJSONRequest) (ai.GenerateJSONResult, error)
}

type Option func(*Service)

type Service struct {
	repository Repository
	membership MembershipProvider
	generator  TaskGenerator
	now        func() time.Time
}

func NewService(repository Repository, options ...Option) *Service {
	service := &Service{repository: repository, now: time.Now}
	for _, option := range options {
		option(service)
	}
	return service
}

func WithMembershipProvider(provider MembershipProvider) Option {
	return func(service *Service) {
		service.membership = provider
	}
}

func WithTaskGenerator(generator TaskGenerator) Option {
	return func(service *Service) {
		service.generator = generator
	}
}

func (s *Service) CreateTask(ctx context.Context, input CreateInput) (Task, error) {
	if s.repository == nil {
		return Task{}, ErrServiceNotReady
	}
	input.SourceType = strings.TrimSpace(input.SourceType)
	input.SourceTitle = strings.TrimSpace(input.SourceTitle)
	input.SourceURL = strings.TrimSpace(input.SourceURL)
	if !validTaskSource(input.SourceType, input.SourceID, input.SourceTitle, input.SourceURL) {
		return Task{}, ErrInvalidTaskSource
	}
	task := Task{
		UserID:         input.UserID,
		Title:          strings.TrimSpace(input.Title),
		Description:    strings.TrimSpace(input.Description),
		Assignee:       strings.TrimSpace(input.Assignee),
		Project:        strings.TrimSpace(input.Project),
		Status:         StatusTodo,
		Priority:       normalizePriority(input.Priority),
		Tags:           normalizeUniqueStrings(input.Tags),
		DueAt:          input.DueAt,
		Tools:          normalizeStrings(input.Tools),
		Learning:       strings.TrimSpace(input.Learning),
		SourceType:     input.SourceType,
		SourceID:       input.SourceID,
		SourceTitle:    input.SourceTitle,
		SourceURL:      input.SourceURL,
		IdempotencyKey: strings.TrimSpace(input.IdempotencyKey),
		Version:        1,
		CreatedAt:      s.now(),
	}
	return s.repository.CreateTask(ctx, task)
}

func (s *Service) CreateTasks(ctx context.Context, input BatchCreateInput) ([]Task, error) {
	if s.repository == nil || input.UserID <= 0 || len(input.Tasks) == 0 || len(input.Tasks) > 20 {
		return nil, ErrInvalidTaskBatch
	}
	created := make([]Task, 0, len(input.Tasks))
	for _, item := range input.Tasks {
		item.UserID = input.UserID
		item.SourceType = strings.TrimSpace(item.SourceType)
		item.SourceTitle = strings.TrimSpace(item.SourceTitle)
		item.SourceURL = strings.TrimSpace(item.SourceURL)
		if !validCreateInput(item) {
			return nil, ErrInvalidTaskBatch
		}
		now := s.now()
		created = append(created, Task{UserID: input.UserID, Title: strings.TrimSpace(item.Title), Description: strings.TrimSpace(item.Description), Assignee: strings.TrimSpace(item.Assignee), Project: strings.TrimSpace(item.Project), Status: StatusTodo, Priority: normalizePriority(item.Priority), Tags: normalizeUniqueStrings(item.Tags), DueAt: item.DueAt, Tools: normalizeStrings(item.Tools), Learning: strings.TrimSpace(item.Learning), SourceType: item.SourceType, SourceID: item.SourceID, SourceTitle: item.SourceTitle, SourceURL: item.SourceURL, IdempotencyKey: strings.TrimSpace(item.IdempotencyKey), Version: 1, CreatedAt: now})
	}
	return s.repository.CreateTasks(ctx, created)
}

func validTaskSource(sourceType string, sourceID *int64, sourceTitle, sourceURL string) bool {
	if sourceType == "" {
		return sourceID == nil && strings.TrimSpace(sourceTitle) == "" && strings.TrimSpace(sourceURL) == ""
	}
	switch sourceType {
	case SourceAnalysisSession, SourceProjectMatch, SourceSandboxSession, SourceCompetitorScan,
		SourceCompetitorMonitoring, SourceEnterpriseDiagnosis, SourceLeadTask, SourceCRMCustomer,
		SourceLearningDiagnosis, SourceGrowthModel:
	case SourceCopilotMessage:
	default:
		return false
	}
	if sourceID != nil && *sourceID <= 0 {
		return false
	}
	sourceTitle = strings.TrimSpace(sourceTitle)
	sourceURL = strings.TrimSpace(sourceURL)
	return sourceTitle != "" && len([]rune(sourceTitle)) <= 200 &&
		len(sourceURL) <= 500 && strings.HasPrefix(sourceURL, "/") && !strings.HasPrefix(sourceURL, "//")
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
	return TaskPage{Tasks: tasks, Total: total, Limit: filters.Limit, Offset: filters.Offset, Sort: filters.Sort, Group: filters.Group}, nil
}

func normalizeListFilters(filters ListFilters) ListFilters {
	filters.Status = strings.TrimSpace(filters.Status)
	filters.Project = strings.TrimSpace(filters.Project)
	filters.Priority = strings.TrimSpace(filters.Priority)
	filters.Tag = strings.TrimSpace(filters.Tag)
	filters.Query = strings.TrimSpace(filters.Query)
	filters.Sort = strings.TrimSpace(filters.Sort)
	filters.Group = strings.TrimSpace(filters.Group)
	if filters.Status != "" {
		filters.Status = normalizeStatus(filters.Status)
	}
	if filters.Priority != "" {
		filters.Priority = normalizePriority(filters.Priority)
	}
	if filters.Sort == "" {
		filters.Sort = TaskSortCreated
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
	current, err := s.repository.GetTask(ctx, userID, id)
	if err != nil {
		return Task{}, err
	}
	if update.Version != nil && *update.Version != current.Version {
		return Task{}, ErrTaskVersionConflict
	}
	if update.Status != nil {
		status := normalizeStatus(*update.Status)
		if !validTaskStatusTransition(current.Status, status) {
			return Task{}, ErrInvalidTaskStatusTransition
		}
		update.Status = &status
	}
	if update.Progress != nil && (*update.Progress < 0 || *update.Progress > 100) {
		return Task{}, ErrInvalidTaskProgress
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

func (s *Service) RestoreTask(ctx context.Context, userID, id int64) error {
	if s.repository == nil {
		return ErrServiceNotReady
	}
	return s.repository.RestoreTask(ctx, userID, id)
}

func (s *Service) BatchUpdateTaskStatus(ctx context.Context, userID int64, ids []int64, status string) (int, error) {
	if s.repository == nil {
		return 0, ErrServiceNotReady
	}
	ids, ok := normalizeBatchTaskIDs(ids)
	if !ok || !validStatus(status) {
		return 0, ErrInvalidTaskBatch
	}
	return s.repository.BatchUpdateTaskStatus(ctx, userID, ids, status)
}

func (s *Service) BatchDeleteTasks(ctx context.Context, userID int64, ids []int64) (int, error) {
	if s.repository == nil {
		return 0, ErrServiceNotReady
	}
	ids, ok := normalizeBatchTaskIDs(ids)
	if !ok {
		return 0, ErrInvalidTaskBatch
	}
	return s.repository.BatchDeleteTasks(ctx, userID, ids)
}

func (s *Service) ListTaskActivities(ctx context.Context, userID, taskID int64, limit, offset int) ([]TaskActivity, int, error) {
	if s.repository == nil {
		return nil, 0, ErrServiceNotReady
	}
	if _, err := s.repository.GetTask(ctx, userID, taskID); err != nil {
		return nil, 0, err
	}
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	activities, err := s.repository.ListTaskActivities(ctx, userID, taskID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	total, err := s.repository.CountTaskActivities(ctx, userID, taskID)
	if err != nil {
		return nil, 0, err
	}
	return activities, total, nil
}

func normalizeBatchTaskIDs(ids []int64) ([]int64, bool) {
	if len(ids) == 0 || len(ids) > 100 {
		return nil, false
	}
	normalized := make([]int64, 0, len(ids))
	seen := make(map[int64]struct{}, len(ids))
	for _, id := range ids {
		if id <= 0 {
			return nil, false
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		normalized = append(normalized, id)
	}
	return normalized, true
}

func normalizeStatus(status string) string {
	switch status {
	case StatusInProgress, StatusReview, StatusBlocked, StatusCompleted, StatusCancelled, StatusReminder:
		return status
	default:
		return StatusTodo
	}
}

func validTaskStatusTransition(from, to string) bool {
	if from == to {
		return true
	}
	if to == StatusCompleted || to == StatusCancelled {
		return !terminalTaskStatus(from)
	}
	switch from {
	case StatusTodo:
		return to == StatusInProgress || to == StatusReminder
	case StatusInProgress:
		return to == StatusTodo || to == StatusReview || to == StatusBlocked || to == StatusReminder
	case StatusReview:
		return to == StatusInProgress || to == StatusBlocked
	case StatusBlocked:
		return to == StatusTodo || to == StatusInProgress
	case StatusReminder:
		return to == StatusTodo || to == StatusInProgress || to == StatusReview || to == StatusBlocked
	case StatusCompleted:
		return to == StatusTodo || to == StatusInProgress
	case StatusCancelled:
		return to == StatusTodo
	default:
		return false
	}
}

func terminalTaskStatus(status string) bool {
	return status == StatusCompleted || status == StatusCancelled
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
