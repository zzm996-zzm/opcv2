package httpserver

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthEndpoints(t *testing.T) {
	router := NewRouter(HealthChecks{
		Ready: func() bool { return true },
	})

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
	})
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
