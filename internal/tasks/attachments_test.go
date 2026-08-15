package tasks

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeTaskAttachmentRepository struct {
	*fakeRepository
	items  []TaskAttachment
	nextID int64
}

type fakeTaskNotificationRepository struct {
	*fakeRepository
	count int
	limit int
}

func (r *fakeTaskNotificationRepository) DispatchTaskNotifications(_ context.Context, _ time.Time, limit int) (int, error) {
	r.limit = limit
	return r.count, r.err
}

func TestDispatchTaskNotificationsUsesBoundedWorkerBatch(t *testing.T) {
	repository := &fakeTaskNotificationRepository{fakeRepository: &fakeRepository{}, count: 3}
	service := NewService(repository)
	count, err := service.DispatchTaskNotifications(context.Background(), 999)
	if err != nil || count != 3 || repository.limit != 500 {
		t.Fatalf("count/error/limit = %d/%v/%d", count, err, repository.limit)
	}
}

func (r *fakeTaskAttachmentRepository) ListTaskAttachments(_ context.Context, userID, taskID int64) ([]TaskAttachment, error) {
	result := make([]TaskAttachment, 0)
	for _, item := range r.items {
		if item.UserID == userID && item.TaskID == taskID && item.DeletedAt == nil {
			result = append(result, item)
		}
	}
	return result, nil
}

func (r *fakeTaskAttachmentRepository) CreateTaskAttachment(_ context.Context, item TaskAttachment) (TaskAttachment, error) {
	r.nextID++
	item.ID = r.nextID
	r.items = append(r.items, item)
	return item, nil
}

func (r *fakeTaskAttachmentRepository) GetTaskAttachment(_ context.Context, userID, taskID, attachmentID int64) (TaskAttachment, error) {
	for _, item := range r.items {
		if item.ID == attachmentID && item.UserID == userID && item.TaskID == taskID && item.DeletedAt == nil {
			return item, nil
		}
	}
	return TaskAttachment{}, ErrTaskAttachmentNotFound
}

func (r *fakeTaskAttachmentRepository) DeleteTaskAttachment(_ context.Context, userID, taskID, attachmentID int64) error {
	for index, item := range r.items {
		if item.ID == attachmentID && item.UserID == userID && item.TaskID == taskID && item.DeletedAt == nil {
			now := time.Now()
			r.items[index].DeletedAt = &now
			return nil
		}
	}
	return ErrTaskAttachmentNotFound
}

type fakeAttachmentStorage struct {
	objects map[string][]byte
}

func (s *fakeAttachmentStorage) Put(_ context.Context, key string, content []byte) error {
	if s.objects == nil {
		s.objects = make(map[string][]byte)
	}
	s.objects[key] = append([]byte(nil), content...)
	return nil
}

func (s *fakeAttachmentStorage) Get(_ context.Context, key string) ([]byte, error) {
	content, ok := s.objects[key]
	if !ok {
		return nil, errors.New("missing object")
	}
	return append([]byte(nil), content...), nil
}

func (s *fakeAttachmentStorage) Delete(_ context.Context, key string) error {
	delete(s.objects, key)
	return nil
}

type unsafeAttachmentScanner struct{}

func (unsafeAttachmentScanner) Scan(context.Context, []byte) error {
	return errors.New("unsafe")
}

func TestTaskAttachmentUploadSignedDownloadAndDelete(t *testing.T) {
	repository := &fakeTaskAttachmentRepository{fakeRepository: &fakeRepository{}, nextID: 10}
	storage := &fakeAttachmentStorage{}
	now := time.Date(2026, time.August, 15, 8, 0, 0, 0, time.UTC)
	service := NewService(repository, WithTaskAttachmentStorage(storage, nil, "test-signing-key"))
	service.now = func() time.Time { return now }

	attachment, err := service.UploadTaskAttachment(context.Background(), CreateTaskAttachmentInput{
		UserID: 42, TaskID: 99, Name: "execution-notes.txt", MIMEType: "text/plain", Data: []byte("customer follow-up"),
	})
	if err != nil {
		t.Fatalf("UploadTaskAttachment() error = %v", err)
	}
	if attachment.ID != 11 || attachment.TaskID != 99 || attachment.StorageKey == "" || attachment.SHA256 == "" {
		t.Fatalf("attachment = %+v", attachment)
	}

	download, err := service.GetTaskAttachmentDownload(context.Background(), 42, 99, attachment.ID)
	if err != nil || download.URL == "" || !download.ExpiresAt.Equal(now.Add(TaskAttachmentURLTTL)) {
		t.Fatalf("download/error = %+v/%v", download, err)
	}
	expires := download.ExpiresAt.Unix()
	signature := service.attachmentSignature(42, 99, attachment.ID, expires)
	_, data, err := service.DownloadTaskAttachment(context.Background(), 42, 99, attachment.ID, expires, signature)
	if err != nil || string(data) != "customer follow-up" {
		t.Fatalf("data/error = %q/%v", data, err)
	}
	if _, _, err := service.DownloadTaskAttachment(context.Background(), 42, 99, attachment.ID, expires, "tampered"); !errors.Is(err, ErrTaskAttachmentSignatureInvalid) {
		t.Fatalf("tampered signature error = %v", err)
	}

	if err := service.DeleteTaskAttachment(context.Background(), 42, 99, attachment.ID); err != nil {
		t.Fatalf("DeleteTaskAttachment() error = %v", err)
	}
	if len(storage.objects) != 0 {
		t.Fatalf("storage objects = %v", storage.objects)
	}
}

func TestTaskAttachmentValidationRejectsUnsupportedUnsafeAndExpiredDownload(t *testing.T) {
	repository := &fakeTaskAttachmentRepository{fakeRepository: &fakeRepository{}}
	storage := &fakeAttachmentStorage{}
	service := NewService(repository, WithTaskAttachmentStorage(storage, unsafeAttachmentScanner{}, "test-signing-key"))

	_, err := service.UploadTaskAttachment(context.Background(), CreateTaskAttachmentInput{
		UserID: 42, TaskID: 99, Name: "payload.exe", MIMEType: "application/octet-stream", Data: []byte("binary"),
	})
	if !errors.Is(err, ErrUnsupportedTaskAttachmentType) {
		t.Fatalf("unsupported error = %v", err)
	}
	_, err = service.UploadTaskAttachment(context.Background(), CreateTaskAttachmentInput{
		UserID: 42, TaskID: 99, Name: "notes.txt", MIMEType: "text/plain", Data: []byte("unsafe"),
	})
	if !errors.Is(err, ErrUnsafeTaskAttachment) {
		t.Fatalf("unsafe error = %v", err)
	}

	repository.items = []TaskAttachment{{ID: 7, TaskID: 99, UserID: 42, StorageKey: "task-attachments/file", Name: "notes.txt", MIMEType: "text/plain"}}
	service.attachmentScanner = nil
	service.now = func() time.Time { return time.Unix(100, 0) }
	signature := service.attachmentSignature(42, 99, 7, 99)
	if _, _, err := service.DownloadTaskAttachment(context.Background(), 42, 99, 7, 99, signature); !errors.Is(err, ErrTaskAttachmentSignatureInvalid) {
		t.Fatalf("expired signature error = %v", err)
	}
}
