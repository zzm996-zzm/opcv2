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

type GrowthScenario struct {
	Name      string  `json:"name"`
	Revenue   int     `json:"revenue"`
	Cost      int     `json:"cost"`
	Margin    float64 `json:"margin"`
	Highlight string  `json:"highlight"`
}

type GrowthScenarios struct {
	ModelID     int64            `json:"model_id"`
	ModelName   string           `json:"model_name"`
	Scenarios   []GrowthScenario `json:"scenarios"`
	GeneratedAt time.Time        `json:"generated_at"`
}

type ForecastMonth struct {
	Month           string `json:"month"`
	Revenue         int    `json:"revenue"`
	Phase           string `json:"phase"`
	ProgressPercent int    `json:"progress_percent"`
}

type GrowthForecast struct {
	ModelID     int64           `json:"model_id"`
	ModelName   string          `json:"model_name"`
	Months      []ForecastMonth `json:"months"`
	GeneratedAt time.Time       `json:"generated_at"`
}

type CostItem struct {
	Name   string `json:"name"`
	Amount int    `json:"amount"`
	Detail string `json:"detail"`
}

type GrowthRecommendations struct {
	ModelID     int64      `json:"model_id"`
	ModelName   string     `json:"model_name"`
	Headline    string     `json:"headline"`
	Summary     string     `json:"summary"`
	CostItems   []CostItem `json:"cost_items"`
	ActionItems []string   `json:"action_items"`
	GeneratedAt time.Time  `json:"generated_at"`
}
