package geo

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hibiken/asynq"
	"github.com/zzm/opcv2/internal/jobs"
)

type fakeAnalysisProcessor struct {
	requestID int64
	err       error
}

func (p *fakeAnalysisProcessor) ProcessAnalysisRequest(_ context.Context, requestID int64) error {
	p.requestID = requestID
	return p.err
}

func TestWorkerProcessesGeoAnalysisEnvelope(t *testing.T) {
	processor := &fakeAnalysisProcessor{}
	handler := NewWorkerHandler(processor)
	payload, err := json.Marshal(jobs.Envelope{
		IdempotencyKey: "geo-analysis-request-7",
		Payload:        map[string]any{"request_id": float64(7)},
	})
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	err = handler.Handle(context.Background(), payload)
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if processor.requestID != 7 {
		t.Fatalf("requestID = %d, want 7", processor.requestID)
	}
}

func TestWorkerRegistersGeoAnalysisHandler(t *testing.T) {
	processor := &fakeAnalysisProcessor{}
	mux := asynq.NewServeMux()
	RegisterWorker(mux, processor)
	payload, err := json.Marshal(jobs.Envelope{
		IdempotencyKey: "geo-analysis-request-7",
		Payload:        map[string]any{"request_id": float64(7)},
	})
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	err = mux.ProcessTask(context.Background(), asynq.NewTask(jobs.TypeGeoAnalysis, payload))
	if err != nil {
		t.Fatalf("ProcessTask() error = %v", err)
	}
	if processor.requestID != 7 {
		t.Fatalf("requestID = %d, want 7", processor.requestID)
	}
}
