package integration_test

import (
	"testing"
)

// The interactive form requires a real terminal; in CI/pipe contexts stdin and
// stderr are not TTYs, so `-i` must fail loudly (exit 2) rather than hang, and
// must not pollute stdout. The form's internal logic is covered by unit tests
// in internal/cli (field model, seeding, parity).

func TestInteractive_NoTTYIsUsageError(t *testing.T) {
	work := t.TempDir()
	r := run(t, work, "generate", "-i")
	if r.code != 2 {
		t.Fatalf("`generate -i` without a terminal should exit 2, got %d (stderr=%s)", r.code, r.stderr)
	}
	if r.stdout != "" {
		t.Errorf("no data should be written to stdout, got: %q", r.stdout)
	}
	mustNotExist(t, work, ".goreleaser.yaml")
}

func TestInteractive_QuietIsUsageError(t *testing.T) {
	work := t.TempDir()
	r := run(t, work, "generate", "-i", "--quiet")
	if r.code != 2 {
		t.Fatalf("`generate -i --quiet` should exit 2, got %d (stderr=%s)", r.code, r.stderr)
	}
}
