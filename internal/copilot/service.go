package copilot

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/zzm/opcv2/internal/ai"
	"github.com/zzm/opcv2/internal/competitor"
	"github.com/zzm/opcv2/internal/growth"
	"github.com/zzm/opcv2/internal/membership"
	"github.com/zzm/opcv2/internal/projects"
	"github.com/zzm/opcv2/internal/tasks"
)

type Repository interface {
	CreateThread(ctx context.Context, thread Thread) (Thread, error)
	ListThreads(ctx context.Context, userID int64, limit int) ([]Thread, error)
	GetThread(ctx context.Context, userID, id int64) (Thread, error)
	UpdateThreadTitle(ctx context.Context, userID, id int64, title string) (Thread, error)
	ArchiveThread(ctx context.Context, userID, id int64) error
	CreateMessage(ctx context.Context, message Message) (Message, error)
	GetMessage(ctx context.Context, userID, threadID, messageID int64) (Message, error)
	ClaimToolPreview(ctx context.Context, userID, threadID, messageID int64) (Message, error)
	UpdateMessageContentMetadata(ctx context.Context, userID, threadID, messageID int64, content string, metadata json.RawMessage) (Message, error)
	ListMessages(ctx context.Context, userID, threadID int64, limit int) ([]Message, error)
	ListMemories(ctx context.Context, userID int64, limit int) ([]Memory, error)
	UpsertMemory(ctx context.Context, memory Memory) (Memory, error)
	DeleteMemory(ctx context.Context, userID, id int64) error
	CreateFile(ctx context.Context, file File) (File, error)
	ListFiles(ctx context.Context, userID int64, limit int) ([]File, error)
	GetFilesByIDs(ctx context.Context, userID int64, ids []int64) ([]File, error)
	GetFile(ctx context.Context, userID, id int64) (File, error)
	DeleteFile(ctx context.Context, userID, id int64) error
	ListRuns(ctx context.Context, userID int64, featurePrefix string, limit int) ([]ai.Run, error)
}

type JSONGenerator interface {
	GenerateJSON(ctx context.Context, request ai.GenerateJSONRequest) (ai.GenerateJSONResult, error)
}

type TextStreamer interface {
	GenerateTextStream(ctx context.Context, request ai.GenerateTextRequest, onDelta func([]byte) error) (ai.GenerateTextResult, error)
}

type ToolExecutor interface {
	Execute(ctx context.Context, userID, sourceMessageID int64, call ToolCall) (ToolExecutionResult, error)
}

type TaskContextProvider interface {
	GetTask(ctx context.Context, userID, id int64) (tasks.Task, error)
	ListSubtasks(ctx context.Context, userID, taskID int64) ([]tasks.Subtask, error)
}

type CompetitorContextProvider interface {
	GetScan(ctx context.Context, userID, id int64) (competitor.Scan, error)
}

type GrowthContextProvider interface {
	GetModel(ctx context.Context, userID, id int64) (growth.Model, error)
	ListModels(ctx context.Context, userID int64, limit int) ([]growth.Model, error)
}

type MonitoringContextProvider interface {
	GetMonitoring(ctx context.Context, userID int64, limit int) (competitor.MonitoringSnapshot, error)
}

type ProjectContextProvider interface {
	GetProject(ctx context.Context, ref string) (projects.Project, error)
	GetProjectMatch(ctx context.Context, userID, id int64) (projects.MatchWorkflowResponse, error)
}

type QuotaConsumer interface {
	CheckAndConsume(ctx context.Context, input membership.ConsumeInput) (membership.UsageItem, error)
	RefundUsage(ctx context.Context, input membership.ConsumeInput) (membership.UsageItem, error)
}

type Option func(*Service)

type Service struct {
	repository  Repository
	generator   JSONGenerator
	streamer    TextStreamer
	tools       ToolExecutor
	taskContext TaskContextProvider
	competitor  CompetitorContextProvider
	growth      GrowthContextProvider
	monitoring  MonitoringContextProvider
	projects    ProjectContextProvider
	models      []ModelOption
	quota       QuotaConsumer
	now         func() time.Time
}

type chatAIResult struct {
	Reply            string            `json:"reply"`
	MemoryCandidates []MemoryCandidate `json:"memory_candidates,omitempty"`
}

type memoryExtractionAIResult struct {
	MemoryCandidates []MemoryCandidate `json:"memory_candidates"`
}

type modelSmokeAIResult struct {
	Reply string `json:"reply"`
}

type messageMetadata struct {
	Kind          string               `json:"kind,omitempty"`
	ReferenceIDs  []int64              `json:"reference_ids,omitempty"`
	TaskID        int64                `json:"task_id,omitempty"`
	CurrentView   string               `json:"current_view,omitempty"`
	ActiveFilters map[string]string    `json:"active_filters,omitempty"`
	ToolResult    *ToolExecutionResult `json:"tool_result,omitempty"`
	ToolPreview   *ToolPreview         `json:"tool_preview,omitempty"`
}

type competitorScanContext struct {
	ID              int64                       `json:"id"`
	Targets         []string                    `json:"targets"`
	Focus           string                      `json:"focus"`
	Status          string                      `json:"status"`
	ProgressPercent int                         `json:"progress_percent"`
	CurrentStep     string                      `json:"current_step"`
	Competitors     []competitor.Competitor     `json:"competitors,omitempty"`
	Conclusions     []competitor.Conclusion     `json:"conclusions,omitempty"`
	EvidenceSources []competitor.EvidenceSource `json:"evidence_sources,omitempty"`
}

type growthModelContext struct {
	Model growth.Model `json:"model"`
}

type monitoringContext struct {
	Watchlist []competitor.WatchItem `json:"watchlist"`
	Events    []competitor.Event     `json:"events"`
}

type projectContext struct {
	Project        *projects.Project               `json:"project,omitempty"`
	Match          *projects.MatchWorkflowResponse `json:"match,omitempty"`
	CurrentSection string                          `json:"current_section,omitempty"`
}

const (
	messageKindCompareQuestion = "compare_question"
	messageKindCompareAnswer   = "compare_answer"
	messageKindCompareSummary  = "compare_summary"
	maxFileContentBytes        = 120000
	maxReferenceFiles          = 5
	maxReferencePromptBytes    = 6000
	maxUploadFileBytes         = 10 * 1024 * 1024
	maxDOCXXMLBytes            = 4 * 1024 * 1024
	maxMemoryKeyRunes          = 80
	maxMemoryValueRunes        = 1000
)

func NewService(repository Repository, generator JSONGenerator, options ...Option) *Service {
	service := &Service{repository: repository, generator: generator, now: time.Now}
	if streamer, ok := generator.(TextStreamer); ok {
		service.streamer = streamer
	}
	for _, option := range options {
		option(service)
	}
	return service
}

func (s *Service) StreamMessage(ctx context.Context, input SendMessageInput, onEvent func(StreamEvent) error) (SendMessageResult, error) {
	if s.repository == nil || s.streamer == nil || onEvent == nil {
		return SendMessageResult{}, ErrServiceNotReady
	}
	content := strings.TrimSpace(input.Content)
	if content == "" {
		return SendMessageResult{}, ErrInvalidInput
	}
	thread, err := s.repository.GetThread(ctx, input.UserID, input.ThreadID)
	if err != nil {
		return SendMessageResult{}, err
	}
	model := strings.TrimSpace(input.Model)
	if model == "" {
		model = thread.Model
	}
	model, err = s.normalizeRequestedModel(model)
	if err != nil {
		return SendMessageResult{}, err
	}
	references, err := s.referenceFiles(ctx, input.UserID, input.ReferenceIDs)
	if err != nil {
		return SendMessageResult{}, err
	}
	taskContext, err := s.loadTaskContext(ctx, input)
	if err != nil {
		return SendMessageResult{}, err
	}
	competitorContext, err := s.loadCompetitorContext(ctx, input)
	if err != nil {
		return SendMessageResult{}, err
	}
	growthContext, err := s.loadGrowthContext(ctx, input)
	if err != nil {
		return SendMessageResult{}, err
	}
	monitoringContext, err := s.loadMonitoringContext(ctx, input)
	if err != nil {
		return SendMessageResult{}, err
	}
	projectContext, err := s.loadProjectContext(ctx, input)
	if err != nil {
		return SendMessageResult{}, err
	}
	memories, _ := s.repository.ListMemories(ctx, input.UserID, 20)
	messages, _ := s.repository.ListMessages(ctx, input.UserID, thread.ID, 12)
	quotaKey := "copilot-message-" + s.normalizeRequestID(input.RequestID, input.UserID, input.ThreadID)
	if err := s.consumeQuota(ctx, input.UserID, membership.FeatureCopilotMessages, 1, quotaKey); err != nil {
		return SendMessageResult{}, err
	}
	userMessage, err := s.repository.CreateMessage(ctx, Message{
		UserID: input.UserID, ThreadID: input.ThreadID, Role: RoleUser, Content: content,
		Status: MessageStatusCompleted, Model: model, Metadata: sendMessageMetadata(input), CreatedAt: s.now(),
	})
	if err != nil {
		s.refundQuota(ctx, input.UserID, membership.FeatureCopilotMessages, 1, quotaKey, "user-message")
		return SendMessageResult{}, err
	}
	if err := onEvent(StreamEvent{Type: StreamEventUserMessage, UserMessage: &userMessage}); err != nil {
		s.refundQuota(ctx, input.UserID, membership.FeatureCopilotMessages, 1, quotaKey, "client")
		return SendMessageResult{}, err
	}
	if toolCall := s.detectToolCall(ctx, input.UserID, content, model, taskContext); toolCall != nil && s.tools != nil {
		assistantMessage, err := s.createToolPreview(ctx, input, model, userMessage.ID, *toolCall)
		if err != nil {
			return SendMessageResult{}, err
		}
		if err := onEvent(StreamEvent{Type: StreamEventAssistantMessage, AssistantMessage: &assistantMessage}); err != nil {
			return SendMessageResult{}, err
		}
		return SendMessageResult{UserMessage: userMessage, AssistantMessage: assistantMessage}, nil
	}
	streamResult, err := s.streamer.GenerateTextStream(ctx, ai.GenerateTextRequest{
		UserID: input.UserID, Feature: "copilot.chat_stream", PromptVersion: "copilot_chat_stream_v1", Model: model,
		SystemPrompt: "你是智活 Copilot，回答要直接、可执行。请使用清晰的 Markdown，不要返回 JSON。",
		UserPrompt:   buildUserPrompt(thread, memories, messages, references, taskContext, competitorContext, growthContext, monitoringContext, projectContext, content, false),
	}, func(delta []byte) error {
		return onEvent(StreamEvent{Type: StreamEventDelta, Delta: string(delta)})
	})
	if err != nil {
		_, _ = s.repository.CreateMessage(ctx, Message{
			UserID: input.UserID, ThreadID: input.ThreadID, Role: RoleAssistant,
			Content: "AI 回复暂时不可用，请稍后重试。", Status: MessageStatusFailed,
			Model: model, ErrorCode: "stream_failed", CreatedAt: s.now(),
		})
		s.refundQuota(ctx, input.UserID, membership.FeatureCopilotMessages, 1, quotaKey, "generation")
		return SendMessageResult{}, fmt.Errorf("%w: %v", ErrInvalidAIResult, err)
	}
	assistantMessage, err := s.repository.CreateMessage(ctx, Message{
		UserID: input.UserID, ThreadID: input.ThreadID, Role: RoleAssistant, Content: strings.TrimSpace(streamResult.Content),
		Status: MessageStatusCompleted, Model: model, InputTokens: streamResult.InputTokens,
		OutputTokens: streamResult.OutputTokens, CreatedAt: s.now(),
	})
	if err != nil {
		s.refundQuota(ctx, input.UserID, membership.FeatureCopilotMessages, 1, quotaKey, "assistant-message")
		return SendMessageResult{}, err
	}
	if err := onEvent(StreamEvent{Type: StreamEventAssistantMessage, AssistantMessage: &assistantMessage}); err != nil {
		return SendMessageResult{}, err
	}
	s.extractAndSaveMemories(ctx, input.UserID, model, content, assistantMessage.Content)
	return SendMessageResult{UserMessage: userMessage, AssistantMessage: assistantMessage}, nil
}

func (s *Service) detectToolCall(ctx context.Context, userID int64, content, model string, taskContext *TaskContext) *ToolCall {
	if s.generator == nil || !likelyToolRequest(content) {
		return nil
	}
	result, err := s.generator.GenerateJSON(ctx, ai.GenerateJSONRequest{
		UserID: userID, Feature: "copilot.tool_intent", PromptVersion: "copilot_tool_intent_v1", Model: model,
		SystemPrompt: "识别用户是否明确要求代执行。只允许 create_task、project_match、none。不得从普通咨询推断执行意图。必须返回 JSON。",
		UserPrompt:   fmt.Sprintf("%s\n用户请求：%s\n返回 {\"tool\":\"create_task|project_match|none\",\"arguments\":{\"title\":\"\",\"description\":\"\",\"priority\":\"low|medium|high\",\"tags\":[],\"intent\":\"\"}}", taskContextPrompt(taskContext), content),
		SchemaName:   "copilot_tool_intent", Validate: validateToolCallJSON, RepairAttempts: 1,
	})
	if err != nil {
		return nil
	}
	var call ToolCall
	if json.Unmarshal(result.Content, &call) != nil {
		return nil
	}
	call = normalizeToolCall(call)
	if call.Tool == ToolNone {
		return nil
	}
	return &call
}

func likelyToolRequest(content string) bool {
	content = strings.ToLower(strings.TrimSpace(content))
	keywords := []string{"创建任务", "新建任务", "添加任务", "建个任务", "项目匹配", "匹配项目", "帮我匹配", "create task", "match project"}
	for _, keyword := range keywords {
		if strings.Contains(content, keyword) {
			return true
		}
	}
	return false
}

func validateToolCallJSON(data []byte) error {
	var call ToolCall
	if err := json.Unmarshal(data, &call); err != nil {
		return err
	}
	call = normalizeToolCall(call)
	switch call.Tool {
	case ToolNone:
		return nil
	case ToolCreateTask:
		if call.Arguments.Title == "" {
			return errors.New("task title is required")
		}
	case ToolProjectMatch:
		if call.Arguments.Intent == "" {
			return errors.New("project match intent is required")
		}
	default:
		return errors.New("tool is not allowed")
	}
	return nil
}

func normalizeToolCall(call ToolCall) ToolCall {
	call.Tool = strings.ToLower(strings.TrimSpace(call.Tool))
	call.Arguments.Title = strings.TrimSpace(call.Arguments.Title)
	call.Arguments.Description = strings.TrimSpace(call.Arguments.Description)
	call.Arguments.Priority = strings.ToLower(strings.TrimSpace(call.Arguments.Priority))
	call.Arguments.Intent = strings.TrimSpace(call.Arguments.Intent)
	return call
}

func toolResultMetadata(result ToolExecutionResult) []byte {
	data, _ := json.Marshal(messageMetadata{ToolResult: &result})
	return data
}

func (s *Service) createToolPreview(ctx context.Context, input SendMessageInput, model string, sourceMessageID int64, call ToolCall) (Message, error) {
	preview := ToolPreview{
		ID:              fmt.Sprintf("tool-%d", sourceMessageID),
		SourceMessageID: sourceMessageID,
		Call:            normalizeToolCall(call),
		Status:          ToolPreviewPending,
		ExpiresAt:       s.now().Add(30 * time.Minute),
	}
	metadata, _ := json.Marshal(messageMetadata{ToolPreview: &preview})
	return s.repository.CreateMessage(ctx, Message{
		UserID: input.UserID, ThreadID: input.ThreadID, Role: RoleAssistant,
		Content: toolPreviewContent(preview.Call), Status: MessageStatusCompleted,
		Model: model, Metadata: metadata, CreatedAt: s.now(),
	})
}

func toolPreviewContent(call ToolCall) string {
	switch call.Tool {
	case ToolCreateTask:
		return fmt.Sprintf("我准备创建任务“%s”。确认后会进入正式任务统计，并可能向相关成员发送通知。", call.Arguments.Title)
	case ToolProjectMatch:
		return fmt.Sprintf("我准备发起项目匹配“%s”。确认后会创建一条正式匹配会话。", call.Arguments.Intent)
	default:
		return "我准备执行一项操作，请确认。"
	}
}

func (s *Service) ConfirmTool(ctx context.Context, input ToolConfirmationInput) (ToolConfirmationResult, error) {
	if s.repository == nil || s.tools == nil || input.UserID <= 0 || input.ThreadID <= 0 || input.MessageID <= 0 {
		return ToolConfirmationResult{}, ErrInvalidInput
	}
	decision := strings.ToLower(strings.TrimSpace(input.Decision))
	if decision != "confirm" && decision != "cancel" {
		return ToolConfirmationResult{}, ErrInvalidInput
	}
	message, err := s.repository.GetMessage(ctx, input.UserID, input.ThreadID, input.MessageID)
	if err != nil {
		return ToolConfirmationResult{}, err
	}
	metadata, preview, err := toolPreviewFromMessage(message)
	if err != nil {
		return ToolConfirmationResult{}, err
	}
	if !s.now().Before(preview.ExpiresAt) && preview.Status == ToolPreviewPending {
		claimed, claimErr := s.repository.ClaimToolPreview(ctx, input.UserID, input.ThreadID, input.MessageID)
		if claimErr != nil {
			return ToolConfirmationResult{}, claimErr
		}
		_, preview, err = toolPreviewFromMessage(claimed)
		if err != nil {
			return ToolConfirmationResult{}, err
		}
		preview.Status = ToolPreviewExpired
		data, _ := json.Marshal(messageMetadata{ToolPreview: preview})
		updated, updateErr := s.repository.UpdateMessageContentMetadata(ctx, input.UserID, input.ThreadID, input.MessageID, "这条执行预览已过期，请重新发起请求。", data)
		if updateErr != nil {
			return ToolConfirmationResult{}, updateErr
		}
		return ToolConfirmationResult{Message: updated}, ErrToolPreviewExpired
	}
	if decision == "cancel" {
		if preview.Status == ToolPreviewCancelled || preview.Status == ToolPreviewExpired {
			return ToolConfirmationResult{Message: message}, nil
		}
		if preview.Status != ToolPreviewPending {
			return ToolConfirmationResult{}, ErrToolPreviewConflict
		}
		claimed, claimErr := s.repository.ClaimToolPreview(ctx, input.UserID, input.ThreadID, input.MessageID)
		if claimErr != nil {
			return ToolConfirmationResult{}, claimErr
		}
		_, preview, err = toolPreviewFromMessage(claimed)
		if err != nil {
			return ToolConfirmationResult{}, err
		}
		preview.Status = ToolPreviewCancelled
		data, _ := json.Marshal(messageMetadata{ToolPreview: preview})
		updated, err := s.repository.UpdateMessageContentMetadata(ctx, input.UserID, input.ThreadID, input.MessageID, "已取消这项操作。", data)
		if err != nil {
			return ToolConfirmationResult{}, err
		}
		return ToolConfirmationResult{Message: updated}, nil
	}
	if preview.Status == ToolPreviewConfirmed && metadata.ToolResult != nil {
		return ToolConfirmationResult{Message: message, ToolResult: metadata.ToolResult}, nil
	}
	if preview.Status != ToolPreviewPending {
		return ToolConfirmationResult{}, ErrToolPreviewConflict
	}
	claimed, err := s.repository.ClaimToolPreview(ctx, input.UserID, input.ThreadID, input.MessageID)
	if err != nil {
		return ToolConfirmationResult{}, err
	}
	_, preview, err = toolPreviewFromMessage(claimed)
	if err != nil {
		return ToolConfirmationResult{}, err
	}
	toolResult, err := s.tools.Execute(ctx, input.UserID, preview.SourceMessageID, preview.Call)
	if err != nil {
		preview.Status = ToolPreviewPending
		preview.Error = "上次执行失败，可以重试。"
		failedMetadata, _ := json.Marshal(messageMetadata{ToolPreview: preview})
		_, _ = s.repository.UpdateMessageContentMetadata(ctx, input.UserID, input.ThreadID, input.MessageID, toolPreviewContent(preview.Call), failedMetadata)
		return ToolConfirmationResult{}, err
	}
	preview.Status = ToolPreviewConfirmed
	preview.Error = ""
	metadata = messageMetadata{ToolPreview: preview, ToolResult: &toolResult}
	data, _ := json.Marshal(metadata)
	updated, err := s.repository.UpdateMessageContentMetadata(ctx, input.UserID, input.ThreadID, input.MessageID, toolResult.Message, data)
	if err != nil {
		return ToolConfirmationResult{}, err
	}
	return ToolConfirmationResult{Message: updated, ToolResult: &toolResult}, nil
}

func toolPreviewFromMessage(message Message) (messageMetadata, *ToolPreview, error) {
	var metadata messageMetadata
	if err := json.Unmarshal(message.Metadata, &metadata); err != nil || metadata.ToolPreview == nil {
		return messageMetadata{}, nil, ErrToolPreviewNotFound
	}
	return metadata, metadata.ToolPreview, nil
}

func (s *Service) loadTaskContext(ctx context.Context, input SendMessageInput) (*TaskContext, error) {
	view := strings.TrimSpace(input.CurrentView)
	filters := normalizeContextFilters(input.ActiveFilters)
	if input.TaskID <= 0 {
		if view == "" && len(filters) == 0 {
			return nil, nil
		}
		return &TaskContext{CurrentView: view, ActiveFilters: filters}, nil
	}
	if s.taskContext == nil {
		return nil, ErrServiceNotReady
	}
	task, err := s.taskContext.GetTask(ctx, input.UserID, input.TaskID)
	if err != nil {
		return nil, err
	}
	items, err := s.taskContext.ListSubtasks(ctx, input.UserID, input.TaskID)
	if err != nil {
		return nil, err
	}
	context := &TaskContext{
		ID: task.ID, Title: task.Title, Description: task.Description,
		Assignee: task.Assignee, Project: task.Project, Status: task.Status,
		Priority: task.Priority, Progress: task.Progress, DueAt: task.DueAt,
		CurrentView: view, ActiveFilters: filters, TotalSubtasks: len(items),
		Subtasks: make([]TaskSubtaskContext, 0, len(items)),
	}
	for _, item := range items {
		if item.Completed {
			context.CompletedSubtasks++
		}
		context.Subtasks = append(context.Subtasks, TaskSubtaskContext{
			ID: item.ID, ParentSubtaskID: item.ParentSubtaskID, Title: item.Title,
			Assignee: item.Assignee, DueAt: item.DueAt, Completed: item.Completed,
		})
	}
	return context, nil
}

func (s *Service) loadCompetitorContext(ctx context.Context, input SendMessageInput) (*competitorScanContext, error) {
	filters := normalizeContextFilters(input.ActiveFilters)
	scanIDValue := strings.TrimSpace(filters["scan_id"])
	if scanIDValue == "" {
		return nil, nil
	}
	scanID, err := strconv.ParseInt(scanIDValue, 10, 64)
	if err != nil || scanID <= 0 {
		return nil, ErrInvalidInput
	}
	if s.competitor == nil {
		return nil, ErrServiceNotReady
	}
	scan, err := s.competitor.GetScan(ctx, input.UserID, scanID)
	if err != nil {
		return nil, err
	}
	return &competitorScanContext{
		ID: scan.ID, Targets: scan.Targets, Focus: scan.Focus, Status: scan.Status,
		ProgressPercent: scan.ProgressPercent, CurrentStep: scan.CurrentStep,
		Competitors: scan.Competitors, Conclusions: scan.Conclusions, EvidenceSources: scan.EvidenceSources,
	}, nil
}

func (s *Service) loadGrowthContext(ctx context.Context, input SendMessageInput) (*growthModelContext, error) {
	filters := normalizeContextFilters(input.ActiveFilters)
	if filters["module"] != "growth" {
		return nil, nil
	}
	if s.growth == nil {
		return nil, ErrServiceNotReady
	}
	modelIDValue := strings.TrimSpace(filters["model_id"])
	if modelIDValue == "" {
		models, err := s.growth.ListModels(ctx, input.UserID, 1)
		if err != nil {
			return nil, err
		}
		if len(models) == 0 {
			return nil, nil
		}
		return &growthModelContext{Model: models[0]}, nil
	}
	modelID, err := strconv.ParseInt(modelIDValue, 10, 64)
	if err != nil || modelID <= 0 {
		return nil, ErrInvalidInput
	}
	model, err := s.growth.GetModel(ctx, input.UserID, modelID)
	if err != nil {
		return nil, err
	}
	return &growthModelContext{Model: model}, nil
}

func (s *Service) loadMonitoringContext(ctx context.Context, input SendMessageInput) (*monitoringContext, error) {
	filters := normalizeContextFilters(input.ActiveFilters)
	if filters["module"] != "monitoring" {
		return nil, nil
	}
	if s.monitoring == nil {
		return nil, ErrServiceNotReady
	}
	snapshot, err := s.monitoring.GetMonitoring(ctx, input.UserID, 20)
	if err != nil {
		return nil, err
	}
	return &monitoringContext{Watchlist: snapshot.Watchlist, Events: snapshot.Events}, nil
}

func (s *Service) loadProjectContext(ctx context.Context, input SendMessageInput) (*projectContext, error) {
	filters := normalizeContextFilters(input.ActiveFilters)
	if filters["module"] != "projects" {
		return nil, nil
	}
	projectRef := strings.TrimSpace(filters["project_ref"])
	matchIDValue := strings.TrimSpace(filters["match_id"])
	if projectRef == "" && matchIDValue == "" {
		return nil, nil
	}
	if s.projects == nil {
		return nil, ErrServiceNotReady
	}
	context := &projectContext{CurrentSection: strings.TrimSpace(filters["section"])}
	if projectRef != "" {
		project, err := s.projects.GetProject(ctx, projectRef)
		if err != nil {
			return nil, err
		}
		context.Project = &project
	}
	if matchIDValue != "" {
		matchID, err := strconv.ParseInt(matchIDValue, 10, 64)
		if err != nil || matchID <= 0 {
			return nil, ErrInvalidInput
		}
		match, err := s.projects.GetProjectMatch(ctx, input.UserID, matchID)
		if err != nil {
			return nil, err
		}
		context.Match = &match
	}
	return context, nil
}

func NewServiceWithModels(repository Repository, generator JSONGenerator, models []ModelOption, options ...Option) *Service {
	service := NewService(repository, generator, options...)
	service.models = normalizeModelOptions(models)
	return service
}

func WithQuotaConsumer(quota QuotaConsumer) Option {
	return func(service *Service) {
		service.quota = quota
	}
}

func WithToolExecutor(executor ToolExecutor) Option {
	return func(service *Service) {
		service.tools = executor
	}
}

func WithTaskContextProvider(provider TaskContextProvider) Option {
	return func(service *Service) {
		service.taskContext = provider
	}
}

func WithCompetitorContextProvider(provider CompetitorContextProvider) Option {
	return func(service *Service) {
		service.competitor = provider
	}
}

func WithGrowthContextProvider(provider GrowthContextProvider) Option {
	return func(service *Service) {
		service.growth = provider
	}
}

func WithMonitoringContextProvider(provider MonitoringContextProvider) Option {
	return func(service *Service) {
		service.monitoring = provider
	}
}

func WithProjectContextProvider(provider ProjectContextProvider) Option {
	return func(service *Service) {
		service.projects = provider
	}
}

func (s *Service) ListModels(_ context.Context) ([]ModelOption, error) {
	if len(s.models) == 0 {
		return []ModelOption{{Name: "Development", Value: "development-model", Provider: "development", IsDefault: true}}, nil
	}
	return append([]ModelOption(nil), s.models...), nil
}

func (s *Service) ListAIRuns(ctx context.Context, userID int64, limit int) ([]AIRun, error) {
	if s.repository == nil {
		return nil, ErrServiceNotReady
	}
	runs, err := s.repository.ListRuns(ctx, userID, "copilot.", normalizeLimit(limit, 20))
	if err != nil {
		return nil, err
	}
	result := make([]AIRun, 0, len(runs))
	for _, run := range runs {
		result = append(result, AIRun{
			ID:            run.ID,
			UserID:        run.UserID,
			Feature:       run.Feature,
			PromptVersion: run.PromptVersion,
			Provider:      run.Provider,
			Model:         run.Model,
			Status:        run.Status,
			ErrorCode:     run.ErrorCode,
			ErrorMessage:  run.ErrorMessage,
			InputTokens:   run.InputTokens,
			OutputTokens:  run.OutputTokens,
			LatencyMS:     run.LatencyMS,
			CreatedAt:     run.CreatedAt,
			UpdatedAt:     run.UpdatedAt,
		})
	}
	return result, nil
}

func (s *Service) CreateThread(ctx context.Context, input CreateThreadInput) (Thread, error) {
	if s.repository == nil {
		return Thread{}, ErrServiceNotReady
	}
	title := strings.TrimSpace(input.Title)
	if title == "" {
		title = "新会话"
	}
	mode := strings.TrimSpace(input.Mode)
	if mode == "" {
		mode = ModeChat
	}
	model, err := s.normalizeRequestedModel(input.Model)
	if err != nil {
		return Thread{}, err
	}
	thread := Thread{
		UserID:    input.UserID,
		Title:     title,
		Mode:      mode,
		Model:     model,
		CreatedAt: s.now(),
	}
	return s.repository.CreateThread(ctx, thread)
}

func (s *Service) ListThreads(ctx context.Context, userID int64, limit int) ([]Thread, error) {
	if s.repository == nil {
		return nil, ErrServiceNotReady
	}
	return s.repository.ListThreads(ctx, userID, normalizeLimit(limit, 20))
}

func (s *Service) GetThread(ctx context.Context, userID, id int64) (Thread, error) {
	if s.repository == nil {
		return Thread{}, ErrServiceNotReady
	}
	return s.repository.GetThread(ctx, userID, id)
}

func (s *Service) RenameThread(ctx context.Context, userID, id int64, title string) (Thread, error) {
	if s.repository == nil {
		return Thread{}, ErrServiceNotReady
	}
	title = strings.TrimSpace(title)
	if title == "" {
		return Thread{}, ErrInvalidInput
	}
	return s.repository.UpdateThreadTitle(ctx, userID, id, title)
}

func (s *Service) ArchiveThread(ctx context.Context, userID, id int64) error {
	if s.repository == nil {
		return ErrServiceNotReady
	}
	return s.repository.ArchiveThread(ctx, userID, id)
}

func (s *Service) ListMessages(ctx context.Context, userID, threadID int64, limit int) ([]Message, error) {
	if s.repository == nil {
		return nil, ErrServiceNotReady
	}
	if _, err := s.repository.GetThread(ctx, userID, threadID); err != nil {
		return nil, err
	}
	return s.repository.ListMessages(ctx, userID, threadID, normalizeLimit(limit, 50))
}

func (s *Service) SendMessage(ctx context.Context, input SendMessageInput) (SendMessageResult, error) {
	if s.repository == nil || s.generator == nil {
		return SendMessageResult{}, ErrServiceNotReady
	}
	content := strings.TrimSpace(input.Content)
	if content == "" {
		return SendMessageResult{}, ErrInvalidInput
	}
	thread, err := s.repository.GetThread(ctx, input.UserID, input.ThreadID)
	if err != nil {
		return SendMessageResult{}, err
	}
	model := strings.TrimSpace(input.Model)
	if model == "" {
		model = thread.Model
	}
	model, err = s.normalizeRequestedModel(model)
	if err != nil {
		return SendMessageResult{}, err
	}
	now := s.now()
	references, err := s.referenceFiles(ctx, input.UserID, input.ReferenceIDs)
	if err != nil {
		return SendMessageResult{}, err
	}
	taskContext, err := s.loadTaskContext(ctx, input)
	if err != nil {
		return SendMessageResult{}, err
	}
	competitorContext, err := s.loadCompetitorContext(ctx, input)
	if err != nil {
		return SendMessageResult{}, err
	}
	growthContext, err := s.loadGrowthContext(ctx, input)
	if err != nil {
		return SendMessageResult{}, err
	}
	monitoringContext, err := s.loadMonitoringContext(ctx, input)
	if err != nil {
		return SendMessageResult{}, err
	}
	projectContext, err := s.loadProjectContext(ctx, input)
	if err != nil {
		return SendMessageResult{}, err
	}
	quotaKey := "copilot-message-" + s.normalizeRequestID(input.RequestID, input.UserID, input.ThreadID)
	if err := s.consumeQuota(ctx, input.UserID, membership.FeatureCopilotMessages, 1, quotaKey); err != nil {
		return SendMessageResult{}, err
	}
	userMessage, err := s.repository.CreateMessage(ctx, Message{
		UserID:    input.UserID,
		ThreadID:  input.ThreadID,
		Role:      RoleUser,
		Content:   content,
		Status:    MessageStatusCompleted,
		Model:     model,
		Metadata:  sendMessageMetadata(input),
		CreatedAt: now,
	})
	if err != nil {
		s.refundQuota(ctx, input.UserID, membership.FeatureCopilotMessages, 1, quotaKey, "user-message")
		return SendMessageResult{}, err
	}
	if toolCall := s.detectToolCall(ctx, input.UserID, content, model, taskContext); toolCall != nil && s.tools != nil {
		assistantMessage, previewErr := s.createToolPreview(ctx, input, model, userMessage.ID, *toolCall)
		if previewErr != nil {
			return SendMessageResult{}, previewErr
		}
		return SendMessageResult{UserMessage: userMessage, AssistantMessage: assistantMessage}, nil
	}
	aiResult, result, err := s.generateReply(ctx, input.UserID, thread, content, model, references, taskContext, competitorContext, growthContext, monitoringContext, projectContext)
	if err != nil {
		_, _ = s.repository.CreateMessage(ctx, Message{
			UserID:    input.UserID,
			ThreadID:  input.ThreadID,
			Role:      RoleAssistant,
			Content:   "AI 回复暂时不可用，请稍后重试。",
			Status:    MessageStatusFailed,
			Model:     model,
			ErrorCode: "invalid_ai_result",
			Metadata:  sendMessageMetadata(input),
			CreatedAt: s.now(),
		})
		s.refundQuota(ctx, input.UserID, membership.FeatureCopilotMessages, 1, quotaKey, "generation")
		return SendMessageResult{}, err
	}
	assistantMessage, err := s.repository.CreateMessage(ctx, Message{
		UserID:       input.UserID,
		ThreadID:     input.ThreadID,
		Role:         RoleAssistant,
		Content:      result.Reply,
		Status:       MessageStatusCompleted,
		Model:        model,
		InputTokens:  aiResult.InputTokens,
		OutputTokens: aiResult.OutputTokens,
		CreatedAt:    s.now(),
	})
	if err != nil {
		s.refundQuota(ctx, input.UserID, membership.FeatureCopilotMessages, 1, quotaKey, "assistant-message")
		return SendMessageResult{}, err
	}
	s.saveMemoryCandidates(ctx, input.UserID, result.MemoryCandidates)
	return SendMessageResult{UserMessage: userMessage, AssistantMessage: assistantMessage}, nil
}

func (s *Service) CompareMessages(ctx context.Context, input CompareMessagesInput) (CompareMessagesResult, error) {
	if s.repository == nil || s.generator == nil {
		return CompareMessagesResult{}, ErrServiceNotReady
	}
	content := strings.TrimSpace(input.Content)
	if content == "" {
		return CompareMessagesResult{}, ErrInvalidInput
	}
	thread, err := s.repository.GetThread(ctx, input.UserID, input.ThreadID)
	if err != nil {
		return CompareMessagesResult{}, err
	}
	models, err := s.normalizeRequestedModels(input.Models)
	if err != nil {
		return CompareMessagesResult{}, err
	}
	quotaKey := "copilot-compare-" + s.normalizeRequestID(input.RequestID, input.UserID, input.ThreadID)
	if err := s.consumeQuota(ctx, input.UserID, membership.FeatureCopilotCompareCalls, len(models), quotaKey); err != nil {
		return CompareMessagesResult{}, err
	}
	now := s.now()
	userMessage, err := s.repository.CreateMessage(ctx, Message{
		UserID:    input.UserID,
		ThreadID:  input.ThreadID,
		Role:      RoleUser,
		Content:   content,
		Status:    MessageStatusCompleted,
		Model:     strings.Join(models, ","),
		Metadata:  messageKindMetadata(messageKindCompareQuestion),
		CreatedAt: now,
	})
	if err != nil {
		s.refundQuota(ctx, input.UserID, membership.FeatureCopilotCompareCalls, len(models), quotaKey, "user-message")
		return CompareMessagesResult{}, err
	}

	answers := make([]CompareAnswer, 0, len(models))
	failedCalls := 0
	for _, model := range models {
		aiResult, result, err := s.generateReply(ctx, input.UserID, thread, content, model, nil, nil, nil, nil, nil, nil)
		if err != nil {
			failedCalls++
			message, _ := s.repository.CreateMessage(ctx, Message{
				UserID:    input.UserID,
				ThreadID:  input.ThreadID,
				Role:      RoleAssistant,
				Content:   "AI 回复暂时不可用，请稍后重试。",
				Status:    MessageStatusFailed,
				Model:     model,
				ErrorCode: "invalid_ai_result",
				Metadata:  messageKindMetadata(messageKindCompareAnswer),
				CreatedAt: s.now(),
			})
			answers = append(answers, CompareAnswer{Model: model, AssistantMessage: message, ErrorCode: "invalid_ai_result"})
			continue
		}
		assistantMessage, err := s.repository.CreateMessage(ctx, Message{
			UserID:       input.UserID,
			ThreadID:     input.ThreadID,
			Role:         RoleAssistant,
			Content:      result.Reply,
			Status:       MessageStatusCompleted,
			Model:        model,
			InputTokens:  aiResult.InputTokens,
			OutputTokens: aiResult.OutputTokens,
			Metadata:     messageKindMetadata(messageKindCompareAnswer),
			CreatedAt:    s.now(),
		})
		if err != nil {
			s.refundQuota(ctx, input.UserID, membership.FeatureCopilotCompareCalls, len(models)-len(answers), quotaKey, "assistant-message")
			return CompareMessagesResult{}, err
		}
		s.saveMemoryCandidates(ctx, input.UserID, result.MemoryCandidates)
		answers = append(answers, CompareAnswer{Model: model, AssistantMessage: assistantMessage})
	}
	if failedCalls > 0 {
		s.refundQuota(ctx, input.UserID, membership.FeatureCopilotCompareCalls, failedCalls, quotaKey, "generation")
	}

	return CompareMessagesResult{UserMessage: userMessage, Answers: answers}, nil
}

func (s *Service) SummarizeComparison(ctx context.Context, input CompareSummaryInput) (CompareSummaryResult, error) {
	if s.repository == nil || s.generator == nil {
		return CompareSummaryResult{}, ErrServiceNotReady
	}
	content := strings.TrimSpace(input.Content)
	if content == "" {
		return CompareSummaryResult{}, ErrInvalidInput
	}
	thread, err := s.repository.GetThread(ctx, input.UserID, input.ThreadID)
	if err != nil {
		return CompareSummaryResult{}, err
	}
	model, err := s.normalizeRequestedModel(input.Model)
	if err != nil {
		return CompareSummaryResult{}, err
	}
	if model == "" {
		model = thread.Model
	}
	model, err = s.normalizeRequestedModel(model)
	if err != nil {
		return CompareSummaryResult{}, err
	}
	prompt, err := buildCompareSummaryPrompt(thread, content, input.Answers)
	if err != nil {
		return CompareSummaryResult{}, err
	}
	quotaKey := "copilot-summary-" + s.normalizeRequestID(input.RequestID, input.UserID, input.ThreadID)
	if err := s.consumeQuota(ctx, input.UserID, membership.FeatureCopilotMessages, 1, quotaKey); err != nil {
		return CompareSummaryResult{}, err
	}
	aiResult, result, err := s.generateSummary(ctx, input.UserID, prompt, model)
	if err != nil {
		s.refundQuota(ctx, input.UserID, membership.FeatureCopilotMessages, 1, quotaKey, "generation")
		return CompareSummaryResult{}, err
	}
	summaryMessage, err := s.repository.CreateMessage(ctx, Message{
		UserID:       input.UserID,
		ThreadID:     input.ThreadID,
		Role:         RoleAssistant,
		Content:      result.Reply,
		Status:       MessageStatusCompleted,
		Model:        model,
		InputTokens:  aiResult.InputTokens,
		OutputTokens: aiResult.OutputTokens,
		Metadata:     messageKindMetadata(messageKindCompareSummary),
		CreatedAt:    s.now(),
	})
	if err != nil {
		s.refundQuota(ctx, input.UserID, membership.FeatureCopilotMessages, 1, quotaKey, "assistant-message")
		return CompareSummaryResult{}, err
	}
	return CompareSummaryResult{SummaryMessage: summaryMessage}, nil
}

func (s *Service) SmokeModel(ctx context.Context, input ModelSmokeInput) (ModelSmokeResult, error) {
	if s.generator == nil {
		return ModelSmokeResult{}, ErrServiceNotReady
	}
	model, err := s.normalizeRequestedModel(input.Model)
	if err != nil {
		return ModelSmokeResult{}, err
	}
	prompt := strings.TrimSpace(input.Prompt)
	if prompt == "" {
		prompt = "ping"
	}
	aiResult, err := s.generator.GenerateJSON(ctx, ai.GenerateJSONRequest{
		UserID:         input.UserID,
		Feature:        "copilot.model_smoke",
		PromptVersion:  "copilot_model_smoke_v1",
		Model:          model,
		SystemPrompt:   "You are a model route smoke test. Return JSON only with the schema {\"reply\":\"pong\"}.",
		UserPrompt:     fmt.Sprintf("Input: %s\nReturn {\"reply\":\"pong\"} if this route is working.", prompt),
		SchemaName:     "copilot_model_smoke_response",
		Validate:       validateModelSmokeJSON,
		RepairAttempts: 1,
	})
	if err != nil {
		return ModelSmokeResult{}, fmt.Errorf("%w: %v", ErrInvalidAIResult, err)
	}
	var result modelSmokeAIResult
	if err := json.Unmarshal(aiResult.Content, &result); err != nil {
		return ModelSmokeResult{}, fmt.Errorf("%w: %v", ErrInvalidAIResult, err)
	}
	if err := validateModelSmokeResult(result); err != nil {
		return ModelSmokeResult{}, fmt.Errorf("%w: %v", ErrInvalidAIResult, err)
	}
	return ModelSmokeResult{
		OK:           true,
		Model:        model,
		Reply:        strings.TrimSpace(result.Reply),
		InputTokens:  aiResult.InputTokens,
		OutputTokens: aiResult.OutputTokens,
	}, nil
}

func (s *Service) consumeQuota(ctx context.Context, userID int64, feature string, amount int, idempotencyKey string) error {
	if s.quota == nil {
		return nil
	}
	_, err := s.quota.CheckAndConsume(ctx, membership.ConsumeInput{
		UserID:         userID,
		FeatureKey:     feature,
		Amount:         amount,
		IdempotencyKey: idempotencyKey,
	})
	return err
}

func (s *Service) refundQuota(ctx context.Context, userID int64, feature string, amount int, idempotencyKey, reason string) {
	if s.quota == nil || amount <= 0 {
		return
	}
	_, _ = s.quota.RefundUsage(ctx, membership.ConsumeInput{
		UserID:         userID,
		FeatureKey:     feature,
		Amount:         amount,
		IdempotencyKey: idempotencyKey + "-refund-" + reason,
	})
}

func (s *Service) normalizeRequestID(requestID string, userID, threadID int64) string {
	requestID = strings.TrimSpace(requestID)
	if requestID != "" {
		if len(requestID) > 96 {
			return requestID[:96]
		}
		return requestID
	}
	return fmt.Sprintf("%d-%d-%d", userID, threadID, s.now().UnixNano())
}

func (s *Service) ListMemories(ctx context.Context, userID int64, limit int) ([]Memory, error) {
	if s.repository == nil {
		return nil, ErrServiceNotReady
	}
	return s.repository.ListMemories(ctx, userID, normalizeLimit(limit, 50))
}

func (s *Service) SaveMemory(ctx context.Context, input MemoryInput) (Memory, error) {
	if s.repository == nil {
		return Memory{}, ErrServiceNotReady
	}
	memory, err := memoryFromInput(input, s.now())
	if err != nil {
		return Memory{}, err
	}
	return s.repository.UpsertMemory(ctx, memory)
}

func (s *Service) DeleteMemory(ctx context.Context, userID, id int64) error {
	if s.repository == nil {
		return ErrServiceNotReady
	}
	return s.repository.DeleteMemory(ctx, userID, id)
}

func (s *Service) SaveFile(ctx context.Context, input FileInput) (File, error) {
	if s.repository == nil {
		return File{}, ErrServiceNotReady
	}
	file, err := fileFromInput(input, s.now())
	if err != nil {
		return File{}, err
	}
	return s.repository.CreateFile(ctx, file)
}

func (s *Service) UploadFile(ctx context.Context, input UploadFileInput) (File, error) {
	if s.repository == nil {
		return File{}, ErrServiceNotReady
	}
	name := normalizeUploadName(input.Name)
	if name == "" || len(input.Data) == 0 {
		return File{}, ErrInvalidInput
	}
	if len(input.Data) > maxUploadFileBytes {
		return File{}, ErrFileTooLarge
	}
	content, mimeType, err := extractUploadedText(name, input.MimeType, input.Data)
	if err != nil {
		return File{}, err
	}
	digest := fmt.Sprintf("%x", sha256.Sum256(input.Data))
	quotaKey := "copilot-file-" + digest
	if err := s.consumeQuota(ctx, input.UserID, membership.FeatureCopilotFileAnalysis, 1, quotaKey); err != nil {
		return File{}, err
	}
	file, err := s.repository.CreateFile(ctx, File{
		UserID:         input.UserID,
		Name:           name,
		MimeType:       mimeType,
		SizeBytes:      len(input.Data),
		Content:        content,
		Status:         FileStatusReady,
		Source:         FileSourceUpload,
		SHA256:         digest,
		ExtractedChars: utf8.RuneCountInString(content),
		CreatedAt:      s.now(),
	})
	if err != nil {
		s.refundQuota(ctx, input.UserID, membership.FeatureCopilotFileAnalysis, 1, quotaKey, "persistence")
		return File{}, err
	}
	return file, nil
}

func (s *Service) ListFiles(ctx context.Context, userID int64, limit int) ([]File, error) {
	if s.repository == nil {
		return nil, ErrServiceNotReady
	}
	return s.repository.ListFiles(ctx, userID, normalizeLimit(limit, 50))
}

func (s *Service) GetFile(ctx context.Context, userID, id int64) (File, error) {
	if s.repository == nil {
		return File{}, ErrServiceNotReady
	}
	return s.repository.GetFile(ctx, userID, id)
}

func (s *Service) DeleteFile(ctx context.Context, userID, id int64) error {
	if s.repository == nil {
		return ErrServiceNotReady
	}
	return s.repository.DeleteFile(ctx, userID, id)
}

func (s *Service) referenceFiles(ctx context.Context, userID int64, ids []int64) ([]File, error) {
	ids = normalizeReferenceIDs(ids)
	if len(ids) == 0 {
		return nil, nil
	}
	files, err := s.repository.GetFilesByIDs(ctx, userID, ids)
	if err != nil {
		return nil, err
	}
	if len(files) != len(ids) {
		return nil, ErrFileNotFound
	}
	return files, nil
}

func (s *Service) generateReply(ctx context.Context, userID int64, thread Thread, content, model string, references []File, taskContext *TaskContext, competitorContext *competitorScanContext, growthContext *growthModelContext, monitoringContext *monitoringContext, projectContext *projectContext) (ai.GenerateJSONResult, chatAIResult, error) {
	memories, _ := s.repository.ListMemories(ctx, userID, 20)
	messages, _ := s.repository.ListMessages(ctx, userID, thread.ID, 12)
	aiResult, err := s.generator.GenerateJSON(ctx, ai.GenerateJSONRequest{
		UserID:         userID,
		Feature:        "copilot.chat",
		PromptVersion:  "copilot_chat_v1",
		Model:          model,
		SystemPrompt:   "你是智活 Copilot，回答要直接、可执行。必须只返回 JSON，字段严格匹配 copilot_chat_response。",
		UserPrompt:     buildUserPrompt(thread, memories, messages, references, taskContext, competitorContext, growthContext, monitoringContext, projectContext, content, true),
		SchemaName:     "copilot_chat_response",
		Validate:       validateChatJSON,
		RepairAttempts: 1,
	})
	if err != nil {
		return ai.GenerateJSONResult{}, chatAIResult{}, fmt.Errorf("%w: %v", ErrInvalidAIResult, err)
	}
	var result chatAIResult
	if err := json.Unmarshal(aiResult.Content, &result); err != nil {
		return ai.GenerateJSONResult{}, chatAIResult{}, fmt.Errorf("%w: %v", ErrInvalidAIResult, err)
	}
	if err := validateChatResult(result); err != nil {
		return ai.GenerateJSONResult{}, chatAIResult{}, fmt.Errorf("%w: %v", ErrInvalidAIResult, err)
	}
	_ = model
	return aiResult, result, nil
}

func (s *Service) generateSummary(ctx context.Context, userID int64, prompt, model string) (ai.GenerateJSONResult, chatAIResult, error) {
	aiResult, err := s.generator.GenerateJSON(ctx, ai.GenerateJSONRequest{
		UserID:         userID,
		Feature:        "copilot.compare_summary",
		PromptVersion:  "copilot_compare_summary_v1",
		Model:          model,
		SystemPrompt:   "你是智活 Copilot，负责综合多个模型的回答。必须只返回 JSON，字段严格匹配 copilot_chat_response。",
		UserPrompt:     prompt,
		SchemaName:     "copilot_chat_response",
		Validate:       validateChatJSON,
		RepairAttempts: 1,
	})
	if err != nil {
		return ai.GenerateJSONResult{}, chatAIResult{}, fmt.Errorf("%w: %v", ErrInvalidAIResult, err)
	}
	var result chatAIResult
	if err := json.Unmarshal(aiResult.Content, &result); err != nil {
		return ai.GenerateJSONResult{}, chatAIResult{}, fmt.Errorf("%w: %v", ErrInvalidAIResult, err)
	}
	if err := validateChatResult(result); err != nil {
		return ai.GenerateJSONResult{}, chatAIResult{}, fmt.Errorf("%w: %v", ErrInvalidAIResult, err)
	}
	return aiResult, result, nil
}

func buildUserPrompt(thread Thread, memories []Memory, messages []Message, references []File, taskContext *TaskContext, competitorContext *competitorScanContext, growthContext *growthModelContext, monitoringContext *monitoringContext, projectContext *projectContext, content string, includeResponseSchema bool) string {
	var builder strings.Builder
	builder.WriteString("会话标题：")
	builder.WriteString(thread.Title)
	builder.WriteString("\n\n用户长期记忆：\n")
	for _, memory := range memories {
		builder.WriteString("- ")
		builder.WriteString(memory.Key)
		builder.WriteString(": ")
		builder.WriteString(memory.Value)
		builder.WriteString("\n")
	}
	builder.WriteString("\n最近消息：\n")
	for _, message := range messages {
		builder.WriteString(message.Role)
		builder.WriteString(": ")
		builder.WriteString(message.Content)
		builder.WriteString("\n")
	}
	if len(references) > 0 {
		builder.WriteString("\n引用文件：\n")
		for _, file := range references {
			builder.WriteString("- ")
			builder.WriteString(file.Name)
			builder.WriteString(" (")
			builder.WriteString(file.MimeType)
			builder.WriteString("):\n")
			builder.WriteString(truncateForPrompt(file.Content, maxReferencePromptBytes))
			builder.WriteString("\n")
		}
	}
	if taskContext != nil {
		builder.WriteString("\n")
		builder.WriteString(taskContextPrompt(taskContext))
		builder.WriteString("\n")
	}
	if competitorContext != nil {
		payload, err := json.Marshal(competitorContext)
		if err == nil {
			builder.WriteString("\n当前竞品分析上下文（已按当前用户权限读取）：")
			builder.Write(payload)
			builder.WriteString("\n")
		}
	}
	if growthContext != nil {
		payload, err := json.Marshal(growthContext)
		if err == nil {
			builder.WriteString("\n当前增长测算上下文（已按当前用户权限读取）：")
			builder.Write(payload)
			builder.WriteString("\n")
		}
	}
	if monitoringContext != nil {
		payload, err := json.Marshal(monitoringContext)
		if err == nil {
			builder.WriteString("\n当前竞品动态监测上下文（已按当前用户权限读取）：")
			builder.Write(payload)
			builder.WriteString("\n")
		}
	}
	if projectContext != nil {
		payload, err := json.Marshal(projectContext)
		if err == nil {
			builder.WriteString("\n当前项目超市上下文（项目目录为已发布内容，匹配记录已按当前用户权限读取）：")
			builder.Write(payload)
			builder.WriteString("\n")
		}
	}
	builder.WriteString("\n用户当前问题：")
	builder.WriteString(content)
	if includeResponseSchema {
		builder.WriteString("\n\n返回 JSON：{\"reply\":\"...\",\"memory_candidates\":[{\"key\":\"...\",\"value\":\"...\",\"confidence\":0.9,\"source\":\"copilot\"}]}")
	}
	return builder.String()
}

func taskContextPrompt(taskContext *TaskContext) string {
	if taskContext == nil {
		return "当前任务上下文：无"
	}
	payload, err := json.Marshal(taskContext)
	if err != nil {
		return "当前任务上下文：无"
	}
	return "当前任务上下文（已按当前用户权限读取）：" + string(payload)
}

func buildCompareSummaryPrompt(thread Thread, content string, answers []CompareAnswer) (string, error) {
	var builder strings.Builder
	builder.WriteString("会话标题：")
	builder.WriteString(thread.Title)
	builder.WriteString("\n\n用户问题：")
	builder.WriteString(content)
	builder.WriteString("\n\n模型回答：\n")
	validAnswers := 0
	for _, answer := range answers {
		body := strings.TrimSpace(answer.AssistantMessage.Content)
		if body == "" {
			continue
		}
		validAnswers++
		builder.WriteString("- 模型 ")
		builder.WriteString(strings.TrimSpace(answer.Model))
		builder.WriteString("：")
		builder.WriteString(body)
		builder.WriteString("\n")
	}
	if validAnswers == 0 {
		return "", ErrInvalidInput
	}
	builder.WriteString("\n请输出综合结论，指出共识、分歧、推荐决策和下一步动作。返回 JSON：{\"reply\":\"...\"}")
	return builder.String(), nil
}

func validateChatJSON(data []byte) error {
	var result chatAIResult
	if err := json.Unmarshal(data, &result); err != nil {
		return err
	}
	return validateChatResult(result)
}

func validateChatResult(result chatAIResult) error {
	if strings.TrimSpace(result.Reply) == "" {
		return errors.New("reply is required")
	}
	for _, candidate := range result.MemoryCandidates {
		if strings.TrimSpace(candidate.Key) == "" || strings.TrimSpace(candidate.Value) == "" {
			return errors.New("memory candidate is missing required fields")
		}
	}
	return nil
}

func validateModelSmokeJSON(data []byte) error {
	var result modelSmokeAIResult
	if err := json.Unmarshal(data, &result); err != nil {
		return err
	}
	return validateModelSmokeResult(result)
}

func validateModelSmokeResult(result modelSmokeAIResult) error {
	if strings.TrimSpace(result.Reply) == "" {
		return errors.New("reply is required")
	}
	return nil
}

func (s *Service) saveMemoryCandidates(ctx context.Context, userID int64, candidates []MemoryCandidate) {
	for _, candidate := range candidates {
		if candidate.Confidence < 0.7 {
			continue
		}
		memory, err := memoryFromInput(MemoryInput{
			UserID:     userID,
			Key:        candidate.Key,
			Value:      candidate.Value,
			Confidence: candidate.Confidence,
			Source:     candidate.Source,
		}, s.now())
		if err != nil {
			continue
		}
		_, _ = s.repository.UpsertMemory(ctx, memory)
	}
}

func (s *Service) extractAndSaveMemories(ctx context.Context, userID int64, model, userMessage, assistantMessage string) {
	if s.generator == nil {
		return
	}
	result, err := s.generator.GenerateJSON(ctx, ai.GenerateJSONRequest{
		UserID:        userID,
		Feature:       "copilot.memory_extract",
		PromptVersion: "copilot_memory_extract_v1",
		Model:         model,
		SystemPrompt:  "你负责提取用户长期记忆。只记录用户明确表达且未来对话仍有帮助的稳定事实、偏好、背景或长期目标；不要记录临时问题、一次性指令、AI 的推断或敏感凭证。最多返回 3 条。必须只返回 JSON。",
		UserPrompt: fmt.Sprintf(
			"用户消息：%s\n\n助手回复：%s\n\n返回 JSON：{\"memory_candidates\":[{\"key\":\"简短分类\",\"value\":\"明确事实或偏好\",\"confidence\":0.9,\"source\":\"copilot\"}]}。没有合适内容时返回空数组。",
			truncateForPrompt(userMessage, 3000),
			truncateForPrompt(assistantMessage, 3000),
		),
		SchemaName:     "copilot_memory_extraction",
		Validate:       validateMemoryExtractionJSON,
		RepairAttempts: 1,
	})
	if err != nil {
		return
	}
	var extracted memoryExtractionAIResult
	if json.Unmarshal(result.Content, &extracted) != nil {
		return
	}
	s.saveMemoryCandidates(ctx, userID, extracted.MemoryCandidates)
}

func validateMemoryExtractionJSON(data []byte) error {
	var result memoryExtractionAIResult
	if err := json.Unmarshal(data, &result); err != nil {
		return err
	}
	if len(result.MemoryCandidates) > 3 {
		return errors.New("too many memory candidates")
	}
	for _, candidate := range result.MemoryCandidates {
		if strings.TrimSpace(candidate.Key) == "" || strings.TrimSpace(candidate.Value) == "" {
			return errors.New("memory candidate is missing required fields")
		}
	}
	return nil
}

func memoryFromInput(input MemoryInput, now time.Time) (Memory, error) {
	key := strings.TrimSpace(input.Key)
	value := strings.TrimSpace(input.Value)
	if key == "" || value == "" || utf8.RuneCountInString(key) > maxMemoryKeyRunes || utf8.RuneCountInString(value) > maxMemoryValueRunes {
		return Memory{}, ErrInvalidInput
	}
	confidence := input.Confidence
	if confidence <= 0 {
		confidence = 1
	}
	if confidence > 1 {
		confidence = 1
	}
	return Memory{
		UserID:     input.UserID,
		Key:        key,
		Value:      value,
		Confidence: confidence,
		Source:     strings.TrimSpace(input.Source),
		CreatedAt:  now,
	}, nil
}

func fileFromInput(input FileInput, now time.Time) (File, error) {
	name := strings.TrimSpace(input.Name)
	content := strings.TrimSpace(input.Content)
	if name == "" || content == "" {
		return File{}, ErrInvalidInput
	}
	if len([]byte(content)) > maxFileContentBytes {
		return File{}, ErrInvalidInput
	}
	mimeType := strings.TrimSpace(input.MimeType)
	if mimeType == "" {
		mimeType = "text/plain"
	}
	digest := fmt.Sprintf("%x", sha256.Sum256([]byte(content)))
	return File{
		UserID:         input.UserID,
		Name:           name,
		MimeType:       mimeType,
		SizeBytes:      len([]byte(content)),
		Content:        content,
		Status:         FileStatusReady,
		Source:         FileSourcePasted,
		SHA256:         digest,
		ExtractedChars: utf8.RuneCountInString(content),
		CreatedAt:      now,
	}, nil
}

func normalizeUploadName(name string) string {
	name = strings.TrimSpace(strings.ReplaceAll(name, "\\", "/"))
	name = filepath.Base(name)
	if name == "." || name == "/" {
		return ""
	}
	return name
}

func extractUploadedText(name, providedMimeType string, data []byte) (string, string, error) {
	extension := strings.ToLower(filepath.Ext(name))
	if extension == ".docx" {
		content, err := extractDOCXText(data)
		if err != nil {
			if errors.Is(err, ErrFileTooLarge) {
				return "", "", err
			}
			return "", "", ErrInvalidFileEncoding
		}
		return limitExtractedText(content), "application/vnd.openxmlformats-officedocument.wordprocessingml.document", nil
	}
	allowedTextExtensions := map[string]bool{
		".txt": true, ".md": true, ".markdown": true, ".csv": true, ".tsv": true,
		".json": true, ".yaml": true, ".yml": true, ".xml": true, ".html": true, ".htm": true,
	}
	if !allowedTextExtensions[extension] {
		return "", "", ErrUnsupportedFileType
	}
	if !utf8.Valid(data) {
		return "", "", ErrInvalidFileEncoding
	}
	content := strings.TrimSpace(strings.TrimPrefix(string(data), "\ufeff"))
	if content == "" {
		return "", "", ErrInvalidInput
	}
	mimeType := strings.TrimSpace(strings.Split(providedMimeType, ";")[0])
	if mimeType == "" || mimeType == "application/octet-stream" {
		mimeType = "text/plain"
	}
	return limitExtractedText(content), mimeType, nil
}

func extractDOCXText(data []byte) (string, error) {
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", err
	}
	for _, part := range reader.File {
		if part.Name != "word/document.xml" {
			continue
		}
		if part.UncompressedSize64 > maxDOCXXMLBytes {
			return "", ErrFileTooLarge
		}
		stream, err := part.Open()
		if err != nil {
			return "", err
		}
		defer stream.Close()
		return decodeDOCXDocument(io.LimitReader(stream, maxDOCXXMLBytes+1))
	}
	return "", ErrInvalidFileEncoding
}

func decodeDOCXDocument(reader io.Reader) (string, error) {
	decoder := xml.NewDecoder(reader)
	var builder strings.Builder
	for {
		token, err := decoder.Token()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return "", err
		}
		switch value := token.(type) {
		case xml.CharData:
			builder.Write([]byte(value))
		case xml.EndElement:
			if value.Name.Local == "p" && builder.Len() > 0 {
				builder.WriteByte('\n')
			}
		}
	}
	content := strings.TrimSpace(builder.String())
	if content == "" {
		return "", ErrInvalidFileEncoding
	}
	return content, nil
}

func limitExtractedText(content string) string {
	if len([]byte(content)) <= maxFileContentBytes {
		return content
	}
	return truncateForPrompt(content, maxFileContentBytes)
}

func normalizeReferenceIDs(ids []int64) []int64 {
	seen := map[int64]struct{}{}
	normalized := make([]int64, 0, min(len(ids), maxReferenceFiles))
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		normalized = append(normalized, id)
		if len(normalized) == maxReferenceFiles {
			break
		}
	}
	return normalized
}

func referenceMetadata(ids []int64) []byte {
	ids = normalizeReferenceIDs(ids)
	if len(ids) == 0 {
		return nil
	}
	data, _ := json.Marshal(messageMetadata{ReferenceIDs: ids})
	return data
}

func sendMessageMetadata(input SendMessageInput) []byte {
	metadata := messageMetadata{
		ReferenceIDs:  normalizeReferenceIDs(input.ReferenceIDs),
		TaskID:        input.TaskID,
		CurrentView:   strings.TrimSpace(input.CurrentView),
		ActiveFilters: normalizeContextFilters(input.ActiveFilters),
	}
	if len(metadata.ReferenceIDs) == 0 && metadata.TaskID == 0 && metadata.CurrentView == "" && len(metadata.ActiveFilters) == 0 {
		return nil
	}
	data, _ := json.Marshal(metadata)
	return data
}

func normalizeContextFilters(filters map[string]string) map[string]string {
	if len(filters) == 0 {
		return nil
	}
	normalized := make(map[string]string, min(len(filters), 12))
	for key, value := range filters {
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key == "" || value == "" || len(normalized) >= 12 {
			continue
		}
		normalized[key] = value
	}
	if len(normalized) == 0 {
		return nil
	}
	return normalized
}

func truncateForPrompt(content string, maxBytes int) string {
	if len([]byte(content)) <= maxBytes {
		return content
	}
	var builder strings.Builder
	for _, next := range content {
		if builder.Len()+len(string(next)) > maxBytes {
			break
		}
		builder.WriteRune(next)
	}
	builder.WriteString("\n[内容已截断]")
	return builder.String()
}

func normalizeLimit(limit, fallback int) int {
	if limit <= 0 {
		return fallback
	}
	if limit > 100 {
		return 100
	}
	return limit
}

func messageKindMetadata(kind string) []byte {
	data, _ := json.Marshal(messageMetadata{Kind: kind})
	return data
}

func normalizeModelOptions(models []ModelOption) []ModelOption {
	normalized := make([]ModelOption, 0, len(models))
	hasDefault := false
	for _, model := range models {
		model.Name = strings.TrimSpace(model.Name)
		model.Value = strings.TrimSpace(model.Value)
		model.Provider = strings.TrimSpace(model.Provider)
		if model.Name == "" {
			model.Name = model.Value
		}
		if model.Value == "" {
			continue
		}
		if model.IsDefault {
			hasDefault = true
		}
		normalized = append(normalized, model)
	}
	if len(normalized) > 0 && !hasDefault {
		normalized[0].IsDefault = true
	}
	return normalized
}

func (s *Service) normalizeRequestedModel(model string) (string, error) {
	model = strings.TrimSpace(model)
	if len(s.models) == 0 {
		return model, nil
	}
	if model == "" {
		for _, option := range s.models {
			if option.IsDefault {
				return option.Value, nil
			}
		}
		return s.models[0].Value, nil
	}
	for _, option := range s.models {
		if option.Value == model {
			return model, nil
		}
	}
	return "", ErrInvalidInput
}

func (s *Service) normalizeRequestedModels(models []string) ([]string, error) {
	seen := map[string]struct{}{}
	normalized := make([]string, 0, len(models))
	for _, model := range models {
		model, err := s.normalizeRequestedModel(model)
		if err != nil {
			return nil, err
		}
		if model == "" {
			continue
		}
		if _, ok := seen[model]; ok {
			continue
		}
		seen[model] = struct{}{}
		normalized = append(normalized, model)
	}
	if len(normalized) == 0 {
		model, err := s.normalizeRequestedModel("")
		if err != nil {
			return nil, err
		}
		if model == "" {
			return nil, ErrInvalidInput
		}
		normalized = append(normalized, model)
	}
	if len(normalized) > 4 {
		return nil, ErrInvalidInput
	}
	return normalized, nil
}
