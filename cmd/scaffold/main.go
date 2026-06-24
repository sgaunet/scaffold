// Command scaffold generates a Go project's tooling/CI/config files from
// templates embedded in the binary. This entry point only wires a cancellable
// context, runs the CLI, prints the final error to stderr, and maps it to an
// exit code.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/sgaunet/scaffold/internal/cli"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	root := cli.NewRootCmd(os.Stdout, os.Stderr)
	err := root.ExecuteContext(ctx)
	// The command tree silences cobra's own error printing (SilenceErrors) so the
	// final, wrapped message is surfaced here once — on stderr, never stdout, so
	// the data stream stays clean (constitution III/IV). A clean cancel returns
	// nil and prints nothing.
	if err != nil {
		fmt.Fprintf(os.Stderr, "scaffold: %v\n", err)
	}
	os.Exit(cli.ExitCode(err))
}
