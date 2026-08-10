package sandbox

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type fakeV2ExportRepository struct {
	*fakeV2Repository
	export  V2Export
	payload []byte
	err     error
}

func (r *fakeV2ExportRepository) CreateV2Export(_ context.Context, export V2Export, _ int64, payload []byte) (V2Export, error) {
	if r.err != nil {
		return V2Export{}, r.err
	}
	export.ID = 700
	r.export = export
	r.payload = append([]byte(nil), payload...)
	return export, nil
}

func (r *fakeV2ExportRepository) GetV2Export(_ context.Context, _ int64, exportID int64) (V2Export, []byte, error) {
	if r.err != nil || r.export.ID != exportID {
		return V2Export{}, nil, r.err
	}
	return r.export, append([]byte(nil), r.payload...), nil
}

func completedV2Run() V2SandboxRun {
	return V2SandboxRun{
		ID: 99, UserID: 42, Name: "社区餐饮订阅", Status: V2StatusDone,
		Product: V2Product{Name: "工作日健康午餐", PriceCents: 2800, Stage: "验证期"},
		Context: V2RunContext{TargetCustomer: "园区白领", Market: "本地生活", Channel: "企业微信"},
		Report: &V2SandboxReport{
			Summary:             "项目具备试点价值，但复购和履约成本仍需真实订单验证。",
			Feasibility:         V2Feasibility{Score: 76, Level: "中高", Basis: "需求清晰，交付链路可控。"},
			PurchaseProbability: V2PurchaseProbability{ValuePct: 63, Basis: "价格处于可接受区间。", IsModelGenerated: true},
			Opportunity:         []V2Insight{{Point: "高频刚需", Reason: "工作日午餐具有稳定消费频次。"}},
			Risk:                []V2ReportRisk{{Point: "履约波动", Severity: "high", Reason: "高峰配送时效不稳定。", Mitigation: "限定首批服务半径。"}},
			Advice:              []V2Advice{{Action: "完成 50 单付费试点", Why: "验证复购与配送成本。", Priority: 1, Effort: "两周"}},
			RoleTakeaways:       []V2RoleTakeaway{{Role: "目标客户", Stance: "support", KeyPoints: []string{"愿意为稳定和健康支付溢价"}}},
			DimensionSummary:    []V2DimensionSummary{{Dimension: "需求强度", Score: 82, Consensus: "多数角色认可", SupportingRoles: []string{"customer"}}},
			Scenarios:           map[string]V2Scenario{"基准情景": {Desc: "先覆盖单一园区", Condition: "周复购率达到 35%"}},
			Assumptions:         []string{"园区内存在至少 200 名稳定目标客户"}, IsModelGenerated: true,
		},
	}
}

func TestCreateV2ExportRendersRealPDFPayload(t *testing.T) {
	run := completedV2Run()
	repository := &fakeV2ExportRepository{fakeV2Repository: newFakeV2Repository()}
	repository.run = run
	rendered := []byte("%PDF-1.7\nrendered")
	service := NewService(repository, nil, WithV2PDFRenderer(func(got V2SandboxRun) ([]byte, error) {
		if got.ID != run.ID || got.Report == nil {
			t.Fatalf("renderer run = %+v", got)
		}
		return rendered, nil
	}))
	service.now = func() time.Time { return time.Date(2026, 8, 10, 12, 0, 0, 0, time.UTC) }

	export, err := service.CreateV2Export(context.Background(), CreateV2ExportInput{UserID: 42, RunID: run.ID, Format: "PDF"})
	if err != nil {
		t.Fatalf("CreateV2Export() error = %v", err)
	}
	if export.ID != 700 || export.Format != "pdf" || !bytes.Equal(repository.payload, rendered) {
		t.Fatalf("export = %+v payload=%q", export, repository.payload)
	}
}

func TestCreateV2ExportKeepsJSONPayload(t *testing.T) {
	run := completedV2Run()
	repository := &fakeV2ExportRepository{fakeV2Repository: newFakeV2Repository()}
	repository.run = run
	service := NewService(repository, nil)

	if _, err := service.CreateV2Export(context.Background(), CreateV2ExportInput{UserID: 42, RunID: run.ID, Format: "json"}); err != nil {
		t.Fatalf("CreateV2Export() error = %v", err)
	}
	if !bytes.Contains(repository.payload, []byte(`"format":"json"`)) || !bytes.Contains(repository.payload, []byte(`"report"`)) {
		t.Fatalf("json payload = %s", repository.payload)
	}
}

func TestGetV2ExportRejectsExpiredPayload(t *testing.T) {
	repository := &fakeV2ExportRepository{fakeV2Repository: newFakeV2Repository()}
	repository.export = V2Export{ID: 700, RunID: 99, Format: "pdf", ExpiresAt: time.Date(2026, 8, 9, 12, 0, 0, 0, time.UTC)}
	service := NewService(repository, nil)
	service.now = func() time.Time { return time.Date(2026, 8, 10, 12, 0, 0, 0, time.UTC) }

	if _, _, err := service.GetV2Export(context.Background(), 42, 700); !errors.Is(err, ErrV2ExportExpired) {
		t.Fatalf("GetV2Export() error = %v", err)
	}
}

type fakeV2ExportHTTPApplication struct {
	*fakeApplication
	export  V2Export
	payload []byte
}

func (a *fakeV2ExportHTTPApplication) CreateV2Export(_ context.Context, input CreateV2ExportInput) (V2Export, error) {
	return a.export, nil
}

func (a *fakeV2ExportHTTPApplication) GetV2Export(_ context.Context, _ int64, _ int64) (V2Export, []byte, error) {
	return a.export, a.payload, nil
}

func TestDownloadV2ExportUsesPDFContentType(t *testing.T) {
	app := &fakeV2ExportHTTPApplication{
		fakeApplication: &fakeApplication{},
		export:          V2Export{ID: 700, RunID: 99, Format: "pdf"},
		payload:         []byte("%PDF-1.7\nrendered"),
	}
	router := sandboxTestRouter(app)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/v1/sandbox-runs/99/exports/700/download", nil))

	if recorder.Code != http.StatusOK || recorder.Header().Get("Content-Type") != "application/pdf" {
		t.Fatalf("status=%d content-type=%q body=%q", recorder.Code, recorder.Header().Get("Content-Type"), recorder.Body.String())
	}
	if disposition := recorder.Header().Get("Content-Disposition"); !strings.Contains(disposition, "sandbox-export-700.pdf") {
		t.Fatalf("content-disposition = %q", disposition)
	}
}

func TestRenderV2SandboxPDFProducesReadableDocument(t *testing.T) {
	if _, err := loadV2PDFFont(); errors.Is(err, errV2PDFFontUnavailable) {
		t.Skip("CJK font is not installed in this test environment")
	}
	payload, err := RenderV2SandboxPDF(completedV2Run())
	if err != nil {
		t.Fatalf("RenderV2SandboxPDF() error = %v", err)
	}
	if len(payload) < 1000 || !bytes.HasPrefix(payload, []byte("%PDF-")) || !bytes.Contains(payload, []byte("/Type /Page")) {
		t.Fatalf("invalid PDF payload: bytes=%d prefix=%q", len(payload), payload[:min(len(payload), 8)])
	}
	if output := strings.TrimSpace(os.Getenv("OPCV2_PDF_TEST_OUTPUT")); output != "" {
		if err := os.MkdirAll(filepath.Dir(output), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(output, payload, 0o644); err != nil {
			t.Fatal(err)
		}
	}
}
