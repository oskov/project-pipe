package service

import (
	"context"
	"fmt"

	"github.com/oskov/project-pipe/internal/git"
)

// GitService provides git workflow operations scoped to a repository working
// directory. Used by developer agents to create branches, commit, and push.
type GitService interface {
	Status(ctx context.Context) (string, error)
	Diff(ctx context.Context, staged bool) (string, error)
	CurrentBranch(ctx context.Context) (string, error)
	CheckoutBranch(ctx context.Context, name string) error
	Add(ctx context.Context, paths []string) error
	Commit(ctx context.Context, message string) error
	Push(ctx context.Context, branch string) error
}

type gitService struct {
	workDir string
	token   string
}

func NewGitService(workDir, token string) GitService {
	return &gitService{workDir: workDir, token: token}
}

func (s *gitService) Status(ctx context.Context) (string, error) {
	out, err := git.Status(ctx, s.workDir)
	if err != nil {
		return "", fmt.Errorf("git status: %w", err)
	}
	return out, nil
}

func (s *gitService) Diff(ctx context.Context, staged bool) (string, error) {
	out, err := git.Diff(ctx, s.workDir, staged)
	if err != nil {
		return "", fmt.Errorf("git diff: %w", err)
	}
	return out, nil
}

func (s *gitService) CurrentBranch(ctx context.Context) (string, error) {
	branch, err := git.CurrentBranch(ctx, s.workDir)
	if err != nil {
		return "", fmt.Errorf("current branch: %w", err)
	}
	return branch, nil
}

func (s *gitService) CheckoutBranch(ctx context.Context, name string) error {
	if name == "" {
		return fmt.Errorf("%w: branch name is required", ErrInvalid)
	}
	if err := git.CheckoutBranch(ctx, s.workDir, name); err != nil {
		return fmt.Errorf("checkout branch: %w", err)
	}
	return nil
}

func (s *gitService) Add(ctx context.Context, paths []string) error {
	if len(paths) == 0 {
		paths = []string{"."}
	}
	if err := git.Add(ctx, s.workDir, paths); err != nil {
		return fmt.Errorf("git add: %w", err)
	}
	return nil
}

func (s *gitService) Commit(ctx context.Context, message string) error {
	if message == "" {
		return fmt.Errorf("%w: commit message is required", ErrInvalid)
	}
	if err := git.Commit(ctx, s.workDir, message); err != nil {
		return fmt.Errorf("git commit: %w", err)
	}
	return nil
}

func (s *gitService) Push(ctx context.Context, branch string) error {
	if branch == "" {
		var err error
		branch, err = git.CurrentBranch(ctx, s.workDir)
		if err != nil {
			return fmt.Errorf("resolve branch: %w", err)
		}
	}
	if err := git.Push(ctx, s.workDir, "origin", branch, s.token); err != nil {
		return fmt.Errorf("git push: %w", err)
	}
	return nil
}
