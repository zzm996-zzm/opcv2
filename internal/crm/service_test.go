package crm

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/zzm/opcv2/internal/ai"
)

type fakeJSONGenerator struct {
	request ai.GenerateJSONRequest
	result  ai.GenerateJSONResult
	err     error
	calls   int
}

func (g *fakeJSONGenerator) GenerateJSON(_ context.Context, request ai.GenerateJSONRequest) (ai.GenerateJSONResult, error) {
	g.calls++
	g.request = request
	return g.result, g.err
}

type memoryRepository struct {
	customers  []Customer
	activities []Activity
	followups  []FollowUp
	nextID     int64
}

func (r *memoryRepository) ImportCustomer(_ context.Context, customer Customer) (Customer, bool, error) {
	for _, existing := range r.customers {
		if existing.UserID == customer.UserID && existing.ImportKey == customer.ImportKey {
			return existing, true, nil
		}
	}
	if r.nextID == 0 {
		r.nextID = 100
	}
	customer.ID = r.nextID
	r.nextID++
	r.customers = append(r.customers, customer)
	return customer, false, nil
}

func (r *memoryRepository) GetCustomer(_ context.Context, userID, customerID int64) (Customer, error) {
	for _, customer := range r.customers {
		if customer.UserID == userID && customer.ID == customerID {
			return customer, nil
		}
	}
	return Customer{}, ErrCustomerNotFound
}

func (r *memoryRepository) UpdateStage(ctx context.Context, userID, customerID int64, stage string, activity Activity) (Customer, error) {
	for index := range r.customers {
		if r.customers[index].UserID == userID && r.customers[index].ID == customerID {
			r.customers[index].Stage = stage
			activity.UserID = userID
			activity.CustomerID = customerID
			r.activities = append(r.activities, activity)
			return r.customers[index], nil
		}
	}
	return Customer{}, ErrCustomerNotFound
}

func (r *memoryRepository) RecordFollowUp(ctx context.Context, followUp FollowUp, activity Activity) (FollowUp, error) {
	for index := range r.customers {
		if r.customers[index].UserID == followUp.UserID && r.customers[index].ID == followUp.CustomerID {
			r.customers[index].NextFollowUpAt = followUp.NextFollowUpAt
			followUp.ID = int64(len(r.followups) + 1)
			r.followups = append(r.followups, followUp)
			r.activities = append(r.activities, activity)
			return followUp, nil
		}
	}
	return FollowUp{}, ErrCustomerNotFound
}

func (r *memoryRepository) ListDueCustomers(_ context.Context, userID int64, dueBefore time.Time, limit int) ([]Customer, error) {
	var customers []Customer
	for _, customer := range r.customers {
		if customer.UserID == userID && !customer.NextFollowUpAt.IsZero() && !customer.NextFollowUpAt.After(dueBefore) {
			customers = append(customers, customer)
		}
	}
	if limit > 0 && len(customers) > limit {
		return customers[:limit], nil
	}
	return customers, nil
}

func TestServiceImportLeadIsIdempotent(t *testing.T) {
	repository := &memoryRepository{}
	service := NewService(repository)
	service.now = func() time.Time { return time.Date(2026, 6, 24, 9, 0, 0, 0, time.UTC) }

	input := ImportLeadInput{
		UserID:       42,
		LeadResultID: 99,
		Name:         "成都启明星教育咨询有限公司",
		Phone:        "028-12345678",
		Email:        "hello@example.com",
		Website:      "https://example.com",
	}
	first, err := service.ImportLead(context.Background(), input)
	if err != nil {
		t.Fatalf("ImportLead() first error = %v", err)
	}
	second, err := service.ImportLead(context.Background(), input)
	if err != nil {
		t.Fatalf("ImportLead() second error = %v", err)
	}
	if first.ID != second.ID || len(repository.customers) != 1 {
		t.Fatalf("first=%+v second=%+v customers=%+v", first, second, repository.customers)
	}
	if first.ImportKey != "lead_result:99" || first.Stage != StageNew {
		t.Fatalf("customer = %+v", first)
	}
}

func TestServiceStageUpdateCreatesActivity(t *testing.T) {
	repository := &memoryRepository{}
	service := NewService(repository)
	service.now = func() time.Time { return time.Date(2026, 6, 24, 10, 0, 0, 0, time.UTC) }
	customer, _, _ := repository.ImportCustomer(context.Background(), Customer{UserID: 42, ImportKey: "lead_result:99", Name: "成都启明星教育", Stage: StageNew})

	updated, err := service.UpdateStage(context.Background(), UpdateStageInput{UserID: 42, CustomerID: customer.ID, Stage: StageContacted, Note: "电话已接通"})
	if err != nil {
		t.Fatalf("UpdateStage() error = %v", err)
	}

	if updated.Stage != StageContacted {
		t.Fatalf("updated = %+v", updated)
	}
	if len(repository.activities) != 1 || repository.activities[0].Type != ActivityStageChanged || repository.activities[0].Note != "电话已接通" {
		t.Fatalf("activities = %+v", repository.activities)
	}
}

func TestServiceFollowUpRecordUpdatesNextFollowUpDate(t *testing.T) {
	repository := &memoryRepository{}
	service := NewService(repository)
	service.now = func() time.Time { return time.Date(2026, 6, 24, 11, 0, 0, 0, time.UTC) }
	next := time.Date(2026, 6, 25, 9, 30, 0, 0, time.UTC)
	customer, _, _ := repository.ImportCustomer(context.Background(), Customer{UserID: 42, ImportKey: "lead_result:99", Name: "成都启明星教育", Stage: StageContacted})

	followUp, err := service.RecordFollowUp(context.Background(), RecordFollowUpInput{UserID: 42, CustomerID: customer.ID, Note: "发送方案", NextFollowUpAt: next})
	if err != nil {
		t.Fatalf("RecordFollowUp() error = %v", err)
	}

	updated, _ := repository.GetCustomer(context.Background(), 42, customer.ID)
	if followUp.NextFollowUpAt != next || updated.NextFollowUpAt != next {
		t.Fatalf("followUp=%+v updated=%+v", followUp, updated)
	}
	if len(repository.activities) != 1 || repository.activities[0].Type != ActivityFollowUpRecorded {
		t.Fatalf("activities = %+v", repository.activities)
	}
}

func TestServiceListsDueAndOverdueCustomers(t *testing.T) {
	repository := &memoryRepository{}
	service := NewService(repository)
	now := time.Date(2026, 6, 24, 12, 0, 0, 0, time.UTC)
	service.now = func() time.Time { return now }
	_, _, _ = repository.ImportCustomer(context.Background(), Customer{UserID: 42, ImportKey: "due", Name: "今日到期", NextFollowUpAt: now})
	_, _, _ = repository.ImportCustomer(context.Background(), Customer{UserID: 42, ImportKey: "overdue", Name: "已经逾期", NextFollowUpAt: now.Add(-time.Hour)})
	_, _, _ = repository.ImportCustomer(context.Background(), Customer{UserID: 42, ImportKey: "future", Name: "明天跟进", NextFollowUpAt: now.Add(24 * time.Hour)})
	_, _, _ = repository.ImportCustomer(context.Background(), Customer{UserID: 7, ImportKey: "other", Name: "其他用户", NextFollowUpAt: now.Add(-time.Hour)})

	customers, err := service.ListDueCustomers(context.Background(), ListDueInput{UserID: 42})
	if err != nil {
		t.Fatalf("ListDueCustomers() error = %v", err)
	}
	if len(customers) != 2 {
		t.Fatalf("customers = %+v, want due and overdue only", customers)
	}
}

func TestServiceRejectsOtherUsersCustomer(t *testing.T) {
	repository := &memoryRepository{}
	service := NewService(repository)
	customer, _, _ := repository.ImportCustomer(context.Background(), Customer{UserID: 7, ImportKey: "lead_result:99", Name: "其他用户客户", Stage: StageNew})

	_, err := service.UpdateStage(context.Background(), UpdateStageInput{UserID: 42, CustomerID: customer.ID, Stage: StageContacted})
	if !errors.Is(err, ErrCustomerNotFound) {
		t.Fatalf("UpdateStage() error = %v, want ErrCustomerNotFound", err)
	}
}

func TestServiceGeneratesFollowUpCopyFromOwnedCustomerContext(t *testing.T) {
	content, err := json.Marshal(FollowUpCopy{
		Subject: "智能客服升级方案跟进",
		Body:    "您好，我们已根据贵司教培转化场景整理了下一步方案。",
		Channel: "wechat",
	})
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	repository := &memoryRepository{}
	customer, _, _ := repository.ImportCustomer(context.Background(), Customer{
		UserID:    42,
		ImportKey: "lead_result:99",
		Name:      "成都启明星教育",
		Phone:     "028-12345678",
		Email:     "hello@example.com",
		Website:   "https://example.com",
		Stage:     StageContacted,
		Source:    SourceLead,
	})
	generator := &fakeJSONGenerator{result: ai.GenerateJSONResult{Content: content}}
	service := NewService(repository, generator)

	copy, err := service.GenerateFollowUpCopy(context.Background(), FollowUpCopyInput{UserID: 42, CustomerID: customer.ID, Goal: "推进方案会"})
	if err != nil {
		t.Fatalf("GenerateFollowUpCopy() error = %v", err)
	}

	if copy.Subject == "" || copy.Body == "" || copy.Channel != "wechat" {
		t.Fatalf("copy = %+v", copy)
	}
	if generator.request.Feature != "crm.follow_up_copy" || generator.request.SchemaName != "crm_follow_up_copy" {
		t.Fatalf("AI request = %+v", generator.request)
	}
	if !strings.Contains(generator.request.UserPrompt, "成都启明星教育") || !strings.Contains(generator.request.UserPrompt, "contacted") {
		t.Fatalf("AI prompt missing customer context: %s", generator.request.UserPrompt)
	}
}

func TestServiceDoesNotGenerateFollowUpCopyForOtherUsersCustomer(t *testing.T) {
	repository := &memoryRepository{}
	customer, _, _ := repository.ImportCustomer(context.Background(), Customer{UserID: 7, ImportKey: "lead_result:99", Name: "其他用户客户", Stage: StageContacted})
	generator := &fakeJSONGenerator{result: ai.GenerateJSONResult{Content: []byte(`{"subject":"x","body":"y","channel":"wechat"}`)}}
	service := NewService(repository, generator)

	_, err := service.GenerateFollowUpCopy(context.Background(), FollowUpCopyInput{UserID: 42, CustomerID: customer.ID, Goal: "推进方案会"})
	if !errors.Is(err, ErrCustomerNotFound) {
		t.Fatalf("GenerateFollowUpCopy() error = %v, want ErrCustomerNotFound", err)
	}
	if generator.calls != 0 {
		t.Fatalf("AI generator calls = %d, want 0", generator.calls)
	}
}

func TestServiceReturnsSafeErrorForInvalidAIFollowUpCopyPayload(t *testing.T) {
	repository := &memoryRepository{}
	customer, _, _ := repository.ImportCustomer(context.Background(), Customer{UserID: 42, ImportKey: "lead_result:99", Name: "成都启明星教育", Stage: StageContacted})
	service := NewService(repository, &fakeJSONGenerator{result: ai.GenerateJSONResult{Content: []byte(`{"subject":"字段不完整"}`)}})

	_, err := service.GenerateFollowUpCopy(context.Background(), FollowUpCopyInput{UserID: 42, CustomerID: customer.ID, Goal: "推进方案会"})
	if !errors.Is(err, ErrInvalidAIResult) {
		t.Fatalf("GenerateFollowUpCopy() error = %v, want ErrInvalidAIResult", err)
	}
}
