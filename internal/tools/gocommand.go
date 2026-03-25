package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// GoCommand runs Go toolchain commands in the workspace directory.
// Only an explicit allowlist of subcommands is permitted for safety.
type GoCommand struct {
	workDir     string
	allowedSubs map[string]bool
}

// NewGoCommand creates a GoCommand tool restricted to safe Go subcommands.
func NewGoCommand(workDir string) *GoCommand {
	allowed := []string{"build", "test", "vet", "fmt", "help", "mod", "run", "generate"}
	m := make(map[string]bool, len(allowed))
	for _, s := range allowed {
		m[s] = true
	}
	return &GoCommand{workDir: workDir, allowedSubs: m}
}

func (t *GoCommand) Name() string        { return "go_command" }
func (t *GoCommand) Description() string { return "Run a Go toolchain command in the workspace (e.g. go build ./..., go test ./..., go vet ./...). Only safe subcommands are allowed." }
func (t *GoCommand) Parameters() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"subcommand": {
				"type": "string",
				"description": "Go subcommand: build, test, vet, fmt, help, mod, run, generate",
				"enum": ["build", "test", "vet", "fmt", "help", "mod", "run", "generate"]
			},
			"args": {
				"type": "array",
				"items": {"type": "string"},
				"description": "Additional arguments, e.g. [\"./...\"] or [\"-v\", \"./...\"]"
			}
		},
		"required": ["subcommand"]
	}`)
}

func (t *GoCommand) Execute(ctx context.Context, argsJSON string) (string, error) {
	var input struct {
		Subcommand string   `json:"subcommand"`
		Args       []string `json:"args"`
	}
	if err := json.Unmarshal([]byte(argsJSON), &input); err != nil {
		return "", fmt.Errorf("parse args: %w", err)
	}

	if !t.allowedSubs[input.Subcommand] {
		allowed := make([]string, 0, len(t.allowedSubs))
		for k := range t.allowedSubs {
			allowed = append(allowed, k)
		}
		return "", fmt.Errorf("subcommand %q is not allowed; permitted: %s", input.Subcommand, strings.Join(allowed, ", "))
	}

	cmdArgs := append([]string{input.Subcommand}, input.Args...)
	cmd := exec.CommandContext(ctx, "go", cmdArgs...)
	cmd.Dir = t.workDir

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()
	output := out.String()
	if output == "" && err != nil {
		output = err.Error()
	}
	// Return output regardless of exit code so the agent can reason about errors.
	return output, nil
}
