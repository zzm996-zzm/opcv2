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
	"strings"
	"time"
	"unicode/utf8"

	"github.com/zzm/opcv2/internal/ai"
	"github.com/zzm/opcv2/internal/membership"
)

type Repository interface {
	CreateThread(ctx context.Context, thread Thread) (Thread, error)
	ListThreads(ctx context.Context, userID int64, limit int) ([]Thread, error)
	GetThread(ctx context.Context, userID, id int64) (Thread, error)
	UpdateThreadTitle(ctx context.Context, userID, id int64, title string) (Thread, error)
	ArchiveThread(ctx context.Context, userID, id int64) error
	CreateMessage(ctx context.Context, message Message) (Message, error)
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

type QuotaConsumer interface {
	CheckAndConsume(ctx context.Context, input membership.ConsumeInput) (membership.UsageItem, error)
	RefundUsage(ctx context.Context, input membership.ConsumeInput) (membership.UsageItem, error)
}

type Option func(*Service)

type Service struct {
	repository Repository
	generator  JSONGenerator
	models     []ModelOption
	quota      QuotaConsumer
	now        func() time.Time
}

type chatAIResult struct {
	Reply            string            `json:"reply"`
	MemoryCandidates []MemoryCandidate `json:"memory_candidates,omitempty"`
}

type modelSmokeAIResult struct {
	Reply string `json:"reply"`
}

type messageMetadata struct {
	Kind         string  `json:"kind,omitempty"`
	ReferenceIDs []int64 `json:"reference_ids,omitempty"`
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
)

func NewService(repository Repository, generator JSONGenerator, options ...Option) *Service {
	service := &Service{repository: repository, generator: generator, now: time.Now}
	for _, option := range options {
		option(service)
	}
	return service
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
		Metadata:  referenceMetadata(input.ReferenceIDs),
		CreatedAt: now,
	})
	if err != nil {
		s.refundQuota(ctx, input.UserID, membership.FeatureCopilotMessages, 1, quotaKey, "user-message")
		return SendMessageResult{}, err
	}
	aiResult, result, err := s.generateReply(ctx, input.UserID, thread, content, model, references)
	if err != nil {
		_, _ = s.repository.CreateMessage(ctx, Message{
			UserID:    input.UserID,
			ThreadID:  input.ThreadID,
			Role:      RoleAssistant,
			Content:   "AI 回复暂时不可用，请稍后重试。",
			Status:    MessageStatusFailed,
			Model:     model,
			ErrorCode: "invalid_ai_result",
			Metadata:  referenceMetadata(input.ReferenceIDs),
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
		aiResult, result, err := s.generateReply(ctx, input.UserID, thread, content, model, nil)
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

func (s *Service) generateReply(ctx context.Context, userID int64, thread Thread, content, model string, references []File) (ai.GenerateJSONResult, chatAIResult, error) {
	memories, _ := s.repository.ListMemories(ctx, userID, 20)
	messages, _ := s.repository.ListMessages(ctx, userID, thread.ID, 12)
	aiResult, err := s.generator.GenerateJSON(ctx, ai.GenerateJSONRequest{
		UserID:         userID,
		Feature:        "copilot.chat",
		PromptVersion:  "copilot_chat_v1",
		Model:          model,
		SystemPrompt:   "你是智活 Copilot，回答要直接、可执行。必须只返回 JSON，字段严格匹配 copilot_chat_response。",
		UserPrompt:     buildUserPrompt(thread, memories, messages, references, content),
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

func buildUserPrompt(thread Thread, memories []Memory, messages []Message, references []File, content string) string {
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
	builder.WriteString("\n用户当前问题：")
	builder.WriteString(content)
	builder.WriteString("\n\n返回 JSON：{\"reply\":\"...\",\"memory_candidates\":[{\"key\":\"...\",\"value\":\"...\",\"confidence\":0.9,\"source\":\"copilot\"}]}")
	return builder.String()
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

func memoryFromInput(input MemoryInput, now time.Time) (Memory, error) {
	key := strings.TrimSpace(input.Key)
	value := strings.TrimSpace(input.Value)
	if key == "" || value == "" {
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
