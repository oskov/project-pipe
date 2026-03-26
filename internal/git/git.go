// Package git provides thin wrappers around the git CLI for repository
// operations. All commands are executed via exec.Command so no CGO is needed.
package git

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"
)

const cloneTimeout = 5 * time.Minute

// Clone clones repoURL into destDir. If token is non-empty it is embedded in
// the HTTPS URL so private repositories can be accessed without SSH keys.
// destDir must not exist yet (git clone creates it).
func Clone(ctx context.Context, repoURL, destDir, token string) error {
	cloneURL, err := injectToken(repoURL, token)
	if err != nil {
		return fmt.Errorf("prepare clone URL: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, cloneTimeout)
	defer cancel()

	if err := run(ctx, "", "git", "clone", "--depth=1", cloneURL, destDir); err != nil {
		return fmt.Errorf("git clone: %w", err)
	}
	return nil
}

// Pull runs "git pull" in workDir, fast-forwarding the current branch.
func Pull(ctx context.Context, workDir, token string) error {
	// Set credential helper inline so we don't pollute the global git config.
	credHelper := credentialHelper(workDir, token)
	args := []string{"pull", "--ff-only"}
	if credHelper != "" {
		args = append([]string{"-c", credHelper}, args...)
	}
	if err := run(ctx, workDir, "git", args...); err != nil {
		return fmt.Errorf("git pull: %w", err)
	}
	return nil
}

// Status returns the output of "git status --short" for workDir.
func Status(ctx context.Context, workDir string) (string, error) {
	out, err := output(ctx, workDir, "git", "status", "--short")
	if err != nil {
		return "", fmt.Errorf("git status: %w", err)
	}
	if out == "" {
		return "nothing to commit, working tree clean", nil
	}
	return out, nil
}

// Diff returns the diff for workDir. When staged is true it returns only the
// staged changes ("git diff --staged").
func Diff(ctx context.Context, workDir string, staged bool) (string, error) {
	args := []string{"diff"}
	if staged {
		args = append(args, "--staged")
	}
	out, err := output(ctx, workDir, "git", args...)
	if err != nil {
		return "", fmt.Errorf("git diff: %w", err)
	}
	return out, nil
}

// CurrentBranch returns the name of the currently checked-out branch.
func CurrentBranch(ctx context.Context, workDir string) (string, error) {
	out, err := output(ctx, workDir, "git", "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", fmt.Errorf("git current branch: %w", err)
	}
	return strings.TrimSpace(out), nil
}

// CheckoutBranch creates a new branch named name and switches to it.
// If the branch already exists it just switches to it.
func CheckoutBranch(ctx context.Context, workDir, name string) error {
	// Try to create; fall back to checkout-only if it already exists.
	err := run(ctx, workDir, "git", "checkout", "-b", name)
	if err != nil {
		if err2 := run(ctx, workDir, "git", "checkout", name); err2 != nil {
			return fmt.Errorf("git checkout branch: %w", err)
		}
	}
	return nil
}

// Add stages the given paths. Pass "." to stage everything.
func Add(ctx context.Context, workDir string, paths []string) error {
	args := append([]string{"add"}, paths...)
	if err := run(ctx, workDir, "git", args...); err != nil {
		return fmt.Errorf("git add: %w", err)
	}
	return nil
}

// Commit creates a commit with message in workDir.
func Commit(ctx context.Context, workDir, message string) error {
	if err := run(ctx, workDir, "git", "commit", "-m", message); err != nil {
		return fmt.Errorf("git commit: %w", err)
	}
	return nil
}

// Push pushes branch to remote (typically "origin").
func Push(ctx context.Context, workDir, remote, branch, token string) error {
	// Configure credential helper for this push only.
	args := []string{"push", remote, branch, "--set-upstream"}
	if token != "" {
		cred := credentialHelper(workDir, token)
		args = append([]string{"-c", cred}, args...)
	}
	if err := run(ctx, workDir, "git", args...); err != nil {
		return fmt.Errorf("git push: %w", err)
	}
	return nil
}

// IsRepo returns true when dir contains a git repository (has a .git entry).
func IsRepo(dir string) bool {
	_, err := os.Stat(fmt.Sprintf("%s/.git", dir))
	return err == nil
}

// ── helpers ────────────────────────────────────────────────────────────────

// injectToken rewrites an HTTPS GitHub URL to embed the token as a password,
// enabling authentication without interactive prompts.
//   https://github.com/org/repo  →  https://x-token:<token>@github.com/org/repo
// SSH URLs and empty tokens are returned unchanged.
func injectToken(rawURL, token string) (string, error) {
	if token == "" || strings.HasPrefix(rawURL, "git@") {
		return rawURL, nil
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	u.User = url.UserPassword("x-token", token)
	return u.String(), nil
}

// credentialHelper returns a git -c argument that supplies the token via a
// credential helper, used for pull/push when the token is needed again.
func credentialHelper(workDir, token string) string {
	if token == "" {
		return ""
	}
	_ = workDir
	return fmt.Sprintf("credential.helper=!f(){ echo username=x-token; echo password=%s; };f", token)
}

// output executes a git command and returns stdout as a string.
func output(ctx context.Context, dir string, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	if dir != "" {
		cmd.Dir = dir
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("%s: %s", strings.Join(args, " "), msg)
	}
	return stdout.String(), nil
}

// run executes a git command, capturing stderr for error messages.
func run(ctx context.Context, dir string, name string, args ...string) error {
	_, err := output(ctx, dir, name, args...)
	return err
}
