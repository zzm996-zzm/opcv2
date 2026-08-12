package research

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

var (
	ErrUnsafeURL    = errors.New("unsafe research URL")
	ErrFetchFailed  = errors.New("research page fetch failed")
	ErrSearchFailed = errors.New("research search failed")
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

type SerperProvider struct {
	baseURL     string
	apiKey      string
	client      *http.Client
	fetchClient *http.Client
}

func NewSerperProvider(baseURL, apiKey string, timeout time.Duration) (*SerperProvider, error) {
	if strings.TrimSpace(apiKey) == "" {
		return nil, fmt.Errorf("%w: missing Serper API key", ErrSearchFailed)
	}
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		baseURL = "https://google.serper.dev"
	}
	return &SerperProvider{
		baseURL: baseURL, apiKey: apiKey, client: &http.Client{Timeout: timeout},
		fetchClient: newSafeFetchClient(timeout),
	}, nil
}

func (p *SerperProvider) Search(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	if limit <= 0 || limit > 20 {
		limit = 10
	}
	body, _ := json.Marshal(map[string]any{"q": strings.TrimSpace(query), "num": limit})
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/search", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSearchFailed, err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-API-KEY", p.apiKey)
	response, err := p.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSearchFailed, err)
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("%w: status %d", ErrSearchFailed, response.StatusCode)
	}
	var payload struct {
		Organic []struct {
			Title   string `json:"title"`
			Link    string `json:"link"`
			Snippet string `json:"snippet"`
		} `json:"organic"`
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, 2<<20)).Decode(&payload); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSearchFailed, err)
	}
	results := make([]SearchResult, 0, len(payload.Organic))
	for _, item := range payload.Organic {
		if strings.TrimSpace(item.Title) == "" || strings.TrimSpace(item.Link) == "" {
			continue
		}
		results = append(results, SearchResult{Title: strings.TrimSpace(item.Title), URL: strings.TrimSpace(item.Link), Snippet: strings.TrimSpace(item.Snippet), Quality: sourceQuality(item.Link)})
	}
	return results, nil
}

func (p *SerperProvider) Fetch(ctx context.Context, rawURL string) (Page, error) {
	canonical, err := CanonicalizeURL(rawURL)
	if err != nil {
		return Page{}, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, canonical, nil)
	if err != nil {
		return Page{}, ErrFetchFailed
	}
	request.Header.Set("User-Agent", "OPCV2Research/1.0")
	response, err := p.fetchClient.Do(request)
	if err != nil {
		return Page{}, fmt.Errorf("%w: %v", ErrFetchFailed, err)
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return Page{}, fmt.Errorf("%w: status %d", ErrFetchFailed, response.StatusCode)
	}
	contentType := strings.ToLower(response.Header.Get("Content-Type"))
	if contentType != "" && !strings.Contains(contentType, "text/html") && !strings.Contains(contentType, "text/plain") && !strings.Contains(contentType, "application/xhtml") {
		return Page{}, fmt.Errorf("%w: unsupported content type", ErrFetchFailed)
	}
	content, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if err != nil {
		return Page{}, fmt.Errorf("%w: %v", ErrFetchFailed, err)
	}
	return Page{URL: canonical, Title: canonical, Publisher: strings.ToLower(request.URL.Hostname()), Content: stripHTML(string(content)), FetchedAt: time.Now()}, nil
}

func newSafeFetchClient(timeout time.Duration) *http.Client {
	dialer := &net.Dialer{Timeout: timeout}
	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
			host, port, err := net.SplitHostPort(address)
			if err != nil {
				return nil, ErrUnsafeURL
			}
			addresses, err := net.DefaultResolver.LookupIPAddr(ctx, host)
			if err != nil || len(addresses) == 0 {
				return nil, ErrUnsafeURL
			}
			for _, address := range addresses {
				if unsafeIP(address.IP) {
					continue
				}
				return dialer.DialContext(ctx, network, net.JoinHostPort(address.IP.String(), port))
			}
			return nil, ErrUnsafeURL
		},
		ResponseHeaderTimeout: timeout,
		TLSHandshakeTimeout:   timeout,
		IdleConnTimeout:       30 * time.Second,
	}
	return &http.Client{
		Timeout:   timeout,
		Transport: transport,
		CheckRedirect: func(request *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return ErrUnsafeURL
			}
			_, err := CanonicalizeURL(request.URL.String())
			return err
		},
	}
}

func unsafeIP(ip net.IP) bool {
	return ip == nil || ip.IsLoopback() || ip.IsPrivate() || ip.IsUnspecified() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsMulticast()
}

func (p *SerperProvider) Extract(_ context.Context, page Page, result SearchResult) (Evidence, error) {
	excerpt := strings.TrimSpace(page.Content)
	if excerpt == "" {
		excerpt = strings.TrimSpace(result.Snippet)
	}
	if len([]rune(excerpt)) > 800 {
		excerpt = string([]rune(excerpt)[:800])
	}
	return Evidence{URL: page.URL, Title: result.Title, Publisher: page.Publisher, Excerpt: excerpt, Quality: result.Quality, UntrustedContent: true}, nil
}

func sourceQuality(rawURL string) float64 {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return 0
	}
	host := strings.ToLower(parsed.Hostname())
	switch {
	case strings.HasSuffix(host, ".gov.cn"), strings.HasSuffix(host, ".edu.cn"):
		return 0.95
	case strings.Contains(host, "stats.gov"), strings.Contains(host, "miit.gov"), strings.Contains(host, "cnnic"):
		return 0.95
	case strings.Contains(host, "36kr"), strings.Contains(host, "caixin"), strings.Contains(host, "iresearch"):
		return 0.75
	default:
		return 0.6
	}
}

func stripHTML(value string) string {
	var output strings.Builder
	inTag := false
	for _, char := range value {
		switch char {
		case '<':
			inTag = true
		case '>':
			inTag = false
			output.WriteRune(' ')
		default:
			if !inTag {
				output.WriteRune(char)
			}
		}
	}
	return strings.Join(strings.Fields(output.String()), " ")
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
