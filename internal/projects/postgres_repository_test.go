package projects

import (
	"context"
	"regexp"
	"testing"
	"time"

	pgxmock "github.com/pashagolub/pgxmock/v4"
)

func TestPostgresRepositoryGetsOpportunityWithStructuredBlocks(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 7, 17, 9, 0, 0, 0, time.UTC)
	db.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, slug, title, summary, industry, tags, budget_band, difficulty, resource_requirements, sections, status, published_at, updated_at
		FROM project_opportunities
		WHERE slug = $1 AND status = 'published'
	`)).
		WithArgs("ai-short-video-studio").
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "slug", "title", "summary", "industry", "tags", "budget_band", "difficulty", "resource_requirements", "sections", "status", "published_at", "updated_at",
		}).AddRow(
			int64(42), "ai-short-video-studio", "AI短视频脚本工作室", "短视频脚本服务", "内容服务",
			[]byte(`["内容创作"]`), "0.8-3万元", "中等", []byte(`["内容策划能力"]`),
			[]byte(`[{"key":"data","title":"当前数据","body":"演示数据","items":["复购率"],"blocks":[{"type":"metrics","title":"关键指标","columns":2,"items":[{"title":"复购率","value":"35.7%","tone":"positive","progress":35.7,"tags":["演示"]}],"series":[{"title":"5月","value":"128","progress":100}]}]}]`),
			OpportunityStatusPublished, &now, now,
		))

	repository := NewPostgresRepository(db)
	opportunity, err := repository.GetOpportunity(context.Background(), "ai-short-video-studio")
	if err != nil {
		t.Fatalf("GetOpportunity() error = %v", err)
	}
	if len(opportunity.Sections) != 1 || opportunity.Sections[0].Key != "data" || len(opportunity.Sections[0].Blocks) != 1 {
		t.Fatalf("sections = %+v", opportunity.Sections)
	}
	block := opportunity.Sections[0].Blocks[0]
	if block.Type != "metrics" || block.Columns != 2 || block.Items[0].Progress != 35.7 || block.Series[0].Value != "128" {
		t.Fatalf("block = %+v", block)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresRepositoryListsCasesByOpportunitySlug(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 7, 17, 9, 0, 0, 0, time.UTC)
	opportunityID := int64(42)
	db.ExpectQuery(regexp.QuoteMeta(`
		SELECT c.id, c.slug, c.opportunity_id,
		       COALESCE(o.slug, ''), COALESCE(o.title, ''), COALESCE(o.industry, ''),
		       c.title, c.summary, c.case_type, c.outcome, c.key_actions, c.lessons, c.pitfalls,
		       c.source_title, c.source_url, c.captured_at, c.status, c.published_at, c.updated_at
		FROM project_cases c
		LEFT JOIN project_opportunities o ON o.id = c.opportunity_id
		WHERE c.status = 'published'
		  AND ($1 = '' OR c.case_type = $1)
		  AND ($2 = '' OR c.opportunity_id = (
			  SELECT id
			  FROM project_opportunities
			  WHERE slug = $2 AND status = 'published'
		  ))
		  AND ($3 = '' OR o.industry = $3)
		ORDER BY c.published_at DESC, c.id ASC
		LIMIT $4
	`)).
		WithArgs("success", "ai-short-video-studio", "内容服务", 20).
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "slug", "opportunity_id", "opportunity_slug", "opportunity_title", "industry", "title", "summary", "case_type", "outcome", "key_actions", "lessons", "pitfalls", "source_title", "source_url", "captured_at", "status", "published_at", "updated_at",
		}).AddRow(
			int64(81), "short-video-first-client", &opportunityID, "ai-short-video-studio", "AI短视频脚本工作室", "内容服务",
			"首个客户", "案例摘要", "success", "完成验证",
			[]byte(`["限定行业"]`), []byte(`["沉淀模板"]`), []byte(`["避免承诺流量"]`),
			"演示来源", "https://example.com/case", now, CaseStatusPublished, &now, now,
		))

	repository := NewPostgresRepository(db)
	cases, err := repository.ListCases(context.Background(), CaseFilters{
		CaseType: "success", OpportunitySlug: "ai-short-video-studio", Industry: "内容服务", Limit: 20,
	})
	if err != nil {
		t.Fatalf("ListCases() error = %v", err)
	}
	if len(cases) != 1 || cases[0].OpportunityID == nil || *cases[0].OpportunityID != 42 ||
		cases[0].OpportunitySlug != "ai-short-video-studio" || cases[0].OpportunityTitle != "AI短视频脚本工作室" || cases[0].Industry != "内容服务" {
		t.Fatalf("cases = %+v", cases)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresRepositoryGetsCaseWithOpportunityMetadata(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 7, 17, 9, 0, 0, 0, time.UTC)
	opportunityID := int64(42)
	db.ExpectQuery(regexp.QuoteMeta(`
		SELECT c.id, c.slug, c.opportunity_id,
		       COALESCE(o.slug, ''), COALESCE(o.title, ''), COALESCE(o.industry, ''),
		       c.title, c.summary, c.case_type, c.outcome, c.key_actions, c.lessons, c.pitfalls,
		       c.source_title, c.source_url, c.captured_at, c.status, c.published_at, c.updated_at
		FROM project_cases c
		LEFT JOIN project_opportunities o ON o.id = c.opportunity_id
		WHERE c.slug = $1 AND c.status = 'published'
	`)).
		WithArgs("short-video-first-client").
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "slug", "opportunity_id", "opportunity_slug", "opportunity_title", "industry", "title", "summary", "case_type", "outcome", "key_actions", "lessons", "pitfalls", "source_title", "source_url", "captured_at", "status", "published_at", "updated_at",
		}).AddRow(
			int64(81), "short-video-first-client", &opportunityID, "ai-short-video-studio", "AI短视频脚本工作室", "内容服务",
			"首个客户", "案例摘要", "success", "完成验证", []byte(`[]`), []byte(`[]`), []byte(`[]`),
			"演示来源", "https://example.com/case", now, CaseStatusPublished, &now, now,
		))

	repository := NewPostgresRepository(db)
	item, err := repository.GetCase(context.Background(), "short-video-first-client")
	if err != nil {
		t.Fatalf("GetCase() error = %v", err)
	}
	if item.OpportunitySlug != "ai-short-video-studio" || item.OpportunityTitle != "AI短视频脚本工作室" || item.Industry != "内容服务" {
		t.Fatalf("case = %+v", item)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresRepositoryCreatesMatchSession(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 6, 24, 9, 0, 0, 0, time.UTC)
	db.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO project_match_sessions (user_id, intent, answers, status, questions, result, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $7)
		RETURNING id
	`)).
		WithArgs(
			int64(42),
			"我擅长内容创作，预算3万以内，每周20小时",
			[]byte(`[]`),
			StatusCompleted,
			[]byte(`[]`),
			pgxmock.AnyArg(),
			now,
		).
		WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(int64(99)))

	repository := NewPostgresRepository(db)
	session, err := repository.CreateSession(context.Background(), MatchSession{
		UserID:    42,
		Intent:    "我擅长内容创作，预算3万以内，每周20小时",
		Status:    StatusCompleted,
		Questions: []Question{},
		Result:    MatchResult{Status: StatusCompleted},
		CreatedAt: now,
	})
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}
	if session.ID != 99 {
		t.Fatalf("session.ID = %d, want 99", session.ID)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresRepositoryListsMatchSessionsForUser(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 6, 24, 9, 0, 0, 0, time.UTC)
	db.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, user_id, intent, answers, status, questions, result, created_at, updated_at
		FROM project_match_sessions
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`)).
		WithArgs(int64(42), 20).
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "user_id", "intent", "answers", "status", "questions", "result", "created_at", "updated_at",
		}).AddRow(
			int64(99),
			int64(42),
			"我的项目",
			[]byte(`[]`),
			StatusCompleted,
			[]byte(`[]`),
			[]byte(`{"status":"completed","projects":[{"rank":1,"title":"AI短视频脚本工作室","score":94}]}`),
			now,
			now,
		))

	repository := NewPostgresRepository(db)
	sessions, err := repository.ListSessions(context.Background(), 42, 20)
	if err != nil {
		t.Fatalf("ListSessions() error = %v", err)
	}
	if len(sessions) != 1 || sessions[0].ID != 99 || sessions[0].Result.Projects[0].Title != "AI短视频脚本工作室" {
		t.Fatalf("sessions = %+v", sessions)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresRepositoryReturnsEmptyMatchSessionList(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	db.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, user_id, intent, answers, status, questions, result, created_at, updated_at
		FROM project_match_sessions
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`)).
		WithArgs(int64(42), 20).
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "user_id", "intent", "answers", "status", "questions", "result", "created_at", "updated_at",
		}))

	repository := NewPostgresRepository(db)
	sessions, err := repository.ListSessions(context.Background(), 42, 20)
	if err != nil {
		t.Fatalf("ListSessions() error = %v", err)
	}
	if sessions == nil || len(sessions) != 0 {
		t.Fatalf("sessions = %#v, want non-nil empty slice", sessions)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresRepositoryGetsMatchSessionForUser(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 6, 24, 9, 0, 0, 0, time.UTC)
	db.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, user_id, intent, answers, status, questions, result, created_at, updated_at
		FROM project_match_sessions
		WHERE user_id = $1 AND id = $2
	`)).
		WithArgs(int64(42), int64(99)).
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "user_id", "intent", "answers", "status", "questions", "result", "created_at", "updated_at",
		}).AddRow(
			int64(99),
			int64(42),
			"我的项目",
			[]byte(`[{"key":"background","value":"内容创作"}]`),
			StatusCompleted,
			[]byte(`[]`),
			[]byte(`{"status":"completed","projects":[{"rank":1,"title":"AI短视频脚本工作室","score":94}]}`),
			now,
			now,
		))

	repository := NewPostgresRepository(db)
	session, err := repository.GetSession(context.Background(), 42, 99)
	if err != nil {
		t.Fatalf("GetSession() error = %v", err)
	}
	if session.ID != 99 || session.Answers[0].Value != "内容创作" || session.Result.Projects[0].Title != "AI短视频脚本工作室" {
		t.Fatalf("session = %+v", session)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresRepositorySavesFavoriteIdempotently(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 6, 24, 9, 0, 0, 0, time.UTC)
	db.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO project_match_favorites (user_id, session_id, created_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, session_id) DO UPDATE SET session_id = EXCLUDED.session_id
		RETURNING id, created_at
	`)).
		WithArgs(int64(42), int64(99), now).
		WillReturnRows(pgxmock.NewRows([]string{"id", "created_at"}).AddRow(int64(7), now))

	repository := NewPostgresRepository(db)
	favorite, err := repository.SaveFavorite(context.Background(), Favorite{
		UserID:    42,
		SessionID: 99,
		CreatedAt: now,
	})
	if err != nil {
		t.Fatalf("SaveFavorite() error = %v", err)
	}
	if favorite.ID != 7 || !favorite.CreatedAt.Equal(now) {
		t.Fatalf("favorite = %+v", favorite)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresRepositoryListsAndDeletesFavorites(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 6, 24, 9, 0, 0, 0, time.UTC)
	db.ExpectQuery("SELECT f\\.id, f\\.user_id, f\\.session_id, f\\.created_at,").
		WithArgs(int64(42), 20).
		WillReturnRows(pgxmock.NewRows([]string{
			"favorite_id", "favorite_user_id", "session_id", "favorite_created_at",
			"id", "user_id", "intent", "status", "questions", "result", "created_at", "updated_at",
		}).AddRow(
			int64(7), int64(42), int64(99), now,
			int64(99), int64(42), "线上轻资产", StatusCompleted, []byte(`[]`), []byte(`{"status":"completed","projects":[]}`), now, now,
		))
	db.ExpectExec(regexp.QuoteMeta(`DELETE FROM project_match_favorites WHERE user_id = $1 AND session_id = $2`)).
		WithArgs(int64(42), int64(99)).WillReturnResult(pgxmock.NewResult("DELETE", 1))

	repository := NewPostgresRepository(db)
	items, err := repository.ListFavorites(context.Background(), 42, 20)
	if err != nil || len(items) != 1 || items[0].Session == nil || items[0].Session.Intent != "线上轻资产" {
		t.Fatalf("favorites = %+v err=%v", items, err)
	}
	if err := repository.DeleteFavorite(context.Background(), 42, 99); err != nil {
		t.Fatalf("DeleteFavorite() error = %v", err)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
