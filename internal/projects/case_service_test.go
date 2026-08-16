package projects

import (
	"context"
	"errors"
	"testing"
)

type evidenceCaseMemoryRepository struct {
	*memoryRepository
	page    EvidenceCasePage
	detail  EvidenceCaseDetail
	filters EvidenceCaseFilters
}

func (r *evidenceCaseMemoryRepository) ListEvidenceCases(_ context.Context, filters EvidenceCaseFilters) (EvidenceCasePage, error) {
	r.filters = filters
	return r.page, nil
}

func (r *evidenceCaseMemoryRepository) GetEvidenceCase(context.Context, string) (EvidenceCaseDetail, error) {
	return r.detail, nil
}

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

func TestServiceEnforcesEvidenceCasePublicationThreshold(t *testing.T) {
	repository := &evidenceCaseMemoryRepository{
		memoryRepository: &memoryRepository{},
		page: EvidenceCasePage{Items: []EvidenceCaseItem{
			{ID: 1, Type: "failure", EvidenceStatus: EvidenceStatusVerified, PrimarySourceCount: 1, PrimarySourceURL: "https://authority.example/case"},
			{ID: 2, Type: "success", EvidenceStatus: EvidenceStatusVerified, ReliableSecondaryCount: 2, PrimarySourceURL: "https://media.example/case"},
			{ID: 3, EvidenceStatus: EvidenceStatusVerified, ReliableSecondaryCount: 1, PrimarySourceURL: "https://media.example/weak"},
			{ID: 4, EvidenceStatus: EvidenceStatusVerified, PrimarySourceCount: 1, HasConflict: true, PrimarySourceURL: "https://authority.example/conflict"},
			{ID: 5, EvidenceStatus: EvidenceStatusVerified, PrimarySourceCount: 1, PrimarySourceURL: "https://loot-drop.io/case"},
		}, Total: 5},
	}
	service := NewService(repository, nil)

	page, err := service.ListEvidenceCases(context.Background(), EvidenceCaseFilters{CaseType: "fail", Page: -1, PageSize: 500})
	if err != nil {
		t.Fatalf("ListEvidenceCases() error = %v", err)
	}
	if repository.filters.CaseType != "failure" || repository.filters.Page != 1 || repository.filters.PageSize != 100 {
		t.Fatalf("filters = %+v", repository.filters)
	}
	if len(page.Items) != 3 || page.Items[0].Type != "fail" {
		t.Fatalf("published items = %+v", page.Items)
	}
}

func TestServiceReturnsEvidenceCaseFactsAnalysesAndSources(t *testing.T) {
	repository := &evidenceCaseMemoryRepository{
		memoryRepository: &memoryRepository{},
		detail: EvidenceCaseDetail{
			EvidenceCaseItem: EvidenceCaseItem{
				ID: 1, Type: "failure", EvidenceStatus: EvidenceStatusVerified, PrimarySourceCount: 1,
				PrimarySourceURL: "https://authority.example/case",
			},
			Facts:    []EvidenceFact{{Field: "died_year", Value: "2024", SourceRefs: []int64{11}}},
			Analyses: []EvidenceAnalysis{{Point: "控制获客成本", IsModelGenerated: true, SourceRefs: []int64{11}}},
			Sources:  []EvidenceSource{{ID: 11, URL: "https://authority.example/case"}},
		},
	}
	service := NewService(repository, nil)

	detail, err := service.GetEvidenceCase(context.Background(), "1")
	if err != nil {
		t.Fatalf("GetEvidenceCase() error = %v", err)
	}
	if len(detail.Facts) != 1 || len(detail.Analyses) != 1 || len(detail.Sources) != 1 || detail.Type != "fail" {
		t.Fatalf("detail = %+v", detail)
	}
}

func TestServiceHidesEvidenceCaseWithUnresolvedConflict(t *testing.T) {
	repository := &evidenceCaseMemoryRepository{
		memoryRepository: &memoryRepository{},
		detail: EvidenceCaseDetail{EvidenceCaseItem: EvidenceCaseItem{
			ID: 1, EvidenceStatus: EvidenceStatusVerified, PrimarySourceCount: 1, HasConflict: true,
			PrimarySourceURL: "https://authority.example/case",
		}},
	}
	service := NewService(repository, nil)

	_, err := service.GetEvidenceCase(context.Background(), "1")
	if !errors.Is(err, ErrCaseNotFound) {
		t.Fatalf("GetEvidenceCase() error = %v, want not found", err)
	}
}
