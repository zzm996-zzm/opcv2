package notifications

import (
	"context"
	"errors"
)

var (
	ErrUserIDRequired        = errors.New("user id required")
	ErrInvalidFilter         = errors.New("invalid notification filter")
	ErrInvalidNotificationID = errors.New("invalid notification id")
	ErrNotificationNotFound  = errors.New("notification not found")
	ErrServiceNotReady       = errors.New("notifications service is not configured")
)

type Repository interface {
	ListNotifications(ctx context.Context, userID int64, filters ListFilters) ([]Notification, error)
	GetNotification(ctx context.Context, userID, id int64) (Notification, error)
	MarkRead(ctx context.Context, userID, id int64) (Notification, error)
	MarkAllRead(ctx context.Context, userID int64) (int, error)
	Summary(ctx context.Context, userID int64) (Summary, error)
}

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) ListNotifications(ctx context.Context, userID int64, filters ListFilters) ([]Notification, error) {
	if err := s.ready(userID); err != nil {
		return nil, err
	}
	if filters.Status == "" {
		filters.Status = StatusAll
	}
	if filters.Limit <= 0 {
		filters.Limit = 20
	}
	if filters.Limit > 100 {
		filters.Limit = 100
	}
	if filters.Type != "" && !validType(filters.Type) {
		return nil, ErrInvalidFilter
	}
	if !validStatus(filters.Status) {
		return nil, ErrInvalidFilter
	}
	return s.repository.ListNotifications(ctx, userID, filters)
}

func (s *Service) GetNotification(ctx context.Context, userID, id int64) (Notification, error) {
	if err := s.ready(userID); err != nil {
		return Notification{}, err
	}
	if id <= 0 {
		return Notification{}, ErrInvalidNotificationID
	}
	return s.repository.GetNotification(ctx, userID, id)
}

func (s *Service) MarkRead(ctx context.Context, userID, id int64) (Notification, error) {
	if err := s.ready(userID); err != nil {
		return Notification{}, err
	}
	if id <= 0 {
		return Notification{}, ErrInvalidNotificationID
	}
	return s.repository.MarkRead(ctx, userID, id)
}

func (s *Service) MarkAllRead(ctx context.Context, userID int64) (int, error) {
	if err := s.ready(userID); err != nil {
		return 0, err
	}
	return s.repository.MarkAllRead(ctx, userID)
}

func (s *Service) Summary(ctx context.Context, userID int64) (Summary, error) {
	if err := s.ready(userID); err != nil {
		return Summary{}, err
	}
	return s.repository.Summary(ctx, userID)
}

func (s *Service) ready(userID int64) error {
	if userID <= 0 {
		return ErrUserIDRequired
	}
	if s.repository == nil {
		return ErrServiceNotReady
	}
	return nil
}

func validType(value string) bool {
	switch value {
	case TypeSystem, TypeTask, TypeAnalysis, TypeLead, TypeCRM, TypeMembership:
		return true
	default:
		return false
	}
}

func validStatus(value string) bool {
	switch value {
	case StatusAll, StatusUnread, StatusRead:
		return true
	default:
		return false
	}
}
