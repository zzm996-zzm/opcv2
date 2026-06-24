package analysis

import (
	"context"
	"errors"
	"testing"

	"github.com/zzm/opcv2/internal/ai"
)

type memoryRepository struct {
	session  Session
	sessions []Session
}

func (r *memoryRepository) CreateSession(_ context.Context, session Session) (Session, error) {
	session.ID = 99
	r.session = session
	r.sessions = append(r.sessions, session)
	return session, nil
}

func (r *memoryRepository) ListSessions(_ context.Context, userID int64, limit int) ([]Session, error) {
	var sessions []Session
	for _, session := range r.sessions {
		if session.UserID == userID {
			sessions = append(sessions, session)
		}
	}
	if limit > 0 && len(sessions) > limit {
		return sessions[:limit], nil
	}
	return sessions, nil
}

func (r *memoryRepository) GetSession(_ context.Context, userID, id int64) (Session, error) {
	for _, session := range r.sessions {
		if session.UserID == userID && session.ID == id {
			return session, nil
		}
	}
	return Session{}, ErrSessionNotFound
}

type fakeJSONGenerator struct {
	request ai.GenerateJSONRequest
	result  ai.GenerateJSONResult
	err     error
}

func (g *fakeJSONGenerator) GenerateJSON(_ context.Context, request ai.GenerateJSONRequest) (ai.GenerateJSONResult, error) {
	g.request = request
	return g.result, g.err
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

func TestServiceGeneratesDirectionCardsThroughAICore(t *testing.T) {
	repository := &memoryRepository{}
	generator := &fakeJSONGenerator{
		result: ai.GenerateJSONResult{Content: []byte(`{
			"cards": [
				{
					"name": "AI短视频脚本工作室",
					"score": 91,
					"reasons": ["能力匹配", "预算匹配", "市场需求"],
					"market_evidence": "短视频脚本需求持续存在。",
					"difficulty": {"level": "中", "notes": ["需要样板", "需要稳定交付"]},
					"benchmarks": ["脚本工作室", "本地商家代运营"],
					"actions": ["做3个样板", "验证报价", "联系10个客户"],
					"upsell": "升级后可生成线索名单。"
				}
			]
		}`)},
	}
	service := NewService(repository, generator)

	result, err := service.StartDirection(context.Background(), DirectionInput{
		UserID: 42,
		Intent: "我有10年教培经验，5万本金，目前在成都，每周能投入20小时，想做线上加线下的创业项目",
	})
	if err != nil {
		t.Fatalf("StartDirection() error = %v", err)
	}
	if generator.request.Feature != "analysis.direction" || generator.request.PromptVersion == "" {
		t.Fatalf("AI request = %+v", generator.request)
	}
	if generator.request.Validate == nil {
		t.Fatal("AI request Validate = nil")
	}
	if result.Status != StatusCompleted || len(result.Cards) != 1 || result.Cards[0].Name != "AI短视频脚本工作室" {
		t.Fatalf("result = %+v", result)
	}
}

func TestServiceRejectsInvalidDirectionCardPayload(t *testing.T) {
	generator := &fakeJSONGenerator{
		result: ai.GenerateJSONResult{Content: []byte(`{"cards":[{"name":"字段不完整"}]}`)},
	}
	service := NewService(&memoryRepository{}, generator)

	_, err := service.StartDirection(context.Background(), DirectionInput{
		UserID: 42,
		Intent: "我有10年教培经验，5万本金，目前在成都，每周能投入20小时，想做线上加线下的创业项目",
	})
	if !errors.Is(err, ErrInvalidAIResult) {
		t.Fatalf("StartDirection() error = %v, want ErrInvalidAIResult", err)
	}
}

func TestServiceReturnsSafeErrorWhenAICoreFails(t *testing.T) {
	generator := &fakeJSONGenerator{err: ai.ErrInvalidModelJSON}
	service := NewService(&memoryRepository{}, generator)

	_, err := service.StartDirection(context.Background(), DirectionInput{
		UserID: 42,
		Intent: "我有10年教培经验，5万本金，目前在成都，每周能投入20小时，想做线上加线下的创业项目",
	})
	if !errors.Is(err, ErrInvalidAIResult) {
		t.Fatalf("StartDirection() error = %v, want ErrInvalidAIResult", err)
	}
}

func TestServiceListsSessionsForUser(t *testing.T) {
	repository := &memoryRepository{sessions: []Session{
		{ID: 1, UserID: 42, Mode: ModeDirection, Intent: "我的项目", Status: StatusCompleted},
		{ID: 2, UserID: 7, Mode: ModeDirection, Intent: "别人的项目", Status: StatusCompleted},
		{ID: 3, UserID: 42, Mode: ModeDirection, Intent: "第二个项目", Status: StatusNeedsInput},
	}}
	service := NewService(repository, NewDevelopmentProvider())

	sessions, err := service.ListSessions(context.Background(), 42, 20)
	if err != nil {
		t.Fatalf("ListSessions() error = %v", err)
	}
	if len(sessions) != 2 {
		t.Fatalf("sessions = %+v, want 2 owned sessions", sessions)
	}
	for _, session := range sessions {
		if session.UserID != 42 {
			t.Fatalf("session leaked from another user: %+v", session)
		}
	}
}

func TestServiceGetsStoredSessionForUser(t *testing.T) {
	repository := &memoryRepository{sessions: []Session{
		{
			ID:     99,
			UserID: 42,
			Mode:   ModeDirection,
			Intent: "我有10年教培经验，5万本金，每周20小时",
			Status: StatusCompleted,
			Result: DirectionResult{
				Status: StatusCompleted,
				Cards:  []DirectionCard{{Name: "本地教培小班陪跑", Score: 91}},
			},
		},
	}}
	service := NewService(repository, NewDevelopmentProvider())

	session, err := service.GetSession(context.Background(), 42, 99)
	if err != nil {
		t.Fatalf("GetSession() error = %v", err)
	}
	if session.ID != 99 || session.Result.Cards[0].Name != "本地教培小班陪跑" {
		t.Fatalf("session = %+v", session)
	}
}

func TestServiceDoesNotReturnAnotherUsersSession(t *testing.T) {
	repository := &memoryRepository{sessions: []Session{
		{ID: 99, UserID: 7, Mode: ModeDirection, Intent: "别人的项目", Status: StatusCompleted},
	}}
	service := NewService(repository, NewDevelopmentProvider())

	_, err := service.GetSession(context.Background(), 42, 99)
	if !errors.Is(err, ErrSessionNotFound) {
		t.Fatalf("GetSession() error = %v, want ErrSessionNotFound", err)
	}
}
