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

func TestPostgresRepositoryGetsAndUpsertsUserProgress(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 7, 11, 16, 0, 0, 0, time.UTC)
	columns := []string{"id", "user_id", "course_slug", "title", "percent", "last_lesson", "recommended_action", "updated_at"}
	db.ExpectQuery(regexp.QuoteMeta(`
		SELECT lp.id, lp.user_id, lp.course_slug, lc.title, lp.percent, lp.last_lesson, lp.recommended_action, lp.updated_at
		FROM learning_progress lp
		JOIN learning_courses lc ON lc.slug = lp.course_slug
		WHERE lp.user_id = $1 AND lp.course_slug = $2
	`)).WithArgs(int64(42), "ai-market-analysis").WillReturnRows(pgxmock.NewRows(columns).AddRow(
		int64(7), int64(42), "ai-market-analysis", "AI行业分析方法", 32, "2.2 行业生命周期", "继续第2章", now,
	))
	db.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO learning_progress (user_id, course_slug, percent, last_lesson, recommended_action, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $6)
		ON CONFLICT (user_id, course_slug) DO UPDATE SET
			percent = EXCLUDED.percent,
			last_lesson = EXCLUDED.last_lesson,
			recommended_action = EXCLUDED.recommended_action,
			updated_at = EXCLUDED.updated_at
		RETURNING id, user_id, course_slug, percent, last_lesson, recommended_action, updated_at
	`)).WithArgs(int64(42), "ai-market-analysis", 38, "2.3 行业规模", "继续第2章", now).WillReturnRows(
		pgxmock.NewRows([]string{"id", "user_id", "course_slug", "percent", "last_lesson", "recommended_action", "updated_at"}).AddRow(
			int64(7), int64(42), "ai-market-analysis", 38, "2.3 行业规模", "继续第2章", now,
		),
	)

	repository := NewPostgresRepository(db)
	progress, err := repository.GetProgress(context.Background(), 42, "ai-market-analysis")
	if err != nil || progress.CourseTitle != "AI行业分析方法" {
		t.Fatalf("GetProgress() progress = %+v err = %v", progress, err)
	}
	updated, err := repository.UpsertProgress(context.Background(), Progress{
		UserID: 42, CourseSlug: "ai-market-analysis", Percent: 38, LastLesson: "2.3 行业规模", RecommendedAction: "继续第2章", UpdatedAt: now,
	})
	if err != nil || updated.Percent != 38 {
		t.Fatalf("UpsertProgress() progress = %+v err = %v", updated, err)
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
		SELECT id, user_id, goal, project, focus_abilities, weekly_time, bottleneck, status, overall_score, dimensions, recommendations, created_at, updated_at
		FROM learning_diagnoses
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`)).
		WithArgs(int64(42)).
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "user_id", "goal", "project", "focus_abilities", "weekly_time", "bottleneck", "status", "overall_score", "dimensions", "recommendations", "created_at", "updated_at",
		}).AddRow(
			int64(99),
			int64(42),
			"提升AI能力",
			"智能客服",
			[]byte(`["数据洞察能力"]`),
			"5-8 小时",
			"缺少案例",
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
	if diagnosis.ID != 99 || diagnosis.UserID != 42 || diagnosis.Dimensions[0].Name == "" || diagnosis.FocusAbilities[0] != "数据洞察能力" || diagnosis.WeeklyTime != "5-8 小时" {
		t.Fatalf("diagnosis = %+v", diagnosis)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresRepositoryCreatesDiagnosisWithIntake(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 7, 11, 15, 0, 0, 0, time.UTC)
	db.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO learning_diagnoses (user_id, goal, project, focus_abilities, weekly_time, bottleneck, status, overall_score, dimensions, recommendations, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $11)
		RETURNING id, updated_at
	`)).WithArgs(
		int64(42),
		"提升AI能力",
		"智能客服",
		[]byte(`["数据洞察能力"]`),
		"5-8 小时",
		"缺少案例",
		DiagnosisCompleted,
		72,
		pgxmock.AnyArg(),
		pgxmock.AnyArg(),
		now,
	).WillReturnRows(pgxmock.NewRows([]string{"id", "updated_at"}).AddRow(int64(99), now))

	repository := NewPostgresRepository(db)
	diagnosis, err := repository.CreateDiagnosis(context.Background(), Diagnosis{
		UserID:         42,
		Goal:           "提升AI能力",
		Project:        "智能客服",
		FocusAbilities: []string{"数据洞察能力"},
		WeeklyTime:     "5-8 小时",
		Bottleneck:     "缺少案例",
		Status:         DiagnosisCompleted,
		OverallScore:   72,
		CreatedAt:      now,
	})
	if err != nil {
		t.Fatalf("CreateDiagnosis() error = %v", err)
	}
	if diagnosis.ID != 99 {
		t.Fatalf("diagnosis = %+v", diagnosis)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
