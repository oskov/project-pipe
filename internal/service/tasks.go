package service

import (
"context"
"fmt"
"log/slog"
"time"

"github.com/google/uuid"
"github.com/oskov/project-pipe/internal/store"
)

// ManagerAgent is satisfied by agent.Agent — defined here to avoid import cycles.
type ManagerAgent interface {
Run(ctx context.Context, taskID, projectID, prompt string) (string, error)
}

// TaskService defines business operations for tasks.
type TaskService interface {
// Create persists a new task with status "created" and returns it.
Create(ctx context.Context, projectID, prompt string) (*store.Task, error)
// GetByID returns the task by ID.
GetByID(ctx context.Context, id string) (*store.Task, error)
// Execute runs the manager agent for an already-claimed task using the
// provided logger and updates the task status to completed or failed.
Execute(ctx context.Context, task *store.Task, logger *slog.Logger) error
}

type taskService struct {
tasks          store.TaskRepository
projects       store.ProjectRepository
managerFactory func(projectID string, logger *slog.Logger) ManagerAgent
}

func NewTaskService(
tasks store.TaskRepository,
projects store.ProjectRepository,
managerFactory func(projectID string, logger *slog.Logger) ManagerAgent,
) TaskService {
return &taskService{
tasks:          tasks,
projects:       projects,
managerFactory: managerFactory,
}
}

func (s *taskService) Create(ctx context.Context, projectID, prompt string) (*store.Task, error) {
if projectID == "" {
return nil, fmt.Errorf("%w: project_id is required", ErrInvalid)
}
if prompt == "" {
return nil, fmt.Errorf("%w: prompt is required", ErrInvalid)
}

if _, err := s.projects.GetByID(ctx, projectID); err != nil {
return nil, fmt.Errorf("%w: project not found", ErrNotFound)
}

now := time.Now().UTC()
task := &store.Task{
ID:        uuid.New().String(),
ProjectID: projectID,
Prompt:    prompt,
Status:    store.TaskStatusCreated,
CreatedAt: now,
UpdatedAt: now,
}

if err := s.tasks.Create(ctx, task); err != nil {
return nil, fmt.Errorf("%w: failed to create task", ErrInternal)
}

return task, nil
}

func (s *taskService) GetByID(ctx context.Context, id string) (*store.Task, error) {
task, err := s.tasks.GetByID(ctx, id)
if err != nil {
return nil, fmt.Errorf("%w: task not found", ErrNotFound)
}
return task, nil
}

func (s *taskService) Execute(ctx context.Context, task *store.Task, logger *slog.Logger) error {
_ = s.tasks.UpdateStatus(ctx, task.ID, store.TaskStatusProcessing)

manager := s.managerFactory(task.ProjectID, logger)

_, err := manager.Run(ctx, task.ID, task.ProjectID, task.Prompt)
if err != nil {
_ = s.tasks.UpdateStatus(ctx, task.ID, store.TaskStatusFailed)
return fmt.Errorf("agent error for task %s: %w", task.ID, err)
}

_ = s.tasks.UpdateStatus(ctx, task.ID, store.TaskStatusCompleted)
return nil
}
