package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"gopkg.in/yaml.v3"
)

// Options controls a load.
type Options struct {
	// ExplicitPath is a user-requested config path (from --config or
	// SCAFFOLD_CONFIG). When set, a missing file is an error (FR-005).
	ExplicitPath string
	// Disabled skips config loading entirely (--no-config).
	Disabled bool
}

// Result is the outcome of a successful load.
type Result struct {
	Config      Config   // parsed defaults (zero value when no file was read)
	Path        string   // absolute path actually read; "" when none
	UnknownKeys []string // unrecognized top-level keys (warn + ignore, FR-008)
}

// Load resolves the config path, reads and decodes it. A missing file at the
// default/derived location yields an empty Result and a nil error (FR-002). A
// missing file at an explicitly requested path (FR-005), a parse failure
// (FR-006), or an invalid value (FR-007) returns a non-nil error; the CLI layer
// maps that to a usage error (exit 2). The file is only ever read (FR-022).
func Load(opts Options) (Result, error) {
	if opts.Disabled {
		return Result{}, nil
	}

	path, explicit := resolvePath(opts.ExplicitPath)
	if path == "" {
		return Result{}, nil // no HOME/XDG and no explicit path: nothing to load
	}

	data, err := os.ReadFile(path) //nolint:gosec // path is user-chosen by design
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			if explicit {
				return Result{}, fmt.Errorf("config file %q does not exist", path)
			}
			return Result{}, nil // default location absent: silent (FR-002)
		}
		return Result{}, fmt.Errorf("read config %q: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Result{}, fmt.Errorf("parse config %q: %w", path, err)
	}

	// Second pass into a generic map to surface unknown top-level keys without
	// failing the load (forward compatibility, FR-008).
	var raw map[string]any
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return Result{}, fmt.Errorf("parse config %q: %w", path, err)
	}

	if err := cfg.validate(); err != nil {
		return Result{}, fmt.Errorf("config %q: %w", path, err)
	}

	return Result{Config: cfg, Path: path, UnknownKeys: unknownKeys(raw)}, nil
}

// resolvePath returns the config path and whether it was explicitly requested.
// Order: explicit path > $XDG_CONFIG_HOME/scaffold/config.yml >
// $HOME/.config/scaffold/config.yml. os.UserConfigDir is intentionally avoided
// (it points at ~/Library/Application Support on macOS), so the documented
// ~/.config path holds on every platform.
func resolvePath(explicit string) (path string, isExplicit bool) {
	if explicit != "" {
		if abs, err := filepath.Abs(explicit); err == nil {
			return abs, true
		}
		return explicit, true
	}
	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil || home == "" {
			return "", false
		}
		base = filepath.Join(home, ".config")
	}
	return filepath.Join(base, "scaffold", "config.yml"), false
}

// unknownKeys returns the sorted top-level keys not recognized by the schema.
func unknownKeys(raw map[string]any) []string {
	var unknown []string
	for k := range raw {
		if !KnownKeys[k] {
			unknown = append(unknown, k)
		}
	}
	sort.Strings(unknown)
	return unknown
}
