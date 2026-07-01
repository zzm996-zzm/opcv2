package ai

import (
	"context"
	"testing"
)

func TestModelRouterRoutesByAliasAndResolvesUpstreamModel(t *testing.T) {
	openaiProvider := &fakeProvider{responses: []ProviderResponse{{Content: []byte(`{"name":"openai"}`)}}}
	deepseekProvider := &fakeProvider{responses: []ProviderResponse{{Content: []byte(`{"name":"deepseek"}`)}}}
	router := NewModelRouter("deepseek", []ModelRoute{
		{Alias: "gpt-main", ProviderName: "openai", Model: "gpt-4o", Provider: openaiProvider},
		{Alias: "deepseek", ProviderName: "deepseek", Model: "deepseek-v4-flash", Provider: deepseekProvider},
	})

	response, err := router.Generate(context.Background(), ProviderRequest{Model: "gpt-main", UserPrompt: "json"})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if string(response.Content) != `{"name":"openai"}` {
		t.Fatalf("Content = %s", response.Content)
	}
	if len(openaiProvider.requests) != 1 || openaiProvider.requests[0].Model != "gpt-4o" {
		t.Fatalf("openai requests = %+v", openaiProvider.requests)
	}

	_, err = router.Generate(context.Background(), ProviderRequest{UserPrompt: "json"})
	if err != nil {
		t.Fatalf("Generate() default route error = %v", err)
	}
	if len(deepseekProvider.requests) != 1 || deepseekProvider.requests[0].Model != "deepseek-v4-flash" {
		t.Fatalf("deepseek requests = %+v", deepseekProvider.requests)
	}
}

func TestModelRouterRejectsUnknownAlias(t *testing.T) {
	router := NewModelRouter("deepseek", []ModelRoute{
		{Alias: "deepseek", ProviderName: "deepseek", Model: "deepseek-v4-flash", Provider: &fakeProvider{}},
	})

	_, err := router.Generate(context.Background(), ProviderRequest{Model: "missing"})

	if err == nil {
		t.Fatal("Generate() error = nil, want unknown model error")
	}
}
