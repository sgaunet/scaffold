package scaffold_test

import (
	"go/parser"
	"go/token"
	"os"
	"strings"
	"testing"

	"github.com/sgaunet/scaffold/internal/scaffold"
)

// TestCoreHasNoCLIDependencies enforces constitution II: the CLI-free core must
// not import the command or interactive-UI libraries. Only internal/cli may.
func TestCoreHasNoCLIDependencies(t *testing.T) {
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

// TestBinaryNameConsistentAcrossFiles asserts the binary name agrees across the
// GoReleaser config, the Dockerfile, and the Taskfile (FR-019). They all draw
// from the single render context, so they cannot drift.
func TestBinaryNameConsistentAcrossFiles(t *testing.T) {
	t.Parallel()
	const bin = "consistent-bin"
	p := scaffold.ProjectProfile{
		ProjectName: bin, Binary: bin, ModulePath: "github.com/acme/x",
		Owner: "acme", Platform: scaffold.PlatformGitHub, Docker: true,
		MainPath: "./cmd/" + bin, Registry: "ghcr.io/acme/" + bin,
		GoVersion: "1.26.1",
	}
	ctx := scaffold.NewRenderContext(p)
	reg := scaffold.NewRegistry()

	wantContain := map[string]bool{"goreleaser": true, "dockerfile": true, "taskfile": true}
	seen := 0
	for _, tmpl := range reg.Applicable(p) {
		if !wantContain[tmpl.Name] {
			continue
		}
		out, err := scaffold.Render(tmpl, ctx)
		if err != nil {
			t.Fatalf("render %s: %v", tmpl.Name, err)
		}
		if !strings.Contains(string(out), bin) {
			t.Fatalf("file %s does not reference binary %q", tmpl.Name, bin)
		}
		seen++
	}
	if seen != len(wantContain) {
		t.Fatalf("checked %d files, expected %d (goreleaser/dockerfile/taskfile)", seen, len(wantContain))
	}
}
