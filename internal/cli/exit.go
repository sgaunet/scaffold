package cli

import (
	"errors"

	"github.com/sgaunet/scaffold/internal/scaffold"
)

// Documented exit codes (contracts/cli.md, constitution IV).
const (
	ExitOK       = 0
	ExitFailure  = 1
	ExitUsage    = 2
	ExitConflict = 10
)

// isConflict reports whether err is (or wraps) the file-conflict sentinel.
func isConflict(err error) bool {
	return errors.Is(err, scaffold.ErrConflict)
}

// ExitCode maps an error returned by the command tree to a process exit code.
func ExitCode(err error) int {
	switch {
	case err == nil:
		return ExitOK
	case errors.Is(err, scaffold.ErrUsage):
		return ExitUsage
	case errors.Is(err, scaffold.ErrConflict):
		return ExitConflict
	default:
		return ExitFailure
	}
}
