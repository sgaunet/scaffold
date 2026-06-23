package cli //nolint:testpackage // white-box unit tests for unexported resolution + form logic (constitution II testability)

import (
	"testing"

	"github.com/sgaunet/scaffold/internal/config"
	"github.com/sgaunet/scaffold/internal/scaffold"
)

// --- Conditional field visibility (T014, FR-011/FR-012) ---

func TestFieldVisibility(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name                           string
		a                              answers
		owner, registry, homebrew, tap bool
	}{
		{"baseline none", answers{platform: "none"}, false, false, false, false},
		{"gitlab no docker", answers{platform: "gitlab"}, true, false, false, false},
		{"github docker, brew question shown", answers{platform: "github", docker: true}, true, true, true, false},
		{"github docker brew enabled", answers{platform: "github", docker: true, homebrew: true}, true, true, true, true},
		{"forgejo hides brew even if set", answers{platform: "forgejo", homebrew: true}, true, false, false, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			a := c.a
			if got := ownerVisible(&a); got != c.owner {
				t.Errorf("ownerVisible: got %v want %v", got, c.owner)
			}
			if got := registryVisible(&a); got != c.registry {
				t.Errorf("registryVisible: got %v want %v", got, c.registry)
			}
			if got := homebrewVisible(&a); got != c.homebrew {
				t.Errorf("homebrewVisible: got %v want %v", got, c.homebrew)
			}
			if got := tapVisible(&a); got != c.tap {
				t.Errorf("tapVisible: got %v want %v", got, c.tap)
			}
		})
	}
}

func TestPlatformIsFirstAndIncludesNone(t *testing.T) {
	t.Parallel()
	// Seeding a baseline profile must present platform as "none" (selectable),
	// confirming the baseline option exists (FR-011).
	seed := buildProfile(profileFlags{dir: t.TempDir()}, config.Config{})
	a := seedAnswers(seed)
	if a.platform != "none" {
		t.Fatalf("baseline seed platform should be 'none', got %q", a.platform)
	}
}

// --- Pre-fill seeding from resolved config (T023, FR-017) ---

func TestSeedAnswersMatchResolvedProfile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	cfg := config.Config{
		Platform: strp("github"), Owner: strp("seedowner"),
		Name: strp("seedtool"), Docker: boolp(true),
	}
	seed := buildProfile(profileFlags{dir: dir}, cfg)
	a := seedAnswers(seed)

	if a.platform != "github" {
		t.Errorf("platform seed: got %q", a.platform)
	}
	if a.owner != "seedowner" {
		t.Errorf("owner seed: got %q", a.owner)
	}
	if a.name != "seedtool" {
		t.Errorf("name seed: got %q", a.name)
	}
	if !a.docker {
		t.Errorf("docker seed should be true")
	}
	if a.registry != seed.Registry {
		t.Errorf("registry seed: got %q want %q", a.registry, seed.Registry)
	}
}

// --- Parity: accept-all reproduces the non-interactive profile (T025, SC-007) ---

func TestAcceptAllDefaultsEqualsNonInteractive(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	configs := []config.Config{
		{},
		{Platform: strp("github"), Owner: strp("me"), Docker: boolp(true)},
		{Platform: strp("github"), Owner: strp("me"), Homebrew: boolp(true), HomebrewTap: strp("taps")},
		{Platform: strp("gitlab"), Owner: strp("grp"), Docker: boolp(true), Registry: strp("reg.example/grp/app")},
		{Platform: strp("none"), Name: strp("libonly")},
	}
	for i, cfg := range configs {
		t.Run(string(rune('A'+i)), func(t *testing.T) {
			t.Parallel()
			seed := buildProfile(profileFlags{dir: dir}, cfg)
			got := seedAnswers(seed).toProfile(seed)
			if got != seed {
				t.Errorf("accept-all profile diverged from non-interactive:\n got=%+v\nwant=%+v", got, seed)
			}
		})
	}
}

// Guard: a baseline ProjectProfile keeps a stable zero registry round-trip.
func TestToProfileBaselineHasNoRegistry(t *testing.T) {
	t.Parallel()
	seed := scaffold.ProjectProfile{ProjectName: "x", Binary: "x", MainPath: "./cmd/x"}
	got := seedAnswers(seed).toProfile(seed)
	if got.Registry != "" {
		t.Errorf("baseline (no docker) should have empty registry, got %q", got.Registry)
	}
}
