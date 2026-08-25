package copilot

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/zzm/opcv2/internal/account"
	"github.com/zzm/opcv2/internal/ai"
	"github.com/zzm/opcv2/internal/competitor"
	"github.com/zzm/opcv2/internal/crm"
	"github.com/zzm/opcv2/internal/growth"
	"github.com/zzm/opcv2/internal/learning"
	"github.com/zzm/opcv2/internal/membership"
	"github.com/zzm/opcv2/internal/projects"
	"github.com/zzm/opcv2/internal/sandbox"
	"github.com/zzm/opcv2/internal/tasks"
)

type fakeRepository struct {
	threads         []Thread
	messages        []Message
	memories        []Memory
	files           []File
	aiRuns          []ai.Run
	createdThread   Thread
	createdMessages []Message
	createdFile     File
	upsertedMemory  Memory
	deletedMemoryID int64
	deletedFileID   int64
	listRunsUserID  int64
	listRunsPrefix  string
	listRunsLimit   int
	err             error
}

func (r *fakeRepository) CreateThread(_ context.Context, thread Thread) (Thread, error) {
	r.createdThread = thread
	thread.ID = 99
	thread.UpdatedAt = thread.CreatedAt
	r.threads = append(r.threads, thread)
	return thread, r.err
}

func (r *fakeRepository) ListThreads(_ context.Context, userID int64, limit int) ([]Thread, error) {
	if r.err != nil {
		return nil, r.err
	}
	var threads []Thread
	for _, thread := range r.threads {
		if thread.UserID == userID {
			threads = append(threads, thread)
		}
	}
	return threads[:min(len(threads), limit)], nil
}

func (r *fakeRepository) GetThread(_ context.Context, userID, id int64) (Thread, error) {
	if r.err != nil {
		return Thread{}, r.err
	}
	for _, thread := range r.threads {
		if thread.UserID == userID && thread.ID == id {
			return thread, nil
		}
	}
	return Thread{}, ErrThreadNotFound
}

func (r *fakeRepository) UpdateThreadTitle(_ context.Context, userID, id int64, title string) (Thread, error) {
	for index, thread := range r.threads {
		if thread.UserID == userID && thread.ID == id {
			r.threads[index].Title = title
			return r.threads[index], nil
		}
	}
	return Thread{}, ErrThreadNotFound
}

func (r *fakeRepository) ArchiveThread(_ context.Context, userID, id int64) error {
	for index, thread := range r.threads {
		if thread.UserID == userID && thread.ID == id {
			now := time.Now()
			r.threads[index].ArchivedAt = &now
			return nil
		}
	}
	return ErrThreadNotFound
}

func (r *fakeRepository) CreateMessage(_ context.Context, message Message) (Message, error) {
	message.ID = int64(len(r.messages) + 1)
	r.messages = append(r.messages, message)
	r.createdMessages = append(r.createdMessages, message)
	return message, r.err
}

func (r *fakeRepository) GetMessage(_ context.Context, userID, threadID, messageID int64) (Message, error) {
	for _, message := range r.messages {
		if message.UserID == userID && message.ThreadID == threadID && message.ID == messageID {
			return message, nil
		}
	}
	return Message{}, ErrToolPreviewNotFound
}

func (r *fakeRepository) ClaimToolPreview(_ context.Context, userID, threadID, messageID int64) (Message, error) {
	for index, message := range r.messages {
		if message.UserID != userID || message.ThreadID != threadID || message.ID != messageID {
			continue
		}
		var metadata messageMetadata
		if json.Unmarshal(message.Metadata, &metadata) != nil || metadata.ToolPreview == nil || metadata.ToolPreview.Status != ToolPreviewPending {
			return Message{}, ErrToolPreviewConflict
		}
		metadata.ToolPreview.Status = ToolPreviewExecuting
		data, _ := json.Marshal(metadata)
		r.messages[index].Metadata = data
		return r.messages[index], nil
	}
	return Message{}, ErrToolPreviewConflict
}

func (r *fakeRepository) UpdateMessageContentMetadata(_ context.Context, userID, threadID, messageID int64, content string, metadata json.RawMessage) (Message, error) {
	for index, message := range r.messages {
		if message.UserID == userID && message.ThreadID == threadID && message.ID == messageID {
			r.messages[index].Content = content
			r.messages[index].Metadata = metadata
			return r.messages[index], nil
		}
	}
	return Message{}, ErrToolPreviewNotFound
}

func (r *fakeRepository) ListMessages(_ context.Context, userID, threadID int64, limit int) ([]Message, error) {
	if r.err != nil {
		return nil, r.err
	}
	var messages []Message
	for _, message := range r.messages {
		if message.UserID == userID && message.ThreadID == threadID {
			messages = append(messages, message)
		}
	}
	return messages[:min(len(messages), limit)], nil
}

func (r *fakeRepository) ListMemories(_ context.Context, userID int64, limit int) ([]Memory, error) {
	if r.err != nil {
		return nil, r.err
	}
	var memories []Memory
	for _, memory := range r.memories {
		if memory.UserID == userID {
			memories = append(memories, memory)
		}
	}
	return memories[:min(len(memories), limit)], nil
}

func (r *fakeRepository) UpsertMemory(_ context.Context, memory Memory) (Memory, error) {
	if memory.Status == "" {
		memory.Status = MemoryStatusActive
	}
	r.upsertedMemory = memory
	memory.ID = 7
	r.memories = append(r.memories, memory)
	return memory, r.err
}

func (r *fakeRepository) UpdateMemory(_ context.Context, input MemoryUpdateInput) (Memory, error) {
	for index, memory := range r.memories {
		if memory.UserID != input.UserID || memory.ID != input.ID {
			continue
		}
		if input.Key != "" {
			r.memories[index].Key = input.Key
		}
		if input.Value != "" {
			r.memories[index].Value = input.Value
		}
		if input.Status != "" {
			r.memories[index].Status = input.Status
		}
		return r.memories[index], nil
	}
	return Memory{}, ErrMemoryNotFound
}

func (r *fakeRepository) DeleteMemory(_ context.Context, userID, id int64) error {
	r.deletedMemoryID = id
	for index, memory := range r.memories {
		if memory.UserID == userID && memory.ID == id {
			r.memories = append(r.memories[:index], r.memories[index+1:]...)
			return nil
		}
	}
	return ErrMemoryNotFound
}

func (r *fakeRepository) CreateFile(_ context.Context, file File) (File, error) {
	r.createdFile = file
	file.ID = 17
	file.UpdatedAt = file.CreatedAt
	r.files = append(r.files, file)
	return file, r.err
}

func (r *fakeRepository) ListFiles(_ context.Context, userID int64, limit int) ([]File, error) {
	if r.err != nil {
		return nil, r.err
	}
	var files []File
	for _, file := range r.files {
		if file.UserID == userID {
			files = append(files, file)
		}
	}
	return files[:min(len(files), limit)], nil
}

func (r *fakeRepository) GetFilesByIDs(_ context.Context, userID int64, ids []int64) ([]File, error) {
	if r.err != nil {
		return nil, r.err
	}
	wanted := map[int64]struct{}{}
	for _, id := range ids {
		wanted[id] = struct{}{}
	}
	var files []File
	for _, file := range r.files {
		if file.UserID != userID {
			continue
		}
		if _, ok := wanted[file.ID]; ok {
			files = append(files, file)
		}
	}
	return files, nil
}

func (r *fakeRepository) GetFile(_ context.Context, userID, id int64) (File, error) {
	for _, file := range r.files {
		if file.UserID == userID && file.ID == id {
			return file, nil
		}
	}
	return File{}, ErrFileNotFound
}

func (r *fakeRepository) DeleteFile(_ context.Context, userID, id int64) error {
	r.deletedFileID = id
	for index, file := range r.files {
		if file.UserID == userID && file.ID == id {
			r.files = append(r.files[:index], r.files[index+1:]...)
			return nil
		}
	}
	return ErrFileNotFound
}

func (r *fakeRepository) ListRuns(_ context.Context, userID int64, featurePrefix string, limit int) ([]ai.Run, error) {
	r.listRunsUserID = userID
	r.listRunsPrefix = featurePrefix
	r.listRunsLimit = limit
	if r.err != nil {
		return nil, r.err
	}
	return r.aiRuns[:min(len(r.aiRuns), limit)], nil
}

type fakeGenerator struct {
	request ai.GenerateJSONRequest
	content []byte
	err     error
}

type fakeTextStreamer struct {
	deltas      []string
	request     ai.GenerateTextRequest
	jsonRequest ai.GenerateJSONRequest
	jsonContent []byte
	err         error
}

func (g *fakeTextStreamer) GenerateJSON(_ context.Context, request ai.GenerateJSONRequest) (ai.GenerateJSONResult, error) {
	g.jsonRequest = request
	if g.err != nil {
		return ai.GenerateJSONResult{}, g.err
	}
	if request.Validate != nil {
		if err := request.Validate(g.jsonContent); err != nil {
			return ai.GenerateJSONResult{}, err
		}
	}
	return ai.GenerateJSONResult{Content: g.jsonContent}, nil
}

type fakeToolExecutor struct {
	call   ToolCall
	userID int64
	calls  int
	result ToolExecutionResult
	err    error
}

func (e *fakeToolExecutor) Execute(_ context.Context, userID, _ int64, call ToolCall) (ToolExecutionResult, error) {
	e.userID = userID
	e.call = call
	e.calls++
	return e.result, e.err
}

type fakeTaskContextProvider struct {
	task     tasks.Task
	subtasks []tasks.Subtask
	userID   int64
	taskID   int64
	err      error
}

type fakeCompetitorContextProvider struct {
	scan   competitor.Scan
	userID int64
	scanID int64
	err    error
}

type fakeGrowthContextProvider struct {
	model     growth.Model
	models    []growth.Model
	userID    int64
	modelID   int64
	listLimit int
}

func (p *fakeGrowthContextProvider) GetModel(_ context.Context, userID, modelID int64) (growth.Model, error) {
	p.userID = userID
	p.modelID = modelID
	return p.model, nil
}

func (p *fakeGrowthContextProvider) ListModels(_ context.Context, userID int64, limit int) ([]growth.Model, error) {
	p.userID = userID
	p.listLimit = limit
	return p.models, nil
}

type fakeMonitoringContextProvider struct {
	snapshot competitor.MonitoringSnapshot
	userID   int64
	limit    int
}

func (p *fakeMonitoringContextProvider) GetMonitoring(_ context.Context, userID int64, limit int) (competitor.MonitoringSnapshot, error) {
	p.userID = userID
	p.limit = limit
	return p.snapshot, nil
}

type fakeProjectContextProvider struct {
	project    projects.Project
	match      projects.MatchWorkflowResponse
	catalog    []projects.Opportunity
	projectRef string
	userID     int64
	matchID    int64
}

type fakeLearningContextProvider struct {
	course          learning.Course
	progress        learning.Progress
	progresses      []learning.Progress
	diagnosis       learning.Diagnosis
	plan            learning.DiagnosisPlan
	courseSlug      string
	progressUserID  int64
	progressSlug    string
	progressUser    int64
	diagnosisUser   int64
	planUserID      int64
	planDiagnosisID int64
	diagnosisErr    error
}

type fakeSandboxContextProvider struct {
	run        sandbox.V2SandboxRun
	runs       []sandbox.V2SandboxRun
	runUserID  int64
	runID      int64
	listUserID int64
	listLimit  int
}

type fakeCRMContextProvider struct {
	customer        crm.Customer
	activities      []crm.Activity
	followUps       []crm.FollowUp
	pipeline        crm.PipelineStats
	userID          int64
	customerID      int64
	activitiesLimit int
	followUpsInput  crm.ListFollowUpsInput
	dueTodayInput   crm.ListDueInput
	pipelineUserID  int64
}

type fakeProfileContextProvider struct {
	profile account.ProfileContext
	userID  int64
	err     error
}

func (p *fakeLearningContextProvider) GetCourse(_ context.Context, slug string) (learning.Course, error) {
	p.courseSlug = slug
	return p.course, nil
}

func (p *fakeLearningContextProvider) GetProgress(_ context.Context, userID int64, courseSlug string) (learning.Progress, error) {
	p.progressUserID = userID
	p.progressSlug = courseSlug
	return p.progress, nil
}

func (p *fakeLearningContextProvider) ListProgress(_ context.Context, userID int64) ([]learning.Progress, error) {
	p.progressUser = userID
	return p.progresses, nil
}

func (p *fakeLearningContextProvider) LatestDiagnosis(_ context.Context, userID int64) (learning.Diagnosis, error) {
	p.diagnosisUser = userID
	return p.diagnosis, p.diagnosisErr
}

func (p *fakeLearningContextProvider) GetPlan(_ context.Context, userID, diagnosisID int64) (learning.DiagnosisPlan, error) {
	p.planUserID = userID
	p.planDiagnosisID = diagnosisID
	return p.plan, nil
}

func (p *fakeSandboxContextProvider) GetV2Run(_ context.Context, userID, runID int64) (sandbox.V2SandboxRun, error) {
	p.runUserID = userID
	p.runID = runID
	return p.run, nil
}

func (p *fakeSandboxContextProvider) ListV2Runs(_ context.Context, userID int64, limit int) ([]sandbox.V2SandboxRun, error) {
	p.listUserID = userID
	p.listLimit = limit
	return p.runs, nil
}

func (p *fakeCRMContextProvider) GetCustomer(_ context.Context, userID, customerID int64) (crm.Customer, error) {
	p.userID = userID
	p.customerID = customerID
	return p.customer, nil
}

func (p *fakeCRMContextProvider) ListActivities(_ context.Context, userID, customerID int64, limit int) ([]crm.Activity, error) {
	p.userID = userID
	p.customerID = customerID
	p.activitiesLimit = limit
	return p.activities, nil
}

func (p *fakeCRMContextProvider) ListFollowUps(_ context.Context, input crm.ListFollowUpsInput) ([]crm.FollowUp, error) {
	p.followUpsInput = input
	return p.followUps, nil
}

func (p *fakeCRMContextProvider) ListDueCustomers(_ context.Context, input crm.ListDueInput) ([]crm.Customer, error) {
	p.dueTodayInput = input
	return nil, nil
}

func (p *fakeCRMContextProvider) PipelineStats(_ context.Context, userID int64) (crm.PipelineStats, error) {
	p.pipelineUserID = userID
	return p.pipeline, nil
}

func (p *fakeProfileContextProvider) GetProfileContext(_ context.Context, userID int64) (account.ProfileContext, error) {
	p.userID = userID
	return p.profile, p.err
}

func (p *fakeProjectContextProvider) GetProject(_ context.Context, ref string) (projects.Project, error) {
	p.projectRef = ref
	return p.project, nil
}

func (p *fakeProjectContextProvider) GetProjectMatch(_ context.Context, userID, matchID int64) (projects.MatchWorkflowResponse, error) {
	p.userID = userID
	p.matchID = matchID
	return p.match, nil
}

func (p *fakeProjectContextProvider) ListOpportunities(_ context.Context, _ projects.OpportunityFilters) ([]projects.Opportunity, error) {
	return p.catalog, nil
}

func (p *fakeCompetitorContextProvider) GetScan(_ context.Context, userID, scanID int64) (competitor.Scan, error) {
	p.userID = userID
	p.scanID = scanID
	return p.scan, p.err
}

func (p *fakeTaskContextProvider) GetTask(_ context.Context, userID, taskID int64) (tasks.Task, error) {
	p.userID = userID
	p.taskID = taskID
	return p.task, p.err
}

func (p *fakeTaskContextProvider) ListSubtasks(_ context.Context, userID, taskID int64) ([]tasks.Subtask, error) {
	p.userID = userID
	p.taskID = taskID
	return p.subtasks, p.err
}

func (g *fakeTextStreamer) GenerateTextStream(_ context.Context, request ai.GenerateTextRequest, onDelta func([]byte) error) (ai.GenerateTextResult, error) {
	g.request = request
	if g.err != nil {
		return ai.GenerateTextResult{}, g.err
	}
	var content strings.Builder
	for _, delta := range g.deltas {
		content.WriteString(delta)
		if err := onDelta([]byte(delta)); err != nil {
			return ai.GenerateTextResult{}, err
		}
	}
	return ai.GenerateTextResult{Content: content.String(), InputTokens: 5, OutputTokens: 8}, nil
}

func (g *fakeGenerator) GenerateJSON(_ context.Context, request ai.GenerateJSONRequest) (ai.GenerateJSONResult, error) {
	g.request = request
	if g.err != nil {
		return ai.GenerateJSONResult{}, g.err
	}
	if request.Validate != nil {
		if err := request.Validate(g.content); err != nil {
			return ai.GenerateJSONResult{}, err
		}
	}
	return ai.GenerateJSONResult{Content: g.content, InputTokens: 12, OutputTokens: 24}, nil
}

type fakeGeneratorQueue struct {
	requests []ai.GenerateJSONRequest
	contents [][]byte
	err      error
}

type fakeQuotaConsumer struct {
	consumed []membership.ConsumeInput
	refunded []membership.ConsumeInput
	err      error
}

func (q *fakeQuotaConsumer) CheckAndConsume(_ context.Context, input membership.ConsumeInput) (membership.UsageItem, error) {
	q.consumed = append(q.consumed, input)
	return membership.UsageItem{}, q.err
}

func (q *fakeQuotaConsumer) RefundUsage(_ context.Context, input membership.ConsumeInput) (membership.UsageItem, error) {
	q.refunded = append(q.refunded, input)
	return membership.UsageItem{}, nil
}

func (g *fakeGeneratorQueue) GenerateJSON(_ context.Context, request ai.GenerateJSONRequest) (ai.GenerateJSONResult, error) {
	g.requests = append(g.requests, request)
	if g.err != nil {
		return ai.GenerateJSONResult{}, g.err
	}
	index := len(g.requests) - 1
	content := []byte(`{"reply":"ok"}`)
	if index < len(g.contents) {
		content = g.contents[index]
	}
	if request.Validate != nil {
		if err := request.Validate(content); err != nil {
			return ai.GenerateJSONResult{}, err
		}
	}
	return ai.GenerateJSONResult{Content: content, InputTokens: 10 + index, OutputTokens: 20 + index}, nil
}

func TestServiceCreatesThreadWithDefaults(t *testing.T) {
	now := time.Date(2026, 6, 30, 10, 0, 0, 0, time.UTC)
	repository := &fakeRepository{}
	service := NewService(repository, nil)
	service.now = func() time.Time { return now }

	thread, err := service.CreateThread(context.Background(), CreateThreadInput{
		UserID: 42,
		Title:  "  智能客服机会分析  ",
	})

	if err != nil {
		t.Fatalf("CreateThread() error = %v", err)
	}
	if thread.ID != 99 || thread.Title != "智能客服机会分析" || thread.Mode != ModeChat {
		t.Fatalf("thread = %+v", thread)
	}
	if repository.createdThread.UserID != 42 || repository.createdThread.CreatedAt != now {
		t.Fatalf("createdThread = %+v", repository.createdThread)
	}
}

func TestServiceSendMessageGeneratesAssistantReplyAndMemory(t *testing.T) {
	payload, _ := json.Marshal(chatAIResult{
		Reply: "建议先从教培机构的高频咨询场景做小范围验证。",
		MemoryCandidates: []MemoryCandidate{{
			Key:        "industry",
			Value:      "教培",
			Confidence: 0.92,
			Source:     "copilot",
		}},
	})
	repository := &fakeRepository{
		threads: []Thread{{ID: 99, UserID: 42, Title: "机会分析", Mode: ModeChat}},
		memories: []Memory{{
			ID:         1,
			UserID:     42,
			Key:        "preferred_style",
			Value:      "直接给执行清单",
			Confidence: 0.8,
		}},
	}
	generator := &fakeGenerator{content: payload}
	service := NewService(repository, generator)

	result, err := service.SendMessage(context.Background(), SendMessageInput{
		UserID:   42,
		ThreadID: 99,
		Content:  "帮我分析智能客服市场机会",
		Model:    "gpt-test",
	})

	if err != nil {
		t.Fatalf("SendMessage() error = %v", err)
	}
	if result.UserMessage.Role != RoleUser || result.AssistantMessage.Role != RoleAssistant {
		t.Fatalf("result = %+v", result)
	}
	if result.AssistantMessage.Content == "" || result.AssistantMessage.OutputTokens != 24 {
		t.Fatalf("assistant message = %+v", result.AssistantMessage)
	}
	if len(repository.createdMessages) != 2 {
		t.Fatalf("createdMessages = %+v", repository.createdMessages)
	}
	if repository.upsertedMemory.Key != "industry" || repository.upsertedMemory.UserID != 42 {
		t.Fatalf("upsertedMemory = %+v", repository.upsertedMemory)
	}
	if generator.request.Feature != "copilot.chat" || generator.request.SchemaName != "copilot_chat_response" {
		t.Fatalf("ai request = %+v", generator.request)
	}
}

func TestServiceSendMessageConsumesQuotaWithRequestID(t *testing.T) {
	repository := &fakeRepository{threads: []Thread{{ID: 99, UserID: 42, Title: "机会分析", Mode: ModeChat}}}
	quota := &fakeQuotaConsumer{}
	service := NewService(repository, &fakeGenerator{content: []byte(`{"reply":"先验证需求。"}`)}, WithQuotaConsumer(quota))

	_, err := service.SendMessage(context.Background(), SendMessageInput{
		UserID: 42, ThreadID: 99, Content: "分析机会", RequestID: "msg-001",
	})

	if err != nil {
		t.Fatalf("SendMessage() error = %v", err)
	}
	if len(quota.consumed) != 1 {
		t.Fatalf("consumed = %+v", quota.consumed)
	}
	consume := quota.consumed[0]
	if consume.FeatureKey != membership.FeatureCopilotMessages || consume.Amount != 1 || consume.IdempotencyKey != "copilot-message-msg-001" {
		t.Fatalf("consume = %+v", consume)
	}
}

func TestServiceStreamsMessageDeltasAndPersistsFinalReply(t *testing.T) {
	repository := &fakeRepository{threads: []Thread{{ID: 99, UserID: 42, Title: "机会分析", Mode: ModeChat}}}
	streamer := &fakeTextStreamer{deltas: []string{"先验证", "客户需求。"}}
	service := NewService(repository, streamer)
	var events []StreamEvent

	result, err := service.StreamMessage(context.Background(), SendMessageInput{
		UserID: 42, ThreadID: 99, Content: "分析机会", RequestID: "stream-001",
	}, func(event StreamEvent) error {
		events = append(events, event)
		return nil
	})

	if err != nil {
		t.Fatalf("StreamMessage() error = %v", err)
	}
	if result.AssistantMessage.Content != "先验证客户需求。" || len(events) != 4 {
		t.Fatalf("result/events = %+v/%+v", result, events)
	}
	if events[0].Type != StreamEventUserMessage || events[1].Delta != "先验证" || events[3].Type != StreamEventAssistantMessage {
		t.Fatalf("events = %+v", events)
	}
	if streamer.request.Feature != "copilot.chat_stream" {
		t.Fatalf("request = %+v", streamer.request)
	}
}

func TestServiceStreamFailureDoesNotPersistAssistantPlaceholder(t *testing.T) {
	repository := &fakeRepository{threads: []Thread{{ID: 99, UserID: 42, Title: "机会分析", Mode: ModeChat}}}
	streamer := &fakeTextStreamer{err: errors.New("provider unavailable")}
	service := NewService(repository, streamer)

	_, err := service.StreamMessage(context.Background(), SendMessageInput{
		UserID: 42, ThreadID: 99, Content: "分析机会", RequestID: "stream-002",
	}, func(StreamEvent) error { return nil })

	if !errors.Is(err, ErrInvalidAIResult) {
		t.Fatalf("err = %v, want ErrInvalidAIResult", err)
	}
	if len(repository.createdMessages) != 1 || repository.createdMessages[0].Role != RoleUser {
		t.Fatalf("failed stream persisted assistant message: %+v", repository.createdMessages)
	}
}

func TestServiceStreamExtractsStableMemoryAfterReply(t *testing.T) {
	repository := &fakeRepository{threads: []Thread{{ID: 99, UserID: 42, Title: "偏好设置", Mode: ModeChat}}}
	streamer := &fakeTextStreamer{
		deltas:      []string{"后续我会先给结论。"},
		jsonContent: []byte(`{"memory_candidates":[{"key":"回答偏好","value":"先给结论，再给步骤","confidence":0.95,"source":"copilot"}]}`),
	}
	service := NewService(repository, streamer)

	_, err := service.StreamMessage(context.Background(), SendMessageInput{
		UserID: 42, ThreadID: 99, Content: "记住我喜欢先看结论，再看步骤", Model: "deepseek-chat",
	}, func(StreamEvent) error { return nil })

	if err != nil {
		t.Fatalf("StreamMessage() error = %v", err)
	}
	if streamer.jsonRequest.Feature != "copilot.memory_extract" || streamer.jsonRequest.SchemaName != "copilot_memory_extraction" {
		t.Fatalf("memory request = %+v", streamer.jsonRequest)
	}
	if repository.upsertedMemory.Key != "回答偏好" || repository.upsertedMemory.Value != "先给结论，再给步骤" {
		t.Fatalf("upserted memory = %+v", repository.upsertedMemory)
	}
	if repository.upsertedMemory.Status != MemoryStatusPending {
		t.Fatalf("upserted memory status = %q, want %q", repository.upsertedMemory.Status, MemoryStatusPending)
	}
	if strings.Contains(streamer.request.UserPrompt, "返回 JSON") {
		t.Fatalf("stream prompt unexpectedly requests JSON: %s", streamer.request.UserPrompt)
	}
}

func TestServiceExcludesUnconfirmedMemoriesFromPrompt(t *testing.T) {
	repository := &fakeRepository{
		threads:  []Thread{{ID: 99, UserID: 42, Title: "记忆控制", Mode: ModeChat}},
		memories: []Memory{{ID: 1, UserID: 42, Key: "待确认事实", Value: "不应传给模型", Status: MemoryStatusPending}, {ID: 2, UserID: 42, Key: "已确认事实", Value: "应传给模型", Status: MemoryStatusActive}},
	}
	streamer := &fakeTextStreamer{deltas: []string{"已处理"}}
	service := NewService(repository, streamer)
	if _, err := service.StreamMessage(context.Background(), SendMessageInput{UserID: 42, ThreadID: 99, Content: "继续"}, func(StreamEvent) error { return nil }); err != nil {
		t.Fatalf("StreamMessage() error = %v", err)
	}
	if strings.Contains(streamer.request.UserPrompt, "不应传给模型") || !strings.Contains(streamer.request.UserPrompt, "应传给模型") {
		t.Fatalf("memory status filtering failed: %s", streamer.request.UserPrompt)
	}
}

func TestServicePreviewsWhitelistedToolAndExecutesOnlyAfterConfirmation(t *testing.T) {
	repository := &fakeRepository{threads: []Thread{{ID: 99, UserID: 42, Title: "执行计划", Mode: ModeChat}}}
	streamer := &fakeTextStreamer{jsonContent: []byte(`{"tool":"create_task","arguments":{"title":"访谈10位客户","description":"记录高频问题","priority":"high"}}`)}
	executor := &fakeToolExecutor{result: ToolExecutionResult{
		Tool: ToolCreateTask, Status: "completed", EntityID: 81, Title: "访谈10位客户", URL: "/tasks", Message: "已创建任务：访谈10位客户",
	}}
	service := NewService(repository, streamer, WithToolExecutor(executor))
	var events []StreamEvent

	result, err := service.StreamMessage(context.Background(), SendMessageInput{
		UserID: 42, ThreadID: 99, Content: "帮我创建任务：访谈10位客户", RequestID: "tool-001",
	}, func(event StreamEvent) error {
		events = append(events, event)
		return nil
	})

	if err != nil {
		t.Fatalf("StreamMessage() error = %v", err)
	}
	if executor.call.Tool != "" || executor.userID != 0 {
		t.Fatalf("tool executed before confirmation: %+v", executor)
	}
	if !strings.Contains(result.AssistantMessage.Content, "准备创建任务") || !strings.Contains(string(result.AssistantMessage.Metadata), "tool_preview") || !strings.Contains(string(result.AssistantMessage.Metadata), ToolPreviewPending) {
		t.Fatalf("assistant = %+v", result.AssistantMessage)
	}
	if len(events) != 2 || events[1].Type != StreamEventAssistantMessage {
		t.Fatalf("events = %+v", events)
	}
	if streamer.request.Feature != "" {
		t.Fatalf("text stream should not run for tool preview: %+v", streamer.request)
	}

	confirmed, err := service.ConfirmTool(context.Background(), ToolConfirmationInput{
		UserID: 42, ThreadID: 99, MessageID: result.AssistantMessage.ID, Decision: "confirm",
	})
	if err != nil {
		t.Fatalf("ConfirmTool() error = %v", err)
	}
	if executor.call.Tool != ToolCreateTask || executor.userID != 42 || executor.call.Arguments.Title != "访谈10位客户" {
		t.Fatalf("executor = %+v", executor)
	}
	if confirmed.ToolResult == nil || confirmed.ToolResult.EntityID != 81 || !strings.Contains(string(confirmed.Message.Metadata), "tool_result") {
		t.Fatalf("confirmed = %+v", confirmed)
	}

	repeated, err := service.ConfirmTool(context.Background(), ToolConfirmationInput{
		UserID: 42, ThreadID: 99, MessageID: result.AssistantMessage.ID, Decision: "confirm",
	})
	if err != nil || repeated.ToolResult == nil || repeated.ToolResult.EntityID != 81 {
		t.Fatalf("repeated confirmation = %+v, %v", repeated, err)
	}
	if executor.calls != 1 {
		t.Fatalf("executor calls = %d, want 1", executor.calls)
	}
}

func TestServiceRecognizesExplicitProjectCreationWithoutAIIntentCall(t *testing.T) {
	repository := &fakeRepository{threads: []Thread{{ID: 99, UserID: 42, Title: "项目计划", Mode: ModeChat}}}
	streamer := &fakeTextStreamer{err: errors.New("provider unavailable")}
	executor := &fakeToolExecutor{}
	service := NewService(repository, streamer, WithToolExecutor(executor))

	result, err := service.StreamMessage(context.Background(), SendMessageInput{
		UserID: 42, ThreadID: 99, Content: "帮我创建一个项目：AI 客服助手，描述：先验证客服问答场景", RequestID: "project-tool-001",
	}, func(StreamEvent) error { return nil })

	if err != nil {
		t.Fatalf("StreamMessage() error = %v", err)
	}
	var metadata messageMetadata
	if err := json.Unmarshal(result.AssistantMessage.Metadata, &metadata); err != nil {
		t.Fatalf("metadata error = %v", err)
	}
	if metadata.ToolPreview == nil || metadata.ToolPreview.Call.Tool != ToolCreateProject {
		t.Fatalf("tool preview = %+v", metadata.ToolPreview)
	}
	if metadata.ToolPreview.Call.Arguments.Title != "AI 客服助手" || metadata.ToolPreview.Call.Arguments.Description != "先验证客服问答场景" {
		t.Fatalf("tool arguments = %+v", metadata.ToolPreview.Call.Arguments)
	}
	if streamer.jsonRequest.Feature != "" || streamer.request.Feature != "" {
		t.Fatalf("explicit project creation unexpectedly called AI: json=%+v stream=%+v", streamer.jsonRequest, streamer.request)
	}
}

func TestExplicitProjectCreationRequiresProjectName(t *testing.T) {
	if call, ok := explicitProjectCreationCall("怎么创建项目？"); ok {
		t.Fatalf("unexpected tool call = %+v", call)
	}
	if call, ok := explicitProjectCreationCall("帮我创建一个项目"); ok {
		t.Fatalf("unexpected nameless tool call = %+v", call)
	}
}

func TestServiceCancelsToolPreviewWithoutExecution(t *testing.T) {
	repository := &fakeRepository{threads: []Thread{{ID: 99, UserID: 42, Title: "执行计划", Mode: ModeChat}}}
	streamer := &fakeTextStreamer{jsonContent: []byte(`{"tool":"create_task","arguments":{"title":"整理访谈记录"}}`)}
	executor := &fakeToolExecutor{}
	service := NewService(repository, streamer, WithToolExecutor(executor))
	result, err := service.SendMessage(context.Background(), SendMessageInput{UserID: 42, ThreadID: 99, Content: "创建任务：整理访谈记录"})
	if err != nil {
		t.Fatalf("SendMessage() error = %v", err)
	}
	confirmed, err := service.ConfirmTool(context.Background(), ToolConfirmationInput{UserID: 42, ThreadID: 99, MessageID: result.AssistantMessage.ID, Decision: "cancel"})
	if err != nil {
		t.Fatalf("ConfirmTool(cancel) error = %v", err)
	}
	if executor.calls != 0 || !strings.Contains(confirmed.Message.Content, "取消") {
		t.Fatalf("cancelled = %+v, executor = %+v", confirmed.Message, executor)
	}
}

func TestServiceIncludesAuthorizedTaskContextInPrompt(t *testing.T) {
	repository := &fakeRepository{threads: []Thread{{ID: 99, UserID: 42, Title: "任务分析", Mode: ModeChat}}}
	streamer := &fakeTextStreamer{deltas: []string{"结合任务上下文回答"}}
	provider := &fakeTaskContextProvider{
		task:     tasks.Task{ID: 7, UserID: 42, Title: "上线任务中心", Status: tasks.StatusInProgress, Priority: tasks.PriorityHigh, Progress: 35},
		subtasks: []tasks.Subtask{{ID: 1, TaskID: 7, Title: "检查接口", Completed: true}, {ID: 2, TaskID: 7, Title: "验证页面", Completed: false}},
	}
	service := NewService(repository, streamer, WithTaskContextProvider(provider))
	_, err := service.StreamMessage(context.Background(), SendMessageInput{
		UserID: 42, ThreadID: 99, Content: "总结当前任务", TaskID: 7, CurrentView: "list",
		ActiveFilters: map[string]string{"status": "in_progress"},
	}, func(StreamEvent) error { return nil })
	if err != nil {
		t.Fatalf("StreamMessage() error = %v", err)
	}
	if provider.userID != 42 || provider.taskID != 7 || !strings.Contains(streamer.request.UserPrompt, "上线任务中心") || !strings.Contains(streamer.request.UserPrompt, `"completed_subtasks":1`) || !strings.Contains(streamer.request.UserPrompt, `"current_view":"list"`) {
		t.Fatalf("task context not included: %s", streamer.request.UserPrompt)
	}
}

func TestServiceIncludesAuthorizedCompetitorContextInPrompt(t *testing.T) {
	repository := &fakeRepository{threads: []Thread{{ID: 99, UserID: 42, Title: "竞品分析", Mode: ModeChat}}}
	streamer := &fakeTextStreamer{deltas: []string{"结合竞品分析上下文回答"}}
	provider := &fakeCompetitorContextProvider{scan: competitor.Scan{
		ID: 13, UserID: 42, Targets: []string{"小鹅通"}, Focus: "产品能力", Status: competitor.StatusRunning,
		ProgressPercent: 60, CurrentStep: "analyzing", Conclusions: []competitor.Conclusion{{Title: "机会", Detail: "强化差异化"}},
	}}
	service := NewService(repository, streamer, WithCompetitorContextProvider(provider))

	_, err := service.StreamMessage(context.Background(), SendMessageInput{
		UserID: 42, ThreadID: 99, Content: "分析当前页面", CurrentView: "/competitor-data/progress?scanId=13",
		ActiveFilters: map[string]string{"module": "data", "view": "progress", "scan_id": "13"},
	}, func(StreamEvent) error { return nil })
	if err != nil {
		t.Fatalf("StreamMessage() error = %v", err)
	}
	if provider.userID != 42 || provider.scanID != 13 || !strings.Contains(streamer.request.UserPrompt, "小鹅通") || !strings.Contains(streamer.request.UserPrompt, "强化差异化") {
		t.Fatalf("competitor context not included: %s", streamer.request.UserPrompt)
	}
}

func TestServiceIncludesAuthorizedGrowthContextInPrompt(t *testing.T) {
	repository := &fakeRepository{threads: []Thread{{ID: 99, UserID: 42, Title: "增长测算", Mode: ModeChat}}}
	streamer := &fakeTextStreamer{deltas: []string{"结合增长测算上下文回答"}}
	provider := &fakeGrowthContextProvider{model: growth.Model{
		ID: 21, UserID: 42, Name: "SaaS 续费提升", BusinessType: "saas", RiskLevel: "medium",
		Assumptions: growth.Assumptions{MonthlyVisits: 12000, LeadRate: 0.08, DealRate: 0.2, AverageOrder: 3800, AcquisitionCost: 600, DeliveryCost: 900},
		Result:      growth.Result{MonthlyRevenue: 729600, Deals: 192, PaybackDays: 38, NetMargin: 0.31},
	}}
	service := NewService(repository, streamer, WithGrowthContextProvider(provider))

	_, err := service.StreamMessage(context.Background(), SendMessageInput{
		UserID: 42, ThreadID: 99, Content: "分析当前测算", CurrentView: "/growth-calculator/report?model_id=21",
		ActiveFilters: map[string]string{"module": "growth", "view": "report", "model_id": "21"},
	}, func(StreamEvent) error { return nil })
	if err != nil {
		t.Fatalf("StreamMessage() error = %v", err)
	}
	if provider.userID != 42 || provider.modelID != 21 || !strings.Contains(streamer.request.UserPrompt, "SaaS 续费提升") || !strings.Contains(streamer.request.UserPrompt, "729600") {
		t.Fatalf("growth context not included: %s", streamer.request.UserPrompt)
	}
}

func TestServiceIncludesAuthorizedMonitoringContextInPrompt(t *testing.T) {
	repository := &fakeRepository{threads: []Thread{{ID: 99, UserID: 42, Title: "动态监测", Mode: ModeChat}}}
	streamer := &fakeTextStreamer{deltas: []string{"结合监测上下文回答"}}
	provider := &fakeMonitoringContextProvider{snapshot: competitor.MonitoringSnapshot{
		Watchlist: []competitor.WatchItem{{ID: 7, Name: "小鹅通", Category: "SaaS", Threat: "强", Signal: "招聘扩张"}},
		Events:    []competitor.Event{{Company: "小鹅通", Title: "招聘增长岗位", Detail: "新增多个增长岗位", Level: "强"}},
	}}
	service := NewService(repository, streamer, WithMonitoringContextProvider(provider))

	_, err := service.StreamMessage(context.Background(), SendMessageInput{
		UserID: 42, ThreadID: 99, Content: "分析当前预警", CurrentView: "/competitor-monitoring",
		ActiveFilters: map[string]string{"module": "monitoring", "view": "home"},
	}, func(StreamEvent) error { return nil })
	if err != nil {
		t.Fatalf("StreamMessage() error = %v", err)
	}
	if provider.userID != 42 || provider.limit != 20 || !strings.Contains(streamer.request.UserPrompt, "招聘增长岗位") || !strings.Contains(streamer.request.UserPrompt, "新增多个增长岗位") {
		t.Fatalf("monitoring context not included: %s", streamer.request.UserPrompt)
	}
}

func TestServiceIncludesPublishedProjectAndOwnedMatchContextInPrompt(t *testing.T) {
	repository := &fakeRepository{threads: []Thread{{ID: 99, UserID: 42, Title: "项目评估", Mode: ModeChat}}}
	streamer := &fakeTextStreamer{deltas: []string{"结合项目上下文回答"}}
	provider := &fakeProjectContextProvider{
		project: projects.Project{
			ID: 11, Slug: "ai-sales", Title: "AI 销售顾问", Summary: "为销售团队梳理线索和跟进流程。",
			BudgetBand: "1 万以内", Difficulty: "中等", Tags: []string{"企业服务"}, ResourceRequirements: []string{"销售经验"},
		},
		match: projects.MatchWorkflowResponse{
			MatchID: 19, Need: "寻找适合一人启动的销售服务", Status: "ready", AnalysisSummary: "已有销售经验，可先验证单一行业。",
			Assumptions: []string{"先与三位目标客户访谈"}, Completeness: 0.8,
		},
	}
	service := NewService(repository, streamer, WithProjectContextProvider(provider))

	_, err := service.StreamMessage(context.Background(), SendMessageInput{
		UserID: 42, ThreadID: 99, Content: "这个项目下一步怎么验证？", CurrentView: "/projects/ai-sales?section=path",
		ActiveFilters: map[string]string{"module": "projects", "view": "detail", "project_ref": "ai-sales", "match_id": "19", "section": "path"},
	}, func(StreamEvent) error { return nil })
	if err != nil {
		t.Fatalf("StreamMessage() error = %v", err)
	}
	if provider.projectRef != "ai-sales" || provider.userID != 42 || provider.matchID != 19 ||
		!strings.Contains(streamer.request.UserPrompt, "AI 销售顾问") ||
		!strings.Contains(streamer.request.UserPrompt, "先与三位目标客户访谈") ||
		!strings.Contains(streamer.request.UserPrompt, `"current_section":"path"`) {
		t.Fatalf("project context not included: %s", streamer.request.UserPrompt)
	}
}

func TestServiceIncludesProjectCatalogForProjectMarketQuestion(t *testing.T) {
	repository := &fakeRepository{threads: []Thread{{ID: 99, UserID: 42, Title: "智活 Copilot", Mode: ModeChat}}}
	streamer := &fakeTextStreamer{deltas: []string{"结合项目目录回答"}}
	provider := &fakeProjectContextProvider{catalog: []projects.Opportunity{
		{Slug: "ai-sales", Title: "AI 销售顾问", Summary: "帮助企业整理销售线索", Industry: "企业服务", Tags: []string{"一人公司"}, BudgetBand: "1 万以内", Difficulty: "低"},
	}}
	service := NewService(repository, streamer, WithProjectContextProvider(provider))

	_, err := service.StreamMessage(context.Background(), SendMessageInput{
		UserID: 42, ThreadID: 99, Content: "请推荐项目超市里的项目",
	}, func(StreamEvent) error { return nil })
	if err != nil {
		t.Fatalf("StreamMessage() error = %v", err)
	}
	if !strings.Contains(streamer.request.UserPrompt, "AI 销售顾问") || !strings.Contains(streamer.request.UserPrompt, "一人公司") {
		t.Fatalf("project catalog not included: %s", streamer.request.UserPrompt)
	}
}

func TestServiceIncludesProjectCatalogForNaturalLanguageProjectSearch(t *testing.T) {
	repository := &fakeRepository{threads: []Thread{{ID: 99, UserID: 42, Title: "智活 Copilot", Mode: ModeChat}}}
	streamer := &fakeTextStreamer{deltas: []string{"结合区块链项目目录回答"}}
	provider := &fakeProjectContextProvider{catalog: []projects.Opportunity{
		{Slug: "chain-credit", Title: "ChainCredit", Summary: "B2B 嵌入式贷款基础设施", Industry: "金融科技", Tags: []string{"区块链"}},
	}}
	service := NewService(repository, streamer, WithProjectContextProvider(provider))

	_, err := service.StreamMessage(context.Background(), SendMessageInput{
		UserID: 42, ThreadID: 99, Content: "我想找一个区块链的项目",
	}, func(StreamEvent) error { return nil })
	if err != nil {
		t.Fatalf("StreamMessage() error = %v", err)
	}
	if !strings.Contains(streamer.request.UserPrompt, "ChainCredit") || !strings.Contains(streamer.request.UserPrompt, "区块链") {
		t.Fatalf("natural-language project catalog not included: %s", streamer.request.UserPrompt)
	}
}

func TestServiceIncludesProjectCatalogForProjectsPageView(t *testing.T) {
	repository := &fakeRepository{threads: []Thread{{ID: 99, UserID: 42, Title: "项目超市", Mode: ModeChat}}}
	streamer := &fakeTextStreamer{deltas: []string{"结合当前项目页回答"}}
	provider := &fakeProjectContextProvider{catalog: []projects.Opportunity{
		{Slug: "ai-sales", Title: "AI 销售顾问", Summary: "销售线索服务"},
	}}
	service := NewService(repository, streamer, WithProjectContextProvider(provider))

	_, err := service.StreamMessage(context.Background(), SendMessageInput{
		UserID: 42, ThreadID: 99, Content: "帮我看看这个页面",
		CurrentView: "/projects",
	}, func(StreamEvent) error { return nil })
	if err != nil {
		t.Fatalf("StreamMessage() error = %v", err)
	}
	if !strings.Contains(streamer.request.UserPrompt, "AI 销售顾问") {
		t.Fatalf("projects page catalog not included: %s", streamer.request.UserPrompt)
	}
}

func TestServiceIncludesAuthorizedLearningContextInPrompt(t *testing.T) {
	repository := &fakeRepository{threads: []Thread{{ID: 99, UserID: 42, Title: "学习计划", Mode: ModeChat}}}
	streamer := &fakeTextStreamer{deltas: []string{"结合学习上下文回答"}}
	provider := &fakeLearningContextProvider{
		course:    learning.Course{ID: 6, Slug: "ai-market-analysis", Title: "AI 行业分析方法", Outline: []string{"行业地图", "竞品拆解"}},
		progress:  learning.Progress{UserID: 42, CourseSlug: "ai-market-analysis", CourseTitle: "AI 行业分析方法", Percent: 35, LastLesson: "行业地图"},
		diagnosis: learning.Diagnosis{ID: 23, UserID: 42, Goal: "提升市场分析能力", Project: "企业服务项目", Recommendations: []string{"完成一次竞品拆解"}},
		plan:      learning.DiagnosisPlan{DiagnosisID: 23, Title: "市场分析学习路径", Stages: []learning.PlanStage{{Number: 1, Title: "行业研究基础", Milestone: "输出市场地图"}}},
	}
	service := NewService(repository, streamer, WithLearningContextProvider(provider))

	_, err := service.StreamMessage(context.Background(), SendMessageInput{
		UserID: 42, ThreadID: 99, Content: "我应该先学什么？", CurrentView: "/learning/plan",
		ActiveFilters: map[string]string{"module": "learning", "view": "plan", "course_slug": "ai-market-analysis"},
	}, func(StreamEvent) error { return nil })
	if err != nil {
		t.Fatalf("StreamMessage() error = %v", err)
	}
	if provider.courseSlug != "ai-market-analysis" || provider.progressUserID != 42 || provider.progressSlug != "ai-market-analysis" ||
		provider.diagnosisUser != 42 || provider.planUserID != 42 || provider.planDiagnosisID != 23 ||
		!strings.Contains(streamer.request.UserPrompt, "AI 行业分析方法") ||
		!strings.Contains(streamer.request.UserPrompt, `"percent":35`) ||
		!strings.Contains(streamer.request.UserPrompt, "完成一次竞品拆解") ||
		!strings.Contains(streamer.request.UserPrompt, "市场分析学习路径") {
		t.Fatalf("learning context not included: %s", streamer.request.UserPrompt)
	}
}

func TestServiceAllowsLearningChatWithoutSavedDiagnosis(t *testing.T) {
	repository := &fakeRepository{threads: []Thread{{ID: 99, UserID: 42, Title: "学习咨询", Mode: ModeChat}}}
	streamer := &fakeTextStreamer{deltas: []string{"可以先从课程目录开始。"}}
	provider := &fakeLearningContextProvider{diagnosisErr: learning.ErrDiagnosisNotFound}
	service := NewService(repository, streamer, WithLearningContextProvider(provider))

	_, err := service.StreamMessage(context.Background(), SendMessageInput{
		UserID: 42, ThreadID: 99, Content: "我该从哪里开始？", ActiveFilters: map[string]string{"module": "learning", "view": "home"},
	}, func(StreamEvent) error { return nil })
	if err != nil {
		t.Fatalf("StreamMessage() error = %v", err)
	}
	if !strings.Contains(streamer.request.UserPrompt, `"view":"home"`) || strings.Contains(streamer.request.UserPrompt, "\"diagnosis\"") {
		t.Fatalf("unexpected learning prompt: %s", streamer.request.UserPrompt)
	}
}

func TestServiceIncludesAuthorizedSandboxRunContextInPrompt(t *testing.T) {
	repository := &fakeRepository{threads: []Thread{{ID: 99, UserID: 42, Title: "沙盘复盘", Mode: ModeChat}}}
	streamer := &fakeTextStreamer{deltas: []string{"结合沙盘推演回答"}}
	provider := &fakeSandboxContextProvider{run: sandbox.V2SandboxRun{
		ID: 31, UserID: 42, Name: "门店 AI 运营助手", Status: sandbox.V2StatusDone,
		Product: sandbox.V2Product{Name: "门店 AI 运营助手", SellingPoint: "自动生成运营动作"},
		Context: sandbox.V2RunContext{TargetCustomer: "连锁门店"}, Roles: []string{"customer", "skeptic"},
		Report: &sandbox.V2SandboxReport{
			Summary:     "建议先验证门店付费意愿。",
			Feasibility: sandbox.V2Feasibility{Score: 72, Level: "可小范围验证", Basis: "需求有待访谈验证。"},
			Advice:      []sandbox.V2Advice{{Action: "访谈十家门店", Why: "验证真实需求", Priority: 1, Effort: "1 周"}},
		},
	}}
	service := NewService(repository, streamer, WithSandboxContextProvider(provider))

	_, err := service.StreamMessage(context.Background(), SendMessageInput{
		UserID: 42, ThreadID: 99, Content: "下一步如何验证？", CurrentView: "/sandbox-runs/31/report",
		ActiveFilters: map[string]string{"module": "sandbox", "view": "report", "run_id": "31"},
	}, func(StreamEvent) error { return nil })
	if err != nil {
		t.Fatalf("StreamMessage() error = %v", err)
	}
	if provider.runUserID != 42 || provider.runID != 31 ||
		!strings.Contains(streamer.request.UserPrompt, "门店 AI 运营助手") ||
		!strings.Contains(streamer.request.UserPrompt, `"score":72`) ||
		!strings.Contains(streamer.request.UserPrompt, "访谈十家门店") {
		t.Fatalf("sandbox context not included: %s", streamer.request.UserPrompt)
	}
}

func TestServiceIncludesAuthorizedSandboxHistoryInPrompt(t *testing.T) {
	repository := &fakeRepository{threads: []Thread{{ID: 99, UserID: 42, Title: "沙盘历史", Mode: ModeChat}}}
	streamer := &fakeTextStreamer{deltas: []string{"结合历史记录回答"}}
	provider := &fakeSandboxContextProvider{runs: []sandbox.V2SandboxRun{{ID: 31, Name: "门店 AI 运营助手", Product: sandbox.V2Product{Name: "门店 AI 运营助手"}, Status: sandbox.V2StatusDone}}}
	service := NewService(repository, streamer, WithSandboxContextProvider(provider))

	_, err := service.StreamMessage(context.Background(), SendMessageInput{
		UserID: 42, ThreadID: 99, Content: "先复盘哪一条？", ActiveFilters: map[string]string{"module": "sandbox", "view": "history"},
	}, func(StreamEvent) error { return nil })
	if err != nil {
		t.Fatalf("StreamMessage() error = %v", err)
	}
	if provider.listUserID != 42 || provider.listLimit != 10 || !strings.Contains(streamer.request.UserPrompt, `"recent_runs"`) || !strings.Contains(streamer.request.UserPrompt, "门店 AI 运营助手") {
		t.Fatalf("sandbox history context not included: %s", streamer.request.UserPrompt)
	}
}

func TestServiceIncludesAuthorizedCRMCustomerContextInPrompt(t *testing.T) {
	repository := &fakeRepository{threads: []Thread{{ID: 99, UserID: 42, Title: "客户跟进", Mode: ModeChat}}}
	streamer := &fakeTextStreamer{deltas: []string{"结合 CRM 上下文回答"}}
	provider := &fakeCRMContextProvider{
		customer:   crm.Customer{ID: 17, UserID: 42, Name: "启明星教育", Stage: crm.StageQualified, Source: crm.SourceLead},
		activities: []crm.Activity{{ID: 3, CustomerID: 17, Type: crm.ActivityFollowUpRecorded, Note: "确认了演示时间"}},
		followUps:  []crm.FollowUp{{ID: 5, CustomerID: 17, Note: "准备演示案例"}},
		pipeline:   crm.PipelineStats{Total: 9, Qualified: 2, DueToday: 1},
	}
	service := NewService(repository, streamer, WithCRMContextProvider(provider))

	_, err := service.StreamMessage(context.Background(), SendMessageInput{
		UserID: 42, ThreadID: 99, Content: "下一步如何跟进？", CurrentView: "/crm?customer_id=17",
		ActiveFilters: map[string]string{"module": "crm", "view": "customers", "customer_id": "17"},
	}, func(StreamEvent) error { return nil })
	if err != nil {
		t.Fatalf("StreamMessage() error = %v", err)
	}
	if provider.userID != 42 || provider.customerID != 17 || provider.activitiesLimit != 20 ||
		provider.followUpsInput.UserID != 42 || provider.followUpsInput.CustomerID != 17 || provider.followUpsInput.Limit != 20 || provider.pipelineUserID != 42 || provider.dueTodayInput.UserID != 0 ||
		!strings.Contains(streamer.request.UserPrompt, "启明星教育") ||
		!strings.Contains(streamer.request.UserPrompt, "确认了演示时间") ||
		!strings.Contains(streamer.request.UserPrompt, "准备演示案例") ||
		!strings.Contains(streamer.request.UserPrompt, `"qualified":2`) {
		t.Fatalf("CRM context not included: %s", streamer.request.UserPrompt)
	}
}

func TestServiceIncludesAuthorizedCRMOverviewContextInPrompt(t *testing.T) {
	repository := &fakeRepository{threads: []Thread{{ID: 99, UserID: 42, Title: "今日跟进", Mode: ModeChat}}}
	streamer := &fakeTextStreamer{deltas: []string{"结合 CRM 概览回答"}}
	provider := &fakeCRMContextProvider{
		pipeline: crm.PipelineStats{Total: 9, DueToday: 2},
	}
	service := NewService(repository, streamer, WithCRMContextProvider(provider))

	_, err := service.StreamMessage(context.Background(), SendMessageInput{
		UserID: 42, ThreadID: 99, Content: "今天先跟进谁？",
		ActiveFilters: map[string]string{"module": "crm", "view": "overview"},
	}, func(StreamEvent) error { return nil })
	if err != nil {
		t.Fatalf("StreamMessage() error = %v", err)
	}
	if provider.customerID != 0 || provider.pipelineUserID != 42 || provider.dueTodayInput.UserID != 42 || provider.dueTodayInput.Limit != 20 || !strings.Contains(streamer.request.UserPrompt, `"due_today":2`) {
		t.Fatalf("CRM overview context not included: %s", streamer.request.UserPrompt)
	}
}

func TestServiceIncludesUserMaintainedProfileContextInPrompt(t *testing.T) {
	repository := &fakeRepository{threads: []Thread{{ID: 99, UserID: 42, Title: "业务规划", Mode: ModeChat}}}
	streamer := &fakeTextStreamer{deltas: []string{"结合业务档案回答"}}
	provider := &fakeProfileContextProvider{profile: account.ProfileContext{
		UserID: 42,
		Groups: []account.ProfileGroup{
			{Key: account.ProfileGroupBusiness, Title: "我的业务/公司", Fields: map[string]string{"company": "智活科技"}},
			{Key: account.ProfileGroupProducts, Title: "产品与服务", Fields: map[string]string{"product": "AI 销售助手"}},
			{Key: account.ProfileGroupGoals, Title: "目标与阶段", Fields: map[string]string{"goal": "验证首批付费客户"}},
		},
	}}
	service := NewService(repository, streamer, WithProfileContextProvider(provider))

	_, err := service.StreamMessage(context.Background(), SendMessageInput{
		UserID: 42, ThreadID: 99, Content: "下一步应该做什么？",
	}, func(StreamEvent) error { return nil })
	if err != nil {
		t.Fatalf("StreamMessage() error = %v", err)
	}
	if provider.userID != 42 ||
		!strings.Contains(streamer.request.UserPrompt, "用户维护的业务档案") ||
		!strings.Contains(streamer.request.UserPrompt, "智活科技") ||
		!strings.Contains(streamer.request.UserPrompt, "AI 销售助手") ||
		!strings.Contains(streamer.request.UserPrompt, "验证首批付费客户") {
		t.Fatalf("profile context not included: %s", streamer.request.UserPrompt)
	}
}

func TestServiceSendMessageRefundsQuotaWhenGenerationFails(t *testing.T) {
	repository := &fakeRepository{threads: []Thread{{ID: 99, UserID: 42, Title: "机会分析", Mode: ModeChat}}}
	quota := &fakeQuotaConsumer{}
	service := NewService(repository, &fakeGenerator{err: errors.New("provider unavailable")}, WithQuotaConsumer(quota))

	_, err := service.SendMessage(context.Background(), SendMessageInput{
		UserID: 42, ThreadID: 99, Content: "分析机会", RequestID: "msg-002",
	})

	if !errors.Is(err, ErrInvalidAIResult) {
		t.Fatalf("err = %v, want ErrInvalidAIResult", err)
	}
	if len(quota.refunded) != 1 || quota.refunded[0].IdempotencyKey != "copilot-message-msg-002-refund-generation" {
		t.Fatalf("refunded = %+v", quota.refunded)
	}
	if len(repository.createdMessages) != 1 || repository.createdMessages[0].Role != RoleUser {
		t.Fatalf("failed generation persisted assistant message: %+v", repository.createdMessages)
	}
}

func TestServiceSendMessageStopsWhenQuotaIsExceeded(t *testing.T) {
	repository := &fakeRepository{threads: []Thread{{ID: 99, UserID: 42, Title: "机会分析", Mode: ModeChat}}}
	quota := &fakeQuotaConsumer{err: membership.ErrQuotaExceeded}
	service := NewService(repository, &fakeGenerator{content: []byte(`{"reply":"不会生成"}`)}, WithQuotaConsumer(quota))

	_, err := service.SendMessage(context.Background(), SendMessageInput{
		UserID: 42, ThreadID: 99, Content: "分析机会", RequestID: "msg-003",
	})

	if !errors.Is(err, membership.ErrQuotaExceeded) {
		t.Fatalf("err = %v, want ErrQuotaExceeded", err)
	}
	if len(repository.createdMessages) != 0 {
		t.Fatalf("createdMessages = %+v, want none", repository.createdMessages)
	}
}

func TestServiceSavesAndListsReferenceFiles(t *testing.T) {
	now := time.Date(2026, 7, 2, 9, 0, 0, 0, time.UTC)
	repository := &fakeRepository{}
	service := NewService(repository, nil)
	service.now = func() time.Time { return now }

	file, err := service.SaveFile(context.Background(), FileInput{
		UserID:   42,
		Name:     " 智能客服竞品功能对比表.txt ",
		MimeType: " text/plain ",
		Content:  "小鹅通：私域工具强；有赞教育：交易能力强。",
	})

	if err != nil {
		t.Fatalf("SaveFile() error = %v", err)
	}
	if file.ID != 17 || file.Name != "智能客服竞品功能对比表.txt" || file.SizeBytes == 0 {
		t.Fatalf("file = %+v", file)
	}
	if repository.createdFile.UserID != 42 || repository.createdFile.CreatedAt != now {
		t.Fatalf("createdFile = %+v", repository.createdFile)
	}

	files, err := service.ListFiles(context.Background(), 42, 10)
	if err != nil {
		t.Fatalf("ListFiles() error = %v", err)
	}
	if len(files) != 1 || files[0].Name != "智能客服竞品功能对比表.txt" {
		t.Fatalf("files = %+v", files)
	}
}

func TestServiceUploadsTextFileAndConsumesAnalysisQuota(t *testing.T) {
	repository := &fakeRepository{}
	quota := &fakeQuotaConsumer{}
	service := NewService(repository, nil, WithQuotaConsumer(quota))

	file, err := service.UploadFile(context.Background(), UploadFileInput{
		UserID: 42, Name: "客户访谈.txt", MimeType: "text/plain", Data: []byte("客户最关注响应速度。"),
	})

	if err != nil {
		t.Fatalf("UploadFile() error = %v", err)
	}
	if file.Name != "客户访谈.txt" || file.Content != "客户最关注响应速度。" || file.SizeBytes != len([]byte("客户最关注响应速度。")) {
		t.Fatalf("file = %+v", file)
	}
	if len(quota.consumed) != 1 || quota.consumed[0].FeatureKey != membership.FeatureCopilotFileAnalysis {
		t.Fatalf("consumed = %+v", quota.consumed)
	}
}

func TestServiceGetsAndDeletesOwnedFile(t *testing.T) {
	repository := &fakeRepository{files: []File{{ID: 17, UserID: 42, Name: "访谈.txt", Content: "正文", Status: FileStatusReady}}}
	service := NewService(repository, nil)

	file, err := service.GetFile(context.Background(), 42, 17)
	if err != nil || file.Content != "正文" {
		t.Fatalf("GetFile() = %+v, %v", file, err)
	}
	if err := service.DeleteFile(context.Background(), 42, 17); err != nil {
		t.Fatalf("DeleteFile() error = %v", err)
	}
	if repository.deletedFileID != 17 || len(repository.files) != 0 {
		t.Fatalf("repository = %+v", repository)
	}
}

func TestServiceUploadsDOCXAndExtractsDocumentText(t *testing.T) {
	var document bytes.Buffer
	writer := zip.NewWriter(&document)
	part, err := writer.Create("word/document.xml")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = part.Write([]byte(`<?xml version="1.0"?><w:document xmlns:w="urn:test"><w:body><w:p><w:r><w:t>第一段</w:t></w:r></w:p><w:p><w:r><w:t>第二段</w:t></w:r></w:p></w:body></w:document>`))
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	service := NewService(&fakeRepository{}, nil)

	file, err := service.UploadFile(context.Background(), UploadFileInput{
		UserID: 42, Name: "访谈.docx", MimeType: "application/vnd.openxmlformats-officedocument.wordprocessingml.document", Data: document.Bytes(),
	})

	if err != nil {
		t.Fatalf("UploadFile() error = %v", err)
	}
	if !strings.Contains(file.Content, "第一段") || !strings.Contains(file.Content, "第二段") {
		t.Fatalf("content = %q", file.Content)
	}
}

func TestServiceUploadFileRejectsUnsupportedExecutable(t *testing.T) {
	service := NewService(&fakeRepository{}, nil)

	_, err := service.UploadFile(context.Background(), UploadFileInput{
		UserID: 42, Name: "tool.exe", MimeType: "application/octet-stream", Data: []byte("MZbinary"),
	})

	if !errors.Is(err, ErrUnsupportedFileType) {
		t.Fatalf("err = %v, want ErrUnsupportedFileType", err)
	}
}

func TestServiceSendMessageIncludesSelectedReferenceFilesInPrompt(t *testing.T) {
	payload, _ := json.Marshal(chatAIResult{Reply: "建议优先补齐私域转化链路。"})
	repository := &fakeRepository{
		threads: []Thread{{ID: 99, UserID: 42, Title: "机会分析", Mode: ModeChat}},
		files: []File{{
			ID:        17,
			UserID:    42,
			Name:      "竞品对比.txt",
			MimeType:  "text/plain",
			SizeBytes: 64,
			Content:   "小鹅通：私域工具强；有赞教育：交易能力强。",
		}},
	}
	generator := &fakeGenerator{content: payload}
	service := NewService(repository, generator)

	_, err := service.SendMessage(context.Background(), SendMessageInput{
		UserID:       42,
		ThreadID:     99,
		Content:      "结合引用文件分析机会",
		Model:        "gpt-test",
		ReferenceIDs: []int64{17},
	})

	if err != nil {
		t.Fatalf("SendMessage() error = %v", err)
	}
	if !strings.Contains(generator.request.UserPrompt, "引用文件") ||
		!strings.Contains(generator.request.UserPrompt, "竞品对比.txt") ||
		!strings.Contains(generator.request.UserPrompt, "小鹅通") {
		t.Fatalf("prompt did not include references:\n%s", generator.request.UserPrompt)
	}
	if len(repository.createdMessages) == 0 || !strings.Contains(string(repository.createdMessages[0].Metadata), "reference_ids") {
		t.Fatalf("user message metadata = %s", repository.createdMessages[0].Metadata)
	}
}

func TestServiceSendMessageRejectsUnknownConfiguredModel(t *testing.T) {
	repository := &fakeRepository{threads: []Thread{{ID: 99, UserID: 42, Title: "机会分析", Mode: ModeChat}}}
	service := NewServiceWithModels(repository, &fakeGenerator{content: []byte(`{"reply":"ok"}`)}, []ModelOption{{
		Name:      "DeepSeek",
		Value:     "deepseek",
		Provider:  "openai-compatible",
		IsDefault: true,
	}})

	_, err := service.SendMessage(context.Background(), SendMessageInput{
		UserID:   42,
		ThreadID: 99,
		Content:  "帮我分析智能客服市场机会",
		Model:    "unknown-model",
	})

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("err = %v, want ErrInvalidInput", err)
	}
	if len(repository.createdMessages) != 0 {
		t.Fatalf("createdMessages = %+v, want none", repository.createdMessages)
	}
}

func TestServiceCompareMessagesGeneratesOneAnswerPerModel(t *testing.T) {
	firstPayload, _ := json.Marshal(chatAIResult{Reply: "先做教培客服场景。"})
	secondPayload, _ := json.Marshal(chatAIResult{Reply: "优先验证留资转化。"})
	repository := &fakeRepository{
		threads: []Thread{{ID: 99, UserID: 42, Title: "机会分析", Mode: ModeCompare}},
	}
	generator := &fakeGeneratorQueue{contents: [][]byte{firstPayload, secondPayload}}
	service := NewServiceWithModels(repository, generator, []ModelOption{
		{Name: "DeepSeek", Value: "deepseek", Provider: "openai-compatible", IsDefault: true},
		{Name: "GPT-4o", Value: "gpt-main", Provider: "openai-responses"},
	})

	result, err := service.CompareMessages(context.Background(), CompareMessagesInput{
		UserID:   42,
		ThreadID: 99,
		Content:  "帮我分析机会",
		Models:   []string{"deepseek", "gpt-main"},
	})

	if err != nil {
		t.Fatalf("CompareMessages() error = %v", err)
	}
	if result.UserMessage.Role != RoleUser || len(result.Answers) != 2 {
		t.Fatalf("result = %+v", result)
	}
	if result.Answers[0].Model != "deepseek" || result.Answers[1].Model != "gpt-main" {
		t.Fatalf("answers = %+v", result.Answers)
	}
	if len(generator.requests) != 2 || generator.requests[0].Model != "deepseek" || generator.requests[1].Model != "gpt-main" {
		t.Fatalf("requests = %+v", generator.requests)
	}
	if len(repository.createdMessages) != 3 {
		t.Fatalf("createdMessages = %+v", repository.createdMessages)
	}
	if string(repository.createdMessages[0].Metadata) != `{"kind":"compare_question"}` {
		t.Fatalf("question metadata = %s", repository.createdMessages[0].Metadata)
	}
	if string(repository.createdMessages[1].Metadata) != `{"kind":"compare_answer"}` ||
		string(repository.createdMessages[2].Metadata) != `{"kind":"compare_answer"}` {
		t.Fatalf("answer metadata = %s / %s", repository.createdMessages[1].Metadata, repository.createdMessages[2].Metadata)
	}
}

func TestServiceCompareMessagesConsumesOneQuotaUnitPerModel(t *testing.T) {
	repository := &fakeRepository{threads: []Thread{{ID: 99, UserID: 42, Title: "机会分析", Mode: ModeCompare}}}
	quota := &fakeQuotaConsumer{}
	service := NewServiceWithModels(repository, &fakeGeneratorQueue{contents: [][]byte{
		[]byte(`{"reply":"回答一"}`), []byte(`{"reply":"回答二"}`),
	}}, []ModelOption{
		{Name: "A", Value: "model-a", IsDefault: true},
		{Name: "B", Value: "model-b"},
	}, WithQuotaConsumer(quota))

	_, err := service.CompareMessages(context.Background(), CompareMessagesInput{
		UserID: 42, ThreadID: 99, Content: "分析机会", Models: []string{"model-a", "model-b"}, RequestID: "compare-001",
	})

	if err != nil {
		t.Fatalf("CompareMessages() error = %v", err)
	}
	if len(quota.consumed) != 1 {
		t.Fatalf("consumed = %+v", quota.consumed)
	}
	consume := quota.consumed[0]
	if consume.FeatureKey != membership.FeatureCopilotCompareCalls || consume.Amount != 2 || consume.IdempotencyKey != "copilot-compare-compare-001" {
		t.Fatalf("consume = %+v", consume)
	}
}

func TestServiceCompareFailureReturnsTransientPlaceholderWithoutPersistingIt(t *testing.T) {
	repository := &fakeRepository{threads: []Thread{{ID: 99, UserID: 42, Title: "机会分析", Mode: ModeCompare}}}
	service := NewServiceWithModels(repository, &fakeGeneratorQueue{err: errors.New("provider unavailable")}, []ModelOption{
		{Name: "A", Value: "model-a", IsDefault: true},
		{Name: "B", Value: "model-b"},
	})

	result, err := service.CompareMessages(context.Background(), CompareMessagesInput{
		UserID: 42, ThreadID: 99, Content: "分析机会", Models: []string{"model-a", "model-b"},
	})

	if err != nil {
		t.Fatalf("CompareMessages() error = %v", err)
	}
	if len(result.Answers) != 2 || result.Answers[0].AssistantMessage.Status != MessageStatusFailed || result.Answers[0].AssistantMessage.ID >= 0 {
		t.Fatalf("answers = %+v", result.Answers)
	}
	if len(repository.createdMessages) != 1 || repository.createdMessages[0].Role != RoleUser {
		t.Fatalf("failed comparisons persisted assistant messages: %+v", repository.createdMessages)
	}
}

func TestServiceCompareMessagesRejectsUnknownModelBeforeCreatingMessages(t *testing.T) {
	repository := &fakeRepository{threads: []Thread{{ID: 99, UserID: 42, Title: "机会分析", Mode: ModeCompare}}}
	service := NewServiceWithModels(repository, &fakeGenerator{content: []byte(`{"reply":"ok"}`)}, []ModelOption{{
		Name:      "DeepSeek",
		Value:     "deepseek",
		Provider:  "openai-compatible",
		IsDefault: true,
	}})

	_, err := service.CompareMessages(context.Background(), CompareMessagesInput{
		UserID:   42,
		ThreadID: 99,
		Content:  "帮我分析机会",
		Models:   []string{"deepseek", "missing"},
	})

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("err = %v, want ErrInvalidInput", err)
	}
	if len(repository.createdMessages) != 0 {
		t.Fatalf("createdMessages = %+v, want none", repository.createdMessages)
	}
}

func TestServiceSummarizesComparisonAnswers(t *testing.T) {
	payload, _ := json.Marshal(chatAIResult{Reply: "综合来看，先用 DeepSeek 做低成本验证，再用 GPT 做方案打磨。"})
	repository := &fakeRepository{
		threads: []Thread{{ID: 99, UserID: 42, Title: "机会分析", Mode: ModeCompare}},
	}
	generator := &fakeGenerator{content: payload}
	service := NewServiceWithModels(repository, generator, []ModelOption{
		{Name: "DeepSeek", Value: "deepseek", Provider: "openai-compatible", IsDefault: true},
		{Name: "GPT-4o", Value: "gpt-main", Provider: "openai-responses"},
	})

	result, err := service.SummarizeComparison(context.Background(), CompareSummaryInput{
		UserID:   42,
		ThreadID: 99,
		Content:  "帮我分析机会",
		Model:    "gpt-main",
		Answers: []CompareAnswer{
			{Model: "deepseek", AssistantMessage: Message{Content: "低成本验证。"}},
			{Model: "gpt-main", AssistantMessage: Message{Content: "打磨执行方案。"}},
		},
	})

	if err != nil {
		t.Fatalf("SummarizeComparison() error = %v", err)
	}
	if result.SummaryMessage.Role != RoleAssistant || result.SummaryMessage.Content == "" {
		t.Fatalf("summary message = %+v", result.SummaryMessage)
	}
	if generator.request.Feature != "copilot.compare_summary" || generator.request.Model != "gpt-main" {
		t.Fatalf("ai request = %+v", generator.request)
	}
	if len(repository.createdMessages) != 1 || repository.createdMessages[0].Model != "gpt-main" {
		t.Fatalf("createdMessages = %+v", repository.createdMessages)
	}
	if string(repository.createdMessages[0].Metadata) != `{"kind":"compare_summary"}` {
		t.Fatalf("summary metadata = %s", repository.createdMessages[0].Metadata)
	}
}

func TestServiceSummarizeComparisonRejectsBlankAnswers(t *testing.T) {
	repository := &fakeRepository{threads: []Thread{{ID: 99, UserID: 42, Title: "机会分析", Mode: ModeCompare}}}
	service := NewServiceWithModels(repository, &fakeGenerator{content: []byte(`{"reply":"ok"}`)}, []ModelOption{{
		Name:      "DeepSeek",
		Value:     "deepseek",
		Provider:  "openai-compatible",
		IsDefault: true,
	}})

	_, err := service.SummarizeComparison(context.Background(), CompareSummaryInput{
		UserID:   42,
		ThreadID: 99,
		Content:  "帮我分析机会",
		Answers:  []CompareAnswer{{Model: "deepseek", AssistantMessage: Message{Content: " "}}},
	})

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("err = %v, want ErrInvalidInput", err)
	}
}

func TestServiceSmokeModelUsesConfiguredAlias(t *testing.T) {
	generator := &fakeGenerator{content: []byte(`{"reply":"pong"}`)}
	service := NewServiceWithModels(&fakeRepository{}, generator, []ModelOption{{
		Name:      "DeepSeek",
		Value:     "deepseek",
		Provider:  "openai-compatible",
		IsDefault: true,
	}})

	result, err := service.SmokeModel(context.Background(), ModelSmokeInput{
		UserID: 42,
		Model:  "deepseek",
		Prompt: "ping",
	})

	if err != nil {
		t.Fatalf("SmokeModel() error = %v", err)
	}
	if !result.OK || result.Model != "deepseek" || result.Reply != "pong" {
		t.Fatalf("result = %+v", result)
	}
	if result.InputTokens != 12 || result.OutputTokens != 24 {
		t.Fatalf("tokens = %d/%d", result.InputTokens, result.OutputTokens)
	}
	if generator.request.Feature != "copilot.model_smoke" ||
		generator.request.PromptVersion != "copilot_model_smoke_v1" ||
		generator.request.Model != "deepseek" {
		t.Fatalf("ai request = %+v", generator.request)
	}
}

func TestServiceSmokeModelRejectsUnknownConfiguredModel(t *testing.T) {
	repository := &fakeRepository{}
	generator := &fakeGenerator{content: []byte(`{"reply":"pong"}`)}
	service := NewServiceWithModels(repository, generator, []ModelOption{{
		Name:      "DeepSeek",
		Value:     "deepseek",
		Provider:  "openai-compatible",
		IsDefault: true,
	}})

	_, err := service.SmokeModel(context.Background(), ModelSmokeInput{
		UserID: 42,
		Model:  "missing",
		Prompt: "ping",
	})

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("err = %v, want ErrInvalidInput", err)
	}
	if generator.request.Feature != "" {
		t.Fatalf("generator should not be called, request = %+v", generator.request)
	}
}

func TestServiceListAIRunsReturnsCopilotRunsForUser(t *testing.T) {
	now := time.Date(2026, 7, 1, 10, 25, 0, 0, time.UTC)
	repository := &fakeRepository{aiRuns: []ai.Run{{
		ID:            11,
		UserID:        42,
		Feature:       "copilot.model_smoke",
		PromptVersion: "copilot_model_smoke_v1",
		Provider:      "openai-compatible",
		Model:         "deepseek",
		Status:        ai.StatusFailed,
		ErrorCode:     "provider_unavailable",
		ErrorMessage:  "AI provider is unavailable",
		InputTokens:   12,
		OutputTokens:  0,
		LatencyMS:     2080,
		CreatedAt:     now,
		UpdatedAt:     now,
	}}}
	service := NewService(repository, nil)

	runs, err := service.ListAIRuns(context.Background(), 42, 500)

	if err != nil {
		t.Fatalf("ListAIRuns() error = %v", err)
	}
	if repository.listRunsUserID != 42 || repository.listRunsPrefix != "copilot." || repository.listRunsLimit != 100 {
		t.Fatalf("list args = user:%d prefix:%q limit:%d", repository.listRunsUserID, repository.listRunsPrefix, repository.listRunsLimit)
	}
	if len(runs) != 1 || runs[0].ID != 11 || runs[0].Model != "deepseek" || runs[0].ErrorCode != "provider_unavailable" {
		t.Fatalf("runs = %+v", runs)
	}
}

func TestServiceCreateThreadRejectsUnknownConfiguredModel(t *testing.T) {
	service := NewServiceWithModels(&fakeRepository{}, nil, []ModelOption{{
		Name:      "DeepSeek",
		Value:     "deepseek",
		Provider:  "openai-compatible",
		IsDefault: true,
	}})

	_, err := service.CreateThread(context.Background(), CreateThreadInput{
		UserID: 42,
		Title:  "机会分析",
		Model:  "unknown-model",
	})

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("err = %v, want ErrInvalidInput", err)
	}
}

func TestServiceSendMessageRejectsOtherUsersThread(t *testing.T) {
	repository := &fakeRepository{threads: []Thread{{ID: 99, UserID: 7, Title: "Other"}}}
	service := NewService(repository, &fakeGenerator{content: []byte(`{"reply":"ok"}`)})

	_, err := service.SendMessage(context.Background(), SendMessageInput{UserID: 42, ThreadID: 99, Content: "hi"})

	if !errors.Is(err, ErrThreadNotFound) {
		t.Fatalf("err = %v, want ErrThreadNotFound", err)
	}
}

func TestServiceReturnsSafeErrorForInvalidAIReply(t *testing.T) {
	repository := &fakeRepository{threads: []Thread{{ID: 99, UserID: 42, Title: "机会分析"}}}
	service := NewService(repository, &fakeGenerator{content: []byte(`{"reply":""}`)})

	_, err := service.SendMessage(context.Background(), SendMessageInput{UserID: 42, ThreadID: 99, Content: "hi"})

	if !errors.Is(err, ErrInvalidAIResult) {
		t.Fatalf("err = %v, want ErrInvalidAIResult", err)
	}
}
