package notifications

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zzm/opcv2/internal/auth"
)

type fakeApplication struct {
	rows    []Notification
	row     Notification
	summary Summary
	filters ListFilters
	userID  int64
	id      int64
	updated int
	err     error
}

func (a *fakeApplication) ListNotifications(_ context.Context, userID int64, filters ListFilters) ([]Notification, error) {
	a.userID = userID
	a.filters = filters
	return a.rows, a.err
}

func (a *fakeApplication) GetNotification(_ context.Context, userID, id int64) (Notification, error) {
	a.userID = userID
	a.id = id
	return a.row, a.err
}

func (a *fakeApplication) MarkRead(_ context.Context, userID, id int64) (Notification, error) {
	a.userID = userID
	a.id = id
	return a.row, a.err
}

func (a *fakeApplication) MarkAllRead(_ context.Context, userID int64) (int, error) {
	a.userID = userID
	return a.updated, a.err
}

func (a *fakeApplication) Summary(_ context.Context, userID int64) (Summary, error) {
	a.userID = userID
	return a.summary, a.err
}

func notificationsTestRouter(app Application) *gin.Engine {
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

func TestListNotificationsEndpointUsesFilters(t *testing.T) {
	now := time.Date(2026, 7, 2, 10, 0, 0, 0, time.UTC)
	app := &fakeApplication{rows: []Notification{{ID: 1, UserID: 42, Type: TypeTask, Title: "任务", CreatedAt: now}}}
	router := notificationsTestRouter(app)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/notifications?type=task&status=unread&limit=500", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.userID != 42 || app.filters.Type != TypeTask || app.filters.Status != StatusUnread || app.filters.Limit != 100 {
		t.Fatalf("user/filters = %d/%+v", app.userID, app.filters)
	}
	if !strings.Contains(recorder.Body.String(), `"notifications"`) {
		t.Fatalf("body = %s", recorder.Body.String())
	}
}

func TestGetNotificationEndpointParsesID(t *testing.T) {
	app := &fakeApplication{row: Notification{ID: 99, UserID: 42, Type: TypeCRM, Title: "跟进提醒"}}
	router := notificationsTestRouter(app)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/notifications/99", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.userID != 42 || app.id != 99 {
		t.Fatalf("user/id = %d/%d", app.userID, app.id)
	}
}

func TestMarkReadEndpointReturnsNotification(t *testing.T) {
	readAt := time.Date(2026, 7, 2, 10, 0, 0, 0, time.UTC)
	app := &fakeApplication{row: Notification{ID: 99, UserID: 42, Type: TypeTask, Title: "任务", ReadAt: &readAt}}
	router := notificationsTestRouter(app)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodPatch, "/api/v1/notifications/99/read", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.id != 99 || !strings.Contains(recorder.Body.String(), `"read_at"`) {
		t.Fatalf("id/body = %d/%s", app.id, recorder.Body.String())
	}
}

func TestMarkAllReadEndpointReturnsUpdatedCount(t *testing.T) {
	app := &fakeApplication{updated: 12}
	router := notificationsTestRouter(app)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/api/v1/notifications/read-all", nil))

	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), `"updated":12`) {
		t.Fatalf("status/body = %d/%s", recorder.Code, recorder.Body.String())
	}
}

func TestSummaryEndpointReturnsCounts(t *testing.T) {
	app := &fakeApplication{summary: Summary{Unread: 3, ByType: []TypeCount{{Type: TypeTask, Count: 2}}}}
	router := notificationsTestRouter(app)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/notifications/summary", nil))

	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), `"unread":3`) {
		t.Fatalf("status/body = %d/%s", recorder.Code, recorder.Body.String())
	}
}

func TestInvalidNotificationIDReturnsBadRequest(t *testing.T) {
	app := &fakeApplication{}
	router := notificationsTestRouter(app)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/notifications/nope", nil))

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
}
