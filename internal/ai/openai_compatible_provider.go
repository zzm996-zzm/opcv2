package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type OpenAICompatibleConfig struct {
	BaseURL string
	APIKey  string
	Model   string
	Timeout time.Duration
}

type OpenAICompatibleProvider struct {
	baseURL string
	apiKey  string
	model   string
	client  *http.Client
}

func NewOpenAICompatibleProvider(config OpenAICompatibleConfig) *OpenAICompatibleProvider {
	timeout := config.Timeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	baseURL := strings.TrimRight(config.BaseURL, "/")
	return &OpenAICompatibleProvider{
		baseURL: baseURL,
		apiKey:  config.APIKey,
		model:   config.Model,
		client:  &http.Client{Timeout: timeout},
	}
}

func (p *OpenAICompatibleProvider) Generate(ctx context.Context, request ProviderRequest) (ProviderResponse, error) {
	body, err := json.Marshal(p.buildRequest(request))
	if err != nil {
		return ProviderResponse{}, fmt.Errorf("%w: encode request", ErrProviderUnavailable)
	}
	httpRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/chat/completions", bytes.NewReader(body))
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

	var payload openAICompatibleResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return ProviderResponse{}, fmt.Errorf("%w: decode response", ErrProviderUnavailable)
	}
	content := strings.TrimSpace(payload.firstContent())
	if content == "" {
		return ProviderResponse{}, fmt.Errorf("%w: empty response", ErrProviderUnavailable)
	}
	return ProviderResponse{
		Content:      []byte(content),
		InputTokens:  payload.Usage.PromptTokens,
		OutputTokens: payload.Usage.CompletionTokens,
	}, nil
}

func (p *OpenAICompatibleProvider) buildRequest(request ProviderRequest) openAICompatibleRequest {
	messages := []openAICompatibleMessage{
		{Role: "system", Content: request.SystemPrompt},
		{Role: "user", Content: request.UserPrompt},
	}
	if strings.TrimSpace(request.RepairInstruction) != "" {
		messages = append(messages, openAICompatibleMessage{Role: "user", Content: request.RepairInstruction})
	}
	model := strings.TrimSpace(request.Model)
	if model == "" {
		model = p.model
	}
	return openAICompatibleRequest{
		Model:    model,
		Messages: messages,
		ResponseFormat: openAICompatibleResponseFormat{
			Type: "json_object",
		},
	}
}

type openAICompatibleRequest struct {
	Model          string                         `json:"model"`
	Messages       []openAICompatibleMessage      `json:"messages"`
	ResponseFormat openAICompatibleResponseFormat `json:"response_format"`
}

type openAICompatibleMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAICompatibleResponseFormat struct {
	Type string `json:"type"`
}

type openAICompatibleResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	} `json:"usage"`
}

func (r openAICompatibleResponse) firstContent() string {
	for _, choice := range r.Choices {
		if choice.Message.Content != "" {
			return choice.Message.Content
		}
	}
	return ""
}
