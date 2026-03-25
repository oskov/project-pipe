package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/oskov/project-pipe/internal/service"
)

// GoCommand runs Go toolchain commands in the workspace directory.
type GoCommand struct {
	svc service.GoToolchainService
}

func NewGoCommand(svc service.GoToolchainService) *GoCommand {
	return &GoCommand{svc: svc}
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
	return t.svc.Run(ctx, input.Subcommand, input.Args)
}
