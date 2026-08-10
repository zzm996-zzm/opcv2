package projects

import (
	"context"
	"testing"
	"time"
)

type actionMemoryRepository struct {
	*memoryRepository
	projects  map[string]Project
	favorites []ProjectFavorite
	compare   []ProjectCompareItem
}

var _ Repository = (*actionMemoryRepository)(nil)
var _ ProjectActionRepository = (*actionMemoryRepository)(nil)
var _ CatalogRepository = (*actionMemoryRepository)(nil)

func (r *actionMemoryRepository) ListProjects(context.Context, ProjectFilters) (ProjectPage, error) {
	return ProjectPage{}, nil
}

func (r *actionMemoryRepository) ListDictionaryItems(context.Context, string) ([]DictionaryItem, error) {
	return nil, nil
}

func (r *actionMemoryRepository) GetProject(_ context.Context, ref string) (Project, error) {
	if item, ok := r.projects[ref]; ok {
		return item, nil
	}
	for _, item := range r.projects {
		if item.Slug == ref {
			return item, nil
		}
	}
	return Project{}, ErrOpportunityNotFound
}

func (r *actionMemoryRepository) SaveProjectFavorite(_ context.Context, item ProjectFavorite) (ProjectFavorite, error) {
	for _, existing := range r.favorites {
		if existing.UserID == item.UserID && existing.ProjectID == item.ProjectID {
			return existing, nil
		}
	}
	item.CreatedAt = time.Now()
	r.favorites = append(r.favorites, item)
	return item, nil
}

func (r *actionMemoryRepository) ListProjectFavorites(context.Context, int64, int) ([]ProjectFavorite, error) {
	return append([]ProjectFavorite(nil), r.favorites...), nil
}

func (r *actionMemoryRepository) DeleteProjectFavorite(_ context.Context, userID, projectID int64) error {
	for index, item := range r.favorites {
		if item.UserID == userID && item.ProjectID == projectID {
			r.favorites = append(r.favorites[:index], r.favorites[index+1:]...)
			break
		}
	}
	return nil
}

func (r *actionMemoryRepository) AddProjectCompareItem(_ context.Context, item ProjectCompareItem, _ int64) (ProjectCompareItem, error) {
	r.compare = append(r.compare, item)
	return item, nil
}

func (r *actionMemoryRepository) ListProjectCompareItems(context.Context, int64) ([]ProjectCompareItem, error) {
	return append([]ProjectCompareItem(nil), r.compare...), nil
}

func (r *actionMemoryRepository) DeleteProjectCompareItem(_ context.Context, _ int64, projectID int64) error {
	for index, item := range r.compare {
		if item.ProjectID == projectID {
			r.compare = append(r.compare[:index], r.compare[index+1:]...)
			break
		}
	}
	return nil
}

func TestServiceDiagnosesProjectFromProfilePatch(t *testing.T) {
	repository := &actionMemoryRepository{memoryRepository: &memoryRepository{}, projects: map[string]Project{
		"ai-sales": {ID: 7, Slug: "ai-sales", Title: "AI 销售", Summary: "面向企业销售团队", Track: "销售", BudgetBand: "0-5k", Difficulty: "低", ResourceRequirements: []string{"销售能力"}},
	}}
	if _, ok := any(repository).(CatalogRepository); !ok {
		t.Fatalf("repository does not implement catalog: %T", repository)
	}
	service := NewService(repository, nil)
	result, err := service.DiagnoseProject(context.Background(), ProjectDiagnosisInput{UserID: 42, ProfilePatch: map[string]any{"industry": "销售", "budget": "5k", "role": "销售"}}, "ai-sales")
	if err != nil {
		t.Fatalf("DiagnoseProject() error = %v", err)
	}
	if result.ProjectID != 7 || result.FitScore <= 0 || result.Verdict == "" || len(result.Reasons) == 0 || len(result.NextSteps) == 0 || !result.IsModelGenerated {
		t.Fatalf("diagnosis = %+v", result)
	}
}

func TestServiceLimitsProjectCompareItemsToFive(t *testing.T) {
	repository := &actionMemoryRepository{memoryRepository: &memoryRepository{}, projects: map[string]Project{}}
	for index := int64(1); index <= 6; index++ {
		slug := "project-" + string(rune('0'+index))
		repository.projects[slug] = Project{ID: index, Slug: slug, Title: slug}
	}
	service := NewService(repository, nil)
	for index := int64(1); index <= 5; index++ {
		if _, err := service.AddProjectCompareItem(context.Background(), 42, "project-"+string(rune('0'+index))); err != nil {
			t.Fatalf("AddProjectCompareItem(%d) error = %v", index, err)
		}
	}
	if _, err := service.AddProjectCompareItem(context.Background(), 42, "project-6"); err != ErrCompareLimit {
		t.Fatalf("sixth compare error = %v, want %v", err, ErrCompareLimit)
	}
}
