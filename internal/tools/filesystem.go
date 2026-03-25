package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/oskov/project-pipe/internal/service"
)

// ReadFile reads the full contents of a file within the workspace.
type ReadFile struct{ svc service.FilesystemService }

func NewReadFile(svc service.FilesystemService) *ReadFile { return &ReadFile{svc: svc} }

func (t *ReadFile) Name() string { return "read_file" }
func (t *ReadFile) Description() string {
	return "Read the full contents of a file (relative to workspace root). Files larger than 32 KB cannot be read whole — use read_file_range to read them in sections."
}
func (t *ReadFile) Parameters() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"path": {"type": "string", "description": "Relative file path within the workspace"}
		},
		"required": ["path"]
	}`)
}

func (t *ReadFile) Execute(_ context.Context, argsJSON string) (string, error) {
	var args struct {
		Path string `json:"path"`
	}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", fmt.Errorf("parse args: %w", err)
	}
	content, err := t.svc.Read(args.Path)
	if err != nil {
		// FileTooBigError has a user-friendly message — return it as result, not error.
		var big *service.FileTooBigError
		if asErr(err, &big) {
			return big.Error(), nil
		}
		return "", err
	}
	return content, nil
}

// ReadFileRange reads a specific range of lines from a file (1-indexed, inclusive).
type ReadFileRange struct{ svc service.FilesystemService }

func NewReadFileRange(svc service.FilesystemService) *ReadFileRange { return &ReadFileRange{svc: svc} }

func (t *ReadFileRange) Name() string { return "read_file_range" }
func (t *ReadFileRange) Description() string {
	return "Read a range of lines from a file (relative to workspace root). Use this for large files or when you only need a specific section. Lines are 1-indexed."
}
func (t *ReadFileRange) Parameters() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"path":       {"type": "string",  "description": "Relative file path within the workspace"},
			"start_line": {"type": "integer", "description": "First line to read (1-indexed, inclusive)"},
			"end_line":   {"type": "integer", "description": "Last line to read (1-indexed, inclusive). Reads to end of file if omitted."}
		},
		"required": ["path", "start_line"]
	}`)
}

func (t *ReadFileRange) Execute(_ context.Context, argsJSON string) (string, error) {
	var args struct {
		Path      string `json:"path"`
		StartLine int    `json:"start_line"`
		EndLine   int    `json:"end_line"`
	}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", fmt.Errorf("parse args: %w", err)
	}
	return t.svc.ReadRange(args.Path, args.StartLine, args.EndLine)
}

// WriteFile writes (or overwrites) a file within the workspace.
type WriteFile struct{ svc service.FilesystemService }

func NewWriteFile(svc service.FilesystemService) *WriteFile { return &WriteFile{svc: svc} }

func (t *WriteFile) Name() string        { return "write_file" }
func (t *WriteFile) Description() string { return "Write content to a file at the given path (relative to workspace root). Creates intermediate directories as needed." }
func (t *WriteFile) Parameters() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"path":    {"type": "string", "description": "Relative file path within the workspace"},
			"content": {"type": "string", "description": "Full file content to write"}
		},
		"required": ["path", "content"]
	}`)
}

func (t *WriteFile) Execute(_ context.Context, argsJSON string) (string, error) {
	var args struct {
		Path    string `json:"path"`
		Content string `json:"content"`
	}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", fmt.Errorf("parse args: %w", err)
	}
	if err := t.svc.Write(args.Path, args.Content); err != nil {
		return "", err
	}
	return fmt.Sprintf("wrote %d bytes to %s", len(args.Content), args.Path), nil
}

// ListFiles lists files in a directory matching an optional glob pattern.
type ListFiles struct{ svc service.FilesystemService }

func NewListFiles(svc service.FilesystemService) *ListFiles { return &ListFiles{svc: svc} }

func (t *ListFiles) Name() string        { return "list_files" }
func (t *ListFiles) Description() string { return "List files and directories at a path (relative to workspace root). Supports optional glob pattern." }
func (t *ListFiles) Parameters() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"path":    {"type": "string", "description": "Directory path relative to workspace (default: '.')"},
			"pattern": {"type": "string", "description": "Optional glob pattern, e.g. '**/*.go'"}
		},
		"required": []
	}`)
}

func (t *ListFiles) Execute(_ context.Context, argsJSON string) (string, error) {
	var args struct {
		Path    string `json:"path"`
		Pattern string `json:"pattern"`
	}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", fmt.Errorf("parse args: %w", err)
	}
	results, err := t.svc.List(args.Path, args.Pattern)
	if err != nil {
		return "", err
	}
	if len(results) == 0 {
		return "(no files found)", nil
	}
	return strings.Join(results, "\n"), nil
}

// SearchCode searches for a text pattern across files in the workspace.
type SearchCode struct{ svc service.FilesystemService }

func NewSearchCode(svc service.FilesystemService) *SearchCode { return &SearchCode{svc: svc} }

func (t *SearchCode) Name() string        { return "search_code" }
func (t *SearchCode) Description() string { return "Search for a text pattern across all files in the workspace. Returns matching lines with file paths and line numbers." }
func (t *SearchCode) Parameters() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"query": {"type": "string", "description": "Text to search for"},
			"path":  {"type": "string", "description": "Subdirectory to restrict the search (default: entire workspace)"},
			"ext":   {"type": "string", "description": "File extension filter, e.g. '.go' (default: all files)"}
		},
		"required": ["query"]
	}`)
}

func (t *SearchCode) Execute(_ context.Context, argsJSON string) (string, error) {
	var args struct {
		Query string `json:"query"`
		Path  string `json:"path"`
		Ext   string `json:"ext"`
	}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return "", fmt.Errorf("parse args: %w", err)
	}
	results, err := t.svc.Search(args.Query, args.Path, args.Ext)
	if err != nil {
		return "", err
	}
	if len(results) == 0 {
		return fmt.Sprintf("no matches found for %q", args.Query), nil
	}
	return strings.Join(results, "\n"), nil
}

// asErr is a helper for errors.As without importing errors in every call site.
func asErr[T error](err error, target *T) bool {
	var e interface{ As(interface{}) bool }
	_ = e
	// Use standard errors.As via type assertion chain.
	type asInterface interface {
		As(any) bool
	}
	for err != nil {
		if t, ok := err.(T); ok {
			*target = t
			return true
		}
		if u, ok := err.(interface{ Unwrap() error }); ok {
			err = u.Unwrap()
		} else {
			break
		}
	}
	return false
}
