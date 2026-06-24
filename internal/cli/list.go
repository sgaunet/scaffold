package cli

import (
	"github.com/sgaunet/scaffold/internal/scaffold"
	"github.com/spf13/cobra"
)

func newListCmd(g *globalOpts) *cobra.Command {
	var cf configFlags
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List the files that would be generated",
		Long: `List the files that would be generated, applying the same input precedence as
generate: env > config file > auto-detection > built-in defaults.

list is read-only and never prompts, so it is the way to preview the file set
in a pipe or CI. Use the config file (--config / $SCAFFOLD_CONFIG) to choose the
platform and options.`,
		Args: cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			cfg, err := g.resolveConfig(cf)
			if err != nil {
				return err
			}
			profile := buildProfile(".", cfg)
			reg := scaffold.NewRegistry()
			report, err := scaffold.List(profile, reg)
			if err != nil {
				return err
			}
			return writeReport(g.out, g.output, report)
		},
	}
	addConfigFlags(&cf, cmd.Flags())
	return cmd
}
