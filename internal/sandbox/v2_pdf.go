package sandbox

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/go-pdf/fpdf"
)

var errV2PDFFontUnavailable = errors.New("sandbox pdf font unavailable")

// RenderV2SandboxPDF turns a completed sandbox run into a paginated PDF. The
// font is loaded at runtime so the same binary works on macOS and in Alpine.
func RenderV2SandboxPDF(run V2SandboxRun) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetTitle("商业沙盘分析报告", true)
	pdf.SetAuthor("OPCV2", true)
	pdf.SetMargins(16, 18, 16)
	pdf.SetAutoPageBreak(true, 18)
	pdf.AliasNbPages("")

	fontBytes, err := loadV2PDFFont()
	if err != nil {
		return nil, err
	}
	pdf.AddUTF8FontFromBytes("sandbox", "", fontBytes)
	pdf.AddUTF8FontFromBytes("sandbox", "B", fontBytes)
	if pdf.Error() != nil {
		return nil, fmt.Errorf("load sandbox pdf font: %w", pdf.Error())
	}

	pdf.SetHeaderFunc(func() {
		pdf.SetFont("sandbox", "", 8)
		pdf.SetTextColor(120, 127, 140)
		pdf.CellFormat(0, 6, "OPCV2 · 商业沙盘", "", 1, "L", false, 0, "")
		pdf.SetDrawColor(225, 229, 235)
		pdf.Line(16, 25, 194, 25)
		pdf.SetTextColor(35, 42, 52)
	})
	pdf.SetFooterFunc(func() {
		pdf.SetY(-13)
		pdf.SetFont("sandbox", "", 8)
		pdf.SetTextColor(130, 137, 150)
		pdf.CellFormat(0, 6, fmt.Sprintf("第 %d 页 / 共 {nb} 页", pdf.PageNo()), "", 0, "C", false, 0, "")
	})

	pdf.AddPage()
	pdf.SetFont("sandbox", "B", 22)
	pdf.SetTextColor(26, 33, 45)
	pdf.CellFormat(0, 12, nonEmpty(run.Name, "未命名项目"), "", 1, "L", false, 0, "")
	pdf.SetFont("sandbox", "", 11)
	pdf.SetTextColor(90, 99, 112)
	pdf.CellFormat(0, 8, "商业沙盘分析报告", "", 1, "L", false, 0, "")
	pdf.Ln(4)

	writeV2PDFMeta(pdf, run)
	writeV2PDFSection(pdf, "核心结论")
	writeV2PDFParagraph(pdf, run.Report.Summary)

	basis := nonEmpty(run.Report.Feasibility.Basis, run.Report.PurchaseProbability.Basis)
	pdf.SetFont("sandbox", "", 9)
	basisLines := len(pdf.SplitText(basis, 166))
	boxHeight := max(30.0, 20+float64(basisLines)*5)
	boxY := pdf.GetY() + 2
	pdf.SetFillColor(246, 248, 251)
	pdf.SetDrawColor(225, 229, 235)
	pdf.RoundedRect(16, boxY, 178, boxHeight, 2, "1234", "DF")
	pdf.SetXY(22, boxY+5)
	pdf.SetFont("sandbox", "B", 15)
	pdf.SetTextColor(32, 86, 153)
	pdf.CellFormat(52, 8, fmt.Sprintf("可行性 %d 分", run.Report.Feasibility.Score), "", 0, "L", false, 0, "")
	pdf.SetFont("sandbox", "", 10)
	pdf.SetTextColor(70, 78, 91)
	pdf.CellFormat(52, 8, nonEmpty(run.Report.Feasibility.Level, "待研判"), "", 0, "L", false, 0, "")
	pdf.SetFont("sandbox", "B", 15)
	pdf.SetTextColor(29, 122, 87)
	pdf.CellFormat(58, 8, fmt.Sprintf("购买概率 %d%%", run.Report.PurchaseProbability.ValuePct), "", 1, "L", false, 0, "")
	pdf.SetXY(22, boxY+15)
	pdf.SetFont("sandbox", "", 9)
	pdf.SetTextColor(90, 99, 112)
	pdf.MultiCell(166, 5, basis, "", "L", false)
	pdf.SetY(boxY + boxHeight + 4)

	writeV2PDFInsights(pdf, "机会", run.Report.Opportunity, func(item V2Insight) (string, string) { return item.Point, item.Reason })
	writeV2PDFRisks(pdf, run.Report.Risk)
	writeV2PDFAdvice(pdf, run.Report.Advice)
	writeV2PDFRoles(pdf, run.Report.RoleTakeaways)
	writeV2PDFDimensions(pdf, run.Report.DimensionSummary)
	writeV2PDFDisagreements(pdf, run.Report.Disagreements)
	writeV2PDFScenarios(pdf, run.Report.Scenarios)
	writeV2PDFList(pdf, "前提与假设", run.Report.Assumptions)

	var output bytes.Buffer
	if err := pdf.Output(&output); err != nil {
		return nil, err
	}
	if pdf.Error() != nil {
		return nil, pdf.Error()
	}
	return output.Bytes(), nil
}

func writeV2PDFMeta(pdf *fpdf.Fpdf, run V2SandboxRun) {
	pdf.SetFont("sandbox", "", 9)
	pdf.SetTextColor(85, 94, 108)
	meta := []string{
		"产品：" + nonEmpty(run.Product.Name, "未填写"),
		"目标客户：" + nonEmpty(run.Context.TargetCustomer, "未填写"),
		"市场：" + nonEmpty(run.Context.Market, "未填写"),
		"渠道：" + nonEmpty(run.Context.Channel, "未填写"),
	}
	for index := 0; index < len(meta); index += 2 {
		writeV2PDFMetaRow(pdf, meta[index], meta[index+1])
	}
	pdf.Ln(3)
}

func writeV2PDFMetaRow(pdf *fpdf.Fpdf, left, right string) {
	const columnWidth = 89.0
	lineHeight := 5.0
	leftLines := max(1, len(pdf.SplitText(left, columnWidth-3)))
	rightLines := max(1, len(pdf.SplitText(right, columnWidth-3)))
	rowHeight := float64(max(leftLines, rightLines)) * lineHeight
	x, y := pdf.GetX(), pdf.GetY()
	pdf.MultiCell(columnWidth, lineHeight, left, "", "L", false)
	pdf.SetXY(x+columnWidth, y)
	pdf.MultiCell(columnWidth, lineHeight, right, "", "L", false)
	pdf.SetXY(x, y+rowHeight+1)
}

func writeV2PDFSection(pdf *fpdf.Fpdf, title string) {
	if pdf.GetY() > 260 {
		pdf.AddPage()
	}
	pdf.SetFont("sandbox", "B", 13)
	pdf.SetTextColor(26, 33, 45)
	pdf.CellFormat(0, 9, title, "", 1, "L", false, 0, "")
	pdf.SetDrawColor(42, 111, 183)
	pdf.SetLineWidth(0.8)
	pdf.Line(16, pdf.GetY(), 31, pdf.GetY())
	pdf.SetLineWidth(0.2)
	pdf.Ln(3)
}

func writeV2PDFParagraph(pdf *fpdf.Fpdf, text string) {
	if strings.TrimSpace(text) == "" {
		return
	}
	pdf.SetFont("sandbox", "", 10)
	pdf.SetTextColor(58, 66, 80)
	pdf.MultiCell(178, 5.5, strings.TrimSpace(text), "", "L", false)
	pdf.Ln(2)
}

func writeV2PDFInsights(pdf *fpdf.Fpdf, title string, items []V2Insight, fields func(V2Insight) (string, string)) {
	if len(items) == 0 {
		return
	}
	writeV2PDFSection(pdf, title)
	for _, item := range items {
		point, reason := fields(item)
		writeV2PDFBullet(pdf, point, reason)
	}
}

func writeV2PDFRisks(pdf *fpdf.Fpdf, items []V2ReportRisk) {
	if len(items) == 0 {
		return
	}
	writeV2PDFSection(pdf, "风险与缓解")
	for _, item := range items {
		label := item.Point
		if item.Severity != "" {
			label += "（" + localizeV2PDFValue(item.Severity) + "）"
		}
		writeV2PDFBullet(pdf, label, strings.TrimSpace(item.Reason+" "+item.Mitigation))
	}
}

func writeV2PDFAdvice(pdf *fpdf.Fpdf, items []V2Advice) {
	if len(items) == 0 {
		return
	}
	writeV2PDFSection(pdf, "行动建议")
	for index, item := range items {
		label := item.Action
		if label == "" {
			label = fmt.Sprintf("建议 %d", index+1)
		}
		reason := item.Why
		if item.Priority > 0 {
			reason = fmt.Sprintf("优先级 %d。%s", item.Priority, reason)
		}
		if item.Effort != "" {
			reason += " 执行投入：" + item.Effort + "。"
		}
		writeV2PDFBullet(pdf, label, reason)
	}
}

func writeV2PDFRoles(pdf *fpdf.Fpdf, items []V2RoleTakeaway) {
	if len(items) == 0 {
		return
	}
	writeV2PDFSection(pdf, "角色结论")
	for _, item := range items {
		label := localizeV2PDFValue(item.Role)
		if item.Stance != "" {
			label += " · " + localizeV2PDFValue(item.Stance)
		}
		reason := strings.Join(item.KeyPoints, "；")
		writeV2PDFBullet(pdf, label, reason)
	}
}

func writeV2PDFDimensions(pdf *fpdf.Fpdf, items []V2DimensionSummary) {
	if len(items) == 0 {
		return
	}
	writeV2PDFSection(pdf, "维度汇总")
	pdf.SetFillColor(240, 244, 248)
	pdf.SetTextColor(48, 58, 72)
	writeV2PDFDimensionRow(pdf, []string{"维度", "分数", "共识"}, true)
	for _, item := range items {
		row := []string{item.Dimension, fmt.Sprintf("%d", item.Score), item.Consensus}
		if pdf.GetY()+v2PDFDimensionRowHeight(pdf, row) > 278 {
			pdf.AddPage()
			writeV2PDFSection(pdf, "维度汇总（续）")
			writeV2PDFDimensionRow(pdf, []string{"维度", "分数", "共识"}, true)
		}
		writeV2PDFDimensionRow(pdf, row, false)
	}
	pdf.Ln(3)
}

func writeV2PDFDimensionRow(pdf *fpdf.Fpdf, cells []string, header bool) {
	widths := []float64{44, 18, 116}
	if header {
		pdf.SetFont("sandbox", "B", 9)
	} else {
		pdf.SetFont("sandbox", "", 9)
	}
	rowHeight := v2PDFDimensionRowHeight(pdf, cells)
	x, y := pdf.GetX(), pdf.GetY()
	style := "D"
	if header {
		style = "DF"
	}
	for index, cell := range cells {
		width := widths[index]
		pdf.Rect(x, y, width, rowHeight, style)
		align := "L"
		if index == 1 {
			align = "C"
		}
		pdf.SetXY(x+2, y+1)
		pdf.MultiCell(width-4, 5, cell, "", align, false)
		x += width
	}
	pdf.SetXY(16, y+rowHeight)
}

func v2PDFDimensionRowHeight(pdf *fpdf.Fpdf, cells []string) float64 {
	widths := []float64{44, 18, 116}
	lines := 1
	for index, cell := range cells {
		lines = max(lines, len(pdf.SplitText(cell, widths[index]-4)))
	}
	return max(7.0, float64(lines)*5+2)
}

func writeV2PDFDisagreements(pdf *fpdf.Fpdf, items []V2Disagreement) {
	if len(items) == 0 {
		return
	}
	writeV2PDFSection(pdf, "分歧与待决策事项")
	for _, item := range items {
		reason := item.DecisionNeeded
		for _, view := range item.Views {
			if reason != "" {
				reason += "；"
			}
			reason += view.Role + "：" + view.Point
		}
		writeV2PDFBullet(pdf, item.Topic, reason)
	}
}

func writeV2PDFScenarios(pdf *fpdf.Fpdf, scenarios map[string]V2Scenario) {
	if len(scenarios) == 0 {
		return
	}
	keys := make([]string, 0, len(scenarios))
	for key := range scenarios {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	writeV2PDFSection(pdf, "情景推演")
	for _, key := range keys {
		item := scenarios[key]
		writeV2PDFBullet(pdf, key, strings.TrimSpace(item.Desc+" "+item.Condition))
	}
}

func writeV2PDFList(pdf *fpdf.Fpdf, title string, items []string) {
	if len(items) == 0 {
		return
	}
	writeV2PDFSection(pdf, title)
	for _, item := range items {
		writeV2PDFBullet(pdf, item, "")
	}
}

func writeV2PDFBullet(pdf *fpdf.Fpdf, title, body string) {
	pdf.SetFont("sandbox", "B", 10)
	pdf.SetTextColor(45, 54, 67)
	pdf.CellFormat(5, 5.5, "-", "", 0, "L", false, 0, "")
	pdf.SetX(22)
	pdf.MultiCell(172, 5.5, strings.TrimSpace(title), "", "L", false)
	if strings.TrimSpace(body) != "" {
		pdf.SetFont("sandbox", "", 9.5)
		pdf.SetTextColor(86, 95, 108)
		pdf.SetX(22)
		pdf.MultiCell(172, 5, strings.TrimSpace(body), "", "L", false)
	}
	pdf.Ln(1.5)
}

func loadV2PDFFont() ([]byte, error) {
	paths := make([]string, 0, 8)
	if configured := strings.TrimSpace(os.Getenv("OPCV2_PDF_FONT_PATH")); configured != "" {
		paths = append(paths, configured)
	}
	paths = append(paths,
		"/usr/share/fonts/wqy-zenhei/wqy-zenhei.ttc",
		"/Library/Fonts/Arial Unicode.ttf",
		"/Library/Fonts/Arial Unicode MS.ttf",
	)
	for _, path := range paths {
		data, err := os.ReadFile(filepath.Clean(path))
		if err != nil {
			continue
		}
		font, err := extractV2TTFFace(data, 0)
		if err == nil && len(font) > 0 {
			return font, nil
		}
	}
	return nil, errV2PDFFontUnavailable
}

// extractV2TTFFace extracts one TrueType face from a TrueType Collection.
// gofpdf embeds TrueType fonts but intentionally does not parse TTC headers.
func extractV2TTFFace(data []byte, face int) ([]byte, error) {
	if len(data) < 4 {
		return nil, errors.New("font data too short")
	}
	if string(data[:4]) != "ttcf" {
		if binary.BigEndian.Uint32(data[:4]) == 0x00010000 || string(data[:4]) == "true" {
			return data, nil
		}
		return nil, errors.New("unsupported font format")
	}
	if len(data) < 16 {
		return nil, errors.New("invalid font collection")
	}
	numFonts := int(binary.BigEndian.Uint32(data[8:12]))
	if face < 0 || face >= numFonts || 12+4*face+4 > len(data) {
		return nil, errors.New("font face out of range")
	}
	offset := int(binary.BigEndian.Uint32(data[12+4*face : 16+4*face]))
	if offset < 0 || offset+12 > len(data) {
		return nil, errors.New("invalid font face offset")
	}
	if binary.BigEndian.Uint32(data[offset:offset+4]) != 0x00010000 && string(data[offset:offset+4]) != "true" {
		return nil, errors.New("font collection face is not TrueType")
	}
	numTables := int(binary.BigEndian.Uint16(data[offset+4 : offset+6]))
	recordsEnd := offset + 12 + numTables*16
	if numTables <= 0 || recordsEnd > len(data) {
		return nil, errors.New("invalid font table directory")
	}
	out := make([]byte, 12+numTables*16)
	copy(out, data[offset:recordsEnd])
	cursor := len(out)
	for i := 0; i < numTables; i++ {
		record := 12 + i*16
		tableOffset := int(binary.BigEndian.Uint32(data[offset+record+8 : offset+record+12]))
		tableLength := int(binary.BigEndian.Uint32(data[offset+record+12 : offset+record+16]))
		if tableOffset < 0 || tableLength < 0 || tableOffset+tableLength > len(data) {
			return nil, errors.New("invalid font table")
		}
		for cursor%4 != 0 {
			out = append(out, 0)
			cursor++
		}
		binary.BigEndian.PutUint32(out[record+8:record+12], uint32(cursor))
		out = append(out, data[tableOffset:tableOffset+tableLength]...)
		cursor += tableLength
	}
	return out, nil
}

func nonEmpty(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func localizeV2PDFValue(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "high":
		return "高"
	case "medium":
		return "中"
	case "low":
		return "低"
	case "support":
		return "支持"
	case "neutral":
		return "中立"
	case "oppose":
		return "反对"
	case "customer":
		return "目标客户"
	case "investor":
		return "投资人"
	case "competitor":
		return "竞争对手"
	case "channel":
		return "渠道方"
	case "supply":
		return "供应链运营"
	case "expert":
		return "行业专家"
	case "skeptic":
		return "悲观者"
	case "partner":
		return "合伙人"
	default:
		return value
	}
}
