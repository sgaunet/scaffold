package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/sgaunet/scaffold/internal/scaffold"
)

const (
	outputText = "text"
	outputJSON = "json"
)

// writeReport renders the report to out in the selected format.
func writeReport(out io.Writer, format string, r scaffold.Report) error {
	if format == outputJSON {
		enc := json.NewEncoder(out)
		enc.SetIndent("", "  ")
		if err := enc.Encode(r); err != nil {
			return fmt.Errorf("encode json report: %w", err)
		}
		return nil
	}
	return writeTextReport(out, r)
}

func writeTextReport(out io.Writer, r scaffold.Report) error {
	tw := tabwriter.NewWriter(out, 0, 2, 2, ' ', 0)
	fmt.Fprintln(tw, "ACTION\tFILE\tTEMPLATE")
	for _, it := range r.Items {
		fmt.Fprintf(tw, "%s\t%s\t%s\n", it.Action, it.Dest, it.Name)
	}
	if err := tw.Flush(); err != nil {
		return fmt.Errorf("flush text report: %w", err)
	}
	s := r.Summary
	fmt.Fprintf(out, "\n%d created, %d skipped, %d overwritten",
		s.Created, s.Skipped, s.Overwritten)
	if r.DryRun {
		fmt.Fprintf(out, " | %d would-create, %d would-skip, %d would-overwrite",
			s.WouldCreate, s.WouldSkip, s.WouldOverwrite)
	}
	fmt.Fprintln(out)
	return nil
}
