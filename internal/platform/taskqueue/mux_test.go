package taskqueue

import (
	"context"
	"testing"

	"github.com/hibiken/asynq"
	"github.com/zzm/opcv2/internal/jobs"
)

func TestNewMuxRejectsUnregisteredTask(t *testing.T) {
	mux := NewMux()
	task := asynq.NewTask("unknown:task", nil)

	if err := mux.ProcessTask(context.Background(), task); err == nil {
		t.Fatal("ProcessTask() error = nil, want unregistered task error")
	}
}

func TestDefaultRegistryIncludesKnownJobTypes(t *testing.T) {
	registry := DefaultRegistry()

	if !registry[jobs.TypeLeadSearch] {
		t.Fatalf("DefaultRegistry() = %+v, want lead search type", registry)
	}
	if !registry[jobs.TypeGeoAnalysis] {
		t.Fatalf("DefaultRegistry() = %+v, want GEO analysis type", registry)
	}
	if !registry[jobs.TypeCompetitorScan] {
		t.Fatalf("DefaultRegistry() = %+v, want competitor scan type", registry)
	}
	if !registry[jobs.TypeSandboxRun] {
		t.Fatalf("DefaultRegistry() = %+v, want sandbox run type", registry)
	}
	if !registry[jobs.TypeProjectMatchGenerate] {
		t.Fatalf("DefaultRegistry() = %+v, want project match generation type", registry)
	}
	if !registry[jobs.TypeProjectExportRender] {
		t.Fatalf("DefaultRegistry() = %+v, want project export type", registry)
	}
	if !registry[jobs.TypeProjectHeatAggregate] {
		t.Fatalf("DefaultRegistry() = %+v, want project heat aggregation type", registry)
	}
}
