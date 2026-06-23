package detect

import (
	"os"
	"path/filepath"
	"strings"
)

// Remote holds the info inferred from a git origin remote.
type Remote struct {
	Host     string
	Owner    string
	Platform string // github | gitlab | forgejo; "" when undeterminable
}

// FromGitConfig parses dir/.git/config for the origin URL and infers
// host/owner/platform. ok is false when there is no git config or origin
// (the spec's "run outside a git repository" fallback).
func FromGitConfig(dir string) (Remote, bool) {
	data, err := os.ReadFile(filepath.Join(dir, ".git", "config"))
	if err != nil {
		return Remote{}, false
	}
	url := originURL(string(data))
	if url == "" {
		return Remote{}, false
	}
	host, owner := parseRemoteURL(url)
	if host == "" {
		return Remote{}, false
	}
	return Remote{Host: host, Owner: owner, Platform: platformForHost(host)}, true
}

// originURL extracts the url of [remote "origin"] from a git config (INI).
func originURL(cfg string) string {
	inOrigin := false
	for _, ln := range strings.Split(cfg, "\n") {
		t := strings.TrimSpace(ln)
		if strings.HasPrefix(t, "[") {
			inOrigin = t == `[remote "origin"]`
			continue
		}
		if inOrigin && strings.HasPrefix(t, "url") {
			if i := strings.Index(t, "="); i >= 0 {
				return strings.TrimSpace(t[i+1:])
			}
		}
	}
	return ""
}

// parseRemoteURL handles scp-like (git@host:owner/repo.git) and URL
// (https://host/owner/repo.git, ssh://git@host/owner/repo) forms.
func parseRemoteURL(url string) (host, owner string) {
	url = strings.TrimSuffix(url, ".git")
	switch {
	case strings.HasPrefix(url, "git@"):
		rest := strings.TrimPrefix(url, "git@")
		parts := strings.SplitN(rest, ":", 2)
		if len(parts) != 2 {
			return "", ""
		}
		return parts[0], firstSegment(parts[1])
	case strings.Contains(url, "://"):
		rest := url[strings.Index(url, "://")+len("://"):]
		if at := strings.Index(rest, "@"); at >= 0 {
			rest = rest[at+1:]
		}
		slash := strings.Index(rest, "/")
		if slash < 0 {
			return "", ""
		}
		return rest[:slash], firstSegment(rest[slash+1:])
	default:
		return "", ""
	}
}

func firstSegment(p string) string {
	p = strings.TrimPrefix(p, "/")
	if i := strings.Index(p, "/"); i >= 0 {
		return p[:i]
	}
	return p
}

// platformForHost maps a host to a forge: github.com → github, gitlab.*/
// gitlab.com → gitlab, everything else (self-hosted Gitea/Forgejo) → forgejo
// (research §3). The value is always overridable via --platform.
func platformForHost(host string) string {
	h := strings.ToLower(host)
	switch {
	case h == "github.com":
		return "github"
	case h == "gitlab.com" || strings.HasPrefix(h, "gitlab."):
		return "gitlab"
	default:
		return "forgejo"
	}
}
