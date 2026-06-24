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

func TestNewAIProviderRejectsUnsupportedProvider(t *testing.T) {
	if _, err := newAIProvider(config.Config{AIProvider: "openai"}); err == nil {
		t.Fatal("newAIProvider() error = nil, want unsupported provider error")
	}
}
