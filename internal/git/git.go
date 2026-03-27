// Package git provides thin wrappers around the git CLI for repository
// operations. All commands are executed via exec.Command so no CGO is needed.
package git

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

const cloneTimeout = 5 * time.Minute

// Clone clones repoURL into destDir. If token is non-empty it is supplied via
// GIT_ASKPASS so the secret is never embedded in the clone URL (and therefore
// never written to .git/config) or exposed on the command line.
// destDir must not exist yet (git clone creates it).
func Clone(ctx context.Context, repoURL, destDir, token string) error {
	ctx, cancel := context.WithTimeout(ctx, cloneTimeout)
	defer cancel()

	env, cleanup, err := credEnv(token)
	if err != nil {
		return fmt.Errorf("prepare credentials: %w", err)
	}
	defer cleanup()

	if err := run(ctx, "", env, "git", "clone", "--depth=1", repoURL, destDir); err != nil {
		return fmt.Errorf("git clone: %w", err)
	}
	return nil
}

// Pull runs "git pull --ff-only" in workDir. If token is non-empty it is
// supplied via GIT_ASKPASS without touching the command line or .git/config.
func Pull(ctx context.Context, workDir, token string) error {
	env, cleanup, err := credEnv(token)
	if err != nil {
		return fmt.Errorf("prepare credentials: %w", err)
	}
	defer cleanup()

	if err := run(ctx, workDir, env, "git", "pull", "--ff-only"); err != nil {
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

// credEnv builds an environment slice for git commands. It always sets
// GIT_TERMINAL_PROMPT=0 so that git fails immediately instead of blocking on
// user input. When token is non-empty it also creates a temporary GIT_ASKPASS
// script that supplies the token as the password, keeping the secret off the
// command line and out of .git/config.
// The caller must invoke the returned cleanup function to remove the temp file.
func credEnv(token string) (env []string, cleanup func(), err error) {
	env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	if token == "" {
		return env, func() {}, nil
	}

	// Write a minimal askpass helper that echoes credentials to git.
	script := "#!/bin/sh\n" +
		"case \"$1\" in\n" +
		"  *Username*) echo x-token ;;\n" +
		"  *Password*) echo " + shellSingleQuote(token) + " ;;\n" +
		"esac\n"

	f, ferr := os.CreateTemp("", "git-askpass-*.sh")
	if ferr != nil {
		return nil, func() {}, fmt.Errorf("create askpass script: %w", ferr)
	}
	if _, werr := f.WriteString(script); werr != nil {
		f.Close()
		os.Remove(f.Name())
		return nil, func() {}, fmt.Errorf("write askpass script: %w", werr)
	}
	f.Close()

	// 0500 = read+execute for owner only; no write to prevent accidental modification.
	if cherr := os.Chmod(f.Name(), 0500); cherr != nil {
		os.Remove(f.Name())
		return nil, func() {}, fmt.Errorf("chmod askpass script: %w", cherr)
	}

	env = append(env, "GIT_ASKPASS="+f.Name())
	return env, func() { os.Remove(f.Name()) }, nil
}

// shellSingleQuote wraps s in single quotes, escaping any embedded single
// quotes, producing a safe shell literal.
func shellSingleQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// run executes a git command with the provided environment, capturing stderr
// for readable error messages.
func run(ctx context.Context, dir string, env []string, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	if dir != "" {
		cmd.Dir = dir
	}
	cmd.Env = env
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
