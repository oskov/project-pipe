package sqlite

import (
"context"

"github.com/jmoiron/sqlx"
"github.com/oskov/project-pipe/internal/store"
)

type projectRepo struct{ db *sqlx.DB }

func (r *projectRepo) Create(ctx context.Context, p *store.Project) error {
_, err := r.db.ExecContext(ctx, `
INSERT INTO projects (id, name, github_repo, created_at, updated_at)
VALUES (?, ?, ?, ?, ?)`,
p.ID, p.Name, p.GithubRepo, p.CreatedAt, p.UpdatedAt,
)
return err
}

func (r *projectRepo) GetByID(ctx context.Context, id string) (*store.Project, error) {
var p store.Project
if err := r.db.GetContext(ctx, &p, `SELECT * FROM projects WHERE id = ?`, id); err != nil {
return nil, err
}
return &p, nil
}

func (r *projectRepo) List(ctx context.Context) ([]*store.Project, error) {
var projects []*store.Project
if err := r.db.SelectContext(ctx, &projects, `SELECT * FROM projects ORDER BY created_at DESC`); err != nil {
return nil, err
}
return projects, nil
}
