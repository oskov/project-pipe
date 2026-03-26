package sqlite

import (
	"context"
	"time"

	"github.com/oskov/project-pipe/internal/store"
)

type pullRequestRepository struct{ db *sqliteStore }

func (r *pullRequestRepository) Create(ctx context.Context, pr *store.PullRequest) error {
	const q = `INSERT INTO pull_requests
		(id, task_id, project_id, github_number, title, url, head_branch, base_branch, status, created_at, updated_at)
		VALUES (:id, :task_id, :project_id, :github_number, :title, :url, :head_branch, :base_branch, :status, :created_at, :updated_at)`
	_, err := r.db.db.NamedExecContext(ctx, q, pr)
	return err
}

func (r *pullRequestRepository) GetByID(ctx context.Context, id string) (*store.PullRequest, error) {
	var pr store.PullRequest
	err := r.db.db.GetContext(ctx, &pr, `SELECT * FROM pull_requests WHERE id = ?`, id)
	if err != nil {
		return nil, err
	}
	return &pr, nil
}

func (r *pullRequestRepository) GetByTaskID(ctx context.Context, taskID string) ([]*store.PullRequest, error) {
	var prs []*store.PullRequest
	err := r.db.db.SelectContext(ctx, &prs,
		`SELECT * FROM pull_requests WHERE task_id = ? ORDER BY created_at ASC`, taskID)
	return prs, err
}

func (r *pullRequestRepository) UpdateStatus(ctx context.Context, id string, status store.PRStatus) error {
	_, err := r.db.db.ExecContext(ctx,
		`UPDATE pull_requests SET status = ?, updated_at = ? WHERE id = ?`,
		status, time.Now().UTC(), id)
	return err
}
