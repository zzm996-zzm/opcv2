package growth

import (
	"context"
	"testing"
	"time"
)

type snapshotRepository struct {
	draftRepository
	snapshot  ModelSnapshot
	snapshots []ModelSnapshot
}

func (r *snapshotRepository) CreateSnapshot(_ context.Context, snapshot ModelSnapshot) (ModelSnapshot, error) {
	snapshot.ID = 501
	r.snapshot = snapshot
	r.snapshots = append(r.snapshots, snapshot)
	return snapshot, nil
}

func (r *snapshotRepository) ListSnapshots(_ context.Context, userID, modelID int64, _ int) ([]ModelSnapshot, error) {
	rows := make([]ModelSnapshot, 0)
	for _, snapshot := range r.snapshots {
		if snapshot.UserID == userID && snapshot.ModelID == modelID {
			rows = append(rows, snapshot)
		}
	}
	return rows, nil
}

func TestCalculateDraftPersistsCompleteSnapshot(t *testing.T) {
	now := time.Date(2026, 7, 11, 10, 0, 0, 0, time.UTC)
	repository := &snapshotRepository{draftRepository: draftRepository{draft: Draft{
		ID: 71, UserID: 42, Input: "企业培训增长测算", Status: DraftStatusReady,
		Assumptions: Assumptions{MonthlyVisits: 12000, LeadRate: 0.08, DealRate: 0.15, AverageOrder: 6000, AcquisitionCost: 80, DeliveryCost: 120000},
	}}}
	service := NewService(repository)
	service.now = func() time.Time { return now }

	calculation, err := service.CalculateDraft(context.Background(), CalculateDraftInput{UserID: 42, DraftID: 71})
	if err != nil {
		t.Fatalf("CalculateDraft() error = %v", err)
	}
	if calculation.Snapshot.ID != 501 || repository.snapshot.ModelID != calculation.Model.ID {
		t.Fatalf("calculation/snapshot = %+v/%+v", calculation, repository.snapshot)
	}
	if len(repository.snapshot.Scenarios.Scenarios) != 3 || len(repository.snapshot.Forecast.Months) != 5 || len(repository.snapshot.Recommendations.ActionItems) == 0 {
		t.Fatalf("snapshot = %+v", repository.snapshot)
	}
}

func TestServiceListsOwnedModelSnapshots(t *testing.T) {
	repository := &snapshotRepository{draftRepository: draftRepository{fakeRepository: fakeRepository{model: Model{ID: 99, UserID: 42}}}, snapshots: []ModelSnapshot{
		{ID: 1, UserID: 42, ModelID: 99},
		{ID: 2, UserID: 7, ModelID: 99},
	}}
	service := NewService(repository)
	snapshots, err := service.ListSnapshots(context.Background(), 42, 99, 20)
	if err != nil || len(snapshots) != 1 || snapshots[0].ID != 1 {
		t.Fatalf("ListSnapshots() = %+v, %v", snapshots, err)
	}
}
