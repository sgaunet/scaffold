package scaffold

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

const dirMode = fs.FileMode(0o755)

// WriteFile atomically writes content to dest: it writes a temp file in the
// destination directory then renames it into place, so an interrupted run never
// leaves a half-written file (spec edge case). Parent dirs are created.
func WriteFile(dest string, content []byte, mode fs.FileMode) error {
	dir := filepath.Dir(dest)
	if err := os.MkdirAll(dir, dirMode); err != nil {
		return fmt.Errorf("create dir %s: %w", dir, err)
	}
	tmp, err := os.CreateTemp(dir, ".scaffold-*.tmp")
	if err != nil {
		return fmt.Errorf("create temp file in %s: %w", dir, err)
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }() // no-op once renamed

	if _, err := tmp.Write(content); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("write temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp file: %w", err)
	}
	if err := os.Chmod(tmpName, mode); err != nil {
		return fmt.Errorf("chmod temp file: %w", err)
	}
	if err := os.Rename(tmpName, dest); err != nil {
		return fmt.Errorf("rename %s to %s: %w", tmpName, dest, err)
	}
	return nil
}
