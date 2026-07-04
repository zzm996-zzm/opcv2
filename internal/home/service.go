package home

import (
	"context"
	"errors"
	"fmt"

	"github.com/zzm/opcv2/internal/membership"
	"github.com/zzm/opcv2/internal/notifications"
	"github.com/zzm/opcv2/internal/tasks"
)

var ErrUserIDRequired = errors.New("user id required")

type NotificationReader interface {
	Summary(ctx context.Context, userID int64) (notifications.Summary, error)
}

type MembershipReader interface {
	CurrentSnapshot(ctx context.Context, userID int64) (membership.Snapshot, error)
	CurrentUsage(ctx context.Context, userID int64) ([]membership.UsageItem, error)
}

type TaskReader interface {
	ListTasks(ctx context.Context, userID int64, filters tasks.ListFilters) ([]tasks.Task, error)
}

type Dependencies struct {
	Notifications NotificationReader
	Membership    MembershipReader
	Tasks         TaskReader
}

type Service struct {
	deps Dependencies
}

func NewService(deps Dependencies) *Service {
	return &Service{deps: deps}
}

func (s *Service) Summary(ctx context.Context, userID int64) (Summary, error) {
	if userID <= 0 {
		return Summary{}, ErrUserIDRequired
	}
	summary := Summary{
		HeroCards:       []Card{},
		Recommendations: []Card{},
		RecentTasks:     []RecentTask{},
		NotificationSummary: NotificationSummary{
			Latest: []NotificationItem{},
		},
		AccountSummary: AccountSummary{
			QuotaWarnings: []QuotaWarning{},
		},
	}

	if s.deps.Notifications != nil {
		if notificationSummary, err := s.deps.Notifications.Summary(ctx, userID); err == nil {
			summary.NotificationSummary.Unread = notificationSummary.Unread
			summary.NotificationSummary.Latest = make([]NotificationItem, 0, len(notificationSummary.Latest))
			for _, item := range notificationSummary.Latest {
				summary.NotificationSummary.Latest = append(summary.NotificationSummary.Latest, NotificationItem{
					ID:        item.ID,
					Type:      item.Type,
					Title:     item.Title,
					Summary:   item.Summary,
					ActionURL: item.ActionURL,
					CreatedAt: item.CreatedAt,
				})
			}
		}
	}

	if s.deps.Membership != nil {
		if snapshot, err := s.deps.Membership.CurrentSnapshot(ctx, userID); err == nil {
			summary.AccountSummary.PlanName = snapshot.Plan.Name
			summary.AccountSummary.CreditBalance = snapshot.CreditBalance
		}
		if usage, err := s.deps.Membership.CurrentUsage(ctx, userID); err == nil {
			for _, item := range usage {
				if item.Limit > 0 && item.Used*100/item.Limit >= 80 {
					summary.AccountSummary.QuotaWarnings = append(summary.AccountSummary.QuotaWarnings, QuotaWarning{
						Key:     item.Key,
						Label:   item.Label,
						Used:    item.Used,
						Limit:   item.Limit,
						Message: fmt.Sprintf("%s 已使用 %d/%d", item.Label, item.Used, item.Limit),
					})
				}
			}
		}
	}

	if s.deps.Tasks != nil {
		if rows, err := s.deps.Tasks.ListTasks(ctx, userID, tasks.ListFilters{Limit: 5}); err == nil {
			summary.RecentTasks = make([]RecentTask, 0, len(rows))
			for _, task := range rows {
				summary.RecentTasks = append(summary.RecentTasks, RecentTask{
					ID:      task.ID,
					Title:   task.Title,
					Project: task.Project,
					Status:  task.Status,
					DueAt:   task.DueAt,
				})
			}
		}
	}

	if summary.RecentTasks == nil {
		summary.RecentTasks = []RecentTask{}
	}
	if summary.NotificationSummary.Latest == nil {
		summary.NotificationSummary.Latest = []NotificationItem{}
	}
	if summary.AccountSummary.QuotaWarnings == nil {
		summary.AccountSummary.QuotaWarnings = []QuotaWarning{}
	}
	return summary, nil
}
