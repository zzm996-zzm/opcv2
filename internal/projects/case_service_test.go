package projects

import (
	"context"
	"testing"
)

func TestServiceListsPublishedCasesWithEvidence(t *testing.T) {
	repository := &memoryRepository{cases: []CaseStudy{
		{ID: 1, Slug: "ai-sales-pilot", Title: "AI销售试点", CaseType: "success", Status: CaseStatusPublished, SourceURL: "https://example.com/source"},
		{ID: 2, Slug: "draft", Title: "草稿案例", Status: CaseStatusDraft},
	}}
	service := NewService(repository, nil)

	cases, err := service.ListCases(context.Background(), CaseFilters{CaseType: " success ", Limit: 20})
	if err != nil || len(cases) != 1 || cases[0].SourceURL == "" {
		t.Fatalf("ListCases() = %+v, %v", cases, err)
	}
}
