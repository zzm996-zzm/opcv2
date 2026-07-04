package geo

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/zzm/opcv2/internal/jobs"
)

var (
	ErrAnalysisRequestNotFound      = errors.New("geo analysis request not found")
	ErrInvalidAnalysisRequest       = errors.New("invalid geo analysis request")
	ErrInvalidAnalysisRequestID     = errors.New("invalid geo analysis request id")
	ErrInvalidAnalysisRequestStatus = errors.New("invalid geo analysis request status")
	ErrServiceNotReady              = errors.New("geo service is not configured")
	ErrUserIDRequired               = errors.New("user id required")
)

const ErrorAnalyzerNotConfigured = "analyzer_not_configured"

type Repository interface {
	Overview(ctx context.Context, userID int64) (Overview, error)
	CreateAnalysisRequest(ctx context.Context, userID int64, input AnalysisRequestInput) (AnalysisRequest, error)
	ListAnalysisRequests(ctx context.Context, userID int64, limit int) ([]AnalysisRequest, error)
	GetAnalysisRequest(ctx context.Context, userID, id int64) (AnalysisRequest, error)
	UpdateAnalysisRequestStatus(ctx context.Context, id int64, status string, errorMessage string) (AnalysisRequest, error)
}

type Queue interface {
	Enqueue(ctx context.Context, job jobs.Job) error
}

type Analyzer interface {
	Analyze(ctx context.Context, request AnalysisRequest) error
}

type Option func(*Service)

type Service struct {
	repository Repository
	queue      Queue
	analyzer   Analyzer
}

func NewService(repository Repository, options ...Option) *Service {
	service := &Service{repository: repository}
	for _, option := range options {
		option(service)
	}
	return service
}

func WithQueue(queue Queue) Option {
	return func(service *Service) {
		service.queue = queue
	}
}

func WithAnalyzer(analyzer Analyzer) Option {
	return func(service *Service) {
		service.analyzer = analyzer
	}
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

func (s *Service) CreateAnalysisRequest(ctx context.Context, userID int64, input AnalysisRequestInput) (AnalysisRequest, error) {
	if userID <= 0 {
		return AnalysisRequest{}, ErrUserIDRequired
	}
	if s.repository == nil {
		return AnalysisRequest{}, ErrServiceNotReady
	}
	input.Target = strings.TrimSpace(input.Target)
	if input.Target == "" {
		return AnalysisRequest{}, ErrInvalidAnalysisRequest
	}
	request, err := s.repository.CreateAnalysisRequest(ctx, userID, input)
	if err != nil {
		return AnalysisRequest{}, err
	}
	if s.queue == nil {
		return request, nil
	}
	err = s.queue.Enqueue(ctx, jobs.Job{
		Type:           jobs.TypeGeoAnalysis,
		IdempotencyKey: fmt.Sprintf("geo-analysis-request-%d", request.ID),
		Payload: map[string]any{
			"request_id": request.ID,
		},
		MaxRetry: 3,
		Timeout:  5 * time.Minute,
	})
	if err != nil {
		return AnalysisRequest{}, err
	}
	return request, nil
}

func (s *Service) ListAnalysisRequests(ctx context.Context, userID int64, limit int) ([]AnalysisRequest, error) {
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
	requests, err := s.repository.ListAnalysisRequests(ctx, userID, limit)
	if requests == nil {
		requests = []AnalysisRequest{}
	}
	return requests, err
}

func (s *Service) GetAnalysisRequest(ctx context.Context, userID, id int64) (AnalysisRequest, error) {
	if userID <= 0 {
		return AnalysisRequest{}, ErrUserIDRequired
	}
	if id <= 0 {
		return AnalysisRequest{}, ErrInvalidAnalysisRequestID
	}
	if s.repository == nil {
		return AnalysisRequest{}, ErrServiceNotReady
	}
	return s.repository.GetAnalysisRequest(ctx, userID, id)
}

func (s *Service) UpdateAnalysisRequestStatus(ctx context.Context, id int64, status string, errorMessage string) (AnalysisRequest, error) {
	if id <= 0 {
		return AnalysisRequest{}, ErrInvalidAnalysisRequestID
	}
	if !validAnalysisRequestStatus(status) {
		return AnalysisRequest{}, ErrInvalidAnalysisRequestStatus
	}
	if s.repository == nil {
		return AnalysisRequest{}, ErrServiceNotReady
	}
	return s.repository.UpdateAnalysisRequestStatus(ctx, id, status, strings.TrimSpace(errorMessage))
}

func (s *Service) ProcessAnalysisRequest(ctx context.Context, id int64) error {
	if id <= 0 {
		return ErrInvalidAnalysisRequestID
	}
	if s.repository == nil {
		return ErrServiceNotReady
	}
	request, err := s.repository.UpdateAnalysisRequestStatus(ctx, id, AnalysisRequestStatusRunning, "")
	if err != nil {
		return err
	}
	if s.analyzer == nil {
		_, err = s.repository.UpdateAnalysisRequestStatus(ctx, id, AnalysisRequestStatusFailed, ErrorAnalyzerNotConfigured)
		return err
	}
	if err := s.analyzer.Analyze(ctx, request); err != nil {
		_, updateErr := s.repository.UpdateAnalysisRequestStatus(ctx, id, AnalysisRequestStatusFailed, strings.TrimSpace(err.Error()))
		if updateErr != nil {
			return updateErr
		}
		return err
	}
	_, err = s.repository.UpdateAnalysisRequestStatus(ctx, id, AnalysisRequestStatusSucceeded, "")
	return err
}

func validAnalysisRequestStatus(status string) bool {
	switch status {
	case AnalysisRequestStatusQueued, AnalysisRequestStatusRunning, AnalysisRequestStatusSucceeded, AnalysisRequestStatusFailed:
		return true
	default:
		return false
	}
}

func emptyOverview() Overview {
	return Overview{
		Stats:        []Metric{},
		Engines:      []EngineCoverage{},
		LeadSignals:  []LeadSignal{},
		Keywords:     []KeywordOpportunity{},
		ContentTasks: []ContentTask{},
	}
}

func ensureOverviewSlices(overview Overview) Overview {
	if overview.Stats == nil {
		overview.Stats = []Metric{}
	}
	if overview.Engines == nil {
		overview.Engines = []EngineCoverage{}
	}
	if overview.LeadSignals == nil {
		overview.LeadSignals = []LeadSignal{}
	}
	if overview.Keywords == nil {
		overview.Keywords = []KeywordOpportunity{}
	}
	if overview.ContentTasks == nil {
		overview.ContentTasks = []ContentTask{}
	}
	return overview
}
