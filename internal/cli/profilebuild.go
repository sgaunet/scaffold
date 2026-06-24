package cli

import (
	"os"
	"path/filepath"

	"github.com/sgaunet/scaffold/internal/config"
	"github.com/sgaunet/scaffold/internal/detect"
	"github.com/sgaunet/scaffold/internal/scaffold"
)

// strOr returns *p when set, else fallback. Models the "unset" config tier.
func strOr(p *string, fallback string) string {
	if p != nil {
		return *p
	}
	return fallback
}

// boolOr returns *p when set, else fallback.
func boolOr(p *bool, fallback bool) bool {
	if p != nil {
		return *p
	}
	return fallback
}

// buildProfile resolves a ProjectProfile from dir applying precedence
// env > config > detected > built-in default (constitution V, FR-003). The CLI
// flag tier was removed when generate became interactive-only; the config file
// is now the way to change the values proposed in the form (FR-018). Detection
// is best-effort and never required. Derived values (registry default, version
// package) are computed after the merge. dir is the detection root and is always
// "." for the real commands; tests pass an isolated directory.
func buildProfile(dir string, cfg config.Config) scaffold.ProjectProfile {
	if dir == "" {
		dir = "."
	}

	mod, hasMod := detect.FromGoMod(dir)
	rem, hasRem := detect.FromGitConfig(dir)

	// module: config > detected
	module := strOr(cfg.Module, "")
	if module == "" && hasMod {
		module = mod.Path
	}

	// name: config > detected (go.mod base, else dir base)
	name := strOr(cfg.Name, "")
	if name == "" {
		switch {
		case hasMod:
			name = mod.Name
		default:
			if abs, err := filepath.Abs(dir); err == nil {
				name = filepath.Base(abs)
			}
		}
	}

	// binary: config > name
	binary := strOr(cfg.Binary, "")
	if binary == "" {
		binary = name
	}

	// platform: env > config > detected ("none" normalizes to baseline)
	platform := os.Getenv("SCAFFOLD_PLATFORM")
	if platform == "" {
		platform = strOr(cfg.Platform, "")
	}
	if platform == "" && hasRem {
		platform = rem.Platform
	}
	if platform == "none" {
		platform = ""
	}

	// owner: env > config > detected
	owner := os.Getenv("SCAFFOLD_OWNER")
	if owner == "" {
		owner = strOr(cfg.Owner, "")
	}
	if owner == "" && hasRem {
		owner = rem.Owner
	}

	host := ""
	if hasRem {
		host = rem.Host
	}

	// docker: env > config > false
	docker := false
	if v := os.Getenv("SCAFFOLD_DOCKER"); v == "1" || v == "true" {
		docker = true
	}
	if !docker {
		docker = boolOr(cfg.Docker, false)
	}

	// registry: env > config > (derived later when docker)
	registry := os.Getenv("SCAFFOLD_REGISTRY")
	if registry == "" {
		registry = strOr(cfg.Registry, "")
	}

	// mainPath: config > default ./cmd/<binary>
	mainPath := strOr(cfg.MainPath, "")
	if mainPath == "" {
		mainPath = "./cmd/" + binary
	}

	// homebrew: env > config > false
	homebrew := false
	if v := os.Getenv("SCAFFOLD_HOMEBREW"); v == "1" || v == "true" {
		homebrew = true
	}
	if !homebrew {
		homebrew = boolOr(cfg.Homebrew, false)
	}

	// homebrewTap: env > config > default
	homebrewTap := os.Getenv("SCAFFOLD_HOMEBREW_TAP")
	if homebrewTap == "" {
		homebrewTap = strOr(cfg.HomebrewTap, "")
	}
	if homebrewTap == "" {
		homebrewTap = "homebrew-tools"
	}

	plat := scaffold.PlatformID(platform)
	if docker && registry == "" {
		registry = scaffold.RegistryDefault(plat, host, owner, binary)
	}

	versionPackage := "internal/cli"
	if module != "" {
		versionPackage = module + "/internal/cli"
	}

	return scaffold.ProjectProfile{
		ProjectName:       name,
		Binary:            binary,
		ModulePath:        module,
		Owner:             owner,
		Host:              host,
		Platform:          plat,
		Docker:            docker,
		MainPath:          mainPath,
		Registry:          registry,
		VersionPackage:    versionPackage,
		GoVersion:         strOr(cfg.GoVersion, scaffold.DefaultGoVersion),
		TaskVersion:       strOr(cfg.TaskVersion, scaffold.DefaultTaskVersion),
		GolangciVersion:   strOr(cfg.GolangciVersion, scaffold.DefaultGolangciVersion),
		GoreleaserVersion: strOr(cfg.GoreleaserVersion, scaffold.DefaultGoreleaserVersion),
		FundingUser:       owner,
		Homebrew:          homebrew,
		HomebrewTap:       homebrewTap,
	}
}
