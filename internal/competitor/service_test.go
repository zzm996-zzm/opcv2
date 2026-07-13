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
	createdScan        Scan
	createdWatch       WatchItem
	deletedWatchUserID int64
	deletedWatchID     int64
	scan               Scan
	scans              []Scan
	watchlist          []WatchItem
	events             []Event
	updates            []scanStatusUpdate
	results            ScanResult
	err                error
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
		RawSnapshots: []RawSnapshot{{
			Platform:   "official_site",
			Payload:    []byte(`{"price_page":"AI assistant added"}`),
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

func (r *fakeRepository) CreateWatchItem(_ context.Context, item WatchItem) (WatchItem, error) {
	r.createdWatch = item
	item.ID = 77
	item.LastSeenAt = item.LastSeenAt.UTC()
	r.watchlist = append([]WatchItem{item}, r.watchlist...)
	return item, r.err
}

func (r *fakeRepository) GetWatchItem(_ context.Context, userID, id int64) (WatchItem, error) {
	if r.err != nil {
		return WatchItem{}, r.err
	}
	for _, item := range r.watchlist {
		if item.UserID == userID && item.ID == id {
			return item, nil
		}
	}
	return WatchItem{}, ErrWatchItemNotFound
}

func (r *fakeRepository) DeleteWatchItem(_ context.Context, userID, id int64) error {
	r.deletedWatchUserID = userID
	r.deletedWatchID = id
	if r.err != nil {
		return r.err
	}
	next := r.watchlist[:0]
	for _, item := range r.watchlist {
		if item.ID != id {
			next = append(next, item)
		}
	}
	r.watchlist = next
	return nil
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
	if len(repository.results.Competitors) != 1 || len(repository.results.Conclusions) != 1 || len(repository.results.EvidenceSources) != 1 || len(repository.results.RawSnapshots) != 1 {
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

func TestServiceRetryScanRequeuesFailedScan(t *testing.T) {
	repository := &fakeRepository{scan: Scan{ID: 99, UserID: 42, Status: StatusFailed, ErrorMessage: "scanner_not_configured"}}
	queue := &fakeQueue{}
	service := NewService(repository, WithQueue(queue))

	scan, err := service.RetryScan(context.Background(), 42, 99)

	if err != nil {
		t.Fatalf("RetryScan() error = %v", err)
	}
	if scan.Status != StatusQueued || scan.ProgressPercent != 0 || scan.CurrentStep != StatusQueued || scan.ErrorMessage != "" {
		t.Fatalf("scan = %+v", scan)
	}
	if len(repository.updates) != 1 {
		t.Fatalf("updates = %+v, want one queued status update", repository.updates)
	}
	if len(queue.jobs) != 1 {
		t.Fatalf("jobs = %+v, want one retry job", queue.jobs)
	}
	job := queue.jobs[0]
	if job.Type != jobs.TypeCompetitorScan || job.IdempotencyKey != "competitor-scan-retry-99" || job.Payload["scan_id"] != int64(99) {
		t.Fatalf("job = %+v", job)
	}
}

func TestServiceRetryScanRejectsOtherUsersScan(t *testing.T) {
	service := NewService(&fakeRepository{scan: Scan{ID: 99, UserID: 7, Status: StatusFailed}}, WithQueue(&fakeQueue{}))

	_, err := service.RetryScan(context.Background(), 42, 99)

	if !errors.Is(err, ErrScanNotFound) {
		t.Fatalf("err = %v, want ErrScanNotFound", err)
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

func TestServiceCreateWatchItemNormalizesInput(t *testing.T) {
	now := time.Date(2026, 7, 7, 9, 30, 0, 0, time.UTC)
	repository := &fakeRepository{}
	service := NewService(repository)
	service.now = func() time.Time { return now }

	item, err := service.CreateWatchItem(context.Background(), CreateWatchItemInput{
		UserID:   42,
		Name:     " 增长雷达 ",
		Category: " 商业情报 ",
		Channels: []string{" 价格页 ", "", "招聘动态"},
	})

	if err != nil {
		t.Fatalf("CreateWatchItem() error = %v", err)
	}
	if item.Name != "增长雷达" || item.Category != "商业情报" || item.Status != "监测中" || item.Threat != "中" {
		t.Fatalf("item = %+v", item)
	}
	if item.LastSeenAt != now || len(item.Channels) != 2 || item.Channels[0] != "价格页" || item.Signal != "已创建监测规则，等待首次巡检。" {
		t.Fatalf("item = %+v", item)
	}
	if repository.createdWatch.UserID != 42 {
		t.Fatalf("created watch = %+v", repository.createdWatch)
	}
}

func TestServiceCreateWatchItemRejectsBlankName(t *testing.T) {
	service := NewService(&fakeRepository{})

	_, err := service.CreateWatchItem(context.Background(), CreateWatchItemInput{
		UserID:   42,
		Name:     " ",
		Channels: []string{"价格页"},
	})

	if !errors.Is(err, ErrInvalidWatchItem) {
		t.Fatalf("err = %v, want ErrInvalidWatchItem", err)
	}
}

func TestServiceDeleteWatchItemUsesAuthenticatedUser(t *testing.T) {
	repository := &fakeRepository{watchlist: []WatchItem{{ID: 77, UserID: 42, Name: "增长雷达"}}}
	service := NewService(repository)

	err := service.DeleteWatchItem(context.Background(), 42, 77)

	if err != nil {
		t.Fatalf("DeleteWatchItem() error = %v", err)
	}
	if repository.deletedWatchUserID != 42 || repository.deletedWatchID != 77 {
		t.Fatalf("deleted user/id = %d/%d", repository.deletedWatchUserID, repository.deletedWatchID)
	}
	if len(repository.watchlist) != 0 {
		t.Fatalf("watchlist = %+v, want deleted item removed", repository.watchlist)
	}
}

func TestServiceDeleteWatchItemRejectsInvalidID(t *testing.T) {
	service := NewService(&fakeRepository{})

	err := service.DeleteWatchItem(context.Background(), 42, 0)

	if !errors.Is(err, ErrInvalidWatchItem) {
		t.Fatalf("err = %v, want ErrInvalidWatchItem", err)
	}
}

func TestServiceStartWatchItemScanCreatesQueuedScan(t *testing.T) {
	repository := &fakeRepository{watchlist: []WatchItem{{ID: 77, UserID: 42, Name: "增长雷达"}}}
	queue := &fakeQueue{}
	service := NewService(repository, WithQueue(queue))

	scan, err := service.StartWatchItemScan(context.Background(), 42, 77)

	if err != nil {
		t.Fatalf("StartWatchItemScan() error = %v", err)
	}
	if scan.Status != StatusQueued || len(scan.Targets) != 1 || scan.Targets[0] != "增长雷达" {
		t.Fatalf("scan = %+v", scan)
	}
	if repository.createdScan.UserID != 42 || repository.createdScan.Focus != "价格、招聘、内容和产品变化" {
		t.Fatalf("created scan = %+v", repository.createdScan)
	}
	if len(queue.jobs) != 1 || queue.jobs[0].Type != jobs.TypeCompetitorScan {
		t.Fatalf("jobs = %+v, want one competitor scan job", queue.jobs)
	}
}

func TestServiceStartWatchItemScanRejectsOtherUsersItem(t *testing.T) {
	service := NewService(&fakeRepository{watchlist: []WatchItem{{ID: 77, UserID: 7, Name: "增长雷达"}}})

	_, err := service.StartWatchItemScan(context.Background(), 42, 77)

	if !errors.Is(err, ErrWatchItemNotFound) {
		t.Fatalf("err = %v, want ErrWatchItemNotFound", err)
	}
}

func TestServiceAddScanCompetitorToWatchlistCreatesWatchItem(t *testing.T) {
	now := time.Date(2026, 7, 7, 10, 30, 0, 0, time.UTC)
	repository := &fakeRepository{scan: Scan{
		ID:     99,
		UserID: 42,
		Competitors: []Competitor{{
			Name:     "增长雷达",
			Category: "商业情报",
			Signal:   "新增自动化竞品预警和任务派发能力",
			Risk:     "强",
		}},
	}}
	service := NewService(repository)
	service.now = func() time.Time { return now }

	item, err := service.AddScanCompetitorToWatchlist(context.Background(), 42, 99, "增长雷达")

	if err != nil {
		t.Fatalf("AddScanCompetitorToWatchlist() error = %v", err)
	}
	if item.Name != "增长雷达" || item.Category != "商业情报" || item.Threat != "强" || item.Signal != "新增自动化竞品预警和任务派发能力" {
		t.Fatalf("item = %+v", item)
	}
	if repository.createdWatch.UserID != 42 || repository.createdWatch.LastSeenAt != now {
		t.Fatalf("created watch = %+v item=%+v", repository.createdWatch, item)
	}
	if len(repository.createdWatch.Channels) != 3 {
		t.Fatalf("channels = %+v", repository.createdWatch.Channels)
	}
}

func TestServiceAddScanCompetitorToWatchlistRejectsMissingCompetitor(t *testing.T) {
	service := NewService(&fakeRepository{scan: Scan{ID: 99, UserID: 42, Competitors: []Competitor{{Name: "增长雷达"}}}})

	_, err := service.AddScanCompetitorToWatchlist(context.Background(), 42, 99, "不存在")

	if !errors.Is(err, ErrInvalidWatchItem) {
		t.Fatalf("err = %v, want ErrInvalidWatchItem", err)
	}
}
