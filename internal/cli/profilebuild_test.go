package cli //nolint:testpackage // white-box unit tests for unexported resolution + form logic (constitution II testability)

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/sgaunet/scaffold/internal/config"
	"github.com/sgaunet/scaffold/internal/scaffold"
)

func strp(s string) *string { return &s }
func boolp(b bool) *bool    { return &b }

// emptyDir is a target with no go.mod/.git so detection contributes nothing,
// isolating the flag/env/config/default tiers.
func emptyDir(t *testing.T) string {
	t.Helper()
	return t.TempDir()
}

func TestBuildProfile_ConfigSuppliesDefaults(t *testing.T) {
	dir := emptyDir(t)
	cfg := config.Config{Platform: strp("github"), Owner: strp("cfgowner")}
	p := buildProfile(profileFlags{dir: dir}, cfg)

	if p.Platform != scaffold.PlatformGitHub {
		t.Errorf("platform: got %q want github", p.Platform)
	}
	if p.Owner != "cfgowner" {
		t.Errorf("owner: got %q want cfgowner", p.Owner)
	}
}

func TestBuildProfile_FlagBeatsConfig(t *testing.T) {
	dir := emptyDir(t)
	cfg := config.Config{Platform: strp("github")}
	p := buildProfile(profileFlags{dir: dir, platform: "gitlab"}, cfg)

	if p.Platform != scaffold.PlatformGitLab {
		t.Errorf("flag must override config: got %q want gitlab", p.Platform)
	}
}

func TestBuildProfile_EnvBeatsConfig(t *testing.T) {
	dir := emptyDir(t)
	t.Setenv("SCAFFOLD_PLATFORM", "forgejo")
	cfg := config.Config{Platform: strp("github")}
	p := buildProfile(profileFlags{dir: dir}, cfg)

	if p.Platform != scaffold.PlatformForgejo {
		t.Errorf("env must override config: got %q want forgejo", p.Platform)
	}
}

func TestBuildProfile_ConfigBeatsDefault_VersionsAndName(t *testing.T) {
	dir := emptyDir(t)
	cfg := config.Config{Name: strp("from-config"), GoVersion: strp("1.27.0")}
	p := buildProfile(profileFlags{dir: dir}, cfg)

	if p.ProjectName != "from-config" {
		t.Errorf("name: got %q want from-config", p.ProjectName)
	}
	if p.GoVersion != "1.27.0" {
		t.Errorf("go version: got %q want 1.27.0", p.GoVersion)
	}
	if p.TaskVersion != scaffold.DefaultTaskVersion {
		t.Errorf("unset version pin should fall back to default, got %q", p.TaskVersion)
	}
}

func TestBuildProfile_UnsetConfigFallsThroughToDirName(t *testing.T) {
	dir := emptyDir(t)
	p := buildProfile(profileFlags{dir: dir}, config.Config{})

	if p.ProjectName != filepath.Base(dir) {
		t.Errorf("name should default to dir base %q, got %q", filepath.Base(dir), p.ProjectName)
	}
}

func TestBuildProfile_ConfigDockerBooleans(t *testing.T) {
	dir := emptyDir(t)

	// docker on with a platform+owner derives a registry default.
	on := buildProfile(profileFlags{dir: dir}, config.Config{
		Docker: boolp(true), Platform: strp("github"), Owner: strp("me"),
	})
	if !on.Docker {
		t.Error("config docker:true should enable docker")
	}
	if on.Registry == "" {
		t.Error("docker on with platform+owner should derive a registry")
	}

	off := buildProfile(profileFlags{dir: dir}, config.Config{Docker: boolp(false)})
	if off.Docker {
		t.Error("config docker:false should keep docker off")
	}
}

func TestBuildProfile_PlatformNoneNormalizes(t *testing.T) {
	dir := emptyDir(t)
	p := buildProfile(profileFlags{dir: dir}, config.Config{Platform: strp("none")})
	if p.Platform != scaffold.PlatformNone {
		t.Errorf("'none' should normalize to baseline (empty), got %q", p.Platform)
	}
}

func TestBuildProfile_MergedValidationCatchesConfigError(t *testing.T) {
	dir := emptyDir(t)
	// homebrew without a github platform is only invalid after the merge.
	cfg := config.Config{Homebrew: boolp(true), Platform: strp("gitlab")}
	p := buildProfile(profileFlags{dir: dir}, cfg)

	err := p.Validate()
	if err == nil || !errors.Is(err, scaffold.ErrUsage) {
		t.Fatalf("expected ErrUsage from merged validation, got %v", err)
	}
}
