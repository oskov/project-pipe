package service

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/oskov/project-pipe/internal/git"
	"github.com/oskov/project-pipe/internal/store"
)

// ProjectService defines business operations for projects.
type ProjectService interface {
	Create(ctx context.Context, name, githubRepo string) (*store.Project, error)
	List(ctx context.Context) ([]*store.Project, error)
	GetByID(ctx context.Context, id string) (*store.Project, error)
}

type projectService struct {
	repo        store.ProjectRepository
	projectsDir string // root directory where repositories are cloned
	githubToken string
}

func NewProjectService(repo store.ProjectRepository, projectsDir, githubToken string) ProjectService {
	return &projectService{repo: repo, projectsDir: projectsDir, githubToken: githubToken}
}

func (s *projectService) Create(ctx context.Context, name, githubRepo string) (*store.Project, error) {
	if name == "" {
		return nil, fmt.Errorf("%w: name is required", ErrInvalid)
	}
	if githubRepo == "" {
		return nil, fmt.Errorf("%w: github_repo is required", ErrInvalid)
	}

	id := uuid.New().String()
	localPath, err := s.cloneRepo(ctx, id, githubRepo)
	if err != nil {
		return nil, fmt.Errorf("%w: clone repository: %s", ErrInternal, err)
	}

	now := time.Now().UTC()
	p := &store.Project{
		ID:         id,
		Name:       name,
		GithubRepo: githubRepo,
		LocalPath:  localPath,
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

// cloneRepo clones githubRepo into {projectsDir}/{projectID} and returns the
// absolute local path. If the directory already exists and is a valid repo,
// it pulls instead of cloning (idempotent on restart).
func (s *projectService) cloneRepo(ctx context.Context, projectID, repoURL string) (string, error) {
	dest, err := filepath.Abs(filepath.Join(s.projectsDir, projectID))
	if err != nil {
		return "", err
	}

	if git.IsRepo(dest) {
		if err := git.Pull(ctx, dest, s.githubToken); err != nil {
			return "", err
		}
		return dest, nil
	}

	if err := git.Clone(ctx, repoURL, dest, s.githubToken); err != nil {
		return "", err
	}
	return dest, nil
}
