// Package integration_test builds the real binary and exercises it end-to-end
// (constitution Testing Standards: integration tests invoke the built binary).
package integration_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

var binPath string

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "scaffold-bin-*")
	if err != nil {
		panic(err)
	}
	binPath = filepath.Join(dir, "scaffold")
	build := exec.CommandContext(context.Background(), "go", "build", "-trimpath", "-o", binPath, "../../cmd/scaffold")
	build.Env = append(os.Environ(), "CGO_ENABLED=0")
	if out, berr := build.CombinedOutput(); berr != nil {
		fmt.Fprintf(os.Stderr, "build failed: %v\n%s", berr, out)
		os.Exit(1)
	}
	code := m.Run()
	_ = os.RemoveAll(dir)
	os.Exit(code)
}

type result struct {
	stdout string
	stderr string
	code   int
}

// run executes the built binary in dir with args, capturing streams + exit code.
func run(t *testing.T, dir string, args ...string) result {
	t.Helper()
	var out, errb bytes.Buffer
	cmd := exec.CommandContext(t.Context(), binPath, args...)
	cmd.Dir = dir
	cmd.Stdout = &out
	cmd.Stderr = &errb
	// Run in a clean dir without inherited SCAFFOLD_* env.
	cmd.Env = cleanEnv()
	err := cmd.Run()
	code := 0
	if err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			code = ee.ExitCode()
		} else {
			t.Fatalf("run %v: %v", args, err)
		}
	}
	return result{out.String(), errb.String(), code}
}

func cleanEnv() []string {
	var env []string
	for _, kv := range os.Environ() {
		if len(kv) >= 9 && kv[:9] == "SCAFFOLD_" {
			continue
		}
		env = append(env, kv)
	}
	return env
}

func mustExist(t *testing.T, dir, rel string) {
	t.Helper()
	if _, err := os.Stat(filepath.Join(dir, rel)); err != nil {
		t.Fatalf("expected %s to exist: %v", rel, err)
	}
}

func mustNotExist(t *testing.T, dir, rel string) {
	t.Helper()
	if _, err := os.Stat(filepath.Join(dir, rel)); err == nil {
		t.Fatalf("expected %s NOT to exist", rel)
	}
}
