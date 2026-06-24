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

func TestOpenAIProviderMapsResponsesAPIRequestAndResponse(t *testing.T) {
	var method string
	var authorization string
	var payload map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method = r.Method
		authorization = r.Header.Get("Authorization")
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"output_text":"{\"name\":\"AI短视频脚本工作室\"}",
			"usage":{"input_tokens":12,"output_tokens":34}
		}`))
	}))
	defer server.Close()

	provider := NewOpenAIProvider(OpenAIConfig{
		BaseURL: server.URL,
		APIKey:  "test-api-key",
		Model:   "gpt-test",
		Timeout: time.Second,
	})
	response, err := provider.Generate(context.Background(), ProviderRequest{
		Feature:           "projects.match",
		PromptVersion:     "project_match_v1",
		SystemPrompt:      "Return JSON only.",
		UserPrompt:        "match projects",
		SchemaName:        "project_match_result",
		Attempt:           1,
		RepairInstruction: "Return valid JSON only.",
	})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if method != http.MethodPost {
		t.Fatalf("method = %q, want POST", method)
	}
	if authorization != "Bearer test-api-key" {
		t.Fatalf("Authorization = %q", authorization)
	}
	if payload["model"] != "gpt-test" {
		t.Fatalf("model payload = %+v", payload)
	}
	if payload["text"].(map[string]any)["format"].(map[string]any)["type"] != "json_object" {
		t.Fatalf("text format payload = %+v", payload["text"])
	}
	input := payload["input"].([]any)
	if len(input) != 3 {
		t.Fatalf("input messages = %+v", input)
	}
	if !strings.Contains(input[2].(map[string]any)["content"].(string), "Return valid JSON only.") {
		t.Fatalf("repair message missing: %+v", input)
	}
	if string(response.Content) != `{"name":"AI短视频脚本工作室"}` {
		t.Fatalf("Content = %s", response.Content)
	}
	if response.InputTokens != 12 || response.OutputTokens != 34 {
		t.Fatalf("tokens = %d/%d", response.InputTokens, response.OutputTokens)
	}
}

func TestOpenAIProviderExtractsNestedOutputText(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"output":[
				{"content":[{"type":"output_text","text":"{\"name\":\"嵌套结果\"}"}]}
			]
		}`))
	}))
	defer server.Close()

	provider := NewOpenAIProvider(OpenAIConfig{BaseURL: server.URL, APIKey: "key", Model: "gpt-test", Timeout: time.Second})
	response, err := provider.Generate(context.Background(), ProviderRequest{SystemPrompt: "json", UserPrompt: "go"})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if string(response.Content) != `{"name":"嵌套结果"}` {
		t.Fatalf("Content = %s", response.Content)
	}
}

func TestOpenAIProviderClassifiesHTTPStatus(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		want       error
	}{
		{name: "rate limited", statusCode: http.StatusTooManyRequests, want: ErrProviderRateLimited},
		{name: "provider unavailable", statusCode: http.StatusBadGateway, want: ErrProviderUnavailable},
		{name: "server error", statusCode: http.StatusInternalServerError, want: ErrProviderUnavailable},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(`{"error":{"message":"provider failed with raw detail"}}`))
			}))
			defer server.Close()

			provider := NewOpenAIProvider(OpenAIConfig{BaseURL: server.URL, APIKey: "key", Model: "gpt-test", Timeout: time.Second})
			_, err := provider.Generate(context.Background(), ProviderRequest{SystemPrompt: "json", UserPrompt: "go"})
			if !errors.Is(err, tt.want) {
				t.Fatalf("Generate() error = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestOpenAIProviderClassifiesTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond)
	}))
	defer server.Close()

	provider := NewOpenAIProvider(OpenAIConfig{BaseURL: server.URL, APIKey: "key", Model: "gpt-test", Timeout: time.Millisecond})
	_, err := provider.Generate(context.Background(), ProviderRequest{SystemPrompt: "json", UserPrompt: "go"})
	if !errors.Is(err, ErrProviderTimeout) {
		t.Fatalf("Generate() error = %v, want ErrProviderTimeout", err)
	}
}
