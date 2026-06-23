package scaffold

// PlatformID identifies a target forge platform. The empty value means
// "no platform" — only the platform-independent baseline is generated (FR-003).
type PlatformID string

// Known platform identifiers.
const (
	PlatformNone    PlatformID = ""
	PlatformGitHub  PlatformID = "github"
	PlatformGitLab  PlatformID = "gitlab"
	PlatformForgejo PlatformID = "forgejo"
)

// Platform is the static description of one forge's CI outputs.
type Platform struct {
	ID              PlatformID
	CIDir           string
	CIFiles         []string
	ExtraFiles      []string
	ReleaseTokenEnv string
}

var platformTable = map[PlatformID]Platform{
	PlatformGitHub: {
		ID:              PlatformGitHub,
		CIDir:           ".github/workflows",
		CIFiles:         []string{"linter", "test", "snapshot", "release"},
		ExtraFiles:      []string{"dependabot", "FUNDING"},
		ReleaseTokenEnv: "GITHUB_TOKEN",
	},
	PlatformGitLab: {
		ID:              PlatformGitLab,
		CIDir:           "",
		CIFiles:         []string{"gitlab-ci"},
		ReleaseTokenEnv: "GITLAB_TOKEN",
	},
	PlatformForgejo: {
		ID:              PlatformForgejo,
		CIDir:           ".forgejo/workflows",
		CIFiles:         []string{"lint", "test", "snapshot", "release"},
		ReleaseTokenEnv: "GITEA_TOKEN",
	},
}

// KnownPlatform returns the Platform for id; ok is false for unknown/none.
func KnownPlatform(id PlatformID) (Platform, bool) {
	p, ok := platformTable[id]
	return p, ok
}

// ValidPlatformIDs lists the selectable platforms, for error messages.
func ValidPlatformIDs() []string {
	return []string{string(PlatformGitHub), string(PlatformGitLab), string(PlatformForgejo)}
}

// RegistryDefault computes the default image registry base for a platform.
func RegistryDefault(id PlatformID, host, owner, name string) string {
	switch id {
	case PlatformGitHub:
		return "ghcr.io/" + owner + "/" + name
	case PlatformGitLab:
		return "registry.gitlab.com/" + owner + "/" + name
	case PlatformForgejo:
		if host == "" {
			host = "git.example.com"
		}
		return host + "/" + owner + "/" + name
	case PlatformNone:
		return ""
	default:
		return ""
	}
}
