package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

type fakeAuthApplication struct {
	sendPhone  string
	loginIn    LoginInput
	registerIn RegisterInput
	result     LoginResult
	err        error
}

func (a *fakeAuthApplication) SendCode(_ context.Context, phone string) error {
	a.sendPhone = phone
	return a.err
}

func (a *fakeAuthApplication) Login(_ context.Context, input LoginInput) (LoginResult, error) {
	a.loginIn = input
	return a.result, a.err
}

func (a *fakeAuthApplication) Register(_ context.Context, input RegisterInput) (LoginResult, error) {
	a.registerIn = input
	return a.result, a.err
}

func (a *fakeAuthApplication) Refresh(_ context.Context, _ string) (LoginResult, error) {
	return a.result, a.err
}

func (a *fakeAuthApplication) Logout(_ context.Context, _ string) error {
	return a.err
}

func (a *fakeAuthApplication) CurrentUser(_ context.Context, userID int64) (User, error) {
	if a.err != nil {
		return User{}, a.err
	}
	user := a.result.User
	user.ID = userID
	return user, nil
}

type staticTokenManager struct {
	userID int64
	err    error
}

func (m staticTokenManager) IssueAccess(User) (string, time.Time, error) {
	return "", time.Time{}, nil
}

func (m staticTokenManager) ParseAccess(string) (int64, error) {
	return m.userID, m.err
}

func newAuthTestRouter(app AuthApplication, tokens TokenManager) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := NewHTTPHandler(app, tokens, false)
	handler.Register(router.Group("/api/v1"))
	return router
}

func TestSendCodeEndpointReturnsAcceptedWithoutExposingCode(t *testing.T) {
	app := &fakeAuthApplication{}
	router := newAuthTestRouter(app, staticTokenManager{})
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/auth/sms/send",
		strings.NewReader(`{"phone":"13800138000"}`),
	)
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusAccepted)
	}
	if app.sendPhone != "13800138000" {
		t.Fatalf("phone = %q", app.sendPhone)
	}
	if strings.Contains(recorder.Body.String(), "246810") {
		t.Fatal("response exposed development code")
	}
}

func TestAuthEndpointsRejectUnknownJSONFields(t *testing.T) {
	app := &fakeAuthApplication{}
	router := newAuthTestRouter(app, staticTokenManager{})
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/auth/login",
		strings.NewReader(`{"account":"deploy_user","password":"secret123","unexpected":true}`),
	)
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	assertAuthErrorResponse(t, recorder, http.StatusBadRequest, "invalid_request", "登录信息格式有误，请检查后重试")
	if app.loginIn != (LoginInput{}) {
		t.Fatalf("login input = %+v, want zero value", app.loginIn)
	}
}

func TestAuthEndpointsRejectMultipleJSONValues(t *testing.T) {
	app := &fakeAuthApplication{}
	router := newAuthTestRouter(app, staticTokenManager{})
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/auth/register",
		strings.NewReader(`{"account":"deploy_user","password":"secret123"}{}`),
	)
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	assertAuthErrorResponse(t, recorder, http.StatusBadRequest, "invalid_request", "注册信息格式有误，请检查后重试")
	if app.registerIn != (RegisterInput{}) {
		t.Fatalf("register input = %+v, want zero value", app.registerIn)
	}
}

func TestLoginEndpointSetsRefreshCookie(t *testing.T) {
	app := &fakeAuthApplication{result: LoginResult{
		User:               User{ID: 42, Nickname: "张晨", Phone: "13800138000", Status: "active"},
		AccessToken:        "access-token",
		AccessTokenExpires: time.Now().Add(15 * time.Minute),
		RefreshToken:       "refresh-token",
		IsNewUser:          true,
	}}
	router := newAuthTestRouter(app, staticTokenManager{})
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/auth/login",
		strings.NewReader(`{
			"nickname":"张晨",
			"phone":"13800138000",
			"code":"246810",
			"agreement_accepted":true
		}`),
	)
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	cookie := recorder.Result().Cookies()[0]
	if cookie.Name != refreshCookieName || cookie.Value != "refresh-token" || !cookie.HttpOnly {
		t.Fatalf("refresh cookie = %+v", cookie)
	}
	var response map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if response["access_token"] != "access-token" {
		t.Fatalf("response = %+v", response)
	}
}

func TestRegisterEndpointAcceptsAccountPasswordAndSetsRefreshCookie(t *testing.T) {
	app := &fakeAuthApplication{result: LoginResult{
		User:               User{ID: 42, Nickname: "部署测试", Account: "deploy_user", Status: "active"},
		AccessToken:        "access-token",
		AccessTokenExpires: time.Now().Add(15 * time.Minute),
		RefreshToken:       "refresh-token",
		IsNewUser:          true,
	}}
	router := newAuthTestRouter(app, staticTokenManager{})
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/auth/register",
		strings.NewReader(`{
			"nickname":"部署测试",
			"account":"deploy_user",
			"password":"secret123",
			"agreement_accepted":true
		}`),
	)
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.registerIn.Account != "deploy_user" || app.registerIn.Password != "secret123" || !app.registerIn.AgreementAccepted {
		t.Fatalf("register input = %+v", app.registerIn)
	}
	cookie := recorder.Result().Cookies()[0]
	if cookie.Name != refreshCookieName || cookie.Value != "refresh-token" || !cookie.HttpOnly {
		t.Fatalf("refresh cookie = %+v", cookie)
	}
}

func TestLoginEndpointAcceptsAccountPassword(t *testing.T) {
	app := &fakeAuthApplication{result: LoginResult{
		User:               User{ID: 42, Nickname: "部署测试", Account: "deploy_user", Status: "active"},
		AccessToken:        "access-token",
		AccessTokenExpires: time.Now().Add(15 * time.Minute),
		RefreshToken:       "refresh-token",
	}}
	router := newAuthTestRouter(app, staticTokenManager{})
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/auth/login",
		strings.NewReader(`{
			"account":"deploy_user",
			"password":"secret123"
		}`),
	)
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.loginIn.Account != "deploy_user" || app.loginIn.Password != "secret123" {
		t.Fatalf("login input = %+v", app.loginIn)
	}
}

func TestMeEndpointRequiresBearerToken(t *testing.T) {
	app := &fakeAuthApplication{result: LoginResult{
		User: User{Nickname: "张晨", Phone: "13800138000", Status: "active"},
	}}
	router := newAuthTestRouter(app, staticTokenManager{userID: 42})

	unauthorized := httptest.NewRecorder()
	router.ServeHTTP(unauthorized, httptest.NewRequest(http.MethodGet, "/api/v1/me", nil))
	if unauthorized.Code != http.StatusUnauthorized {
		t.Fatalf("unauthorized status = %d", unauthorized.Code)
	}

	request := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	request.Header.Set("Authorization", "Bearer access-token")
	authorized := httptest.NewRecorder()
	router.ServeHTTP(authorized, request)
	if authorized.Code != http.StatusOK {
		t.Fatalf("authorized status = %d body=%s", authorized.Code, authorized.Body.String())
	}
}

func TestRefreshEndpointRejectsDisabledUser(t *testing.T) {
	app := &fakeAuthApplication{err: ErrUserDisabled}
	router := newAuthTestRouter(app, staticTokenManager{})
	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", nil)
	request.AddCookie(&http.Cookie{Name: refreshCookieName, Value: "refresh-token"})
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("status = %d body=%s, want %d", recorder.Code, recorder.Body.String(), http.StatusForbidden)
	}
	assertAuthErrorResponse(t, recorder, http.StatusForbidden, "user_disabled", "该账号已被停用，如有疑问请联系客服")
}

func TestAuthEndpointReturnsServiceNotReadyForMissingDependencies(t *testing.T) {
	app := &fakeAuthApplication{err: ErrAuthNotConfigured}
	router := newAuthTestRouter(app, staticTokenManager{})
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/auth/login",
		strings.NewReader(`{"account":"deploy_user","password":"secret123"}`),
	)
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	assertAuthErrorResponse(t, recorder, http.StatusInternalServerError, "service_not_ready", "登录服务暂时不可用，请稍后再试")
}

func assertAuthErrorResponse(t *testing.T, recorder *httptest.ResponseRecorder, status int, code, message string) {
	t.Helper()
	if recorder.Code != status {
		t.Fatalf("status = %d, want %d; body=%s", recorder.Code, status, recorder.Body.String())
	}
	var response struct {
		Error   string `json:"error"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Error != code || response.Message != message {
		t.Fatalf("response = %+v, want error=%q message=%q", response, code, message)
	}
}
