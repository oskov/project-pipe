package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ReadFile reads the contents of a file within the workspace.
type ReadFile struct{ workDir string }

func NewReadFile(workDir string) *ReadFile { return &ReadFile{workDir: workDir} }

func (t *ReadFile) Name() string        { return "read_file" }
func (t *ReadFile) Description() string { return "Read the full contents of a file at the given path (relative to the workspace root)." }
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
	content, err := os.ReadFile(filepath.Join(t.workDir, filepath.Clean(args.Path)))
	if err != nil {
		return "", fmt.Errorf("read file: %w", err)
	}
	return string(content), nil
}

// WriteFile writes (or overwrites) a file within the workspace.
type WriteFile struct{ workDir string }

func NewWriteFile(workDir string) *WriteFile { return &WriteFile{workDir: workDir} }

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
	absPath := filepath.Join(t.workDir, filepath.Clean(args.Path))
	if err := os.MkdirAll(filepath.Dir(absPath), 0o755); err != nil {
		return "", fmt.Errorf("create dirs: %w", err)
	}
	if err := os.WriteFile(absPath, []byte(args.Content), 0o644); err != nil {
		return "", fmt.Errorf("write file: %w", err)
	}
	return fmt.Sprintf("wrote %d bytes to %s", len(args.Content), args.Path), nil
}

// ListFiles lists files in a directory matching an optional glob pattern.
type ListFiles struct{ workDir string }

func NewListFiles(workDir string) *ListFiles { return &ListFiles{workDir: workDir} }

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

	base := t.workDir
	if args.Path != "" {
		base = filepath.Join(t.workDir, filepath.Clean(args.Path))
	}

	var results []string
	if args.Pattern != "" {
		matches, err := filepath.Glob(filepath.Join(base, args.Pattern))
		if err != nil {
			return "", fmt.Errorf("glob: %w", err)
		}
		for _, m := range matches {
			rel, _ := filepath.Rel(t.workDir, m)
			results = append(results, rel)
		}
	} else {
		entries, err := os.ReadDir(base)
		if err != nil {
			return "", fmt.Errorf("read dir: %w", err)
		}
		for _, e := range entries {
			name := e.Name()
			if e.IsDir() {
				name += "/"
			}
			rel, _ := filepath.Rel(t.workDir, filepath.Join(base, name))
			results = append(results, rel)
		}
	}

	if len(results) == 0 {
		return "(no files found)", nil
	}
	return strings.Join(results, "\n"), nil
}

const maxSearchResults = 50
const maxLineLength = 200

// SearchCode searches for a text pattern across files in the workspace.
type SearchCode struct{ workDir string }

func NewSearchCode(workDir string) *SearchCode { return &SearchCode{workDir: workDir} }

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
	if args.Query == "" {
		return "", fmt.Errorf("query is required")
	}

	searchRoot := t.workDir
	if args.Path != "" {
		searchRoot = filepath.Join(t.workDir, filepath.Clean(args.Path))
	}

	var results []string
	count := 0

	err := filepath.WalkDir(searchRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		if strings.HasPrefix(d.Name(), ".") {
			return nil
		}
		if args.Ext != "" && !strings.HasSuffix(d.Name(), args.Ext) {
			return nil
		}
		if count >= maxSearchResults {
			return filepath.SkipAll
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		rel, _ := filepath.Rel(t.workDir, path)
		queryLower := strings.ToLower(args.Query)
		for i, line := range strings.Split(string(data), "\n") {
			if strings.Contains(strings.ToLower(line), queryLower) {
				display := line
				if len(display) > maxLineLength {
					display = display[:maxLineLength] + "…"
				}
				results = append(results, fmt.Sprintf("%s:%d: %s", rel, i+1, display))
				count++
				if count >= maxSearchResults {
					break
				}
			}
		}
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("walk: %w", err)
	}

	if len(results) == 0 {
		return fmt.Sprintf("no matches found for %q", args.Query), nil
	}
	if count >= maxSearchResults {
		results = append(results, fmt.Sprintf("... (truncated at %d results)", maxSearchResults))
	}
	return strings.Join(results, "\n"), nil
}
