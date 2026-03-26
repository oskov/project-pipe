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
// credential helper, used for pull after clone when the token is needed again.
func credentialHelper(workDir, token string) string {
	if token == "" {
		return ""
	}
	// Use a store-based credential: encode the helper as an inline script.
	// This is safe for non-interactive use.
	return fmt.Sprintf("credential.helper=!f(){ echo username=x-token; echo password=%s; };f", token)
}

// run executes a git command, capturing stderr for error messages.
func run(ctx context.Context, dir string, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	if dir != "" {
		cmd.Dir = dir
	}
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return fmt.Errorf("%s: %s", strings.Join(args, " "), msg)
	}
	return nil
}
