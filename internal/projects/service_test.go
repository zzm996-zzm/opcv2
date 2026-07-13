package projects

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/zzm/opcv2/internal/account"
	"github.com/zzm/opcv2/internal/ai"
)

type memoryRepository struct {
	sessions      []MatchSession
	favorites     map[int64]map[int64]Favorite
	opportunities []Opportunity
	cases         []CaseStudy
	comparisons   []Comparison
	exports       []Export
	nextID        int64
}

func (r *memoryRepository) CreateExport(_ context.Context, item Export) (Export, error) {
	item.ID = int64(len(r.exports) + 1)
	r.exports = append(r.exports, item)
	return item, nil
}
func (r *memoryRepository) GetExport(_ context.Context, userID, id int64) (Export, error) {
	for _, item := range r.exports {
		if item.UserID == userID && item.ID == id {
			return item, nil
		}
	}
	return Export{}, ErrExportNotFound
}

func (r *memoryRepository) CreateComparison(_ context.Context, comparison Comparison) (Comparison, error) {
	comparison.ID = int64(len(r.comparisons) + 1)
	r.comparisons = append(r.comparisons, comparison)
	return comparison, nil
}

func (r *memoryRepository) GetComparison(_ context.Context, userID, id int64) (Comparison, error) {
	for _, item := range r.comparisons {
		if item.UserID == userID && item.ID == id {
			return item, nil
		}
	}
	return Comparison{}, ErrComparisonNotFound
}

func (r *memoryRepository) ListCases(_ context.Context, filters CaseFilters) ([]CaseStudy, error) {
	var rows []CaseStudy
	for _, item := range r.cases {
		if item.Status == CaseStatusPublished && (filters.CaseType == "" || item.CaseType == filters.CaseType) {
			rows = append(rows, item)
		}
	}
	return rows, nil
}

func (r *memoryRepository) GetCase(_ context.Context, slug string) (CaseStudy, error) {
	for _, item := range r.cases {
		if item.Status == CaseStatusPublished && item.Slug == slug {
			return item, nil
		}
	}
	return CaseStudy{}, ErrCaseNotFound
}

func (r *memoryRepository) ListOpportunities(_ context.Context, filters OpportunityFilters) ([]Opportunity, error) {
	var rows []Opportunity
	for _, item := range r.opportunities {
		if item.Status == OpportunityStatusPublished && (filters.Query == "" || strings.Contains(item.Title, filters.Query)) {
			rows = append(rows, item)
		}
	}
	return rows, nil
}

func (r *memoryRepository) GetOpportunity(_ context.Context, slug string) (Opportunity, error) {
	for _, item := range r.opportunities {
		if item.Status == OpportunityStatusPublished && item.Slug == slug {
			return item, nil
		}
	}
	return Opportunity{}, ErrOpportunityNotFound
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

func (r *memoryRepository) UpdateSession(_ context.Context, session MatchSession) (MatchSession, error) {
	for index, item := range r.sessions {
		if item.UserID == session.UserID && item.ID == session.ID {
			r.sessions[index] = session
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

type fakeProfileContextProvider struct {
	context account.ProfileContext
	err     error
}

func (p *fakeProfileContextProvider) GetProfileContext(context.Context, int64) (account.ProfileContext, error) {
	return p.context, p.err
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

func TestServiceAnswersFollowUpAndCompletesExistingSession(t *testing.T) {
	payload := MatchResult{Status: StatusCompleted, Projects: []ProjectMatch{{Rank: 1, Title: "AI销售顾问", Score: 90, Tags: []string{"B端"}, Budget: "1万", Reasons: []string{"经验匹配"}, Risk: "需验证获客"}}}
	content, _ := json.Marshal(payload)
	repository := &memoryRepository{sessions: []MatchSession{{ID: 99, UserID: 42, Intent: "想找项目", Status: StatusNeedsInput, Questions: []Question{{Key: "background", Text: "能力", Options: []string{"销售"}}}}}}
	service := NewService(repository, &fakeJSONGenerator{result: ai.GenerateJSONResult{Content: content}})

	result, err := service.AnswerMatch(context.Background(), AnswerMatchInput{UserID: 42, SessionID: 99, Answers: []Answer{{Key: "background", Value: "销售经验"}}})
	if err != nil || result.SessionID != 99 || repository.sessions[0].Status != StatusCompleted || len(repository.sessions[0].Result.Projects) != 1 {
		t.Fatalf("result/session = %+v/%+v err=%v", result, repository.sessions[0], err)
	}
}

func TestServiceAddsProfileContextToMatchPrompt(t *testing.T) {
	payload := MatchResult{
		Status: StatusCompleted,
		Projects: []ProjectMatch{{
			Rank:    1,
			Title:   "AI私域增长顾问",
			Score:   91,
			Tags:    []string{"私域", "企业服务"},
			Budget:  "¥10,000 - ¥30,000",
			Reasons: []string{"行业匹配", "目标清晰"},
			Risk:    "交付依赖案例积累",
		}},
	}
	content, _ := json.Marshal(payload)
	generator := &fakeJSONGenerator{result: ai.GenerateJSONResult{Content: content}}
	profile := &fakeProfileContextProvider{context: account.ProfileContext{
		UserID:    42,
		Completed: true,
		Groups: []account.ProfileGroup{
			{Key: account.ProfileGroupBusiness, Title: "我的业务/公司", Fields: map[string]string{"company": "智活AI", "stage": "启动"}},
			{Key: account.ProfileGroupGoals, Title: "目标与诉求", Fields: map[string]string{"short_term": "验证企业服务项目"}},
		},
	}}
	service := NewService(&memoryRepository{}, generator, WithProfileContextProvider(profile))

	_, err := service.CreateMatch(context.Background(), MatchInput{
		UserID: 42,
		Intent: "我擅长企业服务，预算3万以内，每周能投入20小时，希望做线上轻资产项目",
	})

	if err != nil {
		t.Fatalf("CreateMatch() error = %v", err)
	}
	if !strings.Contains(generator.request.UserPrompt, "用户画像上下文") ||
		!strings.Contains(generator.request.UserPrompt, "智活AI") ||
		!strings.Contains(generator.request.UserPrompt, "验证企业服务项目") {
		t.Fatalf("prompt missing profile context: %s", generator.request.UserPrompt)
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
