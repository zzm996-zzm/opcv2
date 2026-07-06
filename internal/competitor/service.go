package competitor

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/zzm/opcv2/internal/membership"
)

type Repository interface {
	CreateScan(ctx context.Context, scan Scan) (Scan, error)
	ListScans(ctx context.Context, userID int64, limit int) ([]Scan, error)
	GetScan(ctx context.Context, userID, id int64) (Scan, error)
	ListWatchlist(ctx context.Context, userID int64, limit int) ([]WatchItem, error)
	ListEvents(ctx context.Context, userID int64, limit int) ([]Event, error)
}

type QuotaConsumer interface {
	CheckAndConsume(ctx context.Context, input membership.ConsumeInput) (membership.UsageItem, error)
}

type Option func(*Service)

type Service struct {
	repository Repository
	quota      QuotaConsumer
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
	return s.repository.CreateScan(ctx, Scan{
		UserID:      input.UserID,
		Targets:     targets,
		Focus:       focus,
		Status:      StatusCompleted,
		Competitors: defaultCompetitors(targets),
		Conclusions: defaultConclusions(),
		CreatedAt:   now,
	})
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
