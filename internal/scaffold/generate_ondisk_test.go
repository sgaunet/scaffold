package scaffold_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sgaunet/scaffold/internal/scaffold"
)

// These tests cover what the CLI integration suite used to assert by driving the
// binary: that a real Generate writes the correct *dest paths* to disk for each
// platform/docker/homebrew combination. generate is now interactive-only and
// cannot be driven headlessly, so the end-to-end file-set guarantee lives here,
// at the CLI-free core (constitution II). Predicate selection, rendered content,
// and the create/skip/overwrite lifecycle are covered separately by
// registry_test, render_test, and plan_test.

// ondiskProfile builds a fully-formed, validated profile (as the CLI produces
// after the merge) and applies the optional mutation.
func ondiskProfile(t *testing.T, mutate func(*scaffold.ProjectProfile)) scaffold.ProjectProfile {
	t.Helper()
	p := scaffold.ProjectProfile{
		ProjectName: "demo", Binary: "demo", ModulePath: "github.com/acme/demo",
		Owner: "acme", Host: "github.com", Platform: scaffold.PlatformGitHub,
		MainPath: "./cmd/demo", VersionPackage: "github.com/acme/demo/internal/cli",
		GoVersion: "1.26.1", TaskVersion: "3.40.1",
		GolangciVersion: "2.12.1", GoreleaserVersion: "2.16.0",
		FundingUser: "acme", HomebrewTap: "homebrew-tools",
	}
	if mutate != nil {
		mutate(&p)
	}
	if err := p.Validate(); err != nil {
		t.Fatalf("invalid test profile: %v", err)
	}
	return p
}

func generateOnDisk(t *testing.T, p scaffold.ProjectProfile) string {
	t.Helper()
	dir := t.TempDir()
	if _, err := scaffold.Generate(context.Background(), scaffold.NewRegistry(), p, scaffold.Options{Dir: dir}); err != nil {
		t.Fatalf("generate: %v", err)
	}
	return dir
}

func onDisk(dir, rel string) bool {
	_, err := os.Stat(filepath.Join(dir, rel))
	return err == nil
}

func fileHas(t *testing.T, path, sub string) bool {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return strings.Contains(string(b), sub)
}

// TestGenerateWritesBaselineFileSet: a baseline run writes the six tool files and
// no platform/container artifacts (US1).
func TestGenerateWritesBaselineFileSet(t *testing.T) {
	t.Parallel()
	dir := generateOnDisk(t, ondiskProfile(t, func(p *scaffold.ProjectProfile) {
		p.Platform = scaffold.PlatformNone
	}))
	for _, f := range []string{".goreleaser.yaml", "mise.toml", ".golangci.yml", ".pre-commit-config.yaml", "Taskfile.yml", "Taskfile_dev.yml"} {
		if !onDisk(dir, f) {
			t.Errorf("baseline file %s missing", f)
		}
	}
	for _, f := range []string{".github", ".gitlab-ci.yml", ".forgejo", "Dockerfile"} {
		if onDisk(dir, f) {
			t.Errorf("baseline run must not create %s", f)
		}
	}
}

// TestGenerateWritesPlatformFileSet: each platform writes its own dest paths and
// none of the others' (SC-004, FR-006).
func TestGenerateWritesPlatformFileSet(t *testing.T) {
	t.Parallel()
	cases := []struct {
		platform     scaffold.PlatformID
		want, absent []string
	}{
		{scaffold.PlatformGitHub,
			[]string{".github/workflows/linter.yml", ".github/workflows/test.yml", ".github/workflows/snapshot.yml", ".github/workflows/release.yml", ".github/dependabot.yml", ".github/FUNDING.yml"},
			[]string{".gitlab-ci.yml", ".forgejo"}},
		{scaffold.PlatformForgejo,
			[]string{".forgejo/workflows/release.yml"},
			[]string{".github/dependabot.yml", ".github/FUNDING.yml", ".gitlab-ci.yml"}},
		{scaffold.PlatformGitLab,
			[]string{".gitlab-ci.yml"},
			[]string{".github", ".forgejo"}},
	}
	for _, c := range cases {
		t.Run(string(c.platform), func(t *testing.T) {
			t.Parallel()
			plat := c.platform
			dir := generateOnDisk(t, ondiskProfile(t, func(p *scaffold.ProjectProfile) { p.Platform = plat }))
			for _, f := range c.want {
				if !onDisk(dir, f) {
					t.Errorf("%s run missing %s", plat, f)
				}
			}
			for _, f := range c.absent {
				if onDisk(dir, f) {
					t.Errorf("%s run leaked %s", plat, f)
				}
			}
		})
	}
}

// TestGenerateDockerWritesDockerfile: the Dockerfile and goreleaser dockers:
// block appear on disk only with docker enabled (SC-005).
func TestGenerateDockerWritesDockerfile(t *testing.T) {
	t.Parallel()
	on := generateOnDisk(t, ondiskProfile(t, func(p *scaffold.ProjectProfile) {
		p.Docker = true
		p.Registry = "ghcr.io/acme/demo"
	}))
	if !onDisk(on, "Dockerfile") {
		t.Error("Dockerfile missing with docker on")
	}
	if !fileHas(t, filepath.Join(on, ".goreleaser.yaml"), "dockers:") {
		t.Error("expected dockers: block with docker on")
	}

	off := generateOnDisk(t, ondiskProfile(t, nil))
	if onDisk(off, "Dockerfile") {
		t.Error("Dockerfile present with docker off")
	}
	if fileHas(t, filepath.Join(off, ".goreleaser.yaml"), "dockers:") {
		t.Error("dockers: block present with docker off")
	}
}

// TestGenerateHomebrewMarkers: the homebrew_casks block, tap name, and the
// release workflow's tap token appear on disk only with homebrew enabled.
func TestGenerateHomebrewMarkers(t *testing.T) {
	t.Parallel()
	on := generateOnDisk(t, ondiskProfile(t, func(p *scaffold.ProjectProfile) {
		p.Homebrew = true
		p.HomebrewTap = "homebrew-tools"
	}))
	grl := filepath.Join(on, ".goreleaser.yaml")
	if !fileHas(t, grl, "homebrew_casks:") || !fileHas(t, grl, "name: homebrew-tools") {
		t.Error("expected homebrew_casks block with default tap name")
	}
	if !fileHas(t, filepath.Join(on, ".github/workflows/release.yml"), "HOMEBREW_TAP_TOKEN") {
		t.Error("expected HOMEBREW_TAP_TOKEN in release workflow")
	}

	custom := generateOnDisk(t, ondiskProfile(t, func(p *scaffold.ProjectProfile) {
		p.Homebrew = true
		p.HomebrewTap = "homebrew-custom"
	}))
	if !fileHas(t, filepath.Join(custom, ".goreleaser.yaml"), "name: homebrew-custom") {
		t.Error("expected custom tap name homebrew-custom")
	}
}
