package tasks

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
	created, err := s.repository.CreateTasks(ctx, tasks)
	if err != nil {
		return GenerateTasksResult{}, err
	}
	return GenerateTasksResult{Tasks: created}, nil
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
