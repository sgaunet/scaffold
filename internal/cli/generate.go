package cli

import (
	"github.com/spf13/cobra"
)

func newGenerateCmd(g *globalOpts) *cobra.Command {
	var cf configFlags
	cmd := &cobra.Command{
		Use:     "generate",
		Aliases: []string{"gen"},
		Short:   "Generate tooling/CI/config files for a Go project (interactive)",
		Long: `Generate a Go project's tooling, CI, and config files from embedded templates.

generate runs a single guided, platform-first setup form, so it requires a
terminal. Each prompt is pre-filled from the resolved defaults:
  env > config file > auto-detection (go.mod / git remote) > built-in defaults

Edit the config file to change the values proposed in the form. Config file:
$XDG_CONFIG_HOME/scaffold/config.yml (else $HOME/.config/scaffold/config.yml),
overridable with --config / $SCAFFOLD_CONFIG, or disabled with --no-config.

With no terminal (piped, --quiet, or non-interactive) generate exits with a
usage error (exit 2); set the defaults in the config file and run it in an
interactive shell.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := g.resolveConfig(cf)
			if err != nil {
				return err
			}
			if terr := g.requireTerminal(); terr != nil {
				return terr
			}
			return g.runInteractive(cmd.Context(), cfg)
		},
	}

	addConfigFlags(&cf, cmd.Flags())
	return cmd
}
