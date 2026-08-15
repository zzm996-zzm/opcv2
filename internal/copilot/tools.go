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

type ToolRegistry struct {
	tasks    TaskCreator
	projects ProjectMatcher
}

func NewToolRegistry(taskCreator TaskCreator, projectMatcher ProjectMatcher) *ToolRegistry {
	return &ToolRegistry{tasks: taskCreator, projects: projectMatcher}
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
	default:
		return ToolExecutionResult{}, ErrToolNotAvailable
	}
}
