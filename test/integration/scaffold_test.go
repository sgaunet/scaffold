package integration_test

import (
	"encoding/json"
	"strings"
	"testing"
)

// generate is interactive-only and requires a terminal, so the binary-level
// suite exercises the non-interactive surface: version and help. The on-disk
// file-set guarantee for generate lives in the internal/scaffold package tests
// (generate_ondisk_test.go), which drive the CLI-free Generate API.

// TestHelpContract: --help lists the commands and the exit-code table.
func TestHelpContract(t *testing.T) {
	r := run(t, t.TempDir(), "--help")
	for _, want := range []string{"generate", "version", "Exit codes:", "10", "usage error"} {
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
