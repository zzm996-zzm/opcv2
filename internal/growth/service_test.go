package growth

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeRepository struct {
	created    Model
	model      Model
	modelsByID map[int64]Model
	models     []Model
	err        error
}

func (r *fakeRepository) CreateDraft(_ context.Context, draft Draft) (Draft, error) {
	return draft, r.err
}
func (r *fakeRepository) GetDraft(_ context.Context, _, _ int64) (Draft, error) {
	return Draft{}, r.err
}
func (r *fakeRepository) UpdateDraft(_ context.Context, draft Draft) (Draft, error) {
	return draft, r.err
}
func (r *fakeRepository) CreateSnapshot(_ context.Context, snapshot ModelSnapshot) (ModelSnapshot, error) {
	return snapshot, r.err
}
func (r *fakeRepository) ListSnapshots(_ context.Context, _, _ int64, _ int) ([]ModelSnapshot, error) {
	return nil, r.err
}

func (r *fakeRepository) CreateModel(_ context.Context, model Model) (Model, error) {
	r.created = model
	model.ID = 99
	model.UpdatedAt = model.CreatedAt
	r.model = model
	return model, r.err
}

func (r *fakeRepository) ListModels(_ context.Context, userID int64, limit int) ([]Model, error) {
	if r.err != nil {
		return nil, r.err
	}
	rows := make([]Model, 0, len(r.models))
	for _, model := range r.models {
		if model.UserID == userID {
			rows = append(rows, model)
		}
	}
	return rows[:min(len(rows), limit)], nil
}

func (r *fakeRepository) GetModel(_ context.Context, userID, id int64) (Model, error) {
	if r.err != nil {
		return Model{}, r.err
	}
	if model, ok := r.modelsByID[id]; ok {
		if model.UserID != userID {
			return Model{}, ErrModelNotFound
		}
		return model, nil
	}
	if r.model.UserID != userID || r.model.ID != id {
		return Model{}, ErrModelNotFound
	}
	return r.model, nil
}

func TestServiceCreatesModelAndCalculatesResult(t *testing.T) {
	now := time.Date(2026, 6, 30, 13, 0, 0, 0, time.UTC)
	repository := &fakeRepository{}
	service := NewService(repository)
	service.now = func() time.Time { return now }

	model, err := service.CreateModel(context.Background(), CreateInput{
		UserID:          42,
		Name:            "智能客服系统 · 标准方案",
		MonthlyVisits:   24000,
		LeadRate:        0.068,
		DealRate:        0.14,
		AverageOrder:    820,
		AcquisitionCost: 42,
		DeliveryCost:    51000,
	})

	if err != nil {
		t.Fatalf("CreateModel() error = %v", err)
	}
	if model.ID != 99 || model.Result.MonthlyRevenue == 0 || model.Result.Leads == 0 || model.Result.Deals == 0 {
		t.Fatalf("model = %+v", model)
	}
	if repository.created.UserID != 42 || repository.created.CreatedAt != now {
		t.Fatalf("created = %+v", repository.created)
	}
}

func TestServiceListsOnlyUserModels(t *testing.T) {
	repository := &fakeRepository{models: []Model{
		{ID: 1, UserID: 42, Name: "我的模型"},
		{ID: 2, UserID: 7, Name: "别人的模型"},
	}}
	service := NewService(repository)

	models, err := service.ListModels(context.Background(), 42, 20)

	if err != nil {
		t.Fatalf("ListModels() error = %v", err)
	}
	if len(models) != 1 || models[0].Name != "我的模型" {
		t.Fatalf("models = %+v", models)
	}
}

func TestServiceRejectsOtherUsersModel(t *testing.T) {
	service := NewService(&fakeRepository{model: Model{ID: 99, UserID: 7}})

	_, err := service.GetModel(context.Background(), 42, 99)

	if !errors.Is(err, ErrModelNotFound) {
		t.Fatalf("err = %v, want ErrModelNotFound", err)
	}
}

func TestServiceDerivesGrowthViewsFromModel(t *testing.T) {
	model := Model{
		ID:     99,
		UserID: 42,
		Name:   "智能客服系统 · 标准方案",
		Assumptions: Assumptions{
			MonthlyVisits:   24000,
			LeadRate:        0.068,
			DealRate:        0.14,
			AverageOrder:    820,
			AcquisitionCost: 42,
			DeliveryCost:    51000,
		},
		Result:    Result{MonthlyRevenue: 186960, Leads: 1632, Deals: 228, PaybackDays: 20, NetMargin: 0.36},
		UpdatedAt: time.Date(2026, 6, 30, 8, 30, 0, 0, time.UTC),
	}
	service := NewService(&fakeRepository{model: model})

	scenarios, err := service.ModelScenarios(context.Background(), 42, 99)
	if err != nil {
		t.Fatalf("ModelScenarios() error = %v", err)
	}
	if scenarios.ModelID != 99 || len(scenarios.Scenarios) != 3 || scenarios.Scenarios[1].Name != "标准方案" {
		t.Fatalf("scenarios = %+v", scenarios)
	}

	forecast, err := service.ModelForecast(context.Background(), 42, 99)
	if err != nil {
		t.Fatalf("ModelForecast() error = %v", err)
	}
	if forecast.ModelID != 99 || len(forecast.Months) != 5 || forecast.Months[2].Revenue != model.Result.MonthlyRevenue {
		t.Fatalf("forecast = %+v", forecast)
	}

	recommendations, err := service.ModelRecommendations(context.Background(), 42, 99)
	if err != nil {
		t.Fatalf("ModelRecommendations() error = %v", err)
	}
	if recommendations.ModelID != 99 || len(recommendations.CostItems) != 4 || len(recommendations.ActionItems) != 3 {
		t.Fatalf("recommendations = %+v", recommendations)
	}
}

func TestServiceRecalculatesModelIntoNewSnapshot(t *testing.T) {
	now := time.Date(2026, 8, 15, 12, 0, 0, 0, time.UTC)
	repository := &fakeRepository{model: Model{
		ID: 99, UserID: 42, Name: "SaaS 增长模型",
		Assumptions: Assumptions{MonthlyVisits: 24000, LeadRate: 0.068, DealRate: 0.14, AverageOrder: 820, AcquisitionCost: 42, DeliveryCost: 51000},
		UpdatedAt:   now,
	}}
	service := NewService(repository)
	service.now = func() time.Time { return now }

	result, err := service.RecalculateModel(context.Background(), 42, 99)
	if err != nil {
		t.Fatalf("RecalculateModel() error = %v", err)
	}
	if result.Model.ID != 99 || result.Model.Name != "SaaS 增长模型 · 再测算" || result.Snapshot.ModelID != result.Model.ID {
		t.Fatalf("result = %+v", result)
	}
}

func TestServiceExportsStructuredReport(t *testing.T) {
	now := time.Date(2026, 8, 15, 12, 0, 0, 0, time.UTC)
	repository := &fakeRepository{model: Model{
		ID: 99, UserID: 42, Name: "SaaS 增长模型",
		Assumptions: Assumptions{MonthlyVisits: 24000, LeadRate: 0.068, DealRate: 0.14, AverageOrder: 820, AcquisitionCost: 42, DeliveryCost: 51000},
		Result:      Result{MonthlyRevenue: 186960, Leads: 1632, Deals: 228},
		UpdatedAt:   now,
	}}
	service := NewService(repository)
	service.now = func() time.Time { return now }

	export, err := service.ExportModel(context.Background(), ExportModelInput{UserID: 42, ModelID: 99, Format: "json"})
	if err != nil {
		t.Fatalf("ExportModel() error = %v", err)
	}
	if export.Model.ID != 99 || len(export.Scenarios.Scenarios) != 3 || len(export.Forecast.Months) != 5 || export.ModelVersion != "growth-calculator-v1" || export.GeneratedAt != now {
		t.Fatalf("export = %+v", export)
	}
	if export.Disclaimer == "" || len(export.Recommendations.ActionItems) == 0 {
		t.Fatalf("export metadata = %+v", export)
	}
}

func TestServiceRejectsUnsupportedExportFormat(t *testing.T) {
	service := NewService(&fakeRepository{})
	_, err := service.ExportModel(context.Background(), ExportModelInput{UserID: 42, ModelID: 99, Format: "pdf"})
	if !errors.Is(err, ErrInvalidExportFormat) {
		t.Fatalf("err = %v, want ErrInvalidExportFormat", err)
	}
}

func TestServiceComparesOwnedModels(t *testing.T) {
	now := time.Date(2026, 8, 15, 12, 0, 0, 0, time.UTC)
	repository := &fakeRepository{modelsByID: map[int64]Model{
		99:  {ID: 99, UserID: 42, Name: "初版", UpdatedAt: now},
		100: {ID: 100, UserID: 42, Name: "再次测算", UpdatedAt: now},
	}}
	service := NewService(repository)
	service.now = func() time.Time { return now }

	comparison, err := service.CompareModels(context.Background(), CompareModelsInput{UserID: 42, ModelIDs: []int64{99, 100}})
	if err != nil {
		t.Fatalf("CompareModels() error = %v", err)
	}
	if len(comparison.Models) != 2 || comparison.Models[0].ID != 99 || comparison.Models[1].ID != 100 || comparison.GeneratedAt != now {
		t.Fatalf("comparison = %+v", comparison)
	}
}

func TestServiceDerivesRuleBasedRisks(t *testing.T) {
	model := Model{
		ID: 99, UserID: 42, Name: "高风险模型",
		Assumptions: Assumptions{MonthlyVisits: 10000, LeadRate: 0.05, DealRate: 0.06, AverageOrder: 1000, AcquisitionCost: 250, DeliveryCost: 20000},
		Result:      Result{MonthlyRevenue: 30000, PaybackDays: 120, NetMargin: -0.2},
	}
	service := NewService(&fakeRepository{model: model})

	risks, err := service.ModelRisks(context.Background(), 42, 99)
	if err != nil {
		t.Fatalf("ModelRisks() error = %v", err)
	}
	if risks.OverallLevel != "high" || len(risks.Risks) != 4 {
		t.Fatalf("risks = %+v", risks)
	}
	for _, risk := range risks.Risks {
		if risk.Level != "high" || risk.CurrentValue == "" || risk.Threshold == "" || risk.Suggestion == "" {
			t.Fatalf("risk = %+v", risk)
		}
	}
}

func TestServiceRejectsInvalidComparisonSelection(t *testing.T) {
	service := NewService(&fakeRepository{})
	for _, ids := range [][]int64{{99}, {99, 99}, {99, 100, 101, 102, 103}} {
		_, err := service.CompareModels(context.Background(), CompareModelsInput{UserID: 42, ModelIDs: ids})
		if !errors.Is(err, ErrInvalidComparison) {
			t.Fatalf("ids=%v err=%v, want ErrInvalidComparison", ids, err)
		}
	}
}
