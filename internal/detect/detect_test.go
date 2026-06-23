package detect_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sgaunet/scaffold/internal/detect"
)

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestFromGoMod(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "go.mod"), "module github.com/sgaunet/semver\n\ngo 1.25\n")
	mod, ok := detect.FromGoMod(dir)
	if !ok {
		t.Fatal("expected ok")
	}
	if mod.Path != "github.com/sgaunet/semver" || mod.Name != "semver" {
		t.Fatalf("got %+v", mod)
	}
}

func TestFromGoModAbsent(t *testing.T) {
	t.Parallel()
	if _, ok := detect.FromGoMod(t.TempDir()); ok {
		t.Fatal("expected ok=false when go.mod absent")
	}
}

func TestFromGitConfig(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		url      string
		host     string
		owner    string
		platform string
	}{
		{"https github", "https://github.com/sgaunet/semver.git", "github.com", "sgaunet", "github"},
		{"ssh forgejo", "git@git.sylvlab.fr:sgaunet/hikinglist.git", "git.sylvlab.fr", "sgaunet", "forgejo"},
		{"https gitlab", "https://gitlab.com/acme/widget.git", "gitlab.com", "acme", "gitlab"},
		{"ssh scheme", "ssh://git@github.com/owner/repo", "github.com", "owner", "github"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			cfg := "[core]\n\trepositoryformatversion = 0\n[remote \"origin\"]\n\turl = " + tt.url + "\n\tfetch = +refs/heads/*:refs/remotes/origin/*\n"
			writeFile(t, filepath.Join(dir, ".git", "config"), cfg)
			rem, ok := detect.FromGitConfig(dir)
			if !ok {
				t.Fatal("expected ok")
			}
			if rem.Host != tt.host || rem.Owner != tt.owner || rem.Platform != tt.platform {
				t.Fatalf("got %+v, want host=%s owner=%s platform=%s", rem, tt.host, tt.owner, tt.platform)
			}
		})
	}
}

func TestFromGitConfigAbsent(t *testing.T) {
	t.Parallel()
	if _, ok := detect.FromGitConfig(t.TempDir()); ok {
		t.Fatal("expected ok=false when .git/config absent")
	}
}
