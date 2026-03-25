package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/oskov/project-pipe/internal/store"
)

// ProjectService defines business operations for projects.
type ProjectService interface {
	Create(ctx context.Context, name, githubRepo string) (*store.Project, error)
	List(ctx context.Context) ([]*store.Project, error)
	GetByID(ctx context.Context, id string) (*store.Project, error)
}

type projectService struct {
	repo store.ProjectRepository
}

func NewProjectService(repo store.ProjectRepository) ProjectService {
	return &projectService{repo: repo}
}

func (s *projectService) Create(ctx context.Context, name, githubRepo string) (*store.Project, error) {
	if name == "" {
		return nil, fmt.Errorf("%w: name is required", ErrInvalid)
	}

	now := time.Now().UTC()
	p := &store.Project{
		ID:         uuid.New().String(),
		Name:       name,
		GithubRepo: githubRepo,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if err := s.repo.Create(ctx, p); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInternal, err)
	}
	return p, nil
}

func (s *projectService) List(ctx context.Context) ([]*store.Project, error) {
	projects, err := s.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInternal, err)
	}
	return projects, nil
}

func (s *projectService) GetByID(ctx context.Context, id string) (*store.Project, error) {
	p, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrNotFound, err)
	}
	return p, nil
}
