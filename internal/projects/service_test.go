package projects

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/zzm/opcv2/internal/ai"
)

type memoryRepository struct {
	sessions  []MatchSession
	favorites map[int64]map[int64]Favorite
	nextID    int64
}

func (r *memoryRepository) CreateSession(_ context.Context, session MatchSession) (MatchSession, error) {
	if r.nextID == 0 {
		r.nextID = 100
	}
	session.ID = r.nextID
	r.nextID++
	r.sessions = append(r.sessions, session)
	return session, nil
}

func (r *memoryRepository) ListSessions(_ context.Context, userID int64, limit int) ([]MatchSession, error) {
	var sessions []MatchSession
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

func (r *memoryRepository) GetSession(_ context.Context, userID, id int64) (MatchSession, error) {
	for _, session := range r.sessions {
		if session.UserID == userID && session.ID == id {
			return session, nil
		}
	}
	return MatchSession{}, ErrSessionNotFound
}

func (r *memoryRepository) SaveFavorite(_ context.Context, favorite Favorite) (Favorite, error) {
	if r.favorites == nil {
		r.favorites = map[int64]map[int64]Favorite{}
	}
	if r.favorites[favorite.UserID] == nil {
		r.favorites[favorite.UserID] = map[int64]Favorite{}
	}
	if existing, ok := r.favorites[favorite.UserID][favorite.SessionID]; ok {
		return existing, nil
	}
	favorite.ID = int64(len(r.favorites[favorite.UserID]) + 1)
	r.favorites[favorite.UserID][favorite.SessionID] = favorite
	return favorite, nil
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

func TestServiceAsksFollowUpForThinMatchRequest(t *testing.T) {
	repository := &memoryRepository{}
	service := NewService(repository, &fakeJSONGenerator{})

	result, err := service.CreateMatch(context.Background(), MatchInput{
		UserID: 42,
		Intent: "想找项目",
	})
	if err != nil {
		t.Fatalf("CreateMatch() error = %v", err)
	}
	if result.Status != StatusNeedsInput {
		t.Fatalf("status = %q, want %q", result.Status, StatusNeedsInput)
	}
	if len(result.Questions) == 0 || len(result.Questions) > 4 {
		t.Fatalf("questions = %+v", result.Questions)
	}
	if repository.sessions[0].UserID != 42 || repository.sessions[0].Status != StatusNeedsInput {
		t.Fatalf("stored session = %+v", repository.sessions[0])
	}
}

func TestServiceCreatesRankedProjectMatchesThroughAI(t *testing.T) {
	payload := MatchResult{
		Status: StatusCompleted,
		Projects: []ProjectMatch{
			{
				Rank:    1,
				Title:   "AI短视频脚本工作室",
				Score:   94,
				Tags:    []string{"内容创作", "低成本启动"},
				Budget:  "¥1,000 - ¥3,000",
				Reasons: []string{"能力匹配", "启动成本低", "市场需求明确"},
				Risk:    "同质化竞争较多",
			},
		},
	}
	content, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	repository := &memoryRepository{}
	generator := &fakeJSONGenerator{result: ai.GenerateJSONResult{Content: content}}
	service := NewService(repository, generator)

	result, err := service.CreateMatch(context.Background(), MatchInput{
		UserID: 42,
		Intent: "我擅长内容创作，预算3万以内，每周能投入20小时，希望做一人公司线上项目",
	})
	if err != nil {
		t.Fatalf("CreateMatch() error = %v", err)
	}
	if generator.request.Feature != "projects.match" || generator.request.PromptVersion != "project_match_v1" {
		t.Fatalf("AI request = %+v", generator.request)
	}
	if result.Status != StatusCompleted || len(result.Projects) != 1 || result.Projects[0].Title != "AI短视频脚本工作室" {
		t.Fatalf("result = %+v", result)
	}
	if repository.sessions[0].Result.Projects[0].Score != 94 {
		t.Fatalf("stored session = %+v", repository.sessions[0])
	}
}

func TestServiceReturnsSafeErrorForInvalidAIMatchPayload(t *testing.T) {
	service := NewService(&memoryRepository{}, &fakeJSONGenerator{
		result: ai.GenerateJSONResult{Content: []byte(`{"projects":[{"title":"字段不完整"}]}`)},
	})

	_, err := service.CreateMatch(context.Background(), MatchInput{
		UserID: 42,
		Intent: "我擅长内容创作，预算3万以内，每周能投入20小时，希望做一人公司线上项目",
	})
	if !errors.Is(err, ErrInvalidAIResult) {
		t.Fatalf("CreateMatch() error = %v, want ErrInvalidAIResult", err)
	}
}

func TestServiceListsOnlyUserMatchHistory(t *testing.T) {
	repository := &memoryRepository{sessions: []MatchSession{
		{ID: 1, UserID: 42, Intent: "我的项目", Status: StatusCompleted},
		{ID: 2, UserID: 7, Intent: "别人的项目", Status: StatusCompleted},
	}}
	service := NewService(repository, &fakeJSONGenerator{})

	sessions, err := service.ListMatches(context.Background(), 42, 20)
	if err != nil {
		t.Fatalf("ListMatches() error = %v", err)
	}
	if len(sessions) != 1 || sessions[0].UserID != 42 {
		t.Fatalf("sessions = %+v", sessions)
	}
}

func TestServiceFavoritesAreIdempotent(t *testing.T) {
	now := time.Date(2026, 6, 24, 9, 0, 0, 0, time.UTC)
	repository := &memoryRepository{sessions: []MatchSession{
		{ID: 99, UserID: 42, Intent: "我的项目", Status: StatusCompleted},
	}}
	service := NewService(repository, &fakeJSONGenerator{})
	service.now = func() time.Time { return now }

	first, err := service.FavoriteMatch(context.Background(), 42, 99)
	if err != nil {
		t.Fatalf("FavoriteMatch() first error = %v", err)
	}
	second, err := service.FavoriteMatch(context.Background(), 42, 99)
	if err != nil {
		t.Fatalf("FavoriteMatch() second error = %v", err)
	}
	if first.ID != second.ID || len(repository.favorites[42]) != 1 {
		t.Fatalf("favorites first=%+v second=%+v repo=%+v", first, second, repository.favorites)
	}
}
