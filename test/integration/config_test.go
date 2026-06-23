package integration_test

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// runEnv is like run but appends extra environment entries (after cleanEnv), so
// tests can set SCAFFOLD_CONFIG / XDG_CONFIG_HOME deterministically.
func runEnv(t *testing.T, dir string, extraEnv []string, args ...string) result {
	t.Helper()
	var out, errb bytes.Buffer
	cmd := exec.CommandContext(t.Context(), binPath, args...)
	cmd.Dir = dir
	cmd.Stdout = &out
	cmd.Stderr = &errb
	cmd.Env = append(cleanEnv(), extraEnv...)
	code := 0
	if err := cmd.Run(); err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			code = ee.ExitCode()
		} else {
			t.Fatalf("run %v: %v", args, err)
		}
	}
	return result{out.String(), errb.String(), code}
}

func writeConfig(t *testing.T, body string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "config.yml")
	if err := os.WriteFile(p, []byte(body), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	return p
}

func TestConfig_ValuesAppliedWithZeroFlags(t *testing.T) {
	work := t.TempDir()
	cfg := writeConfig(t, "platform: github\nowner: cfgowner\n")

	r := run(t, work, "list", "--config", cfg)
	if r.code != 0 {
		t.Fatalf("exit=%d stderr=%s", r.code, r.stderr)
	}
	if !strings.Contains(r.stdout, ".github/dependabot.yml") {
		t.Errorf("config platform=github not applied; stdout:\n%s", r.stdout)
	}
}

func TestConfig_FlagOverridesConfig(t *testing.T) {
	work := t.TempDir()
	cfg := writeConfig(t, "platform: github\n")

	r := run(t, work, "list", "--config", cfg, "--platform", "gitlab")
	if r.code != 0 {
		t.Fatalf("exit=%d stderr=%s", r.code, r.stderr)
	}
	if !strings.Contains(r.stdout, ".gitlab-ci.yml") {
		t.Errorf("flag --platform gitlab should win; stdout:\n%s", r.stdout)
	}
	if strings.Contains(r.stdout, ".github/dependabot.yml") {
		t.Errorf("config github should have been overridden; stdout:\n%s", r.stdout)
	}
}

func TestConfig_ExplicitMissingPathIsUsageError(t *testing.T) {
	work := t.TempDir()
	r := run(t, work, "list", "--config", filepath.Join(t.TempDir(), "absent.yml"))
	if r.code != 2 {
		t.Fatalf("explicit missing config should exit 2, got %d (stderr=%s)", r.code, r.stderr)
	}
}

func TestConfig_NoConfigIgnoresFile(t *testing.T) {
	work := t.TempDir()
	cfg := writeConfig(t, "platform: github\n")

	r := run(t, work, "list", "--config", cfg, "--no-config")
	if r.code != 0 {
		t.Fatalf("exit=%d stderr=%s", r.code, r.stderr)
	}
	if strings.Contains(r.stdout, ".github/dependabot.yml") || strings.Contains(r.stdout, ".gitlab-ci.yml") {
		t.Errorf("--no-config should ignore the file (baseline only); stdout:\n%s", r.stdout)
	}
	if !strings.Contains(r.stdout, ".goreleaser.yaml") {
		t.Errorf("baseline files should still be listed; stdout:\n%s", r.stdout)
	}
}

func TestConfig_MalformedIsUsageErrorAndWritesNothing(t *testing.T) {
	work := t.TempDir()
	cfg := writeConfig(t, "platform: : broken\n  - nope\n")

	r := run(t, work, "generate", "--config", cfg)
	if r.code != 2 {
		t.Fatalf("malformed config should exit 2, got %d (stderr=%s)", r.code, r.stderr)
	}
	mustNotExist(t, work, ".goreleaser.yaml")
}

func TestConfig_FromEnvVar(t *testing.T) {
	work := t.TempDir()
	cfg := writeConfig(t, "platform: forgejo\n")

	r := runEnv(t, work, []string{"SCAFFOLD_CONFIG=" + cfg}, "list")
	if r.code != 0 {
		t.Fatalf("exit=%d stderr=%s", r.code, r.stderr)
	}
	if !strings.Contains(r.stdout, ".forgejo/workflows") {
		t.Errorf("SCAFFOLD_CONFIG platform=forgejo not applied; stdout:\n%s", r.stdout)
	}
}

func TestConfig_UnknownKeyWarnsOnStderrNotStdout(t *testing.T) {
	work := t.TempDir()
	cfg := writeConfig(t, "platform: github\nbogus-key: 1\n")

	r := run(t, work, "list", "--config", cfg)
	if r.code != 0 {
		t.Fatalf("unknown key must not fail; exit=%d stderr=%s", r.code, r.stderr)
	}
	if !strings.Contains(r.stderr, "bogus-key") {
		t.Errorf("expected unknown-key warning on stderr, got:\n%s", r.stderr)
	}
	if strings.Contains(r.stdout, "bogus-key") {
		t.Errorf("warning leaked into stdout (data stream):\n%s", r.stdout)
	}
}
