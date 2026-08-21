package copilot

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/zzm/opcv2/internal/auth"
	"github.com/zzm/opcv2/internal/membership"
	"github.com/zzm/opcv2/internal/platform/httpapi"
)

type Application interface {
	CreateThread(ctx context.Context, input CreateThreadInput) (Thread, error)
	ListThreads(ctx context.Context, userID int64, limit int) ([]Thread, error)
	GetThread(ctx context.Context, userID, id int64) (Thread, error)
	RenameThread(ctx context.Context, userID, id int64, title string) (Thread, error)
	ArchiveThread(ctx context.Context, userID, id int64) error
	ListMessages(ctx context.Context, userID, threadID int64, limit int) ([]Message, error)
	SendMessage(ctx context.Context, input SendMessageInput) (SendMessageResult, error)
	ConfirmTool(ctx context.Context, input ToolConfirmationInput) (ToolConfirmationResult, error)
	CompareMessages(ctx context.Context, input CompareMessagesInput) (CompareMessagesResult, error)
	SummarizeComparison(ctx context.Context, input CompareSummaryInput) (CompareSummaryResult, error)
	SmokeModel(ctx context.Context, input ModelSmokeInput) (ModelSmokeResult, error)
	ListMemories(ctx context.Context, userID int64, limit int) ([]Memory, error)
	SaveMemory(ctx context.Context, input MemoryInput) (Memory, error)
	UpdateMemory(ctx context.Context, input MemoryUpdateInput) (Memory, error)
	DeleteMemory(ctx context.Context, userID, id int64) error
	ListFiles(ctx context.Context, userID int64, limit int) ([]File, error)
	SaveFile(ctx context.Context, input FileInput) (File, error)
	UploadFile(ctx context.Context, input UploadFileInput) (File, error)
	GetFile(ctx context.Context, userID, id int64) (File, error)
	DeleteFile(ctx context.Context, userID, id int64) error
	ListModels(ctx context.Context) ([]ModelOption, error)
	ListAIRuns(ctx context.Context, userID int64, limit int) ([]AIRun, error)
}

type StreamingApplication interface {
	StreamMessage(ctx context.Context, input SendMessageInput, onEvent func(StreamEvent) error) (SendMessageResult, error)
}

type HTTPHandler struct {
	app Application
}

func NewHTTPHandler(app Application) *HTTPHandler {
	return &HTTPHandler{app: app}
}

func (h *HTTPHandler) Register(router *gin.RouterGroup) {
	router.POST("/copilot/threads", h.createThread)
	router.GET("/copilot/threads", h.listThreads)
	router.GET("/copilot/threads/:id", h.getThread)
	router.PATCH("/copilot/threads/:id", h.renameThread)
	router.DELETE("/copilot/threads/:id", h.archiveThread)
	router.GET("/copilot/threads/:id/messages", h.listMessages)
	router.POST("/copilot/threads/:id/messages", h.sendMessage)
	router.POST("/copilot/threads/:id/messages/stream", h.sendMessageStream)
	router.POST("/copilot/threads/:id/messages/:message_id/tool-confirmation", h.confirmTool)
	router.POST("/copilot/threads/:id/compare", h.compareMessages)
	router.POST("/copilot/threads/:id/compare/summary", h.summarizeComparison)
	router.GET("/copilot/memories", h.listMemories)
	router.POST("/copilot/memories", h.saveMemory)
	router.PATCH("/copilot/memories/:id", h.updateMemory)
	router.DELETE("/copilot/memories/:id", h.deleteMemory)
	router.GET("/copilot/files", h.listFiles)
	router.POST("/copilot/files", h.saveFile)
	router.POST("/copilot/files/upload", h.uploadFile)
	router.GET("/copilot/files/:id", h.getFile)
	router.DELETE("/copilot/files/:id", h.deleteFile)
	router.GET("/copilot/models", h.listModels)
	router.POST("/copilot/models/smoke", h.smokeModel)
	router.GET("/copilot/ai-runs", h.listAIRuns)
}

func (h *HTTPHandler) createThread(c *gin.Context) {
	var request CreateThreadInput
	if err := c.ShouldBindJSON(&request); err != nil {
		httpapi.BadRequest(c, "invalid_request")
		return
	}
	request.UserID = c.GetInt64(auth.UserIDContextKey)
	thread, err := h.app.CreateThread(c.Request.Context(), request)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, thread)
}

func (h *HTTPHandler) listThreads(c *gin.Context) {
	limit, ok := httpapi.QueryLimit(c, 20, 100)
	if !ok {
		return
	}
	threads, err := h.app.ListThreads(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), limit)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"threads": httpapi.EnsureSlice(threads)})
}

func (h *HTTPHandler) getThread(c *gin.Context) {
	id, ok := parseID(c, "id", "invalid_thread_id")
	if !ok {
		return
	}
	thread, err := h.app.GetThread(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, thread)
}

func (h *HTTPHandler) renameThread(c *gin.Context) {
	id, ok := parseID(c, "id", "invalid_thread_id")
	if !ok {
		return
	}
	var request struct {
		Title string `json:"title"`
	}
	if err := c.ShouldBindJSON(&request); err != nil || strings.TrimSpace(request.Title) == "" {
		httpapi.BadRequest(c, "invalid_request")
		return
	}
	thread, err := h.app.RenameThread(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), id, request.Title)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, thread)
}

func (h *HTTPHandler) archiveThread(c *gin.Context) {
	id, ok := parseID(c, "id", "invalid_thread_id")
	if !ok {
		return
	}
	if err := h.app.ArchiveThread(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), id); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *HTTPHandler) listMessages(c *gin.Context) {
	id, ok := parseID(c, "id", "invalid_thread_id")
	if !ok {
		return
	}
	limit, ok := httpapi.QueryLimit(c, 50, 100)
	if !ok {
		return
	}
	messages, err := h.app.ListMessages(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), id, limit)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"messages": httpapi.EnsureSlice(messages)})
}

func (h *HTTPHandler) sendMessage(c *gin.Context) {
	id, ok := parseID(c, "id", "invalid_thread_id")
	if !ok {
		return
	}
	var request SendMessageInput
	if err := c.ShouldBindJSON(&request); err != nil || strings.TrimSpace(request.Content) == "" {
		httpapi.BadRequest(c, "invalid_request")
		return
	}
	request.UserID = c.GetInt64(auth.UserIDContextKey)
	request.ThreadID = id
	result, err := h.app.SendMessage(c.Request.Context(), request)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *HTTPHandler) sendMessageStream(c *gin.Context) {
	id, ok := parseID(c, "id", "invalid_thread_id")
	if !ok {
		return
	}
	var request SendMessageInput
	if err := c.ShouldBindJSON(&request); err != nil || strings.TrimSpace(request.Content) == "" {
		httpapi.BadRequest(c, "invalid_request")
		return
	}
	streamingApp, ok := h.app.(StreamingApplication)
	if !ok {
		httpapi.Error(c, http.StatusNotImplemented, "streaming_not_supported")
		return
	}
	request.UserID = c.GetInt64(auth.UserIDContextKey)
	request.ThreadID = id
	c.Header("Content-Type", "text/event-stream; charset=utf-8")
	c.Header("Cache-Control", "no-cache, no-transform")
	c.Header("X-Accel-Buffering", "no")
	c.Status(http.StatusOK)
	_, err := streamingApp.StreamMessage(c.Request.Context(), request, func(event StreamEvent) error {
		return writeSSEEvent(c, event.Type, event)
	})
	if err != nil {
		_ = writeSSEEvent(c, "error", map[string]string{"error": streamErrorCode(err)})
		return
	}
	_ = writeSSEEvent(c, "done", map[string]bool{"ok": true})
}

func (h *HTTPHandler) confirmTool(c *gin.Context) {
	threadID, ok := parseID(c, "id", "invalid_thread_id")
	if !ok {
		return
	}
	messageID, ok := parseID(c, "message_id", "invalid_message_id")
	if !ok {
		return
	}
	var request ToolConfirmationInput
	if err := c.ShouldBindJSON(&request); err != nil {
		httpapi.BadRequest(c, "invalid_request")
		return
	}
	request.UserID = c.GetInt64(auth.UserIDContextKey)
	request.ThreadID = threadID
	request.MessageID = messageID
	result, err := h.app.ConfirmTool(c.Request.Context(), request)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func writeSSEEvent(c *gin.Context, event string, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(c.Writer, "event: %s\ndata: %s\n\n", event, data); err != nil {
		return err
	}
	c.Writer.Flush()
	return nil
}

func streamErrorCode(err error) string {
	switch {
	case errors.Is(err, membership.ErrQuotaExceeded):
		return "quota_exceeded"
	case errors.Is(err, ErrThreadNotFound):
		return "thread_not_found"
	case errors.Is(err, ErrFileNotFound):
		return "file_not_found"
	case errors.Is(err, ErrInvalidInput):
		return "invalid_request"
	default:
		return "stream_failed"
	}
}

func (h *HTTPHandler) compareMessages(c *gin.Context) {
	id, ok := parseID(c, "id", "invalid_thread_id")
	if !ok {
		return
	}
	var request CompareMessagesInput
	if err := c.ShouldBindJSON(&request); err != nil || strings.TrimSpace(request.Content) == "" {
		httpapi.BadRequest(c, "invalid_request")
		return
	}
	request.UserID = c.GetInt64(auth.UserIDContextKey)
	request.ThreadID = id
	result, err := h.app.CompareMessages(c.Request.Context(), request)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *HTTPHandler) summarizeComparison(c *gin.Context) {
	id, ok := parseID(c, "id", "invalid_thread_id")
	if !ok {
		return
	}
	var request CompareSummaryInput
	if err := c.ShouldBindJSON(&request); err != nil || strings.TrimSpace(request.Content) == "" {
		httpapi.BadRequest(c, "invalid_request")
		return
	}
	request.UserID = c.GetInt64(auth.UserIDContextKey)
	request.ThreadID = id
	result, err := h.app.SummarizeComparison(c.Request.Context(), request)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *HTTPHandler) listMemories(c *gin.Context) {
	limit, ok := httpapi.QueryLimit(c, 50, 100)
	if !ok {
		return
	}
	memories, err := h.app.ListMemories(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), limit)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"memories": httpapi.EnsureSlice(memories)})
}

func (h *HTTPHandler) saveMemory(c *gin.Context) {
	var request MemoryInput
	if err := c.ShouldBindJSON(&request); err != nil || strings.TrimSpace(request.Key) == "" || strings.TrimSpace(request.Value) == "" {
		httpapi.BadRequest(c, "invalid_request")
		return
	}
	request.UserID = c.GetInt64(auth.UserIDContextKey)
	memory, err := h.app.SaveMemory(c.Request.Context(), request)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, memory)
}

func (h *HTTPHandler) deleteMemory(c *gin.Context) {
	id, ok := parseID(c, "id", "invalid_memory_id")
	if !ok {
		return
	}
	if err := h.app.DeleteMemory(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), id); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *HTTPHandler) updateMemory(c *gin.Context) {
	id, ok := parseID(c, "id", "invalid_memory_id")
	if !ok {
		return
	}
	var request MemoryUpdateInput
	if err := c.ShouldBindJSON(&request); err != nil {
		httpapi.BadRequest(c, "invalid_request")
		return
	}
	request.UserID = c.GetInt64(auth.UserIDContextKey)
	request.ID = id
	memory, err := h.app.UpdateMemory(c.Request.Context(), request)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, memory)
}

func (h *HTTPHandler) listFiles(c *gin.Context) {
	limit, ok := httpapi.QueryLimit(c, 50, 100)
	if !ok {
		return
	}
	files, err := h.app.ListFiles(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), limit)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"files": httpapi.EnsureSlice(files)})
}

func (h *HTTPHandler) saveFile(c *gin.Context) {
	var request FileInput
	if err := c.ShouldBindJSON(&request); err != nil || strings.TrimSpace(request.Name) == "" || strings.TrimSpace(request.Content) == "" {
		httpapi.BadRequest(c, "invalid_request")
		return
	}
	request.UserID = c.GetInt64(auth.UserIDContextKey)
	file, err := h.app.SaveFile(c.Request.Context(), request)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, file)
}

func (h *HTTPHandler) uploadFile(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxUploadFileBytes+(1<<20))
	fileHeader, err := c.FormFile("file")
	if err != nil {
		var maxBytesError *http.MaxBytesError
		if errors.As(err, &maxBytesError) {
			writeError(c, ErrFileTooLarge)
			return
		}
		httpapi.BadRequest(c, "invalid_request")
		return
	}
	if c.Request.MultipartForm != nil {
		defer c.Request.MultipartForm.RemoveAll()
	}
	if fileHeader.Size > maxUploadFileBytes {
		writeError(c, ErrFileTooLarge)
		return
	}
	stream, err := fileHeader.Open()
	if err != nil {
		httpapi.BadRequest(c, "invalid_request")
		return
	}
	defer stream.Close()
	data, err := io.ReadAll(io.LimitReader(stream, maxUploadFileBytes+1))
	if err != nil {
		httpapi.BadRequest(c, "invalid_request")
		return
	}
	if len(data) > maxUploadFileBytes {
		writeError(c, ErrFileTooLarge)
		return
	}
	file, err := h.app.UploadFile(c.Request.Context(), UploadFileInput{
		UserID:   c.GetInt64(auth.UserIDContextKey),
		Name:     fileHeader.Filename,
		MimeType: fileHeader.Header.Get("Content-Type"),
		Data:     data,
	})
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, file)
}

func (h *HTTPHandler) getFile(c *gin.Context) {
	id, ok := parseID(c, "id", "invalid_file_id")
	if !ok {
		return
	}
	file, err := h.app.GetFile(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, file)
}

func (h *HTTPHandler) deleteFile(c *gin.Context) {
	id, ok := parseID(c, "id", "invalid_file_id")
	if !ok {
		return
	}
	if err := h.app.DeleteFile(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), id); err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *HTTPHandler) listModels(c *gin.Context) {
	models, err := h.app.ListModels(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"models": httpapi.EnsureSlice(models)})
}

func (h *HTTPHandler) smokeModel(c *gin.Context) {
	var request ModelSmokeInput
	if err := c.ShouldBindJSON(&request); err != nil {
		httpapi.BadRequest(c, "invalid_request")
		return
	}
	request.UserID = c.GetInt64(auth.UserIDContextKey)
	result, err := h.app.SmokeModel(c.Request.Context(), request)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *HTTPHandler) listAIRuns(c *gin.Context) {
	limit, ok := httpapi.QueryLimit(c, 20, 100)
	if !ok {
		return
	}
	runs, err := h.app.ListAIRuns(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), limit)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"runs": httpapi.EnsureSlice(runs)})
}

func parseID(c *gin.Context, name, errorCode string) (int64, bool) {
	id, err := strconv.ParseInt(c.Param(name), 10, 64)
	if err != nil || id <= 0 {
		httpapi.BadRequest(c, errorCode)
		return 0, false
	}
	return id, true
}

func writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrThreadNotFound):
		httpapi.Error(c, http.StatusNotFound, "thread_not_found")
	case errors.Is(err, ErrToolPreviewNotFound):
		httpapi.Error(c, http.StatusNotFound, "tool_preview_not_found")
	case errors.Is(err, ErrToolPreviewExpired):
		httpapi.Error(c, http.StatusConflict, "tool_preview_expired")
	case errors.Is(err, ErrToolPreviewConflict):
		httpapi.Error(c, http.StatusConflict, "tool_preview_conflict")
	case errors.Is(err, ErrMemoryNotFound):
		httpapi.Error(c, http.StatusNotFound, "memory_not_found")
	case errors.Is(err, ErrFileNotFound):
		httpapi.Error(c, http.StatusNotFound, "file_not_found")
	case errors.Is(err, ErrInvalidInput):
		httpapi.BadRequest(c, "invalid_request")
	case errors.Is(err, membership.ErrQuotaExceeded):
		httpapi.Error(c, http.StatusPaymentRequired, "quota_exceeded")
	case errors.Is(err, ErrFileTooLarge):
		httpapi.Error(c, http.StatusRequestEntityTooLarge, "file_too_large")
	case errors.Is(err, ErrUnsupportedFileType):
		httpapi.Error(c, http.StatusUnsupportedMediaType, "unsupported_file_type")
	case errors.Is(err, ErrInvalidFileEncoding):
		httpapi.BadRequest(c, "invalid_file_encoding")
	case errors.Is(err, ErrServiceNotReady):
		httpapi.Error(c, http.StatusInternalServerError, "service_not_ready")
	case errors.Is(err, ErrInvalidAIResult):
		httpapi.Error(c, http.StatusInternalServerError, "invalid_ai_result")
	default:
		httpapi.Error(c, http.StatusInternalServerError, "internal_error")
	}
}
