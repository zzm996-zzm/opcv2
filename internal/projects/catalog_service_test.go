package projects

import (
	"context"
	"testing"
)

func TestServiceListsPublishedOpportunities(t *testing.T) {
	repository := &memoryRepository{opportunities: []Opportunity{
		{ID: 1, Slug: "ai-sales", Title: "AI销售顾问", Status: OpportunityStatusPublished},
		{ID: 2, Slug: "draft", Title: "未发布项目", Status: OpportunityStatusDraft},
	}}
	service := NewService(repository, nil)

	opportunities, err := service.ListOpportunities(context.Background(), OpportunityFilters{Query: "销售", Limit: 20})
	if err != nil || len(opportunities) != 1 || opportunities[0].Slug != "ai-sales" {
		t.Fatalf("ListOpportunities() = %+v, %v", opportunities, err)
	}
}

func TestServiceGetsPublishedOpportunityBySlug(t *testing.T) {
	repository := &memoryRepository{opportunities: []Opportunity{{ID: 1, Slug: "ai-sales", Title: "AI销售顾问", Status: OpportunityStatusPublished}}}
	service := NewService(repository, nil)

	opportunity, err := service.GetOpportunity(context.Background(), " ai-sales ")
	if err != nil || opportunity.Title != "AI销售顾问" {
		t.Fatalf("GetOpportunity() = %+v, %v", opportunity, err)
	}
}
