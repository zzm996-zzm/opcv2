package membership

import (
	"context"
	"errors"
	"strings"
	"time"
)

const (
	PlanFree = "free"
	PlanPro  = "pro"
)

var (
	ErrCodeNotFound    = errors.New("redemption code not found")
	ErrCodeExpired     = errors.New("redemption code expired")
	ErrCodeExhausted   = errors.New("redemption code exhausted")
	ErrInvalidCode     = errors.New("invalid redemption code")
	ErrUserIDRequired  = errors.New("user id required")
	ErrServiceNotReady = errors.New("membership service is not configured")
)

type Plan struct {
	Code                 string `json:"code"`
	Name                 string `json:"name"`
	MonthlyAnalysisLimit int    `json:"monthly_analysis_limit"`
	LeadExportLimit      int    `json:"lead_export_limit"`
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

func normalizeCode(code string) string {
	return strings.ToUpper(strings.TrimSpace(code))
}
