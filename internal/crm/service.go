package crm

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
	ImportCustomer(ctx context.Context, customer Customer) (Customer, bool, error)
	ListCustomers(ctx context.Context, input ListCustomersInput) ([]Customer, error)
	GetCustomer(ctx context.Context, userID, customerID int64) (Customer, error)
	UpdateCustomer(ctx context.Context, input UpdateCustomerInput, activity Activity) (Customer, error)
	UpdateStage(ctx context.Context, userID, customerID int64, stage string, activity Activity) (Customer, error)
	RecordFollowUp(ctx context.Context, followUp FollowUp, activity Activity) (FollowUp, error)
	ListActivities(ctx context.Context, userID, customerID int64, limit int) ([]Activity, error)
	ListFollowUps(ctx context.Context, input ListFollowUpsInput) ([]FollowUp, error)
	ListDueCustomers(ctx context.Context, userID int64, dueBefore time.Time, limit int) ([]Customer, error)
	PipelineStats(ctx context.Context, userID int64, dueBefore time.Time) (PipelineStats, error)
}

type JSONGenerator interface {
	GenerateJSON(ctx context.Context, request ai.GenerateJSONRequest) (ai.GenerateJSONResult, error)
}

type Service struct {
	repository Repository
	generator  JSONGenerator
	now        func() time.Time
}

func NewService(repository Repository, generators ...JSONGenerator) *Service {
	var generator JSONGenerator
	if len(generators) > 0 {
		generator = generators[0]
	}
	return &Service{repository: repository, generator: generator, now: time.Now}
}

func (s *Service) ImportLead(ctx context.Context, input ImportLeadInput) (Customer, error) {
	if s.repository == nil {
		return Customer{}, ErrServiceNotReady
	}
	input.Name = strings.TrimSpace(input.Name)
	if input.UserID <= 0 || input.LeadResultID <= 0 || input.Name == "" {
		return Customer{}, ErrInvalidInput
	}
	customer, _, err := s.repository.ImportCustomer(ctx, Customer{
		UserID:    input.UserID,
		ImportKey: fmt.Sprintf("lead_result:%d", input.LeadResultID),
		Name:      input.Name,
		Phone:     strings.TrimSpace(input.Phone),
		Email:     strings.TrimSpace(input.Email),
		Website:   strings.TrimSpace(input.Website),
		Stage:     StageNew,
		Source:    SourceLead,
		CreatedAt: s.now(),
	})
	return customer, err
}

func (s *Service) UpdateStage(ctx context.Context, input UpdateStageInput) (Customer, error) {
	if s.repository == nil {
		return Customer{}, ErrServiceNotReady
	}
	input.Stage = strings.TrimSpace(input.Stage)
	if input.UserID <= 0 || input.CustomerID <= 0 || !validStage(input.Stage) {
		return Customer{}, ErrInvalidInput
	}
	now := s.now()
	return s.repository.UpdateStage(ctx, input.UserID, input.CustomerID, input.Stage, Activity{
		UserID:     input.UserID,
		CustomerID: input.CustomerID,
		Type:       ActivityStageChanged,
		Note:       strings.TrimSpace(input.Note),
		CreatedAt:  now,
	})
}

func (s *Service) ListCustomers(ctx context.Context, input ListCustomersInput) ([]Customer, error) {
	if s.repository == nil {
		return nil, ErrServiceNotReady
	}
	if input.UserID <= 0 {
		return nil, ErrInvalidInput
	}
	input.Stage = strings.TrimSpace(input.Stage)
	if input.Stage != "" && !validStage(input.Stage) {
		return nil, ErrInvalidInput
	}
	input.Q = strings.TrimSpace(input.Q)
	if input.Limit <= 0 || input.Limit > 100 {
		input.Limit = defaultListLimit
	}
	return s.repository.ListCustomers(ctx, input)
}

func (s *Service) GetCustomer(ctx context.Context, userID, customerID int64) (Customer, error) {
	if s.repository == nil {
		return Customer{}, ErrServiceNotReady
	}
	if userID <= 0 || customerID <= 0 {
		return Customer{}, ErrInvalidInput
	}
	return s.repository.GetCustomer(ctx, userID, customerID)
}

func (s *Service) UpdateCustomer(ctx context.Context, input UpdateCustomerInput) (Customer, error) {
	if s.repository == nil {
		return Customer{}, ErrServiceNotReady
	}
	if input.UserID <= 0 || input.CustomerID <= 0 {
		return Customer{}, ErrInvalidInput
	}
	changed := false
	if input.Name != nil {
		value := strings.TrimSpace(*input.Name)
		if value == "" {
			return Customer{}, ErrInvalidInput
		}
		input.Name = &value
		changed = true
	}
	if input.Phone != nil {
		value := strings.TrimSpace(*input.Phone)
		input.Phone = &value
		changed = true
	}
	if input.Email != nil {
		value := strings.TrimSpace(*input.Email)
		input.Email = &value
		changed = true
	}
	if input.Website != nil {
		value := strings.TrimSpace(*input.Website)
		input.Website = &value
		changed = true
	}
	if !changed {
		return Customer{}, ErrInvalidInput
	}
	now := s.now()
	return s.repository.UpdateCustomer(ctx, input, Activity{
		UserID:     input.UserID,
		CustomerID: input.CustomerID,
		Type:       ActivityCustomerUpdated,
		Note:       "客户资料已更新",
		CreatedAt:  now,
	})
}

func (s *Service) RecordFollowUp(ctx context.Context, input RecordFollowUpInput) (FollowUp, error) {
	if s.repository == nil {
		return FollowUp{}, ErrServiceNotReady
	}
	input.Note = strings.TrimSpace(input.Note)
	if input.UserID <= 0 || input.CustomerID <= 0 || input.Note == "" || input.NextFollowUpAt.IsZero() {
		return FollowUp{}, ErrInvalidInput
	}
	now := s.now()
	followUp := FollowUp{
		UserID:         input.UserID,
		CustomerID:     input.CustomerID,
		Note:           input.Note,
		NextFollowUpAt: input.NextFollowUpAt,
		CreatedAt:      now,
	}
	activity := Activity{
		UserID:     input.UserID,
		CustomerID: input.CustomerID,
		Type:       ActivityFollowUpRecorded,
		Note:       input.Note,
		CreatedAt:  now,
	}
	return s.repository.RecordFollowUp(ctx, followUp, activity)
}

func (s *Service) ListFollowUps(ctx context.Context, input ListFollowUpsInput) ([]FollowUp, error) {
	if s.repository == nil {
		return nil, ErrServiceNotReady
	}
	if input.UserID <= 0 || input.CustomerID < 0 {
		return nil, ErrInvalidInput
	}
	if input.Limit <= 0 || input.Limit > 100 {
		input.Limit = defaultListLimit
	}
	return s.repository.ListFollowUps(ctx, input)
}

func (s *Service) ListActivities(ctx context.Context, userID, customerID int64, limit int) ([]Activity, error) {
	if s.repository == nil {
		return nil, ErrServiceNotReady
	}
	if userID <= 0 || customerID <= 0 {
		return nil, ErrInvalidInput
	}
	if limit <= 0 || limit > 100 {
		limit = defaultListLimit
	}
	return s.repository.ListActivities(ctx, userID, customerID, limit)
}

func (s *Service) ListDueCustomers(ctx context.Context, input ListDueInput) ([]Customer, error) {
	if s.repository == nil {
		return nil, ErrServiceNotReady
	}
	if input.UserID <= 0 {
		return nil, ErrInvalidInput
	}
	limit := input.Limit
	if limit <= 0 || limit > 100 {
		limit = defaultListLimit
	}
	return s.repository.ListDueCustomers(ctx, input.UserID, s.now(), limit)
}

func (s *Service) PipelineStats(ctx context.Context, userID int64) (PipelineStats, error) {
	if s.repository == nil {
		return PipelineStats{}, ErrServiceNotReady
	}
	if userID <= 0 {
		return PipelineStats{}, ErrInvalidInput
	}
	return s.repository.PipelineStats(ctx, userID, s.now())
}

func (s *Service) GenerateFollowUpCopy(ctx context.Context, input FollowUpCopyInput) (FollowUpCopy, error) {
	if s.repository == nil || s.generator == nil {
		return FollowUpCopy{}, ErrServiceNotReady
	}
	input.Goal = strings.TrimSpace(input.Goal)
	if input.UserID <= 0 || input.CustomerID <= 0 || input.Goal == "" {
		return FollowUpCopy{}, ErrInvalidInput
	}
	customer, err := s.repository.GetCustomer(ctx, input.UserID, input.CustomerID)
	if err != nil {
		return FollowUpCopy{}, err
	}
	aiResult, err := s.generator.GenerateJSON(ctx, ai.GenerateJSONRequest{
		UserID:         input.UserID,
		Feature:        "crm.follow_up_copy",
		PromptVersion:  "crm_follow_up_copy_v1",
		SystemPrompt:   "你是 CRM 跟进助手。必须只返回 JSON，字段严格匹配 crm_follow_up_copy。",
		UserPrompt:     followUpCopyPrompt(customer, input.Goal),
		SchemaName:     "crm_follow_up_copy",
		Validate:       validateFollowUpCopyJSON,
		RepairAttempts: 1,
	})
	if err != nil {
		return FollowUpCopy{}, fmt.Errorf("%w: %v", ErrInvalidAIResult, err)
	}
	var copy FollowUpCopy
	if err := json.Unmarshal(aiResult.Content, &copy); err != nil {
		return FollowUpCopy{}, fmt.Errorf("%w: %v", ErrInvalidAIResult, err)
	}
	if err := validateFollowUpCopy(copy); err != nil {
		return FollowUpCopy{}, fmt.Errorf("%w: %v", ErrInvalidAIResult, err)
	}
	return copy, nil
}

func validStage(stage string) bool {
	switch stage {
	case StageNew, StageContacted, StageQualified, StageProposal, StageWon, StageLost:
		return true
	default:
		return false
	}
}

func followUpCopyPrompt(customer Customer, goal string) string {
	parts := []string{
		"客户名称：" + customer.Name,
		"阶段：" + customer.Stage,
		"来源：" + customer.Source,
		"跟进目标：" + goal,
	}
	if customer.Phone != "" {
		parts = append(parts, "电话："+customer.Phone)
	}
	if customer.Email != "" {
		parts = append(parts, "邮箱："+customer.Email)
	}
	if customer.Website != "" {
		parts = append(parts, "网站："+customer.Website)
	}
	return strings.Join(parts, "\n")
}

func validateFollowUpCopyJSON(data []byte) error {
	var copy FollowUpCopy
	if err := json.Unmarshal(data, &copy); err != nil {
		return err
	}
	return validateFollowUpCopy(copy)
}

func validateFollowUpCopy(copy FollowUpCopy) error {
	if strings.TrimSpace(copy.Subject) == "" || strings.TrimSpace(copy.Body) == "" || strings.TrimSpace(copy.Channel) == "" {
		return errors.New("follow-up copy is missing required fields")
	}
	return nil
}
