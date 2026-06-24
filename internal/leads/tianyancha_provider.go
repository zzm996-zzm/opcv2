package leads

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const defaultTianyanchaBaseURL = "https://open.api.tianyancha.com"

var htmlTagPattern = regexp.MustCompile(`<[^>]*>`)

type TianyanchaConfig struct {
	BaseURL string
	APIKey  string
	Timeout time.Duration
}

type TianyanchaProvider struct {
	baseURL string
	apiKey  string
	client  *http.Client
}

func NewTianyanchaProvider(config TianyanchaConfig) *TianyanchaProvider {
	timeout := config.Timeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	baseURL := strings.TrimRight(config.BaseURL, "/")
	if baseURL == "" {
		baseURL = defaultTianyanchaBaseURL
	}
	return &TianyanchaProvider{
		baseURL: baseURL,
		apiKey:  config.APIKey,
		client:  &http.Client{Timeout: timeout},
	}
}

func (p *TianyanchaProvider) SearchLeads(ctx context.Context, input SearchInput) ([]Lead, error) {
	endpoint, err := url.Parse(p.baseURL + "/services/open/search/2.0")
	if err != nil {
		return nil, fmt.Errorf("%w: invalid tianyancha base url", ErrProviderUnavailable)
	}
	query := endpoint.Query()
	query.Set("word", input.Query)
	query.Set("pageSize", "20")
	endpoint.RawQuery = query.Encode()

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("%w: build tianyancha request", ErrProviderUnavailable)
	}
	request.Header.Set("Authorization", p.apiKey)

	response, err := p.client.Do(request)
	if err != nil {
		return nil, classifyHTTPClientError(err)
	}
	defer response.Body.Close()

	if err := classifyHTTPStatus(response.StatusCode); err != nil {
		return nil, err
	}

	var payload tianyanchaSearchResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("%w: decode tianyancha response", ErrProviderUnavailable)
	}
	if err := classifyTianyanchaCode(payload.ErrorCode); err != nil {
		return nil, err
	}

	leads := make([]Lead, 0, len(payload.Result.Items))
	for _, item := range payload.Result.Items {
		name := cleanProviderText(item.Name)
		if name == "" {
			continue
		}
		lead := Lead{
			Name:    name,
			Phone:   item.ContactPhone,
			Email:   item.ContactEmail,
			Website: item.Website,
		}
		if item.ID != 0 {
			lead.Evidence = append(lead.Evidence, Evidence{
				Type:  "provider",
				Title: name,
				URL:   fmt.Sprintf("https://www.tianyancha.com/company/%d", item.ID),
			})
		}
		leads = append(leads, lead)
	}
	return leads, nil
}

type tianyanchaSearchResponse struct {
	ErrorCode int `json:"error_code"`
	Result    struct {
		Items []tianyanchaCompany `json:"items"`
	} `json:"result"`
}

type tianyanchaCompany struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	ContactPhone string `json:"contactPhone"`
	ContactEmail string `json:"contactEmail"`
	Website      string `json:"website"`
}

func classifyTianyanchaCode(code int) error {
	switch code {
	case 0:
		return nil
	case 300004:
		return ErrProviderRateLimited
	case 300006:
		return ErrProviderQuotaExceeded
	default:
		return fmt.Errorf("%w: tianyancha code %d", ErrProviderUnavailable, code)
	}
}

func cleanProviderText(value string) string {
	value = htmlTagPattern.ReplaceAllString(value, "")
	value = html.UnescapeString(value)
	return strings.TrimSpace(value)
}

func classifyHTTPStatus(statusCode int) error {
	switch statusCode {
	case http.StatusTooManyRequests:
		return ErrProviderRateLimited
	case http.StatusPaymentRequired:
		return ErrProviderQuotaExceeded
	}
	if statusCode < 200 || statusCode >= 300 {
		return fmt.Errorf("%w: provider status %d", ErrProviderUnavailable, statusCode)
	}
	return nil
}

func classifyHTTPClientError(err error) error {
	if errors.Is(err, context.DeadlineExceeded) {
		return ErrProviderTimeout
	}
	var urlErr *url.Error
	if errors.As(err, &urlErr) && urlErr.Timeout() {
		return ErrProviderTimeout
	}
	return fmt.Errorf("%w: %v", ErrProviderUnavailable, err)
}
