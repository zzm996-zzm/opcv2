package ai

import (
	"context"
	"fmt"
	"strings"
)

type ModelRoute struct {
	Alias        string
	ProviderName string
	Model        string
	Provider     Provider
}

type ModelRouter struct {
	defaultAlias string
	routes       map[string]ModelRoute
}

func NewModelRouter(defaultAlias string, routes []ModelRoute) *ModelRouter {
	router := &ModelRouter{
		defaultAlias: strings.TrimSpace(defaultAlias),
		routes:       map[string]ModelRoute{},
	}
	for _, route := range routes {
		route.Alias = strings.TrimSpace(route.Alias)
		if route.Alias == "" || route.Provider == nil {
			continue
		}
		route.Model = strings.TrimSpace(route.Model)
		router.routes[route.Alias] = route
		if router.defaultAlias == "" {
			router.defaultAlias = route.Alias
		}
	}
	return router
}

func (r *ModelRouter) Generate(ctx context.Context, request ProviderRequest) (ProviderResponse, error) {
	alias := strings.TrimSpace(request.Model)
	if alias == "" {
		alias = r.defaultAlias
	}
	route, ok := r.routes[alias]
	if !ok || route.Provider == nil {
		return ProviderResponse{}, fmt.Errorf("%w: unknown model alias", ErrProviderUnavailable)
	}
	request.Model = route.Model
	return route.Provider.Generate(ctx, request)
}

func (r *ModelRouter) Stream(ctx context.Context, request ProviderRequest, onDelta func([]byte) error) (ProviderResponse, error) {
	alias := strings.TrimSpace(request.Model)
	if alias == "" {
		alias = r.defaultAlias
	}
	route, ok := r.routes[alias]
	if !ok || route.Provider == nil {
		return ProviderResponse{}, fmt.Errorf("%w: unknown model alias", ErrProviderUnavailable)
	}
	streamingProvider, ok := route.Provider.(StreamingProvider)
	if !ok {
		return ProviderResponse{}, fmt.Errorf("%w: model does not support streaming", ErrProviderUnavailable)
	}
	request.Model = route.Model
	return streamingProvider.Stream(ctx, request, onDelta)
}
