package files

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"sort"
	"strings"
)

// DevelopmentParser handles deterministic local formats without executing document content.
type DevelopmentParser struct{}

func (DevelopmentParser) Parse(_ context.Context, mime string, content []byte) (ParsedDocument, error) {
	switch mime {
	case MIMEText, MIMEMarkdown, MIMECSV:
		return ParsedDocument{Text: string(content), Metadata: map[string]any{"format": mime}}, nil
	case MIMEPDF:
		text := extractPDFStrings(content)
		if strings.TrimSpace(text) == "" {
			return ParsedDocument{}, ErrUnsupportedParsing
		}
		return ParsedDocument{Text: text, Metadata: map[string]any{"format": MIMEPDF}}, nil
	case MIMEDOCX:
		text, err := parseOfficeZip(content, "word/document.xml")
		if err != nil {
			return ParsedDocument{}, err
		}
		return ParsedDocument{Text: text, Metadata: map[string]any{"format": MIMEDOCX}}, nil
	case MIMEPPTX:
		text, err := parsePresentation(content)
		if err != nil {
			return ParsedDocument{}, err
		}
		return ParsedDocument{Text: text, Metadata: map[string]any{"format": MIMEPPTX}}, nil
	case MIMEXLSX:
		text, err := parseSpreadsheet(content)
		if err != nil {
			return ParsedDocument{}, err
		}
		return ParsedDocument{Text: text, Metadata: map[string]any{"format": MIMEXLSX}}, nil
	case MIMEPNG, MIMEJPEG:
		return ParsedDocument{}, ErrOCRUnavailable
	default:
		return ParsedDocument{}, ErrUnsupportedParsing
	}
}

func parseOfficeZip(content []byte, wanted string) (string, error) {
	files, err := openSafeZip(content)
	if err != nil {
		return "", err
	}
	for _, file := range files {
		if file.Name != wanted {
			continue
		}
		data, err := readZipEntry(file)
		if err != nil {
			return "", err
		}
		return xmlText(data), nil
	}
	return "", ErrUnsupportedParsing
}

func parsePresentation(content []byte) (string, error) {
	files, err := openSafeZip(content)
	if err != nil {
		return "", err
	}
	items := make([]*zip.File, 0)
	for _, file := range files {
		if strings.HasPrefix(file.Name, "ppt/slides/slide") && strings.HasSuffix(file.Name, ".xml") {
			items = append(items, file)
		}
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Name < items[j].Name })
	var parts []string
	for _, file := range items {
		data, err := readZipEntry(file)
		if err != nil {
			return "", err
		}
		if text := strings.TrimSpace(xmlText(data)); text != "" {
			parts = append(parts, text)
		}
	}
	if len(parts) == 0 {
		return "", ErrUnsupportedParsing
	}
	return strings.Join(parts, "\n\n"), nil
}

func parseSpreadsheet(content []byte) (string, error) {
	files, err := openSafeZip(content)
	if err != nil {
		return "", err
	}
	shared := []string{}
	for _, file := range files {
		if file.Name != "xl/sharedStrings.xml" {
			continue
		}
		data, readErr := readZipEntry(file)
		if readErr != nil {
			return "", readErr
		}
		shared = xmlTextList(data)
	}
	sheets := make([]*zip.File, 0)
	for _, file := range files {
		if strings.HasPrefix(file.Name, "xl/worksheets/sheet") && strings.HasSuffix(file.Name, ".xml") {
			sheets = append(sheets, file)
		}
	}
	sort.Slice(sheets, func(i, j int) bool { return sheets[i].Name < sheets[j].Name })
	var parts []string
	for _, file := range sheets {
		data, readErr := readZipEntry(file)
		if readErr != nil {
			return "", readErr
		}
		values := xmlValues(data, shared)
		if len(values) > 0 {
			parts = append(parts, strings.Join(values, ","))
		}
	}
	if len(parts) == 0 {
		return "", ErrUnsupportedParsing
	}
	return strings.Join(parts, "\n"), nil
}

func openSafeZip(content []byte) ([]*zip.File, error) {
	reader, err := zip.NewReader(bytes.NewReader(content), int64(len(content)))
	if err != nil {
		return nil, fmt.Errorf("%w: invalid office archive", ErrUnsupportedParsing)
	}
	if len(reader.File) > 256 {
		return nil, fmt.Errorf("%w: too many archive entries", ErrUnsupportedParsing)
	}
	var total uint64
	for _, file := range reader.File {
		if file.UncompressedSize64 > 50<<20 || total > 50<<20-file.UncompressedSize64 {
			return nil, fmt.Errorf("%w: archive is too large", ErrUnsupportedParsing)
		}
		total += file.UncompressedSize64
	}
	return reader.File, nil
}

func readZipEntry(file *zip.File) ([]byte, error) {
	reader, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	return io.ReadAll(io.LimitReader(reader, 50<<20))
}

func xmlText(data []byte) string {
	decoder := xml.NewDecoder(bytes.NewReader(data))
	var builder strings.Builder
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return strings.TrimSpace(builder.String())
		}
		if character, ok := token.(xml.CharData); ok {
			value := strings.TrimSpace(string(character))
			if value != "" {
				if builder.Len() > 0 {
					builder.WriteByte(' ')
				}
				builder.WriteString(value)
			}
		}
	}
	return strings.TrimSpace(builder.String())
}

func xmlTextList(data []byte) []string {
	decoder := xml.NewDecoder(bytes.NewReader(data))
	var values []string
	var current strings.Builder
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}
		switch typed := token.(type) {
		case xml.StartElement:
			if typed.Name.Local == "si" {
				current.Reset()
			}
		case xml.CharData:
			if value := strings.TrimSpace(string(typed)); value != "" {
				current.WriteString(value)
			}
		case xml.EndElement:
			if typed.Name.Local == "si" && current.Len() > 0 {
				values = append(values, current.String())
			}
		}
	}
	return values
}

func xmlValues(data []byte, shared []string) []string {
	decoder := xml.NewDecoder(bytes.NewReader(data))
	var values []string
	sharedType := false
	insideValue := false
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}
		switch typed := token.(type) {
		case xml.StartElement:
			switch typed.Name.Local {
			case "c":
				sharedType = false
				for _, attr := range typed.Attr {
					if attr.Name.Local == "t" && attr.Value == "s" {
						sharedType = true
					}
				}
			case "v":
				insideValue = true
			}
		case xml.CharData:
			if insideValue {
				value := strings.TrimSpace(string(typed))
				if value != "" {
					if sharedType {
						var index int
						if _, scanErr := fmt.Sscanf(value, "%d", &index); scanErr == nil && index >= 0 && index < len(shared) {
							value = shared[index]
						}
					}
					values = append(values, value)
				}
			}
		case xml.EndElement:
			if typed.Name.Local == "v" {
				insideValue = false
			}
		}
	}
	return values
}

func extractPDFStrings(content []byte) string {
	var parts []string
	for index := 0; index < len(content); index++ {
		if content[index] != '(' {
			continue
		}
		index++
		var value strings.Builder
		depth := 1
		for index < len(content) && depth > 0 {
			switch content[index] {
			case '\\':
				index++
				if index < len(content) {
					value.WriteByte(content[index])
				}
			case '(':
				depth++
				value.WriteByte('(')
			case ')':
				depth--
				if depth > 0 {
					value.WriteByte(')')
				}
			default:
				value.WriteByte(content[index])
			}
			index++
		}
		if text := strings.TrimSpace(value.String()); text != "" {
			parts = append(parts, text)
		}
	}
	return strings.Join(parts, " ")
}
