package service

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const maxReadFileSize = 32 * 1024 // 32 KB — ~500-line Go file

// FileReadResult is returned by Read when the file is within size limits.
type FileReadResult struct {
	Content string
}

// FileTooBigError is returned by Read when the file exceeds the size limit.
type FileTooBigError struct {
	Size      int64
	LineCount int
}

func (e *FileTooBigError) Error() string {
	return fmt.Sprintf("file too large to read at once (%d bytes, ~%d lines). Use read_file_range with start_line and end_line to read it in sections.", e.Size, e.LineCount)
}

// FilesystemService provides workspace-scoped file operations.
type FilesystemService interface {
	Read(path string) (string, error)
	ReadRange(path string, startLine, endLine int) (string, error)
	Write(path, content string) error
	List(dir, pattern string) ([]string, error)
	Search(query, subPath, ext string) ([]string, error)
}

type filesystemService struct {
	workDir string
}

func NewFilesystemService(workDir string) FilesystemService {
	return &filesystemService{workDir: workDir}
}

func (s *filesystemService) abs(rel string) string {
	return filepath.Join(s.workDir, filepath.Clean(rel))
}

func (s *filesystemService) Read(path string) (string, error) {
	absPath := s.abs(path)
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
		return "", &FileTooBigError{Size: info.Size(), LineCount: lineCount}
	}
	content, err := os.ReadFile(absPath)
	if err != nil {
		return "", fmt.Errorf("read file: %w", err)
	}
	return string(content), nil
}

func (s *filesystemService) ReadRange(path string, startLine, endLine int) (string, error) {
	if startLine < 1 {
		return "", fmt.Errorf("%w: start_line must be >= 1", ErrInvalid)
	}
	data, err := os.ReadFile(s.abs(path))
	if err != nil {
		return "", fmt.Errorf("read file: %w", err)
	}
	lines := strings.Split(string(data), "\n")
	total := len(lines)

	start := startLine - 1
	end := total
	if endLine > 0 {
		end = endLine
	}
	if start >= total {
		return fmt.Sprintf("start_line %d exceeds file length (%d lines)", startLine, total), nil
	}
	if end > total {
		end = total
	}

	var sb strings.Builder
	for i, line := range lines[start:end] {
		fmt.Fprintf(&sb, "%d: %s\n", startLine+i, line)
	}
	return sb.String(), nil
}

func (s *filesystemService) Write(path, content string) error {
	absPath := s.abs(path)
	if err := os.MkdirAll(filepath.Dir(absPath), 0o755); err != nil {
		return fmt.Errorf("create dirs: %w", err)
	}
	if err := os.WriteFile(absPath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write file: %w", err)
	}
	return nil
}

func (s *filesystemService) List(dir, pattern string) ([]string, error) {
	base := s.workDir
	if dir != "" {
		base = filepath.Join(s.workDir, filepath.Clean(dir))
	}

	var results []string
	if pattern != "" {
		matches, err := filepath.Glob(filepath.Join(base, pattern))
		if err != nil {
			return nil, fmt.Errorf("glob: %w", err)
		}
		for _, m := range matches {
			rel, _ := filepath.Rel(s.workDir, m)
			results = append(results, rel)
		}
		return results, nil
	}

	entries, err := os.ReadDir(base)
	if err != nil {
		return nil, fmt.Errorf("read dir: %w", err)
	}
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() {
			name += "/"
		}
		rel, _ := filepath.Rel(s.workDir, filepath.Join(base, name))
		results = append(results, rel)
	}
	return results, nil
}

const maxSearchResults = 50
const maxLineLength = 200

func (s *filesystemService) Search(query, subPath, ext string) ([]string, error) {
	if query == "" {
		return nil, fmt.Errorf("%w: query is required", ErrInvalid)
	}
	searchRoot := s.workDir
	if subPath != "" {
		searchRoot = filepath.Join(s.workDir, filepath.Clean(subPath))
	}

	var results []string
	count := 0
	queryLower := strings.ToLower(query)

	err := filepath.WalkDir(searchRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		if strings.HasPrefix(d.Name(), ".") {
			return nil
		}
		if ext != "" && !strings.HasSuffix(d.Name(), ext) {
			return nil
		}
		if count >= maxSearchResults {
			return filepath.SkipAll
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		rel, _ := filepath.Rel(s.workDir, path)
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
		return nil, fmt.Errorf("walk: %w", err)
	}
	if count >= maxSearchResults {
		results = append(results, fmt.Sprintf("... (truncated at %d results)", maxSearchResults))
	}
	return results, nil
}
