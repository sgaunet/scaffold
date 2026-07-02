package scaffold_test

import (
	"strings"
	"testing"

	"github.com/sgaunet/scaffold/internal/scaffold"
)

func applicableNames(p scaffold.ProjectProfile) map[string]bool {
	reg := scaffold.NewRegistry()
	names := map[string]bool{}
	for _, t := range reg.Applicable(p) {
		names[t.Name] = true
	}
	return names
}

// TestPlatformIsolation: each platform's set excludes the others; GitHub-only
// extras gate correctly; an unset platform yields zero CI/extra files
// (SC-004, FR-003, FR-006).
func TestPlatformIsolation(t *testing.T) {
	t.Parallel()

	base := scaffold.ProjectProfile{ProjectName: "demo", Binary: "demo", MainPath: "./cmd/demo"}

	github := base
	github.Platform = scaffold.PlatformGitHub
	gh := applicableNames(github)
	if !gh["github/dependabot"] || !gh["github/FUNDING"] {
		t.Fatalf("github run missing dependabot/FUNDING: %v", gh)
	}
	for name := range gh {
		if strings.HasPrefix(name, "gitlab") || strings.HasPrefix(name, "forgejo") {
			t.Fatalf("github run leaked %s", name)
		}
	}

	forgejo := base
	forgejo.Platform = scaffold.PlatformForgejo
	fj := applicableNames(forgejo)
	if fj["github/dependabot"] || fj["github/FUNDING"] {
		t.Fatalf("forgejo run must not have dependabot/FUNDING (FR-006)")
	}
	for name := range fj {
		if strings.HasPrefix(name, "github") || strings.HasPrefix(name, "gitlab") {
			t.Fatalf("forgejo run leaked %s", name)
		}
	}

	gitlab := base
	gitlab.Platform = scaffold.PlatformGitLab
	gl := applicableNames(gitlab)
	if !gl["gitlab-ci"] {
		t.Fatalf("gitlab run missing gitlab-ci")
	}
	for name := range gl {
		if strings.HasPrefix(name, "github") || strings.HasPrefix(name, "forgejo") {
			t.Fatalf("gitlab run leaked %s", name)
		}
	}

	// Unset platform → only the six baseline files, no CI/extras.
	none := applicableNames(base)
	if len(none) != 6 {
		t.Fatalf("unset platform produced %d files, want 6 baseline: %v", len(none), none)
	}
	for name := range none {
		if strings.ContainsAny(name, "/") {
			t.Fatalf("unset platform leaked a platform file: %s", name)
		}
	}
}

// TestForgejoReleaseDisableWiring: only the release descriptor carries a
// DisableIfExists/DisableSuffix pair; lint/test/snapshot are untouched.
func TestForgejoReleaseDisableWiring(t *testing.T) {
	t.Parallel()
	reg := scaffold.NewRegistry()
	for _, tpl := range reg.All() {
		if !strings.HasPrefix(tpl.Name, "forgejo/workflows/") {
			continue
		}
		if tpl.Name == "forgejo/workflows/release" {
			if tpl.DisableIfExists != ".github" || tpl.DisableSuffix != ".disabled" {
				t.Fatalf("release descriptor has wrong disable wiring: %+v", tpl)
			}
			continue
		}
		if tpl.DisableIfExists != "" || tpl.DisableSuffix != "" {
			t.Fatalf("%s must not carry disable wiring: %+v", tpl.Name, tpl)
		}
	}
}

// TestDockerGating: Dockerfile appears only when Docker is on (SC-005).
func TestDockerGating(t *testing.T) {
	t.Parallel()
	p := scaffold.ProjectProfile{ProjectName: "demo", Binary: "demo", MainPath: "./cmd/demo", Platform: scaffold.PlatformGitHub}
	if applicableNames(p)["dockerfile"] {
		t.Fatal("dockerfile present without --docker")
	}
	p.Docker = true
	if !applicableNames(p)["dockerfile"] {
		t.Fatal("dockerfile missing with --docker")
	}
}
