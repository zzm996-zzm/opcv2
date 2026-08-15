package growth

import (
	"context"
	"testing"
	"time"
)

type draftRepository struct {
	fakeRepository
	draft Draft
}

func (r *draftRepository) CreateDraft(_ context.Context, draft Draft) (Draft, error) {
	draft.ID = 71
	r.draft = draft
	return draft, nil
}

func (r *draftRepository) GetDraft(_ context.Context, userID, id int64) (Draft, error) {
	if r.draft.UserID != userID || r.draft.ID != id {
		return Draft{}, ErrDraftNotFound
	}
	return r.draft, nil
}

func (r *draftRepository) UpdateDraft(_ context.Context, draft Draft) (Draft, error) {
	r.draft = draft
	return draft, nil
}

func TestServiceCreatesDraftAndOnlyAsksForMissingAssumptions(t *testing.T) {
	now := time.Date(2026, 7, 11, 9, 0, 0, 0, time.UTC)
	repository := &draftRepository{}
	service := NewService(repository)
	service.now = func() time.Time { return now }

	draft, err := service.CreateDraft(context.Background(), CreateDraftInput{
		UserID: 42,
		Input:  "企业培训服务，每月访问量 12000，线索转化率 8%，成交率 15%，客单价 6000 元。",
	})

	if err != nil {
		t.Fatalf("CreateDraft() error = %v", err)
	}
	if draft.ID != 71 || draft.Status != DraftStatusNeedsInput {
		t.Fatalf("draft = %+v", draft)
	}
	if draft.Assumptions.MonthlyVisits != 12000 || draft.Assumptions.LeadRate != 0.08 || draft.Assumptions.DealRate != 0.15 || draft.Assumptions.AverageOrder != 6000 {
		t.Fatalf("assumptions = %+v", draft.Assumptions)
	}
	if len(draft.Questions) != 2 || draft.Questions[0].Key != "acquisition_cost" || draft.Questions[1].Key != "delivery_cost" {
		t.Fatalf("questions = %+v", draft.Questions)
	}
}

func TestServiceAnswersDraftThenCalculatesOwnedModel(t *testing.T) {
	repository := &draftRepository{draft: Draft{
		ID:     71,
		UserID: 42,
		Input:  "企业培训增长测算",
		Status: DraftStatusNeedsInput,
		Assumptions: Assumptions{
			MonthlyVisits: 12000,
			LeadRate:      0.08,
			DealRate:      0.15,
			AverageOrder:  6000,
		},
	}}
	service := NewService(repository)

	draft, err := service.AnswerDraft(context.Background(), AnswerDraftInput{
		UserID:  42,
		DraftID: 71,
		Answers: map[string]float64{"acquisition_cost": 80, "delivery_cost": 120000},
	})
	if err != nil {
		t.Fatalf("AnswerDraft() error = %v", err)
	}
	if draft.Status != DraftStatusReady || len(draft.Questions) != 0 {
		t.Fatalf("draft = %+v", draft)
	}

	calculation, err := service.CalculateDraft(context.Background(), CalculateDraftInput{UserID: 42, DraftID: 71})
	if err != nil {
		t.Fatalf("CalculateDraft() error = %v", err)
	}
	if calculation.Draft.Status != DraftStatusCalculated || calculation.Model.ID != 99 {
		t.Fatalf("calculation = %+v", calculation)
	}
	if calculation.Model.Name != "企业培训增长测算" || calculation.Model.Result.MonthlyRevenue == 0 {
		t.Fatalf("model = %+v", calculation.Model)
	}
}

func TestServiceAcceptsExplicitZeroCostAnswers(t *testing.T) {
	repository := &draftRepository{draft: Draft{
		ID: 71, UserID: 42, Status: DraftStatusNeedsInput,
		Assumptions: Assumptions{MonthlyVisits: 12000, LeadRate: 0.08, DealRate: 0.15, AverageOrder: 6000},
	}}
	service := NewService(repository)

	draft, err := service.AnswerDraft(context.Background(), AnswerDraftInput{
		UserID: 42, DraftID: 71,
		Answers: map[string]float64{"acquisition_cost": 0, "delivery_cost": 0},
	})
	if err != nil {
		t.Fatalf("AnswerDraft() error = %v", err)
	}
	if draft.Status != DraftStatusReady || len(draft.Questions) != 0 {
		t.Fatalf("draft = %+v", draft)
	}
}
