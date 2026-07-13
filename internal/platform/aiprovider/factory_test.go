package aiprovider

import (
	"testing"

	"github.com/zzm/opcv2/internal/ai"
	"github.com/zzm/opcv2/internal/platform/config"
)

func TestNewCreatesDevelopmentProviderAndModelRouter(t *testing.T) {
	provider, err := New(config.Config{AIProvider: "development", AIModel: "development-model"})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if _, ok := provider.(*ai.DevelopmentProvider); !ok {
		t.Fatalf("provider = %T", provider)
	}
	routed, err := New(config.Config{AIModel: "dev", AIModelRoutes: []config.AIModelRoute{{Alias: "dev", Provider: "development", Model: "development-model"}}})
	if err != nil {
		t.Fatalf("New(routes) error = %v", err)
	}
	if _, ok := routed.(*ai.ModelRouter); !ok {
		t.Fatalf("routed = %T", routed)
	}
}
