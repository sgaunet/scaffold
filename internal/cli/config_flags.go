package cli

import (
	"fmt"
	"os"

	"github.com/sgaunet/scaffold/internal/config"
	"github.com/sgaunet/scaffold/internal/scaffold"
	"github.com/spf13/pflag"
	"golang.org/x/term"
)

// configFlags are the config-file controls shared by generate and list.
type configFlags struct {
	path     string // --config: explicit config file path
	noConfig bool   // --no-config: skip config loading entirely
}

// addConfigFlags registers --config and --no-config on a command.
func addConfigFlags(cf *configFlags, set *pflag.FlagSet) {
	set.StringVar(&cf.path, "config", "",
		"config file path (default: $XDG_CONFIG_HOME/scaffold/config.yml, else $HOME/.config/scaffold/config.yml)")
	set.BoolVar(&cf.noConfig, "no-config", false, "do not load any config file")
}

// resolveConfig loads the config file per flags+env, emits diagnostics on stderr,
// and returns the parsed defaults. A load failure is wrapped as a usage error so
// it maps to exit code 2 (constitution IV); nothing is written. The explicit
// path comes from --config, else SCAFFOLD_CONFIG; --no-config overrides both.
func (g *globalOpts) resolveConfig(cf configFlags) (config.Config, error) {
	opts := config.Options{Disabled: cf.noConfig}
	if !cf.noConfig {
		opts.ExplicitPath = cf.path
		if opts.ExplicitPath == "" {
			opts.ExplicitPath = os.Getenv("SCAFFOLD_CONFIG")
		}
	}

	res, err := config.Load(opts)
	if err != nil {
		return config.Config{}, fmt.Errorf("%w: %w", scaffold.ErrUsage, err)
	}
	if res.Path != "" {
		g.vlogf("loaded config from %s", res.Path)
	}
	for _, k := range res.UnknownKeys {
		g.logf("scaffold: warning: unknown config key %q (ignored)", k)
	}
	return res.Config, nil
}

// isTerminal reports whether fd is an interactive terminal. It is a variable so
// tests can substitute a deterministic stub.
var isTerminal = func(fd uintptr) bool { return term.IsTerminal(int(fd)) }

// requireTerminal enforces that interactive mode runs only on a real terminal
// (FR-014, constitution IV). Returns a usage error otherwise so `-i` in a pipe,
// under --quiet, or with no TTY fails loudly (exit 2) instead of hanging.
func (g *globalOpts) requireTerminal() error {
	if g.quiet || !isTerminal(os.Stdin.Fd()) || !isTerminal(os.Stderr.Fd()) {
		return fmt.Errorf("%w: interactive mode requires a terminal (remove -i, or use flags/config)", scaffold.ErrUsage)
	}
	return nil
}
