package crm

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
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

func (r *memoryRepository) ListCustomers(_ context.Context, input ListCustomersInput) ([]Customer, error) {
	var customers []Customer
	for _, customer := range r.customers {
		if customer.UserID != input.UserID {
			continue
		}
		if input.Stage != "" && customer.Stage != input.Stage {
			continue
		}
		if input.Source != "" && customer.Source != input.Source {
			continue
		}
		if input.Q != "" && !strings.Contains(customer.Name, input.Q) && !strings.Contains(customer.Phone, input.Q) && !strings.Contains(customer.Email, input.Q) && !strings.Contains(customer.Website, input.Q) {
			continue
		}
		customers = append(customers, customer)
	}
	if input.Limit > 0 && len(customers) > input.Limit {
		return customers[:input.Limit], nil
	}
	return customers, nil
}

func (r *memoryRepository) GetCustomer(_ context.Context, userID, customerID int64) (Customer, error) {
	for _, customer := range r.customers {
		if customer.UserID == userID && customer.ID == customerID {
			return customer, nil
		}
	}
	return Customer{}, ErrCustomerNotFound
}

func (r *memoryRepository) UpdateCustomer(_ context.Context, input UpdateCustomerInput, activity Activity) (Customer, error) {
	for index := range r.customers {
		if r.customers[index].UserID == input.UserID && r.customers[index].ID == input.CustomerID {
			if input.Name != nil {
				r.customers[index].Name = *input.Name
			}
			if input.Phone != nil {
				r.customers[index].Phone = *input.Phone
			}
			if input.Email != nil {
				r.customers[index].Email = *input.Email
			}
			if input.Website != nil {
				r.customers[index].Website = *input.Website
			}
			r.activities = append(r.activities, activity)
			return r.customers[index], nil
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

func (r *memoryRepository) RescheduleFollowUp(_ context.Context, userID, followUpID int64, nextFollowUpAt time.Time, activity Activity) (FollowUp, error) {
	for index := range r.followups {
		if r.followups[index].UserID == userID && r.followups[index].ID == followUpID {
			r.followups[index].NextFollowUpAt = nextFollowUpAt
			for customerIndex := range r.customers {
				if r.customers[customerIndex].UserID == userID && r.customers[customerIndex].ID == r.followups[index].CustomerID {
					r.customers[customerIndex].NextFollowUpAt = nextFollowUpAt
					activity.CustomerID = r.followups[index].CustomerID
					r.activities = append(r.activities, activity)
					return r.followups[index], nil
				}
			}
			return FollowUp{}, ErrCustomerNotFound
		}
	}
	return FollowUp{}, ErrCustomerNotFound
}

func (r *memoryRepository) ListActivities(_ context.Context, userID, customerID int64, limit int) ([]Activity, error) {
	var activities []Activity
	for _, activity := range r.activities {
		if activity.UserID == userID && activity.CustomerID == customerID {
			activities = append(activities, activity)
		}
	}
	if limit > 0 && len(activities) > limit {
		return activities[:limit], nil
	}
	return activities, nil
}

func (r *memoryRepository) ListFollowUps(_ context.Context, input ListFollowUpsInput) ([]FollowUp, error) {
	var followups []FollowUp
	query := strings.ToLower(strings.TrimSpace(input.Q))
	for _, followUp := range r.followups {
		if followUp.UserID != input.UserID {
			continue
		}
		if input.CustomerID > 0 && followUp.CustomerID != input.CustomerID {
			continue
		}
		if query != "" && !strings.Contains(strings.ToLower(followUp.Note), query) && strconv.FormatInt(followUp.CustomerID, 10) != query {
			continue
		}
		if input.HasDueFrom && followUp.NextFollowUpAt.Before(input.DueFrom) {
			continue
		}
		if input.HasDueBefore && !followUp.NextFollowUpAt.Before(input.DueBefore) {
			continue
		}
		followups = append(followups, followUp)
	}
	if input.Limit > 0 && len(followups) > input.Limit {
		return followups[:input.Limit], nil
	}
	return followups, nil
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

func (r *memoryRepository) PipelineStats(_ context.Context, userID int64, dueBefore time.Time) (PipelineStats, error) {
	var stats PipelineStats
	for _, customer := range r.customers {
		if customer.UserID != userID {
			continue
		}
		stats.Total++
		switch customer.Stage {
		case StageNew:
			stats.New++
		case StageContacted:
			stats.Contacted++
		case StageQualified:
			stats.Qualified++
		case StageProposal:
			stats.Proposal++
		case StageWon:
			stats.Won++
		case StageLost:
			stats.Lost++
		}
		if !customer.NextFollowUpAt.IsZero() && !customer.NextFollowUpAt.After(dueBefore) {
			stats.DueToday++
		}
	}
	return stats, nil
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

func TestServiceCreatesManualCustomer(t *testing.T) {
	repository := &memoryRepository{}
	service := NewService(repository)
	now := time.Date(2026, 7, 8, 10, 0, 0, 0, time.UTC)
	service.now = func() time.Time { return now }

	customer, err := service.CreateCustomer(context.Background(), CreateCustomerInput{
		UserID:  42,
		Name:    " 成都启明星教育 ",
		Phone:   " 028-12345678 ",
		Email:   " hello@example.com ",
		Website: " https://example.com ",
	})
	if err != nil {
		t.Fatalf("CreateCustomer() error = %v", err)
	}
	if customer.Name != "成都启明星教育" || customer.Phone != "028-12345678" || customer.Email != "hello@example.com" || customer.Website != "https://example.com" {
		t.Fatalf("customer fields = %+v", customer)
	}
	if customer.Stage != StageNew || customer.Source != SourceManual || customer.ImportKey != "manual:1783504800000000000" {
		t.Fatalf("customer defaults = %+v", customer)
	}
}

func TestServiceImportEnterpriseDeliveryIsIdempotent(t *testing.T) {
	repository := &memoryRepository{}
	service := NewService(repository)
	service.now = func() time.Time { return time.Date(2026, 7, 7, 9, 0, 0, 0, time.UTC) }

	input := ImportEnterpriseInput{
		UserID:             42,
		DiagnosisRequestID: 8,
		Need:               "30人销售团队需要AI获客陪跑",
	}
	first, err := service.ImportEnterpriseDelivery(context.Background(), input)
	if err != nil {
		t.Fatalf("ImportEnterpriseDelivery() first error = %v", err)
	}
	second, err := service.ImportEnterpriseDelivery(context.Background(), input)
	if err != nil {
		t.Fatalf("ImportEnterpriseDelivery() second error = %v", err)
	}
	if first.ID != second.ID || len(repository.customers) != 1 {
		t.Fatalf("first=%+v second=%+v customers=%+v", first, second, repository.customers)
	}
	if first.ImportKey != "enterprise_diagnosis_request:8" || first.Stage != StageWon || first.Source != SourceEnterprise {
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

func TestServiceListsCustomersWithFilters(t *testing.T) {
	repository := &memoryRepository{}
	service := NewService(repository)
	_, _, _ = repository.ImportCustomer(context.Background(), Customer{UserID: 42, ImportKey: "a", Name: "成都启明星教育", Phone: "028-12345678", Stage: StageContacted, Source: SourceEnterprise})
	_, _, _ = repository.ImportCustomer(context.Background(), Customer{UserID: 42, ImportKey: "b", Name: "星桥教育集团", Stage: StageQualified, Source: SourceLead})
	_, _, _ = repository.ImportCustomer(context.Background(), Customer{UserID: 7, ImportKey: "c", Name: "其他用户客户", Stage: StageContacted, Source: SourceEnterprise})

	customers, err := service.ListCustomers(context.Background(), ListCustomersInput{UserID: 42, Stage: StageContacted, Source: SourceEnterprise, Q: "启明星"})
	if err != nil {
		t.Fatalf("ListCustomers() error = %v", err)
	}
	if len(customers) != 1 || customers[0].Name != "成都启明星教育" {
		t.Fatalf("customers = %+v", customers)
	}
}

func TestServiceUpdatesCustomerAndCreatesActivity(t *testing.T) {
	repository := &memoryRepository{}
	service := NewService(repository)
	service.now = func() time.Time { return time.Date(2026, 6, 24, 10, 30, 0, 0, time.UTC) }
	customer, _, _ := repository.ImportCustomer(context.Background(), Customer{UserID: 42, ImportKey: "lead_result:99", Name: "旧客户", Stage: StageNew})
	name := " 成都启明星教育 "
	phone := " 028-12345678 "

	updated, err := service.UpdateCustomer(context.Background(), UpdateCustomerInput{UserID: 42, CustomerID: customer.ID, Name: &name, Phone: &phone})
	if err != nil {
		t.Fatalf("UpdateCustomer() error = %v", err)
	}
	if updated.Name != "成都启明星教育" || updated.Phone != "028-12345678" {
		t.Fatalf("updated = %+v", updated)
	}
	if len(repository.activities) != 1 || repository.activities[0].Type != ActivityCustomerUpdated {
		t.Fatalf("activities = %+v", repository.activities)
	}
}

func TestServiceRejectsEmptyCustomerUpdate(t *testing.T) {
	service := NewService(&memoryRepository{})

	_, err := service.UpdateCustomer(context.Background(), UpdateCustomerInput{UserID: 42, CustomerID: 100})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("UpdateCustomer() error = %v, want ErrInvalidInput", err)
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

func TestServiceReschedulesFollowUpAndCustomerNextDate(t *testing.T) {
	repository := &memoryRepository{}
	service := NewService(repository)
	now := time.Date(2026, 6, 24, 11, 30, 0, 0, time.UTC)
	service.now = func() time.Time { return now }
	customer, _, _ := repository.ImportCustomer(context.Background(), Customer{UserID: 42, ImportKey: "lead_result:99", Name: "成都启明星教育", Stage: StageContacted})
	original := time.Date(2026, 6, 25, 9, 30, 0, 0, time.UTC)
	followUp := FollowUp{ID: 1, UserID: 42, CustomerID: customer.ID, Note: "发送方案", NextFollowUpAt: original}
	repository.followups = append(repository.followups, followUp)
	rescheduledAt := time.Date(2026, 6, 27, 15, 0, 0, 0, time.UTC)

	rescheduled, err := service.RescheduleFollowUp(context.Background(), RescheduleFollowUpInput{UserID: 42, FollowUpID: followUp.ID, NextFollowUpAt: rescheduledAt})
	if err != nil {
		t.Fatalf("RescheduleFollowUp() error = %v", err)
	}

	updated, _ := repository.GetCustomer(context.Background(), 42, customer.ID)
	if rescheduled.NextFollowUpAt != rescheduledAt || updated.NextFollowUpAt != rescheduledAt {
		t.Fatalf("rescheduled=%+v updated=%+v", rescheduled, updated)
	}
	if len(repository.activities) != 1 || repository.activities[0].Type != ActivityFollowUpRescheduled || repository.activities[0].CustomerID != customer.ID {
		t.Fatalf("activities = %+v", repository.activities)
	}
}

func TestServiceListsFollowUpsForUser(t *testing.T) {
	repository := &memoryRepository{}
	service := NewService(repository)
	now := time.Date(2026, 6, 24, 11, 0, 0, 0, time.UTC)
	service.now = func() time.Time { return now }
	repository.followups = []FollowUp{
		{ID: 1, UserID: 42, CustomerID: 100, Note: "发送方案", NextFollowUpAt: now.Add(-time.Hour)},
		{ID: 2, UserID: 42, CustomerID: 101, Note: "预约演示", NextFollowUpAt: now.Add(2 * time.Hour)},
		{ID: 4, UserID: 42, CustomerID: 102, Note: "下周复盘", NextFollowUpAt: now.Add(6 * 24 * time.Hour)},
		{ID: 3, UserID: 7, CustomerID: 100, Note: "其他用户", NextFollowUpAt: now},
	}

	followups, err := service.ListFollowUps(context.Background(), ListFollowUpsInput{UserID: 42, CustomerID: 100})
	if err != nil {
		t.Fatalf("ListFollowUps() error = %v", err)
	}
	if len(followups) != 1 || followups[0].Note != "发送方案" {
		t.Fatalf("followups = %+v", followups)
	}

	followups, err = service.ListFollowUps(context.Background(), ListFollowUpsInput{UserID: 42, Q: "预约"})
	if err != nil {
		t.Fatalf("ListFollowUps() with q error = %v", err)
	}
	if len(followups) != 1 || followups[0].CustomerID != 101 {
		t.Fatalf("followups = %+v", followups)
	}

	followups, err = service.ListFollowUps(context.Background(), ListFollowUpsInput{UserID: 42, Due: "overdue"})
	if err != nil {
		t.Fatalf("ListFollowUps() with overdue error = %v", err)
	}
	if len(followups) != 1 || followups[0].CustomerID != 100 {
		t.Fatalf("followups = %+v", followups)
	}

	followups, err = service.ListFollowUps(context.Background(), ListFollowUpsInput{UserID: 42, Due: "week"})
	if err != nil {
		t.Fatalf("ListFollowUps() with week error = %v", err)
	}
	if len(followups) != 2 {
		t.Fatalf("followups = %+v, want two upcoming records", followups)
	}
}

func TestServiceListsActivitiesForCustomer(t *testing.T) {
	now := time.Date(2026, 6, 24, 11, 0, 0, 0, time.UTC)
	repository := &memoryRepository{activities: []Activity{
		{ID: 1, UserID: 42, CustomerID: 100, Type: ActivityStageChanged, Note: "电话已接通", CreatedAt: now},
		{ID: 2, UserID: 42, CustomerID: 101, Type: ActivityStageChanged, Note: "其他客户", CreatedAt: now},
		{ID: 3, UserID: 7, CustomerID: 100, Type: ActivityStageChanged, Note: "其他用户", CreatedAt: now},
	}}
	service := NewService(repository)

	activities, err := service.ListActivities(context.Background(), 42, 100, 20)
	if err != nil {
		t.Fatalf("ListActivities() error = %v", err)
	}
	if len(activities) != 1 || activities[0].Note != "电话已接通" {
		t.Fatalf("activities = %+v", activities)
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

func TestServiceReturnsPipelineStats(t *testing.T) {
	repository := &memoryRepository{}
	service := NewService(repository)
	now := time.Date(2026, 6, 24, 12, 0, 0, 0, time.UTC)
	service.now = func() time.Time { return now }
	_, _, _ = repository.ImportCustomer(context.Background(), Customer{UserID: 42, ImportKey: "new", Name: "新客户", Stage: StageNew})
	_, _, _ = repository.ImportCustomer(context.Background(), Customer{UserID: 42, ImportKey: "won", Name: "成交客户", Stage: StageWon, NextFollowUpAt: now})
	_, _, _ = repository.ImportCustomer(context.Background(), Customer{UserID: 7, ImportKey: "other", Name: "其他用户", Stage: StageWon})

	stats, err := service.PipelineStats(context.Background(), 42)
	if err != nil {
		t.Fatalf("PipelineStats() error = %v", err)
	}
	if stats.Total != 2 || stats.New != 1 || stats.Won != 1 || stats.DueToday != 1 {
		t.Fatalf("stats = %+v", stats)
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
