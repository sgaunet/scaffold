package cli_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestStreamSeparationJSON: with --output json, stdout is valid JSON only and no
// human log leaks into the data stream (constitution IV, Testing Standards:
// stream separation). Driven through list, which is the non-interactive command.
func TestStreamSeparationJSON(t *testing.T) {
	t.Parallel()
	stdout, _, code := runCLI(t, "list", "--no-config", "--output", "json")
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
}

// TestQuietSuppressesStderr: an unknown config key normally warns on stderr;
// --quiet must silence it while leaving the data stream untouched.
func TestQuietSuppressesStderr(t *testing.T) {
	t.Parallel()
	cfg := filepath.Join(t.TempDir(), "config.yml")
	if err := os.WriteFile(cfg, []byte("platform: github\nbogus-key: 1\n"), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	_, stderr, code := runCLI(t, "list", "--config", cfg)
	if code != 0 {
		t.Fatalf("exit = %d, want 0", code)
	}
	if !strings.Contains(stderr, "bogus-key") {
		t.Fatalf("expected unknown-key warning on stderr, got %q", stderr)
	}

	_, qstderr, qcode := runCLI(t, "list", "--config", cfg, "--quiet")
	if qcode != 0 {
		t.Fatalf("quiet exit = %d, want 0", qcode)
	}
	if strings.TrimSpace(qstderr) != "" {
		t.Fatalf("--quiet should suppress stderr, got %q", qstderr)
	}
}
