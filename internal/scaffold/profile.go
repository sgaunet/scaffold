package scaffold

import (
	"fmt"
	"regexp"
	"strings"
)

// nameRe validates project/binary names (also valid as module/image names).
var nameRe = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]*$`)

// ProjectProfile holds the resolved inputs that parameterize one generation run.
// It is built by the CLI layer (flags + env + detection) and validated here
// before planning. See specs/.../data-model.md.
type ProjectProfile struct {
	ProjectName       string
	Binary            string
	ModulePath        string
	Owner             string
	Host              string
	Platform          PlatformID
	Docker            bool
	MainPath          string
	Registry          string
	VersionPackage    string
	GoVersion         string
	TaskVersion       string
	GolangciVersion   string
	GoreleaserVersion string
	FundingUser       string
}

// Validate checks the profile and returns a wrapped ErrUsage on bad input
// (FR-014). Platform is optional; when set it must be one of the known values.
func (p ProjectProfile) Validate() error {
	if !nameRe.MatchString(p.ProjectName) {
		return fmt.Errorf("%w: invalid project name %q: must match %s", ErrUsage, p.ProjectName, nameRe.String())
	}
	if !nameRe.MatchString(p.Binary) {
		return fmt.Errorf("%w: invalid binary name %q: must match %s", ErrUsage, p.Binary, nameRe.String())
	}
	if p.Platform != PlatformNone {
		if _, ok := KnownPlatform(p.Platform); !ok {
			return fmt.Errorf("%w: unknown platform %q: valid platforms are %s",
				ErrUsage, p.Platform, strings.Join(ValidPlatformIDs(), ", "))
		}
	}
	if !strings.HasPrefix(p.MainPath, "./") {
		return fmt.Errorf("%w: invalid main path %q: must be a relative path beginning './'", ErrUsage, p.MainPath)
	}
	if p.Docker && p.Registry == "" {
		return fmt.Errorf("%w: --registry could not be resolved but --docker is set; pass --registry", ErrUsage)
	}
	return nil
}
