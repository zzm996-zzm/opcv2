package tasks

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"
	"unicode"
)

const (
	MaxTaskAttachmentSize  = 10 << 20
	MaxTaskAttachmentCount = 10
	TaskAttachmentURLTTL   = 5 * time.Minute
)

type AttachmentStorage interface {
	Put(context.Context, string, []byte) error
	Get(context.Context, string) ([]byte, error)
	Delete(context.Context, string) error
}

type AttachmentScanner interface {
	Scan(context.Context, []byte) error
}

type TaskAttachmentRepository interface {
	ListTaskAttachments(context.Context, int64, int64) ([]TaskAttachment, error)
	CreateTaskAttachment(context.Context, TaskAttachment) (TaskAttachment, error)
	GetTaskAttachment(context.Context, int64, int64, int64) (TaskAttachment, error)
	DeleteTaskAttachment(context.Context, int64, int64, int64) error
}

type AttachmentOption func(*Service)

func WithTaskAttachmentStorage(storage AttachmentStorage, scanner AttachmentScanner, signingKey string) Option {
	return func(service *Service) {
		service.attachmentStorage = storage
		service.attachmentScanner = scanner
		service.attachmentSigningKey = []byte(strings.TrimSpace(signingKey))
	}
}

func (s *Service) attachmentRepository() (TaskAttachmentRepository, error) {
	repository, ok := s.repository.(TaskAttachmentRepository)
	if !ok || s.attachmentStorage == nil {
		return nil, ErrServiceNotReady
	}
	return repository, nil
}

func (s *Service) ListTaskAttachments(ctx context.Context, userID, taskID int64) ([]TaskAttachment, error) {
	repository, err := s.attachmentRepository()
	if err != nil {
		return nil, err
	}
	return repository.ListTaskAttachments(ctx, userID, taskID)
}

func (s *Service) UploadTaskAttachment(ctx context.Context, input CreateTaskAttachmentInput) (TaskAttachment, error) {
	repository, err := s.attachmentRepository()
	if err != nil {
		return TaskAttachment{}, err
	}
	if input.UserID <= 0 || input.TaskID <= 0 || len(input.Data) == 0 {
		return TaskAttachment{}, ErrInvalidTaskAttachment
	}
	if len(input.Data) > MaxTaskAttachmentSize {
		return TaskAttachment{}, ErrTaskAttachmentTooLarge
	}
	name := sanitizeTaskAttachmentName(input.Name)
	if name == "" {
		return TaskAttachment{}, ErrInvalidTaskAttachment
	}
	mimeType, err := validateTaskAttachmentMIME(name, input.MIMEType, input.Data)
	if err != nil {
		return TaskAttachment{}, err
	}
	if s.attachmentScanner != nil {
		if err := s.attachmentScanner.Scan(ctx, input.Data); err != nil {
			return TaskAttachment{}, ErrUnsafeTaskAttachment
		}
	}
	items, err := repository.ListTaskAttachments(ctx, input.UserID, input.TaskID)
	if err != nil {
		return TaskAttachment{}, err
	}
	if len(items) >= MaxTaskAttachmentCount {
		return TaskAttachment{}, ErrTaskAttachmentCountExceeded
	}
	digest := sha256.Sum256(input.Data)
	checksum := hex.EncodeToString(digest[:])
	for _, item := range items {
		if item.SHA256 == checksum {
			return TaskAttachment{}, ErrInvalidTaskAttachment
		}
	}
	now := s.now()
	randomID := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, randomID); err != nil {
		return TaskAttachment{}, err
	}
	storageKey := fmt.Sprintf("task-attachments/%d/%d-%s", input.UserID, now.UnixNano(), hex.EncodeToString(randomID))
	if err := s.attachmentStorage.Put(ctx, storageKey, input.Data); err != nil {
		return TaskAttachment{}, err
	}
	attachment, err := repository.CreateTaskAttachment(ctx, TaskAttachment{
		TaskID: input.TaskID, CommentID: input.CommentID, UserID: input.UserID,
		Name: name, MIMEType: mimeType, SizeBytes: int64(len(input.Data)), StorageKey: storageKey,
		SHA256: checksum, CreatedAt: now,
	})
	if err != nil {
		_ = s.attachmentStorage.Delete(ctx, storageKey)
		return TaskAttachment{}, err
	}
	return attachment, nil
}

func (s *Service) GetTaskAttachmentDownload(ctx context.Context, userID, taskID, attachmentID int64) (TaskAttachmentDownload, error) {
	repository, err := s.attachmentRepository()
	if err != nil {
		return TaskAttachmentDownload{}, err
	}
	if len(s.attachmentSigningKey) == 0 {
		return TaskAttachmentDownload{}, ErrTaskAttachmentSignatureUnavailable
	}
	attachment, err := repository.GetTaskAttachment(ctx, userID, taskID, attachmentID)
	if err != nil {
		return TaskAttachmentDownload{}, err
	}
	expiresAt := s.now().Add(TaskAttachmentURLTTL)
	expires := expiresAt.Unix()
	signature := s.attachmentSignature(userID, taskID, attachmentID, expires)
	return TaskAttachmentDownload{
		Attachment: attachment,
		URL:        fmt.Sprintf("/api/v1/tasks/%d/attachments/%d/download?expires=%d&signature=%s", taskID, attachmentID, expires, signature),
		ExpiresAt:  expiresAt,
	}, nil
}

func (s *Service) DownloadTaskAttachment(ctx context.Context, userID, taskID, attachmentID, expires int64, signature string) (TaskAttachment, []byte, error) {
	repository, err := s.attachmentRepository()
	if err != nil {
		return TaskAttachment{}, nil, err
	}
	if len(s.attachmentSigningKey) == 0 || expires <= s.now().Unix() || !hmac.Equal([]byte(signature), []byte(s.attachmentSignature(userID, taskID, attachmentID, expires))) {
		return TaskAttachment{}, nil, ErrTaskAttachmentSignatureInvalid
	}
	attachment, err := repository.GetTaskAttachment(ctx, userID, taskID, attachmentID)
	if err != nil {
		return TaskAttachment{}, nil, err
	}
	data, err := s.attachmentStorage.Get(ctx, attachment.StorageKey)
	if err != nil {
		return TaskAttachment{}, nil, ErrTaskAttachmentNotFound
	}
	return attachment, data, nil
}

func (s *Service) DeleteTaskAttachment(ctx context.Context, userID, taskID, attachmentID int64) error {
	repository, err := s.attachmentRepository()
	if err != nil {
		return err
	}
	attachment, err := repository.GetTaskAttachment(ctx, userID, taskID, attachmentID)
	if err != nil {
		return err
	}
	if err := repository.DeleteTaskAttachment(ctx, userID, taskID, attachmentID); err != nil {
		return err
	}
	if err := s.attachmentStorage.Delete(ctx, attachment.StorageKey); err != nil {
		return err
	}
	return nil
}

func (s *Service) attachmentSignature(userID, taskID, attachmentID, expires int64) string {
	payload := fmt.Sprintf("%d:%d:%d:%d", userID, taskID, attachmentID, expires)
	digest := hmac.New(sha256.New, s.attachmentSigningKey)
	_, _ = digest.Write([]byte(payload))
	return hex.EncodeToString(digest.Sum(nil))
}

func sanitizeTaskAttachmentName(raw string) string {
	name := strings.TrimSpace(raw)
	if name == "" || name != filepath.Base(name) || strings.ContainsAny(name, "/\\\x00") {
		return ""
	}
	for _, r := range name {
		if unicode.IsControl(r) {
			return ""
		}
	}
	if len([]rune(name)) > 200 {
		return ""
	}
	return name
}

func validateTaskAttachmentMIME(name, declared string, content []byte) (string, error) {
	ext := strings.ToLower(filepath.Ext(name))
	expected := map[string]string{
		".txt": "text/plain", ".md": "text/markdown", ".markdown": "text/markdown", ".csv": "text/csv",
		".pdf": "application/pdf", ".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", ".pptx": "application/vnd.openxmlformats-officedocument.presentationml.presentation",
		".png": "image/png", ".jpg": "image/jpeg", ".jpeg": "image/jpeg",
	}[ext]
	if expected == "" {
		return "", ErrUnsupportedTaskAttachmentType
	}
	declared = strings.ToLower(strings.TrimSpace(strings.Split(declared, ";")[0]))
	if declared == "" || declared == "application/octet-stream" {
		declared = expected
	}
	if declared != expected {
		return "", ErrTaskAttachmentMIMEMismatch
	}
	detected := strings.ToLower(strings.TrimSpace(strings.Split(http.DetectContentType(content), ";")[0]))
	if expected == "text/plain" || expected == "text/markdown" || expected == "text/csv" {
		if bytes.Contains(content, []byte{0}) {
			return "", ErrTaskAttachmentMIMEMismatch
		}
		return expected, nil
	}
	if strings.HasPrefix(expected, "application/vnd.openxmlformats") {
		if !bytes.HasPrefix(content, []byte("PK\x03\x04")) || !validOfficeAttachment(content, expected) {
			return "", ErrTaskAttachmentMIMEMismatch
		}
		return expected, nil
	}
	if detected != expected {
		return "", ErrTaskAttachmentMIMEMismatch
	}
	return expected, nil
}

func validOfficeAttachment(content []byte, expected string) bool {
	reader, err := zip.NewReader(bytes.NewReader(content), int64(len(content)))
	if err != nil {
		return false
	}
	for _, file := range reader.File {
		switch {
		case expected == "application/vnd.openxmlformats-officedocument.wordprocessingml.document" && file.Name == "word/document.xml":
			return true
		case expected == "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet" && file.Name == "xl/workbook.xml":
			return true
		case expected == "application/vnd.openxmlformats-officedocument.presentationml.presentation" && file.Name == "ppt/presentation.xml":
			return true
		}
	}
	return false
}
