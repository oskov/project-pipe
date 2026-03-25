package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/oskov/project-pipe/internal/service"
)

// GoDefinitions lists all top-level definitions in one or more Go source files.
type GoDefinitions struct{ svc service.GoParseService }

func NewGoDefinitions(svc service.GoParseService) *GoDefinitions { return &GoDefinitions{svc: svc} }

func (t *GoDefinitions) Name() string { return "go_definitions" }
func (t *GoDefinitions) Description() string {
	return "List all top-level definitions (functions, methods, types, variables, constants) in one or more Go source files. Use this before reading a file in detail."
}
func (t *GoDefinitions) Parameters() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"files": {
				"type":  "array",
				"items": {"type": "string"},
				"description": "Relative paths of Go source files to inspect (can be multiple)"
			}
		},
		"required": ["files"]
	}`)
}

func (t *GoDefinitions) Execute(_ context.Context, argsJSON string) (string, error) {
	var args struct {
		Files []string `json:"files"`
	}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", fmt.Errorf("parse args: %w", err)
	}
	return t.svc.ListDefinitions(args.Files)
}

// GoReadDefinition returns the full source of a named top-level definition.
type GoReadDefinition struct{ svc service.GoParseService }

func NewGoReadDefinition(svc service.GoParseService) *GoReadDefinition {
	return &GoReadDefinition{svc: svc}
}

func (t *GoReadDefinition) Name() string { return "go_read_definition" }
func (t *GoReadDefinition) Description() string {
	return "Read the full source code of a named top-level definition from a Go file (function, method, type, var block, const block). Includes the doc comment. For grouped var/const blocks the entire block is returned."
}
func (t *GoReadDefinition) Parameters() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"file": {"type": "string", "description": "Relative path to the Go source file"},
			"name": {"type": "string", "description": "Exact name of the definition (e.g. 'NewProject', 'Project', 'ErrNotFound')"}
		},
		"required": ["file", "name"]
	}`)
}

func (t *GoReadDefinition) Execute(_ context.Context, argsJSON string) (string, error) {
	var args struct {
		File string `json:"file"`
		Name string `json:"name"`
	}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", fmt.Errorf("parse args: %w", err)
	}
	return t.svc.ReadDefinition(args.File, args.Name)
}
