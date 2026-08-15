package growth

import (
	"context"
	"regexp"
	"testing"
	"time"

	pgxmock "github.com/pashagolub/pgxmock/v4"
)

func TestPostgresRepositoryCreatesModel(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 6, 30, 13, 0, 0, 0, time.UTC)
	db.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO growth_models (user_id, name, assumptions, result, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $5)
		RETURNING id
	`)).
		WithArgs(
			int64(42),
			"标准方案",
			pgxmock.AnyArg(),
			pgxmock.AnyArg(),
			now,
		).
		WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(int64(99)))

	repository := NewPostgresRepository(db)
	model, err := repository.CreateModel(context.Background(), Model{
		UserID:      42,
		Name:        "标准方案",
		Assumptions: Assumptions{MonthlyVisits: 24000, LeadRate: 0.068},
		Result:      Result{MonthlyRevenue: 186000},
		CreatedAt:   now,
	})
	if err != nil {
		t.Fatalf("CreateModel() error = %v", err)
	}
	if model.ID != 99 || !model.UpdatedAt.Equal(now) {
		t.Fatalf("model = %+v", model)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresRepositoryCreatesAndGetsOwnedDraft(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()
	now := time.Date(2026, 7, 11, 9, 0, 0, 0, time.UTC)
	db.ExpectQuery("INSERT INTO growth_drafts").
		WithArgs(int64(42), "企业培训", DraftStatusNeedsInput, pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), nil, now).
		WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(int64(71)))
	repository := NewPostgresRepository(db)
	draft, err := repository.CreateDraft(context.Background(), Draft{UserID: 42, Input: "企业培训", Status: DraftStatusNeedsInput, Answers: map[string]float64{}, CreatedAt: now})
	if err != nil || draft.ID != 71 {
		t.Fatalf("CreateDraft() = %+v, %v", draft, err)
	}

	db.ExpectQuery("SELECT id, user_id, input, status, assumptions, questions, answers, model_id, created_at, updated_at").
		WithArgs(int64(42), int64(71)).
		WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "input", "status", "assumptions", "questions", "answers", "model_id", "created_at", "updated_at"}).AddRow(
			int64(71), int64(42), "企业培训", DraftStatusNeedsInput, []byte(`{}`), []byte(`[]`), []byte(`{}`), nil, now, now,
		))
	draft, err = repository.GetDraft(context.Background(), 42, 71)
	if err != nil || draft.ID != 71 {
		t.Fatalf("GetDraft() = %+v, %v", draft, err)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresRepositoryCreatesAndListsSnapshots(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()
	now := time.Date(2026, 7, 11, 10, 0, 0, 0, time.UTC)
	db.ExpectQuery("INSERT INTO growth_model_snapshots").
		WithArgs(int64(42), int64(99), "企业培训", pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), pgxmock.AnyArg(), now).
		WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(int64(501)))
	repository := NewPostgresRepository(db)
	snapshot, err := repository.CreateSnapshot(context.Background(), ModelSnapshot{UserID: 42, ModelID: 99, ModelName: "企业培训", CreatedAt: now})
	if err != nil || snapshot.ID != 501 {
		t.Fatalf("CreateSnapshot() = %+v, %v", snapshot, err)
	}

	db.ExpectQuery("SELECT id, user_id, model_id, model_name, assumptions, result, scenarios, forecast, recommendations, created_at").
		WithArgs(int64(42), int64(99), 20).
		WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "model_id", "model_name", "assumptions", "result", "scenarios", "forecast", "recommendations", "created_at"}).AddRow(
			int64(501), int64(42), int64(99), "企业培训", []byte(`{}`), []byte(`{}`), []byte(`{"scenarios":[]}`), []byte(`{"months":[]}`), []byte(`{"action_items":[]}`), now,
		))
	snapshots, err := repository.ListSnapshots(context.Background(), 42, 99, 20)
	if err != nil || len(snapshots) != 1 || snapshots[0].ID != 501 {
		t.Fatalf("ListSnapshots() = %+v, %v", snapshots, err)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresRepositoryGetsOwnedModel(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 6, 30, 13, 0, 0, 0, time.UTC)
	db.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, user_id, name, assumptions, result, created_at, updated_at
		FROM growth_models
		WHERE user_id = $1 AND id = $2
	`)).
		WithArgs(int64(42), int64(99)).
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "user_id", "name", "assumptions", "result", "created_at", "updated_at",
		}).AddRow(
			int64(99),
			int64(42),
			"标准方案",
			[]byte(`{"monthly_visits":24000,"lead_rate":0.068,"deal_rate":0.14,"average_order":820,"acquisition_cost":42,"delivery_cost":51000}`),
			[]byte(`{"monthly_revenue":186000,"leads":1632,"deals":228,"payback_days":17,"net_margin":0.38}`),
			now,
			now,
		))

	repository := NewPostgresRepository(db)
	model, err := repository.GetModel(context.Background(), 42, 99)
	if err != nil {
		t.Fatalf("GetModel() error = %v", err)
	}
	if model.ID != 99 || model.Result.MonthlyRevenue != 186000 {
		t.Fatalf("model = %+v", model)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresRepositoryListsOwnedModelPage(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()
	now := time.Date(2026, 8, 15, 12, 0, 0, 0, time.UTC)
	db.ExpectQuery("SELECT COUNT\\(\\*\\)").
		WithArgs(int64(42), "SaaS").
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(21))
	db.ExpectQuery("SELECT id, user_id, name, assumptions, result, created_at, updated_at").
		WithArgs(int64(42), "SaaS", 10, 10).
		WillReturnRows(pgxmock.NewRows([]string{
			"id", "user_id", "name", "assumptions", "result", "created_at", "updated_at",
		}).AddRow(
			int64(99), int64(42), "SaaS 增长", []byte(`{}`), []byte(`{}`), now, now,
		))

	page, err := NewPostgresRepository(db).ListModelPage(context.Background(), ListModelsInput{
		UserID: 42, Query: "SaaS", Limit: 10, Offset: 10,
	})
	if err != nil {
		t.Fatalf("ListModelPage() error = %v", err)
	}
	if page.Total != 21 || page.Limit != 10 || page.Offset != 10 || len(page.Models) != 1 || page.Models[0].ID != 99 {
		t.Fatalf("page = %+v", page)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
