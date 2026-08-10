package membership

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	PlanFree = "free"
	PlanPro  = "pro"
)

const (
	FeatureSandboxRuns         = "sandbox_runs"
	FeatureCompetitorScans     = "competitor_scans"
	FeatureLeadTasks           = "lead_tasks"
	FeatureCopilotMessages     = "copilot_messages"
	FeatureCopilotCompareCalls = "copilot_compare_calls"
	FeatureCopilotFileAnalysis = "copilot_file_analysis"
	FeatureGeoAcquisition      = "geo_acquisition"
	FeatureAILeadDevelopment   = "ai_lead_development"
	FeatureAIChat              = "ai_chat"
	FeatureDashboard           = "dashboard"
	FeatureCRM                 = "crm"
)

var (
	ErrCodeNotFound    = errors.New("redemption code not found")
	ErrCodeExpired     = errors.New("redemption code expired")
	ErrCodeExhausted   = errors.New("redemption code exhausted")
	ErrInvalidCode     = errors.New("invalid redemption code")
	ErrInvalidCheckout = errors.New("invalid checkout")
	ErrInvalidConsume  = errors.New("invalid membership usage consume")
	ErrPlanNotFound    = errors.New("membership plan not found")
	ErrQuotaExceeded   = errors.New("membership quota exceeded")
	ErrQuotaNotFound   = errors.New("membership quota not found")
	ErrUserIDRequired  = errors.New("user id required")
	ErrServiceNotReady = errors.New("membership service is not configured")
)

type Plan struct {
	Code                 string `json:"code"`
	Name                 string `json:"name"`
	MonthlyAnalysisLimit int    `json:"monthly_analysis_limit"`
	LeadExportLimit      int    `json:"lead_export_limit"`
}

type PlanQuota struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	Limit int    `json:"limit"`
	Unit  string `json:"unit"`
}

type PlanOption struct {
	ID           int64       `json:"id"`
	Code         string      `json:"code"`
	Name         string      `json:"name"`
	PriceCents   int         `json:"price_cents"`
	BillingCycle string      `json:"billing_cycle"`
	Features     []string    `json:"features"`
	Quotas       []PlanQuota `json:"quotas"`
	Recommended  bool        `json:"recommended"`
}

type UsageItem struct {
	Key     string     `json:"key"`
	Label   string     `json:"label"`
	Used    int        `json:"used"`
	Limit   int        `json:"limit"`
	Unit    string     `json:"unit"`
	ResetAt *time.Time `json:"reset_at,omitempty"`
}

type FeatureAccess struct {
	Key           string `json:"key"`
	Label         string `json:"label"`
	Status        string `json:"status"`
	RequiredPlan  string `json:"required_plan,omitempty"`
	Message       string `json:"message,omitempty"`
	CTA           string `json:"cta,omitempty"`
	UpgradeURL    string `json:"upgrade_url,omitempty"`
	ContactURL    string `json:"contact_url,omitempty"`
	DataPolicy    string `json:"data_policy,omitempty"`
	AllowReadOnly bool   `json:"allow_read_only"`
	AllowWorkflow bool   `json:"allow_workflow"`
}

type FeatureAccessResponse struct {
	Features []FeatureAccess `json:"features"`
}

type ConsumeInput struct {
	UserID         int64
	FeatureKey     string
	Amount         int
	IdempotencyKey string
}

const (
	OrderPending = "pending"
	OrderPaid    = "paid"
)

type Order struct {
	ID          int64      `json:"id"`
	OrderNo     string     `json:"order_no"`
	PlanCode    string     `json:"plan_code"`
	AmountCents int        `json:"amount_cents"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	PaidAt      *time.Time `json:"paid_at,omitempty"`
}

type CheckoutInput struct {
	UserID       int64
	PlanCode     string `json:"plan_code"`
	BillingCycle string `json:"billing_cycle"`
	OrderNo      string `json:"-"`
}

type PaymentInfo struct {
	Mode    string `json:"mode"`
	Message string `json:"message"`
}

type CheckoutResult struct {
	Order   Order       `json:"order"`
	Payment PaymentInfo `json:"payment"`
}

var PlanCatalog = map[string]Plan{
	PlanFree: {
		Code:                 PlanFree,
		Name:                 "免费版",
		MonthlyAnalysisLimit: 3,
		LeadExportLimit:      0,
	},
	PlanPro: {
		Code:                 PlanPro,
		Name:                 "专业版",
		MonthlyAnalysisLimit: 100,
		LeadExportLimit:      1000,
	},
}

type Subscription struct {
	UserID   int64      `json:"user_id"`
	PlanCode string     `json:"plan_code"`
	Status   string     `json:"status"`
	StartsAt time.Time  `json:"starts_at"`
	EndsAt   *time.Time `json:"ends_at,omitempty"`
}

type CreditAccount struct {
	UserID  int64 `json:"user_id"`
	Balance int   `json:"balance"`
}

type CreditTransaction struct {
	UserID int64  `json:"user_id"`
	Amount int    `json:"amount"`
	Reason string `json:"reason"`
}

type RedemptionCode struct {
	Code           string
	PlanCode       string
	CreditAmount   int
	MaxRedemptions int
	RedeemedCount  int
	ExpiresAt      *time.Time
	Active         bool
}

type Redemption struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Code      string    `json:"code"`
	CreatedAt time.Time `json:"created_at"`
}

type Snapshot struct {
	Plan          Plan         `json:"plan"`
	Subscription  Subscription `json:"subscription,omitempty"`
	CreditBalance int          `json:"credit_balance"`
}

type RedeemInput struct {
	UserID int64
	Code   string
}

type RedeemResult struct {
	Snapshot        Snapshot   `json:"snapshot"`
	Redemption      Redemption `json:"redemption"`
	AlreadyRedeemed bool       `json:"already_redeemed"`
}

type Repository interface {
	CurrentSnapshot(ctx context.Context, userID int64, now time.Time) (Snapshot, error)
	RedeemCode(ctx context.Context, input RedeemInput, now time.Time) (RedeemResult, error)
	ListPlans(ctx context.Context) ([]PlanOption, error)
	CurrentUsage(ctx context.Context, userID int64, now time.Time) ([]UsageItem, error)
	CheckAndConsume(ctx context.Context, input ConsumeInput, now time.Time) (UsageItem, error)
	RefundUsage(ctx context.Context, input ConsumeInput, now time.Time) (UsageItem, error)
	ListOrders(ctx context.Context, userID int64, limit int) ([]Order, error)
	CreateCheckout(ctx context.Context, input CheckoutInput, now time.Time) (CheckoutResult, error)
}

type Service struct {
	repository Repository
	now        func() time.Time
}

func NewService(repository Repository) *Service {
	return &Service{
		repository: repository,
		now:        time.Now,
	}
}

func (s *Service) CurrentSnapshot(ctx context.Context, userID int64) (Snapshot, error) {
	if userID <= 0 {
		return Snapshot{}, ErrUserIDRequired
	}
	if s.repository == nil {
		return Snapshot{}, ErrServiceNotReady
	}
	return s.repository.CurrentSnapshot(ctx, userID, s.now())
}

func (s *Service) Redeem(ctx context.Context, input RedeemInput) (RedeemResult, error) {
	if input.UserID <= 0 {
		return RedeemResult{}, ErrUserIDRequired
	}
	input.Code = normalizeCode(input.Code)
	if input.Code == "" {
		return RedeemResult{}, ErrInvalidCode
	}
	if s.repository == nil {
		return RedeemResult{}, ErrServiceNotReady
	}
	return s.repository.RedeemCode(ctx, input, s.now())
}

func (s *Service) ListPlans(ctx context.Context) ([]PlanOption, error) {
	if s.repository == nil {
		return nil, ErrServiceNotReady
	}
	return s.repository.ListPlans(ctx)
}

func (s *Service) CurrentUsage(ctx context.Context, userID int64) ([]UsageItem, error) {
	if userID <= 0 {
		return nil, ErrUserIDRequired
	}
	if s.repository == nil {
		return nil, ErrServiceNotReady
	}
	return s.repository.CurrentUsage(ctx, userID, s.now())
}

func (s *Service) FeatureAccess(ctx context.Context, userID int64, keys []string) (FeatureAccessResponse, error) {
	if userID <= 0 {
		return FeatureAccessResponse{}, ErrUserIDRequired
	}
	planCode := PlanFree
	if s.repository != nil {
		snapshot, err := s.repository.CurrentSnapshot(ctx, userID, s.now())
		if err != nil {
			return FeatureAccessResponse{}, err
		}
		planCode = snapshot.Plan.Code
	}
	requested := map[string]bool{}
	for _, key := range keys {
		key = strings.TrimSpace(key)
		if key != "" {
			requested[key] = true
		}
	}
	features := []FeatureAccess{}
	for _, feature := range defaultFeatureAccess() {
		if len(requested) == 0 || requested[feature.Key] {
			if planCode == PlanPro {
				feature.Status = "available"
				feature.RequiredPlan = ""
				feature.Message = ""
				feature.CTA = ""
				feature.UpgradeURL = ""
				feature.ContactURL = ""
				feature.DataPolicy = "member_workflow"
				feature.AllowReadOnly = true
				feature.AllowWorkflow = true
			}
			features = append(features, feature)
		}
	}
	return FeatureAccessResponse{Features: features}, nil
}

func (s *Service) CheckAndConsume(ctx context.Context, input ConsumeInput) (UsageItem, error) {
	input, err := normalizeConsumeInput(input)
	if err != nil {
		return UsageItem{}, err
	}
	if s.repository == nil {
		return UsageItem{}, ErrServiceNotReady
	}
	return s.repository.CheckAndConsume(ctx, input, s.now())
}

func (s *Service) RefundUsage(ctx context.Context, input ConsumeInput) (UsageItem, error) {
	input, err := normalizeConsumeInput(input)
	if err != nil {
		return UsageItem{}, err
	}
	if s.repository == nil {
		return UsageItem{}, ErrServiceNotReady
	}
	return s.repository.RefundUsage(ctx, input, s.now())
}

func normalizeConsumeInput(input ConsumeInput) (ConsumeInput, error) {
	if input.UserID <= 0 {
		return ConsumeInput{}, ErrUserIDRequired
	}
	input.FeatureKey = strings.TrimSpace(input.FeatureKey)
	input.IdempotencyKey = strings.TrimSpace(input.IdempotencyKey)
	if input.Amount <= 0 {
		input.Amount = 1
	}
	if input.FeatureKey == "" || input.IdempotencyKey == "" {
		return ConsumeInput{}, ErrInvalidConsume
	}
	return input, nil
}

func (s *Service) ListOrders(ctx context.Context, userID int64, limit int) ([]Order, error) {
	if userID <= 0 {
		return nil, ErrUserIDRequired
	}
	if s.repository == nil {
		return nil, ErrServiceNotReady
	}
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return s.repository.ListOrders(ctx, userID, limit)
}

func (s *Service) CreateCheckout(ctx context.Context, input CheckoutInput) (CheckoutResult, error) {
	if input.UserID <= 0 {
		return CheckoutResult{}, ErrUserIDRequired
	}
	input.PlanCode = strings.ToLower(strings.TrimSpace(input.PlanCode))
	input.BillingCycle = strings.ToLower(strings.TrimSpace(input.BillingCycle))
	if input.PlanCode == "" || input.BillingCycle == "" {
		return CheckoutResult{}, ErrInvalidCheckout
	}
	if s.repository == nil {
		return CheckoutResult{}, ErrServiceNotReady
	}
	now := s.now()
	input.OrderNo = fmt.Sprintf("ZS-%s-%06d", now.Format("20060102"), input.UserID)
	return s.repository.CreateCheckout(ctx, input, now)
}

func normalizeCode(code string) string {
	return strings.ToUpper(strings.TrimSpace(code))
}

func defaultFeatureAccess() []FeatureAccess {
	lockedMessage := "该模块当前版本仅开放入口展示，真实工作流暂未对外启用。请升级或预约企业顾问确认开通方式。"
	return []FeatureAccess{
		{
			Key:           FeatureGeoAcquisition,
			Label:         "GEO获客",
			Status:        "locked",
			RequiredPlan:  PlanPro,
			Message:       lockedMessage,
			CTA:           "查看升级方案",
			UpgradeURL:    "/membership",
			ContactURL:    "/enterprise",
			DataPolicy:    "locked_no_workflow",
			AllowReadOnly: false,
			AllowWorkflow: false,
		},
		{
			Key:           FeatureAILeadDevelopment,
			Label:         "AI线索开发",
			Status:        "locked",
			RequiredPlan:  PlanPro,
			Message:       lockedMessage,
			CTA:           "查看升级方案",
			UpgradeURL:    "/membership",
			ContactURL:    "/enterprise",
			DataPolicy:    "locked_no_workflow",
			AllowReadOnly: false,
			AllowWorkflow: false,
		},
		{
			Key:           FeatureDashboard,
			Label:         "仪表盘",
			Status:        "locked",
			RequiredPlan:  PlanPro,
			Message:       lockedMessage,
			CTA:           "查看升级方案",
			UpgradeURL:    "/membership",
			ContactURL:    "/enterprise",
			DataPolicy:    "locked_no_workflow",
			AllowReadOnly: false,
			AllowWorkflow: false,
		},
		{
			Key:           FeatureCRM,
			Label:         "CRM客户管理",
			Status:        "locked",
			RequiredPlan:  PlanPro,
			Message:       lockedMessage,
			CTA:           "查看升级方案",
			UpgradeURL:    "/membership",
			ContactURL:    "/enterprise",
			DataPolicy:    "locked_no_workflow",
			AllowReadOnly: false,
			AllowWorkflow: false,
		},
	}
}
