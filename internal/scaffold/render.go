package scaffold

import (
	"bytes"
	"fmt"
	"text/template"
)

// Custom delimiters so literal GoReleaser `{{ }}` and Actions `${{ }}` pass
// through verbatim (research §1).
const (
	leftDelim  = "[["
	rightDelim = "]]"
)

// RenderContext is the flat data passed to every template. It is ProjectProfile
// plus derived booleans/strings; identical for every template in a run so names
// stay consistent across files (FR-019).
type RenderContext struct {
	ProjectName       string
	Binary            string
	ModulePath        string
	Owner             string
	Host              string
	Platform          string
	IsGitHub          bool
	IsGitLab          bool
	IsForgejo         bool
	Docker            bool
	MainPath          string
	Registry          string
	VersionPackage    string
	GoVersion         string
	TaskVersion       string
	GolangciVersion   string
	GoreleaserVersion string
	FundingUser       string
	TokenEnv          string
	Homebrew          bool
	HomebrewTap       string
}

// NewRenderContext derives the render context from a profile.
func NewRenderContext(p ProjectProfile) RenderContext {
	tokenEnv := ""
	if plat, ok := KnownPlatform(p.Platform); ok {
		tokenEnv = plat.ReleaseTokenEnv
	}
	return RenderContext{
		ProjectName:       p.ProjectName,
		Binary:            p.Binary,
		ModulePath:        p.ModulePath,
		Owner:             p.Owner,
		Host:              p.Host,
		Platform:          string(p.Platform),
		IsGitHub:          p.Platform == PlatformGitHub,
		IsGitLab:          p.Platform == PlatformGitLab,
		IsForgejo:         p.Platform == PlatformForgejo,
		Docker:            p.Docker,
		MainPath:          p.MainPath,
		Registry:          p.Registry,
		VersionPackage:    p.VersionPackage,
		GoVersion:         p.GoVersion,
		TaskVersion:       p.TaskVersion,
		GolangciVersion:   p.GolangciVersion,
		GoreleaserVersion: p.GoreleaserVersion,
		FundingUser:       p.FundingUser,
		TokenEnv:          tokenEnv,
		Homebrew:          p.Homebrew,
		HomebrewTap:       p.HomebrewTap,
	}
}

// renderString renders a single template string (used for file bodies and for
// destination paths).
func renderString(name, text string, ctx RenderContext) (string, error) {
	t, err := template.New(name).Delims(leftDelim, rightDelim).Parse(text)
	if err != nil {
		return "", fmt.Errorf("parse template %s: %w", name, err)
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, ctx); err != nil {
		return "", fmt.Errorf("render template %s: %w", name, err)
	}
	return buf.String(), nil
}

// Render reads and renders a template's body from the embedded FS.
func Render(tmpl Template, ctx RenderContext) ([]byte, error) {
	raw, err := templatesFS.ReadFile(tmpl.Source)
	if err != nil {
		return nil, fmt.Errorf("read embedded template %s: %w", tmpl.Source, err)
	}
	out, err := renderString(tmpl.Name, string(raw), ctx)
	if err != nil {
		return nil, err
	}
	return []byte(out), nil
}
