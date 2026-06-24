package leads

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestTianyanchaProviderMapsSuccessfulPayload(t *testing.T) {
	var requestPath string
	var authHeader string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestPath = r.URL.Path
		authHeader = r.Header.Get("Authorization")
		if got := r.URL.Query().Get("word"); got != "成都 教培" {
			t.Fatalf("word = %q, want 成都 教培", got)
		}
		if got := r.URL.Query().Get("pageSize"); got != "20" {
			t.Fatalf("pageSize = %q, want 20", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"error_code": 0,
			"reason": "ok",
			"result": {
				"items": [
					{
						"id": 13279712,
						"name": "<em>成都启明星教育咨询有限公司</em>",
						"contactPhone": "028-12345678",
						"contactEmail": "hello@example.com",
						"website": "https://example.com",
						"creditCode": "91510100MA00000000",
						"regStatus": "存续",
						"base": "四川"
					},
					{
						"id": 1737464,
						"name": "成都缺字段科技有限公司"
					}
				]
			}
		}`))
	}))
	defer server.Close()

	provider := NewTianyanchaProvider(TianyanchaConfig{
		BaseURL: server.URL,
		APIKey:  "test-tyc-key",
		Timeout: time.Second,
	})

	leads, err := provider.SearchLeads(context.Background(), SearchInput{Query: "成都 教培"})
	if err != nil {
		t.Fatalf("SearchLeads() error = %v", err)
	}

	if requestPath != "/services/open/search/2.0" {
		t.Fatalf("path = %q, want /services/open/search/2.0", requestPath)
	}
	if authHeader != "test-tyc-key" {
		t.Fatalf("Authorization = %q, want API key", authHeader)
	}
	if len(leads) != 2 {
		t.Fatalf("leads length = %d, want 2", len(leads))
	}
	if leads[0].Name != "成都启明星教育咨询有限公司" || leads[0].Phone != "028-12345678" || leads[0].Email != "hello@example.com" || leads[0].Website != "https://example.com" {
		t.Fatalf("first lead = %+v", leads[0])
	}
	if len(leads[0].Evidence) != 1 || leads[0].Evidence[0].URL != "https://www.tianyancha.com/company/13279712" {
		t.Fatalf("first evidence = %+v", leads[0].Evidence)
	}
	if leads[1].Name != "成都缺字段科技有限公司" || leads[1].Phone != "" || leads[1].Email != "" || leads[1].Website != "" {
		t.Fatalf("second lead with optional fields = %+v", leads[1])
	}
}

func TestTianyanchaProviderClassifiesQuotaAndRateLimit(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       string
		want       error
	}{
		{name: "quota response code", statusCode: http.StatusOK, body: `{"error_code":300006,"reason":"余额不足"}`, want: ErrProviderQuotaExceeded},
		{name: "rate limited response code", statusCode: http.StatusOK, body: `{"error_code":300004,"reason":"访问频率过快"}`, want: ErrProviderRateLimited},
		{name: "http rate limited", statusCode: http.StatusTooManyRequests, body: `{"reason":"too many requests"}`, want: ErrProviderRateLimited},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer server.Close()

			provider := NewTianyanchaProvider(TianyanchaConfig{BaseURL: server.URL, APIKey: "key", Timeout: time.Second})
			_, err := provider.SearchLeads(context.Background(), SearchInput{Query: "成都 教培"})
			if !errors.Is(err, tt.want) {
				t.Fatalf("SearchLeads() error = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestTianyanchaProviderClassifiesTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond)
	}))
	defer server.Close()

	provider := NewTianyanchaProvider(TianyanchaConfig{BaseURL: server.URL, APIKey: "key", Timeout: time.Millisecond})

	_, err := provider.SearchLeads(context.Background(), SearchInput{Query: "成都 教培"})
	if !errors.Is(err, ErrProviderTimeout) {
		t.Fatalf("SearchLeads() error = %v, want ErrProviderTimeout", err)
	}
}
