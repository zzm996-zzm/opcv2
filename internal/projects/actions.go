package projects

import (
	"context"
	"strings"
)

const maxProjectCollectionItems = 5

func (s *Service) projectActionRepository() (ProjectActionRepository, error) {
	if s.repository == nil {
		return nil, ErrServiceNotReady
	}
	repository, ok := s.repository.(ProjectActionRepository)
	if !ok {
		return nil, ErrServiceNotReady
	}
	return repository, nil
}

func (s *Service) FavoriteProject(ctx context.Context, userID int64, ref string) (ProjectFavorite, error) {
	repository, err := s.projectActionRepository()
	if err != nil {
		return ProjectFavorite{}, err
	}
	project, err := s.GetProject(ctx, strings.TrimSpace(ref))
	if err != nil {
		return ProjectFavorite{}, err
	}
	return repository.SaveProjectFavorite(ctx, ProjectFavorite{UserID: userID, ProjectID: project.ID, Slug: project.Slug, Title: project.Title, CreatedAt: s.now()})
}

func (s *Service) ListFavoriteProjects(ctx context.Context, userID int64, limit int) ([]ProjectFavorite, error) {
	repository, err := s.projectActionRepository()
	if err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return repository.ListProjectFavorites(ctx, userID, limit)
}

func (s *Service) UnfavoriteProject(ctx context.Context, userID int64, ref string) error {
	repository, err := s.projectActionRepository()
	if err != nil {
		return err
	}
	project, err := s.GetProject(ctx, strings.TrimSpace(ref))
	if err != nil {
		return err
	}
	return repository.DeleteProjectFavorite(ctx, userID, project.ID)
}

func (s *Service) AddProjectCompareItem(ctx context.Context, userID int64, ref string) (ProjectCompareItem, error) {
	repository, err := s.projectActionRepository()
	if err != nil {
		return ProjectCompareItem{}, err
	}
	project, err := s.GetProject(ctx, strings.TrimSpace(ref))
	if err != nil {
		return ProjectCompareItem{}, err
	}
	items, err := repository.ListProjectCompareItems(ctx, userID)
	if err != nil {
		return ProjectCompareItem{}, err
	}
	for _, item := range items {
		if item.ProjectID == project.ID {
			return item, nil
		}
	}
	if len(items) >= maxProjectCollectionItems {
		return ProjectCompareItem{}, ErrCompareLimit
	}
	return repository.AddProjectCompareItem(ctx, ProjectCompareItem{ProjectID: project.ID, Slug: project.Slug, Title: project.Title, AddedAt: s.now()}, userID)
}

func (s *Service) ListProjectCompareItems(ctx context.Context, userID int64) ([]ProjectCompareItem, error) {
	repository, err := s.projectActionRepository()
	if err != nil {
		return nil, err
	}
	return repository.ListProjectCompareItems(ctx, userID)
}

func (s *Service) RemoveProjectCompareItem(ctx context.Context, userID int64, ref string) error {
	repository, err := s.projectActionRepository()
	if err != nil {
		return err
	}
	project, err := s.GetProject(ctx, strings.TrimSpace(ref))
	if err != nil {
		return err
	}
	return repository.DeleteProjectCompareItem(ctx, userID, project.ID)
}
