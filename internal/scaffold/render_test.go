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
		p.HomebrewTap = "homebrew-tools"
		grl := renderNamed(t, p, "goreleaser")
		for _, want := range []string{"homebrew_casks:", "name: homebrew-tools", "owner: acme", "name: demo", "shell_parameter_format: cobra", "HOMEBREW_TAP_TOKEN"} {
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

// TestTaskfileTokenEnvHandling verifies the snapshot task clears every other
// platform's release token env var (so goreleaser's host auto-detection isn't
// confused by multiple tokens set in a dev shell) while leaving the current
// platform's own token untouched, and that the release task carries no token
// prefix at all.
func TestTaskfileTokenEnvHandling(t *testing.T) {
	t.Parallel()
	base := scaffold.ProjectProfile{
		ProjectName: "demo", Binary: "demo", ModulePath: "github.com/acme/demo",
		Owner: "acme", Host: "github.com", MainPath: "./cmd/demo", GoVersion: "1.26.1",
	}

	cases := []struct {
		platform scaffold.PlatformID
		want     []string
		absent   []string
	}{
		{scaffold.PlatformGitHub, []string{`GITLAB_TOKEN=""`, `GITEA_TOKEN=""`}, []string{`GITHUB_TOKEN=""`}},
		{scaffold.PlatformGitLab, []string{`GITHUB_TOKEN=""`, `GITEA_TOKEN=""`}, []string{`GITLAB_TOKEN=""`}},
		{scaffold.PlatformForgejo, []string{`GITHUB_TOKEN=""`, `GITLAB_TOKEN=""`}, []string{`GITEA_TOKEN=""`}},
	}
	for _, c := range cases {
		t.Run(string(c.platform), func(t *testing.T) {
			t.Parallel()
			p := base
			p.Platform = c.platform
			tf := renderNamed(t, p, "taskfile")
			for _, want := range c.want {
				if !strings.Contains(tf, want) {
					t.Fatalf("taskfile for %s missing %q:\n%s", c.platform, want, tf)
				}
			}
			for _, absent := range c.absent {
				if strings.Contains(tf, absent) {
					t.Fatalf("taskfile for %s must not contain %q:\n%s", c.platform, absent, tf)
				}
			}
			if !strings.Contains(tf, "- goreleaser --clean\n") {
				t.Fatalf("release task must be a bare goreleaser --clean:\n%s", tf)
			}
		})
	}
}

// TestWorkflowsProvisionToolsWithMise asserts every Actions-style workflow
// template (GitHub and Forgejo) provisions its toolchain through
// jdx/mise-action and runs work through a task target, instead of installing Go
// or bootstrapping mise by hand. mise.toml stays the single source of truth for
// tool versions.
func TestWorkflowsProvisionToolsWithMise(t *testing.T) {
	t.Parallel()
	const wantWorkflows = 4 // lint(er), test, snapshot, release
	required := []string{"mise-action@v4", "install: true", "cache: true", "run: task "}
	forbidden := []string{"actions/setup-go", "mise.run", "mise trust", "mise exec", "actions/cache@"}

	cases := []struct {
		name     string
		platform scaffold.PlatformID
		docker   bool
	}{
		{"github", scaffold.PlatformGitHub, false},
		{"github-docker", scaffold.PlatformGitHub, true},
		{"forgejo", scaffold.PlatformForgejo, false},
		{"forgejo-docker", scaffold.PlatformForgejo, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			p := scaffold.ProjectProfile{
				ProjectName: "demo", Binary: "demo", ModulePath: "github.com/acme/demo",
				Owner: "acme", Host: "git.example.com", Platform: c.platform,
				Docker: c.docker, MainPath: "./cmd/demo",
				Registry: "git.example.com/acme/demo", GoVersion: "1.26.1",
			}
			ctx := scaffold.NewRenderContext(p)
			seen := 0
			for _, tmpl := range scaffold.NewRegistry().Applicable(p) {
				if !strings.Contains(tmpl.Name, "workflows/") {
					continue
				}
				seen++
				out, err := scaffold.Render(tmpl, ctx)
				if err != nil {
					t.Fatalf("render %s: %v", tmpl.Name, err)
				}
				body := string(out)
				for _, want := range required {
					if !strings.Contains(body, want) {
						t.Fatalf("%s must provision tools with mise, missing %q:\n%s", tmpl.Name, want, body)
					}
				}
				for _, banned := range forbidden {
					if strings.Contains(body, banned) {
						t.Fatalf("%s must not bootstrap tools by hand, found %q:\n%s", tmpl.Name, banned, body)
					}
				}
			}
			if seen != wantWorkflows {
				t.Fatalf("expected %d workflow templates for %s, got %d", wantWorkflows, c.platform, seen)
			}
		})
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
