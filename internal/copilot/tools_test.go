package copilot

import (
	"context"
	"testing"

	"github.com/zzm/opcv2/internal/projects"
	"github.com/zzm/opcv2/internal/tasks"
)

type fakeTaskCreator struct{ input tasks.CreateInput }

func (c *fakeTaskCreator) CreateTask(_ context.Context, input tasks.CreateInput) (tasks.Task, error) {
	c.input = input
	return tasks.Task{ID: 81, Title: input.Title}, nil
}

type fakeProjectMatcher struct{ input projects.MatchInput }

func (m *fakeProjectMatcher) CreateMatch(_ context.Context, input projects.MatchInput) (projects.MatchResult, error) {
	m.input = input
	return projects.MatchResult{SessionID: 91, Status: projects.StatusNeedsInput, Questions: []projects.Question{{Key: "budget", Text: "预算"}}}, nil
}

type fakeProjectCreator struct {
	input projects.CreateUserProjectInput
}

func (c *fakeProjectCreator) CreateUserProject(_ context.Context, input projects.CreateUserProjectInput) (projects.UserProject, error) {
	c.input = input
	return projects.UserProject{ID: 101, Name: input.Name, Status: "draft"}, nil
}

func TestToolRegistryCreatesAuditedTask(t *testing.T) {
	creator := &fakeTaskCreator{}
	registry := NewToolRegistry(creator, nil)

	result, err := registry.Execute(context.Background(), 42, 7, ToolCall{
		Tool:      ToolCreateTask,
		Arguments: ToolArguments{Title: "访谈10位客户", Description: "记录高频问题", Priority: "high", Tags: []string{"验证"}},
	})

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if result.EntityID != 81 || result.URL != "/tasks" || creator.input.SourceType != tasks.SourceCopilotMessage || creator.input.SourceID == nil || *creator.input.SourceID != 7 {
		t.Fatalf("result/input = %+v/%+v", result, creator.input)
	}
}

func TestToolRegistryStartsProjectMatch(t *testing.T) {
	matcher := &fakeProjectMatcher{}
	registry := NewToolRegistry(nil, matcher)

	result, err := registry.Execute(context.Background(), 42, 7, ToolCall{
		Tool: ToolProjectMatch, Arguments: ToolArguments{Intent: "低预算 AI 服务项目"},
	})

	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if matcher.input.UserID != 42 || result.EntityID != 91 || result.URL != "/projects/matches/91" || result.Status != projects.StatusNeedsInput {
		t.Fatalf("result/input = %+v/%+v", result, matcher.input)
	}
}

func TestToolRegistryCreatesUserProjectDraft(t *testing.T) {
	creator := &fakeProjectCreator{}
	registry := NewToolRegistry(nil, nil, creator)

	result, err := registry.Execute(context.Background(), 42, 7, ToolCall{
		Tool: ToolCreateProject, Arguments: ToolArguments{Title: "AI 客户洞察", Description: "整理客户反馈并生成机会清单"},
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if result.EntityID != 101 || result.URL != "/projects/mine" || result.Status != "completed" || creator.input.UserID != 42 || creator.input.SourceID == nil || *creator.input.SourceID != 7 {
		t.Fatalf("result/input = %+v/%+v", result, creator.input)
	}
}
