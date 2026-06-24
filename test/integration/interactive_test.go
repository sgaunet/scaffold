package integration_test

import (
	"strings"
	"testing"
)

// generate is interactive-only and the form requires a real terminal. In
// CI/pipe contexts stdin and stderr are not TTYs, so generate must fail loudly
// (exit 2) rather than hang, and must write nothing to stdout. The form's
// internal logic (field model, seeding, parity) is covered by unit tests in
// internal/cli.

func TestGenerate_NoTTYIsUsageError(t *testing.T) {
	work := t.TempDir()
	r := run(t, work, "generate")
	if r.code != 2 {
		t.Fatalf("`generate` without a terminal should exit 2, got %d (stderr=%s)", r.code, r.stderr)
	}
	if r.stdout != "" {
		t.Errorf("no data should be written to stdout, got: %q", r.stdout)
	}
	// The actionable message must reach the user on stderr (constitution III), not
	// just an exit code.
	if !strings.Contains(r.stderr, "terminal") {
		t.Errorf("expected an actionable 'requires a terminal' message on stderr, got: %q", r.stderr)
	}
	mustNotExist(t, work, ".goreleaser.yaml")
}

func TestGenerate_QuietIsUsageError(t *testing.T) {
	work := t.TempDir()
	r := run(t, work, "generate", "--quiet")
	if r.code != 2 {
		t.Fatalf("`generate --quiet` should exit 2, got %d (stderr=%s)", r.code, r.stderr)
	}
}

// TestGenerate_RemovedFlagIsUsageError: the profile flags were removed; passing
// one is a usage error (exit 2) at the binary level, not a generic failure.
func TestGenerate_RemovedFlagIsUsageError(t *testing.T) {
	work := t.TempDir()
	r := run(t, work, "generate", "--platform", "github")
	if r.code != 2 {
		t.Fatalf("removed --platform flag should exit 2, got %d (stderr=%s)", r.code, r.stderr)
	}
}
