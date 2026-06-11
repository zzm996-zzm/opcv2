package httpserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/zzm/opcv2/internal/auth"
	"github.com/zzm/opcv2/internal/membership"
)

func TestHealthEndpoints(t *testing.T) {
	router := NewRouter(HealthChecks{
		Ready: func() bool { return true },
	}, nil, nil)

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
	}, nil, nil)
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
	router := NewRouter(HealthChecks{}, authHTTP, membershipHTTP)

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
