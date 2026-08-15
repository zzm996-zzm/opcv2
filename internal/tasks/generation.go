package tasks

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/zzm/opcv2/internal/ai"
)

func (s *Service) GenerateTasks(ctx context.Context, input GenerateTasksInput) (GenerateTasksResult, error) {
	input.Goal = strings.TrimSpace(input.Goal)
	input.SourceType = strings.TrimSpace(input.SourceType)
	input.SourceTitle = strings.TrimSpace(input.SourceTitle)
	input.SourceURL = strings.TrimSpace(input.SourceURL)
	if s.repository == nil || s.generator == nil {
		return GenerateTasksResult{}, ErrServiceNotReady
	}
	if !validTaskSource(input.SourceType, input.SourceID, input.SourceTitle, input.SourceURL) {
		return GenerateTasksResult{}, ErrInvalidTaskSource
	}
	result, err := s.generator.GenerateJSON(ctx, ai.GenerateJSONRequest{
		UserID:         input.UserID,
		Feature:        "tasks.generate",
		PromptVersion:  "task_plan_v1",
		SystemPrompt:   taskGenerationSystemPrompt(),
		UserPrompt:     input.Goal,
		SchemaName:     "task_generation_plan",
		Validate:       validateGeneratedTaskPlanJSON,
		RepairAttempts: 1,
	})
	if err != nil {
		return GenerateTasksResult{}, fmt.Errorf("%w: %v", ErrInvalidGeneratedTasks, err)
	}
	var plan GeneratedTaskPlan
	if err := json.Unmarshal(result.Content, &plan); err != nil {
		return GenerateTasksResult{}, fmt.Errorf("%w: %v", ErrInvalidGeneratedTasks, err)
	}
	if err := validateGeneratedTaskPlan(plan); err != nil {
		return GenerateTasksResult{}, fmt.Errorf("%w: %v", ErrInvalidGeneratedTasks, err)
	}

	now := s.now()
	tasks := make([]Task, 0, len(plan.Tasks))
	for _, draft := range plan.Tasks {
		var dueAt *time.Time
		if draft.DueInDays > 0 {
			value := now.AddDate(0, 0, draft.DueInDays)
			dueAt = &value
		}
		tasks = append(tasks, Task{
			UserID:      input.UserID,
			Title:       strings.TrimSpace(draft.Title),
			Description: strings.TrimSpace(draft.Description),
			Project:     strings.TrimSpace(draft.Project),
			Status:      StatusTodo,
			Priority:    normalizePriority(draft.Priority),
			Tags:        normalizeUniqueStrings(draft.Tags),
			DueAt:       dueAt,
			Tools:       normalizeStrings(draft.Tools),
			Learning:    strings.TrimSpace(draft.Learning),
			SourceType:  input.SourceType,
			SourceID:    input.SourceID,
			SourceTitle: input.SourceTitle,
			SourceURL:   input.SourceURL,
			Version:     1,
			CreatedAt:   now,
		})
	}
	draft, err := s.repository.CreateTaskAIDraft(ctx, TaskAIDraft{
		UserID:      input.UserID,
		Goal:        input.Goal,
		SourceType:  input.SourceType,
		SourceID:    input.SourceID,
		SourceTitle: input.SourceTitle,
		SourceURL:   input.SourceURL,
		Tasks:       tasks,
		Status:      TaskAIDraftStatusDraft,
		CreatedAt:   now,
	})
	if err != nil {
		return GenerateTasksResult{}, err
	}
	return GenerateTasksResult{Draft: draft, Tasks: draft.Tasks}, nil
}

func (s *Service) GetTaskAIDraft(ctx context.Context, userID, id int64) (TaskAIDraft, error) {
	if s.repository == nil {
		return TaskAIDraft{}, ErrServiceNotReady
	}
	return s.repository.GetTaskAIDraft(ctx, userID, id)
}

func (s *Service) AdoptTaskAIDraft(ctx context.Context, userID, id int64, input AdoptTaskAIDraftInput, idempotencyKey string) ([]Task, error) {
	if s.repository == nil {
		return nil, ErrServiceNotReady
	}
	draft, err := s.repository.GetTaskAIDraft(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	if draft.Status != TaskAIDraftStatusDraft {
		return s.repository.AdoptTaskAIDraft(ctx, userID, id, nil)
	}
	if len(input.Tasks) == 0 || len(input.Tasks) > len(draft.Tasks) {
		return nil, ErrInvalidTaskAIDraft
	}

	now := s.now()
	created := make([]Task, 0, len(input.Tasks))
	seen := make(map[int]struct{}, len(input.Tasks))
	for _, item := range input.Tasks {
		if item.DraftIndex < 0 || item.DraftIndex >= len(draft.Tasks) {
			return nil, ErrInvalidTaskAIDraft
		}
		if _, exists := seen[item.DraftIndex]; exists {
			return nil, ErrInvalidTaskAIDraft
		}
		seen[item.DraftIndex] = struct{}{}
		base := draft.Tasks[item.DraftIndex]
		item.Title = strings.TrimSpace(item.Title)
		item.Description = strings.TrimSpace(item.Description)
		item.Assignee = strings.TrimSpace(item.Assignee)
		item.Project = strings.TrimSpace(item.Project)
		item.Learning = strings.TrimSpace(item.Learning)
		if item.Title == "" {
			item.Title = base.Title
		}
		if item.Project == "" {
			item.Project = base.Project
		}
		if item.Priority == "" {
			item.Priority = base.Priority
		}
		if item.Description == "" {
			item.Description = base.Description
		}
		if item.DueAt == nil {
			item.DueAt = base.DueAt
		}
		if len(item.Tags) == 0 {
			item.Tags = base.Tags
		}
		if len(item.Tools) == 0 {
			item.Tools = base.Tools
		}
		if item.Learning == "" {
			item.Learning = base.Learning
		}
		createInput := CreateInput{
			UserID: userID, Title: item.Title, Description: item.Description, Assignee: item.Assignee,
			Project: item.Project, Priority: item.Priority, Tags: item.Tags, DueAt: item.DueAt,
			Tools: item.Tools, Learning: item.Learning, SourceType: draft.SourceType, SourceID: draft.SourceID,
			SourceTitle: draft.SourceTitle, SourceURL: draft.SourceURL,
		}
		if !validCreateInput(createInput) {
			return nil, ErrInvalidTaskAIDraft
		}
		key := strings.TrimSpace(idempotencyKey)
		if key != "" {
			key += ":" + strconv.Itoa(item.DraftIndex)
		}
		created = append(created, Task{
			UserID: userID, Title: createInput.Title, Description: createInput.Description, Assignee: createInput.Assignee,
			Project: createInput.Project, Status: StatusTodo, Priority: normalizePriority(createInput.Priority),
			Tags: normalizeUniqueStrings(createInput.Tags), DueAt: createInput.DueAt, Tools: normalizeStrings(createInput.Tools),
			Learning: createInput.Learning, SourceType: createInput.SourceType, SourceID: createInput.SourceID,
			SourceTitle: createInput.SourceTitle, SourceURL: createInput.SourceURL, IdempotencyKey: key,
			Version: 1, CreatedAt: now,
		})
	}
	return s.repository.AdoptTaskAIDraft(ctx, userID, id, created)
}

func taskGenerationSystemPrompt() string {
	return "你是任务拆解助手。根据用户目标生成1到10条可执行任务，按执行顺序排列。必须只返回JSON，字段严格匹配task_generation_plan；due_in_days为0到365的整数。"
}

func validateGeneratedTaskPlanJSON(data []byte) error {
	var plan GeneratedTaskPlan
	if err := json.Unmarshal(data, &plan); err != nil {
		return err
	}
	return validateGeneratedTaskPlan(plan)
}

func validateGeneratedTaskPlan(plan GeneratedTaskPlan) error {
	if len(plan.Tasks) == 0 || len(plan.Tasks) > 10 {
		return errors.New("tasks must contain between 1 and 10 items")
	}
	for _, task := range plan.Tasks {
		if strings.TrimSpace(task.Title) == "" || len([]rune(strings.TrimSpace(task.Title))) > 100 {
			return errors.New("task title is invalid")
		}
		if strings.TrimSpace(task.Project) == "" || !validPriority(task.Priority) {
			return errors.New("task project or priority is invalid")
		}
		if len([]rune(strings.TrimSpace(task.Description))) > 1000 || !validTags(task.Tags) {
			return errors.New("task description or tags are invalid")
		}
		if task.DueInDays < 0 || task.DueInDays > 365 {
			return errors.New("task due_in_days is invalid")
		}
	}
	return nil
}
