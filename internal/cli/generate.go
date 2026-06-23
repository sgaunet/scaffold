package cli

import (
	"github.com/sgaunet/scaffold/internal/scaffold"
	"github.com/spf13/cobra"
)

func newGenerateCmd(g *globalOpts) *cobra.Command {
	var (
		pf     profileFlags
		force  bool
		dryRun bool
		yes    bool
	)
	cmd := &cobra.Command{
		Use:     "generate",
		Aliases: []string{"gen"},
		Short:   "Generate tooling/CI/config files for a Go project",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			profile := buildProfile(pf)
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
	cmd.Flags().BoolVarP(&force, "force", "f", false, "overwrite existing files")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "compute and print the plan; write nothing")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "assume yes to interactive overwrite prompts")
	return cmd
}

func orDot(dir string) string {
	if dir == "" {
		return "."
	}
	return dir
}
