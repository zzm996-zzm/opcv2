package dashboard

import "context"

type Repository interface {
	GetSummary(ctx context.Context, userID int64) (Summary, error)
}

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) GetSummary(ctx context.Context, userID int64) (Summary, error) {
	if s.repository == nil {
		return Summary{}, ErrServiceNotReady
	}
	summary, err := s.repository.GetSummary(ctx, userID)
	if err != nil {
		return Summary{}, err
	}
	return ensureArrays(summary), nil
}

func ensureArrays(summary Summary) Summary {
	if summary.Metrics == nil {
		summary.Metrics = []Metric{}
	}
	if summary.Projects == nil {
		summary.Projects = []ProjectOpportunity{}
	}
	if summary.Trend == nil {
		summary.Trend = []TrendPoint{}
	}
	if summary.Pipeline == nil {
		summary.Pipeline = []PipelineStage{}
	}
	if summary.Alerts == nil {
		summary.Alerts = []Alert{}
	}
	if summary.Actions == nil {
		summary.Actions = []Action{}
	}
	return summary
}
