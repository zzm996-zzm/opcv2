package projects

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

type analyticsMemoryRepository struct {
	*memoryRepository
	events map[string]AnalyticsEvent
	heatID int64
	cutoff time.Time
}

func (r *analyticsMemoryRepository) RecordProjectEvent(_ context.Context, event AnalyticsEvent) (bool, error) {
	if r.events == nil {
		r.events = map[string]AnalyticsEvent{}
	}
	if _, exists := r.events[event.EventID]; exists {
		return false, nil
	}
	r.events[event.EventID] = event
	return true, nil
}

func (r *analyticsMemoryRepository) RecalculateProjectHeat(_ context.Context, projectID int64, cutoff time.Time) error {
	r.heatID, r.cutoff = projectID, cutoff
	return nil
}

func TestRecordProjectEventHashesVisitorAndSuppressesDuplicates(t *testing.T) {
	repository := &analyticsMemoryRepository{memoryRepository: &memoryRepository{}}
	queue := &generationQueue{}
	now := time.Date(2026, 8, 12, 10, 0, 0, 0, time.UTC)
	service := NewService(repository, nil, WithProjectMatchQueue(queue))
	service.now = func() time.Time { return now }
	input := AnalyticsEventInput{
		EventID: "event-12345678", EventName: ProjectEventDetailView,
		VisitorKey: "browser-visitor-123", Route: "/projects/42", RefModule: "catalog",
		Properties: map[string]any{"project_id": json.Number("42"), "tab": "overview"},
	}
	first, err := service.RecordProjectEvent(context.Background(), input)
	if err != nil {
		t.Fatalf("RecordProjectEvent() error = %v", err)
	}
	second, err := service.RecordProjectEvent(context.Background(), input)
	if err != nil {
		t.Fatalf("duplicate RecordProjectEvent() error = %v", err)
	}
	stored := repository.events[input.EventID]
	if !first.Accepted || first.Duplicate || !second.Duplicate || len(queue.jobs) != 1 {
		t.Fatalf("receipts/jobs = %+v %+v %d", first, second, len(queue.jobs))
	}
	if stored.VisitorHash == input.VisitorKey || len(stored.VisitorHash) != 64 || stored.UserID != nil || stored.ProjectID == nil || *stored.ProjectID != 42 {
		t.Fatalf("stored event leaks identity or project mapping is invalid: %+v", stored)
	}
	if queue.jobs[0].Payload["project_id"] != int64(42) {
		t.Fatalf("heat job = %+v", queue.jobs[0])
	}
}

func TestRecordProjectEventRejectsUnknownAndPrivateProperties(t *testing.T) {
	service := NewService(&analyticsMemoryRepository{memoryRepository: &memoryRepository{}}, nil)
	checks := []AnalyticsEventInput{
		{EventID: "event-unknown-1", EventName: "project_fake", VisitorKey: "visitor-123", Route: "/projects", Properties: map[string]any{}},
		{EventID: "event-private-1", EventName: ProjectEventSearch, VisitorKey: "visitor-123", Route: "/projects", Properties: map[string]any{"user_id": 42}},
		{EventID: "event-heat-0001", EventName: ProjectEventDetailView, VisitorKey: "visitor-123", Route: "/projects/42", Properties: map[string]any{"project_id": 42, "heat": 999}},
	}
	for _, input := range checks {
		if _, err := service.RecordProjectEvent(context.Background(), input); err != ErrInvalidEvent {
			t.Fatalf("input=%+v error=%v, want ErrInvalidEvent", input, err)
		}
	}
}

func TestProcessProjectHeatUsesThirtyDayWindow(t *testing.T) {
	repository := &analyticsMemoryRepository{memoryRepository: &memoryRepository{}}
	now := time.Date(2026, 8, 12, 10, 0, 0, 0, time.UTC)
	service := NewService(repository, nil)
	service.now = func() time.Time { return now }
	if err := service.ProcessProjectHeat(context.Background(), 42); err != nil {
		t.Fatalf("ProcessProjectHeat() error = %v", err)
	}
	if repository.heatID != 42 || repository.cutoff != now.Add(-30*24*time.Hour) {
		t.Fatalf("heat input = %d/%s", repository.heatID, repository.cutoff)
	}
}

func TestAnalyticsEventPropertyWhitelistMatchesPRD(t *testing.T) {
	want := []string{ProjectEventHomeView, ProjectEventSearch, ProjectEventFilterApply, ProjectEventCardClick,
		ProjectEventDetailView, ProjectEventTabSwitch, ProjectEventUnlockClick, ProjectEventDiagnoseSubmit,
		ProjectEventDiagnoseResult, ProjectEventBannerSubmit, ProjectEventMatchFileUpload, ProjectEventMatchParseResult,
		ProjectEventMatchStart, ProjectEventMatchAnswer, ProjectEventMatchResearch, ProjectEventMatchResult,
		ProjectEventExploreView, ProjectEventExploreSource, ProjectEventCaseView, ProjectEventCaseSource,
		ProjectEventMatchExport, ProjectEventCompareAdd, ProjectEventCompareView}
	for _, name := range want {
		if _, ok := projectEventProperties[name]; !ok {
			t.Fatalf("event %q is not whitelisted", name)
		}
	}
	if strings.TrimSpace(ProjectEventLockView) == "" {
		t.Fatal("lock event constant is empty")
	}
}
