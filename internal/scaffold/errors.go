package scaffold

import "errors"

// Sentinel errors let the CLI layer map failures to documented exit codes
// (see contracts/cli.md) without importing CLI packages here (constitution II).
var (
	// ErrUsage marks a user/input error → exit code 2.
	ErrUsage = errors.New("usage error")
	// ErrConflict marks a run where one or more files were skipped → exit code 10.
	ErrConflict = errors.New("file conflict")
)
