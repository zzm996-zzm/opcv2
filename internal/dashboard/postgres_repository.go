package dashboard

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
)

type postgresDB interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

type PostgresRepository struct {
	db postgresDB
}

func NewPostgresRepository(db postgresDB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) GetSummary(ctx context.Context, userID int64) (Summary, error) {
	leadTotal, leadCompleted, err := r.leadStats(ctx, userID)
	if err != nil {
		return Summary{}, err
	}
	openTasks, completedTasks, err := r.taskStats(ctx, userID)
	if err != nil {
		return Summary{}, err
	}
	projects, err := r.projects(ctx, userID)
	if err != nil {
		return Summary{}, err
	}
	actions, err := r.actions(ctx, userID)
	if err != nil {
		return Summary{}, err
	}
	return ensureArrays(Summary{
		Metrics: []Metric{
			{Label: "本月收入", Value: "¥0", Change: "+0%"},
			{Label: "新增线索", Value: strconv.FormatInt(leadTotal, 10), Change: "+0%"},
			{Label: "成交客户", Value: strconv.FormatInt(leadCompleted, 10), Change: "+0%"},
			{Label: "待办任务", Value: strconv.FormatInt(openTasks, 10), Change: "-0%"},
		},
		Projects: projects,
		Trend: []TrendPoint{
			{Label: "周一", Value: 34},
			{Label: "周二", Value: 46},
			{Label: "周三", Value: 58},
			{Label: "周四", Value: 52},
			{Label: "周五", Value: 73},
			{Label: "周六", Value: 64},
			{Label: "周日", Value: 88},
		},
		Pipeline: []PipelineStage{
			{Stage: "线索", Count: strconv.FormatInt(leadTotal, 10), Percent: "100%"},
			{Stage: "已触达", Count: strconv.FormatInt(maxInt64(leadCompleted, 0), 10), Percent: percent(leadCompleted, leadTotal)},
			{Stage: "已演示", Count: strconv.FormatInt(maxInt64(leadCompleted/2, 0), 10), Percent: percent(leadCompleted/2, leadTotal)},
			{Stage: "成交", Count: strconv.FormatInt(leadCompleted, 10), Percent: percent(leadCompleted, leadTotal)},
		},
		Alerts: []Alert{
			{Title: "线索跟进延迟", Detail: fmt.Sprintf("%d 个任务仍待推进，请优先处理高意向客户。", openTasks)},
			{Title: "内容任务阻塞", Detail: fmt.Sprintf("已完成 %d 个任务，建议复盘可复用模板。", completedTasks)},
		},
		Actions: actions,
	}), nil
}

func (r *PostgresRepository) leadStats(ctx context.Context, userID int64) (int64, int64, error) {
	var total int64
	var completed int64
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*) FILTER (WHERE status IN ('queued', 'running', 'succeeded')) AS total_leads,
		       COUNT(*) FILTER (WHERE status = 'succeeded') AS completed_leads
		FROM lead_tasks
		WHERE user_id = $1
	`, userID).Scan(&total, &completed)
	return total, completed, err
}

func (r *PostgresRepository) taskStats(ctx context.Context, userID int64) (int64, int64, error) {
	var open int64
	var completed int64
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*) FILTER (WHERE status <> 'completed') AS open_tasks,
		       COUNT(*) FILTER (WHERE status = 'completed') AS completed_tasks
		FROM tasks
		WHERE user_id = $1
	`, userID).Scan(&open, &completed)
	return open, completed, err
}

func (r *PostgresRepository) projects(ctx context.Context, userID int64) ([]ProjectOpportunity, error) {
	rows, err := r.db.Query(ctx, `
		SELECT name, stage, source
		FROM crm_customers
		WHERE user_id = $1
		ORDER BY updated_at DESC
		LIMIT 3
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []ProjectOpportunity
	for rows.Next() {
		var name string
		var stage string
		var source string
		if err := rows.Scan(&name, &stage, &source); err != nil {
			return nil, err
		}
		projects = append(projects, ProjectOpportunity{
			Name:  name,
			Value: "待评估",
			Leads: source,
			Stage: stage,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return projects, nil
}

func (r *PostgresRepository) actions(ctx context.Context, userID int64) ([]Action, error) {
	rows, err := r.db.Query(ctx, `
		SELECT title, project, due_at
		FROM tasks
		WHERE user_id = $1 AND status <> 'completed'
		ORDER BY due_at NULLS LAST, created_at DESC
		LIMIT 3
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var actions []Action
	for rows.Next() {
		var title string
		var project string
		var dueAt *time.Time
		if err := rows.Scan(&title, &project, &dueAt); err != nil {
			return nil, err
		}
		actions = append(actions, Action{Time: actionTime(dueAt), Title: title})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return actions, nil
}

func actionTime(value *time.Time) string {
	if value == nil {
		return "待安排"
	}
	return value.Format("01-02 15:04")
}

func percent(value, total int64) string {
	if total <= 0 {
		return "0%"
	}
	return strconv.FormatInt(value*100/total, 10) + "%"
}

func maxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
