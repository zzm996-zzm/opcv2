package retrieval

import (
	"context"
	"errors"
	"net/url"
	"strings"
)

var ErrNoRetrievalProvider = errors.New("retrieval provider is not configured")

type Document struct {
	ID       string
	Title    string
	Text     string
	Score    float64
	Metadata map[string]string
}

type SearchRequest struct {
	Query                string
	Limit                int
	KnowledgeBaseVersion string
}

type Provider interface {
	Search(context.Context, SearchRequest) ([]Document, error)
}

type EmbeddingProvider interface {
	Dimensions() int
	Embed(context.Context, []string) ([][]float32, error)
}

type DevelopmentProvider struct {
	Documents []Document
}

func (p DevelopmentProvider) Search(_ context.Context, request SearchRequest) ([]Document, error) {
	limit := request.Limit
	if limit <= 0 || limit > 20 {
		limit = 20
	}
	query := strings.ToLower(strings.TrimSpace(request.Query))
	result := make([]Document, 0, limit)
	for _, document := range p.Documents {
		if query != "" && !strings.Contains(strings.ToLower(document.Title+" "+document.Text), query) {
			continue
		}
		result = append(result, document)
		if len(result) == limit {
			break
		}
	}
	return result, nil
}

type DevelopmentEmbedding struct{ Dimension int }

func (p DevelopmentEmbedding) Dimensions() int {
	if p.Dimension <= 0 {
		return 0
	}
	return p.Dimension
}
func (p DevelopmentEmbedding) Embed(_ context.Context, texts []string) ([][]float32, error) {
	if p.Dimensions() <= 0 {
		return nil, ErrNoRetrievalProvider
	}
	result := make([][]float32, len(texts))
	for index, text := range texts {
		vector := make([]float32, p.Dimension)
		for position, char := range []byte(text) {
			vector[position%p.Dimension] += float32(char) / 255
		}
		result[index] = vector
	}
	return result, nil
}

func CanonicalURL(raw string) (string, error) {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		return "", errors.New("invalid http(s) URL")
	}
	parsed.Fragment = ""
	parsed.Scheme = strings.ToLower(parsed.Scheme)
	parsed.Host = strings.ToLower(parsed.Host)
	parsed.Path = strings.TrimRight(parsed.EscapedPath(), "/")
	if parsed.Path == "" {
		parsed.Path = "/"
	}
	return parsed.String(), nil
}
