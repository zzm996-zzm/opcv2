package support

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type postgresDB interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type PostgresRepository struct {
	db postgresDB
}

func NewPostgresRepository(db postgresDB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) CreateTicket(ctx context.Context, userID int64, input TicketInput) (Ticket, error) {
	return scanTicket(r.db.QueryRow(ctx, `
		INSERT INTO support_tickets (user_id, topic, title, body)
		VALUES ($1, $2, $3, $4)
		RETURNING id, user_id, topic, title, body, status, created_at, updated_at
	`, userID, input.Topic, input.Title, input.Body))
}

func (r *PostgresRepository) ListTickets(ctx context.Context, userID int64, limit int) ([]Ticket, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, user_id, topic, title, body, status, created_at, updated_at
		FROM support_tickets
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tickets []Ticket
	for rows.Next() {
		ticket, err := scanTicket(rows)
		if err != nil {
			return nil, err
		}
		tickets = append(tickets, ticket)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if tickets == nil {
		return []Ticket{}, nil
	}
	return tickets, nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanTicket(scanner scanner) (Ticket, error) {
	var ticket Ticket
	err := scanner.Scan(
		&ticket.ID,
		&ticket.UserID,
		&ticket.Topic,
		&ticket.Title,
		&ticket.Body,
		&ticket.Status,
		&ticket.CreatedAt,
		&ticket.UpdatedAt,
	)
	return ticket, err
}
