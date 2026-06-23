package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

// Build metadata, injected via GoReleaser ldflags (-X .../internal/cli.Version=…).
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

func newVersionCmd(g *globalOpts) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version, commit, and build date",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			if g.output == outputJSON {
				enc := json.NewEncoder(g.out)
				enc.SetIndent("", "  ")
				return enc.Encode(map[string]string{
					"version": Version, "commit": Commit, "date": Date,
				})
			}
			fmt.Fprintf(g.out, "scaffold %s (commit %s, built %s)\n", Version, Commit, Date)
			return nil
		},
	}
}
