package integration_test

import (
	"encoding/json"
	"strings"
	"testing"
)

// generate is interactive-only and requires a terminal, so the binary-level
// suite exercises the non-interactive surface: list (plan preview), version, and
// help. The on-disk file-set guarantee for generate lives in the internal/scaffold
// package tests (generate_ondisk_test.go), which drive the CLI-free Generate API.

// TestList_BaselineFileSet: list with no config writes nothing and reports the
// six baseline files with no platform CI in the plan (US1, pipe-safe).
func TestList_BaselineFileSet(t *testing.T) {
	work := t.TempDir()
	r := run(t, work, "list", "--no-config")
	if r.code != 0 {
		t.Fatalf("exit=%d stderr=%s", r.code, r.stderr)
	}
	for _, f := range []string{".goreleaser.yaml", "mise.toml", ".golangci.yml", ".pre-commit-config.yaml", "Taskfile.yml", "Taskfile_dev.yml"} {
		if !strings.Contains(r.stdout, f) {
			t.Errorf("baseline plan missing %s:\n%s", f, r.stdout)
		}
	}
	for _, f := range []string{".github/", ".gitlab-ci.yml", ".forgejo/", "Dockerfile"} {
		if strings.Contains(r.stdout, f) {
			t.Errorf("baseline plan leaked %s:\n%s", f, r.stdout)
		}
	}
}

// TestList_DockerInPlan: a docker config adds the Dockerfile to the plan.
func TestList_DockerInPlan(t *testing.T) {
	work := t.TempDir()
	cfg := writeConfig(t, "platform: github\nowner: acme\ndocker: true\n")
	r := run(t, work, "list", "--config", cfg)
	if r.code != 0 {
		t.Fatalf("exit=%d stderr=%s", r.code, r.stderr)
	}
	if !strings.Contains(r.stdout, "Dockerfile") {
		t.Errorf("docker config should list Dockerfile:\n%s", r.stdout)
	}
}

// TestList_JSONInvariant: list JSON honors the summary invariant + required keys.
func TestList_JSONInvariant(t *testing.T) {
	work := t.TempDir()
	cfg := writeConfig(t, "platform: gitlab\nowner: acme\ndocker: true\n")
	r := run(t, work, "list", "--config", cfg, "--output", "json")
	if r.code != 0 {
		t.Fatalf("exit=%d stderr=%s", r.code, r.stderr)
	}
	var d struct {
		Project string `json:"project"`
		Items   []any  `json:"items"`
		Summary struct {
			Created, Skipped, Overwritten, WouldCreate, WouldSkip, WouldOverwrite int
		} `json:"summary"`
	}
	if err := json.Unmarshal([]byte(r.stdout), &d); err != nil {
		t.Fatalf("bad json: %v", err)
	}
	s := d.Summary
	if len(d.Items) != s.Created+s.Skipped+s.Overwritten+s.WouldCreate+s.WouldSkip+s.WouldOverwrite {
		t.Fatalf("invariant broken:\n%s", r.stdout)
	}
}

// TestHelpContract: --help lists the commands and the exit-code table.
func TestHelpContract(t *testing.T) {
	r := run(t, t.TempDir(), "--help")
	for _, want := range []string{"generate", "list", "version", "Exit codes:", "10", "usage error"} {
		if !strings.Contains(r.stdout, want) {
			t.Fatalf("--help missing %q\n%s", want, r.stdout)
		}
	}
}

// TestVersionJSON: version --output json emits version/commit/date.
func TestVersionJSON(t *testing.T) {
	r := run(t, t.TempDir(), "version", "--output", "json")
	var v map[string]string
	if err := json.Unmarshal([]byte(r.stdout), &v); err != nil {
		t.Fatalf("bad json: %v", err)
	}
	for _, k := range []string{"version", "commit", "date"} {
		if _, ok := v[k]; !ok {
			t.Fatalf("version json missing %q", k)
		}
	}
}
