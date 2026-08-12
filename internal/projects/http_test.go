package projects

import (
	"bytes"
	"context"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zzm/opcv2/internal/auth"
	projectfiles "github.com/zzm/opcv2/internal/projects/files"
)

type fakeApplication struct {
	input              MatchInput
	userID             int64
	matchID            int64
	result             MatchResult
	sessions           []MatchSession
	session            MatchSession
	favorite           Favorite
	favorites          []Favorite
	err                error
	opportunities      []Opportunity
	opportunity        Opportunity
	filters            OpportunityFilters
	cases              []CaseStudy
	caseStudy          CaseStudy
	caseFilters        CaseFilters
	publicConfig       PublicConfig
	dictionaries       []DictionaryItem
	home               ProjectHome
	projectPage        ProjectPage
	catalogProject     Project
	projectFilters     ProjectFilters
	evidenceCasePage   EvidenceCasePage
	evidenceCase       EvidenceCaseDetail
	evidenceFilters    EvidenceCaseFilters
	workflowCreate     CreateProjectMatchInput
	workflowAnswer     AnswerProjectMatchInput
	workflowResponse   MatchWorkflowResponse
	generationResponse MatchGenerationResponse
	progressEvents     []MatchProgressEvent
	afterEventID       int64
	export             Export
	exportInput        CreateExportInput
	analyticsInput     AnalyticsEventInput
	analyticsReceipt   AnalyticsReceipt
	importBatch        ImportBatch
	importInput        CreateImportBatchInput
	operationAudits    []OperationAudit
}

func (a *fakeApplication) RecordProjectEvent(_ context.Context, input AnalyticsEventInput) (AnalyticsReceipt, error) {
	a.analyticsInput = input
	return a.analyticsReceipt, a.err
}

func (a *fakeApplication) CreateImportBatch(_ context.Context, input CreateImportBatchInput) (ImportBatch, error) {
	a.importInput = input
	return a.importBatch, a.err
}
func (a *fakeApplication) ListImportBatches(context.Context, int64, int) ([]ImportBatch, error) {
	return []ImportBatch{a.importBatch}, a.err
}
func (a *fakeApplication) GetImportBatch(context.Context, int64, int64) (ImportBatch, error) {
	return a.importBatch, a.err
}
func (a *fakeApplication) PublishImportBatch(context.Context, int64, int64) (ImportBatch, error) {
	return a.importBatch, a.err
}
func (a *fakeApplication) RollbackImportBatch(context.Context, int64, int64) (ImportBatch, error) {
	return a.importBatch, a.err
}
func (a *fakeApplication) ListProjectOperationAudits(context.Context, int64, int) ([]OperationAudit, error) {
	return a.operationAudits, a.err
}

func (a *fakeApplication) GetPublicConfig(context.Context) PublicConfig { return a.publicConfig }
func (a *fakeApplication) ListDictionaryItems(_ context.Context, _ string) ([]DictionaryItem, error) {
	return a.dictionaries, a.err
}
func (a *fakeApplication) GetProjectHome(context.Context) (ProjectHome, error) { return a.home, a.err }
func (a *fakeApplication) ListProjects(_ context.Context, filters ProjectFilters) (ProjectPage, error) {
	a.projectFilters = filters
	return a.projectPage, a.err
}
func (a *fakeApplication) GetProject(context.Context, string) (Project, error) {
	return a.catalogProject, a.err
}
func (a *fakeApplication) ListEvidenceCases(_ context.Context, filters EvidenceCaseFilters) (EvidenceCasePage, error) {
	a.evidenceFilters = filters
	return a.evidenceCasePage, a.err
}
func (a *fakeApplication) GetEvidenceCase(context.Context, string) (EvidenceCaseDetail, error) {
	return a.evidenceCase, a.err
}

func (a *fakeApplication) ListCases(_ context.Context, filters CaseFilters) ([]CaseStudy, error) {
	a.caseFilters = filters
	return a.cases, a.err
}
func (a *fakeApplication) GetCase(_ context.Context, _ string) (CaseStudy, error) {
	return a.caseStudy, a.err
}

func (a *fakeApplication) ListOpportunities(_ context.Context, filters OpportunityFilters) ([]Opportunity, error) {
	a.filters = filters
	return a.opportunities, a.err
}
func (a *fakeApplication) GetOpportunity(_ context.Context, _ string) (Opportunity, error) {
	return a.opportunity, a.err
}

func (a *fakeApplication) CreateMatch(_ context.Context, input MatchInput) (MatchResult, error) {
	a.input = input
	return a.result, a.err
}

func (a *fakeApplication) ListMatches(_ context.Context, userID int64, limit int) ([]MatchSession, error) {
	a.userID = userID
	return a.sessions, a.err
}

func (a *fakeApplication) GetMatch(_ context.Context, userID, id int64) (MatchSession, error) {
	a.userID = userID
	a.matchID = id
	return a.session, a.err
}
func (a *fakeApplication) AnswerMatch(_ context.Context, input AnswerMatchInput) (MatchResult, error) {
	a.input.UserID = input.UserID
	a.matchID = input.SessionID
	return a.result, a.err
}
func (a *fakeApplication) CreateComparison(_ context.Context, input CreateComparisonInput) (Comparison, error) {
	return Comparison{ID: 61, UserID: input.UserID}, a.err
}
func (a *fakeApplication) GetComparison(_ context.Context, userID, id int64) (Comparison, error) {
	return Comparison{ID: id, UserID: userID}, a.err
}
func (a *fakeApplication) CreateExport(_ context.Context, input CreateExportInput) (Export, error) {
	a.exportInput = input
	if a.export.ID == 0 {
		return Export{ID: 71, UserID: input.UserID, Status: "queued", Format: "pdf"}, a.err
	}
	return a.export, a.err
}
func (a *fakeApplication) GetExport(_ context.Context, userID, id int64) (Export, error) {
	if a.export.ID == 0 {
		return Export{ID: id, UserID: userID, Status: "ready", Format: "json", Payload: []byte(`{}`)}, a.err
	}
	return a.export, a.err
}

func (a *fakeApplication) FavoriteMatch(_ context.Context, userID, id int64) (Favorite, error) {
	a.userID = userID
	a.matchID = id
	return a.favorite, a.err
}

func (a *fakeApplication) ListFavoriteMatches(_ context.Context, userID int64, _ int) ([]Favorite, error) {
	a.userID = userID
	return a.favorites, a.err
}

func (a *fakeApplication) UnfavoriteMatch(_ context.Context, userID, id int64) error {
	a.userID = userID
	a.matchID = id
	return a.err
}

func (a *fakeApplication) CreateProjectMatch(_ context.Context, input CreateProjectMatchInput) (MatchWorkflowResponse, error) {
	a.workflowCreate = input
	return a.workflowResponse, a.err
}

func (a *fakeApplication) AnswerProjectMatch(_ context.Context, input AnswerProjectMatchInput) (MatchWorkflowResponse, error) {
	a.workflowAnswer = input
	return a.workflowResponse, a.err
}

func (a *fakeApplication) GetProjectMatch(_ context.Context, userID, id int64) (MatchWorkflowResponse, error) {
	a.userID, a.matchID = userID, id
	return a.workflowResponse, a.err
}

func (a *fakeApplication) GenerateProjectMatch(_ context.Context, userID, id int64) (MatchGenerationResponse, error) {
	a.userID, a.matchID = userID, id
	return a.generationResponse, a.err
}

func (a *fakeApplication) CancelProjectMatch(_ context.Context, userID, id int64) (MatchGenerationResponse, error) {
	a.userID, a.matchID = userID, id
	return a.generationResponse, a.err
}

func (a *fakeApplication) ListProjectMatchProgress(_ context.Context, userID, id, afterID int64) ([]MatchProgressEvent, error) {
	a.userID, a.matchID, a.afterEventID = userID, id, afterID
	return a.progressEvents, a.err
}

func projectTestRouter(app Application) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	group := router.Group("/api/v1")
	group.Use(func(c *gin.Context) {
		c.Set(auth.UserIDContextKey, int64(42))
		c.Next()
	})
	NewHTTPHandler(app).Register(group)
	return router
}

func projectAdminTestRouter(app Application) *gin.Engine {
	router := projectTestRouter(app)
	group := router.Group("/api/v1")
	group.Use(func(c *gin.Context) {
		c.Set(auth.UserIDContextKey, int64(42))
		c.Next()
	})
	NewHTTPHandler(app).RegisterAdmin(group)
	return router
}

func TestCreateMatchEndpointUsesAuthenticatedUser(t *testing.T) {
	app := &fakeApplication{result: MatchResult{
		SessionID: 99,
		Status:    StatusCompleted,
		Projects:  []ProjectMatch{{Title: "AI短视频脚本工作室", Score: 94}},
	}}
	router := projectTestRouter(app)
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/projects/matches",
		strings.NewReader(`{"intent":"我擅长内容创作，预算3万以内，每周20小时"}`),
	)
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.input.UserID != 42 || app.input.Intent == "" {
		t.Fatalf("input = %+v", app.input)
	}
	if !strings.Contains(recorder.Body.String(), `"title":"AI短视频脚本工作室"`) {
		t.Fatalf("body = %s", recorder.Body.String())
	}
}

func TestProjectAnalyticsEndpointAcceptsWhitelistedEvent(t *testing.T) {
	app := &fakeApplication{analyticsReceipt: AnalyticsReceipt{EventID: "event-12345678", Accepted: true}}
	router := projectTestRouter(app)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/analytics/events", strings.NewReader(`{
		"event_id":"event-12345678","event_name":"project_detail_view","visitor_key":"visitor-12345",
		"route":"/projects/42","ref_module":"catalog","properties":{"project_id":42,"tab":"overview"}
	}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusAccepted || app.analyticsInput.EventName != ProjectEventDetailView {
		t.Fatalf("status=%d input=%+v body=%s", recorder.Code, app.analyticsInput, recorder.Body.String())
	}
}

func TestProjectAnalyticsEndpointRejectsMalformedPayload(t *testing.T) {
	router := projectTestRouter(&fakeApplication{})
	request := httptest.NewRequest(http.MethodPost, "/api/v1/analytics/events", strings.NewReader(`{"event_name":`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusBadRequest || !strings.Contains(recorder.Body.String(), "invalid_event") {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestProjectAnalyticsEndpointRejectsClientIdentityFields(t *testing.T) {
	router := projectTestRouter(&fakeApplication{})
	request := httptest.NewRequest(http.MethodPost, "/api/v1/analytics/events", strings.NewReader(`{
		"event_id":"event-12345678","event_name":"project_detail_view","visitor_key":"visitor-12345",
		"route":"/projects/42","user_id":99,"properties":{"project_id":42}
	}`))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusBadRequest || !strings.Contains(recorder.Body.String(), "invalid_event") {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestCreateProjectImportBatchEndpointUsesAuthenticatedAdmin(t *testing.T) {
	app := &fakeApplication{importBatch: ImportBatch{ID: 81, BatchNo: "2026-08-12-am", Status: ImportBatchReviewing}}
	router := projectAdminTestRouter(app)
	body := `{"batch_no":"2026-08-12-am","items":[{"slug":"example","name":"Example","value_prop_zh":"价值","death_cause_zh":"原因","failure_analysis_zh":"分析","learnings_zh":["教训"],"sources":[{"field_name":"death_cause","url":"https://example.com","kind":"authority"}]}]}`
	request := httptest.NewRequest(http.MethodPost, "/api/v1/admin/import-batches", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusCreated || app.importInput.AdminUserID != 42 || app.importInput.BatchNo != "2026-08-12-am" {
		t.Fatalf("status=%d input=%+v body=%s", recorder.Code, app.importInput, recorder.Body.String())
	}
}

func TestPublishProjectImportBatchMapsPublicationGate(t *testing.T) {
	app := &fakeApplication{err: ErrPublicationGate}
	router := projectAdminTestRouter(app)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/admin/import-batches/81/publish", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusUnprocessableEntity || !strings.Contains(recorder.Body.String(), "publication_gate_failed") {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestProjectImportBatchAdminRequired(t *testing.T) {
	app := &fakeApplication{err: ErrAdminRequired}
	router := projectAdminTestRouter(app)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/admin/import-batches", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusForbidden || !strings.Contains(recorder.Body.String(), "admin_required") {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestProjectMatchFileEndpointsUploadValidateAndEnforceOwnership(t *testing.T) {
	manager := projectfiles.NewManager(projectfiles.NewDevelopmentStorage(), projectfiles.DevelopmentScanner{}, projectfiles.DevelopmentParser{}, projectfiles.DefaultMaxFileSize)
	service := NewService(nil, nil, WithProjectFileManager(manager))
	router := projectTestRouter(service)

	upload := multipartProjectFileRequest(t, "requirements.txt", projectfiles.MIMEText, []byte("预算 3 万元，希望做线上服务"))
	uploadRecorder := httptest.NewRecorder()
	router.ServeHTTP(uploadRecorder, upload)
	if uploadRecorder.Code != http.StatusCreated || !strings.Contains(uploadRecorder.Body.String(), `"parse_status":"ready"`) {
		t.Fatalf("upload status/body = %d/%s", uploadRecorder.Code, uploadRecorder.Body.String())
	}

	mismatch := multipartProjectFileRequest(t, "fake.txt", projectfiles.MIMEText, []byte("%PDF-1.4"))
	mismatchRecorder := httptest.NewRecorder()
	router.ServeHTTP(mismatchRecorder, mismatch)
	if mismatchRecorder.Code != http.StatusBadRequest || !strings.Contains(mismatchRecorder.Body.String(), `"error":"project_file_mime_mismatch"`) {
		t.Fatalf("mismatch status/body = %d/%s", mismatchRecorder.Code, mismatchRecorder.Body.String())
	}

	otherUserFile, err := manager.Upload(context.Background(), 7, "private.txt", projectfiles.MIMEText, []byte("private"), time.Hour)
	if err != nil {
		t.Fatalf("Upload(other user) error = %v", err)
	}
	otherUserRecorder := httptest.NewRecorder()
	router.ServeHTTP(otherUserRecorder, httptest.NewRequest(http.MethodGet, "/api/v1/project-match-files/"+strconv.FormatInt(otherUserFile.ID, 10), nil))
	if otherUserRecorder.Code != http.StatusNotFound || !strings.Contains(otherUserRecorder.Body.String(), `"error":"project_file_not_found"`) {
		t.Fatalf("other user status/body = %d/%s", otherUserRecorder.Code, otherUserRecorder.Body.String())
	}
}

func multipartProjectFileRequest(t *testing.T, name, contentType string, content []byte) *http.Request {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	header := make(textproto.MIMEHeader)
	header.Set("Content-Disposition", `form-data; name="file"; filename="`+name+`"`)
	header.Set("Content-Type", contentType)
	part, err := writer.CreatePart(header)
	if err != nil {
		t.Fatalf("CreatePart() error = %v", err)
	}
	if _, err := part.Write(content); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	request := httptest.NewRequest(http.MethodPost, "/api/v1/project-match-files", &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	return request
}

func TestProjectMatchWorkflowEndpointsUseAuthenticatedUserAndIdempotency(t *testing.T) {
	app := &fakeApplication{workflowResponse: MatchWorkflowResponse{MatchID: 200, Status: MatchStatusClarifying, Revision: 1}}
	router := projectTestRouter(app)

	create := httptest.NewRequest(http.MethodPost, "/api/v1/project-matches", strings.NewReader(`{"need":"做内容项目","profile_patch":{"team_size":1}}`))
	create.Header.Set("Content-Type", "application/json")
	create.Header.Set("Idempotency-Key", "create-42-1")
	createRecorder := httptest.NewRecorder()
	router.ServeHTTP(createRecorder, create)
	if createRecorder.Code != http.StatusOK || app.workflowCreate.UserID != 42 || app.workflowCreate.IdempotencyKey != "create-42-1" {
		t.Fatalf("create status/input = %d/%+v body=%s", createRecorder.Code, app.workflowCreate, createRecorder.Body.String())
	}

	answer := httptest.NewRequest(http.MethodPost, "/api/v1/project-matches/200/answer", strings.NewReader(`{"revision":1,"answers":[{"question_id":"budget","field":"budget_band","value":"0-5k"}]}`))
	answer.Header.Set("Content-Type", "application/json")
	answer.Header.Set("Idempotency-Key", "answer-200-1")
	answerRecorder := httptest.NewRecorder()
	router.ServeHTTP(answerRecorder, answer)
	if answerRecorder.Code != http.StatusOK || app.workflowAnswer.UserID != 42 || app.workflowAnswer.MatchID != 200 || app.workflowAnswer.Revision != 1 || app.workflowAnswer.IdempotencyKey != "answer-200-1" {
		t.Fatalf("answer status/input = %d/%+v body=%s", answerRecorder.Code, app.workflowAnswer, answerRecorder.Body.String())
	}

	getRecorder := httptest.NewRecorder()
	router.ServeHTTP(getRecorder, httptest.NewRequest(http.MethodGet, "/api/v1/project-matches/200", nil))
	if getRecorder.Code != http.StatusOK || app.userID != 42 || app.matchID != 200 {
		t.Fatalf("get status/user/match = %d/%d/%d", getRecorder.Code, app.userID, app.matchID)
	}
}

func TestProjectMatchWorkflowReturnsRevisionConflict(t *testing.T) {
	app := &fakeApplication{err: ErrMatchRevisionConflict}
	router := projectTestRouter(app)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/project-matches/200/answer", strings.NewReader(`{"revision":1,"skip":true}`))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusConflict || !strings.Contains(recorder.Body.String(), "match_revision_conflict") {
		t.Fatalf("status/body = %d/%s", recorder.Code, recorder.Body.String())
	}
}

func TestProjectMatchGenerationAndCancellationEndpoints(t *testing.T) {
	app := &fakeApplication{generationResponse: MatchGenerationResponse{MatchID: 200, Status: MatchStatusQueued, Attempt: 1}}
	router := projectTestRouter(app)
	generate := httptest.NewRecorder()
	router.ServeHTTP(generate, httptest.NewRequest(http.MethodPost, "/api/v1/project-matches/200/generate", nil))
	if generate.Code != http.StatusAccepted || app.userID != 42 || app.matchID != 200 {
		t.Fatalf("generate status/user/match = %d/%d/%d", generate.Code, app.userID, app.matchID)
	}
	cancel := httptest.NewRecorder()
	router.ServeHTTP(cancel, httptest.NewRequest(http.MethodPost, "/api/v1/project-matches/200/cancel", nil))
	if cancel.Code != http.StatusOK || app.userID != 42 || app.matchID != 200 {
		t.Fatalf("cancel status/user/match = %d/%d/%d", cancel.Code, app.userID, app.matchID)
	}
}

func TestProjectMatchStreamReplaysAfterLastEventID(t *testing.T) {
	app := &fakeApplication{progressEvents: []MatchProgressEvent{{ID: 8, MatchID: 200, Event: MatchStepDone, ProgressPercent: 100, Payload: map[string]any{"status": "completed"}}}}
	router := projectTestRouter(app)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/project-matches/200/stream", nil)
	request.Header.Set("Last-Event-ID", "7")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK || app.afterEventID != 7 || !strings.Contains(recorder.Body.String(), "id: 8\nevent: done\n") || !strings.Contains(recorder.Header().Get("Content-Type"), "text/event-stream") {
		t.Fatalf("status/after/body = %d/%d/%s", recorder.Code, app.afterEventID, recorder.Body.String())
	}
}

func TestOpportunityEndpointsReturnPublishedCatalog(t *testing.T) {
	app := &fakeApplication{
		opportunities: []Opportunity{{ID: 42, Slug: "ai-sales", Title: "AI销售顾问"}},
		opportunity: Opportunity{
			ID: 42, Slug: "ai-sales", Title: "AI销售顾问",
			Sections: []OpportunitySection{{
				Key: "data", Title: "当前数据", Body: "接口结构化数据", Items: []string{"预算区间"},
				Blocks: []OpportunitySectionBlock{{
					Type: "metrics", Title: "关键指标", Columns: 2,
					Items: []OpportunitySectionItem{{Title: "启动预算", Value: "1万元", Tone: "positive", Progress: 72, Tags: []string{"接口数据"}}},
				}},
			}},
		},
	}
	router := projectTestRouter(app)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/projects/opportunities?q=AI&industry=%E4%BC%81%E4%B8%9A%E6%9C%8D%E5%8A%A1", nil))
	if recorder.Code != http.StatusOK || app.filters.Query != "AI" || app.filters.Industry != "企业服务" {
		t.Fatalf("status/filters = %d/%+v body=%s", recorder.Code, app.filters, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"slug":"ai-sales"`) {
		t.Fatalf("body = %s", recorder.Body.String())
	}

	recorder = httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/projects/opportunities/ai-sales", nil))
	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), `"title":"AI销售顾问"`) {
		t.Fatalf("status/body = %d/%s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"key":"data"`) ||
		!strings.Contains(recorder.Body.String(), `"type":"metrics"`) ||
		!strings.Contains(recorder.Body.String(), `"progress":72`) {
		t.Fatalf("structured blocks missing from body = %s", recorder.Body.String())
	}
}

func TestProjectMarketPublicEndpoints(t *testing.T) {
	app := &fakeApplication{
		publicConfig:   PublicConfig{FeaturePaywallEnabled: false},
		dictionaries:   []DictionaryItem{{Code: "ai", Kind: "sector", NameZH: "人工智能"}},
		home:           ProjectHome{Hero: ProjectHero{Title: "项目超市"}, Featured: []Project{}},
		projectPage:    ProjectPage{Items: []Project{{ID: 42, Slug: "ai-sales", Title: "AI销售顾问"}}, Page: 2, PageSize: 6, Total: 1},
		catalogProject: Project{ID: 42, Slug: "ai-sales", Title: "AI销售顾问", LockedBlocks: []string{}, IsUnlocked: true},
	}
	router := projectTestRouter(app)

	checks := []struct {
		path string
		body string
	}{
		{path: "/api/v1/config", body: `"feature_paywall_enabled":false`},
		{path: "/api/v1/dicts?kind=sector", body: `"name_zh":"人工智能"`},
		{path: "/api/v1/projects/home", body: `"title":"项目超市"`},
		{path: "/api/v1/projects/42", body: `"slug":"ai-sales"`},
	}
	for _, check := range checks {
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, check.path, nil))
		if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), check.body) {
			t.Fatalf("GET %s status/body = %d/%s", check.path, recorder.Code, recorder.Body.String())
		}
	}

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/projects?keyword=AI&category=service&track=ai&budget=0-5k&difficulty=low&resource=solo&sort=latest&is_featured=true&page=2&page_size=6", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("list status/body = %d/%s", recorder.Code, recorder.Body.String())
	}
	if app.projectFilters.Keyword != "AI" || app.projectFilters.Page != 2 || app.projectFilters.PageSize != 6 || app.projectFilters.Featured == nil || !*app.projectFilters.Featured {
		t.Fatalf("filters = %+v", app.projectFilters)
	}
}

func TestListCasesEndpointFiltersByOpportunitySlug(t *testing.T) {
	app := &fakeApplication{cases: []CaseStudy{{
		ID: 81, Slug: "short-video-first-client", OpportunitySlug: "ai-short-video-studio",
		OpportunityTitle: "AI短视频脚本工作室", Industry: "内容服务", Title: "首个客户", CaseType: "success",
	}}}
	router := projectTestRouter(app)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, httptest.NewRequest(
		http.MethodGet,
		"/api/v1/projects/cases?type=success&opportunity_slug=%20ai-short-video-studio%20&industry=%20%E5%86%85%E5%AE%B9%E6%9C%8D%E5%8A%A1%20",
		nil,
	))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.caseFilters.CaseType != "success" || app.caseFilters.OpportunitySlug != "ai-short-video-studio" || app.caseFilters.Industry != "内容服务" || app.caseFilters.Limit != 20 {
		t.Fatalf("filters = %+v", app.caseFilters)
	}
	if !strings.Contains(recorder.Body.String(), `"slug":"short-video-first-client"`) ||
		!strings.Contains(recorder.Body.String(), `"opportunity_title":"AI短视频脚本工作室"`) ||
		!strings.Contains(recorder.Body.String(), `"industry":"内容服务"`) {
		t.Fatalf("body = %s", recorder.Body.String())
	}
}

func TestEvidenceCaseEndpointsUseIndependentPRDContract(t *testing.T) {
	app := &fakeApplication{
		evidenceCasePage: EvidenceCasePage{Items: []EvidenceCaseItem{{ID: 81, Title: "失败案例", Type: "fail"}}, Page: 2, PageSize: 10, Total: 1},
		evidenceCase: EvidenceCaseDetail{
			EvidenceCaseItem: EvidenceCaseItem{ID: 81, Title: "失败案例", Type: "fail"},
			Facts:            []EvidenceFact{{Field: "died_year", Value: "2024", SourceRefs: []int64{91}}},
			Sources:          []EvidenceSource{{ID: 91, URL: "https://example.com/case"}},
		},
	}
	router := projectTestRouter(app)

	list := httptest.NewRecorder()
	router.ServeHTTP(list, httptest.NewRequest(http.MethodGet, "/api/v1/project-cases?type=fail&industry=AI&scale=solo&page=2&page_size=10", nil))
	if list.Code != http.StatusOK || !strings.Contains(list.Body.String(), `"items"`) || !strings.Contains(list.Body.String(), `"title":"失败案例"`) {
		t.Fatalf("list status/body = %d/%s", list.Code, list.Body.String())
	}
	if app.evidenceFilters.CaseType != "fail" || app.evidenceFilters.Page != 2 || app.evidenceFilters.PageSize != 10 {
		t.Fatalf("filters = %+v", app.evidenceFilters)
	}

	detail := httptest.NewRecorder()
	router.ServeHTTP(detail, httptest.NewRequest(http.MethodGet, "/api/v1/project-cases/81", nil))
	if detail.Code != http.StatusOK || !strings.Contains(detail.Body.String(), `"facts"`) || !strings.Contains(detail.Body.String(), `"source_refs":[91]`) {
		t.Fatalf("detail status/body = %d/%s", detail.Code, detail.Body.String())
	}
}

func TestListMatchesEndpointUsesAuthenticatedUser(t *testing.T) {
	app := &fakeApplication{sessions: []MatchSession{{ID: 99, UserID: 42, Intent: "我的项目", Status: StatusCompleted}}}
	router := projectTestRouter(app)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/projects/matches", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.userID != 42 {
		t.Fatalf("userID = %d, want 42", app.userID)
	}
	if !strings.Contains(recorder.Body.String(), `"intent":"我的项目"`) {
		t.Fatalf("body = %s", recorder.Body.String())
	}
}

func TestGetMatchEndpointReturnsNotFoundForAnotherUser(t *testing.T) {
	app := &fakeApplication{err: ErrSessionNotFound}
	router := projectTestRouter(app)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/projects/matches/99", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestFavoriteMatchEndpointIsUserScoped(t *testing.T) {
	app := &fakeApplication{favorite: Favorite{ID: 7, UserID: 42, SessionID: 99}}
	router := projectTestRouter(app)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/projects/matches/99/favorite", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if app.userID != 42 || app.matchID != 99 {
		t.Fatalf("user/match = %d/%d", app.userID, app.matchID)
	}
	if !strings.Contains(recorder.Body.String(), `"session_id":99`) {
		t.Fatalf("body = %s", recorder.Body.String())
	}
}

func TestListAndDeleteFavoriteMatchEndpoints(t *testing.T) {
	app := &fakeApplication{favorites: []Favorite{{ID: 7, UserID: 42, SessionID: 99, Session: &MatchSession{ID: 99, Intent: "线上轻资产"}}}}
	router := projectTestRouter(app)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/projects/favorites", nil))
	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), `"intent":"线上轻资产"`) {
		t.Fatalf("list status/body = %d/%s", recorder.Code, recorder.Body.String())
	}

	recorder = httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodDelete, "/api/v1/projects/matches/99/favorite", nil))
	if recorder.Code != http.StatusNoContent || app.userID != 42 || app.matchID != 99 {
		t.Fatalf("delete status/user/match = %d/%d/%d", recorder.Code, app.userID, app.matchID)
	}
}

func TestProjectExportLifecycleEndpoints(t *testing.T) {
	expires := time.Date(2026, 8, 19, 12, 0, 0, 0, time.UTC)
	app := &fakeApplication{export: Export{ID: 71, UserID: 42, SourceType: ExportSourceMatch, SourceID: 99, Status: "queued", Format: "pdf", ExpiresAt: expires}}
	router := projectTestRouter(app)

	create := httptest.NewRecorder()
	router.ServeHTTP(create, httptest.NewRequest(http.MethodPost, "/api/v1/projects/exports", strings.NewReader(`{"source_type":"match","source_id":99,"format":"pdf"}`)))
	if create.Code != http.StatusAccepted || !strings.Contains(create.Body.String(), `"status":"queued"`) {
		t.Fatalf("create status/body = %d/%s", create.Code, create.Body.String())
	}
	if app.exportInput.UserID != 42 || app.exportInput.SourceID != 99 || app.exportInput.Format != "pdf" {
		t.Fatalf("create input = %+v", app.exportInput)
	}

	status := httptest.NewRecorder()
	router.ServeHTTP(status, httptest.NewRequest(http.MethodGet, "/api/v1/projects/exports/71", nil))
	if status.Code != http.StatusOK || !strings.Contains(status.Body.String(), `"status":"queued"`) {
		t.Fatalf("get status/body = %d/%s", status.Code, status.Body.String())
	}
}

func TestProjectExportDownloadRequiresReadyPDF(t *testing.T) {
	app := &fakeApplication{export: Export{ID: 71, UserID: 42, Status: "queued", Format: "pdf"}}
	router := projectTestRouter(app)

	queued := httptest.NewRecorder()
	router.ServeHTTP(queued, httptest.NewRequest(http.MethodGet, "/api/v1/projects/exports/71/download", nil))
	if queued.Code != http.StatusConflict || !strings.Contains(queued.Body.String(), `"error":"export_unavailable"`) {
		t.Fatalf("queued status/body = %d/%s", queued.Code, queued.Body.String())
	}

	app.export.Status = "ready"
	app.export.Payload = []byte("%PDF-1.7\nrendered")
	ready := httptest.NewRecorder()
	router.ServeHTTP(ready, httptest.NewRequest(http.MethodGet, "/api/v1/projects/exports/71/download", nil))
	if ready.Code != http.StatusOK || ready.Header().Get("Content-Type") != "application/pdf" || !strings.HasPrefix(ready.Body.String(), "%PDF-") {
		t.Fatalf("ready status/content-type/body = %d/%q/%q", ready.Code, ready.Header().Get("Content-Type"), ready.Body.String())
	}
}

func TestProjectExportEndpointReturnsGoneAfterExpiry(t *testing.T) {
	app := &fakeApplication{err: ErrExportExpired}
	router := projectTestRouter(app)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/projects/exports/71", nil))
	if recorder.Code != http.StatusGone || !strings.Contains(recorder.Body.String(), `"error":"export_expired"`) {
		t.Fatalf("status/body = %d/%s", recorder.Code, recorder.Body.String())
	}
}
