package growth

import (
	"context"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Repository interface {
	CreateModel(ctx context.Context, model Model) (Model, error)
	ListModels(ctx context.Context, userID int64, limit int) ([]Model, error)
	GetModel(ctx context.Context, userID, id int64) (Model, error)
	CreateDraft(ctx context.Context, draft Draft) (Draft, error)
	GetDraft(ctx context.Context, userID, id int64) (Draft, error)
	UpdateDraft(ctx context.Context, draft Draft) (Draft, error)
	CreateSnapshot(ctx context.Context, snapshot ModelSnapshot) (ModelSnapshot, error)
	ListSnapshots(ctx context.Context, userID, modelID int64, limit int) ([]ModelSnapshot, error)
}

type Service struct {
	repository Repository
	now        func() time.Time
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository, now: time.Now}
}

func (s *Service) CreateDraft(ctx context.Context, input CreateDraftInput) (Draft, error) {
	if s.repository == nil {
		return Draft{}, ErrServiceNotReady
	}
	now := s.now()
	assumptions := extractAssumptions(input.Input)
	questions := missingQuestionsFor(assumptions, nil)
	status := DraftStatusNeedsInput
	if len(questions) == 0 {
		status = DraftStatusReady
	}
	return s.repository.CreateDraft(ctx, Draft{
		UserID:      input.UserID,
		Input:       strings.TrimSpace(input.Input),
		Status:      status,
		Assumptions: assumptions,
		Questions:   questions,
		Answers:     map[string]float64{},
		CreatedAt:   now,
		UpdatedAt:   now,
	})
}

func (s *Service) GetDraft(ctx context.Context, userID, id int64) (Draft, error) {
	if s.repository == nil {
		return Draft{}, ErrServiceNotReady
	}
	return s.repository.GetDraft(ctx, userID, id)
}

func (s *Service) AnswerDraft(ctx context.Context, input AnswerDraftInput) (Draft, error) {
	draft, err := s.GetDraft(ctx, input.UserID, input.DraftID)
	if err != nil {
		return Draft{}, err
	}
	if draft.Status == DraftStatusCalculated || !applyAnswers(&draft.Assumptions, input.Answers) {
		return Draft{}, ErrInvalidAnswers
	}
	if draft.Answers == nil {
		draft.Answers = map[string]float64{}
	}
	for key, value := range input.Answers {
		draft.Answers[key] = value
	}
	draft.Questions = missingQuestionsFor(draft.Assumptions, draft.Answers)
	draft.Status = DraftStatusNeedsInput
	if len(draft.Questions) == 0 {
		draft.Status = DraftStatusReady
	}
	draft.UpdatedAt = s.now()
	return s.repository.UpdateDraft(ctx, draft)
}

func (s *Service) CalculateDraft(ctx context.Context, input CalculateDraftInput) (DraftCalculation, error) {
	draft, err := s.GetDraft(ctx, input.UserID, input.DraftID)
	if err != nil {
		return DraftCalculation{}, err
	}
	if draft.Status != DraftStatusReady || len(missingQuestionsFor(draft.Assumptions, draft.Answers)) != 0 {
		return DraftCalculation{}, ErrDraftNotReady
	}
	name := strings.TrimSpace(input.Name)
	if name == "" {
		name = draftName(draft.Input)
	}
	model, err := s.CreateModel(ctx, CreateInput{
		UserID: input.UserID, Name: name,
		MonthlyVisits: draft.Assumptions.MonthlyVisits, LeadRate: draft.Assumptions.LeadRate,
		DealRate: draft.Assumptions.DealRate, AverageOrder: draft.Assumptions.AverageOrder,
		AcquisitionCost: draft.Assumptions.AcquisitionCost, DeliveryCost: draft.Assumptions.DeliveryCost,
	})
	if err != nil {
		return DraftCalculation{}, err
	}
	snapshot, err := s.createSnapshot(ctx, input.UserID, model)
	if err != nil {
		return DraftCalculation{}, err
	}
	draft.Status = DraftStatusCalculated
	draft.ModelID = &model.ID
	draft.UpdatedAt = s.now()
	draft, err = s.repository.UpdateDraft(ctx, draft)
	if err != nil {
		return DraftCalculation{}, err
	}
	return DraftCalculation{Draft: draft, Model: model, Snapshot: snapshot}, nil
}

func (s *Service) RecalculateModel(ctx context.Context, userID, id int64) (RecalculateResult, error) {
	model, err := s.GetModel(ctx, userID, id)
	if err != nil {
		return RecalculateResult{}, err
	}
	model.ID = 0
	model.Name = strings.TrimSpace(model.Name) + " · 再测算"
	model.CreatedAt = s.now()
	model.UpdatedAt = model.CreatedAt
	model, err = s.CreateModel(ctx, CreateInput{
		UserID:          userID,
		Name:            model.Name,
		MonthlyVisits:   model.Assumptions.MonthlyVisits,
		LeadRate:        model.Assumptions.LeadRate,
		DealRate:        model.Assumptions.DealRate,
		AverageOrder:    model.Assumptions.AverageOrder,
		AcquisitionCost: model.Assumptions.AcquisitionCost,
		DeliveryCost:    model.Assumptions.DeliveryCost,
	})
	if err != nil {
		return RecalculateResult{}, err
	}
	snapshot, err := s.createSnapshot(ctx, userID, model)
	if err != nil {
		return RecalculateResult{}, err
	}
	return RecalculateResult{Model: model, Snapshot: snapshot}, nil
}

func (s *Service) createSnapshot(ctx context.Context, userID int64, model Model) (ModelSnapshot, error) {
	scenarios, err := s.ModelScenarios(ctx, userID, model.ID)
	if err != nil {
		return ModelSnapshot{}, err
	}
	forecast, err := s.ModelForecast(ctx, userID, model.ID)
	if err != nil {
		return ModelSnapshot{}, err
	}
	recommendations, err := s.ModelRecommendations(ctx, userID, model.ID)
	if err != nil {
		return ModelSnapshot{}, err
	}
	return s.repository.CreateSnapshot(ctx, ModelSnapshot{
		UserID: userID, ModelID: model.ID, ModelName: model.Name,
		Assumptions: model.Assumptions, Result: model.Result,
		Scenarios: scenarios, Forecast: forecast, Recommendations: recommendations,
		CreatedAt: s.now(),
	})
}

func (s *Service) ListSnapshots(ctx context.Context, userID, modelID int64, limit int) ([]ModelSnapshot, error) {
	if s.repository == nil {
		return nil, ErrServiceNotReady
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if _, err := s.repository.GetModel(ctx, userID, modelID); err != nil {
		return nil, err
	}
	return s.repository.ListSnapshots(ctx, userID, modelID, limit)
}

var assumptionPatterns = map[string]*regexp.Regexp{
	"monthly_visits":   regexp.MustCompile(`(?:月访问量|月流量|每月访问量|月访客)[^0-9]{0,12}([0-9]+)`),
	"lead_rate":        regexp.MustCompile(`(?:线索转化率|留资率)[^0-9]{0,12}([0-9]+(?:\.[0-9]+)?)%`),
	"deal_rate":        regexp.MustCompile(`(?:成交转化率|成交率)[^0-9]{0,12}([0-9]+(?:\.[0-9]+)?)%`),
	"average_order":    regexp.MustCompile(`(?:平均客单价|客单价)[^0-9]{0,12}([0-9]+)`),
	"acquisition_cost": regexp.MustCompile(`(?:单条获客成本|获客成本)[^0-9]{0,12}([0-9]+)`),
	"delivery_cost":    regexp.MustCompile(`(?:月交付成本|交付成本)[^0-9]{0,12}([0-9]+)`),
}

func extractAssumptions(input string) Assumptions {
	values := map[string]float64{}
	for key, pattern := range assumptionPatterns {
		match := pattern.FindStringSubmatch(input)
		if len(match) != 2 {
			continue
		}
		value, err := strconv.ParseFloat(match[1], 64)
		if err == nil {
			values[key] = value
		}
	}
	if value, ok := values["lead_rate"]; ok {
		values["lead_rate"] = value / 100
	}
	if value, ok := values["deal_rate"]; ok {
		values["deal_rate"] = value / 100
	}
	var assumptions Assumptions
	applyAnswers(&assumptions, values)
	return assumptions
}

var clarificationQuestions = []ClarificationQuestion{
	{Key: "monthly_visits", Label: "预计月访问量", Unit: "次", Min: 1},
	{Key: "lead_rate", Label: "线索转化率", Unit: "%", Min: 0.01, Max: 100},
	{Key: "deal_rate", Label: "成交转化率", Unit: "%", Min: 0.01, Max: 100},
	{Key: "average_order", Label: "平均客单价", Unit: "元", Min: 1},
	{Key: "acquisition_cost", Label: "单条线索获客成本", Unit: "元", Min: 0},
	{Key: "delivery_cost", Label: "每月交付成本", Unit: "元", Min: 0},
}

func missingQuestionsFor(input Assumptions, answers map[string]float64) []ClarificationQuestion {
	missing := make([]ClarificationQuestion, 0, len(clarificationQuestions))
	for _, question := range clarificationQuestions {
		missingField := false
		switch question.Key {
		case "monthly_visits":
			missingField = input.MonthlyVisits <= 0
		case "lead_rate":
			missingField = input.LeadRate <= 0
		case "deal_rate":
			missingField = input.DealRate <= 0
		case "average_order":
			missingField = input.AverageOrder <= 0
		case "acquisition_cost":
			_, answered := answers[question.Key]
			missingField = input.AcquisitionCost <= 0 && !answered
		case "delivery_cost":
			_, answered := answers[question.Key]
			missingField = input.DeliveryCost <= 0 && !answered
		}
		if missingField {
			missing = append(missing, question)
		}
	}
	return missing
}

func applyAnswers(target *Assumptions, answers map[string]float64) bool {
	for key, value := range answers {
		if value < 0 {
			return false
		}
		switch key {
		case "monthly_visits":
			target.MonthlyVisits = int(math.Round(value))
		case "lead_rate":
			if value > 1 {
				value /= 100
			}
			if value <= 0 || value > 1 {
				return false
			}
			target.LeadRate = value
		case "deal_rate":
			if value > 1 {
				value /= 100
			}
			if value <= 0 || value > 1 {
				return false
			}
			target.DealRate = value
		case "average_order":
			target.AverageOrder = int(math.Round(value))
		case "acquisition_cost":
			target.AcquisitionCost = int(math.Round(value))
		case "delivery_cost":
			target.DeliveryCost = int(math.Round(value))
		default:
			return false
		}
	}
	return len(answers) > 0
}

func draftName(input string) string {
	runes := []rune(strings.TrimSpace(input))
	if len(runes) > 20 {
		runes = runes[:20]
	}
	if len(runes) == 0 {
		return "增长测算方案"
	}
	return string(runes)
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

type modelPageRepository interface {
	ListModelPage(context.Context, ListModelsInput) (ModelPage, error)
}

func (s *Service) ListModelPage(ctx context.Context, input ListModelsInput) (ModelPage, error) {
	if s.repository == nil {
		return ModelPage{}, ErrServiceNotReady
	}
	input.Query = strings.TrimSpace(input.Query)
	if input.Limit <= 0 || input.Limit > 100 {
		input.Limit = 20
	}
	if input.Offset < 0 {
		input.Offset = 0
	}
	if repository, ok := s.repository.(modelPageRepository); ok {
		return repository.ListModelPage(ctx, input)
	}
	models, err := s.repository.ListModels(ctx, input.UserID, input.Limit+input.Offset)
	if err != nil {
		return ModelPage{}, err
	}
	total := len(models)
	if input.Offset >= total {
		models = []Model{}
	} else {
		models = models[input.Offset:]
	}
	return ModelPage{Models: models, Total: total, Limit: input.Limit, Offset: input.Offset}, nil
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
