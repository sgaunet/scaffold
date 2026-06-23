// Package detect reads go.mod and .git/config directly (no external processes,
// FR-009) to provide auto-detected defaults for a generation run.
package detect

import (
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/mod/modfile"
)

// Module holds the info detected from a go.mod file.
type Module struct {
	Path string // full module path, e.g. github.com/sgaunet/semver
	Name string // base of the module path, e.g. semver
}

// FromGoMod parses dir/go.mod. ok is false when the file is absent or has no
// module path (the spec's "no module file present" fallback).
func FromGoMod(dir string) (Module, bool) {
	data, err := os.ReadFile(filepath.Join(dir, "go.mod"))
	if err != nil {
		return Module{}, false
	}
	path := modfile.ModulePath(data)
	if path == "" {
		return Module{}, false
	}
	name := path
	if i := strings.LastIndex(path, "/"); i >= 0 {
		name = path[i+1:]
	}
	return Module{Path: path, Name: name}, true
}
