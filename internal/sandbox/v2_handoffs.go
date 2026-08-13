package sandbox

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/zzm/opcv2/internal/tasks"
)

type SandboxTaskCreator interface {
	CreateTasks(context.Context, tasks.BatchCreateInput) ([]tasks.Task, error)
}
type V2HandoffApplication interface {
	CreateSandboxTasks(context.Context, SandboxTaskHandoffInput) (SandboxTaskHandoffResult, error)
	CreateSandboxGrowthHandoff(context.Context, int64, int64) (SandboxGrowthHandoff, error)
}

func (s *Service) CreateSandboxTasks(ctx context.Context, input SandboxTaskHandoffInput) (SandboxTaskHandoffResult, error) {
	if s.taskCreator == nil {
		return SandboxTaskHandoffResult{}, ErrServiceNotReady
	}
	if input.UserID <= 0 || input.RunID <= 0 || len(input.AdviceIndexes) == 0 || len(input.AdviceIndexes) > 20 {
		return SandboxTaskHandoffResult{}, ErrV2InvalidRequest
	}
	run, err := s.v2RepositoryGet(ctx, input.UserID, input.RunID)
	if err != nil {
		return SandboxTaskHandoffResult{}, err
	}
	if run.Report == nil || !isV2TerminalStatus(run.Status) {
		return SandboxTaskHandoffResult{}, ErrV2InvalidRequest
	}
	seen := map[int]bool{}
	create := make([]tasks.CreateInput, 0, len(input.AdviceIndexes))
	for _, index := range input.AdviceIndexes {
		if index < 0 || index >= len(run.Report.Advice) || seen[index] {
			return SandboxTaskHandoffResult{}, ErrV2InvalidRequest
		}
		seen[index] = true
		advice := run.Report.Advice[index]
		create = append(create, tasks.CreateInput{Title: advice.Action, Description: advice.Why, Project: run.Product.Name, Priority: taskPriority(advice.Priority), Tags: []string{"商业沙盘", "验证动作"}, SourceType: tasks.SourceSandboxSession, SourceID: &input.RunID, SourceTitle: run.NameOrProduct(), SourceURL: fmt.Sprintf("/sandbox-runs/%d/report", run.ID), IdempotencyKey: fmt.Sprintf("sandbox:%d:advice:%d", run.ID, index)})
	}
	created, err := s.taskCreator.CreateTasks(ctx, tasks.BatchCreateInput{UserID: input.UserID, Tasks: create})
	if err != nil {
		return SandboxTaskHandoffResult{}, err
	}
	result := SandboxTaskHandoffResult{Tasks: make([]map[string]any, 0, len(created))}
	for _, item := range created {
		if item.SourceID == nil {
			return SandboxTaskHandoffResult{}, ErrV2InvalidRequest
		}
		result.Tasks = append(result.Tasks, map[string]any{"id": item.ID, "title": item.Title, "source_type": item.SourceType, "source_id": *item.SourceID})
	}
	return result, nil
}

func (s *Service) CreateSandboxGrowthHandoff(ctx context.Context, userID, runID int64) (SandboxGrowthHandoff, error) {
	run, err := s.v2RepositoryGet(ctx, userID, runID)
	if err != nil {
		return SandboxGrowthHandoff{}, err
	}
	if run.Report == nil || !isV2TerminalStatus(run.Status) {
		return SandboxGrowthHandoff{}, ErrV2InvalidRequest
	}
	query := url.Values{}
	query.Set("sandbox_run", strconv.FormatInt(run.ID, 10))
	query.Set("name", run.NameOrProduct())
	if run.Product.PriceCents > 0 {
		query.Set("price_cents", strconv.FormatInt(run.Product.PriceCents, 10))
	}
	query.Set("channel", run.Context.Channel)
	return SandboxGrowthHandoff{URL: "/growth-calculator?" + query.Encode(), PricingCents: run.Product.PriceCents, Channel: run.Context.Channel}, nil
}

func (s *Service) v2RepositoryGet(ctx context.Context, userID, runID int64) (V2SandboxRun, error) {
	repository, err := s.v2Repository()
	if err != nil {
		return V2SandboxRun{}, err
	}
	return repository.GetV2Run(ctx, userID, runID)
}
func taskPriority(priority int) string {
	if priority <= 1 {
		return tasks.PriorityHigh
	}
	if priority >= 3 {
		return tasks.PriorityLow
	}
	return tasks.PriorityMedium
}
func (run V2SandboxRun) NameOrProduct() string {
	if strings.TrimSpace(run.Name) != "" {
		return run.Name
	}
	return run.Product.Name
}
