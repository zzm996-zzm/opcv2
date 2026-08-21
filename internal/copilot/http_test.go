package copilot

import (
	"bytes"
	"context"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zzm/opcv2/internal/ai"
	"github.com/zzm/opcv2/internal/auth"
	"github.com/zzm/opcv2/internal/membership"
)

type fakeApplication struct {
	createThreadInput     CreateThreadInput
	sendMessageInput      SendMessageInput
	toolConfirmationInput ToolConfirmationInput
	compareInput          CompareMessagesInput
	summaryInput          CompareSummaryInput
	smokeInput            ModelSmokeInput
	memoryInput           MemoryInput
	memoryUpdateInput     MemoryUpdateInput
	fileInput             FileInput
	uploadInput           UploadFileInput
	fileID                int64
	userID                int64
	threadID              int64
	memoryID              int64
	limit                 int
	thread                Thread
	threads               []Thread
	messages              []Message
	memories              []Memory
	files                 []File
	models                []ModelOption
	aiRuns                []AIRun
	sendResult            SendMessageResult
	compareResult         CompareMessagesResult
	summaryResult         CompareSummaryResult
	smokeResult           ModelSmokeResult
	err                   error
}

func (a *fakeApplication) CreateThread(_ context.Context, input CreateThreadInput) (Thread, error) {
	a.createThreadInput = input
	return a.thread, a.err
}

func (a *fakeApplication) ListThreads(_ context.Context, userID int64, limit int) ([]Thread, error) {
	a.userID = userID
	a.limit = limit
	return a.threads, a.err
}

func (a *fakeApplication) GetThread(_ context.Context, userID, id int64) (Thread, error) {
	a.userID = userID
	a.threadID = id
	return a.thread, a.err
}

func (a *fakeApplication) RenameThread(_ context.Context, userID, id int64, title string) (Thread, error) {
	a.userID = userID
	a.threadID = id
	a.createThreadInput.Title = title
	return a.thread, a.err
}

func (a *fakeApplication) ArchiveThread(_ context.Context, userID, id int64) error {
	a.userID = userID
	a.threadID = id
	return a.err
}

func (a *fakeApplication) ListMessages(_ context.Context, userID, threadID int64, limit int) ([]Message, error) {
	a.userID = userID
	a.threadID = threadID
	a.limit = limit
	return a.messages, a.err
}

func (a *fakeApplication) SendMessage(_ context.Context, input SendMessageInput) (SendMessageResult, error) {
	a.sendMessageInput = input
	return a.sendResult, a.err
}

func (a *fakeApplication) ConfirmTool(_ context.Context, input ToolConfirmationInput) (ToolConfirmationResult, error) {
	a.toolConfirmationInput = input
	return ToolConfirmationResult{Message: Message{ID: input.MessageID, ThreadID: input.ThreadID, UserID: input.UserID}}, a.err
}

func (a *fakeApplication) StreamMessage(_ context.Context, input SendMessageInput, onEvent func(StreamEvent) error) (SendMessageResult, error) {
	a.sendMessageInput = input
	user := Message{ID: 1, UserID: input.UserID, ThreadID: input.ThreadID, Role: RoleUser, Content: input.Content}
	assistant := Message{ID: 2, UserID: input.UserID, ThreadID: input.ThreadID, Role: RoleAssistant, Content: "实时回答"}
	_ = onEvent(StreamEvent{Type: StreamEventUserMessage, UserMessage: &user})
	_ = onEvent(StreamEvent{Type: StreamEventDelta, Delta: "实时回答"})
	_ = onEvent(StreamEvent{Type: StreamEventAssistantMessage, AssistantMessage: &assistant})
	return SendMessageResult{UserMessage: user, AssistantMessage: assistant}, a.err
}

func (a *fakeApplication) CompareMessages(_ context.Context, input CompareMessagesInput) (CompareMessagesResult, error) {
	a.compareInput = input
	return a.compareResult, a.err
}

func (a *fakeApplication) SummarizeComparison(_ context.Context, input CompareSummaryInput) (CompareSummaryResult, error) {
	a.summaryInput = input
	return a.summaryResult, a.err
}

func (a *fakeApplication) SmokeModel(_ context.Context, input ModelSmokeInput) (ModelSmokeResult, error) {
	a.smokeInput = input
	return a.smokeResult, a.err
}

func (a *fakeApplication) ListMemories(_ context.Context, userID int64, limit int) ([]Memory, error) {
	a.userID = userID
	a.limit = limit
	return a.memories, a.err
}

func (a *fakeApplication) SaveMemory(_ context.Context, input MemoryInput) (Memory, error) {
	a.memoryInput = input
	return Memory{ID: 7, UserID: input.UserID, Key: input.Key, Value: input.Value}, a.err
}

func (a *fakeApplication) UpdateMemory(_ context.Context, input MemoryUpdateInput) (Memory, error) {
	a.memoryUpdateInput = input
	return Memory{ID: input.ID, UserID: input.UserID, Key: input.Key, Value: input.Value, Status: input.Status}, a.err
}

func (a *fakeApplication) DeleteMemory(_ context.Context, userID, id int64) error {
	a.userID = userID
	a.memoryID = id
	return a.err
}

func (a *fakeApplication) ListFiles(_ context.Context, userID int64, limit int) ([]File, error) {
	a.userID = userID
	a.limit = limit
	return a.files, a.err
}

func (a *fakeApplication) SaveFile(_ context.Context, input FileInput) (File, error) {
	a.fileInput = input
	return File{ID: 17, UserID: input.UserID, Name: input.Name, MimeType: input.MimeType, Content: input.Content}, a.err
}

func (a *fakeApplication) UploadFile(_ context.Context, input UploadFileInput) (File, error) {
	a.uploadInput = input
	return File{ID: 18, UserID: input.UserID, Name: input.Name, MimeType: input.MimeType, SizeBytes: len(input.Data)}, a.err
}

func (a *fakeApplication) GetFile(_ context.Context, userID, id int64) (File, error) {
	a.userID = userID
	a.fileID = id
	return File{ID: id, UserID: userID, Name: "客户访谈.txt", Content: "正文", Status: FileStatusReady}, a.err
}

func (a *fakeApplication) DeleteFile(_ context.Context, userID, id int64) error {
	a.userID = userID
	a.fileID = id
	return a.err
}

func (a *fakeApplication) ListModels(_ context.Context) ([]ModelOption, error) {
	return a.models, a.err
}

func (a *fakeApplication) ListAIRuns(_ context.Context, userID int64, limit int) ([]AIRun, error) {
	a.userID = userID
	a.limit = limit
	return a.aiRuns, a.err
}

func copilotTestRouter(app Application) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	group := router.Group("/api/v1")
	group.Use(func(c *gin.Context) {
		c.Set(auth.UserIDContextKey, int64(42))
		c.Next()
	})
	NewHTTPHandler(app).Register(group)
	return router
}

func TestCreateThreadEndpointUsesAuthenticatedUser(t *testing.T) {
	app := &fakeApplication{thread: Thread{ID: 99, UserID: 42, Title: "新会话", Mode: ModeChat}}
	router := copilotTestRouter(app)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/copilot/threads", strings.NewReader(`{"title":"新会话"}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.createThreadInput.UserID != 42 || app.createThreadInput.Title != "新会话" {
		t.Fatalf("input = %+v", app.createThreadInput)
	}
}

func TestListThreadsEndpointReturnsEmptyArrayAndCapsLimit(t *testing.T) {
	app := &fakeApplication{}
	router := copilotTestRouter(app)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/copilot/threads?limit=500", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.userID != 42 || app.limit != 100 {
		t.Fatalf("user/limit = %d/%d", app.userID, app.limit)
	}
	if !strings.Contains(recorder.Body.String(), `"threads":[]`) {
		t.Fatalf("body = %s", recorder.Body.String())
	}
}

func TestSendMessageEndpointUsesAuthenticatedUserAndThread(t *testing.T) {
	app := &fakeApplication{sendResult: SendMessageResult{
		UserMessage:      Message{ID: 1, UserID: 42, ThreadID: 99, Role: RoleUser, Content: "你好"},
		AssistantMessage: Message{ID: 2, UserID: 42, ThreadID: 99, Role: RoleAssistant, Content: "你好，我可以帮你分析项目。"},
	}}
	router := copilotTestRouter(app)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/copilot/threads/99/messages", strings.NewReader(`{
		"content":"你好",
		"model":"gpt-test"
	}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.sendMessageInput.UserID != 42 || app.sendMessageInput.ThreadID != 99 || app.sendMessageInput.Content != "你好" {
		t.Fatalf("input = %+v", app.sendMessageInput)
	}
}

func TestConfirmToolEndpointBindsAuthenticatedUserThreadAndMessage(t *testing.T) {
	app := &fakeApplication{}
	router := copilotTestRouter(app)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/copilot/threads/99/messages/17/tool-confirmation", strings.NewReader(`{"decision":"confirm"}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	input := app.toolConfirmationInput
	if input.UserID != 42 || input.ThreadID != 99 || input.MessageID != 17 || input.Decision != "confirm" {
		t.Fatalf("input = %+v", input)
	}
}

func TestStreamMessageEndpointWritesSSEEvents(t *testing.T) {
	app := &fakeApplication{}
	router := copilotTestRouter(app)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/copilot/threads/99/messages/stream", strings.NewReader(`{"content":"分析机会","request_id":"stream-001"}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Header().Get("Content-Type"), "text/event-stream") {
		t.Fatalf("status/content-type = %d/%s", recorder.Code, recorder.Header().Get("Content-Type"))
	}
	body := recorder.Body.String()
	if !strings.Contains(body, "event: delta") || !strings.Contains(body, `"delta":"实时回答"`) || !strings.Contains(body, "event: done") {
		t.Fatalf("body = %s", body)
	}
}

func TestSendMessageEndpointReturnsPaymentRequiredWhenQuotaExceeded(t *testing.T) {
	router := copilotTestRouter(&fakeApplication{err: membership.ErrQuotaExceeded})
	request := httptest.NewRequest(http.MethodPost, "/api/v1/copilot/threads/99/messages", strings.NewReader(`{"content":"你好","request_id":"msg-001"}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusPaymentRequired || !strings.Contains(recorder.Body.String(), `"error":"quota_exceeded"`) {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestSendMessageEndpointBindsReferenceIDs(t *testing.T) {
	app := &fakeApplication{sendResult: SendMessageResult{
		UserMessage:      Message{ID: 1, UserID: 42, ThreadID: 99, Role: RoleUser, Content: "你好"},
		AssistantMessage: Message{ID: 2, UserID: 42, ThreadID: 99, Role: RoleAssistant, Content: "你好，我可以帮你分析项目。"},
	}}
	router := copilotTestRouter(app)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/copilot/threads/99/messages", strings.NewReader(`{
		"content":"结合文件分析",
		"model":"gpt-test",
		"reference_ids":[17,18]
	}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if len(app.sendMessageInput.ReferenceIDs) != 2 || app.sendMessageInput.ReferenceIDs[0] != 17 || app.sendMessageInput.ReferenceIDs[1] != 18 {
		t.Fatalf("reference IDs = %+v", app.sendMessageInput.ReferenceIDs)
	}
}

func TestSendMessageEndpointRejectsBlankContent(t *testing.T) {
	app := &fakeApplication{}
	router := copilotTestRouter(app)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/copilot/threads/99/messages", strings.NewReader(`{"content":" "}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.sendMessageInput.UserID != 0 {
		t.Fatalf("SendMessage should not be called, input = %+v", app.sendMessageInput)
	}
}

func TestCopilotFilesEndpointsUseAuthenticatedUser(t *testing.T) {
	app := &fakeApplication{files: []File{{
		ID:        17,
		UserID:    42,
		Name:      "竞品对比.txt",
		MimeType:  "text/plain",
		SizeBytes: 48,
		Content:   "小鹅通：私域工具强。",
		CreatedAt: time.Date(2026, 7, 2, 9, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 7, 2, 9, 0, 0, 0, time.UTC),
	}}}
	router := copilotTestRouter(app)

	listRequest := httptest.NewRequest(http.MethodGet, "/api/v1/copilot/files?limit=500", nil)
	listRecorder := httptest.NewRecorder()
	router.ServeHTTP(listRecorder, listRequest)

	if listRecorder.Code != http.StatusOK {
		t.Fatalf("list status = %d body=%s", listRecorder.Code, listRecorder.Body.String())
	}
	if app.userID != 42 || app.limit != 100 {
		t.Fatalf("user/limit = %d/%d", app.userID, app.limit)
	}
	if !strings.Contains(listRecorder.Body.String(), `"files"`) {
		t.Fatalf("body = %s", listRecorder.Body.String())
	}

	saveRequest := httptest.NewRequest(http.MethodPost, "/api/v1/copilot/files", strings.NewReader(`{
		"name":"竞品对比.txt",
		"mime_type":"text/plain",
		"content":"小鹅通：私域工具强。"
	}`))
	saveRequest.Header.Set("Content-Type", "application/json")
	saveRecorder := httptest.NewRecorder()
	router.ServeHTTP(saveRecorder, saveRequest)

	if saveRecorder.Code != http.StatusOK {
		t.Fatalf("save status = %d body=%s", saveRecorder.Code, saveRecorder.Body.String())
	}
	if app.fileInput.UserID != 42 || app.fileInput.Name != "竞品对比.txt" || app.fileInput.Content == "" {
		t.Fatalf("file input = %+v", app.fileInput)
	}
}

func TestUploadCopilotFileEndpointReadsMultipartFile(t *testing.T) {
	app := &fakeApplication{}
	router := copilotTestRouter(app)
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", "客户访谈.txt")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = part.Write([]byte("客户关注交付周期。"))
	_ = writer.Close()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/copilot/files/upload", &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.uploadInput.UserID != 42 || app.uploadInput.Name != "客户访谈.txt" || string(app.uploadInput.Data) != "客户关注交付周期。" {
		t.Fatalf("uploadInput = %+v", app.uploadInput)
	}
}

func TestCopilotFileDetailAndDeleteEndpointsUseAuthenticatedUser(t *testing.T) {
	app := &fakeApplication{}
	router := copilotTestRouter(app)
	detailRecorder := httptest.NewRecorder()
	router.ServeHTTP(detailRecorder, httptest.NewRequest(http.MethodGet, "/api/v1/copilot/files/17", nil))
	if detailRecorder.Code != http.StatusOK || !strings.Contains(detailRecorder.Body.String(), `"content":"正文"`) {
		t.Fatalf("detail status = %d body=%s", detailRecorder.Code, detailRecorder.Body.String())
	}
	deleteRecorder := httptest.NewRecorder()
	router.ServeHTTP(deleteRecorder, httptest.NewRequest(http.MethodDelete, "/api/v1/copilot/files/17", nil))
	if deleteRecorder.Code != http.StatusNoContent || app.userID != 42 || app.fileID != 17 {
		t.Fatalf("delete status = %d user/file = %d/%d", deleteRecorder.Code, app.userID, app.fileID)
	}
}

func TestCompareMessagesEndpointUsesAuthenticatedUserThreadAndModels(t *testing.T) {
	app := &fakeApplication{compareResult: CompareMessagesResult{
		UserMessage: Message{ID: 1, UserID: 42, ThreadID: 99, Role: RoleUser, Content: "分析机会"},
		Answers: []CompareAnswer{{
			Model:            "deepseek",
			AssistantMessage: Message{ID: 2, UserID: 42, ThreadID: 99, Role: RoleAssistant, Content: "先做验证"},
		}},
	}}
	router := copilotTestRouter(app)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/copilot/threads/99/compare", strings.NewReader(`{
		"content":"分析机会",
		"models":["deepseek","gpt-main"]
	}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.compareInput.UserID != 42 || app.compareInput.ThreadID != 99 || app.compareInput.Content != "分析机会" {
		t.Fatalf("input = %+v", app.compareInput)
	}
	if len(app.compareInput.Models) != 2 || app.compareInput.Models[0] != "deepseek" || app.compareInput.Models[1] != "gpt-main" {
		t.Fatalf("models = %+v", app.compareInput.Models)
	}
	if !strings.Contains(recorder.Body.String(), `"answers"`) {
		t.Fatalf("body = %s", recorder.Body.String())
	}
}

func TestCompareSummaryEndpointUsesAuthenticatedUserThreadAndAnswers(t *testing.T) {
	app := &fakeApplication{summaryResult: CompareSummaryResult{
		SummaryMessage: Message{ID: 3, UserID: 42, ThreadID: 99, Role: RoleAssistant, Content: "综合建议先做验证。"},
	}}
	router := copilotTestRouter(app)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/copilot/threads/99/compare/summary", strings.NewReader(`{
		"content":"分析机会",
		"model":"gpt-main",
		"answers":[{"model":"deepseek","assistant_message":{"content":"先做验证。"}}]
	}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.summaryInput.UserID != 42 || app.summaryInput.ThreadID != 99 || app.summaryInput.Model != "gpt-main" {
		t.Fatalf("input = %+v", app.summaryInput)
	}
	if len(app.summaryInput.Answers) != 1 || app.summaryInput.Answers[0].Model != "deepseek" {
		t.Fatalf("answers = %+v", app.summaryInput.Answers)
	}
	if !strings.Contains(recorder.Body.String(), `"summary_message"`) {
		t.Fatalf("body = %s", recorder.Body.String())
	}
}

func TestMemoryEndpointsUseAuthenticatedUser(t *testing.T) {
	app := &fakeApplication{}
	router := copilotTestRouter(app)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/copilot/memories", strings.NewReader(`{
		"key":"industry",
		"value":"教培",
		"confidence":0.9,
		"source":"manual"
	}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.memoryInput.UserID != 42 || app.memoryInput.Key != "industry" {
		t.Fatalf("input = %+v", app.memoryInput)
	}
}

func TestUpdateMemoryEndpointUsesAuthenticatedUser(t *testing.T) {
	app := &fakeApplication{}
	router := copilotTestRouter(app)
	request := httptest.NewRequest(http.MethodPatch, "/api/v1/copilot/memories/7", strings.NewReader(`{"status":"active"}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.memoryUpdateInput.UserID != 42 || app.memoryUpdateInput.ID != 7 || app.memoryUpdateInput.Status != MemoryStatusActive {
		t.Fatalf("input = %+v", app.memoryUpdateInput)
	}
}

func TestListModelsEndpointReturnsConfiguredModels(t *testing.T) {
	app := &fakeApplication{models: []ModelOption{
		{Name: "DeepSeek", Value: "deepseek", Provider: "openai-compatible", IsDefault: true},
		{Name: "GPT-4o", Value: "gpt-main", Provider: "openai-responses"},
	}}
	router := copilotTestRouter(app)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/copilot/models", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"value":"deepseek"`) ||
		!strings.Contains(recorder.Body.String(), `"is_default":true`) {
		t.Fatalf("body = %s", recorder.Body.String())
	}
}

func TestSmokeModelEndpointUsesAuthenticatedUser(t *testing.T) {
	app := &fakeApplication{smokeResult: ModelSmokeResult{
		OK:           true,
		Model:        "deepseek",
		Reply:        "pong",
		InputTokens:  1,
		OutputTokens: 1,
	}}
	router := copilotTestRouter(app)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/copilot/models/smoke", strings.NewReader(`{
		"model":"deepseek",
		"prompt":"ping"
	}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.smokeInput.UserID != 42 || app.smokeInput.Model != "deepseek" || app.smokeInput.Prompt != "ping" {
		t.Fatalf("input = %+v", app.smokeInput)
	}
	if !strings.Contains(recorder.Body.String(), `"ok":true`) || !strings.Contains(recorder.Body.String(), `"reply":"pong"`) {
		t.Fatalf("body = %s", recorder.Body.String())
	}
}

func TestListAIRunsEndpointUsesAuthenticatedUser(t *testing.T) {
	app := &fakeApplication{aiRuns: []AIRun{{
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
		CreatedAt:     time.Date(2026, 7, 1, 10, 25, 0, 0, time.UTC),
		UpdatedAt:     time.Date(2026, 7, 1, 10, 25, 2, 0, time.UTC),
	}}}
	router := copilotTestRouter(app)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/copilot/ai-runs?limit=500", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.userID != 42 || app.limit != 100 {
		t.Fatalf("user/limit = %d/%d", app.userID, app.limit)
	}
	body := recorder.Body.String()
	if !strings.Contains(body, `"feature":"copilot.model_smoke"`) ||
		!strings.Contains(body, `"error_code":"provider_unavailable"`) {
		t.Fatalf("body = %s", body)
	}
	if strings.Contains(body, `"request"`) || strings.Contains(body, `"response"`) {
		t.Fatalf("body leaked request/response: %s", body)
	}
}

func TestListMessagesEndpointReturnsMetadata(t *testing.T) {
	app := &fakeApplication{messages: []Message{{
		ID:       7,
		UserID:   42,
		ThreadID: 99,
		Role:     RoleAssistant,
		Content:  "综合结论",
		Status:   MessageStatusCompleted,
		Metadata: []byte(`{"kind":"compare_summary"}`),
	}}}
	router := copilotTestRouter(app)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/copilot/threads/99/messages", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"metadata":{"kind":"compare_summary"}`) {
		t.Fatalf("body = %s", recorder.Body.String())
	}
}
