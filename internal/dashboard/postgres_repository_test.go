package dashboard

import (
	"context"
	"regexp"
	"testing"

	pgxmock "github.com/pashagolub/pgxmock/v4"
)

func TestPostgresRepositoryBuildsSummaryFromExistingTables(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	db.ExpectQuery(regexp.QuoteMeta(`
		SELECT COUNT(*) FILTER (WHERE status IN ('queued', 'running', 'succeeded')) AS total_leads,
		       COUNT(*) FILTER (WHERE status = 'succeeded') AS completed_leads
		FROM lead_tasks
		WHERE user_id = $1
	`)).
		WithArgs(int64(42)).
		WillReturnRows(pgxmock.NewRows([]string{"total_leads", "completed_leads"}).AddRow(int64(12), int64(7)))
	db.ExpectQuery(regexp.QuoteMeta(`
		SELECT COUNT(*) FILTER (WHERE status <> 'completed') AS open_tasks,
		       COUNT(*) FILTER (WHERE status = 'completed') AS completed_tasks
		FROM tasks
		WHERE user_id = $1
	`)).
		WithArgs(int64(42)).
		WillReturnRows(pgxmock.NewRows([]string{"open_tasks", "completed_tasks"}).AddRow(int64(5), int64(9)))
	db.ExpectQuery(regexp.QuoteMeta(`
		SELECT name, stage, source
		FROM crm_customers
		WHERE user_id = $1
		ORDER BY updated_at DESC
		LIMIT 3
	`)).
		WithArgs(int64(42)).
		WillReturnRows(pgxmock.NewRows([]string{"name", "stage", "source"}).AddRow("星桥教育集团", "qualified", "lead"))
	db.ExpectQuery(regexp.QuoteMeta(`
		SELECT title, project, due_at
		FROM tasks
		WHERE user_id = $1 AND status <> 'completed'
		ORDER BY due_at NULLS LAST, created_at DESC
		LIMIT 3
	`)).
		WithArgs(int64(42)).
		WillReturnRows(pgxmock.NewRows([]string{"title", "project", "due_at"}).AddRow("跟进星桥教育集团演示邀约", "AI线索开发", nil))

	repository := NewPostgresRepository(db)
	summary, err := repository.GetSummary(context.Background(), 42)

	if err != nil {
		t.Fatalf("GetSummary() error = %v", err)
	}
	if len(summary.Metrics) == 0 || summary.Metrics[1].Value != "12" {
		t.Fatalf("summary metrics = %+v", summary.Metrics)
	}
	if len(summary.Projects) != 1 || summary.Projects[0].Name != "星桥教育集团" {
		t.Fatalf("summary projects = %+v", summary.Projects)
	}
	if len(summary.Actions) != 1 || summary.Actions[0].Title == "" {
		t.Fatalf("summary actions = %+v", summary.Actions)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
