package store

import (
	"context"
	"time"
)

type PRStatus string

const (
	PRStatusOpen   PRStatus = "open"
	PRStatusMerged PRStatus = "merged"
	PRStatusClosed PRStatus = "closed"
)

// PullRequest represents a GitHub PR linked to a task.
type PullRequest struct {
	ID           string    `db:"id"`
	TaskID       string    `db:"task_id"`
	ProjectID    string    `db:"project_id"`
	GithubNumber int       `db:"github_number"`
	Title        string    `db:"title"`
	URL          string    `db:"url"`
	HeadBranch   string    `db:"head_branch"`
	BaseBranch   string    `db:"base_branch"`
	Status       PRStatus  `db:"status"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

// PullRequestRepository defines persistence operations for pull requests.
type PullRequestRepository interface {
	Create(ctx context.Context, pr *PullRequest) error
	GetByID(ctx context.Context, id string) (*PullRequest, error)
	GetByTaskID(ctx context.Context, taskID string) ([]*PullRequest, error)
	UpdateStatus(ctx context.Context, id string, status PRStatus) error
}
