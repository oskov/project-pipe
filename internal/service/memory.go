package service

import (
	"context"
	"fmt"

	"github.com/oskov/project-pipe/internal/store"
)

// MemoryService defines business operations for agent long-term memory.
type MemoryService interface {
	Set(ctx context.Context, projectID string, agentType store.AgentType, key, value string) error
	Get(ctx context.Context, projectID string, agentType store.AgentType, key string) (value string, found bool, err error)
	List(ctx context.Context, projectID string, agentType store.AgentType) ([]*store.AgentMemory, error)
}

type memoryService struct {
	repo store.AgentMemoryRepository
}

func NewMemoryService(repo store.AgentMemoryRepository) MemoryService {
	return &memoryService{repo: repo}
}

func (s *memoryService) Set(ctx context.Context, projectID string, agentType store.AgentType, key, value string) error {
	if projectID == "" || key == "" {
		return fmt.Errorf("%w: project_id and key are required", ErrInvalid)
	}
	if err := s.repo.Set(ctx, projectID, agentType, key, value); err != nil {
		return fmt.Errorf("%w: %s", ErrInternal, err)
	}
	return nil
}

func (s *memoryService) Get(ctx context.Context, projectID string, agentType store.AgentType, key string) (string, bool, error) {
	if projectID == "" || key == "" {
		return "", false, fmt.Errorf("%w: project_id and key are required", ErrInvalid)
	}
	value, found, err := s.repo.Get(ctx, projectID, agentType, key)
	if err != nil {
		return "", false, fmt.Errorf("%w: %s", ErrInternal, err)
	}
	return value, found, nil
}

func (s *memoryService) List(ctx context.Context, projectID string, agentType store.AgentType) ([]*store.AgentMemory, error) {
	if projectID == "" {
		return nil, fmt.Errorf("%w: project_id is required", ErrInvalid)
	}
	entries, err := s.repo.List(ctx, projectID, agentType)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInternal, err)
	}
	return entries, nil
}
