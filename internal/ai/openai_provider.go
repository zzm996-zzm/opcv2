package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const defaultOpenAIBaseURL = "https://api.openai.com/v1"

type OpenAIConfig struct {
	BaseURL string
	APIKey  string
	Model   string
	Timeout time.Duration
}

type OpenAIProvider struct {
	baseURL string
	apiKey  string
	model   string
	client  *http.Client
}

func NewOpenAIProvider(config OpenAIConfig) *OpenAIProvider {
	timeout := config.Timeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	baseURL := strings.TrimRight(config.BaseURL, "/")
	if baseURL == "" {
		baseURL = defaultOpenAIBaseURL
	}
	return &OpenAIProvider{
		baseURL: baseURL,
		apiKey:  config.APIKey,
		model:   config.Model,
		client:  &http.Client{Timeout: timeout},
	}
}

func (p *OpenAIProvider) Generate(ctx context.Context, request ProviderRequest) (ProviderResponse, error) {
	body, err := json.Marshal(p.buildRequest(request))
	if err != nil {
		return ProviderResponse{}, fmt.Errorf("%w: encode request", ErrProviderUnavailable)
	}
	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/responses", bytes.NewReader(body))
	if err != nil {
		return ProviderResponse{}, fmt.Errorf("%w: build request", ErrProviderUnavailable)
	}
	httpRequest.Header.Set("Authorization", "Bearer "+p.apiKey)
	httpRequest.Header.Set("Content-Type", "application/json")

	response, err := p.client.Do(httpRequest)
	if err != nil {
		return ProviderResponse{}, classifyOpenAIClientError(err)
	}
	defer response.Body.Close()

	if err := classifyOpenAIStatus(response.StatusCode); err != nil {
		return ProviderResponse{}, err
	}

	var payload openAIResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return ProviderResponse{}, fmt.Errorf("%w: decode response", ErrProviderUnavailable)
	}
	content := strings.TrimSpace(payload.OutputText)
	if content == "" {
		content = strings.TrimSpace(payload.firstNestedText())
	}
	if content == "" {
		return ProviderResponse{}, fmt.Errorf("%w: empty response", ErrProviderUnavailable)
	}
	return ProviderResponse{
		Content:      []byte(content),
		InputTokens:  payload.Usage.InputTokens,
		OutputTokens: payload.Usage.OutputTokens,
	}, nil
}

func (p *OpenAIProvider) buildRequest(request ProviderRequest) openAIRequest {
	input := []openAIMessage{
		{Role: "system", Content: request.SystemPrompt},
		{Role: "user", Content: request.UserPrompt},
	}
	if strings.TrimSpace(request.RepairInstruction) != "" {
		input = append(input, openAIMessage{Role: "user", Content: request.RepairInstruction})
	}
	model := strings.TrimSpace(request.Model)
	if model == "" {
		model = p.model
	}
	return openAIRequest{
		Model: model,
		Input: input,
		Text: openAITextConfig{
			Format: openAITextFormat{Type: "json_object"},
		},
	}
}

type openAIRequest struct {
	Model string           `json:"model"`
	Input []openAIMessage  `json:"input"`
	Text  openAITextConfig `json:"text"`
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAITextConfig struct {
	Format openAITextFormat `json:"format"`
}

type openAITextFormat struct {
	Type string `json:"type"`
}

type openAIResponse struct {
	OutputText string `json:"output_text"`
	Output     []struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	} `json:"output"`
	Usage struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

func (r openAIResponse) firstNestedText() string {
	for _, output := range r.Output {
		for _, content := range output.Content {
			if content.Text != "" {
				return content.Text
			}
		}
	}
	return ""
}

func classifyOpenAIStatus(statusCode int) error {
	switch statusCode {
	case http.StatusBadRequest:
		return ErrProviderBadRequest
	case http.StatusUnauthorized:
		return ErrProviderAuthentication
	case http.StatusForbidden:
		return ErrProviderPermission
	case http.StatusNotFound:
		return ErrProviderModelNotFound
	case http.StatusTooManyRequests:
		return ErrProviderRateLimited
	}
	if statusCode < 200 || statusCode >= 300 {
		return ErrProviderUnavailable
	}
	return nil
}

func classifyOpenAIClientError(err error) error {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return ErrProviderTimeout
	}
	var urlErr *url.Error
	if errors.As(err, &urlErr) && urlErr.Timeout() {
		return ErrProviderTimeout
	}
	return ErrProviderUnavailable
}
