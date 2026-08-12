package projects

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/zzm/opcv2/internal/jobs"
)

func TestServiceCreatesExportFromOwnedMatchSnapshot(t *testing.T) {
	repository := &memoryRepository{sessions: []MatchSession{{ID: 99, UserID: 42, Intent: "AI项目", Status: StatusCompleted, Result: MatchResult{Status: StatusCompleted}}}}
	queue := &generationQueue{}
	service := NewService(repository, nil, WithProjectMatchQueue(queue))
	export, err := service.CreateExport(context.Background(), CreateExportInput{UserID: 42, SourceType: ExportSourceMatch, SourceID: 99})
	if err != nil || export.ID == 0 || export.Status != "queued" || export.Format != "pdf" || len(export.Snapshot) == 0 {
		t.Fatalf("export = %+v err=%v", export, err)
	}
	if len(queue.jobs) != 1 || queue.jobs[0].Type != jobs.TypeProjectExportRender {
		t.Fatalf("jobs = %+v", queue.jobs)
	}
	second, err := service.CreateExport(context.Background(), CreateExportInput{UserID: 42, SourceType: ExportSourceMatch, SourceID: 99})
	if err != nil || second.ID != export.ID || len(queue.jobs) != 1 {
		t.Fatalf("second=%+v jobs=%+v err=%v", second, queue.jobs, err)
	}
}

func TestProcessProjectExportPersistsRealPDF(t *testing.T) {
	now := time.Date(2026, 8, 12, 12, 0, 0, 0, time.UTC)
	repository := &memoryRepository{exports: []Export{{ID: 71, UserID: 42, SourceType: ExportSourceMatch, SourceID: 99, Status: "queued", Format: "pdf", Snapshot: []byte(`{"id":99}`), ExpiresAt: now.Add(7 * 24 * time.Hour)}}}
	payload := []byte("%PDF-1.7\nrendered")
	service := NewService(repository, nil, WithProjectPDFRenderer(func(item Export) ([]byte, error) {
		if item.ID != 71 || item.Status != "running" {
			t.Fatalf("renderer item = %+v", item)
		}
		return payload, nil
	}))
	service.now = func() time.Time { return now }

	if err := service.ProcessProjectExport(context.Background(), 42, 71); err != nil {
		t.Fatalf("ProcessProjectExport() error = %v", err)
	}
	got, err := service.GetExport(context.Background(), 42, 71)
	if err != nil || got.Status != "ready" || !bytes.Equal(got.Payload, payload) || got.DownloadURL == "" {
		t.Fatalf("export=%+v err=%v", got, err)
	}
}

func TestProcessProjectExportPersistsFailure(t *testing.T) {
	now := time.Date(2026, 8, 12, 12, 0, 0, 0, time.UTC)
	repository := &memoryRepository{exports: []Export{{ID: 71, UserID: 42, Status: "queued", Format: "pdf", Snapshot: []byte(`{"id":99}`), ExpiresAt: now.Add(time.Hour)}}}
	service := NewService(repository, nil, WithProjectPDFRenderer(func(Export) ([]byte, error) { return nil, errors.New("font missing") }))
	service.now = func() time.Time { return now }

	if err := service.ProcessProjectExport(context.Background(), 42, 71); err == nil {
		t.Fatal("ProcessProjectExport() error = nil")
	}
	if repository.exports[0].Status != "failed" || repository.exports[0].ErrorCode != "render_failed" {
		t.Fatalf("export = %+v", repository.exports[0])
	}
}
