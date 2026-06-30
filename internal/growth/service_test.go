package growth

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeRepository struct {
	created Model
	model   Model
	models  []Model
	err     error
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
