package store

import (
"context"
"time"
)

type TicketStatus string

const (
TicketStatusOpen       TicketStatus = "open"
TicketStatusInProgress TicketStatus = "in_progress"
TicketStatusDone       TicketStatus = "done"
TicketStatusClosed     TicketStatus = "closed"
)

type Ticket struct {
ID          string       `db:"id"`
ProjectID   string       `db:"project_id"`
Title       string       `db:"title"`
Description string       `db:"description"`
Status      TicketStatus `db:"status"`
CreatedAt   time.Time    `db:"created_at"`
UpdatedAt   time.Time    `db:"updated_at"`
}

type TicketRepository interface {
Create(ctx context.Context, ticket *Ticket) error
GetByID(ctx context.Context, id string) (*Ticket, error)
// ListByProject returns tickets for a project, optionally filtered by status.
// Pass empty string to return all statuses.
ListByProject(ctx context.Context, projectID string, status TicketStatus) ([]*Ticket, error)
UpdateStatus(ctx context.Context, id string, status TicketStatus) error
}
