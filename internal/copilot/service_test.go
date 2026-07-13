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

	"github.com/zzm/opcv2/internal/ai"
	"github.com/zzm/opcv2/internal/membership"
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
	deltas  []string
	request ai.GenerateTextRequest
	err     error
}

func (g *fakeTextStreamer) GenerateJSON(_ context.Context, _ ai.GenerateJSONRequest) (ai.GenerateJSONResult, error) {
	return ai.GenerateJSONResult{}, errors.New("unexpected GenerateJSON call")
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
