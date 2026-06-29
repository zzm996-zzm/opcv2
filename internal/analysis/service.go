package analysis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/zzm/opcv2/internal/ai"
)

const (
	ModeDirection    = "direction"
	StatusNeedsInput = "needs_input"
	StatusCompleted  = "completed"
)

var (
	ErrServiceNotReady    = errors.New("analysis service is not configured")
	ErrInvalidAIResult    = errors.New("invalid analysis ai result")
	ErrSessionNotFound    = errors.New("analysis session not found")
	ErrActionItemNotFound = errors.New("analysis action item not found")
)

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
	UpdatedAt time.Time       `json:"updated_at"`
}

type ActionItem struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	SessionID int64     `json:"session_id"`
	DayIndex  int       `json:"day_index"`
	Title     string    `json:"title"`
	Detail    string    `json:"detail"`
	Completed bool      `json:"completed"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Repository interface {
	CreateSession(ctx context.Context, session Session) (Session, error)
	ListSessions(ctx context.Context, userID int64, limit int) ([]Session, error)
	GetSession(ctx context.Context, userID, id int64) (Session, error)
	ListActionItems(ctx context.Context, userID, sessionID int64) ([]ActionItem, error)
	CreateActionItems(ctx context.Context, items []ActionItem) ([]ActionItem, error)
	UpdateActionItem(ctx context.Context, userID, sessionID, itemID int64, completed bool) (ActionItem, error)
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

func (s *Service) ListSessions(ctx context.Context, userID int64, limit int) ([]Session, error) {
	if s.repository == nil {
		return nil, ErrServiceNotReady
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return s.repository.ListSessions(ctx, userID, limit)
}

func (s *Service) GetSession(ctx context.Context, userID, id int64) (Session, error) {
	if s.repository == nil {
		return Session{}, ErrServiceNotReady
	}
	return s.repository.GetSession(ctx, userID, id)
}

func (s *Service) ListActionItems(ctx context.Context, userID, sessionID int64) ([]ActionItem, error) {
	if s.repository == nil {
		return nil, ErrServiceNotReady
	}
	session, err := s.repository.GetSession(ctx, userID, sessionID)
	if err != nil {
		return nil, err
	}
	items, err := s.repository.ListActionItems(ctx, userID, session.ID)
	if err != nil {
		return nil, err
	}
	if len(items) > 0 || session.Status != StatusCompleted {
		return items, nil
	}
	return s.repository.CreateActionItems(ctx, defaultActionItems(userID, session.ID, s.now()))
}

func (s *Service) UpdateActionItem(ctx context.Context, userID, sessionID, itemID int64, completed bool) (ActionItem, error) {
	if s.repository == nil {
		return ActionItem{}, ErrServiceNotReady
	}
	if _, err := s.repository.GetSession(ctx, userID, sessionID); err != nil {
		return ActionItem{}, err
	}
	return s.repository.UpdateActionItem(ctx, userID, sessionID, itemID, completed)
}

func (s *Service) StartDirection(ctx context.Context, input DirectionInput) (DirectionResult, error) {
	input.Intent = strings.TrimSpace(input.Intent)
	if s.repository == nil || s.generator == nil {
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

	result, err := s.generateDirection(ctx, input)
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

func (s *Service) generateDirection(ctx context.Context, input DirectionInput) (DirectionResult, error) {
	aiResult, err := s.generator.GenerateJSON(ctx, ai.GenerateJSONRequest{
		UserID:         input.UserID,
		Feature:        "analysis.direction",
		PromptVersion:  "analysis_direction_v1",
		SystemPrompt:   directionSystemPrompt(),
		UserPrompt:     directionUserPrompt(input),
		SchemaName:     "analysis_direction_result",
		Validate:       validateDirectionResultJSON,
		RepairAttempts: 1,
	})
	if err != nil {
		return DirectionResult{}, fmt.Errorf("%w: %v", ErrInvalidAIResult, err)
	}
	var result DirectionResult
	if err := json.Unmarshal(aiResult.Content, &result); err != nil {
		return DirectionResult{}, fmt.Errorf("%w: %v", ErrInvalidAIResult, err)
	}
	if err := validateDirectionResult(result); err != nil {
		return DirectionResult{}, fmt.Errorf("%w: %v", ErrInvalidAIResult, err)
	}
	result.Status = StatusCompleted
	return result, nil
}

func directionSystemPrompt() string {
	return "你是创业方向分析助手。必须只返回 JSON，字段严格匹配 analysis_direction_result。"
}

func directionUserPrompt(input DirectionInput) string {
	if len(input.Answers) == 0 {
		return input.Intent
	}
	return input.Intent + "\n补充回答：" + answerText(input.Answers)
}

func validateDirectionResultJSON(data []byte) error {
	var result DirectionResult
	if err := json.Unmarshal(data, &result); err != nil {
		return err
	}
	return validateDirectionResult(result)
}

func validateDirectionResult(result DirectionResult) error {
	if len(result.Cards) == 0 {
		return errors.New("cards are required")
	}
	for _, card := range result.Cards {
		if card.Name == "" ||
			card.Score == 0 ||
			len(card.Reasons) == 0 ||
			card.MarketEvidence == "" ||
			card.Difficulty.Level == "" ||
			len(card.Difficulty.Notes) == 0 ||
			len(card.Benchmarks) == 0 ||
			len(card.Actions) == 0 ||
			card.Upsell == "" {
			return errors.New("direction card is missing required fields")
		}
	}
	return nil
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

func defaultActionItems(userID, sessionID int64, now time.Time) []ActionItem {
	rows := []struct {
		title  string
		detail string
	}{
		{"整理资源清单", "把技能、预算、时间和人脉写成一页表格"},
		{"确定目标客群", "选择最容易触达的一类客户做首轮验证"},
		{"制作样板", "做出一份客户能看懂的交付样板或诊断表"},
		{"验证市场证据", "收集竞品案例、公开需求和客户痛点"},
		{"触达首批客户", "向 10 个潜在客户发出诊断邀约"},
		{"复盘反馈", "整理价格、信任、交付和时机四类反馈"},
		{"决定下一步", "保留高信号方向，进入获客或 CRM 跟进"},
	}
	items := make([]ActionItem, 0, len(rows))
	for index, row := range rows {
		items = append(items, ActionItem{
			UserID:    userID,
			SessionID: sessionID,
			DayIndex:  index + 1,
			Title:     row.title,
			Detail:    row.detail,
			CreatedAt: now,
			UpdatedAt: now,
		})
	}
	return items
}
