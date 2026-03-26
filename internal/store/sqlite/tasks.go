package sqlite

import (
"context"
"database/sql"
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

// ClaimNext atomically selects the oldest "created" task for a project that has
// no "processing" task and marks it "processing". Returns nil, nil when nothing
// is claimable.
func (r *taskRepo) ClaimNext(ctx context.Context) (*store.Task, error) {
	tx, err := r.db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback() //nolint:errcheck

	// Find oldest created task whose project has no processing task.
	var t store.Task
	err = tx.GetContext(ctx, &t, `
		SELECT id, project_id, ticket_id, prompt, status, created_at, updated_at
		FROM tasks t_outer
		WHERE t_outer.status = 'created'
		  AND t_outer.project_id IS NOT NULL
		  AND NOT EXISTS (
		        SELECT 1
		        FROM tasks t_proc
		        WHERE t_proc.status = 'processing'
		          AND t_proc.project_id = t_outer.project_id
		  )
		ORDER BY t_outer.created_at ASC
		LIMIT 1`)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	now := time.Now().UTC()
	if _, err = tx.ExecContext(ctx, `
		UPDATE tasks SET status = 'processing', updated_at = ? WHERE id = ?`,
		now, t.ID); err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	t.Status = store.TaskStatusProcessing
	t.UpdatedAt = now
	return &t, nil
}
