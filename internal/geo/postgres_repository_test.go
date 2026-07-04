package geo

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	pgxmock "github.com/pashagolub/pgxmock/v4"
)

func TestPostgresRepositoryBuildsOverviewFromGeoTables(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	db.ExpectQuery(regexp.QuoteMeta(`
		SELECT key, label, value
		FROM geo_metrics
		WHERE user_id = $1
		ORDER BY sort_order ASC, id ASC
	`)).
		WithArgs(int64(42)).
		WillReturnRows(pgxmock.NewRows([]string{"key", "label", "value"}).
			AddRow("coverage", "AI引用覆盖", "12%"))
	db.ExpectQuery(regexp.QuoteMeta(`
		SELECT name, coverage_percent, status
		FROM geo_engine_coverages
		WHERE user_id = $1
		ORDER BY sort_order ASC, id ASC
	`)).
		WithArgs(int64(42)).
		WillReturnRows(pgxmock.NewRows([]string{"name", "coverage_percent", "status"}).
			AddRow("Perplexity", 22, "待优化"))
	db.ExpectQuery(regexp.QuoteMeta(`
		SELECT title, detail
		FROM geo_lead_signals
		WHERE user_id = $1
		ORDER BY sort_order ASC, id ASC
	`)).
		WithArgs(int64(42)).
		WillReturnRows(pgxmock.NewRows([]string{"title", "detail"}).
			AddRow("高意向问题", "来自后端的线索信号"))
	db.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, query, intent, coverage, score, action
		FROM geo_keyword_opportunities
		WHERE user_id = $1
		ORDER BY score DESC, id ASC
		LIMIT 50
	`)).
		WithArgs(int64(42)).
		WillReturnRows(pgxmock.NewRows([]string{"id", "query", "intent", "coverage", "score", "action"}).
			AddRow(int64(7), "AI客服选型后端关键词", "选型", "待覆盖", 71, "补充对比页"))
	db.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, type, title, priority, due_at
		FROM geo_content_tasks
		WHERE user_id = $1
		ORDER BY sort_order ASC, id ASC
		LIMIT 50
	`)).
		WithArgs(int64(42)).
		WillReturnRows(pgxmock.NewRows([]string{"id", "type", "title", "priority", "due_at"}).
			AddRow(int64(9), "对比页", "后端返回的内容任务", "高", "2026-07-08"))

	repository := NewPostgresRepository(db)
	overview, err := repository.Overview(context.Background(), 42)

	if err != nil {
		t.Fatalf("Overview() error = %v", err)
	}
	if len(overview.Stats) != 1 || overview.Stats[0].Value != "12%" {
		t.Fatalf("stats = %+v", overview.Stats)
	}
	if len(overview.Engines) != 1 || overview.Engines[0].Name != "Perplexity" {
		t.Fatalf("engines = %+v", overview.Engines)
	}
	if len(overview.LeadSignals) != 1 || overview.LeadSignals[0].Detail == "" {
		t.Fatalf("lead signals = %+v", overview.LeadSignals)
	}
	if len(overview.Keywords) != 1 || overview.Keywords[0].Query != "AI客服选型后端关键词" {
		t.Fatalf("keywords = %+v", overview.Keywords)
	}
	if len(overview.ContentTasks) != 1 || overview.ContentTasks[0].Title != "后端返回的内容任务" {
		t.Fatalf("content tasks = %+v", overview.ContentTasks)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresRepositoryReturnsEmptyOverviewArrays(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	db.ExpectQuery(regexp.QuoteMeta(`
		SELECT key, label, value
		FROM geo_metrics
		WHERE user_id = $1
		ORDER BY sort_order ASC, id ASC
	`)).
		WithArgs(int64(42)).
		WillReturnRows(pgxmock.NewRows([]string{"key", "label", "value"}))
	db.ExpectQuery(regexp.QuoteMeta(`
		SELECT name, coverage_percent, status
		FROM geo_engine_coverages
		WHERE user_id = $1
		ORDER BY sort_order ASC, id ASC
	`)).
		WithArgs(int64(42)).
		WillReturnRows(pgxmock.NewRows([]string{"name", "coverage_percent", "status"}))
	db.ExpectQuery(regexp.QuoteMeta(`
		SELECT title, detail
		FROM geo_lead_signals
		WHERE user_id = $1
		ORDER BY sort_order ASC, id ASC
	`)).
		WithArgs(int64(42)).
		WillReturnRows(pgxmock.NewRows([]string{"title", "detail"}))
	db.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, query, intent, coverage, score, action
		FROM geo_keyword_opportunities
		WHERE user_id = $1
		ORDER BY score DESC, id ASC
		LIMIT 50
	`)).
		WithArgs(int64(42)).
		WillReturnRows(pgxmock.NewRows([]string{"id", "query", "intent", "coverage", "score", "action"}))
	db.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, type, title, priority, due_at
		FROM geo_content_tasks
		WHERE user_id = $1
		ORDER BY sort_order ASC, id ASC
		LIMIT 50
	`)).
		WithArgs(int64(42)).
		WillReturnRows(pgxmock.NewRows([]string{"id", "type", "title", "priority", "due_at"}))

	repository := NewPostgresRepository(db)
	overview, err := repository.Overview(context.Background(), 42)

	if err != nil {
		t.Fatalf("Overview() error = %v", err)
	}
	if overview.Stats == nil || overview.Engines == nil || overview.LeadSignals == nil || overview.Keywords == nil || overview.ContentTasks == nil {
		t.Fatalf("overview should contain safe empty arrays: %+v", overview)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresRepositoryCreatesAnalysisRequest(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()
	now := time.Date(2026, 7, 4, 8, 0, 0, 0, time.UTC)

	db.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO geo_analysis_requests (user_id, target, status)
		VALUES ($1, $2, $3)
		RETURNING id, user_id, target, status, error_message, created_at, updated_at
	`)).
		WithArgs(int64(42), "面向制造业的 AI 质检工具", AnalysisRequestStatusQueued).
		WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "target", "status", "error_message", "created_at", "updated_at"}).
			AddRow(int64(7), int64(42), "面向制造业的 AI 质检工具", AnalysisRequestStatusQueued, "", now, now))

	repository := NewPostgresRepository(db)
	request, err := repository.CreateAnalysisRequest(context.Background(), 42, AnalysisRequestInput{Target: "面向制造业的 AI 质检工具"})

	if err != nil {
		t.Fatalf("CreateAnalysisRequest() error = %v", err)
	}
	if request.ID != 7 || request.UserID != 42 || request.Target != "面向制造业的 AI 质检工具" || request.Status != AnalysisRequestStatusQueued {
		t.Fatalf("request = %+v", request)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresRepositoryListsAnalysisRequests(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()
	now := time.Date(2026, 7, 4, 8, 0, 0, 0, time.UTC)

	db.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, user_id, target, status, error_message, created_at, updated_at
		FROM geo_analysis_requests
		WHERE user_id = $1
		ORDER BY created_at DESC, id DESC
		LIMIT $2
	`)).
		WithArgs(int64(42), 20).
		WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "target", "status", "error_message", "created_at", "updated_at"}).
			AddRow(int64(7), int64(42), "面向制造业的 AI 质检工具", AnalysisRequestStatusQueued, "", now, now))

	repository := NewPostgresRepository(db)
	requests, err := repository.ListAnalysisRequests(context.Background(), 42, 20)

	if err != nil {
		t.Fatalf("ListAnalysisRequests() error = %v", err)
	}
	if len(requests) != 1 || requests[0].ID != 7 || requests[0].Target != "面向制造业的 AI 质检工具" {
		t.Fatalf("requests = %+v", requests)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresRepositoryListsEmptyAnalysisRequests(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	db.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, user_id, target, status, error_message, created_at, updated_at
		FROM geo_analysis_requests
		WHERE user_id = $1
		ORDER BY created_at DESC, id DESC
		LIMIT $2
	`)).
		WithArgs(int64(42), 20).
		WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "target", "status", "error_message", "created_at", "updated_at"}))

	repository := NewPostgresRepository(db)
	requests, err := repository.ListAnalysisRequests(context.Background(), 42, 20)

	if err != nil {
		t.Fatalf("ListAnalysisRequests() error = %v", err)
	}
	if requests == nil {
		t.Fatal("requests should be an empty array, got nil")
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresRepositoryGetsAnalysisRequest(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()
	now := time.Date(2026, 7, 4, 8, 0, 0, 0, time.UTC)

	db.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, user_id, target, status, error_message, created_at, updated_at
		FROM geo_analysis_requests
		WHERE user_id = $1 AND id = $2
	`)).
		WithArgs(int64(42), int64(7)).
		WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "target", "status", "error_message", "created_at", "updated_at"}).
			AddRow(int64(7), int64(42), "面向制造业的 AI 质检工具", AnalysisRequestStatusQueued, "", now, now))

	repository := NewPostgresRepository(db)
	request, err := repository.GetAnalysisRequest(context.Background(), 42, 7)

	if err != nil {
		t.Fatalf("GetAnalysisRequest() error = %v", err)
	}
	if request.ID != 7 || request.UserID != 42 || request.Target == "" {
		t.Fatalf("request = %+v", request)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresRepositoryMapsMissingAnalysisRequest(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	db.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, user_id, target, status, error_message, created_at, updated_at
		FROM geo_analysis_requests
		WHERE user_id = $1 AND id = $2
	`)).
		WithArgs(int64(42), int64(7)).
		WillReturnError(pgx.ErrNoRows)

	repository := NewPostgresRepository(db)
	_, err = repository.GetAnalysisRequest(context.Background(), 42, 7)

	if !errors.Is(err, ErrAnalysisRequestNotFound) {
		t.Fatalf("err = %v, want ErrAnalysisRequestNotFound", err)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresRepositoryUpdatesAnalysisRequestStatus(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()
	now := time.Date(2026, 7, 4, 8, 0, 0, 0, time.UTC)

	db.ExpectQuery(regexp.QuoteMeta(`
		UPDATE geo_analysis_requests
		SET status = $2, error_message = $3, updated_at = NOW()
		WHERE id = $1
		RETURNING id, user_id, target, status, error_message, created_at, updated_at
	`)).
		WithArgs(int64(7), AnalysisRequestStatusFailed, "provider_unavailable").
		WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "target", "status", "error_message", "created_at", "updated_at"}).
			AddRow(int64(7), int64(42), "面向制造业的 AI 质检工具", AnalysisRequestStatusFailed, "provider_unavailable", now, now))

	repository := NewPostgresRepository(db)
	request, err := repository.UpdateAnalysisRequestStatus(context.Background(), 7, AnalysisRequestStatusFailed, "provider_unavailable")

	if err != nil {
		t.Fatalf("UpdateAnalysisRequestStatus() error = %v", err)
	}
	if request.Status != AnalysisRequestStatusFailed || request.ErrorMessage != "provider_unavailable" {
		t.Fatalf("request = %+v", request)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresRepositoryMapsMissingAnalysisRequestStatusUpdate(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	db.ExpectQuery(regexp.QuoteMeta(`
		UPDATE geo_analysis_requests
		SET status = $2, error_message = $3, updated_at = NOW()
		WHERE id = $1
		RETURNING id, user_id, target, status, error_message, created_at, updated_at
	`)).
		WithArgs(int64(7), AnalysisRequestStatusRunning, "").
		WillReturnError(pgx.ErrNoRows)

	repository := NewPostgresRepository(db)
	_, err = repository.UpdateAnalysisRequestStatus(context.Background(), 7, AnalysisRequestStatusRunning, "")

	if !errors.Is(err, ErrAnalysisRequestNotFound) {
		t.Fatalf("err = %v, want ErrAnalysisRequestNotFound", err)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
