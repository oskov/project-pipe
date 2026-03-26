package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	githubclient "github.com/oskov/project-pipe/internal/github"
	"github.com/oskov/project-pipe/internal/store"
)

// PullRequestService manages GitHub PRs linked to tasks.
type PullRequestService interface {
	// CreatePR opens a PR on GitHub and records it in the DB.
	CreatePR(ctx context.Context, taskID, projectID, title, body, headBranch, baseBranch string) (*store.PullRequest, error)
	// GetByTaskID returns all PRs associated with a task.
	GetByTaskID(ctx context.Context, taskID string) ([]*store.PullRequest, error)
}

type pullRequestService struct {
	repo       store.PullRequestRepository
	projects   store.ProjectRepository
	ghClient   *githubclient.Client
}

func NewPullRequestService(
	repo store.PullRequestRepository,
	projects store.ProjectRepository,
	githubToken string,
) PullRequestService {
	return &pullRequestService{
		repo:     repo,
		projects: projects,
		ghClient: githubclient.New(githubToken),
	}
}

func (s *pullRequestService) CreatePR(
	ctx context.Context,
	taskID, projectID, title, body, headBranch, baseBranch string,
) (*store.PullRequest, error) {
	if title == "" {
		return nil, fmt.Errorf("%w: title is required", ErrInvalid)
	}
	if headBranch == "" {
		return nil, fmt.Errorf("%w: head branch is required", ErrInvalid)
	}

	project, err := s.projects.GetByID(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("%w: project not found", ErrNotFound)
	}

	owner, repo, err := parseGithubRepo(project.GithubRepo)
	if err != nil {
		return nil, fmt.Errorf("parse repo: %w", err)
	}

	if baseBranch == "" {
		baseBranch = "main"
	}

	ghPR, err := s.ghClient.CreatePR(ctx, githubclient.CreatePRInput{
		Owner: owner,
		Repo:  repo,
		Title: title,
		Body:  body,
		Head:  headBranch,
		Base:  baseBranch,
	})
	if err != nil {
		return nil, fmt.Errorf("github create pr: %w", err)
	}

	now := time.Now().UTC()
	pr := &store.PullRequest{
		ID:           uuid.New().String(),
		TaskID:       taskID,
		ProjectID:    projectID,
		GithubNumber: ghPR.Number,
		Title:        title,
		URL:          ghPR.HTMLURL,
		HeadBranch:   headBranch,
		BaseBranch:   baseBranch,
		Status:       store.PRStatusOpen,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.repo.Create(ctx, pr); err != nil {
		return nil, fmt.Errorf("%w: save pr", ErrInternal)
	}

	return pr, nil
}

func (s *pullRequestService) GetByTaskID(ctx context.Context, taskID string) ([]*store.PullRequest, error) {
	return s.repo.GetByTaskID(ctx, taskID)
}

// parseGithubRepo splits "https://github.com/owner/repo" or "owner/repo"
// into (owner, repo). Handles .git suffix.
func parseGithubRepo(repoURL string) (owner, repo string, err error) {
	// Strip scheme and host if present.
	s := repoURL
	for _, prefix := range []string{"https://github.com/", "http://github.com/", "git@github.com:"} {
		s = strings.TrimPrefix(s, prefix)
	}
	s = strings.TrimSuffix(s, ".git")
	parts := strings.SplitN(s, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("cannot parse %q as owner/repo", repoURL)
	}
	return parts[0], parts[1], nil
}
