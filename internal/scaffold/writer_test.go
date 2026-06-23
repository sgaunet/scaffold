package scaffold_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sgaunet/scaffold/internal/scaffold"
)

// TestWriteFileAtomicCreatesDirs writes into a nested, not-yet-existing path
// and verifies the content and that no temp file remains.
func TestWriteFileAtomicCreatesDirs(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	dest := filepath.Join(dir, "a", "b", "file.txt")
	if err := scaffold.WriteFile(dest, []byte("hello"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	got, err := os.ReadFile(dest)
	if err != nil || string(got) != "hello" {
		t.Fatalf("read back = %q, %v", got, err)
	}
	entries, _ := os.ReadDir(filepath.Dir(dest))
	if len(entries) != 1 {
		t.Fatalf("expected exactly the target file, got %d entries", len(entries))
	}
}

// TestWriteFileNotWritable fails cleanly on a read-only parent directory.
func TestWriteFileNotWritable(t *testing.T) {
	t.Parallel()
	if os.Geteuid() == 0 {
		t.Skip("running as root; permission check would not fail")
	}
	dir := t.TempDir()
	ro := filepath.Join(dir, "ro")
	if err := os.Mkdir(ro, 0o555); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	err := scaffold.WriteFile(filepath.Join(ro, "x.txt"), []byte("data"), 0o644)
	if err == nil {
		t.Fatal("expected error writing into a read-only dir, got nil")
	}
}
