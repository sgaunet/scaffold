package cli

import (
	"os"
	"path/filepath"

	"github.com/sgaunet/scaffold/internal/config"
	"github.com/sgaunet/scaffold/internal/detect"
	"github.com/sgaunet/scaffold/internal/scaffold"
	"github.com/spf13/pflag"
)

// profileFlags is the set of profile-shaping flags shared by generate and list.
type profileFlags struct {
	name        string
	binary      string
	module      string
	platform    string
	owner       string
	registry    string
	mainPath    string
	dir         string
	docker      bool
	homebrew    bool
	homebrewTap string
}

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

// buildProfile resolves a ProjectProfile applying precedence
// flags > env > config > detected > built-in default (constitution V, FR-003).
// Detection is best-effort and never required; the config tier sits just above
// it (FR-018). Derived values (registry default, version package) are computed
// after the merge, exactly as before.
func buildProfile(f profileFlags, cfg config.Config) scaffold.ProjectProfile {
	dir := f.dir
	if dir == "" {
		dir = "."
	}

	mod, hasMod := detect.FromGoMod(dir)
	rem, hasRem := detect.FromGitConfig(dir)

	// module: flag > config > detected
	module := f.module
	if module == "" {
		module = strOr(cfg.Module, "")
	}
	if module == "" && hasMod {
		module = mod.Path
	}

	// name: flag > config > detected (go.mod base, else dir base)
	name := f.name
	if name == "" {
		name = strOr(cfg.Name, "")
	}
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

	// binary: flag > config > name
	binary := f.binary
	if binary == "" {
		binary = strOr(cfg.Binary, "")
	}
	if binary == "" {
		binary = name
	}

	// platform: flag > env > config > detected ("none" normalizes to baseline)
	platform := f.platform
	if platform == "" {
		platform = os.Getenv("SCAFFOLD_PLATFORM")
	}
	if platform == "" {
		platform = strOr(cfg.Platform, "")
	}
	if platform == "" && hasRem {
		platform = rem.Platform
	}
	if platform == "none" {
		platform = ""
	}

	// owner: flag > env > config > detected
	owner := f.owner
	if owner == "" {
		owner = os.Getenv("SCAFFOLD_OWNER")
	}
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

	// docker: flag > env > config > false
	docker := f.docker
	if !docker {
		if v := os.Getenv("SCAFFOLD_DOCKER"); v == "1" || v == "true" {
			docker = true
		}
	}
	if !docker {
		docker = boolOr(cfg.Docker, false)
	}

	// registry: flag > env > config > (derived later when docker)
	registry := f.registry
	if registry == "" {
		registry = os.Getenv("SCAFFOLD_REGISTRY")
	}
	if registry == "" {
		registry = strOr(cfg.Registry, "")
	}

	// mainPath: flag > config > default ./cmd/<binary>
	mainPath := f.mainPath
	if mainPath == "" {
		mainPath = strOr(cfg.MainPath, "")
	}
	if mainPath == "" {
		mainPath = "./cmd/" + binary
	}

	// homebrew: flag > env > config > false
	homebrew := f.homebrew
	if !homebrew {
		if v := os.Getenv("SCAFFOLD_HOMEBREW"); v == "1" || v == "true" {
			homebrew = true
		}
	}
	if !homebrew {
		homebrew = boolOr(cfg.Homebrew, false)
	}

	// homebrewTap: flag > env > config > default
	homebrewTap := f.homebrewTap
	if homebrewTap == "" {
		homebrewTap = os.Getenv("SCAFFOLD_HOMEBREW_TAP")
	}
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

// addProfileFlags registers the shared profile-shaping flags on a command.
func addProfileFlags(flags *profileFlags, set *pflag.FlagSet) {
	set.StringVar(&flags.name, "name", "", "project name (default: go.mod module base, else dir name)")
	set.StringVar(&flags.binary, "binary", "", "binary/image name (default: --name)")
	set.StringVar(&flags.module, "module", "", "module path (default: parsed from go.mod)")
	set.StringVar(&flags.platform, "platform", "", "forge platform: github|gitlab|forgejo (optional)")
	set.StringVar(&flags.owner, "owner", "", "repo owner (default: derived from remote/module)")
	set.StringVar(&flags.registry, "registry", "", "image registry base (docker only)")
	set.StringVar(&flags.mainPath, "main", "", "main package dir (default: ./cmd/<binary>)")
	set.StringVarP(&flags.dir, "dir", "C", ".", "target project directory")
	set.BoolVar(&flags.docker, "docker", false, "enable container support")
	set.BoolVar(&flags.homebrew, "homebrew", false, "publish a Homebrew formula on release (github only)")
	set.StringVar(&flags.homebrewTap, "homebrew-tap", "", "Homebrew tap repo name (with --homebrew; default homebrew-tools)")
}
