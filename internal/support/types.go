package support

import "time"

const StatusOpen = "open"

type TicketInput struct {
	Topic string `json:"topic"`
	Title string `json:"title"`
	Body  string `json:"body"`
}

type Ticket struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id,omitempty"`
	Topic     string    `json:"topic,omitempty"`
	Title     string    `json:"title"`
	Body      string    `json:"body,omitempty"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}
