package scaffold_test

import (
	"errors"
	"testing"

	"github.com/sgaunet/scaffold/internal/scaffold"
)

func baseProfile() scaffold.ProjectProfile {
	return scaffold.ProjectProfile{
		ProjectName: "demo",
		Binary:      "demo",
		ModulePath:  "github.com/acme/demo",
		MainPath:    "./cmd/demo",
		Platform:    scaffold.PlatformGitHub,
	}
}

func TestProjectProfileValidate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		mutate  func(*scaffold.ProjectProfile)
		wantErr bool
	}{
		{"valid", func(*scaffold.ProjectProfile) {}, false},
		{"empty platform is valid", func(p *scaffold.ProjectProfile) { p.Platform = scaffold.PlatformNone }, false},
		{"empty name", func(p *scaffold.ProjectProfile) { p.ProjectName = "" }, true},
		{"name with spaces", func(p *scaffold.ProjectProfile) { p.ProjectName = "my app" }, true},
		{"name with slash", func(p *scaffold.ProjectProfile) { p.ProjectName = "a/b" }, true},
		{"uppercase name", func(p *scaffold.ProjectProfile) { p.ProjectName = "Demo" }, true},
		{"bad binary", func(p *scaffold.ProjectProfile) { p.Binary = "" }, true},
		{"unknown platform", func(p *scaffold.ProjectProfile) { p.Platform = "bitbucket" }, true},
		{"main path not relative", func(p *scaffold.ProjectProfile) { p.MainPath = "cmd/demo" }, true},
		{"docker without registry", func(p *scaffold.ProjectProfile) { p.Docker = true; p.Registry = "" }, true},
		{"docker with registry", func(p *scaffold.ProjectProfile) { p.Docker = true; p.Registry = "ghcr.io/acme/demo" }, false},
		{"homebrew on github", func(p *scaffold.ProjectProfile) { p.Homebrew = true; p.HomebrewTap = "homebrew-tap" }, false},
		{"homebrew off ignores tap", func(p *scaffold.ProjectProfile) { p.Homebrew = false; p.HomebrewTap = "" }, false},
		{"homebrew not github", func(p *scaffold.ProjectProfile) {
			p.Homebrew = true
			p.HomebrewTap = "homebrew-tap"
			p.Platform = scaffold.PlatformGitLab
		}, true},
		{"homebrew empty tap", func(p *scaffold.ProjectProfile) { p.Homebrew = true; p.HomebrewTap = "" }, true},
		{"homebrew invalid tap", func(p *scaffold.ProjectProfile) { p.Homebrew = true; p.HomebrewTap = "Bad Tap" }, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p := baseProfile()
			tt.mutate(&p)
			err := p.Validate()
			if tt.wantErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantErr && !errors.Is(err, scaffold.ErrUsage) {
				t.Fatalf("error should wrap ErrUsage, got %v", err)
			}
		})
	}
}
