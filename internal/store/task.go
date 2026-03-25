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
}
