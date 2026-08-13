package sandbox

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeV2AnalyticsRepository struct {
	event     SandboxAnalyticsEvent
	recorded  bool
	recordErr error
}

func (r *fakeV2AnalyticsRepository) RecordSandboxEvent(_ context.Context, event SandboxAnalyticsEvent) (bool, error) {
	r.event = event
	return r.recorded, r.recordErr
}

func TestRecordSandboxEventValidatesWhitelistAndHashesVisitor(t *testing.T) {
	repository := &fakeV2AnalyticsRepository{recorded: true}
	v2 := newFakeV2Repository()
	v2.run = V2SandboxRun{ID: 99, UserID: 42}
	service := NewService(&analyticsRepositoryAdapter{fakeV2Repository: v2, analytics: repository}, nil)
	service.now = func() time.Time { return time.Date(2026, 8, 13, 12, 0, 0, 0, time.UTC) }
	receipt, err := service.RecordSandboxEvent(context.Background(), SandboxAnalyticsInput{
		UserID: 42, EventID: "sandbox-event-001", EventName: SandboxEventRolesSelect,
		VisitorKey: "visitor-key-001", Route: "/sandbox/new", RunID: 99,
		Properties: map[string]any{"roles": []any{"customer", "skeptic"}, "count": float64(2)},
	})
	if err != nil || !receipt.Accepted || receipt.Duplicate {
		t.Fatalf("receipt/error = %+v/%v", receipt, err)
	}
	if len(repository.event.VisitorHash) != 64 || repository.event.VisitorHash == "visitor-key-001" || repository.event.EventName != SandboxEventRolesSelect {
		t.Fatalf("event = %+v", repository.event)
	}
}

func TestRecordSandboxEventRejectsUnknownEventsAndProperties(t *testing.T) {
	service := NewService(&analyticsRepositoryAdapter{fakeV2Repository: newFakeV2Repository(), analytics: &fakeV2AnalyticsRepository{}}, nil)
	for _, input := range []SandboxAnalyticsInput{
		{UserID: 42, EventID: "sandbox-event-001", EventName: "sandbox_unknown", VisitorKey: "visitor-key-001", Route: "/sandbox"},
		{UserID: 42, EventID: "sandbox-event-002", EventName: SandboxEventHomeView, VisitorKey: "visitor-key-001", Route: "/sandbox", Properties: map[string]any{"user_id": float64(42)}},
	} {
		if _, err := service.RecordSandboxEvent(context.Background(), input); !errors.Is(err, ErrV2InvalidEvent) {
			t.Fatalf("input/error = %+v/%v", input, err)
		}
	}
}

type analyticsRepositoryAdapter struct {
	*fakeV2Repository
	analytics *fakeV2AnalyticsRepository
}

func (r *analyticsRepositoryAdapter) RecordSandboxEvent(ctx context.Context, event SandboxAnalyticsEvent) (bool, error) {
	return r.analytics.RecordSandboxEvent(ctx, event)
}
