package sqlite

import (
"context"
"database/sql"
"time"

"github.com/jmoiron/sqlx"
"github.com/oskov/project-pipe/internal/store"
)

type agentMemoryRepo struct{ db *sqlx.DB }

func (r *agentMemoryRepo) Set(ctx context.Context, projectID string, agentType store.AgentType, key, value string) error {
now := time.Now().UTC()
_, err := r.db.ExecContext(ctx, `
INSERT INTO agent_memory (id, project_id, agent_type, key, value, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(project_id, agent_type, key)
DO UPDATE SET value = excluded.value, updated_at = excluded.updated_at`,
newID(), projectID, agentType, key, value, now, now,
)
return err
}

func (r *agentMemoryRepo) Get(ctx context.Context, projectID string, agentType store.AgentType, key string) (string, bool, error) {
var value string
err := r.db.QueryRowContext(ctx, `
SELECT value FROM agent_memory
WHERE project_id = ? AND agent_type = ? AND key = ?`,
projectID, agentType, key,
).Scan(&value)
if err == sql.ErrNoRows {
return "", false, nil
}
if err != nil {
return "", false, err
}
return value, true, nil
}

func (r *agentMemoryRepo) List(ctx context.Context, projectID string, agentType store.AgentType) ([]*store.AgentMemory, error) {
var entries []*store.AgentMemory
err := r.db.SelectContext(ctx, &entries, `
SELECT * FROM agent_memory
WHERE project_id = ? AND agent_type = ?
ORDER BY key`,
projectID, agentType,
)
return entries, err
}

func (r *agentMemoryRepo) Delete(ctx context.Context, projectID string, agentType store.AgentType, key string) error {
_, err := r.db.ExecContext(ctx, `
DELETE FROM agent_memory WHERE project_id = ? AND agent_type = ? AND key = ?`,
projectID, agentType, key,
)
return err
}
