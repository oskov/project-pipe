package store

import (
"context"
"time"
)

// AgentMemory is a key-value record scoped to a project and agent type.
type AgentMemory struct {
ID        string    `db:"id"`
ProjectID string    `db:"project_id"`
AgentType AgentType `db:"agent_type"`
Key       string    `db:"key"`
Value     string    `db:"value"`
CreatedAt time.Time `db:"created_at"`
UpdatedAt time.Time `db:"updated_at"`
}

type AgentMemoryRepository interface {
// Set creates or updates a memory entry (upsert).
Set(ctx context.Context, projectID string, agentType AgentType, key, value string) error
// Get retrieves a single memory entry. Returns ("", false, nil) if not found.
Get(ctx context.Context, projectID string, agentType AgentType, key string) (string, bool, error)
// List returns all memory entries for a project/agent.
List(ctx context.Context, projectID string, agentType AgentType) ([]*AgentMemory, error)
// Delete removes a memory entry. No-op if not found.
Delete(ctx context.Context, projectID string, agentType AgentType, key string) error
}
