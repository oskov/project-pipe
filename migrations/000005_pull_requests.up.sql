CREATE TABLE IF NOT EXISTS pull_requests (
    id          TEXT    PRIMARY KEY,
    task_id     TEXT    NOT NULL REFERENCES tasks(id),
    project_id  TEXT    NOT NULL REFERENCES projects(id),
    github_number INTEGER NOT NULL,
    title       TEXT    NOT NULL,
    url         TEXT    NOT NULL,
    head_branch TEXT    NOT NULL,
    base_branch TEXT    NOT NULL,
    status      TEXT    NOT NULL DEFAULT 'open',
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_pull_requests_task_id ON pull_requests(task_id);
