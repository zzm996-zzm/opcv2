package research

import (
	"context"
	"errors"
	"net"
	"net/url"
	"sort"
	"strings"
	"time"
)

var (
	ErrUnsafeURL   = errors.New("unsafe research URL")
	ErrFetchFailed = errors.New("research page fetch failed")
)

type SearchResult struct {
	Title   string
	URL     string
	Snippet string
	Quality float64
}

type Page struct {
	URL       string
	Title     string
	Publisher string
	Content   string
	FetchedAt time.Time
}

type Evidence struct {
	URL              string
	Title            string
	Publisher        string
	Excerpt          string
	Quality          float64
	UntrustedContent bool
}

type Report struct {
	Evidence []Evidence
	Degraded bool
	Errors   []string
}

type SearchProvider interface {
	Search(context.Context, string, int) ([]SearchResult, error)
}

type FetchProvider interface {
	Fetch(context.Context, string) (Page, error)
}

type EvidenceExtractor interface {
	Extract(context.Context, Page, SearchResult) (Evidence, error)
}

type Service struct {
	searcher  SearchProvider
	fetcher   FetchProvider
	extractor EvidenceExtractor
	threshold float64
}

func NewService(searcher SearchProvider, fetcher FetchProvider, extractor EvidenceExtractor, qualityThreshold float64) *Service {
	if qualityThreshold <= 0 || qualityThreshold > 1 {
		qualityThreshold = 0.5
	}
	return &Service{searcher: searcher, fetcher: fetcher, extractor: extractor, threshold: qualityThreshold}
}

func (s *Service) Research(ctx context.Context, query string, limit int) (Report, error) {
	results, err := s.searcher.Search(ctx, strings.TrimSpace(query), limit)
	if err != nil {
		return Report{}, err
	}
	report := Report{Evidence: []Evidence{}, Errors: []string{}}
	seen := map[string]bool{}
	for _, result := range results {
		if result.Quality < s.threshold {
			continue
		}
		canonical, canonicalErr := CanonicalizeURL(result.URL)
		if canonicalErr != nil || seen[canonical] {
			report.Degraded = true
			if canonicalErr != nil {
				report.Errors = append(report.Errors, "unsafe_url")
			}
			continue
		}
		seen[canonical] = true
		page, fetchErr := s.fetcher.Fetch(ctx, canonical)
		if fetchErr != nil {
			report.Degraded = true
			report.Errors = append(report.Errors, "fetch_failed")
			continue
		}
		evidence, extractErr := s.extractor.Extract(ctx, page, result)
		if extractErr != nil {
			report.Degraded = true
			report.Errors = append(report.Errors, "extract_failed")
			continue
		}
		evidence.URL, evidence.Quality, evidence.UntrustedContent = canonical, result.Quality, true
		report.Evidence = append(report.Evidence, evidence)
	}
	return report, nil
}

func CanonicalizeURL(raw string) (string, error) {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || parsed.User != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Hostname() == "" {
		return "", ErrUnsafeURL
	}
	host := strings.ToLower(parsed.Hostname())
	if host == "localhost" || strings.HasSuffix(host, ".localhost") {
		return "", ErrUnsafeURL
	}
	if ip := net.ParseIP(host); ip != nil && (ip.IsLoopback() || ip.IsPrivate() || ip.IsUnspecified() || ip.IsLinkLocalUnicast()) {
		return "", ErrUnsafeURL
	}
	parsed.Scheme = strings.ToLower(parsed.Scheme)
	parsed.Host = strings.ToLower(parsed.Host)
	parsed.Fragment = ""
	parsed.Path = strings.TrimRight(parsed.Path, "/")
	if parsed.Path == "" {
		parsed.Path = "/"
	}
	query := parsed.Query()
	for key := range query {
		lower := strings.ToLower(key)
		if strings.HasPrefix(lower, "utm_") || lower == "fbclid" || lower == "gclid" {
			query.Del(key)
		}
	}
	parsed.RawQuery = query.Encode()
	return parsed.String(), nil
}

type DevelopmentProvider struct {
	Results []SearchResult
	Pages   map[string]Page
	Now     func() time.Time
}

func (p *DevelopmentProvider) Search(_ context.Context, query string, limit int) ([]SearchResult, error) {
	if limit <= 0 || limit > 20 {
		limit = 10
	}
	query = strings.ToLower(strings.TrimSpace(query))
	results := append([]SearchResult(nil), p.Results...)
	sort.SliceStable(results, func(i, j int) bool { return results[i].Quality > results[j].Quality })
	filtered := make([]SearchResult, 0, limit)
	for _, result := range results {
		if query != "" && !strings.Contains(strings.ToLower(result.Title+" "+result.Snippet), query) {
			continue
		}
		filtered = append(filtered, result)
		if len(filtered) == limit {
			break
		}
	}
	return filtered, nil
}

func (p *DevelopmentProvider) Fetch(_ context.Context, rawURL string) (Page, error) {
	canonical, err := CanonicalizeURL(rawURL)
	if err != nil {
		return Page{}, err
	}
	page, ok := p.Pages[canonical]
	if !ok {
		return Page{}, ErrFetchFailed
	}
	page.URL = canonical
	if page.FetchedAt.IsZero() && p.Now != nil {
		page.FetchedAt = p.Now()
	}
	return page, nil
}

func (p *DevelopmentProvider) Extract(_ context.Context, page Page, result SearchResult) (Evidence, error) {
	excerpt := strings.TrimSpace(page.Content)
	if len([]rune(excerpt)) > 500 {
		excerpt = string([]rune(excerpt)[:500])
	}
	return Evidence{URL: page.URL, Title: page.Title, Publisher: page.Publisher, Excerpt: excerpt, Quality: result.Quality, UntrustedContent: true}, nil
}
