package cli_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/sgaunet/scaffold/internal/cli"
	"github.com/sgaunet/scaffold/internal/scaffold"
)

// TestExitCode covers the documented error→exit-code mapping (constitution IV,
// Testing Standards: exit-code selection).
func TestExitCode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		err  error
		want int
	}{
		{"nil → 0", nil, 0},
		{"generic → 1", errors.New("boom"), 1},
		{"usage → 2", fmt.Errorf("bad: %w", scaffold.ErrUsage), 2},
		{"conflict → 10", fmt.Errorf("exists: %w", scaffold.ErrConflict), 10},
		{"wrapped generic → 1", fmt.Errorf("wrap: %w", errors.New("x")), 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := cli.ExitCode(tt.err); got != tt.want {
				t.Fatalf("ExitCode(%v) = %d, want %d", tt.err, got, tt.want)
			}
		})
	}
}
