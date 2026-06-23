package scaffold_test

import (
	"strings"
	"testing"

	"github.com/sgaunet/scaffold/internal/scaffold"
)

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
