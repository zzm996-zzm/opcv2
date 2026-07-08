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

func TestPostgresRepositoryListsCustomersWithFilters(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 6, 24, 10, 0, 0, 0, time.UTC)
	db.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, user_id, import_key, name, phone, email, website, stage, source, next_follow_up_at, created_at, updated_at
		FROM crm_customers
		WHERE user_id = $1
			AND ($2 = '' OR stage = $2)
			AND ($3 = '' OR source = $3)
			AND (
				$4 = ''
				OR name ILIKE '%' || $4 || '%'
				OR phone ILIKE '%' || $4 || '%'
				OR email ILIKE '%' || $4 || '%'
				OR website ILIKE '%' || $4 || '%'
			)
		ORDER BY updated_at DESC, id DESC
		LIMIT $5
	`)).
		WithArgs(int64(42), StageContacted, SourceEnterprise, "启明星", 20).
		WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "import_key", "name", "phone", "email", "website", "stage", "source", "next_follow_up_at", "created_at", "updated_at"}).
			AddRow(int64(100), int64(42), "enterprise_diagnosis_request:8", "成都启明星教育", "028-12345678", "", "", StageContacted, SourceEnterprise, nil, now, now))

	repository := NewPostgresRepository(db)
	customers, err := repository.ListCustomers(context.Background(), ListCustomersInput{UserID: 42, Stage: StageContacted, Source: SourceEnterprise, Q: "启明星", Limit: 20})
	if err != nil {
		t.Fatalf("ListCustomers() error = %v", err)
	}
	if len(customers) != 1 || customers[0].ID != 100 {
		t.Fatalf("customers = %+v", customers)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresRepositoryUpdatesCustomerAndActivityInTransaction(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 6, 24, 10, 30, 0, 0, time.UTC)
	name := "成都启明星教育"
	phone := "028-12345678"
	db.ExpectBegin()
	db.ExpectQuery(regexp.QuoteMeta(`
		UPDATE crm_customers
		SET
			name = COALESCE($3, name),
			phone = COALESCE($4, phone),
			email = COALESCE($5, email),
			website = COALESCE($6, website),
			updated_at = $7
		WHERE user_id = $1 AND id = $2
		RETURNING id, user_id, import_key, name, phone, email, website, stage, source, next_follow_up_at, created_at, updated_at
	`)).
		WithArgs(int64(42), int64(100), name, phone, nil, nil, now).
		WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "import_key", "name", "phone", "email", "website", "stage", "source", "next_follow_up_at", "created_at", "updated_at"}).
			AddRow(int64(100), int64(42), "lead_result:99", name, phone, "", "", StageContacted, SourceLead, nil, now, now))
	db.ExpectExec(regexp.QuoteMeta(`
		INSERT INTO crm_activities (user_id, customer_id, type, note, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`)).
		WithArgs(int64(42), int64(100), ActivityCustomerUpdated, "客户资料已更新", now).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))
	db.ExpectCommit()

	repository := NewPostgresRepository(db)
	customer, err := repository.UpdateCustomer(context.Background(), UpdateCustomerInput{
		UserID:     42,
		CustomerID: 100,
		Name:       &name,
		Phone:      &phone,
	}, Activity{
		UserID:     42,
		CustomerID: 100,
		Type:       ActivityCustomerUpdated,
		Note:       "客户资料已更新",
		CreatedAt:  now,
	})
	if err != nil {
		t.Fatalf("UpdateCustomer() error = %v", err)
	}
	if customer.Name != name || customer.Phone != phone {
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

func TestPostgresRepositoryListsFollowUpsForUser(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 6, 24, 12, 0, 0, 0, time.UTC)
	db.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, user_id, customer_id, note, next_follow_up_at, created_at
		FROM crm_followups
		WHERE user_id = $1
			AND ($2 = 0 OR customer_id = $2)
			AND ($4 = '' OR note ILIKE '%' || $4 || '%' OR customer_id::text = $4)
			AND (NOT $5 OR next_follow_up_at >= $6)
			AND (NOT $7 OR next_follow_up_at < $8)
		ORDER BY next_follow_up_at ASC, created_at DESC, id DESC
		LIMIT $3
	`)).
		WithArgs(int64(42), int64(100), 20, "方案", true, now.Add(-time.Hour), true, now.Add(24*time.Hour)).
		WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "customer_id", "note", "next_follow_up_at", "created_at"}).
			AddRow(int64(1), int64(42), int64(100), "发送方案", now, now))

	repository := NewPostgresRepository(db)
	followUps, err := repository.ListFollowUps(context.Background(), ListFollowUpsInput{
		UserID:       42,
		CustomerID:   100,
		Q:            "方案",
		HasDueFrom:   true,
		DueFrom:      now.Add(-time.Hour),
		HasDueBefore: true,
		DueBefore:    now.Add(24 * time.Hour),
		Limit:        20,
	})
	if err != nil {
		t.Fatalf("ListFollowUps() error = %v", err)
	}
	if len(followUps) != 1 || followUps[0].Note != "发送方案" {
		t.Fatalf("followUps = %+v", followUps)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresRepositoryReschedulesFollowUp(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	createdAt := time.Date(2026, 6, 24, 12, 0, 0, 0, time.UTC)
	next := time.Date(2026, 6, 27, 15, 0, 0, 0, time.UTC)
	activityAt := time.Date(2026, 6, 24, 13, 0, 0, 0, time.UTC)
	db.ExpectBegin()
	db.ExpectQuery(regexp.QuoteMeta(`
		UPDATE crm_followups
		SET next_follow_up_at = $3
		WHERE user_id = $1 AND id = $2
		RETURNING id, user_id, customer_id, note, next_follow_up_at, created_at
	`)).
		WithArgs(int64(42), int64(9), next).
		WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "customer_id", "note", "next_follow_up_at", "created_at"}).
			AddRow(int64(9), int64(42), int64(100), "发送方案", next, createdAt))
	db.ExpectExec(regexp.QuoteMeta(`
		UPDATE crm_customers
		SET next_follow_up_at = $3, updated_at = $4
		WHERE user_id = $1 AND id = $2
	`)).
		WithArgs(int64(42), int64(100), next, activityAt).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))
	db.ExpectExec(regexp.QuoteMeta(`
		INSERT INTO crm_activities (user_id, customer_id, type, note, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`)).
		WithArgs(int64(42), int64(100), ActivityFollowUpRescheduled, "跟进时间已改期", activityAt).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))
	db.ExpectCommit()

	repository := NewPostgresRepository(db)
	followUp, err := repository.RescheduleFollowUp(context.Background(), 42, 9, next, Activity{
		UserID:    42,
		Type:      ActivityFollowUpRescheduled,
		Note:      "跟进时间已改期",
		CreatedAt: activityAt,
	})
	if err != nil {
		t.Fatalf("RescheduleFollowUp() error = %v", err)
	}
	if followUp.ID != 9 || followUp.CustomerID != 100 || !followUp.NextFollowUpAt.Equal(next) {
		t.Fatalf("followUp = %+v", followUp)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresRepositoryListsActivitiesForCustomer(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 6, 24, 12, 0, 0, 0, time.UTC)
	db.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, user_id, customer_id, type, note, created_at
		FROM crm_activities
		WHERE user_id = $1 AND customer_id = $2
		ORDER BY created_at DESC, id DESC
		LIMIT $3
	`)).
		WithArgs(int64(42), int64(100), 20).
		WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "customer_id", "type", "note", "created_at"}).
			AddRow(int64(1), int64(42), int64(100), ActivityCustomerUpdated, "客户资料已更新", now))

	repository := NewPostgresRepository(db)
	activities, err := repository.ListActivities(context.Background(), 42, 100, 20)
	if err != nil {
		t.Fatalf("ListActivities() error = %v", err)
	}
	if len(activities) != 1 || activities[0].Type != ActivityCustomerUpdated {
		t.Fatalf("activities = %+v", activities)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresRepositoryReturnsPipelineStats(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 6, 24, 12, 0, 0, 0, time.UTC)
	db.ExpectQuery(regexp.QuoteMeta(`
		SELECT
			COUNT(*)::int,
			COUNT(*) FILTER (WHERE stage = 'new')::int,
			COUNT(*) FILTER (WHERE stage = 'contacted')::int,
			COUNT(*) FILTER (WHERE stage = 'qualified')::int,
			COUNT(*) FILTER (WHERE stage = 'proposal')::int,
			COUNT(*) FILTER (WHERE stage = 'won')::int,
			COUNT(*) FILTER (WHERE stage = 'lost')::int,
			COUNT(*) FILTER (WHERE next_follow_up_at IS NOT NULL AND next_follow_up_at <= $2)::int
		FROM crm_customers
		WHERE user_id = $1
	`)).
		WithArgs(int64(42), now).
		WillReturnRows(pgxmock.NewRows([]string{"total", "new", "contacted", "qualified", "proposal", "won", "lost", "due_today"}).
			AddRow(3, 1, 1, 0, 0, 1, 0, 2))

	repository := NewPostgresRepository(db)
	stats, err := repository.PipelineStats(context.Background(), 42, now)
	if err != nil {
		t.Fatalf("PipelineStats() error = %v", err)
	}
	if stats.Total != 3 || stats.New != 1 || stats.Won != 1 || stats.DueToday != 2 {
		t.Fatalf("stats = %+v", stats)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
