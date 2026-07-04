package enterprise

import (
	"context"
	"errors"
	"testing"
)

type fakeRepository struct {
	overview Overview
	userID   int64
	err      error
}

func (r *fakeRepository) Overview(_ context.Context, userID int64) (Overview, error) {
	r.userID = userID
	return r.overview, r.err
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
