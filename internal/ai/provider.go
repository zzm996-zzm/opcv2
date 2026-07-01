package ai

import (
	"context"
	"errors"
)

var (
	ErrProviderTimeout        = errors.New("provider timeout")
	ErrProviderRateLimited    = errors.New("provider rate limited")
	ErrProviderAuthentication = errors.New("provider authentication failed")
	ErrProviderPermission     = errors.New("provider permission denied")
	ErrProviderModelNotFound  = errors.New("provider model not found")
	ErrProviderBadRequest     = errors.New("provider bad request")
	ErrProviderUnavailable    = errors.New("provider unavailable")
	ErrInvalidModelJSON       = errors.New("invalid model json")
	ErrServiceNotReady        = errors.New("ai service is not configured")
)

const (
	ErrorProviderTimeout        = "provider_timeout"
	ErrorProviderRateLimited    = "provider_rate_limited"
	ErrorProviderAuthentication = "provider_authentication_failed"
	ErrorProviderPermission     = "provider_permission_denied"
	ErrorProviderModelNotFound  = "provider_model_not_found"
	ErrorProviderBadRequest     = "provider_bad_request"
	ErrorProviderUnavailable    = "provider_unavailable"
	ErrorInvalidModelJSON       = "invalid_model_json"
	ErrorInternal               = "internal_error"
)

type Provider interface {
	Generate(ctx context.Context, request ProviderRequest) (ProviderResponse, error)
}

type ProviderRequest struct {
	Feature           string
	PromptVersion     string
	Model             string
	SystemPrompt      string
	UserPrompt        string
	SchemaName        string
	Attempt           int
	RepairInstruction string
}

type ProviderResponse struct {
	Content      []byte
	InputTokens  int
	OutputTokens int
}
