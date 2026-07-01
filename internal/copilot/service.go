package copilot

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/zzm/opcv2/internal/ai"
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
	ListRuns(ctx context.Context, userID int64, featurePrefix string, limit int) ([]ai.Run, error)
}

type JSONGenerator interface {
	GenerateJSON(ctx context.Context, request ai.GenerateJSONRequest) (ai.GenerateJSONResult, error)
}

type Service struct {
	repository Repository
	generator  JSONGenerator
	models     []ModelOption
	now        func() time.Time
}

type chatAIResult struct {
	Reply            string            `json:"reply"`
	MemoryCandidates []MemoryCandidate `json:"memory_candidates,omitempty"`
}

type modelSmokeAIResult struct {
	Reply string `json:"reply"`
}

const (
	messageKindCompareQuestion = "compare_question"
	messageKindCompareAnswer   = "compare_answer"
	messageKindCompareSummary  = "compare_summary"
)

func NewService(repository Repository, generator JSONGenerator) *Service {
	return &Service{repository: repository, generator: generator, now: time.Now}
}

func NewServiceWithModels(repository Repository, generator JSONGenerator, models []ModelOption) *Service {
	service := NewService(repository, generator)
	service.models = normalizeModelOptions(models)
	return service
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
	userMessage, err := s.repository.CreateMessage(ctx, Message{
		UserID:    input.UserID,
		ThreadID:  input.ThreadID,
		Role:      RoleUser,
		Content:   content,
		Status:    MessageStatusCompleted,
		Model:     model,
		CreatedAt: now,
	})
	if err != nil {
		return SendMessageResult{}, err
	}
	aiResult, result, err := s.generateReply(ctx, input.UserID, thread, content, model)
	if err != nil {
		_, _ = s.repository.CreateMessage(ctx, Message{
			UserID:    input.UserID,
			ThreadID:  input.ThreadID,
			Role:      RoleAssistant,
			Content:   "AI 回复暂时不可用，请稍后重试。",
			Status:    MessageStatusFailed,
			Model:     model,
			ErrorCode: "invalid_ai_result",
			CreatedAt: s.now(),
		})
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
		return CompareMessagesResult{}, err
	}

	answers := make([]CompareAnswer, 0, len(models))
	for _, model := range models {
		aiResult, result, err := s.generateReply(ctx, input.UserID, thread, content, model)
		if err != nil {
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
			return CompareMessagesResult{}, err
		}
		s.saveMemoryCandidates(ctx, input.UserID, result.MemoryCandidates)
		answers = append(answers, CompareAnswer{Model: model, AssistantMessage: assistantMessage})
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
	aiResult, result, err := s.generateSummary(ctx, input.UserID, prompt, model)
	if err != nil {
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

func (s *Service) generateReply(ctx context.Context, userID int64, thread Thread, content, model string) (ai.GenerateJSONResult, chatAIResult, error) {
	memories, _ := s.repository.ListMemories(ctx, userID, 20)
	messages, _ := s.repository.ListMessages(ctx, userID, thread.ID, 12)
	aiResult, err := s.generator.GenerateJSON(ctx, ai.GenerateJSONRequest{
		UserID:         userID,
		Feature:        "copilot.chat",
		PromptVersion:  "copilot_chat_v1",
		Model:          model,
		SystemPrompt:   "你是智活 Copilot，回答要直接、可执行。必须只返回 JSON，字段严格匹配 copilot_chat_response。",
		UserPrompt:     buildUserPrompt(thread, memories, messages, content),
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

func buildUserPrompt(thread Thread, memories []Memory, messages []Message, content string) string {
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
	return []byte(fmt.Sprintf(`{"kind":%q}`, kind))
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
