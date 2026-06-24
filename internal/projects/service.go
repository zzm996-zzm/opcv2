package projects

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/zzm/opcv2/internal/ai"
)

type Repository interface {
	CreateSession(ctx context.Context, session MatchSession) (MatchSession, error)
	ListSessions(ctx context.Context, userID int64, limit int) ([]MatchSession, error)
	GetSession(ctx context.Context, userID, id int64) (MatchSession, error)
	SaveFavorite(ctx context.Context, favorite Favorite) (Favorite, error)
}

type JSONGenerator interface {
	GenerateJSON(ctx context.Context, request ai.GenerateJSONRequest) (ai.GenerateJSONResult, error)
}

type Service struct {
	repository Repository
	generator  JSONGenerator
	now        func() time.Time
}

func NewService(repository Repository, generator JSONGenerator) *Service {
	return &Service{repository: repository, generator: generator, now: time.Now}
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

func (s *Service) generateMatch(ctx context.Context, input MatchInput) (MatchResult, error) {
	aiResult, err := s.generator.GenerateJSON(ctx, ai.GenerateJSONRequest{
		UserID:         input.UserID,
		Feature:        "projects.match",
		PromptVersion:  "project_match_v1",
		SystemPrompt:   "你是项目超市 AI 匹配助手。必须只返回 JSON，字段严格匹配 project_match_result。",
		UserPrompt:     matchUserPrompt(input),
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
	result.Status = StatusCompleted
	return result, nil
}

func matchUserPrompt(input MatchInput) string {
	if len(input.Answers) == 0 {
		return input.Intent
	}
	return input.Intent + "\n补充回答：" + answerText(input.Answers)
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
			Options: []string{"内容创作", "客户资源", "行业经验", "技术能力"},
		})
	}
	if !containsAny(intent, "万", "预算", "资金", "本金") {
		questions = append(questions, Question{
			Key:     "budget",
			Text:    "你计划投入多少启动预算？",
			Options: []string{"1万以内", "1-3万", "3-10万", "10万以上"},
		})
	}
	if !containsAny(intent, "小时", "全职", "兼职", "每周", "每天") {
		questions = append(questions, Question{
			Key:     "time",
			Text:    "你每周可以投入多少时间？",
			Options: []string{"5小时以内", "5-20小时", "20小时以上", "全职"},
		})
	}
	if !containsAny(intent, "线上", "本地", "服务", "产品", "一人公司", "轻资产") {
		questions = append(questions, Question{
			Key:     "preference",
			Text:    "你更偏好哪类项目形态？",
			Options: []string{"线上轻资产", "本地服务", "产品工具", "都可以"},
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
