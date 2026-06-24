package crm

import (
	"context"
	"regexp"
	"testing"
	"time"

	pgxmock "github.com/pashagolub/pgxmock/v4"
)

func TestPostgresRepositoryImportsCustomerIdempotently(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 6, 24, 9, 0, 0, 0, time.UTC)
	db.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO crm_customers (user_id, import_key, name, phone, email, website, stage, source, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $9)
		ON CONFLICT (user_id, import_key) DO UPDATE SET import_key = EXCLUDED.import_key
		RETURNING id, user_id, import_key, name, phone, email, website, stage, source, next_follow_up_at, created_at, updated_at, xmax = 0 AS inserted
	`)).
		WithArgs(int64(42), "lead_result:99", "成都启明星教育", "028-12345678", "hello@example.com", "https://example.com", StageNew, SourceLead, now).
		WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "import_key", "name", "phone", "email", "website", "stage", "source", "next_follow_up_at", "created_at", "updated_at", "inserted"}).
			AddRow(int64(100), int64(42), "lead_result:99", "成都启明星教育", "028-12345678", "hello@example.com", "https://example.com", StageNew, SourceLead, nil, now, now, true))

	repository := NewPostgresRepository(db)
	customer, existed, err := repository.ImportCustomer(context.Background(), Customer{
		UserID:    42,
		ImportKey: "lead_result:99",
		Name:      "成都启明星教育",
		Phone:     "028-12345678",
		Email:     "hello@example.com",
		Website:   "https://example.com",
		Stage:     StageNew,
		Source:    SourceLead,
		CreatedAt: now,
	})
	if err != nil {
		t.Fatalf("ImportCustomer() error = %v", err)
	}
	if existed || customer.ID != 100 {
		t.Fatalf("customer=%+v existed=%v", customer, existed)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresRepositoryUpdatesStageAndActivityInTransaction(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 6, 24, 10, 0, 0, 0, time.UTC)
	db.ExpectBegin()
	db.ExpectQuery(regexp.QuoteMeta(`
		UPDATE crm_customers
		SET stage = $3, updated_at = $4
		WHERE user_id = $1 AND id = $2
		RETURNING id, user_id, import_key, name, phone, email, website, stage, source, next_follow_up_at, created_at, updated_at
	`)).
		WithArgs(int64(42), int64(100), StageContacted, now).
		WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "import_key", "name", "phone", "email", "website", "stage", "source", "next_follow_up_at", "created_at", "updated_at"}).
			AddRow(int64(100), int64(42), "lead_result:99", "成都启明星教育", "", "", "", StageContacted, SourceLead, nil, now, now))
	db.ExpectExec(regexp.QuoteMeta(`
		INSERT INTO crm_activities (user_id, customer_id, type, note, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`)).
		WithArgs(int64(42), int64(100), ActivityStageChanged, "电话已接通", now).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))
	db.ExpectCommit()

	repository := NewPostgresRepository(db)
	customer, err := repository.UpdateStage(context.Background(), 42, 100, StageContacted, Activity{
		UserID:     42,
		CustomerID: 100,
		Type:       ActivityStageChanged,
		Note:       "电话已接通",
		CreatedAt:  now,
	})
	if err != nil {
		t.Fatalf("UpdateStage() error = %v", err)
	}
	if customer.Stage != StageContacted {
		t.Fatalf("customer = %+v", customer)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresRepositoryListsDueCustomersForUser(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 6, 24, 12, 0, 0, 0, time.UTC)
	db.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, user_id, import_key, name, phone, email, website, stage, source, next_follow_up_at, created_at, updated_at
		FROM crm_customers
		WHERE user_id = $1 AND next_follow_up_at IS NOT NULL AND next_follow_up_at <= $2
		ORDER BY next_follow_up_at ASC, id ASC
		LIMIT $3
	`)).
		WithArgs(int64(42), now, 20).
		WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "import_key", "name", "phone", "email", "website", "stage", "source", "next_follow_up_at", "created_at", "updated_at"}).
			AddRow(int64(100), int64(42), "lead_result:99", "成都启明星教育", "", "", "", StageContacted, SourceLead, now.Add(-time.Hour), now, now))

	repository := NewPostgresRepository(db)
	customers, err := repository.ListDueCustomers(context.Background(), 42, now, 20)
	if err != nil {
		t.Fatalf("ListDueCustomers() error = %v", err)
	}
	if len(customers) != 1 || customers[0].ID != 100 {
		t.Fatalf("customers = %+v", customers)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
