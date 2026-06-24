package httpserver

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/zzm/opcv2/internal/auth"
	"github.com/zzm/opcv2/internal/content"
	"github.com/zzm/opcv2/internal/membership"
)

func TestHealthEndpoints(t *testing.T) {
	router := NewRouter(HealthChecks{
		Ready: func() bool { return true },
	}, nil, nil, nil, nil, nil, nil, nil)

	tests := []struct {
		name       string
		path       string
		statusCode int
		body       string
	}{
		{
			name:       "live endpoint reports running process",
			path:       "/health/live",
			statusCode: http.StatusOK,
			body:       `{"status":"ok"}`,
		},
		{
			name:       "ready endpoint reports available dependencies",
			path:       "/health/ready",
			statusCode: http.StatusOK,
			body:       `{"status":"ok"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, tt.path, nil)

			router.ServeHTTP(recorder, request)

			if recorder.Code != tt.statusCode {
				t.Fatalf("status = %d, want %d", recorder.Code, tt.statusCode)
			}
			if recorder.Body.String() != tt.body {
				t.Fatalf("body = %q, want %q", recorder.Body.String(), tt.body)
			}
		})
	}
}

func TestReadyEndpointReportsUnavailableDependency(t *testing.T) {
	router := NewRouter(HealthChecks{
		Ready: func() bool { return false },
	}, nil, nil, nil, nil, nil, nil, nil)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/health/ready", nil)

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusServiceUnavailable)
	}
	if recorder.Body.String() != `{"status":"unavailable"}` {
		t.Fatalf("body = %q, want unavailable status", recorder.Body.String())
	}
}

type fakeAuthApp struct{}

func (fakeAuthApp) SendCode(context.Context, string) error { return nil }
func (fakeAuthApp) Login(context.Context, auth.LoginInput) (auth.LoginResult, error) {
	return auth.LoginResult{}, nil
}
func (fakeAuthApp) Register(context.Context, auth.RegisterInput) (auth.LoginResult, error) {
	return auth.LoginResult{}, nil
}
func (fakeAuthApp) Refresh(context.Context, string) (auth.LoginResult, error) {
	return auth.LoginResult{}, nil
}
func (fakeAuthApp) Logout(context.Context, string) error { return nil }
func (fakeAuthApp) CurrentUser(context.Context, int64) (auth.User, error) {
	return auth.User{}, nil
}

type fakeTokenManager struct{}

func (fakeTokenManager) IssueAccess(auth.User) (string, time.Time, error) {
	return "", time.Time{}, nil
}
func (fakeTokenManager) ParseAccess(string) (int64, error) { return 42, nil }

type fakeMembershipApp struct{}

func (fakeMembershipApp) CurrentSnapshot(context.Context, int64) (membership.Snapshot, error) {
	return membership.Snapshot{
		Plan:          membership.PlanCatalog[membership.PlanFree],
		CreditBalance: 7,
	}, nil
}
func (fakeMembershipApp) Redeem(context.Context, membership.RedeemInput) (membership.RedeemResult, error) {
	return membership.RedeemResult{}, nil
}

func TestMembershipRoutesAreMountedBehindAuth(t *testing.T) {
	authHTTP := auth.NewHTTPHandler(fakeAuthApp{}, fakeTokenManager{}, false)
	membershipHTTP := membership.NewHTTPHandler(fakeMembershipApp{})
	router := NewRouter(HealthChecks{}, authHTTP, membershipHTTP, nil, nil, nil, nil, nil)

	unauthorized := httptest.NewRecorder()
	router.ServeHTTP(unauthorized, httptest.NewRequest(http.MethodGet, "/api/v1/membership/me", nil))
	if unauthorized.Code != http.StatusUnauthorized {
		t.Fatalf("unauthorized status = %d, want %d", unauthorized.Code, http.StatusUnauthorized)
	}

	request := httptest.NewRequest(http.MethodGet, "/api/v1/membership/me", nil)
	request.Header.Set("Authorization", "Bearer access-token")
	authorized := httptest.NewRecorder()
	router.ServeHTTP(authorized, request)
	if authorized.Code != http.StatusOK {
		t.Fatalf("authorized status = %d body=%s", authorized.Code, authorized.Body.String())
	}
	if !strings.Contains(authorized.Body.String(), `"credit_balance":7`) {
		t.Fatalf("body = %s", authorized.Body.String())
	}
}

type fakeContentApp struct{}

func (fakeContentApp) ListArticles(context.Context) ([]content.Article, error) {
	return []content.Article{}, nil
}
func (fakeContentApp) GetArticle(context.Context, string) (content.Article, error) {
	return content.Article{}, content.ErrArticleNotFound
}
func (fakeContentApp) CreateArticle(context.Context, int64, content.ArticleInput) (content.Article, error) {
	return content.Article{}, content.ErrAdminRequired
}
func (fakeContentApp) ListTools(context.Context) ([]content.Tool, error) {
	return []content.Tool{}, nil
}
func (fakeContentApp) UpsertTool(context.Context, int64, content.ToolInput) (content.Tool, error) {
	return content.Tool{}, content.ErrAdminRequired
}
func (fakeContentApp) GetCommunityConfig(context.Context) (content.CommunityConfig, error) {
	return content.CommunityConfig{}, nil
}
func (fakeContentApp) UpdateCommunityConfig(context.Context, int64, content.CommunityConfigInput) (content.CommunityConfig, error) {
	return content.CommunityConfig{}, content.ErrAdminRequired
}
func (fakeContentApp) GetBrand(context.Context) (content.BrandContent, error) {
	return content.BrandContent{Metrics: []content.BrandMetric{}, Cases: []content.BrandCase{}}, nil
}
func (fakeContentApp) UpsertBrandMetric(context.Context, int64, content.BrandMetricInput) (content.BrandMetric, error) {
	return content.BrandMetric{}, content.ErrAdminRequired
}
func (fakeContentApp) UpsertBrandCase(context.Context, int64, content.BrandCaseInput) (content.BrandCase, error) {
	return content.BrandCase{}, content.ErrAdminRequired
}

func TestContentRoutesExposePublicReadsAndProtectAdminWrites(t *testing.T) {
	authHTTP := auth.NewHTTPHandler(fakeAuthApp{}, fakeTokenManager{}, false)
	contentHTTP := content.NewHTTPHandler(fakeContentApp{})
	router := NewRouter(HealthChecks{}, authHTTP, nil, nil, nil, nil, nil, contentHTTP)

	public := httptest.NewRecorder()
	router.ServeHTTP(public, httptest.NewRequest(http.MethodGet, "/api/v1/content/articles", nil))
	if public.Code != http.StatusOK {
		t.Fatalf("public status = %d body=%s", public.Code, public.Body.String())
	}

	unauthorized := httptest.NewRecorder()
	router.ServeHTTP(unauthorized, httptest.NewRequest(http.MethodPost, "/api/v1/admin/content/articles", strings.NewReader(`{}`)))
	if unauthorized.Code != http.StatusUnauthorized {
		t.Fatalf("unauthorized status = %d, want %d", unauthorized.Code, http.StatusUnauthorized)
	}

	request := httptest.NewRequest(http.MethodPost, "/api/v1/admin/content/articles", strings.NewReader(`{}`))
	request.Header.Set("Authorization", "Bearer access-token")
	authorized := httptest.NewRecorder()
	router.ServeHTTP(authorized, request)
	if authorized.Code != http.StatusForbidden {
		t.Fatalf("authorized status = %d body=%s, want forbidden from app role check", authorized.Code, authorized.Body.String())
	}
}

func TestRouterAppliesExplicitCORSOrigins(t *testing.T) {
	router := NewRouter(HealthChecks{
		AllowedOrigins: []string{"https://app.example.com"},
	}, nil, nil, nil, nil, nil, nil, nil)

	allowed := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/health/live", nil)
	request.Header.Set("Origin", "https://app.example.com")
	router.ServeHTTP(allowed, request)
	if allowed.Header().Get("Access-Control-Allow-Origin") != "https://app.example.com" {
		t.Fatalf("allowed origin header = %q", allowed.Header().Get("Access-Control-Allow-Origin"))
	}

	disallowed := httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodGet, "/health/live", nil)
	request.Header.Set("Origin", "https://evil.example.com")
	router.ServeHTTP(disallowed, request)
	if disallowed.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Fatalf("disallowed origin header = %q", disallowed.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestRouterLogsRequestsWithoutSensitiveHeaders(t *testing.T) {
	var buffer bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buffer, nil))
	router := NewRouter(HealthChecks{Logger: logger}, nil, nil, nil, nil, nil, nil, nil)
	request := httptest.NewRequest(http.MethodGet, "/health/live?token=query-secret", nil)
	request.Header.Set("Authorization", "Bearer header-secret")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	logs := buffer.String()
	if !strings.Contains(logs, "request_complete") || !strings.Contains(logs, "path=/health/live") {
		t.Fatalf("logs = %s", logs)
	}
	if strings.Contains(logs, "header-secret") || strings.Contains(logs, "query-secret") {
		t.Fatalf("logs leaked sensitive data: %s", logs)
	}
}

func TestRouterRateLimitsExpensiveEndpoints(t *testing.T) {
	authHTTP := auth.NewHTTPHandler(fakeAuthApp{}, fakeTokenManager{}, false)
	router := NewRouter(HealthChecks{
		ExpensiveEndpointLimit: 1,
	}, authHTTP, nil, nil, nil, nil, nil, nil)

	for i := 0; i < 2; i++ {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost, "/api/v1/analysis/direction", strings.NewReader(`{}`))
		request.Header.Set("Authorization", "Bearer access-token")
		router.ServeHTTP(recorder, request)
		if i == 0 && recorder.Code == http.StatusTooManyRequests {
			t.Fatalf("first request was rate limited: body=%s", recorder.Body.String())
		}
		if i == 1 && recorder.Code != http.StatusTooManyRequests {
			t.Fatalf("second status = %d body=%s, want 429", recorder.Code, recorder.Body.String())
		}
	}
}
