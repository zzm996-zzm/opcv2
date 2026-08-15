package growth

import (
	"errors"
	"time"
)

var (
	ErrServiceNotReady     = errors.New("growth service is not configured")
	ErrModelNotFound       = errors.New("growth model not found")
	ErrDraftNotFound       = errors.New("growth draft not found")
	ErrDraftNotReady       = errors.New("growth draft is not ready")
	ErrInvalidAnswers      = errors.New("growth draft answers are invalid")
	ErrInvalidExportFormat = errors.New("growth export format is invalid")
	ErrInvalidComparison   = errors.New("growth model comparison is invalid")
)

const (
	DraftStatusNeedsInput = "needs_input"
	DraftStatusReady      = "ready"
	DraftStatusCalculated = "calculated"
)

const growthModelVersion = "growth-calculator-v1"

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

type ListModelsInput struct {
	UserID int64
	Query  string
	Limit  int
	Offset int
}

type ModelPage struct {
	Models []Model `json:"models"`
	Total  int     `json:"total"`
	Limit  int     `json:"limit"`
	Offset int     `json:"offset"`
}

type CompareModelsInput struct {
	UserID   int64   `json:"-"`
	ModelIDs []int64 `json:"model_ids"`
}

type GrowthComparison struct {
	Models      []Model   `json:"models"`
	GeneratedAt time.Time `json:"generated_at"`
}

type ClarificationQuestion struct {
	Key   string  `json:"key"`
	Label string  `json:"label"`
	Unit  string  `json:"unit"`
	Min   float64 `json:"min"`
	Max   float64 `json:"max,omitempty"`
}

type Draft struct {
	ID          int64                   `json:"id"`
	UserID      int64                   `json:"user_id"`
	Input       string                  `json:"input"`
	Status      string                  `json:"status"`
	Assumptions Assumptions             `json:"assumptions"`
	Questions   []ClarificationQuestion `json:"questions"`
	Answers     map[string]float64      `json:"answers"`
	ModelID     *int64                  `json:"model_id,omitempty"`
	CreatedAt   time.Time               `json:"created_at"`
	UpdatedAt   time.Time               `json:"updated_at"`
}

type CreateDraftInput struct {
	UserID int64  `json:"-"`
	Input  string `json:"input"`
}

type AnswerDraftInput struct {
	UserID  int64              `json:"-"`
	DraftID int64              `json:"-"`
	Answers map[string]float64 `json:"answers"`
}

type CalculateDraftInput struct {
	UserID  int64  `json:"-"`
	DraftID int64  `json:"-"`
	Name    string `json:"name"`
}

type DraftCalculation struct {
	Draft    Draft         `json:"draft"`
	Model    Model         `json:"model"`
	Snapshot ModelSnapshot `json:"snapshot"`
}

type RecalculateResult struct {
	Model    Model         `json:"model"`
	Snapshot ModelSnapshot `json:"snapshot"`
}

type ExportModelInput struct {
	UserID  int64  `json:"-"`
	ModelID int64  `json:"-"`
	Format  string `json:"format"`
}

type GrowthReportExport struct {
	Model           Model                 `json:"model"`
	Scenarios       GrowthScenarios       `json:"scenarios"`
	Forecast        GrowthForecast        `json:"forecast"`
	Recommendations GrowthRecommendations `json:"recommendations"`
	Disclaimer      string                `json:"disclaimer"`
	ModelVersion    string                `json:"model_version"`
	GeneratedAt     time.Time             `json:"generated_at"`
}

type ModelSnapshot struct {
	ID              int64                 `json:"id"`
	UserID          int64                 `json:"user_id"`
	ModelID         int64                 `json:"model_id"`
	ModelName       string                `json:"model_name"`
	Assumptions     Assumptions           `json:"assumptions"`
	Result          Result                `json:"result"`
	Scenarios       GrowthScenarios       `json:"scenarios"`
	Forecast        GrowthForecast        `json:"forecast"`
	Recommendations GrowthRecommendations `json:"recommendations"`
	CreatedAt       time.Time             `json:"created_at"`
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

type GrowthRisk struct {
	Key          string `json:"key"`
	Name         string `json:"name"`
	Level        string `json:"level"`
	CurrentValue string `json:"current_value"`
	Threshold    string `json:"threshold"`
	Reason       string `json:"reason"`
	Suggestion   string `json:"suggestion"`
}

type GrowthRisks struct {
	ModelID      int64        `json:"model_id"`
	ModelName    string       `json:"model_name"`
	OverallLevel string       `json:"overall_level"`
	Risks        []GrowthRisk `json:"risks"`
	GeneratedAt  time.Time    `json:"generated_at"`
}

type GrowthActionItem struct {
	ID             string `json:"id"`
	Title          string `json:"title"`
	Detail         string `json:"detail"`
	OwnerRole      string `json:"owner_role"`
	TargetMetric   string `json:"target_metric"`
	ExpectedResult string `json:"expected_result"`
}

type GrowthActionPhase struct {
	Key   string             `json:"key"`
	Name  string             `json:"name"`
	Goal  string             `json:"goal"`
	Items []GrowthActionItem `json:"items"`
}

type GrowthActionPlan struct {
	ModelID     int64               `json:"model_id"`
	ModelName   string              `json:"model_name"`
	Phases      []GrowthActionPhase `json:"phases"`
	GeneratedAt time.Time           `json:"generated_at"`
}

type GrowthInputField struct {
	Key             string  `json:"key"`
	Label           string  `json:"label"`
	Value           float64 `json:"value"`
	Unit            string  `json:"unit"`
	Source          string  `json:"source"`
	Confidence      string  `json:"confidence"`
	ConfirmedByUser bool    `json:"confirmed_by_user"`
}

type GrowthInputs struct {
	ModelID             int64              `json:"model_id"`
	ModelName           string             `json:"model_name"`
	CompletenessPercent int                `json:"completeness_percent"`
	Fields              []GrowthInputField `json:"fields"`
	GeneratedAt         time.Time          `json:"generated_at"`
}
