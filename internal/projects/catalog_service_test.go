package projects

import (
	"context"
	"errors"
	"testing"
)

type catalogMemoryRepository struct {
	*memoryRepository
	dictionaries []DictionaryItem
	page         ProjectPage
	project      Project
	filters      ProjectFilters
}

func (r *catalogMemoryRepository) ListDictionaryItems(context.Context, string) ([]DictionaryItem, error) {
	return r.dictionaries, nil
}

func (r *catalogMemoryRepository) ListProjects(_ context.Context, filters ProjectFilters) (ProjectPage, error) {
	r.filters = filters
	return r.page, nil
}

func (r *catalogMemoryRepository) GetProject(context.Context, string) (Project, error) {
	return r.project, nil
}

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

func TestServiceNormalizesProjectCatalogAndEvidence(t *testing.T) {
	repository := &catalogMemoryRepository{
		memoryRepository: &memoryRepository{},
		page: ProjectPage{Items: []Project{
			{ID: 1, Slug: "valid", IsReal: true, SourceURL: "https://example.com/source"},
			{ID: 2, Slug: "loot-drop", IsReal: true, SourceURL: "https://loot-drop.io/startups/2"},
			{ID: 3, Slug: "invalid", IsReal: true, SourceURL: "javascript:alert(1)"},
		}, Total: 3},
	}
	service := NewService(repository, nil)

	page, err := service.ListProjects(context.Background(), ProjectFilters{Page: -1, PageSize: 500, Sort: "fit"})
	if err != nil {
		t.Fatalf("ListProjects() error = %v", err)
	}
	if repository.filters.Page != 1 || repository.filters.PageSize != 100 || repository.filters.Sort != "heat" {
		t.Fatalf("normalized filters = %+v", repository.filters)
	}
	if !page.Items[0].IsReal || page.Items[0].SourceURL == "" || !page.Items[0].IsUnlocked || page.Items[0].LockedBlocks == nil {
		t.Fatalf("valid project = %+v", page.Items[0])
	}
	if !page.Items[1].IsReal || page.Items[1].SourceURL == "" || page.Items[2].IsReal || page.Items[2].SourceURL != "" {
		t.Fatalf("source normalization was incorrect: %+v", page.Items)
	}
}

func TestServiceReturnsProjectMarketHome(t *testing.T) {
	repository := &catalogMemoryRepository{
		memoryRepository: &memoryRepository{},
		page:             ProjectPage{Items: []Project{{ID: 1, Title: "精选项目"}}},
	}
	service := NewService(repository, nil)

	home, err := service.GetProjectHome(context.Background())
	if err != nil {
		t.Fatalf("GetProjectHome() error = %v", err)
	}
	if len(home.QuickTags) != 6 || len(home.Entries) != 3 || len(home.Featured) != 1 {
		t.Fatalf("home = %+v", home)
	}
	if repository.filters.Featured == nil || !*repository.filters.Featured || repository.filters.PageSize != 8 {
		t.Fatalf("featured filters = %+v", repository.filters)
	}
}

func TestServiceRejectsUnknownProjectDictionary(t *testing.T) {
	service := NewService(&catalogMemoryRepository{memoryRepository: &memoryRepository{}}, nil)
	_, err := service.ListDictionaryItems(context.Background(), "unknown")
	if !errors.Is(err, ErrInvalidDictionaryKind) {
		t.Fatalf("ListDictionaryItems() error = %v, want invalid kind", err)
	}
}
