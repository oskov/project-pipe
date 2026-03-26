package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/oskov/project-pipe/internal/service"
)

// CreatePR opens a GitHub pull request for the current feature branch.
type CreatePR struct {
	taskID    string
	projectID string
	svc       service.PullRequestService
}

func NewCreatePR(taskID, projectID string, svc service.PullRequestService) *CreatePR {
	return &CreatePR{taskID: taskID, projectID: projectID, svc: svc}
}

func (t *CreatePR) Name() string { return "create_pr" }
func (t *CreatePR) Description() string {
	return "Open a GitHub Pull Request for the current feature branch. " +
		"Provide a clear title, descriptive body (markdown supported), head branch (your feature branch), " +
		"and base branch (defaults to 'main'). Returns the PR URL."
}
func (t *CreatePR) Parameters() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"title": {"type": "string", "description": "PR title"},
			"body":  {"type": "string", "description": "PR description in Markdown"},
			"head_branch": {"type": "string", "description": "Feature branch with your changes"},
			"base_branch": {"type": "string", "description": "Target branch to merge into. Default: 'main'"}
		},
		"required": ["title", "head_branch"]
	}`)
}

func (t *CreatePR) Execute(ctx context.Context, argsJSON string) (string, error) {
	var args struct {
		Title      string `json:"title"`
		Body       string `json:"body"`
		HeadBranch string `json:"head_branch"`
		BaseBranch string `json:"base_branch"`
	}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", fmt.Errorf("parse args: %w", err)
	}

	pr, err := t.svc.CreatePR(ctx, t.taskID, t.projectID, args.Title, args.Body, args.HeadBranch, args.BaseBranch)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("PR #%d created: %s", pr.GithubNumber, pr.URL), nil
}
