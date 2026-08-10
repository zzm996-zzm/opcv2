package files

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	DefaultMaxFileSize = 10 << 20
	MIMEText           = "text/plain"
	MIMEPDF            = "application/pdf"
	MIMEDOCX           = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
)

var (
	ErrFileNotFound       = errors.New("project match file not found")
	ErrFileExpired        = errors.New("project match file expired")
	ErrFileTooLarge       = errors.New("project match file is too large")
	ErrUnsupportedMIME    = errors.New("unsupported project match file MIME type")
	ErrMIMEMismatch       = errors.New("project match file MIME does not match content")
	ErrUnsafeFile         = errors.New("unsafe project match file")
	ErrUnsupportedParsing = errors.New("development parser does not support this file type")
)

type File struct {
	ID        int64
	UserID    int64
	Name      string
	MIME      string
	Size      int64
	ObjectKey string
	ExpiresAt time.Time
	CreatedAt time.Time
}

type ParsedDocument struct {
	Text             string
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

type Manager struct {
	storage ObjectStorage
	scanner Scanner
	parser  Parser
	maxSize int64
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

func (m *Manager) Upload(ctx context.Context, userID int64, name, declaredMIME string, content []byte, ttl time.Duration) (File, error) {
	if userID <= 0 || len(content) == 0 || int64(len(content)) > m.maxSize {
		return File{}, ErrFileTooLarge
	}
	mime, err := validateMIME(declaredMIME, content)
	if err != nil {
		return File{}, err
	}
	if m.scanner != nil {
		if err := m.scanner.Scan(ctx, content); err != nil {
			return File{}, err
		}
	}
	if ttl <= 0 || ttl > 24*time.Hour {
		ttl = 24 * time.Hour
	}
	m.mu.Lock()
	id := m.nextID
	m.nextID++
	now := m.now()
	file := File{ID: id, UserID: userID, Name: strings.TrimSpace(name), MIME: mime, Size: int64(len(content)), ObjectKey: fmt.Sprintf("project-matches/%d/%d", userID, id), CreatedAt: now, ExpiresAt: now.Add(ttl)}
	m.files[id] = file
	m.mu.Unlock()
	if err := m.storage.Put(ctx, file.ObjectKey, content); err != nil {
		m.mu.Lock()
		delete(m.files, id)
		m.mu.Unlock()
		return File{}, err
	}
	return file, nil
}

func (m *Manager) Parse(ctx context.Context, userID, fileID int64) (ParsedDocument, error) {
	file, err := m.getOwned(userID, fileID)
	if err != nil {
		return ParsedDocument{}, err
	}
	if !m.now().Before(file.ExpiresAt) {
		_ = m.storage.Delete(ctx, file.ObjectKey)
		m.mu.Lock()
		delete(m.files, file.ID)
		m.mu.Unlock()
		return ParsedDocument{}, ErrFileExpired
	}
	content, err := m.storage.Get(ctx, file.ObjectKey)
	if err != nil {
		return ParsedDocument{}, err
	}
	document, err := m.parser.Parse(ctx, file.MIME, content)
	if err != nil {
		return ParsedDocument{}, err
	}
	document.UntrustedContent = true
	return document, nil
}

func (m *Manager) getOwned(userID, fileID int64) (File, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	file, ok := m.files[fileID]
	if !ok || file.UserID != userID {
		return File{}, ErrFileNotFound
	}
	return file, nil
}

func validateMIME(declared string, content []byte) (string, error) {
	declared = strings.ToLower(strings.TrimSpace(strings.Split(declared, ";")[0]))
	switch declared {
	case MIMEText, MIMEPDF, MIMEDOCX:
	default:
		return "", ErrUnsupportedMIME
	}
	detected := strings.ToLower(strings.TrimSpace(strings.Split(http.DetectContentType(content), ";")[0]))
	match := declared == MIMEText && detected == MIMEText
	match = match || declared == MIMEPDF && bytes.HasPrefix(content, []byte("%PDF-"))
	match = match || declared == MIMEDOCX && len(content) >= 4 && bytes.Equal(content[:4], []byte{'P', 'K', 3, 4})
	if !match {
		return "", ErrMIMEMismatch
	}
	return declared, nil
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

type DevelopmentParser struct{}

func (DevelopmentParser) Parse(_ context.Context, mime string, content []byte) (ParsedDocument, error) {
	if mime != MIMEText {
		return ParsedDocument{}, ErrUnsupportedParsing
	}
	return ParsedDocument{Text: string(content), UntrustedContent: true}, nil
}
