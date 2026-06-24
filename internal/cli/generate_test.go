package cli_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/sgaunet/scaffold/internal/cli"
)

// runCLI builds the command tree, runs it with args, and returns stdout/stderr
// and the mapped exit code (shared by the cli unit tests). The go test process
// has no controlling terminal, so interactive generate is gated off here.
func runCLI(t *testing.T, args ...string) (stdout, stderr string, code int) {
	t.Helper()
	var out, errb bytes.Buffer
	root := cli.NewRootCmd(&out, &errb)
	root.SetOut(&out)
	root.SetErr(&errb)
	root.SetArgs(args)
	err := root.ExecuteContext(context.Background())
	return out.String(), errb.String(), cli.ExitCode(err)
}

// TestGenerateRequiresTerminal: generate is interactive-only; with no TTY (as in
// `go test`) it must fail loudly with a usage error and write no data to stdout
// (FR-014, constitution IV) rather than hang waiting for input.
func TestGenerateRequiresTerminal(t *testing.T) {
	t.Parallel()
	stdout, stderr, code := runCLI(t, "generate")
	if code != 2 {
		t.Fatalf("exit = %d, want 2 (no terminal)\nstderr=%s", code, stderr)
	}
	if stdout != "" {
		t.Errorf("generate must not write data to stdout without a terminal, got %q", stdout)
	}
}

// TestGenerateRejectsRemovedFlags: the profile and mode flags were removed when
// generate became interactive-only; passing any of them is a usage error (exit
// 2), wrapped from cobra's flag-parse error by SetFlagErrorFunc.
func TestGenerateRejectsRemovedFlags(t *testing.T) {
	t.Parallel()
	for _, f := range []string{"--name", "--platform", "--docker", "--homebrew", "--force", "--dry-run", "-i", "-C"} {
		_, _, code := runCLI(t, "generate", f, "x")
		if code != 2 {
			t.Errorf("generate %s should be a usage error (exit 2), got %d", f, code)
		}
	}
}

// TestListRejectsRemovedFlags: list lost the same profile flags; it is now
// driven purely by config + detection.
func TestListRejectsRemovedFlags(t *testing.T) {
	t.Parallel()
	for _, f := range []string{"--name", "--platform", "--docker"} {
		_, _, code := runCLI(t, "list", f, "x")
		if code != 2 {
			t.Errorf("list %s should be a usage error (exit 2), got %d", f, code)
		}
	}
}

// TestInvalidOutputFlag → usage error, exit 2 (validated in PersistentPreRunE,
// independent of any subcommand's TTY needs).
func TestInvalidOutputFlag(t *testing.T) {
	t.Parallel()
	_, _, code := runCLI(t, "list", "--output", "yaml")
	if code != 2 {
		t.Fatalf("exit = %d, want 2", code)
	}
}
