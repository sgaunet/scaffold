// Package scaffold is the CLI-free core: it owns the template registry, render
// context, generation plan, atomic writer, and result report. It imports no CLI
// packages (constitution II) so it is fully unit-testable on its own.
package scaffold

import (
	"context"
	"fmt"
	"path/filepath"
)

// Options controls a generation run.
type Options struct {
	Dir    string
	Force  bool
	DryRun bool
}

// Generate computes the plan, renders and writes the applicable files, and
// returns a Report. It honors context cancellation between files. If any file
// was skipped because it already exists, it returns the Report together with a
// wrapped ErrConflict (the non-conflicting files are still written).
func Generate(ctx context.Context, reg *TemplateRegistry, p ProjectProfile, opts Options) (Report, error) {
	if err := p.Validate(); err != nil {
		return Report{}, err
	}
	rctx := NewRenderContext(p)

	mode := modeGenerate
	if opts.DryRun {
		mode = modeDryRun
	}
	plan, err := BuildPlan(reg, p, rctx, opts.Dir, opts.Force, mode)
	if err != nil {
		return Report{}, err
	}

	items := make([]ReportItem, 0, len(plan.Items))
	for _, pi := range plan.Items {
		if err := ctx.Err(); err != nil {
			return Report{}, fmt.Errorf("generation cancelled: %w", err)
		}
		if !opts.DryRun && (pi.Action == ActionCreate || pi.Action == ActionOverwrite) {
			content, rerr := Render(pi.Template, rctx)
			if rerr != nil {
				return Report{}, rerr
			}
			if werr := WriteFile(filepath.Join(opts.Dir, pi.Dest), content, pi.Template.Mode); werr != nil {
				return Report{}, werr
			}
		}
		items = append(items, ReportItem{Name: pi.Template.Name, Dest: pi.Dest, Action: pi.Action, Reason: pi.Reason})
	}

	report := newReport(p, opts.DryRun, items)
	if plan.HasConflicts() {
		return report, fmt.Errorf("%w: one or more files already exist; re-run with --force to overwrite", ErrConflict)
	}
	return report, nil
}

// List returns the report of templates that would be produced for a profile,
// writing nothing (FR-016). Every item is would-create.
func List(p ProjectProfile, reg *TemplateRegistry) (Report, error) {
	if err := p.Validate(); err != nil {
		return Report{}, err
	}
	rctx := NewRenderContext(p)
	plan, err := BuildPlan(reg, p, rctx, "", false, modeList)
	if err != nil {
		return Report{}, err
	}
	items := make([]ReportItem, 0, len(plan.Items))
	for _, pi := range plan.Items {
		items = append(items, ReportItem{Name: pi.Template.Name, Dest: pi.Dest, Action: pi.Action})
	}
	return newReport(p, true, items), nil
}
