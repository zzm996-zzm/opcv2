package competitor

import (
	"context"
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
		INSERT INTO competitor_scans (user_id, targets, focus, status, progress_percent, current_step, error_message, competitors, conclusions, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $10)
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
		CreatedAt:   now,
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
		SELECT id, user_id, targets, focus, status, progress_percent, current_step, error_message, competitors, conclusions, created_at, updated_at
		FROM competitor_scans
		WHERE user_id = $1 AND id = $2
	`)).
		WithArgs(int64(42), int64(99)).
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "user_id", "targets", "focus", "status", "progress_percent", "current_step", "error_message", "competitors", "conclusions", "created_at", "updated_at",
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
			now,
			now,
		))

	repository := NewPostgresRepository(db)
	scan, err := repository.GetScan(context.Background(), 42, 99)
	if err != nil {
		t.Fatalf("GetScan() error = %v", err)
	}
	if scan.ID != 99 || scan.Status != StatusRunning || scan.ProgressPercent != 45 || scan.CurrentStep != "collecting_sources" || scan.Competitors[0].Name != "小鹅通" {
		t.Fatalf("scan = %+v", scan)
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
		SELECT name, category, status, threat, last_seen_at, channels, signal
		FROM competitor_watchlist
		WHERE user_id = $1
		ORDER BY last_seen_at DESC
		LIMIT $2
	`)).
		WithArgs(int64(42), 20).
		WillReturnRows(pgxmock.NewRows([]string{"name", "category", "status", "threat", "last_seen_at", "channels", "signal"}).
			AddRow("小鹅通", "知识付费", "高频变化", "high", now, []byte(`["价格页"]`), "新增 AI 助教"))
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
	if len(watchlist) != 1 || watchlist[0].Name != "小鹅通" || len(events) != 1 || events[0].Title == "" {
		t.Fatalf("watchlist/events = %+v/%+v", watchlist, events)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
