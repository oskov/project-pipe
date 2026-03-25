package service

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// GoParseService provides Go source code inspection using the standard AST.
type GoParseService interface {
	// ListDefinitions returns a formatted summary of top-level definitions in a
	// single Go source file (relative to workDir). kinds optionally restricts
	// output to specific definition kinds: "func", "method", "type", "struct",
	// "interface", "var", "const". Empty or nil means all kinds.
	ListDefinitions(file string, kinds []string) (string, error)
	// ReadDefinition returns the full source (including doc comment) of the
	// named top-level definition from the given file.
	ReadDefinition(file, name string) (string, error)
}

type goParseService struct {
	workDir string
}

func NewGoParseService(workDir string) GoParseService {
	return &goParseService{workDir: workDir}
}

func (s *goParseService) abs(rel string) string {
	return filepath.Join(s.workDir, filepath.Clean(rel))
}

func (s *goParseService) ListDefinitions(file string, kinds []string) (string, error) {
	if file == "" {
		return "", fmt.Errorf("%w: file is required", ErrInvalid)
	}
	filter, err := parseKindFilter(kinds)
	if err != nil {
		return "", err
	}
	result, err := listDefinitionsInFile(s.abs(file), filter)
	if err != nil {
		return "", fmt.Errorf("parse %s: %w", file, err)
	}
	return result, nil
}

func (s *goParseService) ReadDefinition(file, name string) (string, error) {
	if file == "" || name == "" {
		return "", fmt.Errorf("%w: file and name are required", ErrInvalid)
	}
	absPath := s.abs(file)
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
			if d.Name.Name == name {
				return extractSource(fset, src, d.Doc, d.Pos(), d.End()), nil
			}
		case *ast.GenDecl:
			for _, spec := range d.Specs {
				switch s := spec.(type) {
				case *ast.TypeSpec:
					if s.Name.Name == name {
						return extractSource(fset, src, d.Doc, d.Pos(), d.End()), nil
					}
				case *ast.ValueSpec:
					for _, n := range s.Names {
						if n.Name == name {
							return extractSource(fset, src, d.Doc, d.Pos(), d.End()), nil
						}
					}
				}
			}
		}
	}
	return fmt.Sprintf("definition %q not found in %s", name, file), nil
}

// ── helpers ────────────────────────────────────────────────────────────────

// validKinds is the set of accepted filter values.
var validKinds = map[string]bool{
	"func": true, "method": true,
	"type": true, "struct": true, "interface": true,
	"var": true, "const": true,
}

// parseKindFilter validates and normalises the caller-supplied kinds list.
// Returns nil map (= all) when kinds is empty.
func parseKindFilter(kinds []string) (map[string]bool, error) {
	if len(kinds) == 0 {
		return nil, nil
	}
	filter := make(map[string]bool, len(kinds))
	for _, k := range kinds {
		if !validKinds[k] {
			return nil, fmt.Errorf("%w: unknown kind %q (valid: func, method, type, struct, interface, var, const)", ErrInvalid, k)
		}
		filter[k] = true
	}
	// "type" implies struct + interface + all other type sub-kinds.
	// "struct" / "interface" without "type" should still match their typeKind.
	return filter, nil
}

func listDefinitionsInFile(path string, filter map[string]bool) (string, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return "", err
	}

	// wantKind returns true when filter is nil (all) or contains the key.
	wantKind := func(k string) bool { return filter == nil || filter[k] }

	// wantType returns true when we should include a type of the given typeKind string.
	wantType := func(kind string) bool {
		if filter == nil {
			return true
		}
		if filter["type"] {
			return true
		}
		return filter[kind] // e.g. filter["struct"] or filter["interface"]
	}

	var sb strings.Builder
	for _, decl := range f.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			isMethod := d.Recv != nil && len(d.Recv.List) > 0
			if isMethod && !wantKind("method") {
				continue
			}
			if !isMethod && !wantKind("func") {
				continue
			}
			line := fset.Position(d.Pos()).Line
			sig := formatFuncSignature(d)
			fmt.Fprintf(&sb, "  %4d  func   %s\n", line, sig)
		case *ast.GenDecl:
			switch d.Tok {
			case token.TYPE:
				for _, spec := range d.Specs {
					ts := spec.(*ast.TypeSpec)
					kind := typeKind(ts.Type)
					if !wantType(kind) {
						continue
					}
					line := fset.Position(ts.Pos()).Line
					fmt.Fprintf(&sb, "  %4d  type   %-30s %s\n", line, ts.Name.Name, kind)
				}
			case token.VAR:
				if !wantKind("var") {
					continue
				}
				for _, spec := range d.Specs {
					vs := spec.(*ast.ValueSpec)
					line := fset.Position(vs.Pos()).Line
					for _, n := range vs.Names {
						fmt.Fprintf(&sb, "  %4d  var    %s\n", line, n.Name)
					}
				}
			case token.CONST:
				if !wantKind("const") {
					continue
				}
				for _, spec := range d.Specs {
					vs := spec.(*ast.ValueSpec)
					line := fset.Position(vs.Pos()).Line
					for _, n := range vs.Names {
						fmt.Fprintf(&sb, "  %4d  const  %s\n", line, n.Name)
					}
				}
			}
		}
	}
	return sb.String(), nil
}

func extractSource(fset *token.FileSet, src []byte, doc *ast.CommentGroup, pos, end token.Pos) string {
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

func receiverType(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.StarExpr:
		return "*" + receiverType(e.X)
	case *ast.Ident:
		return e.Name
	default:
		return "?"
	}
}

// formatFuncSignature returns e.g. "(*Item).String() string" or "NewItem(name string) (*Item, error)".
func formatFuncSignature(d *ast.FuncDecl) string {
	var sb strings.Builder
	if d.Recv != nil && len(d.Recv.List) > 0 {
		fmt.Fprintf(&sb, "(%s).%s", receiverType(d.Recv.List[0].Type), d.Name.Name)
	} else {
		sb.WriteString(d.Name.Name)
	}
	sb.WriteString(formatParams(d.Type.Params))
	if ret := formatResults(d.Type.Results); ret != "" {
		sb.WriteString(" ")
		sb.WriteString(ret)
	}
	return sb.String()
}

// formatParams formats a parameter list as "(name type, ...)".
func formatParams(fl *ast.FieldList) string {
	if fl == nil {
		return "()"
	}
	var parts []string
	for _, f := range fl.List {
		typ := formatExpr(f.Type)
		if len(f.Names) == 0 {
			parts = append(parts, typ)
		} else {
			var names []string
			for _, n := range f.Names {
				names = append(names, n.Name)
			}
			parts = append(parts, strings.Join(names, ", ")+" "+typ)
		}
	}
	return "(" + strings.Join(parts, ", ") + ")"
}

// formatResults formats return types. Single unnamed result → bare type; otherwise → "(type, ...)".
func formatResults(fl *ast.FieldList) string {
	if fl == nil || len(fl.List) == 0 {
		return ""
	}
	// single unnamed result
	if len(fl.List) == 1 && len(fl.List[0].Names) == 0 {
		return formatExpr(fl.List[0].Type)
	}
	var parts []string
	for _, f := range fl.List {
		typ := formatExpr(f.Type)
		if len(f.Names) == 0 {
			parts = append(parts, typ)
		} else {
			var names []string
			for _, n := range f.Names {
				names = append(names, n.Name)
			}
			parts = append(parts, strings.Join(names, ", ")+" "+typ)
		}
	}
	return "(" + strings.Join(parts, ", ") + ")"
}

// formatExpr converts an ast.Expr to its Go source representation.
func formatExpr(expr ast.Expr) string {
	if expr == nil {
		return ""
	}
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.StarExpr:
		return "*" + formatExpr(e.X)
	case *ast.SelectorExpr:
		return formatExpr(e.X) + "." + e.Sel.Name
	case *ast.ArrayType:
		if e.Len == nil {
			return "[]" + formatExpr(e.Elt)
		}
		return "[" + formatExpr(e.Len) + "]" + formatExpr(e.Elt)
	case *ast.MapType:
		return "map[" + formatExpr(e.Key) + "]" + formatExpr(e.Value)
	case *ast.ChanType:
		switch e.Dir {
		case ast.RECV:
			return "<-chan " + formatExpr(e.Value)
		case ast.SEND:
			return "chan<- " + formatExpr(e.Value)
		default:
			return "chan " + formatExpr(e.Value)
		}
	case *ast.Ellipsis:
		return "..." + formatExpr(e.Elt)
	case *ast.InterfaceType:
		return "interface{}"
	case *ast.StructType:
		return "struct{}"
	case *ast.FuncType:
		return "func" + formatParams(e.Params)
	case *ast.BasicLit:
		return e.Value
	case *ast.ParenExpr:
		return "(" + formatExpr(e.X) + ")"
	case *ast.IndexExpr:
		return formatExpr(e.X) + "[" + formatExpr(e.Index) + "]"
	default:
		return "?"
	}
}

func typeKind(expr ast.Expr) string {
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
