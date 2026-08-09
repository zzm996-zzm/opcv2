package projects

import (
	"context"
	"strings"
	"time"
)

const EvidenceStatusVerified = "verified"

type EvidenceCaseFilters struct {
	CaseType string
	Industry string
	Scale    string
	Page     int
	PageSize int
}

type EvidenceCaseItem struct {
	ID                     int64      `json:"id"`
	Title                  string     `json:"title"`
	CoverURL               string     `json:"cover_url,omitempty"`
	ResultSummary          string     `json:"result_summary"`
	Industry               string     `json:"industry,omitempty"`
	Scale                  string     `json:"scale,omitempty"`
	Type                   string     `json:"type"`
	PrimarySourceURL       string     `json:"primary_source_url"`
	SourceCount            int        `json:"source_count"`
	VerifiedAt             *time.Time `json:"verified_at,omitempty"`
	PublishedAt            *time.Time `json:"published_at,omitempty"`
	ProjectID              *int64     `json:"project_id,omitempty"`
	EvidenceStatus         string     `json:"-"`
	HasConflict            bool       `json:"-"`
	PrimarySourceCount     int        `json:"-"`
	ReliableSecondaryCount int        `json:"-"`
}

type EvidenceCasePage struct {
	Items    []EvidenceCaseItem `json:"items"`
	Page     int                `json:"page"`
	PageSize int                `json:"page_size"`
	Total    int                `json:"total"`
}

type EvidenceSource struct {
	ID          int64      `json:"id"`
	Title       string     `json:"title,omitempty"`
	Publisher   string     `json:"publisher,omitempty"`
	URL         string     `json:"url"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
	FetchedAt   time.Time  `json:"fetched_at"`
	Quality     *float64   `json:"quality,omitempty"`
	Kind        string     `json:"kind"`
	IsPrimary   bool       `json:"is_primary"`
	ClaimFields []string   `json:"claim_fields"`
}

type EvidenceFact struct {
	Field      string  `json:"field"`
	Value      string  `json:"value"`
	SourceRefs []int64 `json:"source_refs"`
}

type EvidenceAnalysis struct {
	Point            string  `json:"point"`
	Detail           string  `json:"detail,omitempty"`
	IsModelGenerated bool    `json:"is_model_generated"`
	SourceRefs       []int64 `json:"source_refs"`
}

type EvidenceCaseDetail struct {
	EvidenceCaseItem
	ContentMD string             `json:"content_md"`
	Facts     []EvidenceFact     `json:"facts"`
	Analyses  []EvidenceAnalysis `json:"analyses"`
	Sources   []EvidenceSource   `json:"sources"`
}

type EvidenceCaseRepository interface {
	ListEvidenceCases(ctx context.Context, filters EvidenceCaseFilters) (EvidenceCasePage, error)
	GetEvidenceCase(ctx context.Context, ref string) (EvidenceCaseDetail, error)
}

type EvidenceCaseApplication interface {
	ListEvidenceCases(context.Context, EvidenceCaseFilters) (EvidenceCasePage, error)
	GetEvidenceCase(context.Context, string) (EvidenceCaseDetail, error)
}

func (s *Service) ListEvidenceCases(ctx context.Context, filters EvidenceCaseFilters) (EvidenceCasePage, error) {
	repository, ok := s.repository.(EvidenceCaseRepository)
	if !ok {
		return EvidenceCasePage{}, ErrServiceNotReady
	}
	filters.CaseType = normalizeEvidenceCaseType(filters.CaseType)
	filters.Industry = strings.TrimSpace(filters.Industry)
	filters.Scale = strings.TrimSpace(filters.Scale)
	if filters.Page <= 0 {
		filters.Page = 1
	}
	if filters.PageSize <= 0 {
		filters.PageSize = 20
	}
	if filters.PageSize > 100 {
		filters.PageSize = 100
	}
	page, err := repository.ListEvidenceCases(ctx, filters)
	if err != nil {
		return EvidenceCasePage{}, err
	}
	page.Page = filters.Page
	page.PageSize = filters.PageSize
	if page.Items == nil {
		page.Items = []EvidenceCaseItem{}
	}
	items := make([]EvidenceCaseItem, 0, len(page.Items))
	for _, item := range page.Items {
		if normalizeEvidenceCaseItem(&item) {
			items = append(items, item)
		}
	}
	page.Items = items
	return page, nil
}

func (s *Service) GetEvidenceCase(ctx context.Context, ref string) (EvidenceCaseDetail, error) {
	repository, ok := s.repository.(EvidenceCaseRepository)
	if !ok {
		return EvidenceCaseDetail{}, ErrServiceNotReady
	}
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return EvidenceCaseDetail{}, ErrCaseNotFound
	}
	detail, err := repository.GetEvidenceCase(ctx, ref)
	if err != nil {
		return EvidenceCaseDetail{}, err
	}
	if !normalizeEvidenceCaseItem(&detail.EvidenceCaseItem) {
		return EvidenceCaseDetail{}, ErrCaseNotFound
	}
	if detail.Facts == nil {
		detail.Facts = []EvidenceFact{}
	}
	if detail.Analyses == nil {
		detail.Analyses = []EvidenceAnalysis{}
	}
	if detail.Sources == nil {
		detail.Sources = []EvidenceSource{}
	}
	verifiedFacts := make([]EvidenceFact, 0, len(detail.Facts))
	for _, fact := range detail.Facts {
		if len(fact.SourceRefs) > 0 {
			verifiedFacts = append(verifiedFacts, fact)
		}
	}
	detail.Facts = verifiedFacts
	for index := range detail.Analyses {
		detail.Analyses[index].IsModelGenerated = true
	}
	validSources := make([]EvidenceSource, 0, len(detail.Sources))
	for index := range detail.Sources {
		detail.Sources[index].URL = eligibleProjectSourceURL(detail.Sources[index].URL)
		if detail.Sources[index].ClaimFields == nil {
			detail.Sources[index].ClaimFields = []string{}
		}
		if detail.Sources[index].URL != "" {
			validSources = append(validSources, detail.Sources[index])
		}
	}
	detail.Sources = validSources
	return detail, nil
}

func normalizeEvidenceCaseType(value string) string {
	switch strings.TrimSpace(value) {
	case "fail", "failure":
		return "failure"
	case "success":
		return "success"
	default:
		return ""
	}
}

func normalizeEvidenceCaseItem(item *EvidenceCaseItem) bool {
	if item.EvidenceStatus != EvidenceStatusVerified || item.HasConflict {
		return false
	}
	if item.PrimarySourceCount < 1 && item.ReliableSecondaryCount < 2 {
		return false
	}
	item.PrimarySourceURL = eligibleProjectSourceURL(item.PrimarySourceURL)
	if item.PrimarySourceURL == "" {
		return false
	}
	if item.Type == "failure" {
		item.Type = "fail"
	}
	return true
}
