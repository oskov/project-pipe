package store

import (
"context"
"time"
)

type AgentRunStatus string

const (
AgentRunStatusRunning   AgentRunStatus = "running"
AgentRunStatusCompleted AgentRunStatus = "completed"
AgentRunStatusFailed    AgentRunStatus = "failed"
)

type AgentType string

const (
	AgentTypeManager         AgentType = "manager"
	AgentTypeArchitect       AgentType = "architect"
	AgentTypeDevManager      AgentType = "dev_manager"
	AgentTypeGolangDeveloper AgentType = "golang_developer"
)

type AgentRun struct {
ID         string         `db:"id"`
TaskID     string         `db:"task_id"`
AgentType  AgentType      `db:"agent_type"`
Status     AgentRunStatus `db:"status"`
Input      string         `db:"input"`
Output     string         `db:"output"`
Error      string         `db:"error"`
StartedAt  time.Time      `db:"started_at"`
FinishedAt *time.Time     `db:"finished_at"`
}

type AgentRunRepository interface {
Create(ctx context.Context, run *AgentRun) error
Complete(ctx context.Context, id string, output string) error
Fail(ctx context.Context, id string, errMsg string) error
}
