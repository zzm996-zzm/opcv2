package leads

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSerperSearchProviderPreservesEvidenceURLs(t *testing.T) {
	var method string
	var apiKey string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method = r.Method
		apiKey = r.Header.Get("X-API-KEY")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"organic": [
				{
					"title": "成都启明星教育咨询有限公司官网",
					"link": "https://example.com",
					"snippet": "招生咨询与企业培训"
				},
				{
					"title": "企查页面",
					"link": "https://directory.example.com/company",
					"snippet": "工商信息"
				}
			]
		}`))
	}))
	defer server.Close()

	provider := NewSerperSearchProvider(SerperConfig{
		BaseURL: server.URL,
		APIKey:  "test-serper-key",
		Timeout: time.Second,
	})

	evidence, err := provider.SearchEvidence(context.Background(), SearchInput{Query: "成都启明星教育咨询有限公司"})
	if err != nil {
		t.Fatalf("SearchEvidence() error = %v", err)
	}

	if method != http.MethodPost {
		t.Fatalf("method = %q, want POST", method)
	}
	if apiKey != "test-serper-key" {
		t.Fatalf("X-API-KEY = %q, want API key", apiKey)
	}
	if len(evidence) != 2 {
		t.Fatalf("evidence length = %d, want 2", len(evidence))
	}
	if evidence[0].Type != "web" || evidence[0].Title != "成都启明星教育咨询有限公司官网" || evidence[0].URL != "https://example.com" {
		t.Fatalf("first evidence = %+v", evidence[0])
	}
	if evidence[1].URL != "https://directory.example.com/company" {
		t.Fatalf("second evidence = %+v", evidence[1])
	}
}

func TestSerperSearchProviderClassifiesErrors(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       string
		want       error
	}{
		{name: "quota exceeded", statusCode: http.StatusPaymentRequired, body: `{"message":"quota exceeded"}`, want: ErrProviderQuotaExceeded},
		{name: "rate limited", statusCode: http.StatusTooManyRequests, body: `{"message":"rate limited"}`, want: ErrProviderRateLimited},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer server.Close()

			provider := NewSerperSearchProvider(SerperConfig{BaseURL: server.URL, APIKey: "key", Timeout: time.Second})
			_, err := provider.SearchEvidence(context.Background(), SearchInput{Query: "成都 教培"})
			if !errors.Is(err, tt.want) {
				t.Fatalf("SearchEvidence() error = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestSerperSearchProviderClassifiesTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond)
	}))
	defer server.Close()

	provider := NewSerperSearchProvider(SerperConfig{BaseURL: server.URL, APIKey: "key", Timeout: time.Millisecond})

	_, err := provider.SearchEvidence(context.Background(), SearchInput{Query: "成都 教培"})
	if !errors.Is(err, ErrProviderTimeout) {
		t.Fatalf("SearchEvidence() error = %v, want ErrProviderTimeout", err)
	}
}
