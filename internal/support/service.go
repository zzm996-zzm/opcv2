package support

import (
	"context"
	"errors"
	"strings"
)

var (
	ErrInvalidInput    = errors.New("invalid support input")
	ErrUserIDRequired  = errors.New("user id required")
	ErrServiceNotReady = errors.New("support service is not configured")
)

type Repository interface {
	CreateTicket(ctx context.Context, userID int64, input TicketInput) (Ticket, error)
	ListTickets(ctx context.Context, userID int64, limit int) ([]Ticket, error)
}

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) CreateTicket(ctx context.Context, userID int64, input TicketInput) (Ticket, error) {
	if userID <= 0 {
		return Ticket{}, ErrUserIDRequired
	}
	if s.repository == nil {
		return Ticket{}, ErrServiceNotReady
	}
	input.Topic = strings.TrimSpace(input.Topic)
	input.Title = strings.TrimSpace(input.Title)
	input.Body = strings.TrimSpace(input.Body)
	if input.Title == "" || input.Body == "" {
		return Ticket{}, ErrInvalidInput
	}
	return s.repository.CreateTicket(ctx, userID, input)
}

func (s *Service) ListTickets(ctx context.Context, userID int64, limit int) ([]Ticket, error) {
	if userID <= 0 {
		return nil, ErrUserIDRequired
	}
	if s.repository == nil {
		return nil, ErrServiceNotReady
	}
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	rows, err := s.repository.ListTickets(ctx, userID, limit)
	if rows == nil {
		rows = []Ticket{}
	}
	return rows, err
}
