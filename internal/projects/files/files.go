package files

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"unicode"
)

const (
	DefaultMaxFileSize  = 20 << 20
	DefaultMaxFileCount = 10
	DefaultMaxTotalSize = 50 << 20
	DefaultFileTTL      = 30 * 24 * time.Hour
	MIMEText            = "text/plain"
	MIMEMarkdown        = "text/markdown"
	MIMECSV             = "text/csv"
	MIMEPDF             = "application/pdf"
	MIMEDOCX            = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	MIMEXLSX            = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	MIMEPPTX            = "application/vnd.openxmlformats-officedocument.presentationml.presentation"
	MIMEPNG             = "image/png"
	MIMEJPEG            = "image/jpeg"
)

const (
	StatusUploading = "uploading"
	StatusScanning  = "scanning"
	StatusParsing   = "parsing"
	StatusReady     = "ready"
	StatusFailed    = "failed"
	StatusDeleted   = "deleted"
)

var (
	ErrFileNotFound        = errors.New("project match file not found")
	ErrFileExpired         = errors.New("project match file expired")
	ErrFileTooLarge        = errors.New("project match file is too large")
	ErrFileCountExceeded   = errors.New("project match file count exceeded")
	ErrTotalSizeExceeded   = errors.New("project match total size exceeded")
	ErrUnsupportedMIME     = errors.New("unsupported project match file MIME type")
	ErrMIMEMismatch        = errors.New("project match file MIME does not match content")
	ErrUnsafeFile          = errors.New("unsafe project match file")
	ErrInvalidFileName     = errors.New("invalid project match file name")
	ErrUnsupportedParsing  = errors.New("project match file type cannot be parsed")
	ErrFileNotReady        = errors.New("project match file is not ready")
	ErrFileAlreadyAttached = errors.New("project match file is already attached")
	ErrOCRUnavailable      = errors.New("OCR provider is not configured")
)

type File struct {
	ID            int64          `json:"id"`
	UserID        int64          `json:"-"`
	MatchID       *int64         `json:"match_id,omitempty"`
	Name          string         `json:"name"`
	MIME          string         `json:"mime_type"`
	DetectedMIME  string         `json:"detected_mime"`
	Size          int64          `json:"size_bytes"`
	ObjectKey     string         `json:"-"`
	SHA256        string         `json:"sha256"`
	ParseStatus   string         `json:"parse_status"`
	ExtractedText string         `json:"extracted_text,omitempty"`
	ExtractedJSON map[string]any `json:"extracted_json,omitempty"`
	ErrorCode     string         `json:"error_code,omitempty"`
	ExpiresAt     time.Time      `json:"expires_at"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

type ParsedDocument struct {
	Text             string
	Metadata         map[string]any
	UntrustedContent bool
}

type ObjectStorage interface {
	Put(context.Context, string, []byte) error
	Get(context.Context, string) ([]byte, error)
	Delete(context.Context, string) error
}

type Scanner interface {
	Scan(context.Context, []byte) error
}

type Parser interface {
	Parse(context.Context, string, []byte) (ParsedDocument, error)
}

// MetadataStore keeps ownership and parse state durable independently from objects.
type MetadataStore interface {
	CreateFile(context.Context, File) (File, error)
	GetFile(context.Context, int64, int64) (File, error)
	ListFiles(context.Context, int64, *int64) ([]File, error)
	UpdateFile(context.Context, File) (File, error)
}

type Manager struct {
	storage ObjectStorage
	scanner Scanner
	parser  Parser
	maxSize int64
	store   MetadataStore
	now     func() time.Time
	mu      sync.RWMutex
	files   map[int64]File
	nextID  int64
}

func NewManager(storage ObjectStorage, scanner Scanner, parser Parser, maxSize int64) *Manager {
	if maxSize <= 0 {
		maxSize = DefaultMaxFileSize
	}
	return &Manager{storage: storage, scanner: scanner, parser: parser, maxSize: maxSize, now: time.Now, files: map[int64]File{}, nextID: 1}
}

func (m *Manager) SetMetadataStore(store MetadataStore) {
	m.mu.Lock()
	m.store = store
	m.mu.Unlock()
}

func (m *Manager) Upload(ctx context.Context, userID int64, name, declaredMIME string, content []byte, ttl time.Duration) (File, error) {
	if userID <= 0 || len(content) == 0 || int64(len(content)) > m.maxSize {
		return File{}, ErrFileTooLarge
	}
	name = sanitizeName(name)
	if name == "" {
		return File{}, ErrInvalidFileName
	}
	mime, detected, err := validateMIME(name, declaredMIME, content)
	if err != nil {
		return File{}, err
	}
	if m.scanner != nil {
		if err := m.scanner.Scan(ctx, content); err != nil {
			return File{}, err
		}
	}
	if ttl <= 0 || ttl > DefaultFileTTL {
		ttl = DefaultFileTTL
	}
	files, listErr := m.listFilesIncludingExpired(ctx, userID)
	if listErr != nil {
		return File{}, listErr
	}
	now := m.now()
	activeCount, activeSize := 0, int64(0)
	contentChecksum := checksum(content)
	for _, existing := range files {
		if existing.ParseStatus == StatusDeleted {
			continue
		}
		if !now.Before(existing.ExpiresAt) {
			existing.ParseStatus, existing.ErrorCode, existing.UpdatedAt = StatusDeleted, "expired", now
			_ = m.storage.Delete(ctx, existing.ObjectKey)
			if _, err := m.updateFile(ctx, existing); err != nil {
				return File{}, err
			}
			continue
		}
		activeCount++
		activeSize += existing.Size
		if existing.SHA256 == contentChecksum {
			return File{}, ErrFileAlreadyAttached
		}
	}
	if activeCount >= DefaultMaxFileCount {
		return File{}, ErrFileCountExceeded
	}
	if activeSize+int64(len(content)) > DefaultMaxTotalSize {
		return File{}, ErrTotalSizeExceeded
	}
	store := m.metadataStore()
	file := File{
		UserID: userID, Name: name, MIME: mime, DetectedMIME: detected,
		Size: int64(len(content)), ObjectKey: fmt.Sprintf("project-matches/%d/%d-%s", userID, now.UnixNano(), name),
		SHA256: contentChecksum, ParseStatus: StatusScanning, ExpiresAt: now.Add(ttl), CreatedAt: now, UpdatedAt: now,
	}
	if store != nil {
		file, err = store.CreateFile(ctx, file)
		if err != nil {
			return File{}, err
		}
	} else {
		m.mu.Lock()
		file.ID = m.nextID
		m.nextID++
		m.files[file.ID] = file
		m.mu.Unlock()
	}
	if err := m.storage.Put(ctx, file.ObjectKey, content); err != nil {
		file.ParseStatus, file.ErrorCode = StatusFailed, "storage_failed"
		_, _ = m.updateFile(ctx, file)
		return File{}, err
	}
	return file, nil
}

func (m *Manager) Parse(ctx context.Context, userID, fileID int64) (ParsedDocument, error) {
	file, err := m.getOwned(ctx, userID, fileID)
	if err != nil {
		return ParsedDocument{}, err
	}
	if !m.now().Before(file.ExpiresAt) {
		file.ParseStatus, file.ErrorCode = StatusDeleted, "expired"
		_ = m.storage.Delete(ctx, file.ObjectKey)
		_, _ = m.updateFile(ctx, file)
		return ParsedDocument{}, ErrFileExpired
	}
	file.ParseStatus, file.ErrorCode, file.UpdatedAt = StatusScanning, "", m.now()
	if _, err := m.updateFile(ctx, file); err != nil {
		return ParsedDocument{}, err
	}
	content, err := m.storage.Get(ctx, file.ObjectKey)
	if err != nil {
		return m.fail(ctx, file, "storage_failed", err)
	}
	if m.scanner != nil {
		if err := m.scanner.Scan(ctx, content); err != nil {
			return m.fail(ctx, file, "unsafe_file", err)
		}
	}
	file.ParseStatus, file.UpdatedAt = StatusParsing, m.now()
	if _, err := m.updateFile(ctx, file); err != nil {
		return ParsedDocument{}, err
	}
	document, err := m.parser.Parse(ctx, file.MIME, content)
	if err != nil {
		code := "parse_failed"
		if errors.Is(err, ErrOCRUnavailable) {
			code = "ocr_unavailable"
		}
		return m.fail(ctx, file, code, err)
	}
	document.UntrustedContent = true
	file.ParseStatus, file.ExtractedText, file.ExtractedJSON, file.UpdatedAt = StatusReady, document.Text, document.Metadata, m.now()
	file.ErrorCode = ""
	if _, err := m.updateFile(ctx, file); err != nil {
		return ParsedDocument{}, err
	}
	return document, nil
}

func (m *Manager) Retry(ctx context.Context, userID, fileID int64) (File, error) {
	file, err := m.getOwned(ctx, userID, fileID)
	if err != nil {
		return File{}, err
	}
	file.ParseStatus, file.ErrorCode, file.UpdatedAt = StatusScanning, "", m.now()
	if _, err := m.updateFile(ctx, file); err != nil {
		return File{}, err
	}
	_, err = m.Parse(ctx, userID, fileID)
	if err != nil {
		file, _ = m.getOwned(ctx, userID, fileID)
		return file, err
	}
	return m.getOwned(ctx, userID, fileID)
}

func (m *Manager) Attach(ctx context.Context, userID, matchID int64, fileIDs []int64) ([]File, error) {
	if matchID <= 0 || len(fileIDs) == 0 || len(fileIDs) > DefaultMaxFileCount {
		return nil, ErrInvalidFileName
	}
	seen := map[int64]bool{}
	files := make([]File, 0, len(fileIDs))
	for _, id := range fileIDs {
		if id <= 0 || seen[id] {
			return nil, ErrFileAlreadyAttached
		}
		seen[id] = true
		file, err := m.getOwned(ctx, userID, id)
		if err != nil {
			return nil, err
		}
		if file.ParseStatus != StatusReady {
			return nil, ErrFileNotReady
		}
		if file.MatchID != nil && *file.MatchID != matchID {
			return nil, ErrFileAlreadyAttached
		}
		files = append(files, file)
	}
	for index := range files {
		files[index].MatchID = &matchID
		files[index].UpdatedAt = m.now()
		updated, err := m.updateFile(ctx, files[index])
		if err != nil {
			return nil, err
		}
		files[index] = updated
	}
	return files, nil
}

func (m *Manager) Delete(ctx context.Context, userID, fileID int64) error {
	file, err := m.getOwned(ctx, userID, fileID)
	if err != nil {
		return err
	}
	_ = m.storage.Delete(ctx, file.ObjectKey)
	file.ParseStatus, file.ErrorCode, file.UpdatedAt = StatusDeleted, "deleted", m.now()
	_, err = m.updateFile(ctx, file)
	return err
}

func (m *Manager) Get(ctx context.Context, userID, fileID int64) (File, error) {
	file, err := m.getOwned(ctx, userID, fileID)
	if err != nil {
		return File{}, err
	}
	if !m.now().Before(file.ExpiresAt) {
		return File{}, ErrFileExpired
	}
	return file, nil
}

func (m *Manager) List(ctx context.Context, userID int64, matchID *int64) ([]File, error) {
	store := m.metadataStore()
	if store != nil {
		files, err := store.ListFiles(ctx, userID, matchID)
		if err != nil {
			return nil, err
		}
		result := make([]File, 0, len(files))
		for _, file := range files {
			if file.ParseStatus != StatusDeleted && m.now().Before(file.ExpiresAt) {
				result = append(result, file)
			}
		}
		return result, nil
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]File, 0)
	for _, file := range m.files {
		if file.UserID == userID && file.ParseStatus != StatusDeleted && m.now().Before(file.ExpiresAt) && (matchID == nil || (file.MatchID != nil && *file.MatchID == *matchID)) {
			result = append(result, file)
		}
	}
	return result, nil
}

func (m *Manager) metadataStore() MetadataStore {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.store
}

func (m *Manager) listFilesIncludingExpired(ctx context.Context, userID int64) ([]File, error) {
	if store := m.metadataStore(); store != nil {
		return store.ListFiles(ctx, userID, nil)
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]File, 0)
	for _, file := range m.files {
		if file.UserID == userID {
			result = append(result, file)
		}
	}
	return result, nil
}

func (m *Manager) getOwned(ctx context.Context, userID, fileID int64) (File, error) {
	if userID <= 0 || fileID <= 0 {
		return File{}, ErrFileNotFound
	}
	if store := m.metadataStore(); store != nil {
		return store.GetFile(ctx, userID, fileID)
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	file, ok := m.files[fileID]
	if !ok || file.UserID != userID || file.ParseStatus == StatusDeleted {
		return File{}, ErrFileNotFound
	}
	return file, nil
}

func (m *Manager) updateFile(ctx context.Context, file File) (File, error) {
	if store := m.metadataStore(); store != nil {
		return store.UpdateFile(ctx, file)
	}
	m.mu.Lock()
	m.files[file.ID] = file
	m.mu.Unlock()
	return file, nil
}

func (m *Manager) fail(ctx context.Context, file File, code string, err error) (ParsedDocument, error) {
	file.ParseStatus, file.ErrorCode, file.UpdatedAt = StatusFailed, code, m.now()
	_, _ = m.updateFile(ctx, file)
	return ParsedDocument{}, err
}

func sanitizeName(raw string) string {
	name := strings.TrimSpace(raw)
	if name == "" || name != filepath.Base(name) || strings.ContainsAny(name, "/\\\x00") {
		return ""
	}
	for _, r := range name {
		if unicode.IsControl(r) {
			return ""
		}
	}
	return name
}

func checksum(content []byte) string {
	digest := sha256.Sum256(content)
	return hex.EncodeToString(digest[:])
}

func validateMIME(name, declared string, content []byte) (string, string, error) {
	ext := strings.ToLower(filepath.Ext(name))
	expected := mimeForExtension(ext)
	if expected == "" {
		return "", "", ErrUnsupportedMIME
	}
	if strings.TrimSpace(declared) == "" {
		declared = expected
	}
	declared = strings.ToLower(strings.TrimSpace(strings.Split(declared, ";")[0]))
	if !supportedMIME(declared) {
		return "", "", ErrUnsupportedMIME
	}
	if declared != expected {
		return "", "", ErrMIMEMismatch
	}
	detected := detectMIME(content, ext)
	if !mimeMatches(declared, detected, content) {
		return "", detected, ErrMIMEMismatch
	}
	return declared, detected, nil
}

func supportedMIME(mime string) bool {
	switch mime {
	case MIMEText, MIMEMarkdown, MIMECSV, MIMEPDF, MIMEDOCX, MIMEXLSX, MIMEPPTX, MIMEPNG, MIMEJPEG:
		return true
	default:
		return false
	}
}

func mimeForExtension(ext string) string {
	switch ext {
	case ".txt":
		return MIMEText
	case ".md", ".markdown":
		return MIMEMarkdown
	case ".csv":
		return MIMECSV
	case ".pdf":
		return MIMEPDF
	case ".docx":
		return MIMEDOCX
	case ".xlsx":
		return MIMEXLSX
	case ".pptx":
		return MIMEPPTX
	case ".png":
		return MIMEPNG
	case ".jpg", ".jpeg":
		return MIMEJPEG
	default:
		return ""
	}
}

func detectMIME(content []byte, ext string) string {
	switch {
	case bytes.HasPrefix(content, []byte("%PDF-")):
		return MIMEPDF
	case len(content) >= 4 && bytes.Equal(content[:4], []byte{'P', 'K', 3, 4}):
		return detectOfficeMIME(content)
	case bytes.HasPrefix(content, []byte("\x89PNG\r\n\x1a\n")):
		return MIMEPNG
	case len(content) >= 3 && bytes.Equal(content[:3], []byte{0xff, 0xd8, 0xff}):
		return MIMEJPEG
	default:
		detected := strings.ToLower(strings.TrimSpace(strings.Split(http.DetectContentType(content), ";")[0]))
		if detected == "application/octet-stream" && mimeForExtension(ext) != "" {
			return mimeForExtension(ext)
		}
		return detected
	}
}

func detectOfficeMIME(content []byte) string {
	reader, err := zip.NewReader(bytes.NewReader(content), int64(len(content)))
	if err != nil {
		return "application/zip"
	}
	for _, file := range reader.File {
		switch file.Name {
		case "word/document.xml":
			return MIMEDOCX
		case "xl/workbook.xml":
			return MIMEXLSX
		case "ppt/presentation.xml":
			return MIMEPPTX
		}
	}
	return "application/zip"
}

func mimeMatches(declared, detected string, content []byte) bool {
	if declared == MIMEText || declared == MIMEMarkdown || declared == MIMECSV {
		return detected == MIMEText || detected == MIMEMarkdown || detected == MIMECSV || (detected == "application/octet-stream" && isText(content))
	}
	return declared == detected
}

func isText(content []byte) bool {
	return !bytes.Contains(content, []byte{0})
}

type DevelopmentStorage struct {
	mu      sync.RWMutex
	objects map[string][]byte
}

func NewDevelopmentStorage() *DevelopmentStorage {
	return &DevelopmentStorage{objects: map[string][]byte{}}
}

func (s *DevelopmentStorage) Put(_ context.Context, key string, content []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.objects[key] = append([]byte(nil), content...)
	return nil
}

func (s *DevelopmentStorage) Get(_ context.Context, key string) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	content, ok := s.objects[key]
	if !ok {
		return nil, ErrFileNotFound
	}
	return append([]byte(nil), content...), nil
}

func (s *DevelopmentStorage) Delete(_ context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.objects, key)
	return nil
}

type DevelopmentScanner struct{}

func (DevelopmentScanner) Scan(_ context.Context, content []byte) error {
	if bytes.Contains(bytes.ToUpper(content), []byte("EICAR-STANDARD-ANTIVIRUS-TEST-FILE")) {
		return ErrUnsafeFile
	}
	return nil
}
