package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// maxReadFileSize is the upper limit for read_file. Files larger than this
// should be read in sections with read_file_range.
// Chosen to comfortably fit a ~500-line Go source file (avg 60 chars/line ≈ 30 KB).
const maxReadFileSize = 32 * 1024 // 32 KB

// ReadFile reads the full contents of a file within the workspace.
// Returns an error hint when the file exceeds maxReadFileSize.
type ReadFile struct{ workDir string }

func NewReadFile(workDir string) *ReadFile { return &ReadFile{workDir: workDir} }

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

	absPath := filepath.Join(t.workDir, filepath.Clean(args.Path))

	info, err := os.Stat(absPath)
	if err != nil {
		return "", fmt.Errorf("stat file: %w", err)
	}
	if info.Size() > maxReadFileSize {
		data, err := os.ReadFile(absPath)
		if err != nil {
			return "", fmt.Errorf("read file: %w", err)
		}
		lineCount := bytes.Count(data, []byte("\n")) + 1
		return fmt.Sprintf(
			"file too large to read at once (%d bytes, ~%d lines). Use read_file_range with start_line and end_line to read it in sections.",
			info.Size(), lineCount,
		), nil
	}

	content, err := os.ReadFile(absPath)
	if err != nil {
		return "", fmt.Errorf("read file: %w", err)
	}
	return string(content), nil
}

// ReadFileRange reads a specific range of lines from a file (1-indexed, inclusive).
type ReadFileRange struct{ workDir string }

func NewReadFileRange(workDir string) *ReadFileRange { return &ReadFileRange{workDir: workDir} }

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
	if args.StartLine < 1 {
		return "", fmt.Errorf("start_line must be >= 1")
	}

	data, err := os.ReadFile(filepath.Join(t.workDir, filepath.Clean(args.Path)))
	if err != nil {
		return "", fmt.Errorf("read file: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	total := len(lines)

	start := args.StartLine - 1 // convert to 0-indexed
	end := total                // default: read to end
	if args.EndLine > 0 {
		end = args.EndLine
	}

	if start >= total {
		return fmt.Sprintf("start_line %d exceeds file length (%d lines)", args.StartLine, total), nil
	}
	if end > total {
		end = total
	}

	selected := lines[start:end]
	var sb strings.Builder
	for i, line := range selected {
		fmt.Fprintf(&sb, "%d: %s\n", args.StartLine+i, line)
	}
	return sb.String(), nil
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
