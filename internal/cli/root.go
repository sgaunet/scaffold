// Package cli is the CLI-thin layer: the only package importing cobra. It
// parses input, calls the scaffold core, formats output, and maps errors to
// exit codes (constitution II/IV).
package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/sgaunet/scaffold/internal/scaffold"
	"github.com/spf13/cobra"
)

// globalOpts holds the flags shared by every command.
type globalOpts struct {
	output  string
	quiet   bool
	verbose bool
	noColor bool
	out     io.Writer
	errw    io.Writer
}

// NewRootCmd builds the command tree, writing data to out and logs to errw.
func NewRootCmd(out, errw io.Writer) *cobra.Command {
	g := &globalOpts{out: out, errw: errw}

	root := &cobra.Command{
		Use:   "scaffold",
		Short: "Generate Go project tooling/CI/config files from embedded templates",
		Long: `scaffold generates a Go project's tooling, CI, and config files from
templates embedded in the binary.

'scaffold generate' runs an interactive, platform-first setup form and requires
a terminal; its prompts are pre-filled from a config file, environment, and
auto-detection. 'scaffold list' previews the file set non-interactively (pipe-
and CI-safe). Existing files are skipped; generate prompts before overwriting.

Exit codes:
  0   success
  1   generic failure
  2   usage error (no terminal for generate, bad flag, invalid name)
  10  conflict (one or more existing files skipped)`,
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
			if _, ok := os.LookupEnv("NO_COLOR"); ok {
				g.noColor = true
			}
			if g.output != outputText && g.output != outputJSON {
				return fmt.Errorf("%w: --output must be %q or %q, got %q",
					scaffold.ErrUsage, outputText, outputJSON, g.output)
			}
			return nil
		},
	}
	root.SetOut(out)
	root.SetErr(errw)

	// Flag-parse failures (unknown/removed flags, missing values) are usage
	// errors: wrap them as ErrUsage so they map to exit code 2, not 1. Inherited
	// by every subcommand via cobra's FlagErrorFunc parent lookup.
	root.SetFlagErrorFunc(func(_ *cobra.Command, err error) error {
		return fmt.Errorf("%w: %w", scaffold.ErrUsage, err)
	})

	root.PersistentFlags().StringVar(&g.output, "output", outputText, "output format for stdout: text|json")
	root.PersistentFlags().BoolVarP(&g.quiet, "quiet", "q", false, "suppress human progress on stderr")
	root.PersistentFlags().BoolVarP(&g.verbose, "verbose", "v", false, "extra diagnostics on stderr")

	root.AddCommand(newGenerateCmd(g))
	root.AddCommand(newListCmd(g))
	root.AddCommand(newVersionCmd(g))
	return root
}

// logf writes a human progress/diagnostic line to stderr unless --quiet.
func (g *globalOpts) logf(format string, args ...any) {
	if g.quiet {
		return
	}
	fmt.Fprintf(g.errw, format+"\n", args...)
}

// vlogf writes an extra-diagnostic line to stderr only under --verbose (and not
// --quiet). Used for low-signal notes like which config file was loaded.
func (g *globalOpts) vlogf(format string, args ...any) {
	if g.quiet || !g.verbose {
		return
	}
	fmt.Fprintf(g.errw, format+"\n", args...)
}
