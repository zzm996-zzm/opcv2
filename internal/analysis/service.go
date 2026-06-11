package analysis

import (
	"context"
	"errors"
	"strings"
	"time"
)

const (
	ModeDirection    = "direction"
	StatusNeedsInput = "needs_input"
	StatusCompleted  = "completed"
)

var ErrServiceNotReady = errors.New("analysis service is not configured")

type DirectionInput struct {
	UserID  int64    `json:"-"`
	Intent  string   `json:"intent"`
	Answers []Answer `json:"answers,omitempty"`
}

type Answer struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Question struct {
	Key     string   `json:"key"`
	Text    string   `json:"text"`
	Options []string `json:"options"`
}

type DirectionResult struct {
	SessionID int64           `json:"session_id"`
	Status    string          `json:"status"`
	Questions []Question      `json:"questions,omitempty"`
	Cards     []DirectionCard `json:"cards,omitempty"`
}

type DirectionCard struct {
	Name           string     `json:"name"`
	Score          int        `json:"score"`
	Reasons        []string   `json:"reasons"`
	MarketEvidence string     `json:"market_evidence"`
	Difficulty     Difficulty `json:"difficulty"`
	Benchmarks     []string   `json:"benchmarks"`
	Actions        []string   `json:"actions"`
	Upsell         string     `json:"upsell"`
}

type Difficulty struct {
	Level string   `json:"level"`
	Notes []string `json:"notes"`
}

type Session struct {
	ID        int64           `json:"id"`
	UserID    int64           `json:"user_id"`
	Mode      string          `json:"mode"`
	Intent    string          `json:"intent"`
	Status    string          `json:"status"`
	Questions []Question      `json:"questions,omitempty"`
	Result    DirectionResult `json:"result,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
}

type Repository interface {
	CreateSession(ctx context.Context, session Session) (Session, error)
}

type Provider interface {
	GenerateDirection(ctx context.Context, input DirectionInput) (DirectionResult, error)
}

type Service struct {
	repository Repository
	provider   Provider
	now        func() time.Time
}

func NewService(repository Repository, provider Provider) *Service {
	return &Service{repository: repository, provider: provider, now: time.Now}
}

func (s *Service) StartDirection(ctx context.Context, input DirectionInput) (DirectionResult, error) {
	input.Intent = strings.TrimSpace(input.Intent)
	if s.repository == nil || s.provider == nil {
		return DirectionResult{}, ErrServiceNotReady
	}

	questions := missingQuestions(input)
	if len(questions) > 0 {
		session, err := s.repository.CreateSession(ctx, Session{
			UserID:    input.UserID,
			Mode:      ModeDirection,
			Intent:    input.Intent,
			Status:    StatusNeedsInput,
			Questions: questions,
			CreatedAt: s.now(),
		})
		if err != nil {
			return DirectionResult{}, err
		}
		return DirectionResult{SessionID: session.ID, Status: StatusNeedsInput, Questions: questions}, nil
	}

	result, err := s.provider.GenerateDirection(ctx, input)
	if err != nil {
		return DirectionResult{}, err
	}
	result.Status = StatusCompleted
	session, err := s.repository.CreateSession(ctx, Session{
		UserID:    input.UserID,
		Mode:      ModeDirection,
		Intent:    input.Intent,
		Status:    StatusCompleted,
		Result:    result,
		CreatedAt: s.now(),
	})
	if err != nil {
		return DirectionResult{}, err
	}
	result.SessionID = session.ID
	return result, nil
}

func missingQuestions(input DirectionInput) []Question {
	intent := input.Intent + " " + answerText(input.Answers)
	var questions []Question
	if len([]rune(strings.TrimSpace(input.Intent))) < 30 {
		questions = append(questions, Question{
			Key:     "background",
			Text:    "你现在最有把握的经验或资源是什么？",
			Options: []string{"行业经验", "客户资源", "可投入资金", "可投入时间"},
		})
	}
	if !containsAny(intent, "万", "预算", "本金", "资金") {
		questions = append(questions, Question{
			Key:     "budget",
			Text:    "你大概有多少启动资金？",
			Options: []string{"1万以内", "1-5万", "5-20万", "20万以上"},
		})
	}
	if !containsAny(intent, "小时", "全职", "兼职", "每周") {
		questions = append(questions, Question{
			Key:     "time",
			Text:    "你每周能投入多少时间？",
			Options: []string{"5小时以内", "5-20小时", "全职"},
		})
	}
	if len(questions) > 3 {
		return questions[:3]
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
