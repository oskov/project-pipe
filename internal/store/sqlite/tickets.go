package sqlite

import (
"context"
"time"

"github.com/jmoiron/sqlx"
"github.com/oskov/project-pipe/internal/store"
)

type ticketRepo struct{ db *sqlx.DB }

func (r *ticketRepo) Create(ctx context.Context, t *store.Ticket) error {
_, err := r.db.ExecContext(ctx, `
INSERT INTO tickets (id, project_id, title, description, status, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?)`,
t.ID, t.ProjectID, t.Title, t.Description, t.Status, t.CreatedAt, t.UpdatedAt,
)
return err
}

func (r *ticketRepo) GetByID(ctx context.Context, id string) (*store.Ticket, error) {
var t store.Ticket
if err := r.db.GetContext(ctx, &t, `SELECT * FROM tickets WHERE id = ?`, id); err != nil {
return nil, err
}
return &t, nil
}

func (r *ticketRepo) ListByProject(ctx context.Context, projectID string, status store.TicketStatus) ([]*store.Ticket, error) {
var tickets []*store.Ticket
if status == "" {
err := r.db.SelectContext(ctx, &tickets,
`SELECT * FROM tickets WHERE project_id = ? ORDER BY created_at DESC`, projectID)
return tickets, err
}
err := r.db.SelectContext(ctx, &tickets,
`SELECT * FROM tickets WHERE project_id = ? AND status = ? ORDER BY created_at DESC`,
projectID, status)
return tickets, err
}

func (r *ticketRepo) UpdateStatus(ctx context.Context, id string, status store.TicketStatus) error {
_, err := r.db.ExecContext(ctx, `
UPDATE tickets SET status = ?, updated_at = ? WHERE id = ?`,
status, time.Now().UTC(), id,
)
return err
}
