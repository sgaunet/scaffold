package cli

import (
	"github.com/sgaunet/scaffold/internal/scaffold"
	"github.com/spf13/cobra"
)

func newGenerateCmd(g *globalOpts) *cobra.Command {
	var (
		pf          profileFlags
		cf          configFlags
		force       bool
		dryRun      bool
		yes         bool
		interactive bool
	)
	cmd := &cobra.Command{
		Use:     "generate",
		Aliases: []string{"gen"},
		Short:   "Generate tooling/CI/config files for a Go project",
		Long: `Generate a Go project's tooling, CI, and config files from embedded templates.

Inputs are resolved with precedence:
  flags > env > config file > auto-detection (go.mod / git remote) > built-in defaults

Config file: $XDG_CONFIG_HOME/scaffold/config.yml (else $HOME/.config/scaffold/config.yml),
overridable with --config / $SCAFFOLD_CONFIG, or disabled with --no-config.

Use --interactive/-i for a guided, platform-first setup form (requires a terminal).`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := g.resolveConfig(cf)
			if err != nil {
				return err
			}

			if interactive {
				if terr := g.requireTerminal(); terr != nil {
					return terr
				}
				return g.runInteractive(cmd.Context(), pf, cfg, yes)
			}

			profile := buildProfile(pf, cfg)
			reg := scaffold.NewRegistry()

			report, genErr := scaffold.Generate(cmd.Context(), reg, profile, scaffold.Options{
				Dir:    orDot(pf.dir),
				Force:  force,
				DryRun: dryRun,
			})
			// A conflict still yields a valid report (non-conflicting files were
			// written); only a hard failure produces no report.
			if genErr == nil || isConflict(genErr) {
				if err := writeReport(g.out, g.output, report); err != nil {
					return err
				}
				g.logf("scaffold: %d created, %d skipped, %d overwritten",
					report.Summary.Created, report.Summary.Skipped, report.Summary.Overwritten)
			}
			return genErr
		},
	}

	addProfileFlags(&pf, cmd.Flags())
	addConfigFlags(&cf, cmd.Flags())
	cmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "guided setup form (platform-first; requires a terminal)")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "overwrite existing files")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "compute and print the plan; write nothing")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "assume yes to the interactive overwrite prompt")
	return cmd
}

func orDot(dir string) string {
	if dir == "" {
		return "."
	}
	return dir
}
