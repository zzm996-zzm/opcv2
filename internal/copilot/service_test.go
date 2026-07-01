package copilot

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/zzm/opcv2/internal/ai"
)

type fakeRepository struct {
	threads         []Thread
	messages        []Message
	memories        []Memory
	aiRuns          []ai.Run
	createdThread   Thread
	createdMessages []Message
	upsertedMemory  Memory
	deletedMemoryID int64
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
	r.upsertedMemory = memory
	memory.ID = 7
	r.memories = append(r.memories, memory)
	return memory, r.err
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
