package cli

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/charmbracelet/huh"
	"github.com/sgaunet/scaffold/internal/config"
	"github.com/sgaunet/scaffold/internal/scaffold"
)

// answers holds the user-editable inputs collected by the interactive form. It
// is a plain struct (no huh types) so the field model is unit-testable without a
// terminal.
type answers struct {
	platform string // github | gitlab | forgejo | none
	name     string
	binary   string
	owner    string
	registry string
	tap      string
	docker   bool
	homebrew bool
}

// seedAnswers derives the form's initial values from an already-resolved profile
// (config > detection > defaults), so the user accepts or edits rather than
// re-typing (FR-017). Platform "" (baseline) is shown as "none".
func seedAnswers(seed scaffold.ProjectProfile) answers {
	plat := string(seed.Platform)
	if plat == "" {
		plat = "none"
	}
	return answers{
		platform: plat,
		name:     seed.ProjectName,
		binary:   seed.Binary,
		owner:    seed.Owner,
		registry: seed.Registry,
		tap:      seed.HomebrewTap,
		docker:   seed.Docker,
		homebrew: seed.Homebrew,
	}
}

// toProfile folds the collected answers back onto the seed profile, preserving
// the non-form fields (module, host, version pins, version package). Derived
// values are recomputed so the result matches a non-interactive run (SC-007).
func (a answers) toProfile(seed scaffold.ProjectProfile) scaffold.ProjectProfile {
	p := seed
	if a.platform == "none" {
		p.Platform = scaffold.PlatformNone
	} else {
		p.Platform = scaffold.PlatformID(a.platform)
	}
	p.ProjectName = a.name
	p.Binary = a.binary
	if p.Binary == "" {
		p.Binary = a.name
	}
	p.Owner = a.owner
	p.FundingUser = a.owner
	p.Docker = a.docker
	p.Homebrew = a.homebrew
	p.HomebrewTap = a.tap
	if a.docker {
		if a.registry != "" {
			p.Registry = a.registry
		} else {
			p.Registry = scaffold.RegistryDefault(p.Platform, p.Host, p.Owner, p.Binary)
		}
	} else {
		p.Registry = ""
	}
	return p
}

// Conditional-visibility predicates (FR-012). Pure functions over current
// answers, unit-tested independently of huh.
func ownerVisible(a *answers) bool    { return a.platform != "none" }
func registryVisible(a *answers) bool { return a.docker }
func homebrewVisible(a *answers) bool { return a.platform == "github" }
func tapVisible(a *answers) bool      { return a.platform == "github" && a.homebrew }

// runInteractive drives the guided setup: collect answers (platform-first,
// conditional), show the dry-run plan, gate overwrite on conflicts, then write
// through the normal Generate path. Cancellation writes nothing (FR-015).
func (g *globalOpts) runInteractive(ctx context.Context, pf profileFlags, cfg config.Config, assumeYes bool) error {
	seed := buildProfile(pf, cfg)
	a := seedAnswers(seed)

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Platform").
				Description("Target forge (or none for the baseline tooling only)").
				Options(
					huh.NewOption("GitHub", "github"),
					huh.NewOption("GitLab", "gitlab"),
					huh.NewOption("Forgejo", "forgejo"),
					huh.NewOption("None (baseline only)", "none"),
				).
				Value(&a.platform),
			huh.NewInput().Title("Project name").Value(&a.name),
			huh.NewInput().Title("Binary name").Value(&a.binary),
		),
		huh.NewGroup(
			huh.NewInput().Title("Repository owner").Value(&a.owner),
		).WithHideFunc(func() bool { return !ownerVisible(&a) }),
		huh.NewGroup(
			huh.NewConfirm().Title("Enable container support?").Value(&a.docker),
		),
		huh.NewGroup(
			huh.NewInput().Title("Image registry base").Value(&a.registry),
		).WithHideFunc(func() bool { return !registryVisible(&a) }),
		huh.NewGroup(
			huh.NewConfirm().Title("Publish a Homebrew formula on release?").Value(&a.homebrew),
		).WithHideFunc(func() bool { return !homebrewVisible(&a) }),
		huh.NewGroup(
			huh.NewInput().Title("Homebrew tap repo").Value(&a.tap),
		).WithHideFunc(func() bool { return !tapVisible(&a) }),
	).WithOutput(g.errw).WithAccessible(os.Getenv("ACCESSIBLE") != "")

	if err := form.RunWithContext(ctx); err != nil {
		return g.handleFormErr(err)
	}

	profile := a.toProfile(seed)
	if err := profile.Validate(); err != nil {
		return err
	}

	return g.generateInteractive(ctx, profile, orDot(pf.dir), assumeYes)
}

// generateInteractive computes the plan, reviews it on stderr, gates overwrite on
// conflicts, then writes through the shared Generate path (FR-013/FR-016).
func (g *globalOpts) generateInteractive(ctx context.Context, profile scaffold.ProjectProfile, dir string, assumeYes bool) error {
	reg := scaffold.NewRegistry()

	preview, perr := scaffold.Generate(ctx, reg, profile, scaffold.Options{Dir: dir, DryRun: true})
	if perr != nil && !isConflict(perr) {
		return perr
	}
	g.logf("\nPlan:")
	for _, it := range preview.Items {
		g.logf("  %-14s %s", it.Action, it.Dest)
	}

	force := assumeYes
	if preview.Summary.WouldSkip > 0 && !assumeYes {
		overwrite := false // default No (skip-existing); FR-016
		confirm := huh.NewForm(huh.NewGroup(
			huh.NewConfirm().
				Title(fmt.Sprintf("%d file(s) already exist. Overwrite them?", preview.Summary.WouldSkip)).
				Description("No keeps existing files (skip); Yes overwrites them.").
				Value(&overwrite),
		)).WithOutput(g.errw).WithAccessible(os.Getenv("ACCESSIBLE") != "")
		if err := confirm.RunWithContext(ctx); err != nil {
			return g.handleFormErr(err)
		}
		force = overwrite
	}

	report, genErr := scaffold.Generate(ctx, reg, profile, scaffold.Options{Dir: dir, Force: force})
	if genErr == nil || isConflict(genErr) {
		if err := writeReport(g.out, g.output, report); err != nil {
			return err
		}
		g.logf("scaffold: %d created, %d skipped, %d overwritten",
			report.Summary.Created, report.Summary.Skipped, report.Summary.Overwritten)
	}
	return genErr
}

// handleFormErr turns a cancelled form into a clean, no-write abort (exit 0,
// FR-015) and surfaces any other failure.
func (g *globalOpts) handleFormErr(err error) error {
	if errors.Is(err, huh.ErrUserAborted) || errors.Is(err, context.Canceled) {
		g.logf("scaffold: cancelled; nothing written")
		return nil
	}
	return fmt.Errorf("interactive form: %w", err)
}
