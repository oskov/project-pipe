package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// GoDefinitions lists all top-level definitions in one or more Go source files.
type GoDefinitions struct{ workDir string }

func NewGoDefinitions(workDir string) *GoDefinitions { return &GoDefinitions{workDir: workDir} }

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
	if len(args.Files) == 0 {
		return "", fmt.Errorf("files list is empty")
	}

	var sb strings.Builder
	for _, f := range args.Files {
		absPath := filepath.Join(t.workDir, filepath.Clean(f))
		result, err := goListDefinitions(absPath)
		if err != nil {
			fmt.Fprintf(&sb, "=== %s ===\nerror: %s\n\n", f, err)
			continue
		}
		fmt.Fprintf(&sb, "=== %s ===\n%s\n", f, result)
	}
	return sb.String(), nil
}

func goListDefinitions(path string) (string, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	for _, decl := range f.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			line := fset.Position(d.Pos()).Line
			if d.Recv != nil && len(d.Recv.List) > 0 {
				recv := goReceiverType(d.Recv.List[0].Type)
				fmt.Fprintf(&sb, "  %4d  func   (%s).%s\n", line, recv, d.Name.Name)
			} else {
				fmt.Fprintf(&sb, "  %4d  func   %s\n", line, d.Name.Name)
			}

		case *ast.GenDecl:
			switch d.Tok {
			case token.TYPE:
				for _, spec := range d.Specs {
					ts := spec.(*ast.TypeSpec)
					line := fset.Position(ts.Pos()).Line
					fmt.Fprintf(&sb, "  %4d  type   %-30s %s\n", line, ts.Name.Name, goTypeKind(ts.Type))
				}
			case token.VAR:
				for _, spec := range d.Specs {
					vs := spec.(*ast.ValueSpec)
					line := fset.Position(vs.Pos()).Line
					for _, name := range vs.Names {
						fmt.Fprintf(&sb, "  %4d  var    %s\n", line, name.Name)
					}
				}
			case token.CONST:
				for _, spec := range d.Specs {
					vs := spec.(*ast.ValueSpec)
					line := fset.Position(vs.Pos()).Line
					for _, name := range vs.Names {
						fmt.Fprintf(&sb, "  %4d  const  %s\n", line, name.Name)
					}
				}
			}
		}
	}
	return sb.String(), nil
}

// GoReadDefinition returns the full source of a named top-level definition.
type GoReadDefinition struct{ workDir string }

func NewGoReadDefinition(workDir string) *GoReadDefinition {
	return &GoReadDefinition{workDir: workDir}
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

	absPath := filepath.Join(t.workDir, filepath.Clean(args.File))
	src, err := os.ReadFile(absPath)
	if err != nil {
		return "", fmt.Errorf("read file: %w", err)
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, absPath, src, parser.ParseComments)
	if err != nil {
		return "", fmt.Errorf("parse file: %w", err)
	}

	for _, decl := range f.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			if d.Name.Name != args.Name {
				continue
			}
			return goExtractSource(fset, src, d.Doc, d.Pos(), d.End()), nil

		case *ast.GenDecl:
			for _, spec := range d.Specs {
				switch s := spec.(type) {
				case *ast.TypeSpec:
					if s.Name.Name != args.Name {
						continue
					}
					return goExtractSource(fset, src, d.Doc, d.Pos(), d.End()), nil
				case *ast.ValueSpec:
					for _, n := range s.Names {
						if n.Name != args.Name {
							continue
						}
						// Return the whole GenDecl — preserves grouping and doc comment.
						return goExtractSource(fset, src, d.Doc, d.Pos(), d.End()), nil
					}
				}
			}
		}
	}

	return fmt.Sprintf("definition %q not found in %s", args.Name, args.File), nil
}

// ── helpers ────────────────────────────────────────────────────────────────

func goExtractSource(fset *token.FileSet, src []byte, doc *ast.CommentGroup, pos, end token.Pos) string {
	start := fset.Position(pos).Offset
	if doc != nil {
		start = fset.Position(doc.Pos()).Offset
	}
	finish := fset.Position(end).Offset
	if finish > len(src) {
		finish = len(src)
	}
	return string(src[start:finish])
}

func goReceiverType(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.StarExpr:
		return "*" + goReceiverType(e.X)
	case *ast.Ident:
		return e.Name
	default:
		return "?"
	}
}

func goTypeKind(expr ast.Expr) string {
	switch expr.(type) {
	case *ast.StructType:
		return "struct"
	case *ast.InterfaceType:
		return "interface"
	case *ast.MapType:
		return "map"
	case *ast.ArrayType:
		return "slice"
	case *ast.ChanType:
		return "chan"
	case *ast.FuncType:
		return "func"
	default:
		return "alias"
	}
}
