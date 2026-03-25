package store

import (
"context"
"time"
)

type TaskStatus string

const (
TaskStatusCreated    TaskStatus = "created"
TaskStatusProcessing TaskStatus = "processing"
TaskStatusCompleted  TaskStatus = "completed"
TaskStatusFailed     TaskStatus = "failed"
)

type Task struct {
ID        string     `db:"id"`
ProjectID string     `db:"project_id"` // required
TicketID  *string    `db:"ticket_id"`  // set after manager assigns the prompt
Prompt    string     `db:"prompt"`
Status    TaskStatus `db:"status"`
CreatedAt time.Time  `db:"created_at"`
UpdatedAt time.Time  `db:"updated_at"`
}

type TaskRepository interface {
Create(ctx context.Context, task *Task) error
GetByID(ctx context.Context, id string) (*Task, error)
UpdateStatus(ctx context.Context, id string, status TaskStatus) error
SetTicket(ctx context.Context, id string, ticketID string) error
	// ClaimNext atomically finds the oldest "created" task for any project
	// that has no "processing" task, marks it "processing" and returns it.
	// Returns nil, nil when no claimable task exists.
	ClaimNext(ctx context.Context) (*Task, error)
}
