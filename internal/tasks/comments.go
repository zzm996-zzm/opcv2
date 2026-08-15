package tasks

import (
	"context"
	"strings"
)

func (s *Service) commentRepository() (TaskCommentRepository, error) {
	repository, ok := s.repository.(TaskCommentRepository)
	if !ok {
		return nil, ErrServiceNotReady
	}
	return repository, nil
}

func (s *Service) ListTaskComments(ctx context.Context, userID, taskID int64, limit, offset int) ([]TaskComment, int, error) {
	repository, err := s.commentRepository()
	if err != nil {
		return nil, 0, err
	}
	if _, err := s.repository.GetTask(ctx, userID, taskID); err != nil {
		return nil, 0, err
	}
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	comments, err := repository.ListTaskComments(ctx, userID, taskID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	total, err := repository.CountTaskComments(ctx, userID, taskID)
	if err != nil {
		return nil, 0, err
	}
	return comments, total, nil
}

func (s *Service) CreateTaskComment(ctx context.Context, input CreateTaskCommentInput) (TaskComment, error) {
	repository, err := s.commentRepository()
	if err != nil {
		return TaskComment{}, err
	}
	input.Content = strings.TrimSpace(input.Content)
	if input.UserID <= 0 || input.TaskID <= 0 || input.Content == "" || len([]rune(input.Content)) > 2000 {
		return TaskComment{}, ErrInvalidTaskComment
	}
	return repository.CreateTaskComment(ctx, TaskComment{
		TaskID: input.TaskID, UserID: input.UserID, ParentCommentID: input.ParentCommentID, Content: input.Content,
	})
}

func (s *Service) UpdateTaskComment(ctx context.Context, userID, taskID, id int64, input UpdateTaskCommentInput) (TaskComment, error) {
	repository, err := s.commentRepository()
	if err != nil {
		return TaskComment{}, err
	}
	input.Content = strings.TrimSpace(input.Content)
	if userID <= 0 || taskID <= 0 || id <= 0 || input.Content == "" || len([]rune(input.Content)) > 2000 {
		return TaskComment{}, ErrInvalidTaskComment
	}
	return repository.UpdateTaskComment(ctx, userID, taskID, id, input.Content)
}

func (s *Service) DeleteTaskComment(ctx context.Context, userID, taskID, id int64) error {
	repository, err := s.commentRepository()
	if err != nil {
		return err
	}
	return repository.DeleteTaskComment(ctx, userID, taskID, id)
}
