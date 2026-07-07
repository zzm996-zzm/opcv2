package competitor

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
	input      CreateScanInput
	watchInput CreateWatchItemInput
	userID     int64
	scanID     int64
	watchID    int64
	limit      int
	scan       Scan
	watchItem  WatchItem
	scans      []Scan
	monitoring MonitoringSnapshot
	err        error
}

func (a *fakeApplication) CreateScan(_ context.Context, input CreateScanInput) (Scan, error) {
	a.input = input
	return a.scan, a.err
}

func (a *fakeApplication) ListScans(_ context.Context, userID int64, limit int) ([]Scan, error) {
	a.userID = userID
	a.limit = limit
	return a.scans, a.err
}

func (a *fakeApplication) GetScan(_ context.Context, userID, id int64) (Scan, error) {
	a.userID = userID
	a.scanID = id
	return a.scan, a.err
}

func (a *fakeApplication) RetryScan(_ context.Context, userID, id int64) (Scan, error) {
	a.userID = userID
	a.scanID = id
	return a.scan, a.err
}

func (a *fakeApplication) CreateWatchItem(_ context.Context, input CreateWatchItemInput) (WatchItem, error) {
	a.watchInput = input
	return a.watchItem, a.err
}

func (a *fakeApplication) DeleteWatchItem(_ context.Context, userID, id int64) error {
	a.userID = userID
	a.watchID = id
	return a.err
}

func (a *fakeApplication) StartWatchItemScan(_ context.Context, userID, id int64) (Scan, error) {
	a.userID = userID
	a.watchID = id
	return a.scan, a.err
}

func (a *fakeApplication) AddScanCompetitorToWatchlist(_ context.Context, userID, scanID int64, competitorName string) (WatchItem, error) {
	a.userID = userID
	a.scanID = scanID
	a.watchInput.Name = competitorName
	return a.watchItem, a.err
}

func (a *fakeApplication) GetMonitoring(_ context.Context, userID int64, limit int) (MonitoringSnapshot, error) {
	a.userID = userID
	a.limit = limit
	return a.monitoring, a.err
}

func competitorTestRouter(app Application) *gin.Engine {
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

func TestCreateScanEndpointUsesAuthenticatedUser(t *testing.T) {
	app := &fakeApplication{scan: Scan{ID: 99, UserID: 42, Status: StatusCompleted}}
	router := competitorTestRouter(app)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/competitor/scans", strings.NewReader(`{
		"targets":["小鹅通","有赞教育"],
		"focus":"价格、案例、招聘和 AI 功能"
	}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.input.UserID != 42 || len(app.input.Targets) != 2 {
		t.Fatalf("input = %+v", app.input)
	}
}

func TestCreateScanEndpointRejectsBlankTargets(t *testing.T) {
	app := &fakeApplication{scan: Scan{ID: 99, UserID: 42, Status: StatusCompleted}}
	router := competitorTestRouter(app)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/competitor/scans", strings.NewReader(`{
		"targets":[" ",""],
		"focus":"价格、案例、招聘和 AI 功能"
	}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if len(app.input.Targets) != 0 {
		t.Fatalf("CreateScan should not be called, input = %+v", app.input)
	}
}

func TestCreateScanEndpointRejectsMissingFocus(t *testing.T) {
	app := &fakeApplication{scan: Scan{ID: 99, UserID: 42, Status: StatusCompleted}}
	router := competitorTestRouter(app)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/competitor/scans", strings.NewReader(`{
		"targets":["小鹅通"],
		"focus":" "
	}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestListScansEndpointUsesAuthenticatedUser(t *testing.T) {
	app := &fakeApplication{scans: []Scan{{ID: 99, UserID: 42, Status: StatusCompleted}}}
	router := competitorTestRouter(app)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/competitor/scans", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.userID != 42 {
		t.Fatalf("userID = %d, want 42", app.userID)
	}
	if !strings.Contains(recorder.Body.String(), `"scans"`) {
		t.Fatalf("body = %s", recorder.Body.String())
	}
}

func TestListScansEndpointReturnsEmptyArrayAndCapsLimit(t *testing.T) {
	app := &fakeApplication{}
	router := competitorTestRouter(app)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/competitor/scans?limit=500", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.userID != 42 || app.limit != 100 {
		t.Fatalf("user/limit = %d/%d", app.userID, app.limit)
	}
	if !strings.Contains(recorder.Body.String(), `"scans":[]`) {
		t.Fatalf("body = %s", recorder.Body.String())
	}
}

func TestGetScanEndpointReturnsNotFoundForOtherUser(t *testing.T) {
	app := &fakeApplication{err: ErrScanNotFound}
	router := competitorTestRouter(app)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/competitor/scans/99", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestRetryScanEndpointUsesAuthenticatedUser(t *testing.T) {
	app := &fakeApplication{scan: Scan{ID: 99, UserID: 42, Status: StatusQueued}}
	router := competitorTestRouter(app)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/competitor/scans/99/retry", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.userID != 42 || app.scanID != 99 {
		t.Fatalf("user/scan = %d/%d", app.userID, app.scanID)
	}
	if !strings.Contains(recorder.Body.String(), `"status":"queued"`) {
		t.Fatalf("body = %s", recorder.Body.String())
	}
}

func TestMonitoringEndpointUsesAuthenticatedUser(t *testing.T) {
	app := &fakeApplication{monitoring: MonitoringSnapshot{
		Watchlist: []WatchItem{{Name: "小鹅通", Threat: "high"}},
		Events:    []Event{{Company: "小鹅通", Title: "价格页新增 AI 助教权益"}},
	}}
	router := competitorTestRouter(app)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/competitor/monitoring", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.userID != 42 || !strings.Contains(recorder.Body.String(), `"watchlist"`) {
		t.Fatalf("user/body = %d/%s", app.userID, recorder.Body.String())
	}
}

func TestCreateWatchItemEndpointUsesAuthenticatedUser(t *testing.T) {
	app := &fakeApplication{watchItem: WatchItem{Name: "增长雷达", Category: "商业情报", Status: "监测中", Threat: "中", Channels: []string{"价格页"}}}
	router := competitorTestRouter(app)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/competitor/monitoring/watchlist", strings.NewReader(`{
		"name":"增长雷达",
		"category":"商业情报",
		"channels":["价格页","招聘动态"]
	}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.watchInput.UserID != 42 || app.watchInput.Name != "增长雷达" || len(app.watchInput.Channels) != 2 {
		t.Fatalf("watch input = %+v", app.watchInput)
	}
	if !strings.Contains(recorder.Body.String(), `"name":"增长雷达"`) {
		t.Fatalf("body = %s", recorder.Body.String())
	}
}

func TestCreateWatchItemEndpointRejectsBlankName(t *testing.T) {
	app := &fakeApplication{}
	router := competitorTestRouter(app)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/competitor/monitoring/watchlist", strings.NewReader(`{
		"name":" ",
		"channels":["价格页"]
	}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.watchInput.UserID != 0 {
		t.Fatalf("CreateWatchItem should not be called, input = %+v", app.watchInput)
	}
}

func TestDeleteWatchItemEndpointUsesAuthenticatedUser(t *testing.T) {
	app := &fakeApplication{}
	router := competitorTestRouter(app)
	request := httptest.NewRequest(http.MethodDelete, "/api/v1/competitor/monitoring/watchlist/77", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.userID != 42 || app.watchID != 77 {
		t.Fatalf("user/watch = %d/%d", app.userID, app.watchID)
	}
	if !strings.Contains(recorder.Body.String(), `"deleted":true`) {
		t.Fatalf("body = %s", recorder.Body.String())
	}
}

func TestDeleteWatchItemEndpointRejectsInvalidID(t *testing.T) {
	app := &fakeApplication{}
	router := competitorTestRouter(app)
	request := httptest.NewRequest(http.MethodDelete, "/api/v1/competitor/monitoring/watchlist/0", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.watchID != 0 {
		t.Fatalf("DeleteWatchItem should not be called, watchID = %d", app.watchID)
	}
}

func TestStartWatchItemScanEndpointUsesAuthenticatedUser(t *testing.T) {
	app := &fakeApplication{scan: Scan{ID: 99, UserID: 42, Status: StatusQueued, Targets: []string{"增长雷达"}}}
	router := competitorTestRouter(app)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/competitor/monitoring/watchlist/77/scan", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.userID != 42 || app.watchID != 77 {
		t.Fatalf("user/watch = %d/%d", app.userID, app.watchID)
	}
	if !strings.Contains(recorder.Body.String(), `"status":"queued"`) {
		t.Fatalf("body = %s", recorder.Body.String())
	}
}

func TestStartWatchItemScanEndpointRejectsInvalidID(t *testing.T) {
	app := &fakeApplication{}
	router := competitorTestRouter(app)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/competitor/monitoring/watchlist/0/scan", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.watchID != 0 {
		t.Fatalf("StartWatchItemScan should not be called, watchID = %d", app.watchID)
	}
}

func TestAddScanCompetitorToWatchlistEndpointUsesAuthenticatedUser(t *testing.T) {
	app := &fakeApplication{watchItem: WatchItem{ID: 77, Name: "增长雷达", Category: "商业情报", Status: "监测中"}}
	router := competitorTestRouter(app)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/competitor/scans/99/watchlist", strings.NewReader(`{
		"competitor_name":"增长雷达"
	}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.userID != 42 || app.scanID != 99 || app.watchInput.Name != "增长雷达" {
		t.Fatalf("user/scan/name = %d/%d/%s", app.userID, app.scanID, app.watchInput.Name)
	}
	if !strings.Contains(recorder.Body.String(), `"name":"增长雷达"`) {
		t.Fatalf("body = %s", recorder.Body.String())
	}
}

func TestAddScanCompetitorToWatchlistEndpointRejectsBlankName(t *testing.T) {
	app := &fakeApplication{}
	router := competitorTestRouter(app)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/competitor/scans/99/watchlist", strings.NewReader(`{
		"competitor_name":" "
	}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.scanID != 0 {
		t.Fatalf("AddScanCompetitorToWatchlist should not be called, scanID = %d", app.scanID)
	}
}
