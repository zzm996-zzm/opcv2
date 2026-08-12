package projects

import (
	"context"
	"testing"
	"time"
)

type contentAdminMemoryRepository struct {
	*memoryRepository
	admin   bool
	created ImportBatch
	items   []ImportFailureInput
	audits  []OperationAudit
}

func (r *contentAdminMemoryRepository) IsProjectAdmin(context.Context, int64) (bool, error) {
	return r.admin, nil
}

func (r *contentAdminMemoryRepository) CreateImportBatch(_ context.Context, batch ImportBatch, items []ImportFailureInput) (ImportBatch, error) {
	batch.ID = 81
	r.created, r.items = batch, append([]ImportFailureInput(nil), items...)
	return batch, nil
}

func (r *contentAdminMemoryRepository) ListImportBatches(context.Context, int) ([]ImportBatch, error) {
	return []ImportBatch{r.created}, nil
}

func (r *contentAdminMemoryRepository) GetImportBatch(context.Context, int64) (ImportBatch, error) {
	return r.created, nil
}

func (r *contentAdminMemoryRepository) PublishImportBatch(_ context.Context, id, _ int64, at time.Time) (ImportBatch, error) {
	r.created.ID, r.created.Status, r.created.PublishedAt = id, ImportBatchPublished, &at
	return r.created, nil
}

func (r *contentAdminMemoryRepository) RollbackImportBatch(_ context.Context, id, _ int64, at time.Time) (ImportBatch, error) {
	r.created.ID, r.created.Status, r.created.RolledBackAt = id, ImportBatchRolledBack, &at
	return r.created, nil
}

func (r *contentAdminMemoryRepository) ListProjectOperationAudits(context.Context, int) ([]OperationAudit, error) {
	return r.audits, nil
}

func validImportFailure(slug string) ImportFailureInput {
	return ImportFailureInput{
		Slug: slug, Name: "Example", ValuePropZH: "提供服务", DeathCauseZH: "获客成本过高",
		FailureAnalysisZH: "渠道与客单价不匹配", LearningsZH: []string{"先验证渠道"},
		Sources: []ImportSourceInput{{FieldName: "death_cause", URL: "https://example.com/failure", Kind: "authority"}},
	}
}

func TestCreateImportBatchPartiallyAcceptsValidItems(t *testing.T) {
	repository := &contentAdminMemoryRepository{memoryRepository: &memoryRepository{}, admin: true}
	service := NewService(repository, nil)
	service.now = func() time.Time { return time.Date(2026, 8, 12, 12, 0, 0, 0, time.UTC) }
	invalid := validImportFailure("bad slug")
	invalid.Sources[0].URL = "javascript:alert(1)"
	batch, err := service.CreateImportBatch(context.Background(), CreateImportBatchInput{
		AdminUserID: 42, BatchNo: "2026-08-12-am", Items: []ImportFailureInput{validImportFailure("valid-company"), invalid},
	})
	if err != nil {
		t.Fatalf("CreateImportBatch() error = %v", err)
	}
	if batch.SucceededCount != 1 || batch.FailedCount != 1 || batch.Status != ImportBatchReviewing || len(batch.ValidationErrors) < 1 || len(repository.items) != 1 {
		t.Fatalf("batch/items = %+v/%+v", batch, repository.items)
	}
}

func TestCreateImportBatchRejectsIneligibleEvidenceDomain(t *testing.T) {
	repository := &contentAdminMemoryRepository{memoryRepository: &memoryRepository{}, admin: true}
	service := NewService(repository, nil)
	item := validImportFailure("invalid-evidence")
	item.Sources[0].URL = "https://loot-drop.io/example"
	batch, err := service.CreateImportBatch(context.Background(), CreateImportBatchInput{AdminUserID: 42, BatchNo: "blocked-source", Items: []ImportFailureInput{item}})
	if err != nil {
		t.Fatalf("CreateImportBatch() error = %v", err)
	}
	if batch.SucceededCount != 0 || batch.FailedCount != 1 || len(repository.items) != 0 {
		t.Fatalf("batch/items = %+v/%+v", batch, repository.items)
	}
}

func TestProjectContentAdminRequiresAdminRole(t *testing.T) {
	repository := &contentAdminMemoryRepository{memoryRepository: &memoryRepository{}, admin: false}
	service := NewService(repository, nil)
	if _, err := service.CreateImportBatch(context.Background(), CreateImportBatchInput{AdminUserID: 42, BatchNo: "batch", Items: []ImportFailureInput{validImportFailure("valid")}}); err != ErrAdminRequired {
		t.Fatalf("CreateImportBatch() error = %v, want ErrAdminRequired", err)
	}
	if _, err := service.ListProjectOperationAudits(context.Background(), 42, 50); err != ErrAdminRequired {
		t.Fatalf("ListProjectOperationAudits() error = %v, want ErrAdminRequired", err)
	}
}

func TestImportBatchPublishAndRollbackUseAuditedStateChanges(t *testing.T) {
	repository := &contentAdminMemoryRepository{memoryRepository: &memoryRepository{}, admin: true}
	service := NewService(repository, nil)
	now := time.Date(2026, 8, 12, 12, 30, 0, 0, time.UTC)
	service.now = func() time.Time { return now }
	published, err := service.PublishImportBatch(context.Background(), 42, 81)
	if err != nil || published.Status != ImportBatchPublished || published.PublishedAt == nil {
		t.Fatalf("PublishImportBatch() = %+v/%v", published, err)
	}
	rolledBack, err := service.RollbackImportBatch(context.Background(), 42, 81)
	if err != nil || rolledBack.Status != ImportBatchRolledBack || rolledBack.RolledBackAt == nil {
		t.Fatalf("RollbackImportBatch() = %+v/%v", rolledBack, err)
	}
}
