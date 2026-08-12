package projects

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRenderProjectPDFProducesReadableMatchAndComparisonReports(t *testing.T) {
	if _, err := loadProjectPDFFont(); errors.Is(err, ErrProjectPDFFontUnavailable) {
		t.Skip("CJK font is not installed in this test environment")
	}
	matchSnapshot, err := json.Marshal(MatchRun{
		ID: 99, Need: "预算 3 万元，每周投入 20 小时，优先线上获客。", Status: MatchStatusCompleted,
		ParsedProfile: map[string]any{"budget_band": "2w以上", "time_per_week": "20小时以上", "risk_preference": "低"},
		Result: MatchResult{Status: StatusCompleted,
			Projects: []ProjectMatch{{Rank: 1, Title: "AI 销售线索服务", Score: 86, Budget: "1-3 万元", Reasons: []string{"符合销售获客目标", "可在线交付"}, Risk: "需先验证客户付费意愿"}},
			Evidence: []MatchEvidence{{Title: "项目案例", Publisher: "OPCV2", Excerpt: "案例包含真实交付记录。", URL: "https://example.com/case", Quality: 0.9}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	comparisonSnapshot, err := json.Marshal(Comparison{ID: 61, Items: []Opportunity{
		{ID: 1, Title: "AI 销售", Summary: "销售流程自动化服务", BudgetBand: "1-3 万元", Difficulty: "中等", ResourceRequirements: []string{"销售经验", "AI 工具"}},
		{ID: 2, Title: "AI 内容", Summary: "企业内容生产服务", BudgetBand: "0.5-2 万元", Difficulty: "低", ResourceRequirements: []string{"内容策划"}},
	}})
	if err != nil {
		t.Fatal(err)
	}

	checks := []struct {
		name string
		item Export
	}{
		{name: "match", item: Export{SourceType: ExportSourceMatch, Snapshot: matchSnapshot}},
		{name: "comparison", item: Export{SourceType: ExportSourceComparison, Snapshot: comparisonSnapshot}},
	}
	for _, check := range checks {
		t.Run(check.name, func(t *testing.T) {
			payload, err := RenderProjectPDF(check.item)
			if err != nil {
				t.Fatalf("RenderProjectPDF() error = %v", err)
			}
			if len(payload) < 1000 || !bytes.HasPrefix(payload, []byte("%PDF-")) || !bytes.Contains(payload, []byte("/Type /Page")) {
				t.Fatalf("invalid PDF payload: bytes=%d prefix=%q", len(payload), payload[:min(len(payload), 8)])
			}
			if outputDir := strings.TrimSpace(os.Getenv("OPCV2_PROJECT_PDF_TEST_OUTPUT_DIR")); outputDir != "" {
				if err := os.MkdirAll(outputDir, 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(outputDir, "project-"+check.name+".pdf"), payload, 0o644); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}

func TestRenderProjectPDFKeepsContinuationContentBelowHeader(t *testing.T) {
	if _, err := loadProjectPDFFont(); errors.Is(err, ErrProjectPDFFontUnavailable) {
		t.Skip("CJK font is not installed in this test environment")
	}
	evidence := make([]MatchEvidence, 0, 12)
	for index := 0; index < 12; index++ {
		evidence = append(evidence, MatchEvidence{
			Title: "跨页证据", Publisher: "项目超市", Quality: 0.9,
			Excerpt: strings.Repeat("这是一段用于验证自动分页页眉间距的证据说明。", 8),
		})
	}
	snapshot, err := json.Marshal(MatchRun{ID: 99, Need: "长报告分页验证", Status: MatchStatusCompleted, Result: MatchResult{Status: StatusCompleted, Evidence: evidence}})
	if err != nil {
		t.Fatal(err)
	}
	payload, err := RenderProjectPDF(Export{SourceType: ExportSourceMatch, Snapshot: snapshot})
	if err != nil {
		t.Fatalf("RenderProjectPDF() error = %v", err)
	}
	if bytes.Count(payload, []byte("/Type /Page")) < 2 {
		t.Fatalf("expected multi-page PDF, bytes=%d", len(payload))
	}
	if outputDir := strings.TrimSpace(os.Getenv("OPCV2_PROJECT_PDF_TEST_OUTPUT_DIR")); outputDir != "" {
		if err := os.MkdirAll(outputDir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(outputDir, "project-continuation.pdf"), payload, 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestProjectPDFAutomaticPageBreakRestoresContentMargin(t *testing.T) {
	if _, err := loadProjectPDFFont(); errors.Is(err, ErrProjectPDFFontUnavailable) {
		t.Skip("CJK font is not installed in this test environment")
	}
	pdf, err := newProjectPDF("自动分页验证")
	if err != nil {
		t.Fatalf("newProjectPDF() error = %v", err)
	}
	pdf.AddPage()
	if got := pdf.GetY(); got != 36 {
		t.Fatalf("first page content margin = %.1f, want 36", got)
	}
	pdf.SetY(275)
	projectPDFParagraph(pdf, strings.Repeat("自动分页后正文必须位于页眉分隔线下方。", 12))
	if pdf.PageNo() < 2 {
		t.Fatal("expected paragraph to trigger an automatic page break")
	}
	if got := pdf.GetY(); got < 41.3 {
		t.Fatalf("continuation page cursor = %.1f, want at least one line below 36", got)
	}
}

func TestOpportunityEvidenceExcerptUsesReadableLabels(t *testing.T) {
	excerpt := opportunityEvidenceExcerpt(Opportunity{
		Summary: "为中小企业自动化重复报表流程。", Industry: "企业服务",
		Tags: []string{"办公效率", "可复制"}, BudgetBand: "0.5-2万元", Difficulty: "中等",
		ResourceRequirements: []string{"Excel函数", "业务流程梳理"},
	})
	for _, want := range []string{
		"行业：企业服务", "标签：办公效率、可复制", "启动预算：0.5-2万元",
		"难度：中等", "资源要求：Excel函数、业务流程梳理",
	} {
		if !strings.Contains(excerpt, want) {
			t.Fatalf("excerpt = %q, want %q", excerpt, want)
		}
	}
	if strings.ContainsAny(excerpt, "[]{}\"") {
		t.Fatalf("excerpt exposes raw JSON syntax: %q", excerpt)
	}
}
