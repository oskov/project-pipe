package service

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

var allowedSubcommands = map[string]bool{
	"build":    true,
	"test":     true,
	"vet":      true,
	"fmt":      true,
	"help":     true,
	"mod":      true,
	"run":      true,
	"generate": true,
}

// GoToolchainService runs Go toolchain commands in the workspace directory.
type GoToolchainService interface {
	Run(ctx context.Context, subcommand string, args []string) (string, error)
}

type goToolchainService struct {
	workDir string
}

func NewGoToolchainService(workDir string) GoToolchainService {
	return &goToolchainService{workDir: workDir}
}

func (s *goToolchainService) Run(ctx context.Context, subcommand string, args []string) (string, error) {
	if !allowedSubcommands[subcommand] {
		allowed := make([]string, 0, len(allowedSubcommands))
		for k := range allowedSubcommands {
			allowed = append(allowed, k)
		}
		return "", fmt.Errorf("%w: subcommand %q is not allowed; permitted: %s",
			ErrInvalid, subcommand, strings.Join(allowed, ", "))
	}

	cmdArgs := append([]string{subcommand}, args...)
	cmd := exec.CommandContext(ctx, "go", cmdArgs...)
	cmd.Dir = s.workDir

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	// Intentionally ignore the exit code — return combined output so the
	// caller (agent) can reason about compiler errors, test failures, etc.
	_ = cmd.Run()
	output := out.String()
	if output == "" {
		output = fmt.Sprintf("go %s: no output", subcommand)
	}
	return output, nil
}
