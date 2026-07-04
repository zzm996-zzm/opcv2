package geo

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/zzm/opcv2/internal/jobs"
)

type fakeQueue struct {
	jobs []jobs.Job
	err  error
}

func (q *fakeQueue) Enqueue(_ context.Context, job jobs.Job) error {
	q.jobs = append(q.jobs, job)
	return q.err
}

type fakeAnalyzer struct {
	request AnalysisRequest
	err     error
}

func (a *fakeAnalyzer) Analyze(_ context.Context, request AnalysisRequest) error {
	a.request = request
	return a.err
}

type fakeRepository struct {
	overview        Overview
	analysisRequest AnalysisRequest
	requests        []AnalysisRequest
	userID          int64
	requestID       int64
	input           AnalysisRequestInput
	status          string
	errorMessage    string
	limit           int
	err             error
}

func (r *fakeRepository) Overview(_ context.Context, userID int64) (Overview, error) {
	r.userID = userID
	return r.overview, r.err
}

func (r *fakeRepository) CreateAnalysisRequest(_ context.Context, userID int64, input AnalysisRequestInput) (AnalysisRequest, error) {
	r.userID = userID
	r.input = input
	return r.analysisRequest, r.err
}

func (r *fakeRepository) ListAnalysisRequests(_ context.Context, userID int64, limit int) ([]AnalysisRequest, error) {
	r.userID = userID
	r.limit = limit
	return r.requests, r.err
}

func (r *fakeRepository) GetAnalysisRequest(_ context.Context, userID, id int64) (AnalysisRequest, error) {
	r.userID = userID
	r.requestID = id
	return r.analysisRequest, r.err
}

func (r *fakeRepository) UpdateAnalysisRequestStatus(_ context.Context, id int64, status string, errorMessage string) (AnalysisRequest, error) {
	r.requestID = id
	r.status = status
	r.errorMessage = errorMessage
	return r.analysisRequest, r.err
}

func TestServiceReturnsSafeEmptyOverviewWithoutRepository(t *testing.T) {
	service := NewService(nil)

	overview, err := service.Overview(context.Background(), 42)

	if err != nil {
		t.Fatalf("Overview() error = %v", err)
	}
	if overview.Stats == nil || overview.Engines == nil || overview.LeadSignals == nil || overview.Keywords == nil || overview.ContentTasks == nil {
		t.Fatalf("overview should contain safe empty slices: %+v", overview)
	}
}

func TestServiceUsesRepository(t *testing.T) {
	repository := &fakeRepository{overview: Overview{Stats: []Metric{{Key: "coverage", Label: "AI引用覆盖", Value: "12%"}}}}
	service := NewService(repository)

	overview, err := service.Overview(context.Background(), 42)

	if err != nil {
		t.Fatalf("Overview() error = %v", err)
	}
	if repository.userID != 42 || len(overview.Stats) != 1 {
		t.Fatalf("user/overview = %d/%+v", repository.userID, overview)
	}
}

func TestServiceRejectsMissingUserID(t *testing.T) {
	service := NewService(nil)

	_, err := service.Overview(context.Background(), 0)

	if !errors.Is(err, ErrUserIDRequired) {
		t.Fatalf("err = %v, want ErrUserIDRequired", err)
	}
}

func TestServiceCreatesAnalysisRequest(t *testing.T) {
	repository := &fakeRepository{analysisRequest: AnalysisRequest{ID: 7, UserID: 42, Target: "面向制造业的 AI 质检工具", Status: AnalysisRequestStatusQueued}}
	queue := &fakeQueue{}
	service := NewService(repository, WithQueue(queue))

	request, err := service.CreateAnalysisRequest(context.Background(), 42, AnalysisRequestInput{Target: "  面向制造业的 AI 质检工具  "})

	if err != nil {
		t.Fatalf("CreateAnalysisRequest() error = %v", err)
	}
	if repository.userID != 42 || repository.input.Target != "面向制造业的 AI 质检工具" {
		t.Fatalf("repository call = user:%d input:%+v", repository.userID, repository.input)
	}
	if request.ID != 7 || request.Status != AnalysisRequestStatusQueued {
		t.Fatalf("request = %+v", request)
	}
	if len(queue.jobs) != 1 {
		t.Fatalf("jobs = %+v, want one GEO analysis job", queue.jobs)
	}
	job := queue.jobs[0]
	if job.Type != jobs.TypeGeoAnalysis || job.IdempotencyKey != "geo-analysis-request-7" || job.Payload["request_id"] != int64(7) {
		t.Fatalf("job = %+v", job)
	}
}

func TestServiceRejectsInvalidAnalysisRequest(t *testing.T) {
	service := NewService(&fakeRepository{})

	_, err := service.CreateAnalysisRequest(context.Background(), 42, AnalysisRequestInput{Target: "   "})

	if !errors.Is(err, ErrInvalidAnalysisRequest) {
		t.Fatalf("err = %v, want ErrInvalidAnalysisRequest", err)
	}
}

func TestServiceRejectsAnalysisRequestWithoutRepository(t *testing.T) {
	service := NewService(nil)

	_, err := service.CreateAnalysisRequest(context.Background(), 42, AnalysisRequestInput{Target: "AI 获客"})

	if !errors.Is(err, ErrServiceNotReady) {
		t.Fatalf("err = %v, want ErrServiceNotReady", err)
	}
}

func TestServiceListsAnalysisRequests(t *testing.T) {
	now := time.Date(2026, 7, 4, 8, 0, 0, 0, time.UTC)
	repository := &fakeRepository{requests: []AnalysisRequest{{ID: 7, UserID: 42, Target: "面向制造业的 AI 质检工具", Status: AnalysisRequestStatusQueued, CreatedAt: now, UpdatedAt: now}}}
	service := NewService(repository)

	requests, err := service.ListAnalysisRequests(context.Background(), 42, 200)

	if err != nil {
		t.Fatalf("ListAnalysisRequests() error = %v", err)
	}
	if repository.userID != 42 || repository.limit != 100 {
		t.Fatalf("repository call = user:%d limit:%d", repository.userID, repository.limit)
	}
	if len(requests) != 1 || requests[0].ID != 7 {
		t.Fatalf("requests = %+v", requests)
	}
}

func TestServiceReturnsEmptyAnalysisRequestArrays(t *testing.T) {
	service := NewService(&fakeRepository{})

	requests, err := service.ListAnalysisRequests(context.Background(), 42, 0)

	if err != nil {
		t.Fatalf("ListAnalysisRequests() error = %v", err)
	}
	if requests == nil {
		t.Fatal("requests should be an empty array, got nil")
	}
}

func TestServiceGetsAnalysisRequest(t *testing.T) {
	repository := &fakeRepository{analysisRequest: AnalysisRequest{ID: 7, UserID: 42, Target: "面向制造业的 AI 质检工具", Status: AnalysisRequestStatusQueued}}
	service := NewService(repository)

	request, err := service.GetAnalysisRequest(context.Background(), 42, 7)

	if err != nil {
		t.Fatalf("GetAnalysisRequest() error = %v", err)
	}
	if repository.userID != 42 || repository.requestID != 7 {
		t.Fatalf("repository call = user:%d request:%d", repository.userID, repository.requestID)
	}
	if request.ID != 7 || request.Target == "" {
		t.Fatalf("request = %+v", request)
	}
}

func TestServiceRejectsInvalidAnalysisRequestID(t *testing.T) {
	service := NewService(&fakeRepository{})

	_, err := service.GetAnalysisRequest(context.Background(), 42, 0)

	if !errors.Is(err, ErrInvalidAnalysisRequestID) {
		t.Fatalf("err = %v, want ErrInvalidAnalysisRequestID", err)
	}
}

func TestServiceUpdatesAnalysisRequestStatus(t *testing.T) {
	repository := &fakeRepository{analysisRequest: AnalysisRequest{ID: 7, Status: AnalysisRequestStatusRunning}}
	service := NewService(repository)

	request, err := service.UpdateAnalysisRequestStatus(context.Background(), 7, AnalysisRequestStatusRunning, "")

	if err != nil {
		t.Fatalf("UpdateAnalysisRequestStatus() error = %v", err)
	}
	if repository.requestID != 7 || repository.status != AnalysisRequestStatusRunning || repository.errorMessage != "" {
		t.Fatalf("repository call = id:%d status:%q error:%q", repository.requestID, repository.status, repository.errorMessage)
	}
	if request.Status != AnalysisRequestStatusRunning {
		t.Fatalf("request = %+v", request)
	}
}

func TestServiceRejectsInvalidAnalysisRequestStatus(t *testing.T) {
	service := NewService(&fakeRepository{})

	_, err := service.UpdateAnalysisRequestStatus(context.Background(), 7, "done", "")

	if !errors.Is(err, ErrInvalidAnalysisRequestStatus) {
		t.Fatalf("err = %v, want ErrInvalidAnalysisRequestStatus", err)
	}
}

func TestServiceProcessAnalysisRequestFailsTransparentlyWithoutAnalyzer(t *testing.T) {
	repository := &fakeRepository{analysisRequest: AnalysisRequest{ID: 7, UserID: 42, Target: "面向制造业的 AI 质检工具"}}
	service := NewService(repository)

	err := service.ProcessAnalysisRequest(context.Background(), 7)

	if err != nil {
		t.Fatalf("ProcessAnalysisRequest() error = %v", err)
	}
	if repository.requestID != 7 || repository.status != AnalysisRequestStatusFailed || repository.errorMessage != ErrorAnalyzerNotConfigured {
		t.Fatalf("final status update = id:%d status:%q error:%q", repository.requestID, repository.status, repository.errorMessage)
	}
}

func TestServiceProcessAnalysisRequestMarksSucceededAfterAnalyzer(t *testing.T) {
	repository := &fakeRepository{analysisRequest: AnalysisRequest{ID: 7, UserID: 42, Target: "面向制造业的 AI 质检工具"}}
	analyzer := &fakeAnalyzer{}
	service := NewService(repository, WithAnalyzer(analyzer))

	err := service.ProcessAnalysisRequest(context.Background(), 7)

	if err != nil {
		t.Fatalf("ProcessAnalysisRequest() error = %v", err)
	}
	if analyzer.request.ID != 7 || repository.status != AnalysisRequestStatusSucceeded || repository.errorMessage != "" {
		t.Fatalf("analyzer=%+v final status=%q error=%q", analyzer.request, repository.status, repository.errorMessage)
	}
}
