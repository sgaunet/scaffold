package scaffold

import (
	"os"
	"path/filepath"
)

// Action is the decision for a single file.
type Action string

// Actions: past-tense for a real generate run; would-* for dry-run.
const (
	ActionCreate         Action = "created"
	ActionSkip           Action = "skipped"
	ActionOverwrite      Action = "overwritten"
	ActionWouldCreate    Action = "would-create"
	ActionWouldSkip      Action = "would-skip"
	ActionWouldOverwrite Action = "would-overwrite"
)

const skipReason = "exists, no --force"

// PlanItem is one (template → destination, action) decision.
type PlanItem struct {
	Template Template
	Dest     string
	Action   Action
	Reason   string
}

// GenerationPlan is the computed, reviewable list of decisions before any write.
type GenerationPlan struct {
	Items []PlanItem
}

// HasConflicts reports whether any file was (or would be) skipped because it
// already exists — drives the conflict exit code 10 (FR-013).
func (g GenerationPlan) HasConflicts() bool {
	for _, it := range g.Items {
		if it.Action == ActionSkip || it.Action == ActionWouldSkip {
			return true
		}
	}
	return false
}

// planMode selects how actions are labelled.
type planMode int

const (
	modeGenerate planMode = iota // real write decisions
	modeDryRun                   // would-* reflecting on-disk state
)

// BuildPlan computes the plan for the applicable templates.
func BuildPlan(reg *TemplateRegistry, p ProjectProfile, ctx RenderContext, dir string, force bool, mode planMode) (GenerationPlan, error) {
	var plan GenerationPlan
	for _, t := range reg.Applicable(p) {
		dest, err := renderString(t.Name+":dest", t.DestTmpl, ctx)
		if err != nil {
			return GenerationPlan{}, err
		}
		if t.DisableIfExists != "" && fileExists(filepath.Join(dir, t.DisableIfExists)) {
			dest += t.DisableSuffix
		}
		item := PlanItem{Template: t, Dest: dest}

		exists := fileExists(filepath.Join(dir, dest))
		dry := mode == modeDryRun
		switch {
		case !exists:
			item.Action = pick(dry, ActionWouldCreate, ActionCreate)
		case force:
			item.Action = pick(dry, ActionWouldOverwrite, ActionOverwrite)
		default:
			item.Action = pick(dry, ActionWouldSkip, ActionSkip)
			item.Reason = skipReason
		}
		plan.Items = append(plan.Items, item)
	}
	return plan, nil
}

func pick(dry bool, dryAction, realAction Action) Action {
	if dry {
		return dryAction
	}
	return realAction
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
