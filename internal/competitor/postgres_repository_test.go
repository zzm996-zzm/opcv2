package competitor

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	pgxmock "github.com/pashagolub/pgxmock/v4"
)

func TestPostgresRepositoryCreatesScan(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 6, 30, 14, 0, 0, 0, time.UTC)
	db.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO competitor_scans (user_id, targets, focus, status, progress_percent, current_step, error_message, competitors, conclusions, evidence_sources, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $11)
		RETURNING id
	`)).
		WithArgs(
			int64(42),
			[]byte(`["小鹅通"]`),
			"价格变化",
			StatusQueued,
			0,
			"queued",
			"",
			pgxmock.AnyArg(),
			pgxmock.AnyArg(),
			pgxmock.AnyArg(),
			now,
		).
		WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(int64(99)))

	repository := NewPostgresRepository(db)
	scan, err := repository.CreateScan(context.Background(), Scan{
		UserID:      42,
		Targets:     []string{"小鹅通"},
		Focus:       "价格变化",
		Status:      StatusQueued,
		CurrentStep: "queued",
		Competitors: []Competitor{{Name: "小鹅通", Score: 91}},
		Conclusions: []Conclusion{{Title: "定位变化", Detail: "AI 私域增长"}},
		EvidenceSources: []EvidenceSource{{
			SourceType: "official_site",
			Title:      "小鹅通价格页",
			URL:        "https://example.com/pricing",
			Summary:    "套餐页新增 AI 助教权益",
			CapturedAt: now,
		}},
		CreatedAt: now,
	})
	if err != nil {
		t.Fatalf("CreateScan() error = %v", err)
	}
	if scan.ID != 99 {
		t.Fatalf("scan.ID = %d, want 99", scan.ID)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresRepositoryGetsOwnedScan(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 6, 30, 14, 0, 0, 0, time.UTC)
	db.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, user_id, targets, focus, status, progress_percent, current_step, error_message, competitors, conclusions, evidence_sources, created_at, updated_at
		FROM competitor_scans
		WHERE user_id = $1 AND id = $2
	`)).
		WithArgs(int64(42), int64(99)).
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "user_id", "targets", "focus", "status", "progress_percent", "current_step", "error_message", "competitors", "conclusions", "evidence_sources", "created_at", "updated_at",
		}).AddRow(
			int64(99),
			int64(42),
			[]byte(`["小鹅通"]`),
			"价格变化",
			StatusRunning,
			45,
			"collecting_sources",
			"",
			[]byte(`[{"name":"小鹅通","category":"知识付费","score":91,"signal":"新增 AI 助教","risk":"high","tags":["价格页更新"]}]`),
			[]byte(`[{"title":"定位变化","detail":"AI 私域增长"}]`),
			[]byte(`[{"source_type":"official_site","title":"小鹅通价格页","url":"https://example.com/pricing","summary":"套餐页新增 AI 助教权益","captured_at":"2026-06-30T14:00:00Z"}]`),
			now,
			now,
		))

	repository := NewPostgresRepository(db)
	scan, err := repository.GetScan(context.Background(), 42, 99)
	if err != nil {
		t.Fatalf("GetScan() error = %v", err)
	}
	if scan.ID != 99 || scan.Status != StatusRunning || scan.ProgressPercent != 45 || scan.CurrentStep != "collecting_sources" || scan.Competitors[0].Name != "小鹅通" || scan.EvidenceSources[0].Title != "小鹅通价格页" {
		t.Fatalf("scan = %+v", scan)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresRepositoryUpdatesScanStatus(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 6, 30, 14, 0, 0, 0, time.UTC)
	db.ExpectQuery(regexp.QuoteMeta(`
		UPDATE competitor_scans
		SET status = $2, progress_percent = $3, current_step = $4, error_message = $5, updated_at = NOW()
		WHERE id = $1
		RETURNING id, user_id, targets, focus, status, progress_percent, current_step, error_message, competitors, conclusions, evidence_sources, created_at, updated_at
	`)).
		WithArgs(int64(99), StatusRunning, 30, "collecting_sources", "").
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "user_id", "targets", "focus", "status", "progress_percent", "current_step", "error_message", "competitors", "conclusions", "evidence_sources", "created_at", "updated_at",
		}).AddRow(
			int64(99),
			int64(42),
			[]byte(`["小鹅通"]`),
			"价格变化",
			StatusRunning,
			30,
			"collecting_sources",
			"",
			[]byte(`[]`),
			[]byte(`[]`),
			[]byte(`[]`),
			now,
			now,
		))

	repository := NewPostgresRepository(db)
	scan, err := repository.UpdateScanStatus(context.Background(), 99, StatusRunning, 30, "collecting_sources", "")
	if err != nil {
		t.Fatalf("UpdateScanStatus() error = %v", err)
	}
	if scan.Status != StatusRunning || scan.ProgressPercent != 30 || scan.CurrentStep != "collecting_sources" {
		t.Fatalf("scan = %+v", scan)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresRepositoryStoresScanResults(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	db.ExpectExec(regexp.QuoteMeta(`
		UPDATE competitor_scans
		SET competitors = $2, conclusions = $3, evidence_sources = $4, updated_at = NOW()
		WHERE id = $1
	`)).
		WithArgs(
			int64(99),
			pgxmock.AnyArg(),
			pgxmock.AnyArg(),
			pgxmock.AnyArg(),
		).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	repository := NewPostgresRepository(db)
	err = repository.StoreScanResults(context.Background(), 99, ScanResult{
		Competitors: []Competitor{{Name: "小鹅通", Category: "知识付费", Score: 91, Risk: "high"}},
		Conclusions: []Conclusion{{Title: "定位变化", Detail: "竞品正在强化 AI 私域能力。"}},
		EvidenceSources: []EvidenceSource{{
			SourceType: "official_site",
			Title:      "小鹅通价格页",
			URL:        "https://example.com/pricing",
			Summary:    "套餐页新增 AI 助教权益",
			CapturedAt: time.Date(2026, 6, 30, 14, 0, 0, 0, time.UTC),
		}},
	})
	if err != nil {
		t.Fatalf("StoreScanResults() error = %v", err)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresRepositoryListsMonitoringRows(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 6, 30, 14, 0, 0, 0, time.UTC)
	db.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, name, category, status, threat, last_seen_at, channels, signal
		FROM competitor_watchlist
		WHERE user_id = $1
		ORDER BY last_seen_at DESC
		LIMIT $2
	`)).
		WithArgs(int64(42), 20).
		WillReturnRows(pgxmock.NewRows([]string{"id", "name", "category", "status", "threat", "last_seen_at", "channels", "signal"}).
			AddRow(int64(77), "小鹅通", "知识付费", "高频变化", "high", now, []byte(`["价格页"]`), "新增 AI 助教"))
	db.ExpectQuery(regexp.QuoteMeta(`
		SELECT occurred_at, company, title, detail, level
		FROM competitor_events
		WHERE user_id = $1
		ORDER BY occurred_at DESC
		LIMIT $2
	`)).
		WithArgs(int64(42), 20).
		WillReturnRows(pgxmock.NewRows([]string{"occurred_at", "company", "title", "detail", "level"}).
			AddRow(now, "小鹅通", "价格页新增 AI 助教权益", "套餐页新增权益", "high"))

	repository := NewPostgresRepository(db)
	watchlist, err := repository.ListWatchlist(context.Background(), 42, 20)
	if err != nil {
		t.Fatalf("ListWatchlist() error = %v", err)
	}
	events, err := repository.ListEvents(context.Background(), 42, 20)
	if err != nil {
		t.Fatalf("ListEvents() error = %v", err)
	}
	if len(watchlist) != 1 || watchlist[0].ID != 77 || watchlist[0].Name != "小鹅通" || len(events) != 1 || events[0].Title == "" {
		t.Fatalf("watchlist/events = %+v/%+v", watchlist, events)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresRepositoryCreatesWatchItem(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 7, 7, 9, 30, 0, 0, time.UTC)
	db.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO competitor_watchlist (user_id, name, category, status, threat, last_seen_at, channels, signal, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $9)
		RETURNING id, name, category, status, threat, last_seen_at, channels, signal
	`)).
		WithArgs(
			int64(42),
			"增长雷达",
			"商业情报",
			"监测中",
			"中",
			now,
			[]byte(`["价格页","招聘动态"]`),
			"已创建监测规则，等待首次巡检。",
			now,
		).
		WillReturnRows(pgxmock.NewRows([]string{"id", "name", "category", "status", "threat", "last_seen_at", "channels", "signal"}).
			AddRow(int64(77), "增长雷达", "商业情报", "监测中", "中", now, []byte(`["价格页","招聘动态"]`), "已创建监测规则，等待首次巡检。"))

	repository := NewPostgresRepository(db)
	item, err := repository.CreateWatchItem(context.Background(), WatchItem{
		UserID:     42,
		Name:       "增长雷达",
		Category:   "商业情报",
		Status:     "监测中",
		Threat:     "中",
		LastSeenAt: now,
		Channels:   []string{"价格页", "招聘动态"},
		Signal:     "已创建监测规则，等待首次巡检。",
	})
	if err != nil {
		t.Fatalf("CreateWatchItem() error = %v", err)
	}
	if item.ID != 77 || item.Name != "增长雷达" || len(item.Channels) != 2 {
		t.Fatalf("item = %+v", item)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresRepositoryDeletesWatchItemForUser(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	db.ExpectExec(regexp.QuoteMeta(`
		DELETE FROM competitor_watchlist
		WHERE user_id = $1 AND id = $2
	`)).
		WithArgs(int64(42), int64(77)).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	repository := NewPostgresRepository(db)
	if err := repository.DeleteWatchItem(context.Background(), 42, 77); err != nil {
		t.Fatalf("DeleteWatchItem() error = %v", err)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresRepositoryGetsWatchItemForUser(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 7, 7, 9, 30, 0, 0, time.UTC)
	db.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, name, category, status, threat, last_seen_at, channels, signal
		FROM competitor_watchlist
		WHERE user_id = $1 AND id = $2
	`)).
		WithArgs(int64(42), int64(77)).
		WillReturnRows(pgxmock.NewRows([]string{"id", "name", "category", "status", "threat", "last_seen_at", "channels", "signal"}).
			AddRow(int64(77), "增长雷达", "商业情报", "监测中", "中", now, []byte(`["价格页"]`), "等待首次巡检"))

	repository := NewPostgresRepository(db)
	item, err := repository.GetWatchItem(context.Background(), 42, 77)
	if err != nil {
		t.Fatalf("GetWatchItem() error = %v", err)
	}
	if item.ID != 77 || item.Name != "增长雷达" {
		t.Fatalf("item = %+v", item)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresRepositoryDeleteWatchItemReturnsNotFoundWhenNoRows(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	db.ExpectExec(regexp.QuoteMeta(`
		DELETE FROM competitor_watchlist
		WHERE user_id = $1 AND id = $2
	`)).
		WithArgs(int64(42), int64(77)).
		WillReturnResult(pgxmock.NewResult("DELETE", 0))

	repository := NewPostgresRepository(db)
	err = repository.DeleteWatchItem(context.Background(), 42, 77)
	if !errors.Is(err, ErrWatchItemNotFound) {
		t.Fatalf("err = %v, want ErrWatchItemNotFound", err)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
