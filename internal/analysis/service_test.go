package analysis

import (
	"context"
	"testing"
)

type memoryRepository struct {
	session Session
}

func (r *memoryRepository) CreateSession(_ context.Context, session Session) (Session, error) {
	session.ID = 99
	r.session = session
	return session, nil
}

func TestServiceAsksFollowUpWhenInputIsTooThin(t *testing.T) {
	service := NewService(&memoryRepository{}, NewDevelopmentProvider())

	result, err := service.StartDirection(context.Background(), DirectionInput{
		UserID: 42,
		Intent: "想创业",
	})
	if err != nil {
		t.Fatalf("StartDirection() error = %v", err)
	}
	if result.Status != StatusNeedsInput {
		t.Fatalf("status = %q, want %q", result.Status, StatusNeedsInput)
	}
	if len(result.Questions) == 0 || len(result.Questions) > 3 {
		t.Fatalf("questions = %+v", result.Questions)
	}
}

func TestServiceGeneratesThreeDirectionCards(t *testing.T) {
	repository := &memoryRepository{}
	service := NewService(repository, NewDevelopmentProvider())

	result, err := service.StartDirection(context.Background(), DirectionInput{
		UserID: 42,
		Intent: "我有10年教培经验，5万本金，目前在成都，每周能投入20小时，想做线上加线下的创业项目",
	})
	if err != nil {
		t.Fatalf("StartDirection() error = %v", err)
	}
	if result.Status != StatusCompleted {
		t.Fatalf("status = %q, want %q", result.Status, StatusCompleted)
	}
	if len(result.Cards) != 3 {
		t.Fatalf("cards = %d, want 3", len(result.Cards))
	}
	card := result.Cards[0]
	if card.Name == "" || card.Score == 0 || len(card.Reasons) != 3 ||
		card.MarketEvidence == "" || card.Difficulty.Level == "" ||
		len(card.Benchmarks) < 2 || len(card.Actions) != 3 || card.Upsell == "" {
		t.Fatalf("card missing required fields: %+v", card)
	}
	if repository.session.Status != StatusCompleted || repository.session.Result.Cards[0].Name == "" {
		t.Fatalf("stored session = %+v", repository.session)
	}
}
