package scaffold_test

import (
	"strings"
	"testing"

	"github.com/sgaunet/scaffold/internal/scaffold"
)

// TestRenderMatrixNoResidualDelimiters renders every applicable template across
// the platform × docker matrix and asserts the project values are substituted
// with zero leftover [[ / ]] delimiters (SC-002). This is the assertion-based
// stand-in for the golden-file suite.
func TestRenderMatrixNoResidualDelimiters(t *testing.T) {
	t.Parallel()
	platforms := []scaffold.PlatformID{
		scaffold.PlatformNone, scaffold.PlatformGitHub,
		scaffold.PlatformGitLab, scaffold.PlatformForgejo,
	}
	for _, plat := range platforms {
		for _, docker := range []bool{false, true} {
			name := string(plat)
			if name == "" {
				name = "none"
			}
			if docker {
				name += "-docker"
			}
			t.Run(name, func(t *testing.T) {
				t.Parallel()
				p := scaffold.ProjectProfile{
					ProjectName: "demo", Binary: "demo",
					ModulePath: "github.com/acme/demo", Owner: "acme",
					Host: "github.com", Platform: plat, Docker: docker,
					MainPath: "./cmd/demo", Registry: "ghcr.io/acme/demo",
					VersionPackage: "github.com/acme/demo/internal/cli",
					GoVersion:      "1.26.1", TaskVersion: "3.40.1",
					GolangciVersion: "2.1.6", GoreleaserVersion: "2.5.1",
					FundingUser: "acme",
				}
				ctx := scaffold.NewRenderContext(p)
				reg := scaffold.NewRegistry()
				for _, tmpl := range reg.Applicable(p) {
					out, err := scaffold.Render(tmpl, ctx)
					if err != nil {
						t.Fatalf("render %s: %v", tmpl.Name, err)
					}
					s := string(out)
					if strings.Contains(s, "[[") || strings.Contains(s, "]]") {
						t.Fatalf("template %s has residual delimiters:\n%s", tmpl.Name, s)
					}
				}
			})
		}
	}
}

// TestBinaryNameSubstituted verifies the binary name reaches the key files.
func TestBinaryNameSubstituted(t *testing.T) {
	t.Parallel()
	p := scaffold.ProjectProfile{
		ProjectName: "widget", Binary: "widget", MainPath: "./cmd/widget",
		Platform: scaffold.PlatformGitHub, GoVersion: "1.26.1",
	}
	ctx := scaffold.NewRenderContext(p)
	reg := scaffold.NewRegistry()
	want := map[string]bool{"goreleaser": true, "taskfile": true}
	for _, tmpl := range reg.Applicable(p) {
		if !want[tmpl.Name] {
			continue
		}
		out, err := scaffold.Render(tmpl, ctx)
		if err != nil {
			t.Fatalf("render %s: %v", tmpl.Name, err)
		}
		if !strings.Contains(string(out), "widget") {
			t.Fatalf("template %s does not contain binary name", tmpl.Name)
		}
	}
}
