package cli

import (
	"os"
	"path/filepath"

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

// buildProfile resolves a ProjectProfile with precedence flags > env > detected
// > default (constitution V). Detection is best-effort and never required.
func buildProfile(f profileFlags) scaffold.ProjectProfile {
	dir := f.dir
	if dir == "" {
		dir = "."
	}

	mod, hasMod := detect.FromGoMod(dir)
	rem, hasRem := detect.FromGitConfig(dir)

	module := f.module
	if module == "" && hasMod {
		module = mod.Path
	}

	name := f.name
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

	binary := f.binary
	if binary == "" {
		binary = name
	}

	platform := f.platform
	if platform == "" {
		platform = os.Getenv("SCAFFOLD_PLATFORM")
	}
	if platform == "" && hasRem {
		platform = rem.Platform
	}

	owner := f.owner
	if owner == "" {
		owner = os.Getenv("SCAFFOLD_OWNER")
	}
	if owner == "" && hasRem {
		owner = rem.Owner
	}

	host := ""
	if hasRem {
		host = rem.Host
	}

	docker := f.docker
	if !docker {
		if v := os.Getenv("SCAFFOLD_DOCKER"); v == "1" || v == "true" {
			docker = true
		}
	}

	registry := f.registry
	if registry == "" {
		registry = os.Getenv("SCAFFOLD_REGISTRY")
	}

	mainPath := f.mainPath
	if mainPath == "" {
		mainPath = "./cmd/" + binary
	}

	homebrew := f.homebrew
	if !homebrew {
		if v := os.Getenv("SCAFFOLD_HOMEBREW"); v == "1" || v == "true" {
			homebrew = true
		}
	}

	homebrewTap := f.homebrewTap
	if homebrewTap == "" {
		homebrewTap = os.Getenv("SCAFFOLD_HOMEBREW_TAP")
	}
	if homebrewTap == "" {
		homebrewTap = "homebrew-tap"
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
		GoVersion:         scaffold.DefaultGoVersion,
		TaskVersion:       scaffold.DefaultTaskVersion,
		GolangciVersion:   scaffold.DefaultGolangciVersion,
		GoreleaserVersion: scaffold.DefaultGoreleaserVersion,
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
	set.StringVar(&flags.homebrewTap, "homebrew-tap", "", "Homebrew tap repo name (with --homebrew; default homebrew-tap)")
}
