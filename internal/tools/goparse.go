package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/oskov/project-pipe/internal/service"
)

// GoDefinitions lists all top-level definitions in one or more Go source files.
// The tool enforces maxToolOutputBytes: if multiple files are requested and the
// combined output exceeds the limit, an error is returned asking the agent to
// request fewer files. For a single file the full output is always returned.
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
				"description": "Relative paths of Go source files to inspect (1–N files)"
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
	if len(args.Files) == 0 {
		return "", fmt.Errorf("files list is empty")
	}

	type result struct {
		file string
		body string
		err  error
	}
	results := make([]result, len(args.Files))
	for i, f := range args.Files {
		body, err := t.svc.ListDefinitions(f)
		results[i] = result{file: f, body: body, err: err}
	}

	var sb strings.Builder
	for _, r := range results {
		if r.err != nil {
			fmt.Fprintf(&sb, "=== %s ===\nerror: %s\n\n", r.file, r.err)
		} else {
			fmt.Fprintf(&sb, "=== %s ===\n%s\n", r.file, r.body)
		}
	}
	out := sb.String()

	if len(out) > maxToolOutputBytes && len(args.Files) > 1 {
		return fmt.Sprintf(
			"response too large (%d bytes for %d files). Request fewer files at a time.",
			len(out), len(args.Files),
		), nil
	}
	return out, nil
}

// GoReadDefinition returns the full source of one or more named top-level definitions.
// The tool enforces maxToolOutputBytes: if multiple names are requested and the
// combined output exceeds the limit, an error is returned asking the agent to
// request fewer definitions. For a single name the full source is always returned.
type GoReadDefinition struct{ svc service.GoParseService }

func NewGoReadDefinition(svc service.GoParseService) *GoReadDefinition {
	return &GoReadDefinition{svc: svc}
}

func (t *GoReadDefinition) Name() string { return "go_read_definition" }
func (t *GoReadDefinition) Description() string {
	return "Read the full source code of one or more named top-level definitions from a Go file (functions, methods, types, var blocks, const blocks). Includes doc comments. Pass multiple names to fetch several at once; if the combined output is too large you will be asked to request fewer."
}
func (t *GoReadDefinition) Parameters() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"file": {"type": "string", "description": "Relative path to the Go source file"},
			"names": {
				"type": "array",
				"items": {"type": "string"},
				"description": "Names of the definitions to read (e.g. ['NewProject', 'Project', 'ErrNotFound'])"
			}
		},
		"required": ["file", "names"]
	}`)
}

func (t *GoReadDefinition) Execute(_ context.Context, argsJSON string) (string, error) {
	var args struct {
		File  string   `json:"file"`
		Names []string `json:"names"`
	}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", fmt.Errorf("parse args: %w", err)
	}
	if args.File == "" {
		return "", fmt.Errorf("file is required")
	}
	if len(args.Names) == 0 {
		return "", fmt.Errorf("names list is empty")
	}

	type result struct {
		name string
		src  string
		err  error
	}
	results := make([]result, len(args.Names))
	for i, name := range args.Names {
		src, err := t.svc.ReadDefinition(args.File, name)
		results[i] = result{name: name, src: src, err: err}
	}

	var sb strings.Builder
	for _, r := range results {
		if r.err != nil {
			fmt.Fprintf(&sb, "// error reading %q: %s\n\n", r.name, r.err)
		} else {
			fmt.Fprintf(&sb, "%s\n\n", r.src)
		}
	}
	out := sb.String()

	if len(out) > maxToolOutputBytes && len(args.Names) > 1 {
		return fmt.Sprintf(
			"response too large (%d bytes for %d definitions). Request fewer definitions at a time.",
			len(out), len(args.Names),
		), nil
	}
	return out, nil
}
