// Package config loads optional scaffold defaults from a YAML config file and
// exposes them as one tier in the resolution precedence
// (flags > env > config > auto-detection > built-in defaults). It is CLI-free
// (constitution II): it imports only the standard library and gopkg.in/yaml.v3,
// never cobra or huh, so it is fully unit-testable without the command layer.
package config

import (
	"fmt"
	"regexp"
	"strings"
)

// nameRe mirrors the project/binary name rule enforced by the scaffold core; it
// is duplicated here (rather than imported) to keep this package independent and
// give config-context error messages. The merged ProjectProfile.Validate remains
// the authoritative cross-field check.
var nameRe = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]*$`)

// validPlatforms are the selectable forges; "" and "none" mean baseline-only.
var validPlatforms = map[string]bool{"github": true, "gitlab": true, "forgejo": true}

// Config is the typed view of config.yml. Every field is optional: a nil pointer
// means "unset" so the value falls through to the next precedence tier. Booleans
// use *bool so an explicit false is distinguishable from absent. Keys mirror the
// `generate` flag names.
type Config struct {
	Name              *string `yaml:"name"`
	Binary            *string `yaml:"binary"`
	Module            *string `yaml:"module"`
	Platform          *string `yaml:"platform"`
	Owner             *string `yaml:"owner"`
	Registry          *string `yaml:"registry"`
	MainPath          *string `yaml:"main"`
	Docker            *bool   `yaml:"docker"`
	Homebrew          *bool   `yaml:"homebrew"`
	HomebrewTap       *string `yaml:"homebrew-tap"`       //nolint:tagliatelle // kebab keys mirror the documented config schema and CLI flag names
	GoVersion         *string `yaml:"go-version"`         //nolint:tagliatelle // kebab keys mirror the documented config schema and CLI flag names
	TaskVersion       *string `yaml:"task-version"`       //nolint:tagliatelle // kebab keys mirror the documented config schema and CLI flag names
	GolangciVersion   *string `yaml:"golangci-version"`   //nolint:tagliatelle // kebab keys mirror the documented config schema and CLI flag names
	GoreleaserVersion *string `yaml:"goreleaser-version"` //nolint:tagliatelle // kebab keys mirror the documented config schema and CLI flag names
}

// KnownKeys is the set of recognized top-level YAML keys. Anything else is an
// unknown key (warned and ignored, FR-008).
var KnownKeys = map[string]bool{
	"name": true, "binary": true, "module": true, "platform": true,
	"owner": true, "registry": true, "main": true, "docker": true,
	"homebrew": true, "homebrew-tap": true, "go-version": true,
	"task-version": true, "golangci-version": true, "goreleaser-version": true,
}

// validate applies per-field shape checks, naming the offending key (FR-007).
// Cross-field rules (e.g. homebrew requires a GitHub platform) are deferred to
// the merged ProjectProfile.Validate, since they depend on values that may come
// from flags/env/detection rather than the file.
func (c Config) validate() error {
	for key, val := range map[string]*string{
		"name": c.Name, "binary": c.Binary, "homebrew-tap": c.HomebrewTap,
	} {
		if val != nil && !nameRe.MatchString(*val) {
			return fmt.Errorf("invalid %s %q: must match %s", key, *val, nameRe.String())
		}
	}
	if c.Platform != nil {
		p := *c.Platform
		if p != "" && p != "none" && !validPlatforms[p] {
			return fmt.Errorf("invalid platform %q: must be one of github, gitlab, forgejo (or none)", p)
		}
	}
	if c.MainPath != nil && *c.MainPath != "" && !strings.HasPrefix(*c.MainPath, "./") {
		return fmt.Errorf("invalid main %q: must be a relative path beginning './'", *c.MainPath)
	}
	return nil
}
