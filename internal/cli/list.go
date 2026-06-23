package cli

import (
	"github.com/sgaunet/scaffold/internal/scaffold"
	"github.com/spf13/cobra"
)

func newListCmd(g *globalOpts) *cobra.Command {
	var (
		pf profileFlags
		cf configFlags
	)
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List the files that would be generated for the given options",
		Long: `List the files that would be generated, applying the same input precedence as
generate: flags > env > config file > auto-detection > built-in defaults.`,
		Args: cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			cfg, err := g.resolveConfig(cf)
			if err != nil {
				return err
			}
			profile := buildProfile(pf, cfg)
			reg := scaffold.NewRegistry()
			report, err := scaffold.List(profile, reg)
			if err != nil {
				return err
			}
			return writeReport(g.out, g.output, report)
		},
	}
	addProfileFlags(&pf, cmd.Flags())
	addConfigFlags(&cf, cmd.Flags())
	return cmd
}
