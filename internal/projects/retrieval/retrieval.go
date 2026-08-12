package retrieval

import (
	"context"
	"errors"
	"net/url"
	"sort"
	"strings"
	"unicode"
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
	return RankDocuments(request.Query, p.Documents, limit), nil
}

// RankDocuments applies deterministic lexical ranking to provider candidates.
// Provider supplied scores are retained as a secondary quality signal.
func RankDocuments(query string, documents []Document, limit int) []Document {
	if limit <= 0 || limit > 20 {
		limit = 20
	}
	terms := searchTerms(query)
	ranked := make([]Document, 0, len(documents))
	for _, document := range documents {
		haystack := strings.ToLower(document.Title + " " + document.Text)
		score := document.Score
		for _, term := range terms {
			if strings.Contains(strings.ToLower(document.Title), term) {
				score += 0.35
			}
			if strings.Contains(haystack, term) {
				score += 0.15
			}
		}
		if len(terms) > 0 && score <= document.Score {
			continue
		}
		document.Score = minFloat(score, 1)
		ranked = append(ranked, document)
	}
	sort.SliceStable(ranked, func(i, j int) bool {
		if ranked[i].Score == ranked[j].Score {
			return ranked[i].ID < ranked[j].ID
		}
		return ranked[i].Score > ranked[j].Score
	})
	if len(ranked) > limit {
		ranked = ranked[:limit]
	}
	return ranked
}

func searchTerms(value string) []string {
	fields := strings.FieldsFunc(strings.ToLower(strings.TrimSpace(value)), func(char rune) bool {
		return unicode.IsSpace(char) || strings.ContainsRune(",，。；;:：/|()（）[]【】", char)
	})
	seen := map[string]bool{}
	terms := make([]string, 0, len(fields))
	for _, field := range fields {
		field = strings.TrimSpace(field)
		if field != "" && !seen[field] {
			seen[field] = true
			terms = append(terms, field)
		}
		runes := []rune(field)
		if len(runes) >= 4 {
			for index := 0; index+1 < len(runes); index++ {
				bigram := string(runes[index : index+2])
				if !seen[bigram] {
					seen[bigram] = true
					terms = append(terms, bigram)
				}
			}
		}
	}
	return terms
}

func minFloat(left, right float64) float64 {
	if left < right {
		return left
	}
	return right
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
