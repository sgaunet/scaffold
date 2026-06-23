package config_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sgaunet/scaffold/internal/config"
)

func writeCfg(t *testing.T, dir, body string) string {
	t.Helper()
	p := filepath.Join(dir, "config.yml")
	if err := os.WriteFile(p, []byte(body), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	return p
}

// --- Path resolution (T004) ---

func TestLoad_DisabledSkipsEverything(t *testing.T) {
	t.Parallel()
	res, err := config.Load(config.Options{Disabled: true, ExplicitPath: "/nope/missing.yml"})
	if err != nil {
		t.Fatalf("disabled load should not error: %v", err)
	}
	if res.Path != "" || res.Config.Platform != nil {
		t.Fatalf("disabled load should be empty, got %+v", res)
	}
}

func TestLoad_DefaultLocationFromXDG(t *testing.T) {
	home := t.TempDir()
	cfgDir := filepath.Join(home, "scaffold")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cfgDir, "config.yml"), []byte("owner: xdguser\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("XDG_CONFIG_HOME", home)

	res, err := config.Load(config.Options{})
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if res.Config.Owner == nil || *res.Config.Owner != "xdguser" {
		t.Fatalf("expected owner=xdguser, got %+v", res.Config.Owner)
	}
}

func TestLoad_DefaultLocationFallsBackToHomeConfig(t *testing.T) {
	home := t.TempDir()
	cfgDir := filepath.Join(home, ".config", "scaffold")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cfgDir, "config.yml"), []byte("owner: homeuser\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("HOME", home)

	res, err := config.Load(config.Options{})
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if res.Config.Owner == nil || *res.Config.Owner != "homeuser" {
		t.Fatalf("expected owner=homeuser, got %+v", res.Config.Owner)
	}
}

func TestLoad_DefaultMissingIsSilent(t *testing.T) {
	home := t.TempDir() // empty: no config.yml
	t.Setenv("XDG_CONFIG_HOME", home)

	res, err := config.Load(config.Options{})
	if err != nil {
		t.Fatalf("missing default config must be silent, got error: %v", err)
	}
	if res.Path != "" {
		t.Fatalf("expected no path, got %q", res.Path)
	}
}

func TestLoad_ExplicitMissingIsError(t *testing.T) {
	t.Parallel()
	_, err := config.Load(config.Options{ExplicitPath: filepath.Join(t.TempDir(), "absent.yml")})
	if err == nil {
		t.Fatal("explicitly requested missing config must error (FR-005)")
	}
	if !strings.Contains(err.Error(), "does not exist") {
		t.Fatalf("error should explain the missing file, got: %v", err)
	}
}

func TestLoad_ExplicitPathLoads(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	p := writeCfg(t, dir, "platform: gitlab\nowner: me\n")

	res, err := config.Load(config.Options{ExplicitPath: p})
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if res.Config.Platform == nil || *res.Config.Platform != "gitlab" {
		t.Fatalf("expected platform=gitlab, got %+v", res.Config.Platform)
	}
}

// --- Decode, unknown keys, errors (T005) ---

func TestLoad_FullValidFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	p := writeCfg(t, dir, strings.Join([]string{
		"name: my-tool",
		"binary: my-tool",
		"platform: github",
		"owner: sgaunet",
		"docker: true",
		"homebrew: false",
		"go-version: \"1.27.0\"",
	}, "\n"))

	res, err := config.Load(config.Options{ExplicitPath: p})
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	c := res.Config
	if c.Name == nil || *c.Name != "my-tool" {
		t.Errorf("name: got %v", c.Name)
	}
	if c.Docker == nil || *c.Docker != true {
		t.Errorf("docker: got %v", c.Docker)
	}
	if c.Homebrew == nil || *c.Homebrew != false {
		t.Errorf("homebrew should be explicitly false, got %v", c.Homebrew)
	}
	if c.GoVersion == nil || *c.GoVersion != "1.27.0" {
		t.Errorf("go-version: got %v", c.GoVersion)
	}
	if len(res.UnknownKeys) != 0 {
		t.Errorf("unexpected unknown keys: %v", res.UnknownKeys)
	}
}

func TestLoad_UnknownKeysWarnNotError(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	p := writeCfg(t, dir, "owner: me\ncolour: blue\nfuture-flag: 1\n")

	res, err := config.Load(config.Options{ExplicitPath: p})
	if err != nil {
		t.Fatalf("unknown keys must not error: %v", err)
	}
	want := []string{"colour", "future-flag"}
	if strings.Join(res.UnknownKeys, ",") != strings.Join(want, ",") {
		t.Fatalf("unknown keys: got %v want %v", res.UnknownKeys, want)
	}
	if res.Config.Owner == nil || *res.Config.Owner != "me" {
		t.Fatalf("recognized keys should still apply, got %+v", res.Config.Owner)
	}
}

func TestLoad_MalformedYAMLIsError(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	p := writeCfg(t, dir, "owner: : : broken\n  - nope\n")

	if _, err := config.Load(config.Options{ExplicitPath: p}); err == nil {
		t.Fatal("malformed YAML must error (FR-006)")
	}
}

func TestLoad_InvalidValueIsError(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	p := writeCfg(t, dir, "platform: githubx\n")

	_, err := config.Load(config.Options{ExplicitPath: p})
	if err == nil {
		t.Fatal("invalid platform must error (FR-007)")
	}
	if !strings.Contains(err.Error(), "platform") {
		t.Fatalf("error must name the offending key, got: %v", err)
	}
}

func TestLoad_WrongTypeIsError(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	p := writeCfg(t, dir, "docker: maybe\n") // not a bool

	if _, err := config.Load(config.Options{ExplicitPath: p}); err == nil {
		t.Fatal("a non-bool docker value must error")
	}
}
