package dashboard

import (
	"context"
	"testing"
)

type fakeRepository struct {
	userID  int64
	summary Summary
	err     error
}

func (r *fakeRepository) GetSummary(_ context.Context, userID int64) (Summary, error) {
	r.userID = userID
	return r.summary, r.err
}

func TestServiceReturnsUserScopedSummary(t *testing.T) {
	repository := &fakeRepository{summary: Summary{
		Metrics:  []Metric{{Label: "新增线索", Value: "328", Change: "+41%"}},
		Projects: []ProjectOpportunity{{Name: "智能客服系统", Value: "¥86万", Leads: "线索 128", Stage: "进行中"}},
		Trend:    []TrendPoint{{Label: "周一", Value: 34}},
		Pipeline: []PipelineStage{{Stage: "线索", Count: "328", Percent: "100%"}},
		Alerts:   []Alert{{Title: "线索跟进延迟", Detail: "12 条高意向线索超过 24 小时未触达"}},
		Actions:  []Action{{Time: "今天 14:00", Title: "跟进星桥教育集团演示邀约"}},
	}}
	service := NewService(repository)

	summary, err := service.GetSummary(context.Background(), 42)

	if err != nil {
		t.Fatalf("GetSummary() error = %v", err)
	}
	if repository.userID != 42 {
		t.Fatalf("userID = %d, want 42", repository.userID)
	}
	if summary.Metrics[0].Label != "新增线索" || summary.Actions[0].Title == "" {
		t.Fatalf("summary = %+v", summary)
	}
}

func TestServiceReturnsEmptyArrays(t *testing.T) {
	service := NewService(&fakeRepository{})

	summary, err := service.GetSummary(context.Background(), 42)

	if err != nil {
		t.Fatalf("GetSummary() error = %v", err)
	}
	if summary.Metrics == nil || summary.Projects == nil || summary.Trend == nil || summary.Pipeline == nil || summary.Alerts == nil || summary.Actions == nil {
		t.Fatalf("summary has nil slices: %+v", summary)
	}
}
