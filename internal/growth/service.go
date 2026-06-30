package growth

import (
	"context"
	"math"
	"strings"
	"time"
)

type Repository interface {
	CreateModel(ctx context.Context, model Model) (Model, error)
	ListModels(ctx context.Context, userID int64, limit int) ([]Model, error)
	GetModel(ctx context.Context, userID, id int64) (Model, error)
}

type Service struct {
	repository Repository
	now        func() time.Time
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository, now: time.Now}
}

func (s *Service) CreateModel(ctx context.Context, input CreateInput) (Model, error) {
	if s.repository == nil {
		return Model{}, ErrServiceNotReady
	}
	assumptions := Assumptions{
		MonthlyVisits:   maxInt(input.MonthlyVisits, 0),
		LeadRate:        clampRate(input.LeadRate),
		DealRate:        clampRate(input.DealRate),
		AverageOrder:    maxInt(input.AverageOrder, 0),
		AcquisitionCost: maxInt(input.AcquisitionCost, 0),
		DeliveryCost:    maxInt(input.DeliveryCost, 0),
	}
	now := s.now()
	return s.repository.CreateModel(ctx, Model{
		UserID:      input.UserID,
		Name:        strings.TrimSpace(input.Name),
		Assumptions: assumptions,
		Result:      calculate(assumptions),
		CreatedAt:   now,
	})
}

func (s *Service) ListModels(ctx context.Context, userID int64, limit int) ([]Model, error) {
	if s.repository == nil {
		return nil, ErrServiceNotReady
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return s.repository.ListModels(ctx, userID, limit)
}

func (s *Service) GetModel(ctx context.Context, userID, id int64) (Model, error) {
	if s.repository == nil {
		return Model{}, ErrServiceNotReady
	}
	return s.repository.GetModel(ctx, userID, id)
}

func calculate(input Assumptions) Result {
	leads := int(math.Round(float64(input.MonthlyVisits) * input.LeadRate))
	deals := int(math.Round(float64(leads) * input.DealRate))
	revenue := deals * input.AverageOrder
	cost := leads*input.AcquisitionCost + input.DeliveryCost
	margin := 0.0
	if revenue > 0 {
		margin = float64(revenue-cost) / float64(revenue)
	}
	paybackDays := 0
	if revenue > 0 {
		paybackDays = int(math.Ceil(float64(cost) / (float64(revenue) / 30)))
	}
	return Result{
		MonthlyRevenue: revenue,
		Leads:          leads,
		Deals:          deals,
		PaybackDays:    paybackDays,
		NetMargin:      math.Round(margin*100) / 100,
	}
}

func clampRate(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}
	return value
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
