package ai

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"
)

type Repository interface {
	CreateRun(ctx context.Context, run Run) (Run, error)
	CompleteRun(ctx context.Context, id int64, result RunResult) error
	FailRun(ctx context.Context, id int64, failure RunFailure) error
}

type Config struct {
	Provider string
	Model    string
	Logger   *slog.Logger
}

type Service struct {
	repository Repository
	provider   Provider
	config     Config
	now        func() time.Time
}

type GenerateJSONRequest struct {
	UserID         int64
	Feature        string
	PromptVersion  string
	SystemPrompt   string
	UserPrompt     string
	SchemaName     string
	Validate       func([]byte) error
	RepairAttempts int
}

type GenerateJSONResult struct {
	Content      []byte
	InputTokens  int
	OutputTokens int
}

func NewService(repository Repository, provider Provider, config Config) *Service {
	return &Service{
		repository: repository,
		provider:   provider,
		config:     config,
		now:        time.Now,
	}
}

func (s *Service) GenerateJSON(ctx context.Context, request GenerateJSONRequest) (GenerateJSONResult, error) {
	if s.repository == nil || s.provider == nil {
		return GenerateJSONResult{}, ErrServiceNotReady
	}
	if request.Validate == nil {
		return GenerateJSONResult{}, fmt.Errorf("validate json: %w", ErrInvalidModelJSON)
	}

	startedAt := s.now()
	run, err := s.repository.CreateRun(ctx, Run{
		UserID:        request.UserID,
		Feature:       request.Feature,
		PromptVersion: request.PromptVersion,
		Provider:      s.config.Provider,
		Model:         s.config.Model,
		Status:        StatusPending,
		Request:       buildRunRequestJSON(request),
		CreatedAt:     startedAt,
	})
	if err != nil {
		return GenerateJSONResult{}, err
	}

	result, failure, err := s.generateValidJSON(ctx, request)
	if err != nil {
		if failure.Code == "" {
			failure = RunFailure{
				Code:    failureCode(err),
				Message: safeFailureMessage(err),
			}
		}
		failure.LatencyMS = elapsedMS(startedAt, s.now())
		_ = s.repository.FailRun(ctx, run.ID, failure)
		s.logFailure(request, failure)
		return GenerateJSONResult{}, err
	}

	if err := s.repository.CompleteRun(ctx, run.ID, RunResult{
		Response:     result.Content,
		InputTokens:  result.InputTokens,
		OutputTokens: result.OutputTokens,
		LatencyMS:    elapsedMS(startedAt, s.now()),
	}); err != nil {
		return GenerateJSONResult{}, err
	}
	return result, nil
}

func (s *Service) generateValidJSON(ctx context.Context, request GenerateJSONRequest) (GenerateJSONResult, RunFailure, error) {
	maxAttempts := request.RepairAttempts + 1
	if maxAttempts < 1 {
		maxAttempts = 1
	}
	var lastValidationErr error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		providerRequest := ProviderRequest{
			Feature:       request.Feature,
			PromptVersion: request.PromptVersion,
			SystemPrompt:  request.SystemPrompt,
			UserPrompt:    request.UserPrompt,
			SchemaName:    request.SchemaName,
			Attempt:       attempt,
		}
		if attempt > 0 {
			providerRequest.RepairInstruction = "Return valid JSON only and match the expected schema. Previous response failed validation."
		}
		response, err := s.provider.Generate(ctx, providerRequest)
		if err != nil {
			return GenerateJSONResult{}, RunFailure{}, err
		}
		if err := request.Validate(response.Content); err != nil {
			lastValidationErr = err
			continue
		}
		return GenerateJSONResult{
			Content:      response.Content,
			InputTokens:  response.InputTokens,
			OutputTokens: response.OutputTokens,
		}, RunFailure{}, nil
	}
	return GenerateJSONResult{}, RunFailure{
		Code:    ErrorInvalidModelJSON,
		Message: "AI response did not match the expected schema",
	}, fmt.Errorf("%w: %v", ErrInvalidModelJSON, lastValidationErr)
}

func (s *Service) logFailure(request GenerateJSONRequest, failure RunFailure) {
	if s.config.Logger == nil {
		return
	}
	s.config.Logger.Warn("ai_run_failed",
		"user_id", request.UserID,
		"feature", request.Feature,
		"prompt_version", request.PromptVersion,
		"schema_name", request.SchemaName,
		"provider", s.config.Provider,
		"model", s.config.Model,
		"error_code", failure.Code,
		"latency_ms", failure.LatencyMS,
	)
}

func buildRunRequestJSON(request GenerateJSONRequest) []byte {
	return []byte(fmt.Sprintf(
		`{"feature":%q,"prompt_version":%q,"schema_name":%q}`,
		request.Feature,
		request.PromptVersion,
		request.SchemaName,
	))
}

func failureCode(err error) string {
	switch {
	case errors.Is(err, ErrProviderTimeout), errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		return ErrorProviderTimeout
	case errors.Is(err, ErrProviderRateLimited):
		return ErrorProviderRateLimited
	case errors.Is(err, ErrProviderUnavailable):
		return ErrorProviderUnavailable
	case errors.Is(err, ErrInvalidModelJSON):
		return ErrorInvalidModelJSON
	default:
		return ErrorInternal
	}
}

func safeFailureMessage(err error) string {
	switch {
	case errors.Is(err, ErrProviderTimeout), errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		return "AI provider timed out"
	case errors.Is(err, ErrProviderRateLimited):
		return "AI provider rate limit was reached"
	case errors.Is(err, ErrProviderUnavailable):
		return "AI provider is unavailable"
	case errors.Is(err, ErrInvalidModelJSON):
		return "AI response did not match the expected schema"
	default:
		return "AI generation failed"
	}
}

func elapsedMS(start, end time.Time) int {
	if end.Before(start) {
		return 0
	}
	return int(end.Sub(start).Milliseconds())
}
