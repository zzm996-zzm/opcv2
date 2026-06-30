package growth

import (
	"errors"
	"time"
)

var (
	ErrServiceNotReady = errors.New("growth service is not configured")
	ErrModelNotFound   = errors.New("growth model not found")
)

type CreateInput struct {
	UserID          int64   `json:"-"`
	Name            string  `json:"name"`
	MonthlyVisits   int     `json:"monthly_visits"`
	LeadRate        float64 `json:"lead_rate"`
	DealRate        float64 `json:"deal_rate"`
	AverageOrder    int     `json:"average_order"`
	AcquisitionCost int     `json:"acquisition_cost"`
	DeliveryCost    int     `json:"delivery_cost"`
}

type Assumptions struct {
	MonthlyVisits   int     `json:"monthly_visits"`
	LeadRate        float64 `json:"lead_rate"`
	DealRate        float64 `json:"deal_rate"`
	AverageOrder    int     `json:"average_order"`
	AcquisitionCost int     `json:"acquisition_cost"`
	DeliveryCost    int     `json:"delivery_cost"`
}

type Result struct {
	MonthlyRevenue int     `json:"monthly_revenue"`
	Leads          int     `json:"leads"`
	Deals          int     `json:"deals"`
	PaybackDays    int     `json:"payback_days"`
	NetMargin      float64 `json:"net_margin"`
}

type Model struct {
	ID          int64       `json:"id"`
	UserID      int64       `json:"user_id"`
	Name        string      `json:"name"`
	Assumptions Assumptions `json:"assumptions"`
	Result      Result      `json:"result"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}
