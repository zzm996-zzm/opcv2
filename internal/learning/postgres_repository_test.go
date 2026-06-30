package learning

import (
	"context"
	"regexp"
	"testing"
	"time"

	pgxmock "github.com/pashagolub/pgxmock/v4"
)

func TestPostgresRepositoryListsProgressForUser(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 6, 30, 15, 0, 0, 0, time.UTC)
	db.ExpectQuery(regexp.QuoteMeta(`
		SELECT lp.id, lp.user_id, lp.course_slug, lc.title, lp.percent, lp.last_lesson, lp.recommended_action, lp.updated_at
		FROM learning_progress lp
		JOIN learning_courses lc ON lc.slug = lp.course_slug
		WHERE lp.user_id = $1
		ORDER BY lp.updated_at DESC
	`)).
		WithArgs(int64(42)).
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "user_id", "course_slug", "title", "percent", "last_lesson", "recommended_action", "updated_at",
		}).AddRow(
			int64(7),
			int64(42),
			"ai-basics",
			"AI基础入门",
			68,
			"市场判断框架",
			"继续学习提示词工程",
			now,
		))

	repository := NewPostgresRepository(db)
	progress, err := repository.ListProgress(context.Background(), 42)
	if err != nil {
		t.Fatalf("ListProgress() error = %v", err)
	}
	if len(progress) != 1 || progress[0].UserID != 42 || progress[0].CourseSlug != "ai-basics" {
		t.Fatalf("progress = %+v", progress)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresRepositoryGetsLatestDiagnosisForUser(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 6, 30, 15, 0, 0, 0, time.UTC)
	db.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, user_id, goal, project, status, overall_score, dimensions, recommendations, created_at, updated_at
		FROM learning_diagnoses
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`)).
		WithArgs(int64(42)).
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "user_id", "goal", "project", "status", "overall_score", "dimensions", "recommendations", "created_at", "updated_at",
		}).AddRow(
			int64(99),
			int64(42),
			"提升AI能力",
			"智能客服",
			DiagnosisCompleted,
			72,
			[]byte(`[{"name":"市场分析能力","score":78,"gap":12,"summary":"具备基础判断能力"}]`),
			[]byte(`["优先学习 AI行业分析方法"]`),
			now,
			now,
		))

	repository := NewPostgresRepository(db)
	diagnosis, err := repository.LatestDiagnosis(context.Background(), 42)
	if err != nil {
		t.Fatalf("LatestDiagnosis() error = %v", err)
	}
	if diagnosis.ID != 99 || diagnosis.UserID != 42 || diagnosis.Dimensions[0].Name == "" {
		t.Fatalf("diagnosis = %+v", diagnosis)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
