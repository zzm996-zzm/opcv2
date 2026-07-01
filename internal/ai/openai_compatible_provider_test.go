package ai

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestOpenAICompatibleProviderMapsChatCompletionJSONRequestAndResponse(t *testing.T) {
	var path string
	var authorization string
	var payload map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		authorization = r.Header.Get("Authorization")
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"choices":[{"message":{"content":"{\"name\":\"DeepSeek JSON\"}"}}],
			"usage":{"prompt_tokens":12,"completion_tokens":34}
		}`))
	}))
	defer server.Close()

	provider := NewOpenAICompatibleProvider(OpenAICompatibleConfig{
		BaseURL: server.URL,
		APIKey:  "test-api-key",
		Model:   "deepseek-v4-flash",
		Timeout: time.Second,
	})
	response, err := provider.Generate(context.Background(), ProviderRequest{
		Model:             "deepseek-v4-flash",
		SystemPrompt:      "Return JSON only.",
		UserPrompt:        "match projects",
		RepairInstruction: "Return valid JSON only.",
	})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if path != "/chat/completions" {
		t.Fatalf("path = %q, want /chat/completions", path)
	}
	if authorization != "Bearer test-api-key" {
		t.Fatalf("Authorization = %q", authorization)
	}
	if payload["model"] != "deepseek-v4-flash" {
		t.Fatalf("model payload = %+v", payload)
	}
	if payload["response_format"].(map[string]any)["type"] != "json_object" {
		t.Fatalf("response_format payload = %+v", payload["response_format"])
	}
	messages := payload["messages"].([]any)
	if len(messages) != 3 {
		t.Fatalf("messages = %+v", messages)
	}
	if !strings.Contains(messages[2].(map[string]any)["content"].(string), "Return valid JSON only.") {
		t.Fatalf("repair message missing: %+v", messages)
	}
	if string(response.Content) != `{"name":"DeepSeek JSON"}` {
		t.Fatalf("Content = %s", response.Content)
	}
	if response.InputTokens != 12 || response.OutputTokens != 34 {
		t.Fatalf("tokens = %d/%d", response.InputTokens, response.OutputTokens)
	}
}

func TestOpenAICompatibleProviderClassifiesHTTPStatus(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		want       error
	}{
		{name: "bad request", statusCode: http.StatusBadRequest, want: ErrProviderBadRequest},
		{name: "unauthorized", statusCode: http.StatusUnauthorized, want: ErrProviderAuthentication},
		{name: "forbidden", statusCode: http.StatusForbidden, want: ErrProviderPermission},
		{name: "model not found", statusCode: http.StatusNotFound, want: ErrProviderModelNotFound},
		{name: "rate limited", statusCode: http.StatusTooManyRequests, want: ErrProviderRateLimited},
		{name: "provider unavailable", statusCode: http.StatusBadGateway, want: ErrProviderUnavailable},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(`{"error":{"message":"raw provider detail"}}`))
			}))
			defer server.Close()

			provider := NewOpenAICompatibleProvider(OpenAICompatibleConfig{
				BaseURL: server.URL,
				APIKey:  "key",
				Model:   "deepseek-v4-flash",
				Timeout: time.Second,
			})
			_, err := provider.Generate(context.Background(), ProviderRequest{SystemPrompt: "json", UserPrompt: "go"})
			if !errors.Is(err, tt.want) {
				t.Fatalf("Generate() error = %v, want %v", err, tt.want)
			}
		})
	}
}
