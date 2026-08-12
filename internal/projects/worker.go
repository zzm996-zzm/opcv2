package projects

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/hibiken/asynq"
	"github.com/zzm/opcv2/internal/jobs"
)

var ErrInvalidProjectMatchJob = errors.New("invalid project match job payload")

type ProjectMatchProcessor interface {
	ProcessProjectMatch(context.Context, int64, int64, int) error
	ProcessProjectExport(context.Context, int64, int64) error
	ProcessProjectHeat(context.Context, int64) error
	ProcessProjectKBReindex(context.Context, int64) error
	ProcessProjectContentBatch(context.Context, int64) error
}

type ProjectMatchWorker struct {
	processor ProjectMatchProcessor
}

func (w *ProjectMatchWorker) HandleExport(ctx context.Context, envelope jobs.Envelope) error {
	userID, userOK := projectMatchJobNumber(envelope.Payload["user_id"])
	exportID, exportOK := projectMatchJobNumber(envelope.Payload["export_id"])
	if !userOK || !exportOK {
		return ErrInvalidProjectMatchJob
	}
	return w.processor.ProcessProjectExport(ctx, userID, exportID)
}

func (w *ProjectMatchWorker) HandleHeat(ctx context.Context, envelope jobs.Envelope) error {
	projectID, ok := projectMatchJobNumber(envelope.Payload["project_id"])
	if !ok {
		return ErrInvalidProjectMatchJob
	}
	return w.processor.ProcessProjectHeat(ctx, projectID)
}

func (w *ProjectMatchWorker) HandleKBReindex(ctx context.Context, envelope jobs.Envelope) error {
	id, ok := projectMatchJobNumber(envelope.Payload["reindex_id"])
	if !ok {
		return ErrInvalidProjectMatchJob
	}
	return w.processor.ProcessProjectKBReindex(ctx, id)
}

func (w *ProjectMatchWorker) HandleContentBatch(ctx context.Context, envelope jobs.Envelope) error {
	id, ok := projectMatchJobNumber(envelope.Payload["run_id"])
	if !ok {
		return ErrInvalidProjectMatchJob
	}
	return w.processor.ProcessProjectContentBatch(ctx, id)
}

func NewProjectMatchWorker(processor ProjectMatchProcessor) *ProjectMatchWorker {
	return &ProjectMatchWorker{processor: processor}
}

func (w *ProjectMatchWorker) Handle(ctx context.Context, envelope jobs.Envelope) error {
	userID, userOK := projectMatchJobNumber(envelope.Payload["user_id"])
	matchID, matchOK := projectMatchJobNumber(envelope.Payload["match_id"])
	attempt, attemptOK := projectMatchJobNumber(envelope.Payload["attempt"])
	if !userOK || !matchOK || !attemptOK {
		return ErrInvalidProjectMatchJob
	}
	return w.processor.ProcessProjectMatch(ctx, userID, matchID, int(attempt))
}

func RegisterWorker(mux *asynq.ServeMux, processor ProjectMatchProcessor) {
	worker := NewProjectMatchWorker(processor)
	jobs.Register(mux, jobs.Handler{Type: jobs.TypeProjectMatchGenerate, Handle: worker.Handle})
	jobs.Register(mux, jobs.Handler{Type: jobs.TypeProjectExportRender, Handle: worker.HandleExport})
	jobs.Register(mux, jobs.Handler{Type: jobs.TypeProjectHeatAggregate, Handle: worker.HandleHeat})
	jobs.Register(mux, jobs.Handler{Type: jobs.TypeProjectKBReindex, Handle: worker.HandleKBReindex})
	jobs.Register(mux, jobs.Handler{Type: jobs.TypeProjectContentBatch, Handle: worker.HandleContentBatch})
}

func projectMatchJobNumber(value any) (int64, bool) {
	switch typed := value.(type) {
	case float64:
		parsed := int64(typed)
		return parsed, parsed > 0 && typed == float64(parsed)
	case int:
		return int64(typed), typed > 0
	case int64:
		return typed, typed > 0
	case json.Number:
		parsed, err := typed.Int64()
		return parsed, err == nil && parsed > 0
	default:
		return 0, false
	}
}
