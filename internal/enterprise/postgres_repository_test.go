package enterprise

import (
	"context"
	"regexp"
	"testing"
	"time"

	pgxmock "github.com/pashagolub/pgxmock/v4"
)

func TestPostgresRepositoryBuildsOverviewFromEnterpriseTables(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	db.ExpectQuery(regexp.QuoteMeta(`
		SELECT key, label, value
		FROM enterprise_metrics
		WHERE user_id = $1
		ORDER BY sort_order ASC, id ASC
	`)).
		WithArgs(int64(42)).
		WillReturnRows(pgxmock.NewRows([]string{"key", "label", "value"}).
			AddRow("companies", "服务企业数", "3"))
	db.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, title, audience, price_label, focus, result
		FROM enterprise_plans
		WHERE user_id = $1
		ORDER BY sort_order ASC, id ASC
		LIMIT 50
	`)).
		WithArgs(int64(42)).
		WillReturnRows(pgxmock.NewRows([]string{"id", "title", "audience", "price_label", "focus", "result"}).
			AddRow(int64(7), "后端陪跑方案", "增长团队", "待报价", []string{"诊断", "训练"}, "完成系统上线"))
	db.ExpectQuery(regexp.QuoteMeta(`
		SELECT stage, count, detail
		FROM enterprise_delivery_board
		WHERE user_id = $1
		ORDER BY sort_order ASC, id ASC
	`)).
		WithArgs(int64(42)).
		WillReturnRows(pgxmock.NewRows([]string{"stage", "count", "detail"}).
			AddRow("诊断中", 2, "后端交付阶段"))
	db.ExpectQuery(regexp.QuoteMeta(`
		SELECT time_label, title, detail
		FROM enterprise_milestones
		WHERE user_id = $1
		ORDER BY sort_order ASC, id ASC
		LIMIT 50
	`)).
		WithArgs(int64(42)).
		WillReturnRows(pgxmock.NewRows([]string{"time_label", "title", "detail"}).
			AddRow("第1周", "后端里程碑", "完成诊断"))
	db.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, company, result
		FROM enterprise_cases
		WHERE user_id = $1
		ORDER BY sort_order ASC, id ASC
		LIMIT 50
	`)).
		WithArgs(int64(42)).
		WillReturnRows(pgxmock.NewRows([]string{"id", "company", "result"}).
			AddRow(int64(9), "后端企业案例", "完成落地复盘"))

	repository := NewPostgresRepository(db)
	overview, err := repository.Overview(context.Background(), 42)

	if err != nil {
		t.Fatalf("Overview() error = %v", err)
	}
	if len(overview.Stats) != 1 || overview.Stats[0].Value != "3" {
		t.Fatalf("stats = %+v", overview.Stats)
	}
	if len(overview.Plans) != 1 || overview.Plans[0].Title != "后端陪跑方案" || len(overview.Plans[0].Focus) != 2 {
		t.Fatalf("plans = %+v", overview.Plans)
	}
	if len(overview.DeliveryBoard) != 1 || overview.DeliveryBoard[0].Count != 2 {
		t.Fatalf("delivery board = %+v", overview.DeliveryBoard)
	}
	if len(overview.Milestones) != 1 || overview.Milestones[0].Title != "后端里程碑" {
		t.Fatalf("milestones = %+v", overview.Milestones)
	}
	if len(overview.Cases) != 1 || overview.Cases[0].Company != "后端企业案例" {
		t.Fatalf("cases = %+v", overview.Cases)
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
		FROM enterprise_metrics
		WHERE user_id = $1
		ORDER BY sort_order ASC, id ASC
	`)).
		WithArgs(int64(42)).
		WillReturnRows(pgxmock.NewRows([]string{"key", "label", "value"}))
	db.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, title, audience, price_label, focus, result
		FROM enterprise_plans
		WHERE user_id = $1
		ORDER BY sort_order ASC, id ASC
		LIMIT 50
	`)).
		WithArgs(int64(42)).
		WillReturnRows(pgxmock.NewRows([]string{"id", "title", "audience", "price_label", "focus", "result"}))
	db.ExpectQuery(regexp.QuoteMeta(`
		SELECT stage, count, detail
		FROM enterprise_delivery_board
		WHERE user_id = $1
		ORDER BY sort_order ASC, id ASC
	`)).
		WithArgs(int64(42)).
		WillReturnRows(pgxmock.NewRows([]string{"stage", "count", "detail"}))
	db.ExpectQuery(regexp.QuoteMeta(`
		SELECT time_label, title, detail
		FROM enterprise_milestones
		WHERE user_id = $1
		ORDER BY sort_order ASC, id ASC
		LIMIT 50
	`)).
		WithArgs(int64(42)).
		WillReturnRows(pgxmock.NewRows([]string{"time_label", "title", "detail"}))
	db.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, company, result
		FROM enterprise_cases
		WHERE user_id = $1
		ORDER BY sort_order ASC, id ASC
		LIMIT 50
	`)).
		WithArgs(int64(42)).
		WillReturnRows(pgxmock.NewRows([]string{"id", "company", "result"}))

	repository := NewPostgresRepository(db)
	overview, err := repository.Overview(context.Background(), 42)

	if err != nil {
		t.Fatalf("Overview() error = %v", err)
	}
	if overview.Stats == nil || overview.Plans == nil || overview.DeliveryBoard == nil || overview.Milestones == nil || overview.Cases == nil {
		t.Fatalf("overview should contain safe empty arrays: %+v", overview)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresRepositoryCreatesDiagnosisRequest(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 7, 7, 9, 30, 0, 0, time.UTC)
	db.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO enterprise_diagnosis_requests (user_id, need)
		VALUES ($1, $2)
		RETURNING id, user_id, need, status, created_at, updated_at
	`)).
		WithArgs(int64(42), "30人销售团队需要AI获客陪跑").
		WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "need", "status", "created_at", "updated_at"}).
			AddRow(int64(7), int64(42), "30人销售团队需要AI获客陪跑", "submitted", now, now))

	repository := NewPostgresRepository(db)
	request, err := repository.CreateDiagnosisRequest(context.Background(), 42, DiagnosisRequestInput{Need: "30人销售团队需要AI获客陪跑"})

	if err != nil {
		t.Fatalf("CreateDiagnosisRequest() error = %v", err)
	}
	if request.ID != 7 || request.UserID != 42 || request.Status != "submitted" || request.CreatedAt != "2026-07-07T09:30:00Z" {
		t.Fatalf("request = %+v", request)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresRepositoryListsDiagnosisRequests(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 7, 7, 10, 30, 0, 0, time.UTC)
	db.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, user_id, need, status, created_at, updated_at
		FROM enterprise_diagnosis_requests
		WHERE user_id = $1
		ORDER BY created_at DESC, id DESC
		LIMIT $2
	`)).
		WithArgs(int64(42), 5).
		WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "need", "status", "created_at", "updated_at"}).
			AddRow(int64(7), int64(42), "30人销售团队需要AI获客陪跑", "submitted", now, now))

	repository := NewPostgresRepository(db)
	requests, err := repository.ListDiagnosisRequests(context.Background(), 42, 5)

	if err != nil {
		t.Fatalf("ListDiagnosisRequests() error = %v", err)
	}
	if len(requests) != 1 || requests[0].ID != 7 || requests[0].CreatedAt != "2026-07-07T10:30:00Z" {
		t.Fatalf("requests = %+v", requests)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
