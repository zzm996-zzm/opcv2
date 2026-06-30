package learning

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/zzm/opcv2/internal/auth"
)

type fakeApp struct {
	course         Course
	courses        []Course
	courseFilter   CourseFilter
	progress       []Progress
	diagnosis      Diagnosis
	diagnosisInput CreateDiagnosisInput
	progressUserID int64
	latestUserID   int64
	err            error
}

func (a *fakeApp) ListCourses(_ context.Context, filter CourseFilter) ([]Course, error) {
	a.courseFilter = filter
	return a.courses, a.err
}
func (a *fakeApp) GetCourse(context.Context, string) (Course, error) {
	return a.course, a.err
}
func (a *fakeApp) ListProgress(_ context.Context, userID int64) ([]Progress, error) {
	a.progressUserID = userID
	return a.progress, a.err
}
func (a *fakeApp) CreateDiagnosis(_ context.Context, input CreateDiagnosisInput) (Diagnosis, error) {
	a.diagnosisInput = input
	return a.diagnosis, a.err
}
func (a *fakeApp) LatestDiagnosis(_ context.Context, userID int64) (Diagnosis, error) {
	a.latestUserID = userID
	return a.diagnosis, a.err
}

func TestHTTPHandlerListsCourses(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	NewHTTPHandler(&fakeApp{courses: []Course{{Slug: "ai-basics", Title: "AI基础入门"}}}).RegisterPublic(router.Group("/api/v1"))

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/learning/courses", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"courses"`) || !strings.Contains(recorder.Body.String(), "AI基础入门") {
		t.Fatalf("body = %s", recorder.Body.String())
	}
}

func TestHTTPHandlerListsCoursesWithEmptyArrayAndCappedLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	app := &fakeApp{}
	NewHTTPHandler(app).RegisterPublic(router.Group("/api/v1"))

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/learning/courses?limit=500", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.courseFilter.Limit != 100 {
		t.Fatalf("limit = %d, want 100", app.courseFilter.Limit)
	}
	if !strings.Contains(recorder.Body.String(), `"courses":[]`) {
		t.Fatalf("body = %s", recorder.Body.String())
	}
}

func TestHTTPHandlerCreatesDiagnosisForAuthenticatedUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	group := router.Group("/api/v1")
	group.Use(func(c *gin.Context) {
		c.Set(auth.UserIDContextKey, int64(42))
		c.Next()
	})
	app := &fakeApp{diagnosis: Diagnosis{ID: 99, UserID: 42, Status: DiagnosisCompleted}}
	NewHTTPHandler(app).RegisterProtected(group)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/learning/diagnoses", strings.NewReader(`{"user_id":99,"goal":"提升AI能力","project":"智能客服"}`))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.diagnosisInput.UserID != 42 {
		t.Fatalf("diagnosis userID = %d, want authenticated user 42", app.diagnosisInput.UserID)
	}
	if !strings.Contains(recorder.Body.String(), `"id":99`) {
		t.Fatalf("body = %s", recorder.Body.String())
	}
}

func TestHTTPHandlerRejectsDiagnosisWithoutGoal(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	group := router.Group("/api/v1")
	group.Use(func(c *gin.Context) {
		c.Set(auth.UserIDContextKey, int64(42))
		c.Next()
	})
	NewHTTPHandler(&fakeApp{diagnosis: Diagnosis{ID: 99, UserID: 42, Status: DiagnosisCompleted}}).RegisterProtected(group)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/learning/diagnoses", strings.NewReader(`{"goal":"","project":"智能客服"}`))
	request.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"error":"invalid_request"`) {
		t.Fatalf("body = %s", recorder.Body.String())
	}
}

func TestHTTPHandlerListsProgressForAuthenticatedUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	group := router.Group("/api/v1")
	group.Use(func(c *gin.Context) {
		c.Set(auth.UserIDContextKey, int64(42))
		c.Next()
	})
	app := &fakeApp{progress: []Progress{{ID: 7, UserID: 42, CourseSlug: "ai-basics"}}}
	NewHTTPHandler(app).RegisterProtected(group)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/learning/progress", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.progressUserID != 42 {
		t.Fatalf("progress userID = %d, want authenticated user 42", app.progressUserID)
	}
}

func TestHTTPHandlerGetsLatestDiagnosisForAuthenticatedUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	group := router.Group("/api/v1")
	group.Use(func(c *gin.Context) {
		c.Set(auth.UserIDContextKey, int64(42))
		c.Next()
	})
	app := &fakeApp{diagnosis: Diagnosis{ID: 99, UserID: 42, Status: DiagnosisCompleted}}
	NewHTTPHandler(app).RegisterProtected(group)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/learning/diagnoses/latest", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.latestUserID != 42 {
		t.Fatalf("latest diagnosis userID = %d, want authenticated user 42", app.latestUserID)
	}
}
