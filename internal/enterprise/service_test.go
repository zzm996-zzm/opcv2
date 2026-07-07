package enterprise

import (
	"context"
	"errors"
	"testing"
)

type fakeRepository struct {
	overview       Overview
	diagnosisInput DiagnosisRequestInput
	diagnosis      DiagnosisRequest
	userID         int64
	err            error
}

func (r *fakeRepository) Overview(_ context.Context, userID int64) (Overview, error) {
	r.userID = userID
	return r.overview, r.err
}

func (r *fakeRepository) CreateDiagnosisRequest(_ context.Context, userID int64, input DiagnosisRequestInput) (DiagnosisRequest, error) {
	r.userID = userID
	r.diagnosisInput = input
	return r.diagnosis, r.err
}

func TestServiceReturnsSafeEmptyOverviewWithoutRepository(t *testing.T) {
	service := NewService(nil)

	overview, err := service.Overview(context.Background(), 42)

	if err != nil {
		t.Fatalf("Overview() error = %v", err)
	}
	if overview.Stats == nil || overview.Plans == nil || overview.DeliveryBoard == nil || overview.Milestones == nil || overview.Cases == nil {
		t.Fatalf("overview should contain safe empty slices: %+v", overview)
	}
}

func TestServiceUsesRepository(t *testing.T) {
	repository := &fakeRepository{overview: Overview{Stats: []Metric{{Key: "companies", Label: "服务企业数", Value: "2"}}}}
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

func TestServiceCreatesDiagnosisRequest(t *testing.T) {
	repository := &fakeRepository{diagnosis: DiagnosisRequest{ID: 8, Status: "submitted"}}
	service := NewService(repository)

	request, err := service.CreateDiagnosisRequest(context.Background(), 42, DiagnosisRequestInput{Need: "  30人销售团队需要AI获客陪跑  "})

	if err != nil {
		t.Fatalf("CreateDiagnosisRequest() error = %v", err)
	}
	if request.ID != 8 || repository.userID != 42 || repository.diagnosisInput.Need != "30人销售团队需要AI获客陪跑" {
		t.Fatalf("request/repository = %+v/%+v", request, repository)
	}
}

func TestServiceRejectsInvalidDiagnosisRequest(t *testing.T) {
	service := NewService(&fakeRepository{})

	_, err := service.CreateDiagnosisRequest(context.Background(), 42, DiagnosisRequestInput{Need: "   "})

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("err = %v, want ErrInvalidInput", err)
	}
}
