package cli_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/sgaunet/scaffold/internal/cli"
)

// runCLI builds the command tree, runs it with args, and returns stdout/stderr
// and the mapped exit code (shared by the cli unit tests).
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

// TestGeneratePlatformOptional: no --platform → baseline only, exit 0
// (FR-003, US1). Argument parsing + default resolution (constitution Testing
// Standards: argument parsing).
func TestGeneratePlatformOptional(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	_, _, code := runCLI(t, "generate", "--name", "demo", "-C", dir)
	if code != 0 {
		t.Fatalf("exit = %d, want 0", code)
	}
	for _, f := range []string{".goreleaser.yaml", "mise.toml", "Taskfile.yml"} {
		if _, err := os.Stat(filepath.Join(dir, f)); err != nil {
			t.Fatalf("missing baseline file %s", f)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, ".github")); err == nil {
		t.Fatal("no-platform run must not create .github/")
	}
}

// TestGenerateInvalidPlatform → usage error, exit 2.
func TestGenerateInvalidPlatform(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	_, _, code := runCLI(t, "generate", "--name", "demo", "--platform", "bitbucket", "-C", dir)
	if code != 2 {
		t.Fatalf("exit = %d, want 2", code)
	}
}

// TestPlatformPrecedenceEnv: env SCAFFOLD_PLATFORM is used when the flag is
// absent (flags > env > detected).
func TestPlatformPrecedenceEnv(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("SCAFFOLD_PLATFORM", "github")
	_, _, code := runCLI(t, "generate", "--name", "demo", "-C", dir)
	if code != 0 {
		t.Fatalf("exit = %d, want 0", code)
	}
	if _, err := os.Stat(filepath.Join(dir, ".github", "workflows", "release.yml")); err != nil {
		t.Fatalf("env platform not honored: %v", err)
	}
}

// TestInvalidOutputFlag → usage error, exit 2.
func TestInvalidOutputFlag(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	_, _, code := runCLI(t, "generate", "--name", "demo", "-C", dir, "--output", "yaml")
	if code != 2 {
		t.Fatalf("exit = %d, want 2", code)
	}
}
