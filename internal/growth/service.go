package growth

import (
	"context"
	"fmt"
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

func (s *Service) ModelScenarios(ctx context.Context, userID, id int64) (GrowthScenarios, error) {
	model, err := s.GetModel(ctx, userID, id)
	if err != nil {
		return GrowthScenarios{}, err
	}
	scenarios := []GrowthScenario{
		buildScenario("保守方案", scaleAssumptions(model.Assumptions, 0.72, 0.9, 0.9, 1, 0.85, 0.78), "适合冷启动：少投放，多依赖内容和社群转化。"),
		buildScenario("标准方案", model.Assumptions, "当前推荐：投放验证关键词，私域承接高意向线索。"),
		buildScenario("进攻方案", scaleAssumptions(model.Assumptions, 1.35, 1.1, 1.08, 1.05, 1.25, 1.35), "适合预算充足：快速放量，但需要客服和交付能力同步扩容。"),
	}
	return GrowthScenarios{
		ModelID:     model.ID,
		ModelName:   model.Name,
		Scenarios:   scenarios,
		GeneratedAt: model.UpdatedAt,
	}, nil
}

func (s *Service) ModelForecast(ctx context.Context, userID, id int64) (GrowthForecast, error) {
	model, err := s.GetModel(ctx, userID, id)
	if err != nil {
		return GrowthForecast{}, err
	}
	phases := []string{"验证渠道", "优化转化", "稳定投放", "扩大渠道", "复购加成"}
	ratios := []float64{0.45, 0.75, 1, 1.18, 1.35}
	progress := []int{28, 44, 60, 78, 90}
	months := make([]ForecastMonth, 0, len(ratios))
	for index, ratio := range ratios {
		months = append(months, ForecastMonth{
			Month:           fmt.Sprintf("第%d月", index+1),
			Revenue:         int(math.Round(float64(model.Result.MonthlyRevenue) * ratio)),
			Phase:           phases[index],
			ProgressPercent: progress[index],
		})
	}
	return GrowthForecast{
		ModelID:     model.ID,
		ModelName:   model.Name,
		Months:      months,
		GeneratedAt: model.UpdatedAt,
	}, nil
}

func (s *Service) ModelRecommendations(ctx context.Context, userID, id int64) (GrowthRecommendations, error) {
	model, err := s.GetModel(ctx, userID, id)
	if err != nil {
		return GrowthRecommendations{}, err
	}
	actionItems := []string{
		fmt.Sprintf("把客单价从 ¥%d 提升到 ¥%d，利润率会更稳。", model.Assumptions.AverageOrder, int(math.Round(float64(model.Assumptions.AverageOrder)*1.2))),
		"优先优化线索到成交转化率，比单纯买流量更划算。",
		"把高意向线索同步到 CRM，并设置 24 小时跟进提醒。",
	}
	if model.Assumptions.DealRate < 0.12 {
		actionItems[1] = "成交转化率偏低，先补齐顾问话术、案例证明和跟进节奏。"
	}
	return GrowthRecommendations{
		ModelID:   model.ID,
		ModelName: model.Name,
		Headline:  "先提成交率，再扩大预算",
		Summary: fmt.Sprintf(
			"当前模型每月预计产生 %d 条线索、%d 个成交，优先提升转化质量再放大渠道预算。",
			model.Result.Leads,
			model.Result.Deals,
		),
		CostItems: []CostItem{
			{Name: "内容生产", Amount: maxInt(12000, model.Assumptions.DeliveryCost/4), Detail: "短视频、文章、案例页"},
			{Name: "投放预算", Amount: model.Result.Leads * model.Assumptions.AcquisitionCost, Detail: "搜索词和信息流测试"},
			{Name: "工具订阅", Amount: 3200, Detail: "线索、CRM、自动化工具"},
			{Name: "交付人力", Amount: maxInt(9800, model.Assumptions.DeliveryCost/5), Detail: "顾问跟进与客户成功"},
		},
		ActionItems: actionItems,
		GeneratedAt: model.UpdatedAt,
	}, nil
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

func buildScenario(name string, assumptions Assumptions, highlight string) GrowthScenario {
	result := calculate(assumptions)
	return GrowthScenario{
		Name:      name,
		Revenue:   result.MonthlyRevenue,
		Cost:      costFor(assumptions, result),
		Margin:    result.NetMargin,
		Highlight: highlight,
	}
}

func scaleAssumptions(input Assumptions, visits, leadRate, dealRate, averageOrder, acquisitionCost, deliveryCost float64) Assumptions {
	return Assumptions{
		MonthlyVisits:   int(math.Round(float64(input.MonthlyVisits) * visits)),
		LeadRate:        clampRate(input.LeadRate * leadRate),
		DealRate:        clampRate(input.DealRate * dealRate),
		AverageOrder:    int(math.Round(float64(input.AverageOrder) * averageOrder)),
		AcquisitionCost: int(math.Round(float64(input.AcquisitionCost) * acquisitionCost)),
		DeliveryCost:    int(math.Round(float64(input.DeliveryCost) * deliveryCost)),
	}
}

func costFor(assumptions Assumptions, result Result) int {
	return result.Leads*assumptions.AcquisitionCost + assumptions.DeliveryCost
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
