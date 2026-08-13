package sandbox

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/zzm/opcv2/internal/ai"
)

type V2FollowUpRepository interface {
	V2Repository
	CreateV2FollowUp(context.Context, V2FollowUp) (V2FollowUp, error)
	ListV2FollowUps(context.Context, int64, int64) ([]V2FollowUp, error)
}
type V2FollowUpApplication interface {
	ListV2FollowUps(context.Context, int64, int64) ([]V2FollowUp, error)
	AskV2Role(context.Context, AskV2RoleInput) (V2FollowUp, error)
}

func (s *Service) ListV2FollowUps(ctx context.Context, userID, runID int64) ([]V2FollowUp, error) {
	repository, ok := s.repository.(V2FollowUpRepository)
	if !ok {
		return nil, ErrServiceNotReady
	}
	if _, err := repository.GetV2Run(ctx, userID, runID); err != nil {
		return nil, err
	}
	return repository.ListV2FollowUps(ctx, userID, runID)
}

func (s *Service) AskV2Role(ctx context.Context, input AskV2RoleInput) (V2FollowUp, error) {
	repository, ok := s.repository.(V2FollowUpRepository)
	if !ok || s.generator == nil {
		return V2FollowUp{}, ErrServiceNotReady
	}
	input.RoleCode = strings.TrimSpace(input.RoleCode)
	input.Question = strings.TrimSpace(input.Question)
	if input.UserID <= 0 || input.RunID <= 0 || input.RoleCode == "" || input.Question == "" || len([]rune(input.Question)) > 2000 {
		return V2FollowUp{}, ErrV2InvalidRequest
	}
	run, err := repository.GetV2Run(ctx, input.UserID, input.RunID)
	if err != nil {
		return V2FollowUp{}, err
	}
	if !isV2TerminalStatus(run.Status) {
		return V2FollowUp{}, ErrV2InvalidRequest
	}
	var selected *V2RunRole
	for i := range run.RunRoles {
		if run.RunRoles[i].RoleCode == input.RoleCode {
			selected = &run.RunRoles[i]
			break
		}
	}
	if selected == nil || selected.Status != "done" || selected.Output == nil {
		return V2FollowUp{}, ErrV2InvalidRequest
	}
	prompt, _ := json.Marshal(map[string]any{"product": run.Product, "context": run.Context, "assumptions": run.Assumptions, "role_code": selected.RoleCode, "role_profile": selected.SystemPrompt, "role_output": selected.Output, "question": input.Question})
	result, err := s.generator.GenerateJSON(ctx, ai.GenerateJSONRequest{UserID: input.UserID, Feature: "sandbox.follow_up_v2", PromptVersion: "sandbox_follow_up_v2", SystemPrompt: "你是商业沙盘指定角色追问助手。只根据给定角色配置和该角色本轮输出回答，不得提及或推断其他角色。只返回 JSON：{\"answer\":\"...\"}。", UserPrompt: string(prompt), SchemaName: "sandbox_v2_follow_up", RepairAttempts: 1, Validate: validateRoleAnswerJSON})
	if err != nil {
		return V2FollowUp{}, fmt.Errorf("%w: %v", ErrInvalidAIResult, err)
	}
	var answer roleAnswer
	if err := json.Unmarshal(result.Content, &answer); err != nil || strings.TrimSpace(answer.Answer) == "" {
		return V2FollowUp{}, ErrInvalidAIResult
	}
	return repository.CreateV2FollowUp(ctx, V2FollowUp{RunID: input.RunID, UserID: input.UserID, RoleCode: input.RoleCode, Question: input.Question, Answer: strings.TrimSpace(answer.Answer), InputContextHash: selected.InputHash, CreatedAt: s.now().UTC()})
}
