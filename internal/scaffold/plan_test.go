package scaffold_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/sgaunet/scaffold/internal/scaffold"
)

func genProfile() scaffold.ProjectProfile {
	return scaffold.ProjectProfile{
		ProjectName: "demo", Binary: "demo", ModulePath: "github.com/acme/demo",
		MainPath: "./cmd/demo", Platform: scaffold.PlatformGitHub,
		GoVersion: "1.26.1",
	}
}

// TestPlanDecisions exercises create/skip/overwrite + HasConflicts (via the
// ErrConflict return) through the exported Generate API (black-box).
func TestPlanDecisions(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	reg := scaffold.NewRegistry()
	p := genProfile()

	// First run: everything created, no conflict.
	rep, err := scaffold.Generate(context.Background(), reg, p, scaffold.Options{Dir: dir})
	if err != nil {
		t.Fatalf("first generate: %v", err)
	}
	if rep.Summary.Created == 0 || rep.Summary.Skipped != 0 {
		t.Fatalf("first run summary = %+v, want all created", rep.Summary)
	}
	for _, it := range rep.Items {
		if it.Action != scaffold.ActionCreate {
			t.Fatalf("item %s action = %s, want created", it.Name, it.Action)
		}
	}

	// Second run: all skipped → ErrConflict (exit 10).
	rep2, err := scaffold.Generate(context.Background(), reg, p, scaffold.Options{Dir: dir})
	if !errors.Is(err, scaffold.ErrConflict) {
		t.Fatalf("re-run error = %v, want ErrConflict", err)
	}
	if rep2.Summary.Created != 0 || rep2.Summary.Skipped == 0 {
		t.Fatalf("re-run summary = %+v, want all skipped", rep2.Summary)
	}

	// Third run with force: all overwritten, no conflict.
	rep3, err := scaffold.Generate(context.Background(), reg, p, scaffold.Options{Dir: dir, Force: true})
	if err != nil {
		t.Fatalf("force generate: %v", err)
	}
	if rep3.Summary.Overwritten == 0 {
		t.Fatalf("force summary = %+v, want overwritten", rep3.Summary)
	}
}

// TestDryRunWritesNothing verifies dry-run uses would-* and touches no disk.
func TestDryRunWritesNothing(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	reg := scaffold.NewRegistry()
	rep, err := scaffold.Generate(context.Background(), reg, genProfile(), scaffold.Options{Dir: dir, DryRun: true})
	if err != nil {
		t.Fatalf("dry-run: %v", err)
	}
	if !rep.DryRun || rep.Summary.WouldCreate == 0 || rep.Summary.Created != 0 {
		t.Fatalf("dry-run summary = %+v", rep.Summary)
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) != 0 {
		t.Fatalf("dry-run wrote %d entries, want 0", len(entries))
	}
	// Invariant: items == sum of all six counters.
	s := rep.Summary
	if total := s.Created + s.Skipped + s.Overwritten + s.WouldCreate + s.WouldSkip + s.WouldOverwrite; total != len(rep.Items) {
		t.Fatalf("invariant broken: items=%d sum=%d", len(rep.Items), total)
	}
}

// TestAtomicWriteLeavesNoTemp ensures no stray temp files remain after writes.
func TestAtomicWriteLeavesNoTemp(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	reg := scaffold.NewRegistry()
	if _, err := scaffold.Generate(context.Background(), reg, genProfile(), scaffold.Options{Dir: dir}); err != nil {
		t.Fatalf("generate: %v", err)
	}
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if filepath.Ext(e.Name()) == ".tmp" {
			t.Fatalf("stray temp file left behind: %s", e.Name())
		}
	}
}
