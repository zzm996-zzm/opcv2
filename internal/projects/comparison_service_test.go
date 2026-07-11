package projects

import (
	"context"
	"testing"
)

func TestServiceCreatesComparisonFromPublishedOpportunitySnapshots(t *testing.T) {
	repository := &memoryRepository{opportunities: []Opportunity{
		{ID: 1, Slug: "ai-sales", Title: "AI销售", Status: OpportunityStatusPublished},
		{ID: 2, Slug: "ai-content", Title: "AI内容", Status: OpportunityStatusPublished},
	}}
	service := NewService(repository, nil)
	comparison, err := service.CreateComparison(context.Background(), CreateComparisonInput{UserID: 42, OpportunitySlugs: []string{"ai-sales", "ai-content"}})
	if err != nil || comparison.ID == 0 || len(comparison.Items) != 2 || len(repository.comparisons) != 1 {
		t.Fatalf("comparison = %+v err=%v", comparison, err)
	}
}
