package enterprise

import (
	"context"
	"errors"
	"strings"

	"github.com/zzm/opcv2/internal/crm"
)

var ErrUserIDRequired = errors.New("user id required")
var ErrInvalidInput = errors.New("invalid input")
var ErrDiagnosisRequestNotFound = errors.New("enterprise diagnosis request not found")
var ErrPublicCaseNotFound = errors.New("enterprise public case not found")

type Repository interface {
	Overview(ctx context.Context, userID int64) (Overview, error)
	PublicOverview(ctx context.Context) (PublicOverview, error)
	ListPublicCases(ctx context.Context, limit int) ([]PublicCase, error)
	GetPublicCase(ctx context.Context, slug string) (PublicCase, error)
	ContactConfig(ctx context.Context) (ContactConfig, int64, error)
	CreateInquiry(ctx context.Context, input InquiryInput) (Inquiry, error)
	UpdateInquiryCRMCustomer(ctx context.Context, inquiryID int64, customerID int64) (Inquiry, error)
	CreateDiagnosisRequest(ctx context.Context, userID int64, input DiagnosisRequestInput) (DiagnosisRequest, error)
	ListDiagnosisRequests(ctx context.Context, userID int64, limit int) ([]DiagnosisRequest, error)
	GetDiagnosisRequest(ctx context.Context, userID int64, requestID int64) (DiagnosisRequest, error)
	UpdateDiagnosisRequest(ctx context.Context, userID int64, requestID int64, input DiagnosisRequestUpdateInput) (DiagnosisRequest, error)
}

type CRMImporter interface {
	ImportEnterpriseDelivery(ctx context.Context, input crm.ImportEnterpriseInput) (crm.Customer, error)
	ImportEnterpriseInquiry(ctx context.Context, input crm.ImportEnterpriseInquiryInput) (crm.Customer, error)
}

type Service struct {
	repository Repository
	crm        CRMImporter
}

func NewService(repository Repository, crmImporters ...CRMImporter) *Service {
	var crmImporter CRMImporter
	if len(crmImporters) > 0 {
		crmImporter = crmImporters[0]
	}
	return &Service{repository: repository, crm: crmImporter}
}

func (s *Service) Overview(ctx context.Context, userID int64) (Overview, error) {
	if userID <= 0 {
		return Overview{}, ErrUserIDRequired
	}
	if s.repository == nil {
		return emptyOverview(), nil
	}
	overview, err := s.repository.Overview(ctx, userID)
	if err != nil {
		return Overview{}, err
	}
	return ensureOverviewSlices(overview), nil
}

func (s *Service) PublicOverview(ctx context.Context) (PublicOverview, error) {
	if s.repository == nil {
		return emptyPublicOverview(), nil
	}
	overview, err := s.repository.PublicOverview(ctx)
	if err != nil {
		return PublicOverview{}, err
	}
	return ensurePublicOverviewSlices(overview), nil
}

func (s *Service) ListPublicCases(ctx context.Context, limit int) (PublicCasesResponse, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	if s.repository == nil {
		return PublicCasesResponse{Cases: []PublicCase{}}, nil
	}
	cases, err := s.repository.ListPublicCases(ctx, limit)
	if err != nil {
		return PublicCasesResponse{}, err
	}
	if cases == nil {
		cases = []PublicCase{}
	}
	for index := range cases {
		cases[index] = ensurePublicCaseSlices(cases[index])
	}
	return PublicCasesResponse{Cases: cases}, nil
}

func (s *Service) GetPublicCase(ctx context.Context, slug string) (PublicCase, error) {
	slug = strings.TrimSpace(slug)
	if slug == "" {
		return PublicCase{}, ErrInvalidInput
	}
	if s.repository == nil {
		return PublicCase{}, ErrPublicCaseNotFound
	}
	item, err := s.repository.GetPublicCase(ctx, slug)
	if err != nil {
		return PublicCase{}, err
	}
	return ensurePublicCaseSlices(item), nil
}

func (s *Service) ContactConfig(ctx context.Context) (ContactConfig, error) {
	if s.repository == nil {
		return ContactConfig{}, nil
	}
	config, _, err := s.repository.ContactConfig(ctx)
	return config, err
}

func (s *Service) CreateInquiry(ctx context.Context, input InquiryInput) (Inquiry, error) {
	if s.repository == nil {
		return Inquiry{}, ErrInvalidInput
	}
	input.Company = strings.TrimSpace(input.Company)
	input.Name = strings.TrimSpace(input.Name)
	input.Phone = strings.TrimSpace(input.Phone)
	input.Email = strings.TrimSpace(input.Email)
	input.Wechat = strings.TrimSpace(input.Wechat)
	input.Need = strings.TrimSpace(input.Need)
	input.Budget = strings.TrimSpace(input.Budget)
	input.Timeline = strings.TrimSpace(input.Timeline)
	input.SourcePage = strings.TrimSpace(input.SourcePage)
	if input.Name == "" || input.Need == "" || (input.Phone == "" && input.Email == "" && input.Wechat == "") {
		return Inquiry{}, ErrInvalidInput
	}
	if input.SourcePage == "" {
		input.SourcePage = "/enterprise"
	}
	if !strings.HasPrefix(input.SourcePage, "/") || strings.HasPrefix(input.SourcePage, "//") {
		return Inquiry{}, ErrInvalidInput
	}
	inquiry, err := s.repository.CreateInquiry(ctx, input)
	if err != nil {
		return Inquiry{}, err
	}
	if s.crm == nil {
		return inquiry, nil
	}
	_, ownerUserID, err := s.repository.ContactConfig(ctx)
	if err != nil {
		return Inquiry{}, err
	}
	if ownerUserID <= 0 {
		return inquiry, nil
	}
	customer, err := s.crm.ImportEnterpriseInquiry(ctx, crm.ImportEnterpriseInquiryInput{
		UserID:    ownerUserID,
		InquiryID: inquiry.ID,
		Company:   inquiry.Company,
		Name:      inquiry.Name,
		Phone:     inquiry.Phone,
		Email:     inquiry.Email,
		Need:      inquiry.Need,
	})
	if err != nil {
		return Inquiry{}, err
	}
	return s.repository.UpdateInquiryCRMCustomer(ctx, inquiry.ID, customer.ID)
}

func (s *Service) CreateDiagnosisRequest(ctx context.Context, userID int64, input DiagnosisRequestInput) (DiagnosisRequest, error) {
	if userID <= 0 {
		return DiagnosisRequest{}, ErrUserIDRequired
	}
	if s.repository == nil {
		return DiagnosisRequest{}, ErrInvalidInput
	}
	input.Need = strings.TrimSpace(input.Need)
	if input.Need == "" {
		return DiagnosisRequest{}, ErrInvalidInput
	}
	return s.repository.CreateDiagnosisRequest(ctx, userID, input)
}

func (s *Service) ListDiagnosisRequests(ctx context.Context, userID int64, limit int) (DiagnosisRequestsResponse, error) {
	if userID <= 0 {
		return DiagnosisRequestsResponse{}, ErrUserIDRequired
	}
	if limit <= 0 || limit > 50 {
		limit = 10
	}
	if s.repository == nil {
		return DiagnosisRequestsResponse{Requests: []DiagnosisRequest{}}, nil
	}
	requests, err := s.repository.ListDiagnosisRequests(ctx, userID, limit)
	if err != nil {
		return DiagnosisRequestsResponse{}, err
	}
	if requests == nil {
		requests = []DiagnosisRequest{}
	}
	return DiagnosisRequestsResponse{Requests: requests}, nil
}

func (s *Service) UpdateDiagnosisRequest(ctx context.Context, userID int64, requestID int64, input DiagnosisRequestUpdateInput) (DiagnosisRequest, error) {
	if userID <= 0 {
		return DiagnosisRequest{}, ErrUserIDRequired
	}
	if s.repository == nil || requestID <= 0 {
		return DiagnosisRequest{}, ErrInvalidInput
	}
	input.Status = strings.TrimSpace(input.Status)
	if input.Status != "follow_up_created" && input.Status != "in_delivery" && input.Status != "completed" {
		return DiagnosisRequest{}, ErrInvalidInput
	}
	return s.repository.UpdateDiagnosisRequest(ctx, userID, requestID, input)
}

func (s *Service) ImportDiagnosisRequestCustomer(ctx context.Context, userID int64, requestID int64) (crm.Customer, error) {
	if userID <= 0 {
		return crm.Customer{}, ErrUserIDRequired
	}
	if s.repository == nil || s.crm == nil || requestID <= 0 {
		return crm.Customer{}, ErrInvalidInput
	}
	request, err := s.repository.GetDiagnosisRequest(ctx, userID, requestID)
	if err != nil {
		return crm.Customer{}, err
	}
	if request.Status != "completed" {
		return crm.Customer{}, ErrInvalidInput
	}
	return s.crm.ImportEnterpriseDelivery(ctx, crm.ImportEnterpriseInput{
		UserID:             userID,
		DiagnosisRequestID: request.ID,
		Need:               request.Need,
	})
}

func emptyOverview() Overview {
	return Overview{
		Stats:         []Metric{},
		Plans:         []Plan{},
		DeliveryBoard: []DeliveryItem{},
		Milestones:    []Milestone{},
		Cases:         []Case{},
	}
}

func emptyPublicOverview() PublicOverview {
	return PublicOverview{
		ProofPoints:  []PublicProofPoint{},
		Stats:        []PublicStat{},
		ServiceSteps: []PublicServiceStep{},
	}
}

func ensureOverviewSlices(overview Overview) Overview {
	if overview.Stats == nil {
		overview.Stats = []Metric{}
	}
	if overview.Plans == nil {
		overview.Plans = []Plan{}
	}
	for index := range overview.Plans {
		if overview.Plans[index].Focus == nil {
			overview.Plans[index].Focus = []string{}
		}
	}
	if overview.DeliveryBoard == nil {
		overview.DeliveryBoard = []DeliveryItem{}
	}
	if overview.Milestones == nil {
		overview.Milestones = []Milestone{}
	}
	if overview.Cases == nil {
		overview.Cases = []Case{}
	}
	return overview
}

func ensurePublicOverviewSlices(overview PublicOverview) PublicOverview {
	if overview.ProofPoints == nil {
		overview.ProofPoints = []PublicProofPoint{}
	}
	if overview.Stats == nil {
		overview.Stats = []PublicStat{}
	}
	if overview.ServiceSteps == nil {
		overview.ServiceSteps = []PublicServiceStep{}
	}
	return overview
}

func ensurePublicCaseSlices(item PublicCase) PublicCase {
	if item.Services == nil {
		item.Services = []string{}
	}
	if item.Metrics == nil {
		item.Metrics = []PublicCaseMetric{}
	}
	return item
}
