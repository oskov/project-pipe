package tools

import "context"

type contextKey string

const taskIDKey contextKey = "agent_task_id"

// ContextWithTaskID injects the current task ID into the context so sub-agent
// tools can propagate it to child agent runs.
func ContextWithTaskID(ctx context.Context, taskID string) context.Context {
	return context.WithValue(ctx, taskIDKey, taskID)
}

// TaskIDFromContext retrieves the task ID previously set by ContextWithTaskID.
// Returns an empty string if not present.
func TaskIDFromContext(ctx context.Context) string {
	v, _ := ctx.Value(taskIDKey).(string)
	return v
}
