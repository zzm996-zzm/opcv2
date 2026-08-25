package copilot

import (
	"context"
	"fmt"
	"strings"

	"github.com/zzm/opcv2/internal/projects"
	"github.com/zzm/opcv2/internal/tasks"
)

type TaskCreator interface {
	CreateTask(ctx context.Context, input tasks.CreateInput) (tasks.Task, error)
}

type ProjectMatcher interface {
	CreateMatch(ctx context.Context, input projects.MatchInput) (projects.MatchResult, error)
}

type ProjectCreator interface {
	CreateUserProject(ctx context.Context, input projects.CreateUserProjectInput) (projects.UserProject, error)
}

type ToolRegistry struct {
	tasks    TaskCreator
	projects ProjectMatcher
	creator  ProjectCreator
}

func NewToolRegistry(taskCreator TaskCreator, projectMatcher ProjectMatcher, projectCreators ...ProjectCreator) *ToolRegistry {
	registry := &ToolRegistry{tasks: taskCreator, projects: projectMatcher}
	if len(projectCreators) > 0 {
		registry.creator = projectCreators[0]
	}
	return registry
}

func (r *ToolRegistry) Execute(ctx context.Context, userID, sourceMessageID int64, call ToolCall) (ToolExecutionResult, error) {
	call = normalizeToolCall(call)
	switch call.Tool {
	case ToolCreateTask:
		if r == nil || r.tasks == nil {
			return ToolExecutionResult{}, ErrToolNotAvailable
		}
		sourceID := sourceMessageID
		item, err := r.tasks.CreateTask(ctx, tasks.CreateInput{
			UserID: userID, Title: call.Arguments.Title, Description: call.Arguments.Description,
			Priority: call.Arguments.Priority, Tags: call.Arguments.Tags,
			SourceType: tasks.SourceCopilotMessage, SourceID: &sourceID,
			SourceTitle: "Copilot 代执行", SourceURL: "/copilot",
			IdempotencyKey: fmt.Sprintf("copilot-tool-task-%d", sourceMessageID),
		})
		if err != nil {
			return ToolExecutionResult{}, err
		}
		return ToolExecutionResult{
			Tool: call.Tool, Status: "completed", EntityID: item.ID, Title: item.Title,
			URL: "/tasks", Message: fmt.Sprintf("已创建任务：%s", item.Title),
		}, nil
	case ToolProjectMatch:
		if r == nil || r.projects == nil {
			return ToolExecutionResult{}, ErrToolNotAvailable
		}
		result, err := r.projects.CreateMatch(ctx, projects.MatchInput{UserID: userID, Intent: call.Arguments.Intent})
		if err != nil {
			return ToolExecutionResult{}, err
		}
		message := "已完成项目匹配"
		if result.Status == projects.StatusNeedsInput {
			message = fmt.Sprintf("已发起项目匹配，还需要补充 %d 项信息", len(result.Questions))
		}
		return ToolExecutionResult{
			Tool: call.Tool, Status: result.Status, EntityID: result.SessionID,
			Title: strings.TrimSpace(call.Arguments.Intent), URL: fmt.Sprintf("/projects/matches/%d", result.SessionID), Message: message,
		}, nil
	case ToolCreateProject:
		if r == nil || r.creator == nil {
			return ToolExecutionResult{}, ErrToolNotAvailable
		}
		item, err := r.creator.CreateUserProject(ctx, projects.CreateUserProjectInput{
			UserID: userID, Name: call.Arguments.Title, Description: call.Arguments.Description,
			SourceType: "copilot_message", SourceID: &sourceMessageID,
			IdempotencyKey: fmt.Sprintf("copilot-tool-project-%d", sourceMessageID),
		})
		if err != nil {
			return ToolExecutionResult{}, err
		}
		return ToolExecutionResult{
			Tool: call.Tool, Status: "completed", EntityID: item.ID, Title: item.Name,
			URL: "/projects/mine", Message: fmt.Sprintf("已保存项目草稿：%s", item.Name),
		}, nil
	default:
		return ToolExecutionResult{}, ErrToolNotAvailable
	}
}
