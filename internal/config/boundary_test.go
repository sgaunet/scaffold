package config_test

import (
	"go/parser"
	"go/token"
	"os"
	"strings"
	"testing"
)

// TestNoCLIDependencies enforces the constitution II boundary: the CLI-free
// config package must not import the command/UI libraries. Only internal/cli may
// depend on cobra or huh.
func TestNoCLIDependencies(t *testing.T) {
	t.Parallel()
	forbidden := []string{"spf13/cobra", "charmbracelet/huh"}

	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read package dir: %v", err)
	}
	fset := token.NewFileSet()
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		f, perr := parser.ParseFile(fset, name, nil, parser.ImportsOnly)
		if perr != nil {
			t.Fatalf("parse %s: %v", name, perr)
		}
		for _, imp := range f.Imports {
			for _, bad := range forbidden {
				if strings.Contains(imp.Path.Value, bad) {
					t.Errorf("%s imports forbidden CLI dependency %s", name, imp.Path.Value)
				}
			}
		}
	}
}
