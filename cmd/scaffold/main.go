// Command scaffold generates a Go project's tooling/CI/config files from
// templates embedded in the binary. This entry point only wires a cancellable
// context, runs the CLI, and maps the resulting error to an exit code.
package main

import (
	"context"
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
	os.Exit(cli.ExitCode(err))
}
