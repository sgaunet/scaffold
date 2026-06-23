package integration_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestBaseline (T035): generate with no platform → six baseline files, no CI,
// no residual delimiters (US1).
func TestBaseline(t *testing.T) {
	dir := t.TempDir()
	r := run(t, dir, "generate", "--name", "demo")
	if r.code != 0 {
		t.Fatalf("exit=%d stderr=%s", r.code, r.stderr)
	}
	for _, f := range []string{".goreleaser.yaml", "mise.toml", ".golangci.yml", ".pre-commit-config.yaml", "Taskfile.yml", "Taskfile_dev.yml"} {
		mustExist(t, dir, f)
	}
	mustNotExist(t, dir, ".github")
	mustNotExist(t, dir, ".gitlab-ci.yml")
	assertNoDelimiters(t, dir)
}

// TestPlatformMatrix (T044): each platform's files + GitHub-only extras.
func TestPlatformMatrix(t *testing.T) {
	t.Run("github", func(t *testing.T) {
		dir := t.TempDir()
		run(t, dir, "generate", "--name", "demo", "--platform", "github")
		for _, f := range []string{".github/workflows/linter.yml", ".github/workflows/test.yml", ".github/workflows/snapshot.yml", ".github/workflows/release.yml", ".github/dependabot.yml", ".github/FUNDING.yml"} {
			mustExist(t, dir, f)
		}
	})
	t.Run("forgejo", func(t *testing.T) {
		dir := t.TempDir()
		run(t, dir, "generate", "--name", "demo", "--platform", "forgejo")
		mustExist(t, dir, ".forgejo/workflows/release.yml")
		mustNotExist(t, dir, ".github/dependabot.yml")
		mustNotExist(t, dir, ".github/FUNDING.yml")
	})
	t.Run("gitlab", func(t *testing.T) {
		dir := t.TempDir()
		run(t, dir, "generate", "--name", "demo", "--platform", "gitlab")
		mustExist(t, dir, ".gitlab-ci.yml")
		mustNotExist(t, dir, ".github")
		mustNotExist(t, dir, ".forgejo")
	})
}

// TestDockerToggle (T051): container artifacts appear only with --docker.
func TestDockerToggle(t *testing.T) {
	on := t.TempDir()
	run(t, on, "generate", "--name", "demo", "--platform", "github", "--docker", "--owner", "acme")
	mustExist(t, on, "Dockerfile")
	if !grepFile(t, filepath.Join(on, ".goreleaser.yaml"), "dockers:") {
		t.Fatal("expected dockers: block with --docker")
	}

	off := t.TempDir()
	run(t, off, "generate", "--name", "demo", "--platform", "github")
	mustNotExist(t, off, "Dockerfile")
	if grepFile(t, filepath.Join(off, ".goreleaser.yaml"), "dockers:") {
		t.Fatal("dockers: block must be absent without --docker")
	}
}

// TestRerunExit10 (T057): re-run skips all → exit 10; --force → exit 0.
func TestRerunExit10(t *testing.T) {
	dir := t.TempDir()
	if r := run(t, dir, "generate", "--name", "demo", "--platform", "github"); r.code != 0 {
		t.Fatalf("first run exit=%d", r.code)
	}
	if r := run(t, dir, "generate", "--name", "demo", "--platform", "github"); r.code != 10 {
		t.Fatalf("re-run exit=%d, want 10", r.code)
	}
	if r := run(t, dir, "generate", "--name", "demo", "--platform", "github", "--force"); r.code != 0 {
		t.Fatalf("force exit=%d, want 0", r.code)
	}
}

// TestDryRunStreams (T058): dry-run writes nothing; JSON on stdout, logs on stderr.
func TestDryRunStreams(t *testing.T) {
	dir := t.TempDir()
	r := run(t, dir, "generate", "--name", "demo", "--platform", "github", "--dry-run", "--output", "json")
	if r.code != 0 {
		t.Fatalf("exit=%d", r.code)
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) != 0 {
		t.Fatalf("dry-run wrote %d entries", len(entries))
	}
	var report map[string]any
	if err := json.Unmarshal([]byte(r.stdout), &report); err != nil {
		t.Fatalf("stdout not JSON: %v", err)
	}
	if strings.Contains(r.stdout, "created,") {
		t.Fatal("human summary leaked to stdout")
	}
}

// TestJSONInvariant (T056): generate + list JSON honor the summary invariant
// and required keys.
func TestJSONInvariant(t *testing.T) {
	check := func(t *testing.T, raw string) {
		t.Helper()
		var d struct {
			Project string `json:"project"`
			Items   []any  `json:"items"`
			Summary struct {
				Created, Skipped, Overwritten, WouldCreate, WouldSkip, WouldOverwrite int
			} `json:"summary"`
		}
		if err := json.Unmarshal([]byte(raw), &d); err != nil {
			t.Fatalf("bad json: %v", err)
		}
		s := d.Summary
		if len(d.Items) != s.Created+s.Skipped+s.Overwritten+s.WouldCreate+s.WouldSkip+s.WouldOverwrite {
			t.Fatalf("invariant broken for %s", raw)
		}
	}
	dir := t.TempDir()
	gen := run(t, dir, "generate", "--name", "demo", "--platform", "gitlab", "--docker", "--owner", "acme", "--output", "json")
	check(t, gen.stdout)
	lst := run(t, t.TempDir(), "list", "--name", "demo", "--platform", "forgejo", "--output", "json")
	check(t, lst.stdout)
}

// TestHelpContract (T071): --help lists the commands and the exit-code table.
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

func assertNoDelimiters(t *testing.T, dir string) {
	t.Helper()
	_ = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		b, _ := os.ReadFile(path)
		if bytes.Contains(b, []byte("[[")) || bytes.Contains(b, []byte("]]")) {
			t.Fatalf("residual delimiters in %s", path)
		}
		return nil
	})
}

func grepFile(t *testing.T, path, substr string) bool {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return bytes.Contains(b, []byte(substr))
}
