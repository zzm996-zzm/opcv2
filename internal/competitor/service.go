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
	IsAdmin(ctx context.Context, userID int64) (bool, error)
	CreateScan(ctx context.Context, scan Scan) (Scan, error)
	ListScans(ctx context.Context, userID int64, limit int) ([]Scan, error)
	GetScan(ctx context.Context, userID, id int64) (Scan, error)
	UpdateScanStatus(ctx context.Context, id int64, status string, progressPercent int, currentStep string, errorMessage string) (Scan, error)
	StoreScanResults(ctx context.Context, id int64, result ScanResult) error
	CreateWatchItem(ctx context.Context, item WatchItem) (WatchItem, error)
	GetWatchItem(ctx context.Context, userID, id int64) (WatchItem, error)
	DeleteWatchItem(ctx context.Context, userID, id int64) error
	ListWatchlist(ctx context.Context, userID int64, limit int) ([]WatchItem, error)
	ListEvents(ctx context.Context, userID int64, limit int) ([]Event, error)
	ListScriptAccounts(ctx context.Context, platform string, limit int) ([]ScriptAccount, error)
	UpsertScriptAccount(ctx context.Context, account ScriptAccount) (ScriptAccount, error)
	AcquireScriptAccount(ctx context.Context, platform string, scanID int64, now time.Time) (ScriptAccount, int64, error)
	FinishScriptAccountRun(ctx context.Context, accountID, runID int64, status, errorCode string, cooldownUntil *time.Time, now time.Time) error
}

func (s *Service) ListScriptAccounts(ctx context.Context, adminUserID int64, platform string, limit int) ([]ScriptAccount, error) {
	if err := s.requireAdmin(ctx, adminUserID); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return s.repository.ListScriptAccounts(ctx, strings.TrimSpace(platform), limit)
}

func (s *Service) UpsertScriptAccount(ctx context.Context, input ScriptAccountInput) (ScriptAccount, error) {
	if err := s.requireAdmin(ctx, input.AdminUserID); err != nil {
		return ScriptAccount{}, err
	}
	input.Platform = strings.TrimSpace(input.Platform)
	input.AccountLabel = strings.TrimSpace(input.AccountLabel)
	input.CredentialRef = strings.TrimSpace(input.CredentialRef)
	input.Status = strings.TrimSpace(input.Status)
	if input.Status == "" {
		input.Status = ScriptAccountAvailable
	}
	if input.MaxRunsPerHour == 0 {
		input.MaxRunsPerHour = 10
	}
	if input.Platform == "" || input.AccountLabel == "" || input.MaxRunsPerHour <= 0 || !validScriptAccountStatus(input.Status) || (input.ID == 0 && input.CredentialRef == "") || (input.CredentialRef != "" && !validCredentialRef(input.CredentialRef)) {
		return ScriptAccount{}, ErrInvalidScriptAccount
	}
	now := s.now()
	item, err := s.repository.UpsertScriptAccount(ctx, ScriptAccount{
		ID:             input.ID,
		Platform:       input.Platform,
		AccountLabel:   input.AccountLabel,
		CredentialRef:  input.CredentialRef,
		HasCredential:  input.CredentialRef != "",
		Status:         input.Status,
		MaxRunsPerHour: input.MaxRunsPerHour,
		CreatedAt:      now,
		UpdatedAt:      now,
	})
	if err != nil {
		return ScriptAccount{}, err
	}
	item.CredentialRef = ""
	return item, nil
}

func validCredentialRef(value string) bool {
	for _, prefix := range []string{"op://", "vault://", "secret://"} {
		if strings.HasPrefix(value, prefix) && len(value) > len(prefix) {
			return true
		}
	}
	return false
}

func (s *Service) requireAdmin(ctx context.Context, userID int64) error {
	if s.repository == nil || userID <= 0 {
		return ErrAdminRequired
	}
	ok, err := s.repository.IsAdmin(ctx, userID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrAdminRequired
	}
	return nil
}

func validScriptAccountStatus(status string) bool {
	switch status {
	case ScriptAccountAvailable, ScriptAccountCooldown, ScriptAccountDisabled:
		return true
	default:
		return false
	}
}

type Queue interface {
	Enqueue(ctx context.Context, job jobs.Job) error
}

type Scanner interface {
	Platform() string
	Scan(ctx context.Context, scan Scan, account *ScriptAccount) (ScanResult, error)
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

func (s *Service) DeleteWatchItem(ctx context.Context, userID, id int64) error {
	if s.repository == nil {
		return ErrServiceNotReady
	}
	if id <= 0 {
		return ErrInvalidWatchItem
	}
	return s.repository.DeleteWatchItem(ctx, userID, id)
}

func (s *Service) StartWatchItemScan(ctx context.Context, userID, id int64) (Scan, error) {
	if s.repository == nil {
		return Scan{}, ErrServiceNotReady
	}
	if id <= 0 {
		return Scan{}, ErrInvalidWatchItem
	}
	item, err := s.repository.GetWatchItem(ctx, userID, id)
	if err != nil {
		return Scan{}, err
	}
	return s.CreateScan(ctx, CreateScanInput{
		UserID:  userID,
		Targets: []string{item.Name},
		Focus:   "价格、招聘、内容和产品变化",
	})
}

func (s *Service) AddScanCompetitorToWatchlist(ctx context.Context, userID, scanID int64, competitorName string) (WatchItem, error) {
	if s.repository == nil {
		return WatchItem{}, ErrServiceNotReady
	}
	name := strings.TrimSpace(competitorName)
	if name == "" {
		return WatchItem{}, ErrInvalidWatchItem
	}
	scan, err := s.repository.GetScan(ctx, userID, scanID)
	if err != nil {
		return WatchItem{}, err
	}
	for _, competitor := range scan.Competitors {
		if competitor.Name != name {
			continue
		}
		category := strings.TrimSpace(competitor.Category)
		if category == "" {
			category = "未分类竞品"
		}
		threat := strings.TrimSpace(competitor.Risk)
		if threat == "" {
			threat = "中"
		}
		signal := strings.TrimSpace(competitor.Signal)
		if signal == "" {
			signal = "已从全盘破解结果加入动态监测。"
		}
		return s.repository.CreateWatchItem(ctx, WatchItem{
			UserID:     userID,
			Name:       competitor.Name,
			Category:   category,
			Status:     "监测中",
			Threat:     threat,
			LastSeenAt: s.now(),
			Channels:   []string{"产品页", "价格页", "内容矩阵"},
			Signal:     signal,
		})
	}
	return WatchItem{}, ErrInvalidWatchItem
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
	var account *ScriptAccount
	var runID int64
	platform := strings.TrimSpace(s.scanner.Platform())
	if platform != "" {
		leased, leasedRunID, acquireErr := s.repository.AcquireScriptAccount(ctx, platform, id, s.now())
		if acquireErr != nil {
			_, updateErr := s.repository.UpdateScanStatus(ctx, id, StatusFailed, 100, "failed", "script_account_unavailable")
			if updateErr != nil {
				return updateErr
			}
			return acquireErr
		}
		account = &leased
		runID = leasedRunID
	}
	result, err := s.scanner.Scan(ctx, scan, account)
	if err != nil {
		if account != nil {
			status, cooldownUntil := failedAccountState(*account, s.now())
			if finishErr := s.repository.FinishScriptAccountRun(ctx, account.ID, runID, status, "scanner_failed", cooldownUntil, s.now()); finishErr != nil {
				return finishErr
			}
		}
		_, updateErr := s.repository.UpdateScanStatus(ctx, id, StatusFailed, 100, "failed", "scanner_failed")
		if updateErr != nil {
			return updateErr
		}
		return err
	}
	if account != nil {
		if err := s.repository.FinishScriptAccountRun(ctx, account.ID, runID, ScriptAccountAvailable, "", nil, s.now()); err != nil {
			return err
		}
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
	if result.RawSnapshots == nil {
		result.RawSnapshots = []RawSnapshot{}
	}
	if err := s.repository.StoreScanResults(ctx, id, result); err != nil {
		return err
	}
	_, err = s.repository.UpdateScanStatus(ctx, id, StatusSucceeded, 100, StatusSucceeded, "")
	return err
}

func failedAccountState(account ScriptAccount, now time.Time) (string, *time.Time) {
	if account.FailureCount+1 >= 5 {
		return ScriptAccountDisabled, nil
	}
	multiplier := account.FailureCount + 1
	if multiplier > 24 {
		multiplier = 24
	}
	cooldown := now.Add(time.Duration(multiplier) * 15 * time.Minute)
	return ScriptAccountCooldown, &cooldown
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
