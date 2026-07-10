package tasks

import (
	"context"
	"strings"
)

type SubtaskRepository interface {
	ListSubtasks(context.Context, int64, int64) ([]Subtask, error)
	CreateSubtask(context.Context, Subtask) (Subtask, error)
	UpdateSubtask(context.Context, int64, int64, int64, SubtaskUpdate) (Subtask, error)
	DeleteSubtask(context.Context, int64, int64, int64) error
}

func (s *Service) subtaskRepository() (SubtaskRepository, error) {
	repository, ok := s.repository.(SubtaskRepository)
	if !ok {
		return nil, ErrServiceNotReady
	}
	return repository, nil
}

func (s *Service) ListSubtasks(ctx context.Context, userID, taskID int64) ([]Subtask, error) {
	repository, err := s.subtaskRepository()
	if err != nil {
		return nil, err
	}
	return repository.ListSubtasks(ctx, userID, taskID)
}

func (s *Service) CreateSubtask(ctx context.Context, input CreateSubtaskInput) (Subtask, error) {
	repository, err := s.subtaskRepository()
	if err != nil {
		return Subtask{}, err
	}
	return repository.CreateSubtask(ctx, Subtask{
		TaskID:    input.TaskID,
		UserID:    input.UserID,
		Title:     strings.TrimSpace(input.Title),
		Assignee:  strings.TrimSpace(input.Assignee),
		DueAt:     input.DueAt,
		CreatedAt: s.now(),
	})
}

func (s *Service) UpdateSubtask(ctx context.Context, userID, taskID, id int64, update SubtaskUpdate) (Subtask, error) {
	repository, err := s.subtaskRepository()
	if err != nil {
		return Subtask{}, err
	}
	if update.Title != nil {
		value := strings.TrimSpace(*update.Title)
		update.Title = &value
	}
	if update.Assignee != nil {
		value := strings.TrimSpace(*update.Assignee)
		update.Assignee = &value
	}
	return repository.UpdateSubtask(ctx, userID, taskID, id, update)
}

func (s *Service) DeleteSubtask(ctx context.Context, userID, taskID, id int64) error {
	repository, err := s.subtaskRepository()
	if err != nil {
		return err
	}
	return repository.DeleteSubtask(ctx, userID, taskID, id)
}
