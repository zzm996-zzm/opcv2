package enterprise

import (
	"context"
	"errors"
	"testing"

	"github.com/zzm/opcv2/internal/crm"
)

type fakeRepository struct {
	overview       Overview
	publicOverview PublicOverview
	publicCases    []PublicCase
	publicCase     PublicCase
	contactConfig  ContactConfig
	diagnosisInput DiagnosisRequestInput
	inquiryInput   InquiryInput
	updateInput    DiagnosisRequestUpdateInput
	diagnosis      DiagnosisRequest
	diagnoses      []DiagnosisRequest
	inquiry        Inquiry
	userID         int64
	requestID      int64
	slug           string
	ownerUserID    int64
	customerID     int64
	limit          int
	err            error
}

func (r *fakeRepository) Overview(_ context.Context, userID int64) (Overview, error) {
	r.userID = userID
	return r.overview, r.err
}

func (r *fakeRepository) PublicOverview(_ context.Context) (PublicOverview, error) {
	return r.publicOverview, r.err
}

func (r *fakeRepository) ListPublicCases(_ context.Context, limit int) ([]PublicCase, error) {
	r.limit = limit
	return r.publicCases, r.err
}

func (r *fakeRepository) GetPublicCase(_ context.Context, slug string) (PublicCase, error) {
	r.slug = slug
	return r.publicCase, r.err
}

func (r *fakeRepository) ContactConfig(_ context.Context) (ContactConfig, int64, error) {
	return r.contactConfig, r.ownerUserID, r.err
}

func (r *fakeRepository) CreateInquiry(_ context.Context, input InquiryInput) (Inquiry, error) {
	r.inquiryInput = input
	return r.inquiry, r.err
}

func (r *fakeRepository) UpdateInquiryCRMCustomer(_ context.Context, inquiryID int64, customerID int64) (Inquiry, error) {
	r.requestID = inquiryID
	r.customerID = customerID
	r.inquiry.CRMCustomerID = customerID
	r.inquiry.Status = "crm_synced"
	return r.inquiry, r.err
}

func (r *fakeRepository) CreateDiagnosisRequest(_ context.Context, userID int64, input DiagnosisRequestInput) (DiagnosisRequest, error) {
	r.userID = userID
	r.diagnosisInput = input
	return r.diagnosis, r.err
}

func (r *fakeRepository) ListDiagnosisRequests(_ context.Context, userID int64, limit int) ([]DiagnosisRequest, error) {
	r.userID = userID
	r.limit = limit
	return r.diagnoses, r.err
}

func (r *fakeRepository) GetDiagnosisRequest(_ context.Context, userID int64, requestID int64) (DiagnosisRequest, error) {
	r.userID = userID
	r.requestID = requestID
	return r.diagnosis, r.err
}

func (r *fakeRepository) UpdateDiagnosisRequest(_ context.Context, userID int64, requestID int64, input DiagnosisRequestUpdateInput) (DiagnosisRequest, error) {
	r.userID = userID
	r.requestID = requestID
	r.updateInput = input
	return r.diagnosis, r.err
}

type fakeCRMImporter struct {
	input        crm.ImportEnterpriseInput
	inquiryInput crm.ImportEnterpriseInquiryInput
	customer     crm.Customer
	err          error
}

func (i *fakeCRMImporter) ImportEnterpriseDelivery(_ context.Context, input crm.ImportEnterpriseInput) (crm.Customer, error) {
	i.input = input
	return i.customer, i.err
}

func (i *fakeCRMImporter) ImportEnterpriseInquiry(_ context.Context, input crm.ImportEnterpriseInquiryInput) (crm.Customer, error) {
	i.inquiryInput = input
	return i.customer, i.err
}

func TestServiceReturnsSafeEmptyOverviewWithoutRepository(t *testing.T) {
	service := NewService(nil)

	overview, err := service.Overview(context.Background(), 42)

	if err != nil {
		t.Fatalf("Overview() error = %v", err)
	}
	if overview.Stats == nil || overview.Plans == nil || overview.DeliveryBoard == nil || overview.Milestones == nil || overview.Cases == nil {
		t.Fatalf("overview should contain safe empty slices: %+v", overview)
	}
}

func TestServiceUsesRepository(t *testing.T) {
	repository := &fakeRepository{overview: Overview{Stats: []Metric{{Key: "companies", Label: "服务企业数", Value: "2"}}}}
	service := NewService(repository)

	overview, err := service.Overview(context.Background(), 42)

	if err != nil {
		t.Fatalf("Overview() error = %v", err)
	}
	if repository.userID != 42 || len(overview.Stats) != 1 {
		t.Fatalf("user/overview = %d/%+v", repository.userID, overview)
	}
}

func TestServiceRejectsMissingUserID(t *testing.T) {
	service := NewService(nil)

	_, err := service.Overview(context.Background(), 0)

	if !errors.Is(err, ErrUserIDRequired) {
		t.Fatalf("err = %v, want ErrUserIDRequired", err)
	}
}

func TestServiceReturnsPublishedPublicOverview(t *testing.T) {
	repository := &fakeRepository{publicOverview: PublicOverview{Headline: "企业AI落地陪跑"}}
	service := NewService(repository)

	overview, err := service.PublicOverview(context.Background())

	if err != nil {
		t.Fatalf("PublicOverview() error = %v", err)
	}
	if overview.Headline != "企业AI落地陪跑" || overview.ProofPoints == nil || overview.Stats == nil || overview.ServiceSteps == nil {
		t.Fatalf("overview = %+v", overview)
	}
}

func TestServiceListsPublicCases(t *testing.T) {
	repository := &fakeRepository{publicCases: []PublicCase{{Slug: "ai-sales", Title: "AI销售流程搭建"}}}
	service := NewService(repository)

	response, err := service.ListPublicCases(context.Background(), 500)

	if err != nil {
		t.Fatalf("ListPublicCases() error = %v", err)
	}
	if repository.limit != 20 || len(response.Cases) != 1 || response.Cases[0].Services == nil || response.Cases[0].Metrics == nil {
		t.Fatalf("repository/response = %+v/%+v", repository, response)
	}
}

func TestServiceGetsPublicCaseBySlug(t *testing.T) {
	repository := &fakeRepository{publicCase: PublicCase{Slug: "ai-sales", Title: "AI销售流程搭建"}}
	service := NewService(repository)

	item, err := service.GetPublicCase(context.Background(), " ai-sales ")

	if err != nil {
		t.Fatalf("GetPublicCase() error = %v", err)
	}
	if repository.slug != "ai-sales" || item.Slug != "ai-sales" {
		t.Fatalf("repository/item = %+v/%+v", repository, item)
	}
}

func TestServiceCreatesInquiryAndHandsOffToCRM(t *testing.T) {
	repository := &fakeRepository{
		ownerUserID: 99,
		inquiry:     Inquiry{ID: 10, Company: "启明星教育", Name: "张总", Phone: "13800138000", Need: "AI销售陪跑", Status: "submitted"},
	}
	importer := &fakeCRMImporter{customer: crm.Customer{ID: 300, Source: crm.SourceEnterprise, Stage: crm.StageNew}}
	service := NewService(repository, importer)

	inquiry, err := service.CreateInquiry(context.Background(), InquiryInput{
		Company: " 启明星教育 ",
		Name:    " 张总 ",
		Phone:   " 13800138000 ",
		Need:    " AI销售陪跑 ",
	})

	if err != nil {
		t.Fatalf("CreateInquiry() error = %v", err)
	}
	if repository.inquiryInput.Company != "启明星教育" || repository.inquiryInput.SourcePage != "/enterprise" {
		t.Fatalf("input = %+v", repository.inquiryInput)
	}
	if importer.inquiryInput.UserID != 99 || importer.inquiryInput.InquiryID != 10 || importer.inquiryInput.Company != "启明星教育" {
		t.Fatalf("crm input = %+v", importer.inquiryInput)
	}
	if inquiry.CRMCustomerID != 300 || inquiry.Status != "crm_synced" || repository.requestID != 10 || repository.customerID != 300 {
		t.Fatalf("inquiry/repository = %+v/%+v", inquiry, repository)
	}
}

func TestServiceCreatesInquiryWithoutCRMWhenOwnerMissing(t *testing.T) {
	repository := &fakeRepository{inquiry: Inquiry{ID: 10, Name: "张总", Wechat: "wx", Need: "AI销售陪跑", Status: "submitted"}}
	importer := &fakeCRMImporter{}
	service := NewService(repository, importer)

	inquiry, err := service.CreateInquiry(context.Background(), InquiryInput{Name: "张总", Wechat: "wx", Need: "AI销售陪跑"})

	if err != nil {
		t.Fatalf("CreateInquiry() error = %v", err)
	}
	if inquiry.Status != "submitted" || importer.inquiryInput.InquiryID != 0 {
		t.Fatalf("inquiry/importer = %+v/%+v", inquiry, importer.inquiryInput)
	}
}

func TestServiceRejectsInvalidInquiry(t *testing.T) {
	service := NewService(&fakeRepository{})

	_, err := service.CreateInquiry(context.Background(), InquiryInput{Name: "张总", Need: "AI销售陪跑", SourcePage: "//evil.example"})

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("err = %v, want ErrInvalidInput", err)
	}
}

func TestServiceCreatesDiagnosisRequest(t *testing.T) {
	repository := &fakeRepository{diagnosis: DiagnosisRequest{ID: 8, Status: "submitted"}}
	service := NewService(repository)

	request, err := service.CreateDiagnosisRequest(context.Background(), 42, DiagnosisRequestInput{Need: "  30人销售团队需要AI获客陪跑  "})

	if err != nil {
		t.Fatalf("CreateDiagnosisRequest() error = %v", err)
	}
	if request.ID != 8 || repository.userID != 42 || repository.diagnosisInput.Need != "30人销售团队需要AI获客陪跑" {
		t.Fatalf("request/repository = %+v/%+v", request, repository)
	}
}

func TestServiceRejectsInvalidDiagnosisRequest(t *testing.T) {
	service := NewService(&fakeRepository{})

	_, err := service.CreateDiagnosisRequest(context.Background(), 42, DiagnosisRequestInput{Need: "   "})

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("err = %v, want ErrInvalidInput", err)
	}
}

func TestServiceListsDiagnosisRequests(t *testing.T) {
	repository := &fakeRepository{diagnoses: []DiagnosisRequest{{ID: 7, Status: "submitted"}}}
	service := NewService(repository)

	response, err := service.ListDiagnosisRequests(context.Background(), 42, 5)

	if err != nil {
		t.Fatalf("ListDiagnosisRequests() error = %v", err)
	}
	if repository.userID != 42 || repository.limit != 5 || len(response.Requests) != 1 {
		t.Fatalf("repository/response = %+v/%+v", repository, response)
	}
}

func TestServiceDefaultsInvalidDiagnosisRequestLimit(t *testing.T) {
	repository := &fakeRepository{}
	service := NewService(repository)

	_, err := service.ListDiagnosisRequests(context.Background(), 42, 0)

	if err != nil {
		t.Fatalf("ListDiagnosisRequests() error = %v", err)
	}
	if repository.limit != 10 {
		t.Fatalf("limit = %d, want 10", repository.limit)
	}
}

func TestServiceUpdatesDiagnosisRequestStatus(t *testing.T) {
	repository := &fakeRepository{diagnosis: DiagnosisRequest{ID: 7, Status: "follow_up_created"}}
	service := NewService(repository)

	request, err := service.UpdateDiagnosisRequest(context.Background(), 42, 7, DiagnosisRequestUpdateInput{Status: " follow_up_created "})

	if err != nil {
		t.Fatalf("UpdateDiagnosisRequest() error = %v", err)
	}
	if request.Status != "follow_up_created" || repository.userID != 42 || repository.requestID != 7 || repository.updateInput.Status != "follow_up_created" {
		t.Fatalf("request/repository = %+v/%+v", request, repository)
	}
}

func TestServiceAllowsCompletedDiagnosisRequestStatus(t *testing.T) {
	repository := &fakeRepository{diagnosis: DiagnosisRequest{ID: 7, Status: "completed"}}
	service := NewService(repository)

	request, err := service.UpdateDiagnosisRequest(context.Background(), 42, 7, DiagnosisRequestUpdateInput{Status: "completed"})

	if err != nil {
		t.Fatalf("UpdateDiagnosisRequest() error = %v", err)
	}
	if request.Status != "completed" || repository.updateInput.Status != "completed" {
		t.Fatalf("request/repository = %+v/%+v", request, repository)
	}
}

func TestServiceRejectsInvalidDiagnosisRequestStatus(t *testing.T) {
	service := NewService(&fakeRepository{})

	_, err := service.UpdateDiagnosisRequest(context.Background(), 42, 7, DiagnosisRequestUpdateInput{Status: "submitted"})

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("err = %v, want ErrInvalidInput", err)
	}
}

func TestServiceImportsCompletedDiagnosisRequestToCRM(t *testing.T) {
	repository := &fakeRepository{diagnosis: DiagnosisRequest{ID: 7, UserID: 42, Need: "30人销售团队需要AI获客陪跑", Status: "completed"}}
	importer := &fakeCRMImporter{customer: crm.Customer{ID: 100, Name: "30人销售团队需要AI获客陪跑", Stage: crm.StageWon, Source: crm.SourceEnterprise}}
	service := NewService(repository, importer)

	customer, err := service.ImportDiagnosisRequestCustomer(context.Background(), 42, 7)

	if err != nil {
		t.Fatalf("ImportDiagnosisRequestCustomer() error = %v", err)
	}
	if customer.ID != 100 || repository.userID != 42 || repository.requestID != 7 {
		t.Fatalf("customer/repository = %+v/%+v", customer, repository)
	}
	if importer.input.UserID != 42 || importer.input.DiagnosisRequestID != 7 || importer.input.Need != "30人销售团队需要AI获客陪跑" {
		t.Fatalf("import input = %+v", importer.input)
	}
}

func TestServiceRejectsUncompletedDiagnosisRequestCRMImport(t *testing.T) {
	repository := &fakeRepository{diagnosis: DiagnosisRequest{ID: 7, UserID: 42, Need: "30人销售团队需要AI获客陪跑", Status: "in_delivery"}}
	importer := &fakeCRMImporter{}
	service := NewService(repository, importer)

	_, err := service.ImportDiagnosisRequestCustomer(context.Background(), 42, 7)

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("err = %v, want ErrInvalidInput", err)
	}
	if importer.input.UserID != 0 {
		t.Fatalf("importer should not be called: %+v", importer.input)
	}
}
