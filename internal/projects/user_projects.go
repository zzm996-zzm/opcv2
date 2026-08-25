package projects

import (
	"context"
	"errors"
	"strings"
	"time"
)

var (
	ErrInvalidUserProject      = errors.New("invalid user project")
	ErrUserProjectNotFound     = errors.New("user project not found")
	ErrInvalidUserProjectState = errors.New("invalid user project state")
)

type UserProject struct {
	ID             int64     `json:"id"`
	UserID         int64     `json:"user_id"`
	Name           string    `json:"name"`
	Description    string    `json:"description,omitempty"`
	Status         string    `json:"status"`
	SourceType     string    `json:"source_type,omitempty"`
	SourceID       *int64    `json:"source_id,omitempty"`
	IdempotencyKey string    `json:"-"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type CreateUserProjectInput struct {
	UserID         int64
	Name           string
	Description    string
	SourceType     string
	SourceID       *int64
	IdempotencyKey string
}

type UserProjectRepository interface {
	CreateUserProject(context.Context, UserProject) (UserProject, error)
	ListUserProjects(context.Context, int64, int) ([]UserProject, error)
	UpdateUserProject(context.Context, int64, int64, string, string) (UserProject, error)
	SetUserProjectStatus(context.Context, int64, int64, string) (UserProject, error)
	DeleteUserProject(context.Context, int64, int64) error
}

func (s *Service) CreateUserProject(ctx context.Context, input CreateUserProjectInput) (UserProject, error) {
	repository, ok := s.repository.(UserProjectRepository)
	if !ok || repository == nil {
		return UserProject{}, ErrServiceNotReady
	}
	input.Name = strings.TrimSpace(input.Name)
	input.Description = strings.TrimSpace(input.Description)
	input.SourceType = strings.TrimSpace(input.SourceType)
	input.IdempotencyKey = strings.TrimSpace(input.IdempotencyKey)
	if input.UserID <= 0 || input.Name == "" || len([]rune(input.Name)) > 120 || len([]rune(input.Description)) > 10000 {
		return UserProject{}, ErrInvalidUserProject
	}
	now := s.now()
	return repository.CreateUserProject(ctx, UserProject{
		UserID: input.UserID, Name: input.Name, Description: input.Description, Status: "draft",
		SourceType: input.SourceType, SourceID: input.SourceID, IdempotencyKey: input.IdempotencyKey,
		CreatedAt: now, UpdatedAt: now,
	})
}

func (s *Service) ListUserProjects(ctx context.Context, userID int64, limit int) ([]UserProject, error) {
	repository, ok := s.repository.(UserProjectRepository)
	if !ok || repository == nil {
		return nil, ErrServiceNotReady
	}
	if userID <= 0 || limit <= 0 {
		return nil, ErrInvalidUserProject
	}
	if limit > 100 {
		limit = 100
	}
	return repository.ListUserProjects(ctx, userID, limit)
}

func (s *Service) UpdateUserProject(ctx context.Context, userID, id int64, name, description string) (UserProject, error) {
	repository, ok := s.repository.(UserProjectRepository)
	if !ok || repository == nil {
		return UserProject{}, ErrServiceNotReady
	}
	name = strings.TrimSpace(name)
	description = strings.TrimSpace(description)
	if userID <= 0 || id <= 0 || name == "" || len([]rune(name)) > 120 || len([]rune(description)) > 10000 {
		return UserProject{}, ErrInvalidUserProject
	}
	return repository.UpdateUserProject(ctx, userID, id, name, description)
}

func (s *Service) ArchiveUserProject(ctx context.Context, userID, id int64) (UserProject, error) {
	return s.setUserProjectStatus(ctx, userID, id, "archived")
}

func (s *Service) PublishUserProject(ctx context.Context, userID, id int64) (UserProject, error) {
	return s.setUserProjectStatus(ctx, userID, id, "active")
}

func (s *Service) setUserProjectStatus(ctx context.Context, userID, id int64, status string) (UserProject, error) {
	repository, ok := s.repository.(UserProjectRepository)
	if !ok || repository == nil {
		return UserProject{}, ErrServiceNotReady
	}
	if userID <= 0 || id <= 0 || (status != "active" && status != "archived") {
		return UserProject{}, ErrInvalidUserProjectState
	}
	return repository.SetUserProjectStatus(ctx, userID, id, status)
}

func (s *Service) DeleteUserProject(ctx context.Context, userID, id int64) error {
	repository, ok := s.repository.(UserProjectRepository)
	if !ok || repository == nil {
		return ErrServiceNotReady
	}
	if userID <= 0 || id <= 0 {
		return ErrInvalidUserProject
	}
	return repository.DeleteUserProject(ctx, userID, id)
}
