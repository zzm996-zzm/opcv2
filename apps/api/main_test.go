package main

import (
	"testing"

	"github.com/zzm/opcv2/internal/ai"
	"github.com/zzm/opcv2/internal/platform/config"
)

func TestNewAIProviderUsesDevelopmentProvider(t *testing.T) {
	provider, err := newAIProvider(config.Config{AIProvider: "development"})
	if err != nil {
		t.Fatalf("newAIProvider() error = %v", err)
	}
	if _, ok := provider.(*ai.DevelopmentProvider); !ok {
		t.Fatalf("provider = %T, want *ai.DevelopmentProvider", provider)
	}
}

func TestNewAIProviderUsesOpenAIProvider(t *testing.T) {
	provider, err := newAIProvider(config.Config{
		AIProvider:       "openai",
		AIAPIKey:         "test-key",
		AIModel:          "gpt-test",
		AIBaseURL:        "https://api.example.test/v1",
		AITimeoutSeconds: 7,
	})
	if err != nil {
		t.Fatalf("newAIProvider() error = %v", err)
	}
	if _, ok := provider.(*ai.OpenAIProvider); !ok {
		t.Fatalf("provider = %T, want *ai.OpenAIProvider", provider)
	}
}

func TestNewAIProviderBuildsModelRouter(t *testing.T) {
	provider, err := newAIProvider(config.Config{
		AIModel: "deepseek",
		AIModelRoutes: []config.AIModelRoute{
			{
				Alias:    "deepseek",
				Provider: "openai-compatible",
				Model:    "deepseek-v4-flash",
				BaseURL:  "https://api.deepseek.com",
				APIKey:   "deepseek-key",
			},
			{
				Alias:    "gpt-main",
				Provider: "openai-responses",
				Model:    "gpt-4o",
				BaseURL:  "https://api.openai.com/v1",
				APIKey:   "openai-key",
			},
		},
	})
	if err != nil {
		t.Fatalf("newAIProvider() error = %v", err)
	}
	if _, ok := provider.(*ai.ModelRouter); !ok {
		t.Fatalf("provider = %T, want *ai.ModelRouter", provider)
	}
}

func TestNewAIProviderRejectsUnsupportedProvider(t *testing.T) {
	if _, err := newAIProvider(config.Config{AIProvider: "anthropic"}); err == nil {
		t.Fatal("newAIProvider() error = nil, want unsupported provider error")
	}
}

func TestCopilotModelOptionsUseConfiguredRoutes(t *testing.T) {
	options := copilotModelOptions(config.Config{
		AIModel: "deepseek",
		AIModelRoutes: []config.AIModelRoute{
			{Alias: "deepseek", Provider: "openai-compatible", Model: "deepseek-v4-flash"},
			{Alias: "gpt-main", Provider: "openai-responses", Model: "gpt-4o"},
		},
	})

	if len(options) != 2 {
		t.Fatalf("options = %+v, want 2", options)
	}
	if options[0].Value != "deepseek" || !options[0].IsDefault || options[1].Name != "GPT-4o" {
		t.Fatalf("options = %+v", options)
	}
}
