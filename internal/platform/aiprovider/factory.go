package aiprovider

import (
	"errors"
	"time"

	"github.com/zzm/opcv2/internal/ai"
	"github.com/zzm/opcv2/internal/platform/config"
)

func New(cfg config.Config) (ai.Provider, error) {
	if len(cfg.AIModelRoutes) > 0 {
		routes := make([]ai.ModelRoute, 0, len(cfg.AIModelRoutes))
		for _, route := range cfg.AIModelRoutes {
			provider, err := providerForRoute(route, cfg)
			if err != nil {
				return nil, err
			}
			routes = append(routes, ai.ModelRoute{Alias: route.Alias, ProviderName: route.Provider, Model: route.Model, Provider: provider})
		}
		return ai.NewModelRouter(cfg.AIModel, routes), nil
	}
	switch cfg.AIProvider {
	case "development":
		return ai.NewDevelopmentProvider(), nil
	case "openai":
		return ai.NewOpenAIProvider(ai.OpenAIConfig{BaseURL: cfg.AIBaseURL, APIKey: cfg.AIAPIKey, Model: cfg.AIModel, Timeout: time.Duration(cfg.AITimeoutSeconds) * time.Second}), nil
	default:
		return nil, errors.New("unsupported AI provider")
	}
}

func providerForRoute(route config.AIModelRoute, cfg config.Config) (ai.Provider, error) {
	timeout := time.Duration(cfg.AITimeoutSeconds) * time.Second
	switch route.Provider {
	case "openai-responses":
		return ai.NewOpenAIProvider(ai.OpenAIConfig{BaseURL: route.BaseURL, APIKey: route.APIKey, Model: route.Model, Timeout: timeout}), nil
	case "openai-compatible":
		return ai.NewOpenAICompatibleProvider(ai.OpenAICompatibleConfig{BaseURL: route.BaseURL, APIKey: route.APIKey, Model: route.Model, Timeout: timeout}), nil
	case "development":
		return ai.NewDevelopmentProvider(), nil
	default:
		return nil, errors.New("unsupported AI model route provider")
	}
}
