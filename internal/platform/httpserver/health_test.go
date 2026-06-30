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
	"github.com/zzm/opcv2/internal/competitor"
	"github.com/zzm/opcv2/internal/content"
	"github.com/zzm/opcv2/internal/dashboard"
	"github.com/zzm/opcv2/internal/growth"
	"github.com/zzm/opcv2/internal/learning"
	"github.com/zzm/opcv2/internal/membership"
	"github.com/zzm/opcv2/internal/sandbox"
	"github.com/zzm/opcv2/internal/tasks"
)

func TestHealthEndpoints(t *testing.T) {
	router := NewRouter(HealthChecks{
		Ready: func() bool { return true },
	}, Handlers{})

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
	}, Handlers{})
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
	router := NewRouter(HealthChecks{}, Handlers{Auth: authHTTP, Membership: membershipHTTP})

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
	router := NewRouter(HealthChecks{}, Handlers{Auth: authHTTP, Content: contentHTTP})

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

type fakeSandboxApp struct{}

func (fakeSandboxApp) CreateSession(context.Context, sandbox.CreateInput) (sandbox.Session, error) {
	return sandbox.Session{ID: 99, UserID: 42, Status: sandbox.StatusDraft}, nil
}
func (fakeSandboxApp) RunSession(context.Context, int64, int64) (sandbox.Session, error) {
	return sandbox.Session{ID: 99, UserID: 42, Status: sandbox.StatusCompleted, Report: sandbox.Report{Score: 83}}, nil
}
func (fakeSandboxApp) ListSessions(context.Context, int64, int) ([]sandbox.Session, error) {
	return []sandbox.Session{{ID: 99, UserID: 42, Status: sandbox.StatusDraft}}, nil
}
func (fakeSandboxApp) GetSession(context.Context, int64, int64) (sandbox.Session, error) {
	return sandbox.Session{ID: 99, UserID: 42, Status: sandbox.StatusDraft}, nil
}

func TestSandboxRoutesAreMountedBehindAuth(t *testing.T) {
	authHTTP := auth.NewHTTPHandler(fakeAuthApp{}, fakeTokenManager{}, false)
	sandboxHTTP := sandbox.NewHTTPHandler(fakeSandboxApp{})
	router := NewRouter(HealthChecks{}, Handlers{Auth: authHTTP, Sandbox: sandboxHTTP})

	unauthorized := httptest.NewRecorder()
	router.ServeHTTP(unauthorized, httptest.NewRequest(http.MethodGet, "/api/v1/sandbox/sessions", nil))
	if unauthorized.Code != http.StatusUnauthorized {
		t.Fatalf("unauthorized status = %d, want %d", unauthorized.Code, http.StatusUnauthorized)
	}

	request := httptest.NewRequest(http.MethodGet, "/api/v1/sandbox/sessions", nil)
	request.Header.Set("Authorization", "Bearer access-token")
	authorized := httptest.NewRecorder()
	router.ServeHTTP(authorized, request)
	if authorized.Code != http.StatusOK {
		t.Fatalf("authorized status = %d body=%s", authorized.Code, authorized.Body.String())
	}
	if !strings.Contains(authorized.Body.String(), `"sessions"`) {
		t.Fatalf("body = %s", authorized.Body.String())
	}
}

type fakeTasksApp struct{}

func (fakeTasksApp) CreateTask(context.Context, tasks.CreateInput) (tasks.Task, error) {
	return tasks.Task{ID: 99, UserID: 42, Title: "整理客户名单", Status: tasks.StatusTodo}, nil
}
func (fakeTasksApp) ListTasks(context.Context, int64, int) ([]tasks.Task, error) {
	return []tasks.Task{{ID: 99, UserID: 42, Title: "整理客户名单", Status: tasks.StatusTodo}}, nil
}
func (fakeTasksApp) GetTask(context.Context, int64, int64) (tasks.Task, error) {
	return tasks.Task{ID: 99, UserID: 42, Title: "整理客户名单", Status: tasks.StatusTodo}, nil
}
func (fakeTasksApp) UpdateTask(context.Context, int64, int64, tasks.TaskUpdate) (tasks.Task, error) {
	return tasks.Task{ID: 99, UserID: 42, Title: "整理客户名单", Status: tasks.StatusCompleted}, nil
}

func TestTaskRoutesAreMountedBehindAuth(t *testing.T) {
	authHTTP := auth.NewHTTPHandler(fakeAuthApp{}, fakeTokenManager{}, false)
	tasksHTTP := tasks.NewHTTPHandler(fakeTasksApp{})
	router := NewRouter(HealthChecks{}, Handlers{Auth: authHTTP, Tasks: tasksHTTP})

	unauthorized := httptest.NewRecorder()
	router.ServeHTTP(unauthorized, httptest.NewRequest(http.MethodGet, "/api/v1/tasks", nil))
	if unauthorized.Code != http.StatusUnauthorized {
		t.Fatalf("unauthorized status = %d, want %d", unauthorized.Code, http.StatusUnauthorized)
	}

	request := httptest.NewRequest(http.MethodGet, "/api/v1/tasks", nil)
	request.Header.Set("Authorization", "Bearer access-token")
	authorized := httptest.NewRecorder()
	router.ServeHTTP(authorized, request)
	if authorized.Code != http.StatusOK {
		t.Fatalf("authorized status = %d body=%s", authorized.Code, authorized.Body.String())
	}
	if !strings.Contains(authorized.Body.String(), `"tasks"`) {
		t.Fatalf("body = %s", authorized.Body.String())
	}
}

func TestRouterAcceptsGroupedHandlers(t *testing.T) {
	authHTTP := auth.NewHTTPHandler(fakeAuthApp{}, fakeTokenManager{}, false)
	tasksHTTP := tasks.NewHTTPHandler(fakeTasksApp{})
	router := NewRouter(HealthChecks{}, Handlers{
		Auth:  authHTTP,
		Tasks: tasksHTTP,
	})

	request := httptest.NewRequest(http.MethodGet, "/api/v1/tasks", nil)
	request.Header.Set("Authorization", "Bearer access-token")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
}

type fakeDashboardApp struct{}

func (fakeDashboardApp) GetSummary(context.Context, int64) (dashboard.Summary, error) {
	return dashboard.Summary{Metrics: []dashboard.Metric{{Label: "新增线索", Value: "328", Change: "+41%"}}}, nil
}

func TestDashboardRoutesAreMountedBehindAuth(t *testing.T) {
	authHTTP := auth.NewHTTPHandler(fakeAuthApp{}, fakeTokenManager{}, false)
	dashboardHTTP := dashboard.NewHTTPHandler(fakeDashboardApp{})
	router := NewRouter(HealthChecks{}, Handlers{Auth: authHTTP, Dashboard: dashboardHTTP})

	unauthorized := httptest.NewRecorder()
	router.ServeHTTP(unauthorized, httptest.NewRequest(http.MethodGet, "/api/v1/dashboard/summary", nil))
	if unauthorized.Code != http.StatusUnauthorized {
		t.Fatalf("unauthorized status = %d, want %d", unauthorized.Code, http.StatusUnauthorized)
	}

	request := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard/summary", nil)
	request.Header.Set("Authorization", "Bearer access-token")
	authorized := httptest.NewRecorder()
	router.ServeHTTP(authorized, request)
	if authorized.Code != http.StatusOK {
		t.Fatalf("authorized status = %d body=%s", authorized.Code, authorized.Body.String())
	}
	if !strings.Contains(authorized.Body.String(), `"metrics"`) {
		t.Fatalf("body = %s", authorized.Body.String())
	}
}

type fakeGrowthApp struct{}

func (fakeGrowthApp) CreateModel(context.Context, growth.CreateInput) (growth.Model, error) {
	return growth.Model{ID: 99, UserID: 42, Name: "标准方案", Result: growth.Result{MonthlyRevenue: 186000}}, nil
}
func (fakeGrowthApp) ListModels(context.Context, int64, int) ([]growth.Model, error) {
	return []growth.Model{{ID: 99, UserID: 42, Name: "标准方案"}}, nil
}
func (fakeGrowthApp) GetModel(context.Context, int64, int64) (growth.Model, error) {
	return growth.Model{ID: 99, UserID: 42, Name: "标准方案"}, nil
}

func TestGrowthRoutesAreMountedBehindAuth(t *testing.T) {
	authHTTP := auth.NewHTTPHandler(fakeAuthApp{}, fakeTokenManager{}, false)
	growthHTTP := growth.NewHTTPHandler(fakeGrowthApp{})
	router := NewRouter(HealthChecks{}, Handlers{Auth: authHTTP, Growth: growthHTTP})

	unauthorized := httptest.NewRecorder()
	router.ServeHTTP(unauthorized, httptest.NewRequest(http.MethodGet, "/api/v1/growth/models", nil))
	if unauthorized.Code != http.StatusUnauthorized {
		t.Fatalf("unauthorized status = %d, want %d", unauthorized.Code, http.StatusUnauthorized)
	}

	request := httptest.NewRequest(http.MethodGet, "/api/v1/growth/models", nil)
	request.Header.Set("Authorization", "Bearer access-token")
	authorized := httptest.NewRecorder()
	router.ServeHTTP(authorized, request)
	if authorized.Code != http.StatusOK {
		t.Fatalf("authorized status = %d body=%s", authorized.Code, authorized.Body.String())
	}
	if !strings.Contains(authorized.Body.String(), `"models"`) {
		t.Fatalf("body = %s", authorized.Body.String())
	}
}

type fakeCompetitorApp struct{}

func (fakeCompetitorApp) CreateScan(context.Context, competitor.CreateScanInput) (competitor.Scan, error) {
	return competitor.Scan{ID: 99, UserID: 42, Status: competitor.StatusCompleted}, nil
}
func (fakeCompetitorApp) ListScans(context.Context, int64, int) ([]competitor.Scan, error) {
	return []competitor.Scan{{ID: 99, UserID: 42, Status: competitor.StatusCompleted}}, nil
}
func (fakeCompetitorApp) GetScan(context.Context, int64, int64) (competitor.Scan, error) {
	return competitor.Scan{ID: 99, UserID: 42, Status: competitor.StatusCompleted}, nil
}
func (fakeCompetitorApp) GetMonitoring(context.Context, int64, int) (competitor.MonitoringSnapshot, error) {
	return competitor.MonitoringSnapshot{Watchlist: []competitor.WatchItem{{Name: "小鹅通", Threat: "high"}}}, nil
}

func TestCompetitorRoutesAreMountedBehindAuth(t *testing.T) {
	authHTTP := auth.NewHTTPHandler(fakeAuthApp{}, fakeTokenManager{}, false)
	competitorHTTP := competitor.NewHTTPHandler(fakeCompetitorApp{})
	router := NewRouter(HealthChecks{}, Handlers{Auth: authHTTP, Competitor: competitorHTTP})

	unauthorized := httptest.NewRecorder()
	router.ServeHTTP(unauthorized, httptest.NewRequest(http.MethodGet, "/api/v1/competitor/monitoring", nil))
	if unauthorized.Code != http.StatusUnauthorized {
		t.Fatalf("unauthorized status = %d, want %d", unauthorized.Code, http.StatusUnauthorized)
	}

	request := httptest.NewRequest(http.MethodGet, "/api/v1/competitor/monitoring", nil)
	request.Header.Set("Authorization", "Bearer access-token")
	authorized := httptest.NewRecorder()
	router.ServeHTTP(authorized, request)
	if authorized.Code != http.StatusOK {
		t.Fatalf("authorized status = %d body=%s", authorized.Code, authorized.Body.String())
	}
	if !strings.Contains(authorized.Body.String(), `"watchlist"`) {
		t.Fatalf("body = %s", authorized.Body.String())
	}
}

type fakeLearningApp struct{}

func (fakeLearningApp) ListCourses(context.Context, learning.CourseFilter) ([]learning.Course, error) {
	return []learning.Course{{Slug: "ai-basics", Title: "AI基础入门"}}, nil
}
func (fakeLearningApp) GetCourse(context.Context, string) (learning.Course, error) {
	return learning.Course{Slug: "ai-basics", Title: "AI基础入门"}, nil
}
func (fakeLearningApp) ListProgress(context.Context, int64) ([]learning.Progress, error) {
	return []learning.Progress{{CourseSlug: "ai-basics", CourseTitle: "AI基础入门", Percent: 42}}, nil
}
func (fakeLearningApp) CreateDiagnosis(context.Context, learning.CreateDiagnosisInput) (learning.Diagnosis, error) {
	return learning.Diagnosis{ID: 99, UserID: 42, Status: learning.DiagnosisCompleted}, nil
}
func (fakeLearningApp) LatestDiagnosis(context.Context, int64) (learning.Diagnosis, error) {
	return learning.Diagnosis{ID: 99, UserID: 42, Status: learning.DiagnosisCompleted}, nil
}

func TestLearningRoutesExposeCoursesAndProtectUserData(t *testing.T) {
	authHTTP := auth.NewHTTPHandler(fakeAuthApp{}, fakeTokenManager{}, false)
	learningHTTP := learning.NewHTTPHandler(fakeLearningApp{})
	router := NewRouter(HealthChecks{}, Handlers{Auth: authHTTP, Learning: learningHTTP})

	public := httptest.NewRecorder()
	router.ServeHTTP(public, httptest.NewRequest(http.MethodGet, "/api/v1/learning/courses", nil))
	if public.Code != http.StatusOK {
		t.Fatalf("public status = %d body=%s", public.Code, public.Body.String())
	}

	unauthorized := httptest.NewRecorder()
	router.ServeHTTP(unauthorized, httptest.NewRequest(http.MethodGet, "/api/v1/learning/progress", nil))
	if unauthorized.Code != http.StatusUnauthorized {
		t.Fatalf("unauthorized status = %d, want %d", unauthorized.Code, http.StatusUnauthorized)
	}

	request := httptest.NewRequest(http.MethodGet, "/api/v1/learning/progress", nil)
	request.Header.Set("Authorization", "Bearer access-token")
	authorized := httptest.NewRecorder()
	router.ServeHTTP(authorized, request)
	if authorized.Code != http.StatusOK {
		t.Fatalf("authorized status = %d body=%s", authorized.Code, authorized.Body.String())
	}
	if !strings.Contains(authorized.Body.String(), `"progress"`) {
		t.Fatalf("body = %s", authorized.Body.String())
	}
}

func TestRouterAppliesExplicitCORSOrigins(t *testing.T) {
	router := NewRouter(HealthChecks{
		AllowedOrigins: []string{"https://app.example.com"},
	}, Handlers{})

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

func TestRouterCORSAllowsPatchRequests(t *testing.T) {
	router := NewRouter(HealthChecks{
		AllowedOrigins: []string{"https://app.example.com"},
	}, Handlers{})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodOptions, "/api/v1/tasks/99", nil)
	request.Header.Set("Origin", "https://app.example.com")
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	methods := recorder.Header().Get("Access-Control-Allow-Methods")
	if !strings.Contains(methods, "PATCH") {
		t.Fatalf("Access-Control-Allow-Methods = %q, want PATCH", methods)
	}
}

func TestRouterLogsRequestsWithoutSensitiveHeaders(t *testing.T) {
	var buffer bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buffer, nil))
	router := NewRouter(HealthChecks{Logger: logger}, Handlers{})
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
	}, Handlers{Auth: authHTTP})

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
