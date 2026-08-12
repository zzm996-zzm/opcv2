package files

import (
	"archive/zip"
	"bytes"
	"context"
	"testing"
)

func TestDevelopmentParserExtractsSupportedDocuments(t *testing.T) {
	parser := DevelopmentParser{}
	tests := []struct {
		name    string
		mime    string
		content []byte
		want    string
	}{
		{name: "text", mime: MIMEText, content: []byte("预算 3 万元"), want: "预算 3 万元"},
		{name: "markdown", mime: MIMEMarkdown, content: []byte("# 项目计划"), want: "# 项目计划"},
		{name: "csv", mime: MIMECSV, content: []byte("项目,预算\n咨询,30000"), want: "咨询,30000"},
		{name: "pdf", mime: MIMEPDF, content: []byte("%PDF-1.4\nBT (Project budget 30000) Tj ET\n%%EOF"), want: "Project budget 30000"},
		{name: "docx", mime: MIMEDOCX, content: officeArchive(t, map[string]string{
			"word/document.xml": `<w:document xmlns:w="w"><w:body><w:p><w:r><w:t>项目需求</w:t></w:r></w:p></w:body></w:document>`,
		}), want: "项目需求"},
		{name: "pptx", mime: MIMEPPTX, content: officeArchive(t, map[string]string{
			"ppt/presentation.xml":  `<p:presentation xmlns:p="p"/>`,
			"ppt/slides/slide1.xml": `<p:sld xmlns:p="p" xmlns:a="a"><a:t>市场方案</a:t></p:sld>`,
		}), want: "市场方案"},
		{name: "xlsx", mime: MIMEXLSX, content: officeArchive(t, map[string]string{
			"xl/workbook.xml":          `<workbook/>`,
			"xl/sharedStrings.xml":     `<sst><si><t>启动预算</t></si></sst>`,
			"xl/worksheets/sheet1.xml": `<worksheet><sheetData><row><c t="s"><v>0</v></c><c><v>30000</v></c></row></sheetData></worksheet>`,
		}), want: "启动预算,30000"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			document, err := parser.Parse(context.Background(), test.mime, test.content)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}
			if !bytes.Contains([]byte(document.Text), []byte(test.want)) {
				t.Fatalf("Parse() text = %q, want %q", document.Text, test.want)
			}
		})
	}
}

func TestDevelopmentParserReportsUnavailableOCRAndInvalidOfficeArchive(t *testing.T) {
	parser := DevelopmentParser{}
	if _, err := parser.Parse(context.Background(), MIMEPNG, []byte("png")); err != ErrOCRUnavailable {
		t.Fatalf("Parse(PNG) error = %v", err)
	}
	if _, err := parser.Parse(context.Background(), MIMEDOCX, []byte("not zip")); err == nil {
		t.Fatal("Parse(invalid DOCX) returned nil error")
	}
}

func officeArchive(t *testing.T, entries map[string]string) []byte {
	t.Helper()
	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	for name, content := range entries {
		entry, err := writer.Create(name)
		if err != nil {
			t.Fatalf("Create(%q) error = %v", name, err)
		}
		if _, err := entry.Write([]byte(content)); err != nil {
			t.Fatalf("Write(%q) error = %v", name, err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	return buffer.Bytes()
}
