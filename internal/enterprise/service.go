package enterprise

import (
	"context"
	"errors"
)

var ErrUserIDRequired = errors.New("user id required")

type Repository interface {
	Overview(ctx context.Context, userID int64) (Overview, error)
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
