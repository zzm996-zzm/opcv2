package competitor

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/zzm/opcv2/internal/membership"
)

type fakeRepository struct {
	createdScan Scan
	scan        Scan
	scans       []Scan
	watchlist   []WatchItem
	events      []Event
	err         error
}

type fakeQuotaConsumer struct {
	consumed []membership.ConsumeInput
	err      error
}

func (c *fakeQuotaConsumer) CheckAndConsume(_ context.Context, input membership.ConsumeInput) (membership.UsageItem, error) {
	c.consumed = append(c.consumed, input)
	if c.err != nil {
		return membership.UsageItem{}, c.err
	}
	return membership.UsageItem{Key: input.FeatureKey, Used: 1, Limit: 200}, nil
}

func (r *fakeRepository) CreateScan(_ context.Context, scan Scan) (Scan, error) {
	r.createdScan = scan
	scan.ID = 99
	scan.UpdatedAt = scan.CreatedAt
	r.scan = scan
	return scan, r.err
}

func (r *fakeRepository) ListScans(_ context.Context, userID int64, limit int) ([]Scan, error) {
	if r.err != nil {
		return nil, r.err
	}
	rows := make([]Scan, 0, len(r.scans))
	for _, scan := range r.scans {
		if scan.UserID == userID {
			rows = append(rows, scan)
		}
	}
	return rows[:min(len(rows), limit)], nil
}

func (r *fakeRepository) GetScan(_ context.Context, userID, id int64) (Scan, error) {
	if r.err != nil {
		return Scan{}, r.err
	}
	if r.scan.UserID != userID || r.scan.ID != id {
		return Scan{}, ErrScanNotFound
	}
	return r.scan, nil
}

func (r *fakeRepository) ListWatchlist(_ context.Context, userID int64, limit int) ([]WatchItem, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.watchlist[:min(len(r.watchlist), limit)], nil
}

func (r *fakeRepository) ListEvents(_ context.Context, userID int64, limit int) ([]Event, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.events[:min(len(r.events), limit)], nil
}

func TestServiceCreateScanConsumesCompetitorQuota(t *testing.T) {
	repository := &fakeRepository{}
	quota := &fakeQuotaConsumer{}
	service := NewService(repository, WithQuotaConsumer(quota))

	_, err := service.CreateScan(context.Background(), CreateScanInput{
		UserID:  42,
		Targets: []string{"小鹅通"},
		Focus:   "价格变化",
	})

	if err != nil {
		t.Fatalf("CreateScan() error = %v", err)
	}
	if len(quota.consumed) != 1 {
		t.Fatalf("consumed = %+v, want one quota consume", quota.consumed)
	}
	consumed := quota.consumed[0]
	if consumed.FeatureKey != membership.FeatureCompetitorScans || consumed.IdempotencyKey == "" {
		t.Fatalf("consumed = %+v", consumed)
	}
}

func TestServiceCreateScanStopsWhenCompetitorQuotaExceeded(t *testing.T) {
	repository := &fakeRepository{}
	service := NewService(repository, WithQuotaConsumer(&fakeQuotaConsumer{err: membership.ErrQuotaExceeded}))

	_, err := service.CreateScan(context.Background(), CreateScanInput{
		UserID:  42,
		Targets: []string{"小鹅通"},
		Focus:   "价格变化",
	})

	if !errors.Is(err, membership.ErrQuotaExceeded) {
		t.Fatalf("err = %v, want ErrQuotaExceeded", err)
	}
	if repository.createdScan.UserID != 0 {
		t.Fatalf("created scan = %+v, want no scan created", repository.createdScan)
	}
}

func TestServiceCreatesScanWithDevelopmentResult(t *testing.T) {
	now := time.Date(2026, 6, 30, 14, 0, 0, 0, time.UTC)
	repository := &fakeRepository{}
	service := NewService(repository)
	service.now = func() time.Time { return now }

	scan, err := service.CreateScan(context.Background(), CreateScanInput{
		UserID:  42,
		Targets: []string{"小鹅通", "有赞教育"},
		Focus:   "价格、案例、招聘和 AI 功能",
	})

	if err != nil {
		t.Fatalf("CreateScan() error = %v", err)
	}
	if scan.ID != 99 || scan.Status != StatusCompleted || len(scan.Competitors) == 0 || len(scan.Conclusions) == 0 {
		t.Fatalf("scan = %+v", scan)
	}
	if repository.createdScan.UserID != 42 || repository.createdScan.CreatedAt != now {
		t.Fatalf("created = %+v", repository.createdScan)
	}
}

func TestServiceListsOnlyUserScans(t *testing.T) {
	repository := &fakeRepository{scans: []Scan{
		{ID: 1, UserID: 42, Focus: "我的扫描"},
		{ID: 2, UserID: 7, Focus: "别人的扫描"},
	}}
	service := NewService(repository)

	scans, err := service.ListScans(context.Background(), 42, 20)

	if err != nil {
		t.Fatalf("ListScans() error = %v", err)
	}
	if len(scans) != 1 || scans[0].Focus != "我的扫描" {
		t.Fatalf("scans = %+v", scans)
	}
}

func TestServiceRejectsOtherUsersScan(t *testing.T) {
	service := NewService(&fakeRepository{scan: Scan{ID: 99, UserID: 7}})

	_, err := service.GetScan(context.Background(), 42, 99)

	if !errors.Is(err, ErrScanNotFound) {
		t.Fatalf("err = %v, want ErrScanNotFound", err)
	}
}

func TestServiceReturnsMonitoringSnapshot(t *testing.T) {
	repository := &fakeRepository{
		watchlist: []WatchItem{{Name: "小鹅通", Threat: "high"}},
		events:    []Event{{Company: "小鹅通", Title: "价格页新增 AI 助教权益", Level: "high"}},
	}
	service := NewService(repository)

	snapshot, err := service.GetMonitoring(context.Background(), 42, 20)

	if err != nil {
		t.Fatalf("GetMonitoring() error = %v", err)
	}
	if len(snapshot.Watchlist) != 1 || len(snapshot.Events) != 1 {
		t.Fatalf("snapshot = %+v", snapshot)
	}
}
