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
	if model.ID != 99 {
		t.Fatalf("model.ID = %d, want 99", model.ID)
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
