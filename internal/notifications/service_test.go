package notifications

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeRepository struct {
	rows    []Notification
	row     Notification
	summary Summary
	filters ListFilters
	userID  int64
	id      int64
	updated int
	total   int
	err     error
}

func (r *fakeRepository) ListNotifications(_ context.Context, userID int64, filters ListFilters) ([]Notification, error) {
	r.userID = userID
	r.filters = filters
	return r.rows, r.err
}

func (r *fakeRepository) CountNotifications(_ context.Context, userID int64, filters ListFilters) (int, error) {
	r.userID = userID
	r.filters = filters
	return r.total, r.err
}

func (r *fakeRepository) GetNotification(_ context.Context, userID, id int64) (Notification, error) {
	r.userID = userID
	r.id = id
	return r.row, r.err
}

func (r *fakeRepository) MarkRead(_ context.Context, userID, id int64) (Notification, error) {
	r.userID = userID
	r.id = id
	return r.row, r.err
}

func (r *fakeRepository) MarkAllRead(_ context.Context, userID int64) (int, error) {
	r.userID = userID
	return r.updated, r.err
}

func (r *fakeRepository) DeleteNotification(_ context.Context, userID, id int64) error {
	r.userID = userID
	r.id = id
	return r.err
}

func (r *fakeRepository) Summary(_ context.Context, userID int64) (Summary, error) {
	r.userID = userID
	return r.summary, r.err
}

func TestServiceRejectsInvalidFilters(t *testing.T) {
	service := NewService(&fakeRepository{})

	_, err := service.ListNotifications(context.Background(), 42, ListFilters{Type: "unknown", Status: StatusUnread, Limit: 20})
	if !errors.Is(err, ErrInvalidFilter) {
		t.Fatalf("type err = %v, want ErrInvalidFilter", err)
	}

	_, err = service.ListNotifications(context.Background(), 42, ListFilters{Type: TypeTask, Status: "missing", Limit: 20})
	if !errors.Is(err, ErrInvalidFilter) {
		t.Fatalf("status err = %v, want ErrInvalidFilter", err)
	}
}

func TestServiceListsNotificationsWithCappedLimit(t *testing.T) {
	now := time.Date(2026, 7, 2, 10, 0, 0, 0, time.UTC)
	repository := &fakeRepository{rows: []Notification{{ID: 1, UserID: 42, Type: TypeTask, Title: "任务", CreatedAt: now}}, total: 27}
	service := NewService(repository)

	page, err := service.ListNotifications(context.Background(), 42, ListFilters{Type: TypeTask, Status: StatusUnread, Limit: 500, Offset: 20})

	if err != nil {
		t.Fatalf("ListNotifications() error = %v", err)
	}
	if repository.userID != 42 || repository.filters.Limit != 100 || repository.filters.Offset != 20 || len(page.Notifications) != 1 || page.Total != 27 || page.Limit != 100 || page.Offset != 20 {
		t.Fatalf("user/filters/page = %d/%+v/%+v", repository.userID, repository.filters, page)
	}
}

func TestServiceRejectsInvalidID(t *testing.T) {
	service := NewService(&fakeRepository{})

	_, err := service.GetNotification(context.Background(), 42, 0)
	if !errors.Is(err, ErrInvalidNotificationID) {
		t.Fatalf("get err = %v, want ErrInvalidNotificationID", err)
	}

	_, err = service.MarkRead(context.Background(), 42, -1)
	if !errors.Is(err, ErrInvalidNotificationID) {
		t.Fatalf("mark err = %v, want ErrInvalidNotificationID", err)
	}

	err = service.DeleteNotification(context.Background(), 42, 0)
	if !errors.Is(err, ErrInvalidNotificationID) {
		t.Fatalf("delete err = %v, want ErrInvalidNotificationID", err)
	}
}

func TestServiceDeletesNotificationForUser(t *testing.T) {
	repository := &fakeRepository{}
	service := NewService(repository)

	err := service.DeleteNotification(context.Background(), 42, 99)

	if err != nil {
		t.Fatalf("DeleteNotification() error = %v", err)
	}
	if repository.userID != 42 || repository.id != 99 {
		t.Fatalf("user/id = %d/%d", repository.userID, repository.id)
	}
}

func TestServiceMarkAllReadReturnsUpdatedCount(t *testing.T) {
	repository := &fakeRepository{updated: 3}
	service := NewService(repository)

	updated, err := service.MarkAllRead(context.Background(), 42)

	if err != nil {
		t.Fatalf("MarkAllRead() error = %v", err)
	}
	if updated != 3 || repository.userID != 42 {
		t.Fatalf("updated/userID = %d/%d", updated, repository.userID)
	}
}
