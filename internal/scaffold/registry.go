package scaffold

// TemplateRegistry is the ordered collection of all template descriptors,
// built once over the embedded FS. Applicable() is used by both generate and
// list, so any newly-registered template appears everywhere automatically
// (FR-015, SC-008).
type TemplateRegistry struct {
	templates []Template
}

// NewRegistry builds the default registry (baseline + platform + docker).
func NewRegistry() *TemplateRegistry {
	r := &TemplateRegistry{}
	r.registerBaseline()
	r.registerPlatform()
	r.registerDocker()
	return r
}

// Register appends a template descriptor.
func (r *TemplateRegistry) Register(t Template) {
	r.templates = append(r.templates, t)
}

// Applicable returns the templates that apply to a profile, in registration order.
func (r *TemplateRegistry) Applicable(p ProjectProfile) []Template {
	out := make([]Template, 0, len(r.templates))
	for _, t := range r.templates {
		if t.Applies == nil || t.Applies(p) {
			out = append(out, t)
		}
	}
	return out
}

// All returns every registered template (used by tests/diagnostics).
func (r *TemplateRegistry) All() []Template {
	return r.templates
}

func (r *TemplateRegistry) registerBaseline() {
	r.Register(Template{Name: "goreleaser", Source: "templates/goreleaser.yaml.tmpl", DestTmpl: ".goreleaser.yaml", Applies: always, Mode: fileMode})
	r.Register(Template{Name: "mise", Source: "templates/mise.toml.tmpl", DestTmpl: "mise.toml", Applies: always, Mode: fileMode})
	r.Register(Template{Name: "golangci", Source: "templates/golangci.yml.tmpl", DestTmpl: ".golangci.yml", Applies: always, Mode: fileMode})
	r.Register(Template{Name: "pre-commit", Source: "templates/pre-commit-config.yaml.tmpl", DestTmpl: ".pre-commit-config.yaml", Applies: always, Mode: fileMode})
	r.Register(Template{Name: "taskfile", Source: "templates/Taskfile.yml.tmpl", DestTmpl: "Taskfile.yml", Applies: always, Mode: fileMode})
	r.Register(Template{Name: "taskfile-dev", Source: "templates/Taskfile_dev.yml.tmpl", DestTmpl: "Taskfile_dev.yml", Applies: always, Mode: fileMode})
	r.Register(Template{Name: "vscode-settings", Source: "templates/vscode/settings.json.tmpl", DestTmpl: ".vscode/settings.json", Applies: always, Mode: fileMode})
	r.Register(Template{Name: "vscode-extensions", Source: "templates/vscode/extensions.json.tmpl", DestTmpl: ".vscode/extensions.json", Applies: always, Mode: fileMode})
}

func (r *TemplateRegistry) registerPlatform() {
	// GitHub workflows + extras.
	for _, wf := range []string{"linter", "test", "snapshot", "release"} {
		r.Register(Template{
			Name:     "github/workflows/" + wf,
			Source:   "templates/github/workflows/" + wf + ".yml.tmpl",
			DestTmpl: ".github/workflows/" + wf + ".yml",
			Applies:  platformIs(PlatformGitHub),
			Mode:     fileMode,
		})
	}
	r.Register(Template{Name: "github/dependabot", Source: "templates/github/dependabot.yml.tmpl", DestTmpl: ".github/dependabot.yml", Applies: platformIs(PlatformGitHub), Mode: fileMode})
	r.Register(Template{Name: "github/FUNDING", Source: "templates/github/FUNDING.yml.tmpl", DestTmpl: ".github/FUNDING.yml", Applies: platformIs(PlatformGitHub), Mode: fileMode})

	// Forgejo workflows. The release workflow is written but neutralized
	// (suffixed, so Forgejo's *.yml/*.yaml glob ignores it) when a .github
	// directory already exists, since GitHub then owns releases and goreleaser
	// can only target one host per run.
	for _, wf := range []string{"lint", "test", "snapshot", "release"} {
		tpl := Template{
			Name:     "forgejo/workflows/" + wf,
			Source:   "templates/forgejo/workflows/" + wf + ".yml.tmpl",
			DestTmpl: ".forgejo/workflows/" + wf + ".yml",
			Applies:  platformIs(PlatformForgejo),
			Mode:     fileMode,
		}
		if wf == "release" {
			tpl.DisableIfExists = ".github"
			tpl.DisableSuffix = ".disabled"
		}
		r.Register(tpl)
	}

	// GitLab single pipeline.
	r.Register(Template{Name: "gitlab-ci", Source: "templates/gitlab/gitlab-ci.yml.tmpl", DestTmpl: ".gitlab-ci.yml", Applies: platformIs(PlatformGitLab), Mode: fileMode})
}

func (r *TemplateRegistry) registerDocker() {
	r.Register(Template{Name: "dockerfile", Source: "templates/Dockerfile.tmpl", DestTmpl: "Dockerfile", Applies: dockerOn, Mode: fileMode})
}
