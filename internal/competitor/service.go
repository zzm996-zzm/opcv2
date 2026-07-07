package competitor

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/zzm/opcv2/internal/jobs"
	"github.com/zzm/opcv2/internal/membership"
)

type Repository interface {
	CreateScan(ctx context.Context, scan Scan) (Scan, error)
	ListScans(ctx context.Context, userID int64, limit int) ([]Scan, error)
	GetScan(ctx context.Context, userID, id int64) (Scan, error)
	UpdateScanStatus(ctx context.Context, id int64, status string, progressPercent int, currentStep string, errorMessage string) (Scan, error)
	StoreScanResults(ctx context.Context, id int64, result ScanResult) error
	CreateWatchItem(ctx context.Context, item WatchItem) (WatchItem, error)
	ListWatchlist(ctx context.Context, userID int64, limit int) ([]WatchItem, error)
	ListEvents(ctx context.Context, userID int64, limit int) ([]Event, error)
}

type Queue interface {
	Enqueue(ctx context.Context, job jobs.Job) error
}

type Scanner interface {
	Scan(ctx context.Context, scan Scan) (ScanResult, error)
}

type QuotaConsumer interface {
	CheckAndConsume(ctx context.Context, input membership.ConsumeInput) (membership.UsageItem, error)
}

type Option func(*Service)

type Service struct {
	repository Repository
	quota      QuotaConsumer
	queue      Queue
	scanner    Scanner
	now        func() time.Time
}

func NewService(repository Repository, options ...Option) *Service {
	service := &Service{repository: repository, now: time.Now}
	for _, option := range options {
		option(service)
	}
	return service
}

func WithQuotaConsumer(quota QuotaConsumer) Option {
	return func(service *Service) {
		service.quota = quota
	}
}

func WithQueue(queue Queue) Option {
	return func(service *Service) {
		service.queue = queue
	}
}

func WithScanner(scanner Scanner) Option {
	return func(service *Service) {
		service.scanner = scanner
	}
}

func (s *Service) CreateScan(ctx context.Context, input CreateScanInput) (Scan, error) {
	if s.repository == nil {
		return Scan{}, ErrServiceNotReady
	}
	targets := normalizeStrings(input.Targets)
	focus := strings.TrimSpace(input.Focus)
	if s.quota != nil {
		if _, err := s.quota.CheckAndConsume(ctx, membership.ConsumeInput{
			UserID:         input.UserID,
			FeatureKey:     membership.FeatureCompetitorScans,
			Amount:         1,
			IdempotencyKey: competitorScanIdempotencyKey(input.UserID, targets, focus),
		}); err != nil {
			return Scan{}, err
		}
	}
	now := s.now()
	scan, err := s.repository.CreateScan(ctx, Scan{
		UserID:          input.UserID,
		Targets:         targets,
		Focus:           focus,
		Status:          StatusQueued,
		ProgressPercent: 0,
		CurrentStep:     StatusQueued,
		Competitors:     []Competitor{},
		Conclusions:     []Conclusion{},
		EvidenceSources: []EvidenceSource{},
		CreatedAt:       now,
	})
	if err != nil {
		return Scan{}, err
	}
	if s.queue == nil {
		return scan, nil
	}
	if err := s.queue.Enqueue(ctx, jobs.Job{
		Type:           jobs.TypeCompetitorScan,
		IdempotencyKey: fmt.Sprintf("competitor-scan-%d", scan.ID),
		Payload: map[string]any{
			"scan_id": scan.ID,
		},
		MaxRetry: 3,
		Timeout:  10 * time.Minute,
	}); err != nil {
		return Scan{}, err
	}
	return scan, nil
}

func (s *Service) ListScans(ctx context.Context, userID int64, limit int) ([]Scan, error) {
	if s.repository == nil {
		return nil, ErrServiceNotReady
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return s.repository.ListScans(ctx, userID, limit)
}

func (s *Service) GetScan(ctx context.Context, userID, id int64) (Scan, error) {
	if s.repository == nil {
		return Scan{}, ErrServiceNotReady
	}
	return s.repository.GetScan(ctx, userID, id)
}

func (s *Service) RetryScan(ctx context.Context, userID, id int64) (Scan, error) {
	if s.repository == nil || s.queue == nil {
		return Scan{}, ErrServiceNotReady
	}
	scan, err := s.repository.GetScan(ctx, userID, id)
	if err != nil {
		return Scan{}, err
	}
	queued, err := s.repository.UpdateScanStatus(ctx, scan.ID, StatusQueued, 0, StatusQueued, "")
	if err != nil {
		return Scan{}, err
	}
	if err := s.queue.Enqueue(ctx, jobs.Job{
		Type:           jobs.TypeCompetitorScan,
		IdempotencyKey: fmt.Sprintf("competitor-scan-retry-%d", scan.ID),
		Payload: map[string]any{
			"scan_id": scan.ID,
		},
		MaxRetry: 3,
		Timeout:  10 * time.Minute,
	}); err != nil {
		return Scan{}, err
	}
	return queued, nil
}

func (s *Service) GetMonitoring(ctx context.Context, userID int64, limit int) (MonitoringSnapshot, error) {
	if s.repository == nil {
		return MonitoringSnapshot{}, ErrServiceNotReady
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	watchlist, err := s.repository.ListWatchlist(ctx, userID, limit)
	if err != nil {
		return MonitoringSnapshot{}, err
	}
	events, err := s.repository.ListEvents(ctx, userID, limit)
	if err != nil {
		return MonitoringSnapshot{}, err
	}
	if watchlist == nil {
		watchlist = []WatchItem{}
	}
	if events == nil {
		events = []Event{}
	}
	return MonitoringSnapshot{Watchlist: watchlist, Events: events}, nil
}

func (s *Service) CreateWatchItem(ctx context.Context, input CreateWatchItemInput) (WatchItem, error) {
	if s.repository == nil {
		return WatchItem{}, ErrServiceNotReady
	}
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return WatchItem{}, ErrInvalidWatchItem
	}
	category := strings.TrimSpace(input.Category)
	if category == "" {
		category = "未分类竞品"
	}
	channels := normalizeStrings(input.Channels)
	if len(channels) == 0 {
		channels = []string{"官网 / 价格页"}
	}
	return s.repository.CreateWatchItem(ctx, WatchItem{
		UserID:     input.UserID,
		Name:       name,
		Category:   category,
		Status:     "监测中",
		Threat:     "中",
		LastSeenAt: s.now(),
		Channels:   channels,
		Signal:     "已创建监测规则，等待首次巡检。",
	})
}

func (s *Service) ProcessScan(ctx context.Context, id int64) error {
	if s.repository == nil {
		return ErrServiceNotReady
	}
	scan, err := s.repository.UpdateScanStatus(ctx, id, StatusRunning, 30, "collecting_sources", "")
	if err != nil {
		return err
	}
	if s.scanner == nil {
		_, err = s.repository.UpdateScanStatus(ctx, id, StatusFailed, 100, "failed", "scanner_not_configured")
		return err
	}
	result, err := s.scanner.Scan(ctx, scan)
	if err != nil {
		_, updateErr := s.repository.UpdateScanStatus(ctx, id, StatusFailed, 100, "failed", strings.TrimSpace(err.Error()))
		if updateErr != nil {
			return updateErr
		}
		return err
	}
	if result.Competitors == nil {
		result.Competitors = []Competitor{}
	}
	if result.Conclusions == nil {
		result.Conclusions = []Conclusion{}
	}
	if result.EvidenceSources == nil {
		result.EvidenceSources = []EvidenceSource{}
	}
	if err := s.repository.StoreScanResults(ctx, id, result); err != nil {
		return err
	}
	_, err = s.repository.UpdateScanStatus(ctx, id, StatusSucceeded, 100, StatusSucceeded, "")
	return err
}

func defaultCompetitors(targets []string) []Competitor {
	if len(targets) == 0 {
		targets = []string{"小鹅通", "有赞教育", "企微管家"}
	}
	competitors := make([]Competitor, 0, len(targets))
	for index, target := range targets {
		score := 91 - index*7
		if score < 60 {
			score = 60
		}
		risk := "medium"
		if score >= 88 {
			risk = "high"
		}
		competitors = append(competitors, Competitor{
			Name:     target,
			Category: "公开竞品",
			Score:    score,
			Signal:   "近期公开页面和内容矩阵出现 AI、私域或增长相关信号。",
			Risk:     risk,
			Tags:     []string{"价格页更新", "内容变化", "招聘信号"},
		})
	}
	return competitors
}

func defaultConclusions() []Conclusion {
	return []Conclusion{
		{Title: "定位变化", Detail: "竞品正在从单点工具转向 AI + 私域增长方案。"},
		{Title: "价格策略", Detail: "低门槛入口用于获客，高阶能力绑定数据分析和私域运营。"},
		{Title: "反击建议", Detail: "优先补齐案例页和销售对比材料，并把高风险变化同步到任务中心。"},
	}
}

func normalizeStrings(values []string) []string {
	normalized := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			normalized = append(normalized, value)
		}
	}
	return normalized
}

func competitorScanIdempotencyKey(userID int64, targets []string, focus string) string {
	hash := sha256.Sum256([]byte(fmt.Sprintf("%d:%s:%s", userID, strings.Join(targets, "\x00"), focus)))
	return "competitor-scan-" + hex.EncodeToString(hash[:8])
}
