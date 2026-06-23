package scaffold

// ReportItem is one file's result, serialized for --output (see
// contracts/output.schema.json).
type ReportItem struct {
	Name   string `json:"name"`
	Dest   string `json:"dest"`
	Action Action `json:"action"`
	Reason string `json:"reason,omitempty"`
}

// ReportSummary holds the per-action counters. The invariant is
// len(items) == created+skipped+overwritten+wouldCreate+wouldSkip+wouldOverwrite.
type ReportSummary struct {
	Created        int `json:"created"`
	Skipped        int `json:"skipped"`
	Overwritten    int `json:"overwritten"`
	WouldCreate    int `json:"wouldCreate"`
	WouldSkip      int `json:"wouldSkip"`
	WouldOverwrite int `json:"wouldOverwrite"`
}

// Report is the machine-readable result of a run.
type Report struct {
	Project  string        `json:"project"`
	Platform string        `json:"platform"`
	Docker   bool          `json:"docker"`
	DryRun   bool          `json:"dryRun"`
	Items    []ReportItem  `json:"items"`
	Summary  ReportSummary `json:"summary"`
}

// newReport assembles a Report and tallies the summary counters.
func newReport(p ProjectProfile, dryRun bool, items []ReportItem) Report {
	r := Report{
		Project:  p.ProjectName,
		Platform: string(p.Platform),
		Docker:   p.Docker,
		DryRun:   dryRun,
		Items:    items,
	}
	for _, it := range items {
		switch it.Action {
		case ActionCreate:
			r.Summary.Created++
		case ActionSkip:
			r.Summary.Skipped++
		case ActionOverwrite:
			r.Summary.Overwritten++
		case ActionWouldCreate:
			r.Summary.WouldCreate++
		case ActionWouldSkip:
			r.Summary.WouldSkip++
		case ActionWouldOverwrite:
			r.Summary.WouldOverwrite++
		}
	}
	return r
}
