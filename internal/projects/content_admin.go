package projects

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"
)

const (
	ImportBatchCollecting = "collecting"
	ImportBatchReviewing  = "reviewing"
	ImportBatchPublished  = "published"
	ImportBatchRolledBack = "rolled_back"
)

var importSourceKinds = map[string]bool{
	"primary": true, "authority": true, "research": true, "media": true,
	"vertical": true, "community": true, "secondary": true,
}

type ImportSourceInput struct {
	FieldName string `json:"field_name"`
	URL       string `json:"url"`
	Name      string `json:"name,omitempty"`
	Kind      string `json:"kind"`
	Excerpt   string `json:"excerpt,omitempty"`
}

type ImportFailureInput struct {
	Slug              string              `json:"slug"`
	Name              string              `json:"name"`
	NameZH            string              `json:"name_zh,omitempty"`
	CountryCode       string              `json:"country_code,omitempty"`
	SectorCode        string              `json:"sector_code,omitempty"`
	ProductTypeCode   string              `json:"product_type_code,omitempty"`
	BurnedCents       *int64              `json:"burned_cents,omitempty"`
	Currency          string              `json:"currency,omitempty"`
	FoundedYear       *int                `json:"founded_year,omitempty"`
	DiedYear          *int                `json:"died_year,omitempty"`
	ValuePropZH       string              `json:"value_prop_zh"`
	DeathCauseZH      string              `json:"death_cause_zh"`
	FailureAnalysisZH string              `json:"failure_analysis_zh"`
	LearningsZH       []string            `json:"learnings_zh"`
	Sources           []ImportSourceInput `json:"sources"`
}

type CreateImportBatchInput struct {
	AdminUserID int64                `json:"-"`
	BatchNo     string               `json:"batch_no"`
	Planned     int                  `json:"planned_count,omitempty"`
	Items       []ImportFailureInput `json:"items"`
}

type ImportValidationError struct {
	Index int    `json:"index"`
	Slug  string `json:"slug,omitempty"`
	Field string `json:"field,omitempty"`
	Code  string `json:"code"`
}

type ImportBatch struct {
	ID               int64                   `json:"id"`
	BatchNo          string                  `json:"batch_no"`
	PlannedCount     int                     `json:"planned_count"`
	SucceededCount   int                     `json:"succeeded_count"`
	FailedCount      int                     `json:"failed_count"`
	Status           string                  `json:"status"`
	OperatorID       *int64                  `json:"operator_id,omitempty"`
	ValidationErrors []ImportValidationError `json:"validation_errors"`
	CreatedAt        time.Time               `json:"created_at"`
	UpdatedAt        time.Time               `json:"updated_at"`
	PublishedAt      *time.Time              `json:"published_at,omitempty"`
	RolledBackAt     *time.Time              `json:"rolled_back_at,omitempty"`
}

type OperationAudit struct {
	ID         int64          `json:"id"`
	OperatorID *int64         `json:"operator_id,omitempty"`
	Action     string         `json:"action"`
	TargetType string         `json:"target_type"`
	TargetID   string         `json:"target_id"`
	Detail     map[string]any `json:"detail"`
	CreatedAt  time.Time      `json:"created_at"`
}

type ProjectContentAdminRepository interface {
	IsProjectAdmin(context.Context, int64) (bool, error)
	CreateImportBatch(context.Context, ImportBatch, []ImportFailureInput) (ImportBatch, error)
	ListImportBatches(context.Context, int) ([]ImportBatch, error)
	GetImportBatch(context.Context, int64) (ImportBatch, error)
	PublishImportBatch(context.Context, int64, int64, time.Time) (ImportBatch, error)
	RollbackImportBatch(context.Context, int64, int64, time.Time) (ImportBatch, error)
	ListProjectOperationAudits(context.Context, int) ([]OperationAudit, error)
}

type ProjectContentAdminApplication interface {
	CreateImportBatch(context.Context, CreateImportBatchInput) (ImportBatch, error)
	ListImportBatches(context.Context, int64, int) ([]ImportBatch, error)
	GetImportBatch(context.Context, int64, int64) (ImportBatch, error)
	PublishImportBatch(context.Context, int64, int64) (ImportBatch, error)
	RollbackImportBatch(context.Context, int64, int64) (ImportBatch, error)
	ListProjectOperationAudits(context.Context, int64, int) ([]OperationAudit, error)
}

func (s *Service) contentAdminRepository() (ProjectContentAdminRepository, error) {
	repository, ok := s.repository.(ProjectContentAdminRepository)
	if !ok || repository == nil {
		return nil, ErrServiceNotReady
	}
	return repository, nil
}

func (s *Service) requireProjectAdmin(ctx context.Context, userID int64) (ProjectContentAdminRepository, error) {
	repository, err := s.contentAdminRepository()
	if err != nil {
		return nil, err
	}
	if userID <= 0 {
		return nil, ErrAdminRequired
	}
	ok, err := repository.IsProjectAdmin(ctx, userID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrAdminRequired
	}
	return repository, nil
}

func (s *Service) CreateImportBatch(ctx context.Context, input CreateImportBatchInput) (ImportBatch, error) {
	repository, err := s.requireProjectAdmin(ctx, input.AdminUserID)
	if err != nil {
		return ImportBatch{}, err
	}
	input.BatchNo = strings.TrimSpace(input.BatchNo)
	if input.BatchNo == "" || len(input.BatchNo) > 100 || len(input.Items) == 0 || len(input.Items) > 100 {
		return ImportBatch{}, ErrInvalidImportBatch
	}
	if input.Planned == 0 {
		input.Planned = len(input.Items)
	}
	if input.Planned < len(input.Items) || input.Planned > 1000 {
		return ImportBatch{}, ErrInvalidImportBatch
	}
	validItems := make([]ImportFailureInput, 0, len(input.Items))
	errorsList := make([]ImportValidationError, 0)
	seen := map[string]bool{}
	for index, raw := range input.Items {
		item, validation := normalizeImportFailure(index, raw)
		if seen[item.Slug] && item.Slug != "" {
			validation = append(validation, ImportValidationError{Index: index, Slug: item.Slug, Field: "slug", Code: "duplicate_slug_in_batch"})
		}
		seen[item.Slug] = true
		if len(validation) > 0 {
			errorsList = append(errorsList, validation...)
			continue
		}
		validItems = append(validItems, item)
	}
	operatorID := input.AdminUserID
	now := s.now()
	batch := ImportBatch{
		BatchNo: input.BatchNo, PlannedCount: input.Planned, SucceededCount: len(validItems),
		FailedCount: len(input.Items) - len(validItems), Status: ImportBatchReviewing,
		OperatorID: &operatorID, ValidationErrors: errorsList, CreatedAt: now, UpdatedAt: now,
	}
	return repository.CreateImportBatch(ctx, batch, validItems)
}

func normalizeImportFailure(index int, input ImportFailureInput) (ImportFailureInput, []ImportValidationError) {
	input.Slug = strings.ToLower(strings.TrimSpace(input.Slug))
	input.Name = strings.TrimSpace(input.Name)
	input.NameZH = strings.TrimSpace(input.NameZH)
	input.CountryCode = strings.TrimSpace(input.CountryCode)
	input.SectorCode = strings.TrimSpace(input.SectorCode)
	input.ProductTypeCode = strings.TrimSpace(input.ProductTypeCode)
	input.Currency = strings.ToUpper(strings.TrimSpace(input.Currency))
	input.ValuePropZH = strings.TrimSpace(input.ValuePropZH)
	input.DeathCauseZH = strings.TrimSpace(input.DeathCauseZH)
	input.FailureAnalysisZH = strings.TrimSpace(input.FailureAnalysisZH)
	if input.Currency == "" {
		input.Currency = "USD"
	}
	fail := func(field, code string) ImportValidationError {
		return ImportValidationError{Index: index, Slug: input.Slug, Field: field, Code: code}
	}
	validation := make([]ImportValidationError, 0)
	if input.Slug == "" || len(input.Slug) > 160 {
		validation = append(validation, fail("slug", "required_or_too_long"))
	}
	for _, char := range input.Slug {
		if !((char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '-') {
			validation = append(validation, fail("slug", "invalid_format"))
			break
		}
	}
	if input.Name == "" || input.ValuePropZH == "" || input.DeathCauseZH == "" || input.FailureAnalysisZH == "" || len(input.LearningsZH) == 0 {
		validation = append(validation, fail("content", "required_fields_missing"))
	}
	if len(input.Currency) != 3 || input.BurnedCents != nil && *input.BurnedCents < 0 || input.FoundedYear != nil && input.DiedYear != nil && *input.DiedYear < *input.FoundedYear {
		validation = append(validation, fail("timeline_or_money", "invalid_value"))
	}
	if len(input.Sources) == 0 {
		validation = append(validation, fail("sources", "source_required"))
	}
	for sourceIndex := range input.Sources {
		source := &input.Sources[sourceIndex]
		source.FieldName = strings.TrimSpace(source.FieldName)
		source.URL = strings.TrimSpace(source.URL)
		source.Name = strings.TrimSpace(source.Name)
		source.Kind = strings.TrimSpace(source.Kind)
		source.Excerpt = strings.TrimSpace(source.Excerpt)
		parsed, err := url.Parse(source.URL)
		host := ""
		if parsed != nil {
			host = strings.ToLower(parsed.Hostname())
		}
		if source.FieldName == "" || err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || host == "" || host == "loot-drop.io" || strings.HasSuffix(host, ".loot-drop.io") || !importSourceKinds[source.Kind] {
			validation = append(validation, fail(fmt.Sprintf("sources.%d", sourceIndex), "invalid_source"))
		}
	}
	return input, validation
}

func (s *Service) ListImportBatches(ctx context.Context, adminUserID int64, limit int) ([]ImportBatch, error) {
	repository, err := s.requireProjectAdmin(ctx, adminUserID)
	if err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	items, err := repository.ListImportBatches(ctx, limit)
	if items == nil {
		items = []ImportBatch{}
	}
	return items, err
}

func (s *Service) GetImportBatch(ctx context.Context, adminUserID, id int64) (ImportBatch, error) {
	repository, err := s.requireProjectAdmin(ctx, adminUserID)
	if err != nil {
		return ImportBatch{}, err
	}
	if id <= 0 {
		return ImportBatch{}, ErrImportBatchNotFound
	}
	return repository.GetImportBatch(ctx, id)
}

func (s *Service) PublishImportBatch(ctx context.Context, adminUserID, id int64) (ImportBatch, error) {
	repository, err := s.requireProjectAdmin(ctx, adminUserID)
	if err != nil {
		return ImportBatch{}, err
	}
	return repository.PublishImportBatch(ctx, id, adminUserID, s.now())
}

func (s *Service) RollbackImportBatch(ctx context.Context, adminUserID, id int64) (ImportBatch, error) {
	repository, err := s.requireProjectAdmin(ctx, adminUserID)
	if err != nil {
		return ImportBatch{}, err
	}
	return repository.RollbackImportBatch(ctx, id, adminUserID, s.now())
}

func (s *Service) ListProjectOperationAudits(ctx context.Context, adminUserID int64, limit int) ([]OperationAudit, error) {
	repository, err := s.requireProjectAdmin(ctx, adminUserID)
	if err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	items, err := repository.ListProjectOperationAudits(ctx, limit)
	if items == nil {
		items = []OperationAudit{}
	}
	return items, err
}
