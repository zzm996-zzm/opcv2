package projects

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/zzm/opcv2/internal/account"
	"github.com/zzm/opcv2/internal/ai"
)

type Repository interface {
	ListOpportunities(ctx context.Context, filters OpportunityFilters) ([]Opportunity, error)
	GetOpportunity(ctx context.Context, slug string) (Opportunity, error)
	ListCases(ctx context.Context, filters CaseFilters) ([]CaseStudy, error)
	GetCase(ctx context.Context, slug string) (CaseStudy, error)
	CreateComparison(ctx context.Context, comparison Comparison) (Comparison, error)
	GetComparison(ctx context.Context, userID, id int64) (Comparison, error)
	CreateExport(ctx context.Context, item Export) (Export, error)
	GetExport(ctx context.Context, userID, id int64) (Export, error)
	CreateSession(ctx context.Context, session MatchSession) (MatchSession, error)
	ListSessions(ctx context.Context, userID int64, limit int) ([]MatchSession, error)
	GetSession(ctx context.Context, userID, id int64) (MatchSession, error)
	UpdateSession(ctx context.Context, session MatchSession) (MatchSession, error)
	SaveFavorite(ctx context.Context, favorite Favorite) (Favorite, error)
	ListFavorites(ctx context.Context, userID int64, limit int) ([]Favorite, error)
	DeleteFavorite(ctx context.Context, userID, sessionID int64) error
}

func (s *Service) CreateExport(ctx context.Context, input CreateExportInput) (Export, error) {
	if s.repository == nil {
		return Export{}, ErrServiceNotReady
	}
	if input.SourceID <= 0 {
		return Export{}, ErrInvalidExport
	}
	var source any
	switch input.SourceType {
	case ExportSourceMatch:
		item, err := s.repository.GetSession(ctx, input.UserID, input.SourceID)
		if err != nil {
			return Export{}, err
		}
		source = item
	case ExportSourceComparison:
		item, err := s.repository.GetComparison(ctx, input.UserID, input.SourceID)
		if err != nil {
			return Export{}, err
		}
		source = item
	default:
		return Export{}, ErrInvalidExport
	}
	payload, err := json.Marshal(source)
	if err != nil {
		return Export{}, err
	}
	item, err := s.repository.CreateExport(ctx, Export{UserID: input.UserID, SourceType: input.SourceType, SourceID: input.SourceID, Status: "ready", Payload: payload, CreatedAt: s.now()})
	if err != nil {
		return Export{}, err
	}
	item.DownloadURL = fmt.Sprintf("/api/v1/projects/exports/%d/download", item.ID)
	return item, nil
}
func (s *Service) GetExport(ctx context.Context, userID, id int64) (Export, error) {
	if s.repository == nil {
		return Export{}, ErrServiceNotReady
	}
	return s.repository.GetExport(ctx, userID, id)
}

func (s *Service) CreateComparison(ctx context.Context, input CreateComparisonInput) (Comparison, error) {
	if s.repository == nil {
		return Comparison{}, ErrServiceNotReady
	}
	seen := map[string]bool{}
	slugs := make([]string, 0, len(input.OpportunitySlugs))
	for _, raw := range input.OpportunitySlugs {
		slug := strings.TrimSpace(raw)
		if slug != "" && !seen[slug] {
			seen[slug] = true
			slugs = append(slugs, slug)
		}
	}
	if len(slugs) < 2 || len(slugs) > 4 {
		return Comparison{}, ErrInvalidComparison
	}
	items := make([]Opportunity, 0, len(slugs))
	for _, slug := range slugs {
		item, err := s.repository.GetOpportunity(ctx, slug)
		if err != nil {
			return Comparison{}, err
		}
		items = append(items, item)
	}
	return s.repository.CreateComparison(ctx, Comparison{UserID: input.UserID, Items: items, CreatedAt: s.now()})
}
func (s *Service) GetComparison(ctx context.Context, userID, id int64) (Comparison, error) {
	if s.repository == nil {
		return Comparison{}, ErrServiceNotReady
	}
	return s.repository.GetComparison(ctx, userID, id)
}

func (s *Service) AnswerMatch(ctx context.Context, input AnswerMatchInput) (MatchResult, error) {
	if s.repository == nil || s.generator == nil {
		return MatchResult{}, ErrServiceNotReady
	}
	session, err := s.repository.GetSession(ctx, input.UserID, input.SessionID)
	if err != nil {
		return MatchResult{}, err
	}
	if session.Status != StatusNeedsInput || !answersCoverQuestions(session.Questions, input.Answers) {
		return MatchResult{}, ErrInvalidMatchAnswers
	}
	result, err := s.generateMatch(ctx, MatchInput{UserID: input.UserID, Intent: session.Intent, Answers: input.Answers})
	if err != nil {
		return MatchResult{}, err
	}
	result.SessionID = session.ID
	session.Status = StatusCompleted
	session.Answers = append([]Answer(nil), input.Answers...)
	session.Questions = []Question{}
	session.Result = result
	session.UpdatedAt = s.now()
	if _, err := s.repository.UpdateSession(ctx, session); err != nil {
		return MatchResult{}, err
	}
	return result, nil
}

func answersCoverQuestions(questions []Question, answers []Answer) bool {
	values := map[string]string{}
	for _, answer := range answers {
		values[strings.TrimSpace(answer.Key)] = strings.TrimSpace(answer.Value)
	}
	for _, question := range questions {
		if values[question.Key] == "" {
			return false
		}
	}
	return len(questions) > 0
}

func (s *Service) ListCases(ctx context.Context, filters CaseFilters) ([]CaseStudy, error) {
	if s.repository == nil {
		return nil, ErrServiceNotReady
	}
	filters.CaseType = strings.TrimSpace(filters.CaseType)
	if filters.Limit <= 0 || filters.Limit > 100 {
		filters.Limit = 20
	}
	return s.repository.ListCases(ctx, filters)
}

func (s *Service) GetCase(ctx context.Context, slug string) (CaseStudy, error) {
	if s.repository == nil {
		return CaseStudy{}, ErrServiceNotReady
	}
	slug = strings.TrimSpace(slug)
	if slug == "" {
		return CaseStudy{}, ErrCaseNotFound
	}
	return s.repository.GetCase(ctx, slug)
}

func (s *Service) ListOpportunities(ctx context.Context, filters OpportunityFilters) ([]Opportunity, error) {
	if s.repository == nil {
		return nil, ErrServiceNotReady
	}
	filters.Query = strings.TrimSpace(filters.Query)
	filters.Industry = strings.TrimSpace(filters.Industry)
	if filters.Limit <= 0 || filters.Limit > 100 {
		filters.Limit = 20
	}
	return s.repository.ListOpportunities(ctx, filters)
}

func (s *Service) GetOpportunity(ctx context.Context, slug string) (Opportunity, error) {
	if s.repository == nil {
		return Opportunity{}, ErrServiceNotReady
	}
	slug = strings.TrimSpace(slug)
	if slug == "" {
		return Opportunity{}, ErrOpportunityNotFound
	}
	return s.repository.GetOpportunity(ctx, slug)
}

type JSONGenerator interface {
	GenerateJSON(ctx context.Context, request ai.GenerateJSONRequest) (ai.GenerateJSONResult, error)
}

type ProfileContextProvider interface {
	GetProfileContext(ctx context.Context, userID int64) (account.ProfileContext, error)
}

type Option func(*Service)

type Service struct {
	repository            Repository
	generator             JSONGenerator
	matchQueue            ProjectMatchQueue
	profile               ProfileContextProvider
	featurePaywallEnabled bool
	now                   func() time.Time
}

func WithProjectMatchQueue(queue ProjectMatchQueue) Option {
	return func(service *Service) {
		service.matchQueue = queue
	}
}

func NewService(repository Repository, generator JSONGenerator, options ...Option) *Service {
	service := &Service{repository: repository, generator: generator, now: time.Now}
	for _, option := range options {
		option(service)
	}
	return service
}

func WithProfileContextProvider(provider ProfileContextProvider) Option {
	return func(service *Service) {
		service.profile = provider
	}
}

func WithFeaturePaywallEnabled(enabled bool) Option {
	return func(service *Service) {
		service.featurePaywallEnabled = enabled
	}
}

func (s *Service) CreateMatch(ctx context.Context, input MatchInput) (MatchResult, error) {
	input.Intent = strings.TrimSpace(input.Intent)
	if s.repository == nil || s.generator == nil {
		return MatchResult{}, ErrServiceNotReady
	}
	questions := missingQuestions(input)
	if len(questions) > 0 {
		session, err := s.repository.CreateSession(ctx, MatchSession{
			UserID:    input.UserID,
			Intent:    input.Intent,
			Status:    StatusNeedsInput,
			Questions: questions,
			CreatedAt: s.now(),
		})
		if err != nil {
			return MatchResult{}, err
		}
		return MatchResult{SessionID: session.ID, Status: StatusNeedsInput, Questions: questions}, nil
	}

	result, err := s.generateMatch(ctx, input)
	if err != nil {
		return MatchResult{}, err
	}
	session, err := s.repository.CreateSession(ctx, MatchSession{
		UserID:    input.UserID,
		Intent:    input.Intent,
		Status:    StatusCompleted,
		Result:    result,
		CreatedAt: s.now(),
	})
	if err != nil {
		return MatchResult{}, err
	}
	result.SessionID = session.ID
	session.Result = result
	session.UpdatedAt = s.now()
	if _, err := s.repository.UpdateSession(ctx, session); err != nil {
		return MatchResult{}, err
	}
	return result, nil
}

func (s *Service) ListMatches(ctx context.Context, userID int64, limit int) ([]MatchSession, error) {
	if s.repository == nil {
		return nil, ErrServiceNotReady
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return s.repository.ListSessions(ctx, userID, limit)
}

func (s *Service) GetMatch(ctx context.Context, userID, id int64) (MatchSession, error) {
	if s.repository == nil {
		return MatchSession{}, ErrServiceNotReady
	}
	return s.repository.GetSession(ctx, userID, id)
}

func (s *Service) FavoriteMatch(ctx context.Context, userID, id int64) (Favorite, error) {
	if s.repository == nil {
		return Favorite{}, ErrServiceNotReady
	}
	if _, err := s.repository.GetSession(ctx, userID, id); err != nil {
		return Favorite{}, err
	}
	return s.repository.SaveFavorite(ctx, Favorite{UserID: userID, SessionID: id, CreatedAt: s.now()})
}

func (s *Service) ListFavoriteMatches(ctx context.Context, userID int64, limit int) ([]Favorite, error) {
	if s.repository == nil {
		return nil, ErrServiceNotReady
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return s.repository.ListFavorites(ctx, userID, limit)
}

func (s *Service) UnfavoriteMatch(ctx context.Context, userID, id int64) error {
	if s.repository == nil {
		return ErrServiceNotReady
	}
	if _, err := s.repository.GetSession(ctx, userID, id); err != nil {
		return err
	}
	return s.repository.DeleteFavorite(ctx, userID, id)
}

func (s *Service) generateMatch(ctx context.Context, input MatchInput) (MatchResult, error) {
	profilePrompt, err := s.profilePrompt(ctx, input.UserID)
	if err != nil {
		return MatchResult{}, err
	}
	catalog, err := s.repository.ListOpportunities(ctx, OpportunityFilters{Limit: 100})
	if err != nil {
		return MatchResult{}, err
	}
	userPrompt := appendPromptSection(matchUserPrompt(input), profilePrompt)
	userPrompt = appendPromptSection(userPrompt, matchCatalogPrompt(catalog))
	aiResult, err := s.generator.GenerateJSON(ctx, ai.GenerateJSONRequest{
		UserID:         input.UserID,
		Feature:        "projects.match",
		PromptVersion:  "project_match_v2",
		SystemPrompt:   "你是项目超市 AI 匹配助手。必须只从提供的已发布项目目录中推荐，返回 opportunity_slug，并且只返回字段严格匹配 project_match_result 的 JSON。",
		UserPrompt:     userPrompt,
		SchemaName:     "project_match_result",
		Validate:       validateMatchResultJSON,
		RepairAttempts: 1,
	})
	if err != nil {
		return MatchResult{}, fmt.Errorf("%w: %v", ErrInvalidAIResult, err)
	}
	var result MatchResult
	if err := json.Unmarshal(aiResult.Content, &result); err != nil {
		return MatchResult{}, fmt.Errorf("%w: %v", ErrInvalidAIResult, err)
	}
	if err := validateMatchResult(result); err != nil {
		return MatchResult{}, fmt.Errorf("%w: %v", ErrInvalidAIResult, err)
	}
	if err := bindCatalogMatches(&result, catalog); err != nil {
		return MatchResult{}, fmt.Errorf("%w: %v", ErrInvalidAIResult, err)
	}
	result.Status = StatusCompleted
	return result, nil
}

func matchCatalogPrompt(catalog []Opportunity) string {
	if len(catalog) == 0 {
		return ""
	}
	type catalogItem struct {
		Slug                 string   `json:"opportunity_slug"`
		Title                string   `json:"title"`
		Industry             string   `json:"industry"`
		Tags                 []string `json:"tags"`
		BudgetBand           string   `json:"budget_band"`
		Difficulty           string   `json:"difficulty"`
		ResourceRequirements []string `json:"resource_requirements"`
	}
	items := make([]catalogItem, 0, len(catalog))
	for _, item := range catalog {
		items = append(items, catalogItem{
			Slug: item.Slug, Title: item.Title, Industry: item.Industry, Tags: item.Tags,
			BudgetBand: item.BudgetBand, Difficulty: item.Difficulty, ResourceRequirements: item.ResourceRequirements,
		})
	}
	data, err := json.Marshal(items)
	if err != nil {
		return ""
	}
	return "已发布项目目录（只能从中选择，opportunity_slug 必须原样返回）：\n" + string(data)
}

func bindCatalogMatches(result *MatchResult, catalog []Opportunity) error {
	if len(catalog) == 0 {
		return nil
	}
	bySlug := make(map[string]Opportunity, len(catalog))
	byTitle := make(map[string]Opportunity, len(catalog))
	for _, item := range catalog {
		bySlug[item.Slug] = item
		byTitle[item.Title] = item
	}
	for index := range result.Projects {
		project := &result.Projects[index]
		item, ok := bySlug[strings.TrimSpace(project.OpportunitySlug)]
		if !ok {
			item, ok = byTitle[strings.TrimSpace(project.Title)]
		}
		if !ok {
			return fmt.Errorf("project %q is not in the published catalog", project.Title)
		}
		project.OpportunitySlug = item.Slug
		project.Title = item.Title
	}
	return nil
}

func (s *Service) profilePrompt(ctx context.Context, userID int64) (string, error) {
	if s.profile == nil {
		return "", nil
	}
	profile, err := s.profile.GetProfileContext(ctx, userID)
	if err != nil {
		return "", err
	}
	return formatProfileContext(profile), nil
}

func matchUserPrompt(input MatchInput) string {
	if len(input.Answers) == 0 {
		return input.Intent
	}
	return input.Intent + "\n补充回答：" + answerText(input.Answers)
}

func appendPromptSection(base string, section string) string {
	section = strings.TrimSpace(section)
	if section == "" {
		return base
	}
	return base + "\n\n" + section
}

func formatProfileContext(profile account.ProfileContext) string {
	if len(profile.Groups) == 0 {
		return ""
	}
	lines := []string{"用户画像上下文："}
	for _, group := range profile.Groups {
		if len(group.Fields) == 0 {
			continue
		}
		fields := make([]string, 0, len(group.Fields))
		for key, value := range group.Fields {
			value = strings.TrimSpace(value)
			if value == "" {
				continue
			}
			fields = append(fields, key+"="+value)
		}
		if len(fields) == 0 {
			continue
		}
		sort.Strings(fields)
		lines = append(lines, "- "+group.Title+"："+strings.Join(fields, "；"))
	}
	if len(lines) == 1 {
		return ""
	}
	return strings.Join(lines, "\n")
}

func validateMatchResultJSON(data []byte) error {
	var result MatchResult
	if err := json.Unmarshal(data, &result); err != nil {
		return err
	}
	return validateMatchResult(result)
}

func validateMatchResult(result MatchResult) error {
	if len(result.Projects) == 0 {
		return errors.New("projects are required")
	}
	for _, project := range result.Projects {
		if project.Rank == 0 ||
			project.Title == "" ||
			project.Score == 0 ||
			len(project.Tags) == 0 ||
			project.Budget == "" ||
			len(project.Reasons) == 0 ||
			project.Risk == "" {
			return errors.New("project match is missing required fields")
		}
	}
	return nil
}

func missingQuestions(input MatchInput) []Question {
	intent := input.Intent + " " + answerText(input.Answers)
	var questions []Question
	if len([]rune(input.Intent)) < 30 {
		questions = append(questions, Question{
			Key:     "background",
			Text:    "你现在最明确的能力、经验或资源是什么？",
			Options: []string{"内容创作", "客户或行业资源", "技术能力"},
		})
	}
	if !containsAny(intent, "万", "预算", "资金", "本金") {
		questions = append(questions, Question{
			Key:     "budget",
			Text:    "你计划投入多少启动预算？",
			Options: []string{"1万以内", "1-3万", "3万以上"},
		})
	}
	if !containsAny(intent, "小时", "全职", "兼职", "每周", "每天") {
		questions = append(questions, Question{
			Key:     "time",
			Text:    "你每周可以投入多少时间？",
			Options: []string{"5小时以内", "5-20小时", "20小时以上"},
		})
	}
	if !containsAny(intent, "线上", "本地", "服务", "产品", "一人公司", "轻资产") {
		questions = append(questions, Question{
			Key:     "preference",
			Text:    "你更偏好哪类项目形态？",
			Options: []string{"线上轻资产", "本地服务", "都可以"},
		})
	}
	if len(questions) > 4 {
		return questions[:4]
	}
	return questions
}

func answerText(answers []Answer) string {
	parts := make([]string, 0, len(answers))
	for _, answer := range answers {
		parts = append(parts, answer.Value)
	}
	return strings.Join(parts, " ")
}

func containsAny(value string, needles ...string) bool {
	for _, needle := range needles {
		if strings.Contains(value, needle) {
			return true
		}
	}
	return false
}
