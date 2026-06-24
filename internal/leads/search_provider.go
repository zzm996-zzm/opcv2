package leads

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const defaultSerperBaseURL = "https://google.serper.dev"

type SearchProvider interface {
	SearchEvidence(ctx context.Context, input SearchInput) ([]Evidence, error)
}

type SerperConfig struct {
	BaseURL string
	APIKey  string
	Timeout time.Duration
}

type SerperSearchProvider struct {
	baseURL string
	apiKey  string
	client  *http.Client
}

func NewSerperSearchProvider(config SerperConfig) *SerperSearchProvider {
	timeout := config.Timeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	baseURL := strings.TrimRight(config.BaseURL, "/")
	if baseURL == "" {
		baseURL = defaultSerperBaseURL
	}
	return &SerperSearchProvider{
		baseURL: baseURL,
		apiKey:  config.APIKey,
		client:  &http.Client{Timeout: timeout},
	}
}

func (p *SerperSearchProvider) SearchEvidence(ctx context.Context, input SearchInput) ([]Evidence, error) {
	body, err := json.Marshal(map[string]string{"q": input.Query})
	if err != nil {
		return nil, fmt.Errorf("%w: encode serper request", ErrProviderUnavailable)
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/search", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("%w: build serper request", ErrProviderUnavailable)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-API-KEY", p.apiKey)

	response, err := p.client.Do(request)
	if err != nil {
		return nil, classifyHTTPClientError(err)
	}
	defer response.Body.Close()

	if err := classifyHTTPStatus(response.StatusCode); err != nil {
		return nil, err
	}

	var payload serperSearchResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("%w: decode serper response", ErrProviderUnavailable)
	}

	evidence := make([]Evidence, 0, len(payload.Organic))
	for _, result := range payload.Organic {
		title := strings.TrimSpace(result.Title)
		link := strings.TrimSpace(result.Link)
		if title == "" || link == "" {
			continue
		}
		evidence = append(evidence, Evidence{Type: "web", Title: title, URL: link})
	}
	return evidence, nil
}

type serperSearchResponse struct {
	Organic []struct {
		Title string `json:"title"`
		Link  string `json:"link"`
	} `json:"organic"`
}
