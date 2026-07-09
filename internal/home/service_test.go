package home

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/zzm/opcv2/internal/competitor"
	"github.com/zzm/opcv2/internal/crm"
	"github.com/zzm/opcv2/internal/leads"
	"github.com/zzm/opcv2/internal/membership"
	"github.com/zzm/opcv2/internal/notifications"
	"github.com/zzm/opcv2/internal/sandbox"
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
	stats   tasks.Stats
	err     error
	userID  int64
	filters tasks.ListFilters
	statsID int64
}

func (f *fakeTasks) ListTasks(_ context.Context, userID int64, filters tasks.ListFilters) ([]tasks.Task, error) {
	f.userID = userID
	f.filters = filters
	return f.rows, f.err
}

func (f *fakeTasks) TaskStats(_ context.Context, userID int64) (tasks.Stats, error) {
	f.statsID = userID
	if f.err != nil {
		return tasks.Stats{}, f.err
	}
	return f.stats, nil
}

type fakeLeads struct {
	rows   []leads.Task
	err    error
	userID int64
	limit  int
}

func (f *fakeLeads) ListTasks(_ context.Context, userID int64, limit int) ([]leads.Task, error) {
	f.userID = userID
	f.limit = limit
	return f.rows, f.err
}

type fakeSandbox struct {
	rows   []sandbox.Session
	err    error
	userID int64
	limit  int
}

func (f *fakeSandbox) ListSessions(_ context.Context, userID int64, limit int) ([]sandbox.Session, error) {
	f.userID = userID
	f.limit = limit
	return f.rows, f.err
}

type fakeCompetitor struct {
	rows   []competitor.Scan
	err    error
	userID int64
	limit  int
}

func (f *fakeCompetitor) ListScans(_ context.Context, userID int64, limit int) ([]competitor.Scan, error) {
	f.userID = userID
	f.limit = limit
	return f.rows, f.err
}

type fakeCRM struct {
	rows    []crm.Customer
	stats   crm.PipelineStats
	err     error
	input   crm.ListDueInput
	statsID int64
}

func (f *fakeCRM) ListDueCustomers(_ context.Context, input crm.ListDueInput) ([]crm.Customer, error) {
	f.input = input
	return f.rows, f.err
}

func (f *fakeCRM) PipelineStats(_ context.Context, userID int64) (crm.PipelineStats, error) {
	f.statsID = userID
	if f.err != nil {
		return crm.PipelineStats{}, f.err
	}
	return f.stats, nil
}

func TestServiceBuildsSummaryFromDependencies(t *testing.T) {
	now := time.Date(2026, 7, 2, 10, 0, 0, 0, time.UTC)
	notificationReader := &fakeNotifications{summary: notifications.Summary{
		Unread: 2,
		Latest: []notifications.Notification{{ID: 7, Type: notifications.TypeTask, Title: "任务提醒", ActionURL: "/tasks", CreatedAt: now}},
	}}
	membershipReader := &fakeMembership{
		snapshot: membership.Snapshot{Plan: membership.Plan{Code: membership.PlanPro, Name: "会员版"}, CreditBalance: 88},
		usage:    []membership.UsageItem{{Key: "lead_tasks", Label: "AI线索任务", Used: 28, Limit: 30}},
	}
	taskReader := &fakeTasks{rows: []tasks.Task{
		{ID: 9, Title: "整理客户名单", Project: "AI线索开发", Status: tasks.StatusTodo, CreatedAt: now},
		{ID: 10, Title: "联调工作台", Project: "工作台", Status: tasks.StatusInProgress, CreatedAt: now},
	}, stats: tasks.Stats{Todo: 4, InProgress: 3}}
	leadReader := &fakeLeads{rows: []leads.Task{{ID: 11, Query: "成都 教培 私域转化", Status: leads.StatusSucceeded}}}
	sandboxReader := &fakeSandbox{rows: []sandbox.Session{{ID: 12, Goal: "验证 AI 低卡代餐奶昔", Status: sandbox.StatusCompleted}}}
	competitorReader := &fakeCompetitor{rows: []competitor.Scan{{ID: 13, Targets: []string{"小鹅通"}, Status: competitor.StatusRunning}}}
	crmReader := &fakeCRM{rows: []crm.Customer{
		{ID: 14, Name: "星河教育", Stage: crm.StageContacted},
		{ID: 15, Name: "星火咨询", Stage: crm.StageNew},
	}, stats: crm.PipelineStats{DueToday: 8}}
	service := NewService(Dependencies{
		Notifications: notificationReader,
		Membership:    membershipReader,
		Tasks:         taskReader,
		Leads:         leadReader,
		Sandbox:       sandboxReader,
		Competitor:    competitorReader,
		CRM:           crmReader,
	})

	summary, err := service.Summary(context.Background(), 42)

	if err != nil {
		t.Fatalf("Summary() error = %v", err)
	}
	if summary.NotificationSummary.Unread != 2 || summary.AccountSummary.PlanName != "会员版" || len(summary.RecentTasks) != 2 {
		t.Fatalf("summary = %+v", summary)
	}
	if len(summary.Metrics) != 3 || summary.Metrics[0].Label != "进行中任务" || summary.Metrics[0].Value != "3" || summary.Metrics[1].Label != "待办任务" || summary.Metrics[1].Value != "4" || summary.Metrics[2].Label != "今日跟进" || summary.Metrics[2].Value != "8" {
		t.Fatalf("metrics = %+v", summary.Metrics)
	}
	if len(summary.HeroCards) != 3 || len(summary.Recommendations) != 6 {
		t.Fatalf("cards/recommendations = %+v/%+v", summary.HeroCards, summary.Recommendations)
	}
	if len(summary.ActionItems) != 6 {
		t.Fatalf("action items = %+v", summary.ActionItems)
	}
	if summary.Recommendations[2].Title != "查看最新 AI 线索结果" || summary.Recommendations[3].URL != "/sandbox/sessions/12/report" || summary.Recommendations[4].Title != "查看竞品采集进度" || summary.Recommendations[5].Title != "跟进今日客户" {
		t.Fatalf("recommendations = %+v", summary.Recommendations)
	}
	if summary.Recommendations[5].Summary != "星河教育 等 2 位客户待跟进" {
		t.Fatalf("crm recommendation = %+v", summary.Recommendations[5])
	}
	if summary.ActionItems[0].Type != "membership" || summary.ActionItems[0].Priority != "high" || summary.ActionItems[0].CTA != "查看会员权益" {
		t.Fatalf("membership action item = %+v", summary.ActionItems[0])
	}
	if summary.ActionItems[2].Type != "leads" || summary.ActionItems[2].Priority != "medium" || summary.ActionItems[2].CTA != "查看线索" {
		t.Fatalf("lead action item = %+v", summary.ActionItems[2])
	}
	if summary.ActionItems[5].Type != "crm" || summary.ActionItems[5].Priority != "high" || summary.ActionItems[5].CTA != "去跟进" {
		t.Fatalf("crm action item = %+v", summary.ActionItems[5])
	}
	if notificationReader.userID != 42 || membershipReader.userID != 42 || taskReader.userID != 42 || taskReader.statsID != 42 || taskReader.filters.Limit != 5 || leadReader.limit != 1 || sandboxReader.limit != 1 || competitorReader.limit != 1 || crmReader.input.UserID != 42 || crmReader.input.Limit != 3 || crmReader.statsID != 42 {
		t.Fatalf("dependency calls = %d/%d/%d/%d/%d/%d/%d/%d/%d/%d/%d", notificationReader.userID, membershipReader.userID, taskReader.userID, taskReader.statsID, taskReader.filters.Limit, leadReader.limit, sandboxReader.limit, competitorReader.limit, crmReader.input.UserID, crmReader.input.Limit, crmReader.statsID)
	}
}

func TestServiceDegradesPartialDependencyFailures(t *testing.T) {
	service := NewService(Dependencies{
		Notifications: &fakeNotifications{err: errors.New("notifications down")},
		Membership:    &fakeMembership{err: errors.New("membership down")},
		Tasks:         &fakeTasks{err: errors.New("tasks down")},
		Leads:         &fakeLeads{err: errors.New("leads down")},
		Sandbox:       &fakeSandbox{err: errors.New("sandbox down")},
		Competitor:    &fakeCompetitor{err: errors.New("competitor down")},
		CRM:           &fakeCRM{err: errors.New("crm down")},
	})

	summary, err := service.Summary(context.Background(), 42)

	if err != nil {
		t.Fatalf("Summary() error = %v", err)
	}
	if summary.Metrics == nil || summary.HeroCards == nil || summary.Recommendations == nil || summary.ActionItems == nil || summary.RecentTasks == nil || summary.NotificationSummary.Latest == nil || summary.AccountSummary.QuotaWarnings == nil {
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
