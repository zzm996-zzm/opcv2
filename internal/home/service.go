package home

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/zzm/opcv2/internal/competitor"
	"github.com/zzm/opcv2/internal/crm"
	"github.com/zzm/opcv2/internal/leads"
	"github.com/zzm/opcv2/internal/membership"
	"github.com/zzm/opcv2/internal/notifications"
	"github.com/zzm/opcv2/internal/sandbox"
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
	TaskStats(ctx context.Context, userID int64) (tasks.Stats, error)
}

type LeadTaskReader interface {
	ListTasks(ctx context.Context, userID int64, limit int) ([]leads.Task, error)
}

type SandboxReader interface {
	ListSessions(ctx context.Context, userID int64, limit int) ([]sandbox.Session, error)
}

type CompetitorReader interface {
	ListScans(ctx context.Context, userID int64, limit int) ([]competitor.Scan, error)
}

type CRMReader interface {
	ListDueCustomers(ctx context.Context, input crm.ListDueInput) ([]crm.Customer, error)
	PipelineStats(ctx context.Context, userID int64) (crm.PipelineStats, error)
}

type Dependencies struct {
	Notifications NotificationReader
	Membership    MembershipReader
	Tasks         TaskReader
	Leads         LeadTaskReader
	Sandbox       SandboxReader
	Competitor    CompetitorReader
	CRM           CRMReader
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
			{Label: "进行中任务", Value: "0", Icon: "folder"},
			{Label: "待办任务", Value: "0", Icon: "inbox"},
			{Label: "今日跟进", Value: "0", Icon: "trend"},
		},
		HeroCards: []Card{
			{Title: "项目确定及拆解", Summary: "洞察机会，精准定位，科学拆解", URL: "/projects"},
			{Title: "落地", Summary: "工具赋能，咨询陪跑，高效执行", URL: "/tasks"},
			{Title: "增长", Summary: "获客转化，客户运营，持续增长", URL: "/leads"},
		},
		Recommendations: []Card{},
		ActionItems:     []ActionItem{},
		RecentTasks:     []RecentTask{},
		NotificationSummary: NotificationSummary{
			ByType: []NotificationTypeCount{},
			Latest: []NotificationItem{},
		},
		AccountSummary: AccountSummary{
			QuotaWarnings: []QuotaWarning{},
		},
	}

	if s.deps.Notifications != nil {
		if notificationSummary, err := s.deps.Notifications.Summary(ctx, userID); err == nil {
			summary.NotificationSummary.Unread = notificationSummary.Unread
			summary.NotificationSummary.ByType = make([]NotificationTypeCount, 0, len(notificationSummary.ByType))
			for _, item := range notificationSummary.ByType {
				summary.NotificationSummary.ByType = append(summary.NotificationSummary.ByType, NotificationTypeCount{
					Type:  item.Type,
					Count: item.Count,
				})
			}
			summary.NotificationSummary.Latest = make([]NotificationItem, 0, len(notificationSummary.Latest))
			for _, item := range notificationSummary.Latest {
				summary.NotificationSummary.Latest = append(summary.NotificationSummary.Latest, NotificationItem{
					ID:          item.ID,
					Type:        item.Type,
					Title:       item.Title,
					Summary:     item.Summary,
					ActionLabel: item.ActionLabel,
					ActionURL:   item.ActionURL,
					ReadAt:      item.ReadAt,
					CreatedAt:   item.CreatedAt,
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
				card := Card{
					Title:   "会员额度即将用完",
					Summary: warning.Message,
					URL:     "/membership",
				}
				summary.Recommendations = append(summary.Recommendations, card)
				summary.ActionItems = append(summary.ActionItems, actionItemFromCard("membership", "high", "查看会员权益", card))
			}
		}
	}

	if s.deps.Tasks != nil {
		if rows, err := s.deps.Tasks.ListTasks(ctx, userID, tasks.ListFilters{Limit: 5}); err == nil {
			summary.RecentTasks = make([]RecentTask, 0, len(rows))
			todoCount := 0
			inProgressCount := 0
			now := time.Now()
			for _, task := range rows {
				summary.RecentTasks = append(summary.RecentTasks, RecentTask{
					ID:        task.ID,
					Title:     task.Title,
					Project:   task.Project,
					Status:    task.Status,
					Priority:  task.Priority,
					DueAt:     task.DueAt,
					IsOverdue: task.DueAt != nil && task.DueAt.Before(now) && task.Status != tasks.StatusCompleted,
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
			if stats, err := s.deps.Tasks.TaskStats(ctx, userID); err == nil {
				todoCount = stats.Todo
				inProgressCount = stats.InProgress
				summary.Metrics[0].Value = fmt.Sprintf("%d", stats.InProgress)
				summary.Metrics[1].Value = fmt.Sprintf("%d", stats.Todo)
			}
			if len(rows) == 0 {
				cards := []Card{
					{Title: "浏览项目超市", Summary: "先选择一个想验证的项目方向", URL: "/projects"},
					{Title: "打开商业沙盘", Summary: "拆解商业模式、成本和落地路径", URL: "/sandbox"},
				}
				summary.Recommendations = append(summary.Recommendations, cards...)
				summary.ActionItems = append(summary.ActionItems,
					actionItemFromCard("project", "medium", "去选择项目", cards[0]),
					actionItemFromCard("sandbox", "medium", "开始推演", cards[1]),
				)
			} else {
				card := Card{
					Title:   "继续推进任务",
					Summary: fmt.Sprintf("当前有 %d 个待办、%d 个进行中任务", todoCount, inProgressCount),
					URL:     "/tasks",
				}
				summary.Recommendations = append(summary.Recommendations, card)
				summary.ActionItems = append(summary.ActionItems, actionItemFromCard("task", "high", "查看任务", card))
			}
		}
	}
	if s.deps.Leads != nil {
		if rows, err := s.deps.Leads.ListTasks(ctx, userID, 1); err == nil && len(rows) > 0 {
			latest := rows[0]
			card := Card{
				Title:   leadRecommendationTitle(latest),
				Summary: fmt.Sprintf("最新线索任务：%s", latest.Query),
				URL:     "/leads",
			}
			summary.Recommendations = append(summary.Recommendations, card)
			summary.ActionItems = append(summary.ActionItems, actionItemFromCard("leads", leadRecommendationPriority(latest), leadRecommendationCTA(latest), card))
		}
	}
	if s.deps.Sandbox != nil {
		if rows, err := s.deps.Sandbox.ListSessions(ctx, userID, 1); err == nil && len(rows) > 0 {
			latest := rows[0]
			card := Card{
				Title:   sandboxRecommendationTitle(latest),
				Summary: latest.Goal,
				URL:     sandboxRecommendationURL(latest),
			}
			summary.Recommendations = append(summary.Recommendations, card)
			summary.ActionItems = append(summary.ActionItems, actionItemFromCard("sandbox", sandboxRecommendationPriority(latest), sandboxRecommendationCTA(latest), card))
		}
	}
	if s.deps.Competitor != nil {
		if rows, err := s.deps.Competitor.ListScans(ctx, userID, 1); err == nil && len(rows) > 0 {
			latest := rows[0]
			card := Card{
				Title:   competitorRecommendationTitle(latest),
				Summary: competitorRecommendationSummary(latest),
				URL:     "/competitor-data",
			}
			summary.Recommendations = append(summary.Recommendations, card)
			summary.ActionItems = append(summary.ActionItems, actionItemFromCard("competitor", competitorRecommendationPriority(latest), competitorRecommendationCTA(latest), card))
		}
	}
	if s.deps.CRM != nil {
		if stats, err := s.deps.CRM.PipelineStats(ctx, userID); err == nil {
			summary.Metrics[2].Value = fmt.Sprintf("%d", stats.DueToday)
		}
		if rows, err := s.deps.CRM.ListDueCustomers(ctx, crm.ListDueInput{UserID: userID, Limit: 3}); err == nil && len(rows) > 0 {
			card := Card{
				Title:   "跟进今日客户",
				Summary: crmRecommendationSummary(rows),
				URL:     "/crm",
			}
			summary.Recommendations = append(summary.Recommendations, card)
			summary.ActionItems = append(summary.ActionItems, actionItemFromCard("crm", "high", "去跟进", card))
		}
	}

	if summary.Metrics == nil {
		summary.Metrics = []Metric{}
	}
	if summary.RecentTasks == nil {
		summary.RecentTasks = []RecentTask{}
	}
	if summary.ActionItems == nil {
		summary.ActionItems = []ActionItem{}
	}
	if summary.NotificationSummary.Latest == nil {
		summary.NotificationSummary.Latest = []NotificationItem{}
	}
	if summary.NotificationSummary.ByType == nil {
		summary.NotificationSummary.ByType = []NotificationTypeCount{}
	}
	if summary.AccountSummary.QuotaWarnings == nil {
		summary.AccountSummary.QuotaWarnings = []QuotaWarning{}
	}
	return summary, nil
}

func actionItemFromCard(itemType, priority, cta string, card Card) ActionItem {
	return ActionItem{
		Type:     itemType,
		Priority: priority,
		Title:    card.Title,
		Summary:  card.Summary,
		URL:      card.URL,
		CTA:      cta,
	}
}

func leadRecommendationTitle(task leads.Task) string {
	switch task.Status {
	case leads.StatusSucceeded:
		return "查看最新 AI 线索结果"
	case leads.StatusFailed, leads.StatusRefunded:
		return "重新发起 AI 线索任务"
	default:
		return "查看 AI 线索采集进度"
	}
}

func leadRecommendationPriority(task leads.Task) string {
	switch task.Status {
	case leads.StatusSucceeded:
		return "medium"
	case leads.StatusFailed, leads.StatusRefunded:
		return "high"
	default:
		return "low"
	}
}

func leadRecommendationCTA(task leads.Task) string {
	switch task.Status {
	case leads.StatusSucceeded:
		return "查看线索"
	case leads.StatusFailed, leads.StatusRefunded:
		return "重新发起"
	default:
		return "查看进度"
	}
}

func sandboxRecommendationTitle(session sandbox.Session) string {
	if session.Status == sandbox.StatusCompleted {
		return "查看最新商业沙盘报告"
	}
	return "继续完成商业沙盘推演"
}

func sandboxRecommendationPriority(session sandbox.Session) string {
	if session.Status == sandbox.StatusCompleted {
		return "medium"
	}
	return "high"
}

func sandboxRecommendationCTA(session sandbox.Session) string {
	if session.Status == sandbox.StatusCompleted {
		return "查看报告"
	}
	return "继续推演"
}

func sandboxRecommendationURL(session sandbox.Session) string {
	if session.ID > 0 && session.Status == sandbox.StatusCompleted {
		return fmt.Sprintf("/sandbox/sessions/%d/report", session.ID)
	}
	return "/sandbox/start"
}

func competitorRecommendationTitle(scan competitor.Scan) string {
	switch scan.Status {
	case competitor.StatusSucceeded:
		return "查看最新竞品破解结论"
	case competitor.StatusFailed:
		return "重新发起竞品采集"
	default:
		return "查看竞品采集进度"
	}
}

func competitorRecommendationPriority(scan competitor.Scan) string {
	switch scan.Status {
	case competitor.StatusFailed:
		return "high"
	case competitor.StatusSucceeded:
		return "medium"
	default:
		return "low"
	}
}

func competitorRecommendationCTA(scan competitor.Scan) string {
	switch scan.Status {
	case competitor.StatusSucceeded:
		return "查看结论"
	case competitor.StatusFailed:
		return "重新采集"
	default:
		return "查看进度"
	}
}

func competitorRecommendationSummary(scan competitor.Scan) string {
	if len(scan.Targets) == 0 {
		return "竞品采集任务已更新"
	}
	return fmt.Sprintf("目标：%s", scan.Targets[0])
}

func crmRecommendationSummary(customers []crm.Customer) string {
	if len(customers) == 1 {
		return fmt.Sprintf("%s 已到跟进时间", customers[0].Name)
	}
	return fmt.Sprintf("%s 等 %d 位客户待跟进", customers[0].Name, len(customers))
}
