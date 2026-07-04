package support

import (
	"context"
	"errors"
	"testing"
)

type fakeRepository struct {
	input  TicketInput
	ticket Ticket
	rows   []Ticket
	userID int64
	limit  int
	err    error
}

func (r *fakeRepository) CreateTicket(_ context.Context, userID int64, input TicketInput) (Ticket, error) {
	r.userID = userID
	r.input = input
	return r.ticket, r.err
}

func (r *fakeRepository) ListTickets(_ context.Context, userID int64, limit int) ([]Ticket, error) {
	r.userID = userID
	r.limit = limit
	return r.rows, r.err
}

func TestServiceCreatesTicketAndTrimsInput(t *testing.T) {
	repository := &fakeRepository{ticket: Ticket{ID: 22, Status: StatusOpen}}
	service := NewService(repository)

	ticket, err := service.CreateTicket(context.Background(), 42, TicketInput{
		Topic: " 套餐与额度 ",
		Title: " 额度没有更新 ",
		Body:  " 我兑换后额度仍未变化。 ",
	})

	if err != nil {
		t.Fatalf("CreateTicket() error = %v", err)
	}
	if ticket.ID != 22 || repository.userID != 42 || repository.input.Title != "额度没有更新" {
		t.Fatalf("ticket/input = %+v/%+v", ticket, repository.input)
	}
}

func TestServiceRejectsInvalidTicket(t *testing.T) {
	service := NewService(&fakeRepository{})

	_, err := service.CreateTicket(context.Background(), 42, TicketInput{Title: " ", Body: "正文"})

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("err = %v, want ErrInvalidInput", err)
	}
}

func TestServiceListsTicketsWithCappedLimit(t *testing.T) {
	repository := &fakeRepository{rows: []Ticket{{ID: 1, Status: StatusOpen}}}
	service := NewService(repository)

	rows, err := service.ListTickets(context.Background(), 42, 500)

	if err != nil {
		t.Fatalf("ListTickets() error = %v", err)
	}
	if len(rows) != 1 || repository.limit != 100 {
		t.Fatalf("rows/limit = %+v/%d", rows, repository.limit)
	}
}
