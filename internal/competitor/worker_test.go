package competitor

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hibiken/asynq"
	"github.com/zzm/opcv2/internal/jobs"
)

type fakeScanProcessor struct {
	scanID int64
	err    error
}

func (p *fakeScanProcessor) ProcessScan(_ context.Context, scanID int64) error {
	p.scanID = scanID
	return p.err
}

func TestWorkerProcessesCompetitorScanEnvelope(t *testing.T) {
	processor := &fakeScanProcessor{}
	handler := NewWorkerHandler(processor)
	payload, err := json.Marshal(jobs.Envelope{
		IdempotencyKey: "competitor-scan-99",
		Payload:        map[string]any{"scan_id": float64(99)},
	})
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	err = handler.Handle(context.Background(), payload)

	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if processor.scanID != 99 {
		t.Fatalf("scanID = %d, want 99", processor.scanID)
	}
}

func TestWorkerRegistersCompetitorScanHandler(t *testing.T) {
	processor := &fakeScanProcessor{}
	mux := asynq.NewServeMux()
	RegisterWorker(mux, processor)
	payload, err := json.Marshal(jobs.Envelope{
		IdempotencyKey: "competitor-scan-99",
		Payload:        map[string]any{"scan_id": float64(99)},
	})
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	err = mux.ProcessTask(context.Background(), asynq.NewTask(jobs.TypeCompetitorScan, payload))

	if err != nil {
		t.Fatalf("ProcessTask() error = %v", err)
	}
	if processor.scanID != 99 {
		t.Fatalf("scanID = %d, want 99", processor.scanID)
	}
}
