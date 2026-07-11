package tasks

import (
	"context"
	"strings"
	"time"

	"github.com/zzm/opcv2/internal/membership"
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
	recurrence := strings.ToLower(strings.TrimSpace(input.Recurrence))
	if recurrence == "" {
		recurrence = ReminderRecurrenceOnce
	}
	if !validReminderRecurrence(recurrence) {
		return TaskReminder{}, ErrInvalidReminderRecurrence
	}
	if recurrence != ReminderRecurrenceOnce {
		if s.membership == nil {
			return TaskReminder{}, ErrServiceNotReady
		}
		snapshot, err := s.membership.CurrentSnapshot(ctx, input.UserID)
		if err != nil {
			return TaskReminder{}, err
		}
		if snapshot.Plan.Code == "" || snapshot.Plan.Code == membership.PlanFree {
			return TaskReminder{}, ErrRecurringReminderRequiresMembership
		}
	}
	if _, err := s.repository.GetTask(ctx, input.UserID, input.TaskID); err != nil {
		return TaskReminder{}, err
	}
	return repository.UpsertTaskReminder(ctx, TaskReminder{
		TaskID:     input.TaskID,
		UserID:     input.UserID,
		RemindAt:   input.RemindAt,
		Recurrence: recurrence,
		CreatedAt:  now,
	})
}

func validReminderRecurrence(value string) bool {
	switch value {
	case ReminderRecurrenceOnce, ReminderRecurrenceDaily, ReminderRecurrenceWeekly:
		return true
	default:
		return false
	}
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
