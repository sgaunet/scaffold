package cli

import (
	"github.com/sgaunet/scaffold/internal/scaffold"
	"github.com/spf13/cobra"
)

func newListCmd(g *globalOpts) *cobra.Command {
	var pf profileFlags
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List the files that would be generated for the given options",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			profile := buildProfile(pf)
			reg := scaffold.NewRegistry()
			report, err := scaffold.List(profile, reg)
			if err != nil {
				return err
			}
			return writeReport(g.out, g.output, report)
		},
	}
	addProfileFlags(&pf, cmd.Flags())
	return cmd
}
