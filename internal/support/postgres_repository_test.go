package support

import (
	"context"
	"regexp"
	"testing"
	"time"

	pgxmock "github.com/pashagolub/pgxmock/v4"
)

func TestPostgresRepositoryCreatesTicket(t *testing.T) {
	db, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 7, 2, 10, 0, 0, 0, time.UTC)
	db.ExpectQuery(regexp.QuoteMeta(`
		INSERT INTO support_tickets (user_id, topic, title, body)
		VALUES ($1, $2, $3, $4)
		RETURNING id, user_id, topic, title, body, status, created_at, updated_at
	`)).
		WithArgs(int64(42), "套餐与额度", "额度没有更新", "正文").
		WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "topic", "title", "body", "status", "created_at", "updated_at"}).
			AddRow(int64(22), int64(42), "套餐与额度", "额度没有更新", "正文", StatusOpen, now, now))

	repository := NewPostgresRepository(db)
	ticket, err := repository.CreateTicket(context.Background(), 42, TicketInput{Topic: "套餐与额度", Title: "额度没有更新", Body: "正文"})
	if err != nil {
		t.Fatalf("CreateTicket() error = %v", err)
	}
	if ticket.ID != 22 || ticket.Status != StatusOpen {
		t.Fatalf("ticket = %+v", ticket)
	}
	if err := db.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
