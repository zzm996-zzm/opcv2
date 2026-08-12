package projects

import (
	"context"
	"errors"
	"testing"
	"time"
)

type kbMemoryRepository struct {
	*contentAdminMemoryRepository
	version           int
	reindex           KBReindexJob
	answers           []ProjectAIAnswer
	contentRun        ContentProductionRun
	failedCode        string
	contentFailedCode string
	rebuildErr        error
	produceErr        error
}

func (r *kbMemoryRepository) GetActiveProjectKBVersion(context.Context) (int, error) {
	return r.version, nil
}
func (r *kbMemoryRepository) CreateProjectKBReindex(_ context.Context, userID int64, key string, at time.Time) (KBReindexJob, error) {
	if r.reindex.ID != 0 {
		return r.reindex, nil
	}
	r.reindex = KBReindexJob{ID: 91, TaskKey: key, RequestedBy: &userID, SourceVersion: r.version, TargetVersion: r.version + 1, Status: OperationsStatusQueued, CreatedAt: at, UpdatedAt: at, created: true}
	return r.reindex, nil
}
func (r *kbMemoryRepository) GetProjectKBReindex(context.Context, int64) (KBReindexJob, error) {
	return r.reindex, nil
}
func (r *kbMemoryRepository) MarkProjectKBReindexFailed(_ context.Context, _ int64, code string, _ time.Time) error {
	r.failedCode = code
	return nil
}
func (r *kbMemoryRepository) RebuildProjectKB(_ context.Context, id int64, at time.Time) (KBReindexJob, error) {
	if r.rebuildErr != nil {
		return KBReindexJob{}, r.rebuildErr
	}
	r.version++
	r.reindex.ID, r.reindex.TargetVersion, r.reindex.Status, r.reindex.FinishedAt = id, r.version, OperationsStatusReady, &at
	return r.reindex, nil
}
func (r *kbMemoryRepository) ListProjectAIAnswers(context.Context, int) ([]ProjectAIAnswer, error) {
	return r.answers, nil
}
func (r *kbMemoryRepository) SaveProjectAIAnswer(_ context.Context, item ProjectAIAnswer) error {
	r.answers = append(r.answers, item)
	return nil
}
func (r *kbMemoryRepository) CreateContentProductionRun(_ context.Context, key, date string, planned int, at time.Time) (ContentProductionRun, error) {
	if r.contentRun.ID != 0 {
		return r.contentRun, nil
	}
	r.contentRun = ContentProductionRun{ID: 92, TaskKey: key, ScheduleDate: date, Status: OperationsStatusQueued, PlannedCount: planned, CreatedAt: at, UpdatedAt: at, created: true}
	return r.contentRun, nil
}
func (r *kbMemoryRepository) ProduceProjectContent(_ context.Context, id int64, at time.Time) (ContentProductionRun, error) {
	if r.produceErr != nil {
		return ContentProductionRun{}, r.produceErr
	}
	r.contentRun.ID, r.contentRun.Status, r.contentRun.ProducedCount, r.contentRun.FinishedAt = id, OperationsStatusReady, 30, &at
	return r.contentRun, nil
}
func (r *kbMemoryRepository) MarkContentProductionFailed(_ context.Context, _ int64, code string, _ time.Time) error {
	r.contentFailedCode = code
	return nil
}

func TestRequestProjectKBReindexUsesVersionedIdempotencyKey(t *testing.T) {
	repository := &kbMemoryRepository{contentAdminMemoryRepository: &contentAdminMemoryRepository{memoryRepository: &memoryRepository{}, admin: true}, version: 7}
	queue := &generationQueue{}
	service := NewService(repository, nil, WithProjectMatchQueue(queue))
	job, err := service.RequestProjectKBReindex(context.Background(), 42)
	if err != nil {
		t.Fatalf("RequestProjectKBReindex() error = %v", err)
	}
	if job.SourceVersion != 7 || job.TargetVersion != 8 || job.TaskKey != "project-kb-reindex-v8" || len(queue.jobs) != 1 || queue.jobs[0].IdempotencyKey != job.TaskKey {
		t.Fatalf("job/queue = %+v/%+v", job, queue.jobs)
	}
}

func TestProcessProjectKBReindexIncrementsActiveVersion(t *testing.T) {
	repository := &kbMemoryRepository{contentAdminMemoryRepository: &contentAdminMemoryRepository{memoryRepository: &memoryRepository{}, admin: true}, version: 7, reindex: KBReindexJob{ID: 91, Status: OperationsStatusQueued}}
	service := NewService(repository, nil)
	if err := service.ProcessProjectKBReindex(context.Background(), 91); err != nil {
		t.Fatalf("ProcessProjectKBReindex() error = %v", err)
	}
	if repository.version != 8 || repository.reindex.Status != OperationsStatusReady {
		t.Fatalf("repository = %+v", repository)
	}
}

func TestProjectOperationsPersistTerminalFailure(t *testing.T) {
	repository := &kbMemoryRepository{
		contentAdminMemoryRepository: &contentAdminMemoryRepository{memoryRepository: &memoryRepository{}, admin: true},
		version:                      1, rebuildErr: errors.New("rebuild failed"), produceErr: errors.New("production failed"),
	}
	service := NewService(repository, nil)
	if err := service.ProcessProjectKBReindex(context.Background(), 91); err == nil || repository.failedCode != "reindex_failed" {
		t.Fatalf("reindex error/code = %v/%q", err, repository.failedCode)
	}
	if err := service.ProcessProjectContentBatch(context.Background(), 92); err == nil || repository.contentFailedCode != "production_failed" {
		t.Fatalf("content error/code = %v/%q", err, repository.contentFailedCode)
	}
}

func TestScheduleContentProductionOnlyQueuesReviewingProductionRun(t *testing.T) {
	repository := &kbMemoryRepository{contentAdminMemoryRepository: &contentAdminMemoryRepository{memoryRepository: &memoryRepository{}, admin: true}, version: 1}
	queue := &generationQueue{}
	service := NewService(repository, nil, WithProjectMatchQueue(queue))
	at := time.Date(2026, 8, 11, 2, 0, 0, 0, projectOperationsLocation())
	run, err := service.ScheduleContentProduction(context.Background(), at)
	if err != nil {
		t.Fatalf("ScheduleContentProduction() error = %v", err)
	}
	if run.TaskKey != "project-content-batch-2026-08-11" || run.PlannedCount != 30 || len(queue.jobs) != 1 || queue.jobs[0].Type != "projects.content_batch" {
		t.Fatalf("run/queue = %+v/%+v", run, queue.jobs)
	}
}

func TestRunProjectContentSchedulerUsesTuesdayAndFridayWindow(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	repository := &kbMemoryRepository{contentAdminMemoryRepository: &contentAdminMemoryRepository{memoryRepository: &memoryRepository{}, admin: true}, version: 1}
	queue := &generationQueue{}
	service := NewService(repository, nil, WithProjectMatchQueue(queue))
	friday := time.Date(2026, 8, 14, 2, 15, 0, 0, projectOperationsLocation())
	go RunProjectContentScheduler(ctx, service, time.Hour, func() time.Time { return friday }, nil)
	deadline := time.Now().Add(time.Second)
	for len(queue.jobs) == 0 && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	cancel()
	if len(queue.jobs) != 1 || queue.jobs[0].Type != "projects.content_batch" {
		t.Fatalf("jobs = %+v", queue.jobs)
	}
}
