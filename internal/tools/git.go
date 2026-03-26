package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/oskov/project-pipe/internal/service"
)

// GitStatus returns the short git status of the working tree.
type GitStatus struct{ svc service.GitService }

func NewGitStatus(svc service.GitService) *GitStatus { return &GitStatus{svc: svc} }
func (t *GitStatus) Name() string                    { return "git_status" }
func (t *GitStatus) Description() string {
	return "Return the short git status of the repository. Shows modified, added, and untracked files."
}
func (t *GitStatus) Parameters() json.RawMessage {
	return json.RawMessage(`{"type":"object","properties":{}}`)
}
func (t *GitStatus) Execute(ctx context.Context, _ string) (string, error) {
	return t.svc.Status(ctx)
}

// GitDiff shows unstaged (or staged) changes.
type GitDiff struct{ svc service.GitService }

func NewGitDiff(svc service.GitService) *GitDiff { return &GitDiff{svc: svc} }
func (t *GitDiff) Name() string                  { return "git_diff" }
func (t *GitDiff) Description() string {
	return "Show file diffs. Set staged=true to show only staged (indexed) changes."
}
func (t *GitDiff) Parameters() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"staged": {"type": "boolean", "description": "If true, show only staged changes (git diff --staged). Default: false."}
		}
	}`)
}
func (t *GitDiff) Execute(ctx context.Context, argsJSON string) (string, error) {
	var args struct {
		Staged bool `json:"staged"`
	}
	_ = json.Unmarshal([]byte(argsJSON), &args)
	out, err := t.svc.Diff(ctx, args.Staged)
	if err != nil {
		return "", err
	}
	if out == "" {
		return "no changes", nil
	}
	return out, nil
}

// GitCheckoutBranch creates and switches to a new branch.
type GitCheckoutBranch struct{ svc service.GitService }

func NewGitCheckoutBranch(svc service.GitService) *GitCheckoutBranch {
	return &GitCheckoutBranch{svc: svc}
}
func (t *GitCheckoutBranch) Name() string { return "git_checkout_branch" }
func (t *GitCheckoutBranch) Description() string {
	return "Create and switch to a new git branch. If the branch already exists, just switch to it."
}
func (t *GitCheckoutBranch) Parameters() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"name": {"type": "string", "description": "Branch name, e.g. 'feat/add-user-auth'"}
		},
		"required": ["name"]
	}`)
}
func (t *GitCheckoutBranch) Execute(ctx context.Context, argsJSON string) (string, error) {
	var args struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", fmt.Errorf("parse args: %w", err)
	}
	if err := t.svc.CheckoutBranch(ctx, args.Name); err != nil {
		return "", err
	}
	return fmt.Sprintf("switched to branch %q", args.Name), nil
}

// GitAdd stages files for commit.
type GitAdd struct{ svc service.GitService }

func NewGitAdd(svc service.GitService) *GitAdd { return &GitAdd{svc: svc} }
func (t *GitAdd) Name() string                 { return "git_add" }
func (t *GitAdd) Description() string {
	return "Stage files for the next commit. Pass specific paths or omit to stage all changes ('git add .')."
}
func (t *GitAdd) Parameters() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"paths": {
				"type": "array",
				"items": {"type": "string"},
				"description": "File paths to stage. Omit or pass empty array to stage everything."
			}
		}
	}`)
}
func (t *GitAdd) Execute(ctx context.Context, argsJSON string) (string, error) {
	var args struct {
		Paths []string `json:"paths"`
	}
	_ = json.Unmarshal([]byte(argsJSON), &args)
	if err := t.svc.Add(ctx, args.Paths); err != nil {
		return "", err
	}
	if len(args.Paths) == 0 {
		return "staged all changes", nil
	}
	return fmt.Sprintf("staged: %s", strings.Join(args.Paths, ", ")), nil
}

// GitCommit creates a commit with the given message.
type GitCommit struct{ svc service.GitService }

func NewGitCommit(svc service.GitService) *GitCommit { return &GitCommit{svc: svc} }
func (t *GitCommit) Name() string                    { return "git_commit" }
func (t *GitCommit) Description() string {
	return "Create a git commit with the staged changes. Provide a clear, descriptive commit message."
}
func (t *GitCommit) Parameters() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"message": {"type": "string", "description": "Commit message"}
		},
		"required": ["message"]
	}`)
}
func (t *GitCommit) Execute(ctx context.Context, argsJSON string) (string, error) {
	var args struct {
		Message string `json:"message"`
	}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", fmt.Errorf("parse args: %w", err)
	}
	if err := t.svc.Commit(ctx, args.Message); err != nil {
		return "", err
	}
	return fmt.Sprintf("committed: %q", args.Message), nil
}

// GitPush pushes the current (or specified) branch to origin.
type GitPush struct{ svc service.GitService }

func NewGitPush(svc service.GitService) *GitPush { return &GitPush{svc: svc} }
func (t *GitPush) Name() string                  { return "git_push" }
func (t *GitPush) Description() string {
	return "Push the current branch to the origin remote. Optionally specify a branch name; defaults to the current branch."
}
func (t *GitPush) Parameters() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"branch": {"type": "string", "description": "Branch to push. Defaults to current branch if omitted."}
		}
	}`)
}
func (t *GitPush) Execute(ctx context.Context, argsJSON string) (string, error) {
	var args struct {
		Branch string `json:"branch"`
	}
	_ = json.Unmarshal([]byte(argsJSON), &args)
	if err := t.svc.Push(ctx, args.Branch); err != nil {
		return "", err
	}
	branch := args.Branch
	if branch == "" {
		branch = "current branch"
	}
	return fmt.Sprintf("pushed %s to origin", branch), nil
}
