package tasks

import (
	"context"
	"time"
)

type ReminderRepository interface {
	GetTaskReminder(context.Context, int64, int64) (*TaskReminder, error)
	UpsertTaskReminder(context.Context, TaskReminder) (TaskReminder, error)
	DeleteTaskReminder(context.Context, int64, int64) error
}

type ReminderDispatcherRepository interface {
	DispatchDueTaskReminders(context.Context, time.Time, int) (int, error)
}

func (s *Service) reminderRepository() (ReminderRepository, error) {
	repository, ok := s.repository.(ReminderRepository)
	if !ok {
		return nil, ErrServiceNotReady
	}
	return repository, nil
}

func (s *Service) GetTaskReminder(ctx context.Context, userID, taskID int64) (*TaskReminder, error) {
	repository, err := s.reminderRepository()
	if err != nil {
		return nil, err
	}
	if _, err := s.repository.GetTask(ctx, userID, taskID); err != nil {
		return nil, err
	}
	return repository.GetTaskReminder(ctx, userID, taskID)
}

func (s *Service) UpsertTaskReminder(ctx context.Context, input UpsertTaskReminderInput) (TaskReminder, error) {
	repository, err := s.reminderRepository()
	if err != nil {
		return TaskReminder{}, err
	}
	now := s.now()
	if input.RemindAt.IsZero() || !input.RemindAt.After(now) {
		return TaskReminder{}, ErrInvalidReminderTime
	}
	if _, err := s.repository.GetTask(ctx, input.UserID, input.TaskID); err != nil {
		return TaskReminder{}, err
	}
	return repository.UpsertTaskReminder(ctx, TaskReminder{
		TaskID:    input.TaskID,
		UserID:    input.UserID,
		RemindAt:  input.RemindAt,
		CreatedAt: now,
	})
}

func (s *Service) DeleteTaskReminder(ctx context.Context, userID, taskID int64) error {
	repository, err := s.reminderRepository()
	if err != nil {
		return err
	}
	if _, err := s.repository.GetTask(ctx, userID, taskID); err != nil {
		return err
	}
	return repository.DeleteTaskReminder(ctx, userID, taskID)
}

func (s *Service) DispatchDueTaskReminders(ctx context.Context, limit int) (int, error) {
	repository, ok := s.repository.(ReminderDispatcherRepository)
	if !ok {
		return 0, ErrServiceNotReady
	}
	if limit <= 0 {
		limit = 100
	}
	if limit > 500 {
		limit = 500
	}
	return repository.DispatchDueTaskReminders(ctx, s.now(), limit)
}
