package enterprise

import (
	"context"
	"errors"
	"strings"
)

var ErrUserIDRequired = errors.New("user id required")
var ErrInvalidInput = errors.New("invalid input")

type Repository interface {
	Overview(ctx context.Context, userID int64) (Overview, error)
	CreateDiagnosisRequest(ctx context.Context, userID int64, input DiagnosisRequestInput) (DiagnosisRequest, error)
	ListDiagnosisRequests(ctx context.Context, userID int64, limit int) ([]DiagnosisRequest, error)
	UpdateDiagnosisRequest(ctx context.Context, userID int64, requestID int64, input DiagnosisRequestUpdateInput) (DiagnosisRequest, error)
}

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) Overview(ctx context.Context, userID int64) (Overview, error) {
	if userID <= 0 {
		return Overview{}, ErrUserIDRequired
	}
	if s.repository == nil {
		return emptyOverview(), nil
	}
	overview, err := s.repository.Overview(ctx, userID)
	if err != nil {
		return Overview{}, err
	}
	return ensureOverviewSlices(overview), nil
}

func (s *Service) CreateDiagnosisRequest(ctx context.Context, userID int64, input DiagnosisRequestInput) (DiagnosisRequest, error) {
	if userID <= 0 {
		return DiagnosisRequest{}, ErrUserIDRequired
	}
	if s.repository == nil {
		return DiagnosisRequest{}, ErrInvalidInput
	}
	input.Need = strings.TrimSpace(input.Need)
	if input.Need == "" {
		return DiagnosisRequest{}, ErrInvalidInput
	}
	return s.repository.CreateDiagnosisRequest(ctx, userID, input)
}

func (s *Service) ListDiagnosisRequests(ctx context.Context, userID int64, limit int) (DiagnosisRequestsResponse, error) {
	if userID <= 0 {
		return DiagnosisRequestsResponse{}, ErrUserIDRequired
	}
	if limit <= 0 || limit > 50 {
		limit = 10
	}
	if s.repository == nil {
		return DiagnosisRequestsResponse{Requests: []DiagnosisRequest{}}, nil
	}
	requests, err := s.repository.ListDiagnosisRequests(ctx, userID, limit)
	if err != nil {
		return DiagnosisRequestsResponse{}, err
	}
	if requests == nil {
		requests = []DiagnosisRequest{}
	}
	return DiagnosisRequestsResponse{Requests: requests}, nil
}

func (s *Service) UpdateDiagnosisRequest(ctx context.Context, userID int64, requestID int64, input DiagnosisRequestUpdateInput) (DiagnosisRequest, error) {
	if userID <= 0 {
		return DiagnosisRequest{}, ErrUserIDRequired
	}
	if s.repository == nil || requestID <= 0 {
		return DiagnosisRequest{}, ErrInvalidInput
	}
	input.Status = strings.TrimSpace(input.Status)
	if input.Status != "follow_up_created" && input.Status != "in_delivery" {
		return DiagnosisRequest{}, ErrInvalidInput
	}
	return s.repository.UpdateDiagnosisRequest(ctx, userID, requestID, input)
}

func emptyOverview() Overview {
	return Overview{
		Stats:         []Metric{},
		Plans:         []Plan{},
		DeliveryBoard: []DeliveryItem{},
		Milestones:    []Milestone{},
		Cases:         []Case{},
	}
}

func ensureOverviewSlices(overview Overview) Overview {
	if overview.Stats == nil {
		overview.Stats = []Metric{}
	}
	if overview.Plans == nil {
		overview.Plans = []Plan{}
	}
	for index := range overview.Plans {
		if overview.Plans[index].Focus == nil {
			overview.Plans[index].Focus = []string{}
		}
	}
	if overview.DeliveryBoard == nil {
		overview.DeliveryBoard = []DeliveryItem{}
	}
	if overview.Milestones == nil {
		overview.Milestones = []Milestone{}
	}
	if overview.Cases == nil {
		overview.Cases = []Case{}
	}
	return overview
}
