package projects

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-pdf/fpdf"
)

var ErrProjectPDFFontUnavailable = errors.New("project pdf font unavailable")

func RenderProjectPDF(item Export) ([]byte, error) {
	switch item.SourceType {
	case ExportSourceMatch:
		var run MatchRun
		if err := json.Unmarshal(item.Snapshot, &run); err == nil && run.ID > 0 {
			return renderProjectMatchRunPDF(run)
		}
		var session MatchSession
		if err := json.Unmarshal(item.Snapshot, &session); err == nil && session.ID > 0 {
			run = MatchRun{ID: session.ID, Need: session.Intent, Status: session.Status, Result: session.Result, CreatedAt: session.CreatedAt, UpdatedAt: session.UpdatedAt}
			return renderProjectMatchRunPDF(run)
		}
	case ExportSourceComparison:
		var comparison Comparison
		if err := json.Unmarshal(item.Snapshot, &comparison); err == nil && comparison.ID > 0 {
			return renderProjectComparisonPDF(comparison)
		}
	}
	return nil, ErrInvalidExport
}

func newProjectPDF(title string) (*fpdf.Fpdf, error) {
	font, err := loadProjectPDFFont()
	if err != nil {
		return nil, err
	}
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.AddUTF8FontFromBytes("project", "", font)
	pdf.AddUTF8FontFromBytes("project", "B", font)
	pdf.SetTitle(title, true)
	pdf.SetAuthor("OPCV2", true)
	pdf.SetMargins(16, 36, 16)
	pdf.SetAutoPageBreak(true, 18)
	pdf.AliasNbPages("")
	pdf.SetHeaderFuncMode(func() {
		pdf.SetY(14)
		pdf.SetFont("project", "", 8)
		pdf.SetTextColor(112, 121, 134)
		pdf.CellFormat(0, 6, "OPCV2 · 项目超市", "", 1, "L", false, 0, "")
		pdf.SetDrawColor(223, 228, 235)
		pdf.Line(16, 25, 194, 25)
	}, true)
	pdf.SetFooterFunc(func() {
		pdf.SetY(-13)
		pdf.SetFont("project", "", 8)
		pdf.SetTextColor(120, 128, 140)
		pdf.CellFormat(0, 6, fmt.Sprintf("第 %d 页 / 共 {nb} 页", pdf.PageNo()), "", 0, "C", false, 0, "")
	})
	return pdf, nil
}

func renderProjectMatchRunPDF(run MatchRun) ([]byte, error) {
	pdf, err := newProjectPDF("项目匹配分析报告")
	if err != nil {
		return nil, err
	}
	pdf.AddPage()
	pdf.SetY(36)
	pdf.SetFont("project", "B", 22)
	pdf.SetTextColor(26, 34, 47)
	pdf.CellFormat(0, 12, "项目匹配分析报告", "", 1, "L", false, 0, "")
	pdf.SetFont("project", "", 9)
	pdf.SetTextColor(92, 101, 115)
	pdf.CellFormat(0, 7, fmt.Sprintf("匹配记录 #%d · 状态 %s", run.ID, run.Status), "", 1, "L", false, 0, "")
	pdf.Ln(4)
	projectPDFSection(pdf, "需求摘要")
	projectPDFParagraph(pdf, run.Need)
	if len(run.ParsedProfile) > 0 {
		projectPDFSection(pdf, "已确认画像")
		keys := []string{"budget_band", "time_per_week", "risk_preference", "team_size", "skills"}
		for _, key := range keys {
			if value, ok := run.ParsedProfile[key]; ok {
				projectPDFBullet(pdf, projectPDFFieldLabel(key), fmt.Sprint(value))
			}
		}
	}
	projectPDFSection(pdf, "推荐项目")
	for _, project := range run.Result.Projects {
		projectPDFBullet(pdf, fmt.Sprintf("%d. %s（%d 分）", project.Rank, project.Title, project.Score), strings.Join(project.Reasons, "；"))
		projectPDFParagraph(pdf, "启动预算："+project.Budget+"\n风险提示："+project.Risk)
	}
	if len(run.Result.Evidence) > 0 {
		projectPDFSection(pdf, "证据与来源")
		for index, evidence := range run.Result.Evidence {
			detail := evidence.Excerpt
			if evidence.Publisher != "" {
				detail = evidence.Publisher + "。" + detail
			}
			if evidence.URL != "" {
				detail += "\n" + evidence.URL
			}
			projectPDFBullet(pdf, fmt.Sprintf("[%d] %s（质量 %.0f%%）", index+1, evidence.Title, evidence.Quality*100), detail)
		}
	}
	return outputProjectPDF(pdf)
}

func renderProjectComparisonPDF(comparison Comparison) ([]byte, error) {
	pdf, err := newProjectPDF("项目对比报告")
	if err != nil {
		return nil, err
	}
	pdf.AddPage()
	pdf.SetY(36)
	pdf.SetFont("project", "B", 22)
	pdf.CellFormat(0, 12, "项目对比报告", "", 1, "L", false, 0, "")
	for index, project := range comparison.Items {
		projectPDFSection(pdf, fmt.Sprintf("%d. %s", index+1, project.Title))
		projectPDFParagraph(pdf, project.Summary)
		projectPDFBullet(pdf, "预算", project.BudgetBand)
		projectPDFBullet(pdf, "难度", project.Difficulty)
		projectPDFBullet(pdf, "资源要求", strings.Join(project.ResourceRequirements, "、"))
	}
	return outputProjectPDF(pdf)
}

func projectPDFSection(pdf *fpdf.Fpdf, title string) {
	if pdf.GetY() > 260 {
		pdf.AddPage()
	}
	pdf.SetFont("project", "B", 13)
	pdf.SetTextColor(27, 36, 50)
	pdf.CellFormat(0, 9, title, "", 1, "L", false, 0, "")
	pdf.SetDrawColor(35, 115, 190)
	pdf.SetLineWidth(0.8)
	pdf.Line(16, pdf.GetY(), 31, pdf.GetY())
	pdf.SetLineWidth(0.2)
	pdf.Ln(3)
}

func projectPDFParagraph(pdf *fpdf.Fpdf, text string) {
	projectPDFParagraphWidth(pdf, text, 178)
}

func projectPDFParagraphWidth(pdf *fpdf.Fpdf, text string, width float64) {
	if strings.TrimSpace(text) == "" {
		return
	}
	pdf.SetFont("project", "", 9.5)
	pdf.SetTextColor(72, 81, 95)
	projectPDFWrappedText(pdf, strings.TrimSpace(text), width, 5.3)
	pdf.Ln(2)
}

func projectPDFWrappedText(pdf *fpdf.Fpdf, text string, width, lineHeight float64) {
	startX := pdf.GetX()
	paragraphs := strings.Split(strings.ReplaceAll(text, "\r", ""), "\n")
	for _, paragraph := range paragraphs {
		lines := pdf.SplitText(paragraph, width)
		if len(lines) == 0 {
			lines = []string{""}
		}
		for _, line := range lines {
			if pdf.GetY()+lineHeight > 278 {
				pdf.AddPage()
				pdf.SetY(36)
			}
			pdf.SetX(startX)
			pdf.CellFormat(width, lineHeight, line, "", 1, "L", false, 0, "")
		}
	}
}

func projectPDFBullet(pdf *fpdf.Fpdf, title, detail string) {
	projectPDFKeepBulletTogether(pdf, title, detail)
	pdf.SetFont("project", "B", 10)
	pdf.SetTextColor(43, 52, 66)
	pdf.MultiCell(178, 5.5, "- "+strings.TrimSpace(title), "", "L", false)
	if strings.TrimSpace(detail) != "" {
		pdf.SetX(22)
		projectPDFParagraphWidth(pdf, detail, 172)
	}
}

func projectPDFKeepBulletTogether(pdf *fpdf.Fpdf, title, detail string) {
	pdf.SetFont("project", "B", 10)
	height := float64(max(1, len(pdf.SplitText("- "+strings.TrimSpace(title), 178)))) * 5.5
	if strings.TrimSpace(detail) != "" {
		pdf.SetFont("project", "", 9.5)
		height += float64(max(1, len(pdf.SplitText(strings.TrimSpace(detail), 172))))*5.3 + 2
	}
	if pdf.GetY()+height > 278 {
		pdf.AddPage()
		pdf.SetY(36)
	}
}

func projectPDFFieldLabel(key string) string {
	labels := map[string]string{"budget_band": "启动预算", "time_per_week": "每周投入", "risk_preference": "风险偏好", "team_size": "团队规模", "skills": "能力方向"}
	return labels[key]
}

func outputProjectPDF(pdf *fpdf.Fpdf) ([]byte, error) {
	var output bytes.Buffer
	if err := pdf.Output(&output); err != nil {
		return nil, err
	}
	return output.Bytes(), pdf.Error()
}

func loadProjectPDFFont() ([]byte, error) {
	paths := []string{
		strings.TrimSpace(os.Getenv("OPCV2_PDF_FONT_PATH")),
		"/usr/share/fonts/wqy-zenhei/wqy-zenhei.ttc",
		"/Library/Fonts/Arial Unicode.ttf",
		"/Library/Fonts/Arial Unicode MS.ttf",
		"/System/Library/Fonts/STHeiti Medium.ttc",
		"/System/Library/Fonts/Hiragino Sans GB.ttc",
	}
	for _, path := range paths {
		if path == "" {
			continue
		}
		data, err := os.ReadFile(filepath.Clean(path))
		if err != nil {
			continue
		}
		font, err := extractProjectTTFFace(data, 0)
		if err == nil {
			return font, nil
		}
	}
	return nil, ErrProjectPDFFontUnavailable
}

func extractProjectTTFFace(data []byte, face int) ([]byte, error) {
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
	if face < 0 || face >= numFonts || 16+4*face > len(data) {
		return nil, errors.New("font face out of range")
	}
	offset := int(binary.BigEndian.Uint32(data[12+4*face : 16+4*face]))
	if offset+12 > len(data) {
		return nil, errors.New("invalid font face offset")
	}
	numTables := int(binary.BigEndian.Uint16(data[offset+4 : offset+6]))
	recordsEnd := offset + 12 + numTables*16
	if numTables <= 0 || recordsEnd > len(data) {
		return nil, errors.New("invalid font directory")
	}
	out := make([]byte, 12+numTables*16)
	copy(out, data[offset:recordsEnd])
	cursor := len(out)
	for index := 0; index < numTables; index++ {
		record := 12 + index*16
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
