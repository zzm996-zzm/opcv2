package home

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/zzm/opcv2/internal/membership"
	"github.com/zzm/opcv2/internal/notifications"
	"github.com/zzm/opcv2/internal/tasks"
)

type fakeNotifications struct {
	summary notifications.Summary
	err     error
	userID  int64
}

func (f *fakeNotifications) Summary(_ context.Context, userID int64) (notifications.Summary, error) {
	f.userID = userID
	return f.summary, f.err
}

type fakeMembership struct {
	snapshot membership.Snapshot
	usage    []membership.UsageItem
	err      error
	userID   int64
}

func (f *fakeMembership) CurrentSnapshot(_ context.Context, userID int64) (membership.Snapshot, error) {
	f.userID = userID
	return f.snapshot, f.err
}

func (f *fakeMembership) CurrentUsage(_ context.Context, userID int64) ([]membership.UsageItem, error) {
	f.userID = userID
	return f.usage, f.err
}

type fakeTasks struct {
	rows    []tasks.Task
	err     error
	userID  int64
	filters tasks.ListFilters
}

func (f *fakeTasks) ListTasks(_ context.Context, userID int64, filters tasks.ListFilters) ([]tasks.Task, error) {
	f.userID = userID
	f.filters = filters
	return f.rows, f.err
}

func TestServiceBuildsSummaryFromDependencies(t *testing.T) {
	now := time.Date(2026, 7, 2, 10, 0, 0, 0, time.UTC)
	notificationReader := &fakeNotifications{summary: notifications.Summary{
		Unread: 2,
		Latest: []notifications.Notification{{ID: 7, Type: notifications.TypeTask, Title: "任务提醒", ActionURL: "/tasks", CreatedAt: now}},
	}}
	membershipReader := &fakeMembership{
		snapshot: membership.Snapshot{Plan: membership.Plan{Code: membership.PlanPro, Name: "会员版"}, CreditBalance: 88},
		usage:    []membership.UsageItem{{Key: "lead_tasks", Label: "AI线索任务", Used: 8, Limit: 30}},
	}
	taskReader := &fakeTasks{rows: []tasks.Task{{ID: 9, Title: "整理客户名单", Project: "AI线索开发", Status: tasks.StatusTodo, CreatedAt: now}}}
	service := NewService(Dependencies{Notifications: notificationReader, Membership: membershipReader, Tasks: taskReader})

	summary, err := service.Summary(context.Background(), 42)

	if err != nil {
		t.Fatalf("Summary() error = %v", err)
	}
	if summary.NotificationSummary.Unread != 2 || summary.AccountSummary.PlanName != "会员版" || len(summary.RecentTasks) != 1 {
		t.Fatalf("summary = %+v", summary)
	}
	if notificationReader.userID != 42 || membershipReader.userID != 42 || taskReader.userID != 42 || taskReader.filters.Limit != 5 {
		t.Fatalf("dependency calls = %d/%d/%d/%d", notificationReader.userID, membershipReader.userID, taskReader.userID, taskReader.filters.Limit)
	}
}

func TestServiceDegradesPartialDependencyFailures(t *testing.T) {
	service := NewService(Dependencies{
		Notifications: &fakeNotifications{err: errors.New("notifications down")},
		Membership:    &fakeMembership{err: errors.New("membership down")},
		Tasks:         &fakeTasks{err: errors.New("tasks down")},
	})

	summary, err := service.Summary(context.Background(), 42)

	if err != nil {
		t.Fatalf("Summary() error = %v", err)
	}
	if summary.HeroCards == nil || summary.Recommendations == nil || summary.RecentTasks == nil || summary.NotificationSummary.Latest == nil || summary.AccountSummary.QuotaWarnings == nil {
		t.Fatalf("summary should contain safe empty slices: %+v", summary)
	}
}

func TestServiceRejectsMissingUserID(t *testing.T) {
	service := NewService(Dependencies{})

	_, err := service.Summary(context.Background(), 0)

	if !errors.Is(err, ErrUserIDRequired) {
		t.Fatalf("err = %v, want ErrUserIDRequired", err)
	}
}
