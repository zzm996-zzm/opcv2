package competitor

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/zzm/opcv2/internal/jobs"
	"github.com/zzm/opcv2/internal/membership"
)

type fakeRepository struct {
	createdScan Scan
	scan        Scan
	scans       []Scan
	watchlist   []WatchItem
	events      []Event
	updates     []scanStatusUpdate
	results     ScanResult
	err         error
}

type scanStatusUpdate struct {
	id              int64
	status          string
	progressPercent int
	currentStep     string
	errorMessage    string
}

type fakeQueue struct {
	jobs []jobs.Job
	err  error
}

type fakeScanner struct {
	scan Scan
	err  error
}

type fakeQuotaConsumer struct {
	consumed []membership.ConsumeInput
	err      error
}

func (q *fakeQueue) Enqueue(_ context.Context, job jobs.Job) error {
	q.jobs = append(q.jobs, job)
	return q.err
}

func (s *fakeScanner) Scan(_ context.Context, scan Scan) (ScanResult, error) {
	s.scan = scan
	if s.err != nil {
		return ScanResult{}, s.err
	}
	return ScanResult{
		Competitors: []Competitor{{Name: "小鹅通", Category: "知识付费", Score: 91, Risk: "high"}},
		Conclusions: []Conclusion{{Title: "定位变化", Detail: "竞品正在强化 AI 私域能力。"}},
		EvidenceSources: []EvidenceSource{{
			SourceType: "official_site",
			Title:      "小鹅通价格页",
			URL:        "https://example.com/pricing",
			Summary:    "套餐页新增 AI 助教权益",
			CapturedAt: time.Date(2026, 6, 30, 14, 0, 0, 0, time.UTC),
		}},
	}, nil
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

func (r *fakeRepository) UpdateScanStatus(_ context.Context, id int64, status string, progressPercent int, currentStep string, errorMessage string) (Scan, error) {
	r.updates = append(r.updates, scanStatusUpdate{
		id:              id,
		status:          status,
		progressPercent: progressPercent,
		currentStep:     currentStep,
		errorMessage:    errorMessage,
	})
	r.scan.Status = status
	r.scan.ProgressPercent = progressPercent
	r.scan.CurrentStep = currentStep
	r.scan.ErrorMessage = errorMessage
	return r.scan, r.err
}

func (r *fakeRepository) StoreScanResults(_ context.Context, id int64, result ScanResult) error {
	r.results = result
	r.scan.ID = id
	r.scan.Competitors = result.Competitors
	r.scan.Conclusions = result.Conclusions
	return r.err
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

func TestServiceCreatesQueuedScanTask(t *testing.T) {
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
	if scan.ID != 99 || scan.Status != StatusQueued || scan.ProgressPercent != 0 || scan.CurrentStep != "queued" {
		t.Fatalf("scan = %+v", scan)
	}
	if len(scan.Competitors) != 0 || len(scan.Conclusions) != 0 {
		t.Fatalf("queued scan should not include generated results: %+v", scan)
	}
	if repository.createdScan.UserID != 42 || repository.createdScan.CreatedAt != now {
		t.Fatalf("created = %+v", repository.createdScan)
	}
}

func TestServiceCreateScanEnqueuesCompetitorScanJob(t *testing.T) {
	repository := &fakeRepository{}
	queue := &fakeQueue{}
	service := NewService(repository, WithQueue(queue))

	scan, err := service.CreateScan(context.Background(), CreateScanInput{
		UserID:  42,
		Targets: []string{"小鹅通"},
		Focus:   "价格变化",
	})

	if err != nil {
		t.Fatalf("CreateScan() error = %v", err)
	}
	if scan.ID != 99 {
		t.Fatalf("scan.ID = %d, want 99", scan.ID)
	}
	if len(queue.jobs) != 1 {
		t.Fatalf("jobs = %+v, want one competitor scan job", queue.jobs)
	}
	job := queue.jobs[0]
	if job.Type != jobs.TypeCompetitorScan || job.IdempotencyKey != "competitor-scan-99" || job.Payload["scan_id"] != int64(99) {
		t.Fatalf("job = %+v", job)
	}
}

func TestServiceProcessScanStoresResultsAndMarksSucceeded(t *testing.T) {
	repository := &fakeRepository{scan: Scan{ID: 99, UserID: 42, Targets: []string{"小鹅通"}, Focus: "价格变化", Status: StatusQueued}}
	scanner := &fakeScanner{}
	service := NewService(repository, WithScanner(scanner))

	err := service.ProcessScan(context.Background(), 99)

	if err != nil {
		t.Fatalf("ProcessScan() error = %v", err)
	}
	if scanner.scan.ID != 99 || scanner.scan.Status != StatusRunning {
		t.Fatalf("scanner scan = %+v", scanner.scan)
	}
	if len(repository.results.Competitors) != 1 || len(repository.results.Conclusions) != 1 || len(repository.results.EvidenceSources) != 1 {
		t.Fatalf("results = %+v", repository.results)
	}
	if len(repository.updates) != 2 {
		t.Fatalf("updates = %+v, want running and succeeded", repository.updates)
	}
	if repository.updates[0].status != StatusRunning || repository.updates[0].progressPercent != 30 || repository.updates[0].currentStep != "collecting_sources" {
		t.Fatalf("running update = %+v", repository.updates[0])
	}
	if repository.updates[1].status != StatusSucceeded || repository.updates[1].progressPercent != 100 || repository.updates[1].currentStep != "succeeded" {
		t.Fatalf("succeeded update = %+v", repository.updates[1])
	}
}

func TestServiceProcessScanMarksFailedWhenScannerFails(t *testing.T) {
	cause := errors.New("script account pool unavailable")
	repository := &fakeRepository{scan: Scan{ID: 99, UserID: 42, Targets: []string{"小鹅通"}, Status: StatusQueued}}
	service := NewService(repository, WithScanner(&fakeScanner{err: cause}))

	err := service.ProcessScan(context.Background(), 99)

	if !errors.Is(err, cause) {
		t.Fatalf("ProcessScan() error = %v, want scanner cause", err)
	}
	if len(repository.updates) != 2 {
		t.Fatalf("updates = %+v, want running and failed", repository.updates)
	}
	failed := repository.updates[1]
	if failed.status != StatusFailed || failed.progressPercent != 100 || failed.currentStep != "failed" || failed.errorMessage != cause.Error() {
		t.Fatalf("failed update = %+v", failed)
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
