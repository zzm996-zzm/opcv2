package account

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/zzm/opcv2/internal/auth"
)

type fakeApplication struct {
	profile     ProfilePayload
	update      ProfileUpdate
	onboarding  OnboardingState
	preferences Preferences
	prefsUpdate PreferencesUpdate
	quotas      []Quota
	content     []ContentItem
	deletion    DeletionStatus
	userID      int64
	limit       int
	err         error
}

func (a *fakeApplication) GetProfile(_ context.Context, userID int64) (ProfilePayload, error) {
	a.userID = userID
	return a.profile, a.err
}

func (a *fakeApplication) UpdateProfile(_ context.Context, userID int64, update ProfileUpdate) (ProfilePayload, error) {
	a.userID = userID
	a.update = update
	return a.profile, a.err
}

func (a *fakeApplication) GetOnboarding(_ context.Context, userID int64) (OnboardingState, error) {
	a.userID = userID
	return a.onboarding, a.err
}

func (a *fakeApplication) SaveOnboarding(_ context.Context, userID int64, state OnboardingState) (OnboardingState, error) {
	a.userID = userID
	a.onboarding = state
	return state, a.err
}

func (a *fakeApplication) CompleteOnboarding(_ context.Context, userID int64) (OnboardingState, error) {
	a.userID = userID
	a.onboarding.Completed = true
	return a.onboarding, a.err
}

func (a *fakeApplication) GetPreferences(_ context.Context, userID int64) (Preferences, error) {
	a.userID = userID
	return a.preferences, a.err
}

func (a *fakeApplication) UpdatePreferences(_ context.Context, userID int64, update PreferencesUpdate) (Preferences, error) {
	a.userID = userID
	a.prefsUpdate = update
	return a.preferences, a.err
}

func (a *fakeApplication) ListQuotas(_ context.Context, userID int64) ([]Quota, error) {
	a.userID = userID
	return a.quotas, a.err
}

func (a *fakeApplication) ListContent(_ context.Context, userID int64, limit int) ([]ContentItem, error) {
	a.userID = userID
	a.limit = limit
	return a.content, a.err
}

func (a *fakeApplication) DeleteAccount(_ context.Context, userID int64) (DeletionStatus, error) {
	a.userID = userID
	return a.deletion, a.err
}

func accountTestRouter(app Application) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	group := router.Group("/api/v1")
	group.Use(func(c *gin.Context) {
		c.Set(auth.UserIDContextKey, int64(42))
		c.Next()
	})
	NewHTTPHandler(app).Register(group)
	return router
}

func TestGetProfileEndpointUsesAuthenticatedUser(t *testing.T) {
	app := &fakeApplication{profile: ProfilePayload{Profile: Profile{ID: 42, Nickname: "张晨"}}}
	router := accountTestRouter(app)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/account/profile", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.userID != 42 || !strings.Contains(recorder.Body.String(), `"nickname":"张晨"`) {
		t.Fatalf("user/body = %d/%s", app.userID, recorder.Body.String())
	}
}

func TestUpdateProfileEndpointPassesPatchFields(t *testing.T) {
	app := &fakeApplication{profile: ProfilePayload{Profile: Profile{ID: 42, Nickname: "张晨"}}}
	router := accountTestRouter(app)
	request := httptest.NewRequest(http.MethodPatch, "/api/v1/account/profile", strings.NewReader(`{"nickname":"张晨","email":"founder@example.com"}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.userID != 42 || app.update.Nickname == nil || *app.update.Nickname != "张晨" {
		t.Fatalf("user/update = %d/%+v", app.userID, app.update)
	}
}

func TestOnboardingEndpointsRoundTripState(t *testing.T) {
	app := &fakeApplication{}
	router := accountTestRouter(app)
	request := httptest.NewRequest(http.MethodPut, "/api/v1/account/onboarding", strings.NewReader(`{
		"sections":[{"key":"identity","title":"基本身份","fields":{"role":"创始人"}}]
	}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.userID != 42 || len(app.onboarding.Sections) != 1 {
		t.Fatalf("user/onboarding = %d/%+v", app.userID, app.onboarding)
	}

	recorder = httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/api/v1/account/onboarding/complete", nil))
	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), `"completed":true`) {
		t.Fatalf("complete status/body = %d/%s", recorder.Code, recorder.Body.String())
	}
}

func TestPreferencesEndpointPassesPatchFields(t *testing.T) {
	app := &fakeApplication{preferences: DefaultPreferences()}
	router := accountTestRouter(app)
	request := httptest.NewRequest(http.MethodPatch, "/api/v1/account/preferences", strings.NewReader(`{"default_model":"deepseek"}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.prefsUpdate.DefaultModel == nil || *app.prefsUpdate.DefaultModel != "deepseek" {
		t.Fatalf("prefs update = %+v", app.prefsUpdate)
	}
}

func TestListContentEndpointCapsLimitAndReturnsEmptyArray(t *testing.T) {
	app := &fakeApplication{}
	router := accountTestRouter(app)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/account/content?limit=500", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.userID != 42 || app.limit != 100 || !strings.Contains(recorder.Body.String(), `"items":[]`) {
		t.Fatalf("user/limit/body = %d/%d/%s", app.userID, app.limit, recorder.Body.String())
	}
}

func TestDeleteAccountEndpointReturnsAccepted(t *testing.T) {
	app := &fakeApplication{deletion: DeletionStatus{Status: "pending_deletion"}}
	router := accountTestRouter(app)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodDelete, "/api/v1/account", nil))

	if recorder.Code != http.StatusAccepted {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.userID != 42 || !strings.Contains(recorder.Body.String(), `"pending_deletion"`) {
		t.Fatalf("user/body = %d/%s", app.userID, recorder.Body.String())
	}
}
