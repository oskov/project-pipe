package service

import (
	"errors"
	"strings"
	"testing"
)

// fixture is the path to the test Go file relative to the service package.
const fixture = "testdata/sample.go"

func newTestGoParseService(t *testing.T) GoParseService {
	t.Helper()
	// workDir is the package directory — testdata/sample.go is resolved from here.
	return NewGoParseService(".")
}

// ── ListDefinitions ────────────────────────────────────────────────────────

func TestGoParseService_ListDefinitions_AllKinds(t *testing.T) {
	svc := newTestGoParseService(t)
	out, err := svc.ListDefinitions([]string{fixture})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cases := []struct {
		want string
		desc string
	}{
		{"func   NewItem", "standalone function"},
		{"func   (*Item).String", "pointer receiver method"},
		{"func   (*Item).Validate", "pointer receiver method (Validate)"},
		{"type   Item", "struct type"},
		{"type   Repository", "interface type"},
		{"type   mapStore", "type alias"},
		{"const  MaxItems", "single const"},
		{"const  ErrNotFound", "grouped const (ErrNotFound)"},
		{"const  ErrInvalid", "grouped const (ErrInvalid)"},
		{"var    DefaultTimeout", "single var"},
		{"var    RequestCount", "grouped var (RequestCount)"},
		{"var    ErrorCount", "grouped var (ErrorCount)"},
	}

	for _, tc := range cases {
		if !strings.Contains(out, tc.want) {
			t.Errorf("ListDefinitions: expected %q (%s) in output\ngot:\n%s", tc.want, tc.desc, out)
		}
	}
}

func TestGoParseService_ListDefinitions_TypeKinds(t *testing.T) {
	svc := newTestGoParseService(t)
	out, err := svc.ListDefinitions([]string{fixture})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out, "struct") {
		t.Errorf("expected 'struct' kind in output, got:\n%s", out)
	}
	if !strings.Contains(out, "interface") {
		t.Errorf("expected 'interface' kind in output, got:\n%s", out)
	}
}

func TestGoParseService_ListDefinitions_MultipleFiles(t *testing.T) {
	svc := newTestGoParseService(t)
	// Pass the same file twice — both sections should appear.
	out, err := svc.ListDefinitions([]string{fixture, fixture})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	count := strings.Count(out, "=== "+fixture+" ===")
	if count != 2 {
		t.Errorf("expected 2 file sections, got %d\noutput:\n%s", count, out)
	}
}

func TestGoParseService_ListDefinitions_NonExistentFile(t *testing.T) {
	svc := newTestGoParseService(t)
	out, err := svc.ListDefinitions([]string{"testdata/nonexistent.go"})
	// Non-existent file is surfaced as an error line in the output, not as err.
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "error:") {
		t.Errorf("expected 'error:' in output for missing file, got:\n%s", out)
	}
}

func TestGoParseService_ListDefinitions_EmptyList(t *testing.T) {
	svc := newTestGoParseService(t)
	_, err := svc.ListDefinitions(nil)
	if !errors.Is(err, ErrInvalid) {
		t.Errorf("expected ErrInvalid for empty files list, got: %v", err)
	}
}

// ── ReadDefinition ─────────────────────────────────────────────────────────

func TestGoParseService_ReadDefinition_Function(t *testing.T) {
	svc := newTestGoParseService(t)
	out, err := svc.ReadDefinition(fixture, "NewItem")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "func NewItem") {
		t.Errorf("expected func signature in output, got:\n%s", out)
	}
	// Doc comment should be included.
	if !strings.Contains(out, "// NewItem creates") {
		t.Errorf("expected doc comment in output, got:\n%s", out)
	}
}

func TestGoParseService_ReadDefinition_Method(t *testing.T) {
	svc := newTestGoParseService(t)
	out, err := svc.ReadDefinition(fixture, "String")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "func (i *Item) String()") {
		t.Errorf("expected method signature in output, got:\n%s", out)
	}
}

func TestGoParseService_ReadDefinition_Struct(t *testing.T) {
	svc := newTestGoParseService(t)
	out, err := svc.ReadDefinition(fixture, "Item")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "type Item struct") {
		t.Errorf("expected struct definition in output, got:\n%s", out)
	}
	// Fields should be present.
	if !strings.Contains(out, "ID") || !strings.Contains(out, "Value") {
		t.Errorf("expected struct fields in output, got:\n%s", out)
	}
	if !strings.Contains(out, "// Item represents") {
		t.Errorf("expected doc comment in output, got:\n%s", out)
	}
}

func TestGoParseService_ReadDefinition_Interface(t *testing.T) {
	svc := newTestGoParseService(t)
	out, err := svc.ReadDefinition(fixture, "Repository")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "type Repository interface") {
		t.Errorf("expected interface definition in output, got:\n%s", out)
	}
	if !strings.Contains(out, "Get(id string)") {
		t.Errorf("expected interface method in output, got:\n%s", out)
	}
}

func TestGoParseService_ReadDefinition_SingleConst(t *testing.T) {
	svc := newTestGoParseService(t)
	out, err := svc.ReadDefinition(fixture, "MaxItems")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "MaxItems") {
		t.Errorf("expected const name in output, got:\n%s", out)
	}
	if !strings.Contains(out, "100") {
		t.Errorf("expected const value in output, got:\n%s", out)
	}
}

func TestGoParseService_ReadDefinition_GroupedConst(t *testing.T) {
	svc := newTestGoParseService(t)
	// ErrNotFound lives in a grouped const block — whole block should be returned.
	out, err := svc.ReadDefinition(fixture, "ErrNotFound")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "ErrNotFound") {
		t.Errorf("expected ErrNotFound in output, got:\n%s", out)
	}
	// The whole block includes ErrInvalid too.
	if !strings.Contains(out, "ErrInvalid") {
		t.Errorf("expected sibling const ErrInvalid in grouped block output, got:\n%s", out)
	}
}

func TestGoParseService_ReadDefinition_GroupedVar(t *testing.T) {
	svc := newTestGoParseService(t)
	out, err := svc.ReadDefinition(fixture, "RequestCount")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "RequestCount") {
		t.Errorf("expected RequestCount in output, got:\n%s", out)
	}
	if !strings.Contains(out, "ErrorCount") {
		t.Errorf("expected sibling var ErrorCount in grouped block output, got:\n%s", out)
	}
}

func TestGoParseService_ReadDefinition_NotFound(t *testing.T) {
	svc := newTestGoParseService(t)
	out, err := svc.ReadDefinition(fixture, "DoesNotExist")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "not found") {
		t.Errorf("expected 'not found' message, got:\n%s", out)
	}
}

func TestGoParseService_ReadDefinition_EmptyArgs(t *testing.T) {
	svc := newTestGoParseService(t)

	_, err := svc.ReadDefinition("", "NewItem")
	if !errors.Is(err, ErrInvalid) {
		t.Errorf("expected ErrInvalid for empty file, got: %v", err)
	}

	_, err = svc.ReadDefinition(fixture, "")
	if !errors.Is(err, ErrInvalid) {
		t.Errorf("expected ErrInvalid for empty name, got: %v", err)
	}
}

func TestGoParseService_ReadDefinition_NonExistentFile(t *testing.T) {
	svc := newTestGoParseService(t)
	_, err := svc.ReadDefinition("testdata/nonexistent.go", "Foo")
	if err == nil {
		t.Error("expected error for non-existent file, got nil")
	}
}
