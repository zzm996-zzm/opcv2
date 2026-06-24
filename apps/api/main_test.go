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

func TestNewAIProviderRejectsUnsupportedProvider(t *testing.T) {
	if _, err := newAIProvider(config.Config{AIProvider: "anthropic"}); err == nil {
		t.Fatal("newAIProvider() error = nil, want unsupported provider error")
	}
}
