package cli //nolint:testpackage // white-box unit tests for unexported resolution + form logic (constitution II testability)

import (
	"errors"
	"io"
	"testing"

	"github.com/sgaunet/scaffold/internal/scaffold"
)

// withStubTerminal swaps the package isTerminal probe for the duration of a test.
func withStubTerminal(t *testing.T, result bool) {
	t.Helper()
	orig := isTerminal
	isTerminal = func(uintptr) bool { return result }
	t.Cleanup(func() { isTerminal = orig })
}

func TestRequireTerminal_NoTTYIsUsageError(t *testing.T) {
	withStubTerminal(t, false)
	g := &globalOpts{out: io.Discard, errw: io.Discard}
	err := g.requireTerminal()
	if err == nil || !errors.Is(err, scaffold.ErrUsage) {
		t.Fatalf("non-terminal -i must be a usage error, got %v", err)
	}
}

func TestRequireTerminal_QuietIsUsageError(t *testing.T) {
	withStubTerminal(t, true) // a terminal, but --quiet is set
	g := &globalOpts{out: io.Discard, errw: io.Discard, quiet: true}
	err := g.requireTerminal()
	if err == nil || !errors.Is(err, scaffold.ErrUsage) {
		t.Fatalf("--quiet with -i must be a usage error, got %v", err)
	}
}

func TestRequireTerminal_TTYPasses(t *testing.T) {
	withStubTerminal(t, true)
	g := &globalOpts{out: io.Discard, errw: io.Discard}
	if err := g.requireTerminal(); err != nil {
		t.Fatalf("a real terminal should pass, got %v", err)
	}
}
