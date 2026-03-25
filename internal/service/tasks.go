package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/oskov/project-pipe/internal/agent"
	"github.com/oskov/project-pipe/internal/store"
)

// CreateTaskResult holds the outcome of a processed task.
type CreateTaskResult struct {
	TaskID   string
	Status   store.TaskStatus
	Response string
}

// TaskService defines business operations for tasks.
type TaskService interface {
	Create(ctx context.Context, projectID, prompt string) (*CreateTaskResult, error)
}

type taskService struct {
	tasks          store.TaskRepository
	projects       store.ProjectRepository
	managerFactory func(projectID string) *agent.Agent
}

func NewTaskService(
	tasks store.TaskRepository,
	projects store.ProjectRepository,
	managerFactory func(projectID string) *agent.Agent,
) TaskService {
	return &taskService{
		tasks:          tasks,
		projects:       projects,
		managerFactory: managerFactory,
	}
}

func (s *taskService) Create(ctx context.Context, projectID, prompt string) (*CreateTaskResult, error) {
	if projectID == "" {
		return nil, fmt.Errorf("%w: project_id is required", ErrInvalid)
	}
	if prompt == "" {
		return nil, fmt.Errorf("%w: prompt is required", ErrInvalid)
	}

	if _, err := s.projects.GetByID(ctx, projectID); err != nil {
		return nil, fmt.Errorf("%w: project not found", ErrNotFound)
	}

	taskID := uuid.New().String()
	now := time.Now().UTC()

	task := &store.Task{
		ID:        taskID,
		ProjectID: projectID,
		Prompt:    prompt,
		Status:    store.TaskStatusCreated,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.tasks.Create(ctx, task); err != nil {
		return nil, fmt.Errorf("%w: failed to create task", ErrInternal)
	}

	_ = s.tasks.UpdateStatus(ctx, taskID, store.TaskStatusProcessing)

	manager := s.managerFactory(projectID)

	response, err := manager.Run(ctx, taskID, projectID, prompt)
	if err != nil {
		_ = s.tasks.UpdateStatus(ctx, taskID, store.TaskStatusFailed)
		return nil, fmt.Errorf("%w: agent error: %s", ErrInternal, err)
	}

	_ = s.tasks.UpdateStatus(ctx, taskID, store.TaskStatusCompleted)

	return &CreateTaskResult{
		TaskID:   taskID,
		Status:   store.TaskStatusCompleted,
		Response: response,
	}, nil
}
