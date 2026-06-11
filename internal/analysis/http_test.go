package analysis

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type fakeApplication struct {
	input  DirectionInput
	result DirectionResult
	err    error
}

func (a *fakeApplication) StartDirection(_ context.Context, input DirectionInput) (DirectionResult, error) {
	a.input = input
	return a.result, a.err
}

func TestDirectionEndpointUsesAuthenticatedUser(t *testing.T) {
	app := &fakeApplication{result: DirectionResult{
		SessionID: 99,
		Status:    StatusNeedsInput,
		Questions: []Question{{Key: "budget", Text: "启动资金大概多少？"}},
	}}
	router := testRouter(app)
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/analysis/direction",
		strings.NewReader(`{"intent":"想创业"}`),
	)
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.input.UserID != 42 || app.input.Intent != "想创业" {
		t.Fatalf("input = %+v", app.input)
	}
	if !strings.Contains(recorder.Body.String(), `"status":"needs_input"`) {
		t.Fatalf("body = %s", recorder.Body.String())
	}
}
