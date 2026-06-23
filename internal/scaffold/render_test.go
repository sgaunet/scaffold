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
					GolangciVersion: "2.1.6", GoreleaserVersion: "2.16.0",
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

// renderNamed renders the single registered template with the given name.
func renderNamed(t *testing.T, p scaffold.ProjectProfile, name string) string {
	t.Helper()
	ctx := scaffold.NewRenderContext(p)
	for _, tmpl := range scaffold.NewRegistry().Applicable(p) {
		if tmpl.Name != name {
			continue
		}
		out, err := scaffold.Render(tmpl, ctx)
		if err != nil {
			t.Fatalf("render %s: %v", name, err)
		}
		return string(out)
	}
	t.Fatalf("template %q not applicable to profile", name)
	return ""
}

// TestHomebrewRendering verifies the brews block is gated by the Homebrew flag
// and that the release workflow exposes the tap token only when enabled.
func TestHomebrewRendering(t *testing.T) {
	t.Parallel()
	base := scaffold.ProjectProfile{
		ProjectName: "demo", Binary: "demo", ModulePath: "github.com/acme/demo",
		Owner: "acme", Host: "github.com", Platform: scaffold.PlatformGitHub,
		MainPath: "./cmd/demo", GoVersion: "1.26.1",
	}

	t.Run("on", func(t *testing.T) {
		t.Parallel()
		p := base
		p.Homebrew = true
		p.HomebrewTap = "homebrew-tap"
		grl := renderNamed(t, p, "goreleaser")
		for _, want := range []string{"homebrew_casks:", "name: homebrew-tap", "owner: acme", "name: demo", "shell_parameter_format: cobra", "HOMEBREW_TAP_TOKEN"} {
			if !strings.Contains(grl, want) {
				t.Fatalf("goreleaser missing %q:\n%s", want, grl)
			}
		}
		if strings.Contains(grl, "[[") || strings.Contains(grl, "]]") {
			t.Fatalf("residual delimiters in homebrew_casks block:\n%s", grl)
		}
		rel := renderNamed(t, p, "github/workflows/release")
		if !strings.Contains(rel, "HOMEBREW_TAP_TOKEN: ${{ secrets.HOMEBREW_TAP_TOKEN }}") {
			t.Fatalf("release workflow missing tap token:\n%s", rel)
		}
	})

	t.Run("custom tap", func(t *testing.T) {
		t.Parallel()
		p := base
		p.Homebrew = true
		p.HomebrewTap = "homebrew-tools"
		if grl := renderNamed(t, p, "goreleaser"); !strings.Contains(grl, "name: homebrew-tools") {
			t.Fatalf("goreleaser missing custom tap name:\n%s", grl)
		}
	})

	t.Run("off", func(t *testing.T) {
		t.Parallel()
		p := base // Homebrew defaults to false
		if grl := renderNamed(t, p, "goreleaser"); strings.Contains(grl, "homebrew_casks:") {
			t.Fatalf("homebrew_casks block present without --homebrew:\n%s", grl)
		}
		if rel := renderNamed(t, p, "github/workflows/release"); strings.Contains(rel, "HOMEBREW_TAP_TOKEN") {
			t.Fatalf("tap token present without --homebrew:\n%s", rel)
		}
	})
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
