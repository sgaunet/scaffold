package cli_test

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestStreamSeparationJSON: with --output json, stdout is valid JSON only and
// the human log goes to stderr (constitution IV, Testing Standards: stream
// separation).
func TestStreamSeparationJSON(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	stdout, stderr, code := runCLI(t, "generate", "--name", "demo", "-C", dir, "--output", "json")
	if code != 0 {
		t.Fatalf("exit = %d, want 0", code)
	}
	var report map[string]any
	if err := json.Unmarshal([]byte(stdout), &report); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\n%s", err, stdout)
	}
	if strings.Contains(stdout, "scaffold:") {
		t.Fatal("human log leaked into stdout")
	}
	if !strings.Contains(stderr, "created") {
		t.Fatalf("expected progress log on stderr, got %q", stderr)
	}
}

// TestQuietSuppressesStderr: --quiet silences the human progress line.
func TestQuietSuppressesStderr(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	_, stderr, code := runCLI(t, "generate", "--name", "demo", "-C", dir, "--quiet")
	if code != 0 {
		t.Fatalf("exit = %d, want 0", code)
	}
	if strings.TrimSpace(stderr) != "" {
		t.Fatalf("--quiet should suppress stderr, got %q", stderr)
	}
}

// TestDryRunJSONVocabulary: dry-run reports would-* actions and counters.
func TestDryRunJSONVocabulary(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	stdout, _, code := runCLI(t, "generate", "--name", "demo", "--platform", "github", "-C", dir, "--dry-run", "--output", "json")
	if code != 0 {
		t.Fatalf("exit = %d, want 0", code)
	}
	var r struct {
		DryRun  bool                      `json:"dryRun"`
		Items   []struct{ Action string } `json:"items"`
		Summary struct {
			Created, WouldCreate int
		} `json:"summary"`
	}
	if err := json.Unmarshal([]byte(stdout), &r); err != nil {
		t.Fatalf("bad json: %v", err)
	}
	if !r.DryRun || r.Summary.WouldCreate == 0 || r.Summary.Created != 0 {
		t.Fatalf("unexpected dry-run summary: %+v", r.Summary)
	}
	for _, it := range r.Items {
		if it.Action != "would-create" {
			t.Fatalf("dry-run action = %s, want would-create", it.Action)
		}
	}
}
