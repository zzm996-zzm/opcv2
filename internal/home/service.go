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
		Metrics: []Metric{
			{Label: "进行中项目", Value: "0", Icon: "folder"},
			{Label: "待办事项", Value: "0", Icon: "inbox"},
			{Label: "额度预警", Value: "0", Icon: "trend"},
		},
		HeroCards: []Card{
			{Title: "项目确定及拆解", Summary: "洞察机会，精准定位，科学拆解", URL: "/projects"},
			{Title: "落地", Summary: "工具赋能，咨询陪跑，高效执行", URL: "/tasks"},
			{Title: "增长", Summary: "获客转化，客户运营，持续增长", URL: "/leads"},
		},
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
			if len(summary.AccountSummary.QuotaWarnings) > 0 {
				warning := summary.AccountSummary.QuotaWarnings[0]
				summary.Recommendations = append(summary.Recommendations, Card{
					Title:   "会员额度即将用完",
					Summary: warning.Message,
					URL:     "/membership",
				})
			}
		}
	}

	if s.deps.Tasks != nil {
		if rows, err := s.deps.Tasks.ListTasks(ctx, userID, tasks.ListFilters{Limit: 5}); err == nil {
			summary.RecentTasks = make([]RecentTask, 0, len(rows))
			todoCount := 0
			inProgressCount := 0
			for _, task := range rows {
				summary.RecentTasks = append(summary.RecentTasks, RecentTask{
					ID:      task.ID,
					Title:   task.Title,
					Project: task.Project,
					Status:  task.Status,
					DueAt:   task.DueAt,
				})
				switch task.Status {
				case tasks.StatusTodo:
					todoCount++
				case tasks.StatusInProgress:
					inProgressCount++
				}
			}
			summary.Metrics[0].Value = fmt.Sprintf("%d", inProgressCount)
			summary.Metrics[1].Value = fmt.Sprintf("%d", todoCount)
			if len(rows) == 0 {
				summary.Recommendations = append(summary.Recommendations,
					Card{Title: "浏览项目超市", Summary: "先选择一个想验证的项目方向", URL: "/projects"},
					Card{Title: "打开商业沙盘", Summary: "拆解商业模式、成本和落地路径", URL: "/sandbox"},
				)
			} else {
				summary.Recommendations = append(summary.Recommendations, Card{
					Title:   "继续推进任务",
					Summary: fmt.Sprintf("当前有 %d 个待办、%d 个进行中任务", todoCount, inProgressCount),
					URL:     "/tasks",
				})
			}
		}
	}
	summary.Metrics[2].Value = fmt.Sprintf("%d", len(summary.AccountSummary.QuotaWarnings))

	if summary.Metrics == nil {
		summary.Metrics = []Metric{}
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
