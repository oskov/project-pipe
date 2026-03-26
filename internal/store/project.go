package store

import (
"context"
"time"
)

type Project struct {
ID         string    `db:"id"`
Name       string    `db:"name"`
GithubRepo string    `db:"github_repo"`
LocalPath  string    `db:"local_path"`  // absolute path of the cloned repository
CreatedAt  time.Time `db:"created_at"`
UpdatedAt  time.Time `db:"updated_at"`
}

type ProjectRepository interface {
Create(ctx context.Context, project *Project) error
GetByID(ctx context.Context, id string) (*Project, error)
List(ctx context.Context) ([]*Project, error)
}
