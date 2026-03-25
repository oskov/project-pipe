package sqlite

import (
"context"
"time"

"github.com/jmoiron/sqlx"
"github.com/oskov/project-pipe/internal/store"
)

type taskRepo struct{ db *sqlx.DB }

func (r *taskRepo) Create(ctx context.Context, t *store.Task) error {
_, err := r.db.ExecContext(ctx, `
INSERT INTO tasks (id, project_id, prompt, status, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?)`,
t.ID, t.ProjectID, t.Prompt, t.Status, t.CreatedAt, t.UpdatedAt,
)
return err
}

func (r *taskRepo) GetByID(ctx context.Context, id string) (*store.Task, error) {
var t store.Task
if err := r.db.GetContext(ctx, &t, `SELECT * FROM tasks WHERE id = ?`, id); err != nil {
return nil, err
}
return &t, nil
}

func (r *taskRepo) UpdateStatus(ctx context.Context, id string, status store.TaskStatus) error {
_, err := r.db.ExecContext(ctx, `
UPDATE tasks SET status = ?, updated_at = ? WHERE id = ?`,
status, time.Now().UTC(), id,
)
return err
}

func (r *taskRepo) SetTicket(ctx context.Context, id string, ticketID string) error {
_, err := r.db.ExecContext(ctx, `
UPDATE tasks SET ticket_id = ?, updated_at = ? WHERE id = ?`,
ticketID, time.Now().UTC(), id,
)
return err
}
