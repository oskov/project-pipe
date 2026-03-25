package sqlite

import (
"context"
"time"

"github.com/jmoiron/sqlx"
"github.com/oskov/project-pipe/internal/store"
)

type agentRunRepo struct{ db *sqlx.DB }

func (r *agentRunRepo) Create(ctx context.Context, run *store.AgentRun) error {
_, err := r.db.ExecContext(ctx, `
INSERT INTO agent_runs (id, task_id, agent_type, status, input, output, error, started_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
run.ID, run.TaskID, run.AgentType, run.Status,
run.Input, run.Output, run.Error, run.StartedAt,
)
return err
}

func (r *agentRunRepo) Complete(ctx context.Context, id string, output string) error {
_, err := r.db.ExecContext(ctx, `
UPDATE agent_runs SET status = ?, output = ?, finished_at = ? WHERE id = ?`,
store.AgentRunStatusCompleted, output, time.Now().UTC(), id,
)
return err
}

func (r *agentRunRepo) Fail(ctx context.Context, id string, errMsg string) error {
_, err := r.db.ExecContext(ctx, `
UPDATE agent_runs SET status = ?, error = ?, finished_at = ? WHERE id = ?`,
store.AgentRunStatusFailed, errMsg, time.Now().UTC(), id,
)
return err
}
